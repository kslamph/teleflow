# PromptConfig Rendering Engine Design

## Overview

The PromptConfig rendering engine is the core component responsible for transforming declarative `PromptConfig` objects into actual Telegram messages with images and keyboards. This engine bridges the gap between the simplified new API and the sophisticated existing TeleFlow infrastructure.

## Engine Architecture

### 1. Core Components

```go
// PromptRenderer is the main engine interface
type PromptRenderer struct {
    bot              *Bot
    templateEngine   *TemplateEngine
    keyboardBuilder  *KeyboardBuilder
    imageHandler     *ImageHandler
    cacheManager     *CacheManager
}

// PromptRenderContext provides execution context for rendering
type PromptRenderContext struct {
    ctx           *Context
    promptConfig  *PromptConfig
    stepName      string
    flowName      string
    renderTime    time.Time
    cacheKey      string
}
```

### 2. Rendering Pipeline

```
PromptConfig → Message Evaluation → Image Processing → Keyboard Generation → Telegram Delivery
     ↓              ↓                    ↓                ↓                    ↓
 Validate       Template/Func        Base64/URL         Dynamic/Static    API Call
 Config         Execution            Handling           Generation        Formatting
```

## Detailed Component Design

### 1. Message Evaluation Engine

```go
type MessageEvaluator struct {
    templateEngine *template.Template
    funcCache      map[string]MessageFuncResult
    mu             sync.RWMutex
}

type MessageFuncResult struct {
    Result    string
    Error     error
    Timestamp time.Time
    TTL       time.Duration
}

func (me *MessageEvaluator) EvaluateMessage(config *PromptConfig, ctx *Context) (string, error) {
    switch msg := config.Message.(type) {
    case string:
        return me.processStaticMessage(msg, ctx)
    case func(*Context) string:
        return me.processDynamicMessage(msg, ctx)
    default:
        return "", fmt.Errorf("unsupported message type: %T", msg)
    }
}

func (me *MessageEvaluator) processStaticMessage(msg string, ctx *Context) (string, error) {
    // Check if message contains template syntax
    if strings.Contains(msg, "{{") {
        return me.executeTemplate(msg, ctx)
    }
    return msg, nil
}

func (me *MessageEvaluator) processDynamicMessage(msgFunc func(*Context) string, ctx *Context) (string, error) {
    // Execute function with error recovery
    var result string
    var err error
    
    func() {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("message function panic: %v", r)
            }
        }()
        result = msgFunc(ctx)
    }()
    
    if err != nil {
        return "", err
    }
    
    return result, nil
}

func (me *MessageEvaluator) executeTemplate(templateStr string, ctx *Context) (string, error) {
    // Integration with existing template system
    tmplName := fmt.Sprintf("prompt_%s_%d", ctx.UserID(), time.Now().UnixNano())
    
    // Use existing Bot.addTemplateInternal for consistency
    if err := ctx.Bot.addTemplateInternal(tmplName, templateStr, ParseModeNone, false); err != nil {
        return "", fmt.Errorf("template creation failed: %w", err)
    }
    
    // Execute using existing Context.executeTemplate
    result, _, err := ctx.executeTemplate(tmplName, ctx.data)
    return result, err
}
```

### 2. Image Processing Engine

```go
type ImageHandler struct {
    maxSize      int64  // Maximum image size in bytes
    allowedTypes []string
    cacheDir     string
    urlClient    *http.Client
}

type ImageProcessResult struct {
    Data     []byte
    Type     ImageType
    CacheKey string
}

type ImageType int

const (
    ImageTypeBase64 ImageType = iota
    ImageTypeURL
    ImageTypeFile
    ImageTypeFunction
)

func (ih *ImageHandler) ProcessImage(imageSpec interface{}, ctx *Context) (*ImageProcessResult, error) {
    switch img := imageSpec.(type) {
    case string:
        return ih.processImageString(img, ctx)
    case func(*Context) string:
        return ih.processImageFunction(img, ctx)
    default:
        return nil, fmt.Errorf("unsupported image type: %T", img)
    }
}

func (ih *ImageHandler) processImageString(img string, ctx *Context) (*ImageProcessResult, error) {
    // Detect image type
    if strings.HasPrefix(img, "data:image/") {
        return ih.processBase64Image(img)
    } else if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
        return ih.processURLImage(img, ctx)
    } else if strings.HasPrefix(img, "/") || strings.HasPrefix(img, "./") {
        return ih.processFileImage(img)
    }
    
    // Assume base64 without prefix
    return ih.processBase64Image("data:image/jpeg;base64," + img)
}

func (ih *ImageHandler) processBase64Image(base64Data string) (*ImageProcessResult, error) {
    // Extract and validate base64 data
    parts := strings.Split(base64Data, ",")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid base64 image format")
    }
    
    data, err := base64.StdEncoding.DecodeString(parts[1])
    if err != nil {
        return nil, fmt.Errorf("base64 decode failed: %w", err)
    }
    
    // Validate size
    if int64(len(data)) > ih.maxSize {
        return nil, fmt.Errorf("image size %d exceeds maximum %d", len(data), ih.maxSize)
    }
    
    return &ImageProcessResult{
        Data:     data,
        Type:     ImageTypeBase64,
        CacheKey: fmt.Sprintf("base64_%x", sha256.Sum256(data)),
    }, nil
}

func (ih *ImageHandler) processURLImage(url string, ctx *Context) (*ImageProcessResult, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("url_%x", sha256.Sum256([]byte(url)))
    
    // Download with timeout and size limits
    req, err := http.NewRequestWithContext(ctx.Update.FromChat().Context(), "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := ih.urlClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to download image: %w", err)
    }
    defer resp.Body.Close()
    
    // Read with size limit
    data, err := io.ReadAll(io.LimitReader(resp.Body, ih.maxSize))
    if err != nil {
        return nil, fmt.Errorf("failed to read image data: %w", err)
    }
    
    return &ImageProcessResult{
        Data:     data,
        Type:     ImageTypeURL,
        CacheKey: cacheKey,
    }, nil
}
```

### 3. Keyboard Generation Engine

```go
type KeyboardBuilder struct {
    cache     map[string]KeyboardCacheEntry
    mu        sync.RWMutex
    cacheTTL  time.Duration
}

type KeyboardCacheEntry struct {
    Keyboard  interface{}
    Generated time.Time
}

func (kb *KeyboardBuilder) BuildKeyboard(keyboardFunc KeyboardFunc, ctx *Context) (interface{}, error) {
    if keyboardFunc == nil {
        return nil, nil
    }
    
    // Generate cache key based on context state
    cacheKey := kb.generateCacheKey(ctx)
    
    // Check cache
    if cached, exists := kb.getCachedKeyboard(cacheKey); exists {
        return cached, nil
    }
    
    // Execute keyboard function with error recovery
    var keyboardData map[string]interface{}
    var err error
    
    func() {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("keyboard function panic: %v", r)
            }
        }()
        keyboardData = keyboardFunc(ctx)
    }()
    
    if err != nil {
        return nil, err
    }
    
    // Convert to Telegram keyboard format
    keyboard, err := kb.convertToTelegramKeyboard(keyboardData)
    if err != nil {
        return nil, fmt.Errorf("keyboard conversion failed: %w", err)
    }
    
    // Cache result
    kb.cacheKeyboard(cacheKey, keyboard)
    
    return keyboard, nil
}

func (kb *KeyboardBuilder) convertToTelegramKeyboard(data map[string]interface{}) (interface{}, error) {
    // Detect keyboard type based on data structure
    if kb.isInlineKeyboard(data) {
        return kb.buildInlineKeyboard(data)
    }
    return kb.buildReplyKeyboard(data)
}

func (kb *KeyboardBuilder) buildInlineKeyboard(data map[string]interface{}) (*InlineKeyboard, error) {
    keyboard := NewInlineKeyboard()
    
    for text, callbackData := range data {
        // Handle different callback data types
        switch cd := callbackData.(type) {
        case string:
            keyboard.AddButton(text, cd)
        case map[string]interface{}:
            // Extract callback_data and other properties
            if callback, ok := cd["callback_data"].(string); ok {
                keyboard.AddButton(text, callback)
            } else {
                return nil, fmt.Errorf("missing callback_data for button: %s", text)
            }
        default:
            // Convert to string
            keyboard.AddButton(text, fmt.Sprintf("%v", cd))
        }
    }
    
    return keyboard, nil
}

func (kb *KeyboardBuilder) generateCacheKey(ctx *Context) string {
    // Create cache key based on user state and flow context
    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("user_%d", ctx.UserID())))
    
    // Include relevant context data
    for key, value := range ctx.data {
        h.Write([]byte(fmt.Sprintf("%s=%v", key, value)))
    }
    
    return fmt.Sprintf("kb_%x", h.Sum(nil))
}
```

### 4. Cache Management System

```go
type CacheManager struct {
    messageCache   map[string]MessageCacheEntry
    keyboardCache  map[string]KeyboardCacheEntry
    imageCache     map[string]ImageCacheEntry
    mu             sync.RWMutex
    cleanupTicker  *time.Ticker
}

type CacheEntry interface {
    IsExpired() bool
    GetSize() int64
}

func (cm *CacheManager) StartCleanup() {
    cm.cleanupTicker = time.NewTicker(5 * time.Minute)
    go func() {
        for range cm.cleanupTicker.C {
            cm.cleanup()
        }
    }()
}

func (cm *CacheManager) cleanup() {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    
    // Clean expired entries
    for key, entry := range cm.messageCache {
        if entry.IsExpired() {
            delete(cm.messageCache, key)
        }
    }
    
    // Similar cleanup for other caches
}
```

## Integration with Existing Systems

### 1. FlowManager Integration

```go
func (fm *FlowManager) renderPromptConfig(ctx *Context, config *PromptConfig) error {
    // Create rendering context
    renderCtx := &PromptRenderContext{
        ctx:          ctx,
        promptConfig: config,
        stepName:     fm.getCurrentStepName(ctx.UserID()),
        flowName:     fm.getCurrentFlowName(ctx.UserID()),
        renderTime:   time.Now(),
    }
    
    // Use the prompt renderer
    return ctx.Bot.promptRenderer.Render(renderCtx)
}
```

### 2. Context System Integration

```go
func (c *Context) RenderPromptConfig(config *PromptConfig) error {
    renderCtx := &PromptRenderContext{
        ctx:          c,
        promptConfig: config,
        renderTime:   time.Now(),
    }
    
    return c.Bot.promptRenderer.Render(renderCtx)
}
```

## Performance Considerations

### 1. Caching Strategy
- **Message Functions**: Cache results with configurable TTL
- **Keyboard Functions**: Cache based on context state hash
- **Images**: Persistent cache with LRU eviction
- **Templates**: Leverage existing template cache

### 2. Resource Management
- **Memory Limits**: Configure maximum cache sizes
- **Image Limits**: Restrict image size and download timeouts
- **Function Timeouts**: Prevent long-running keyboard/message functions

### 3. Error Recovery
- **Function Panics**: Graceful recovery with error messages
- **Network Failures**: Fallback to cached content
- **Invalid Data**: Clear error messages for debugging

## Configuration

```go
type PromptEngineConfig struct {
    MaxImageSize      int64
    CacheTTL          time.Duration
    MaxCacheSize      int64
    FunctionTimeout   time.Duration
    EnableCache       bool
    ImageDownloadTimeout time.Duration
}

func NewPromptRenderer(bot *Bot, config PromptEngineConfig) *PromptRenderer {
    return &PromptRenderer{
        bot:             bot,
        templateEngine:  bot.templates,
        keyboardBuilder: NewKeyboardBuilder(config.CacheTTL),
        imageHandler:    NewImageHandler(config.MaxImageSize),
        cacheManager:    NewCacheManager(config.MaxCacheSize),
    }
}
```

## Testing Strategy

### 1. Unit Tests
- Message evaluation with different input types
- Image processing for all supported formats
- Keyboard generation with various data structures
- Cache behavior and cleanup

### 2. Integration Tests
- End-to-end prompt rendering
- Error handling and recovery
- Performance under load
- Memory usage patterns

### 3. Mock External Services
- Template function calls to databases
- Image downloads from URLs
- Keyboard generation with API calls

## Next Steps

1. **Implement core PromptRenderer interface**
2. **Create MessageEvaluator with template integration**
3. **Build ImageHandler with multiple format support**
4. **Develop KeyboardBuilder with caching**
5. **Add comprehensive error handling and recovery**
6. **Create performance benchmarks and optimization**

This engine design provides a robust, performant foundation for the new Step-Prompt-Process API while maintaining full integration with existing TeleFlow systems.