# Template System Separation & Enhancement Plan

## 🎯 Objective
Separate template management from Bot struct and enhance template support in message rendering with hybrid data binding approach.

## 📋 Current Issues
1. Templates tightly coupled to `Bot` struct via `b.templates` and global `templateRegistry`
2. `context.send()` doesn't handle template parsing modes
3. `MessageSpec` only supports strings and functions, no template support
4. No clean way to pass template variables

## 🏗️ Solution Architecture

### Core Design Decisions
- **Template Detection**: Support `"template:templateName"` format in `MessageSpec`
- **Data Binding**: Hybrid approach - `TemplateData` field + Context data auto-binding
- **Data Precedence**: `TemplateData` takes precedence over Context data
- **Parse Mode**: Templates carry their own parse mode, applied during rendering

## 📝 Implementation Plan

### Phase 1: Create Standalone Template Manager

#### 1.1 Create Template Manager Interface
**File**: `core/template_manager.go`
```go
type TemplateManager interface {
    // Template registration
    AddTemplate(name, templateText string, parseMode ParseMode) error
    HasTemplate(name string) bool
    GetTemplateInfo(name string) *TemplateInfo
    ListTemplates() []string
    
    // Template rendering
    RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error)
}

type templateManager struct {
    templates *template.Template
    registry  map[string]*TemplateInfo
}
```

#### 1.2 Move Template Logic from Bot
- Move `addTemplateInternal`, template functions, and validation from `core/templates.go`
- Remove `templates` field and template methods from `Bot` struct
- Create global `defaultTemplateManager` instance
- Update `Bot.AddTemplate()` to delegate to template manager

#### 1.3 Template Data Merging Logic
```go
func (tm *templateManager) mergeTemplateData(templateData map[string]interface{}, contextData map[string]interface{}) map[string]interface{} {
    merged := make(map[string]interface{})
    
    // First add context data
    for k, v := range contextData {
        merged[k] = v
    }
    
    // Then add template data (overwrites context data)
    for k, v := range templateData {
        merged[k] = v
    }
    
    return merged
}
```

### Phase 2: Enhanced PromptConfig Structure

#### 2.1 Update PromptConfig
**File**: `core/flow_types.go`
```go
type PromptConfig struct {
    Message      MessageSpec               // Can be string, func(*Context) string, or "template:name"
    Image        ImageSpec                 // Can be string, func(*Context) string, or nil
    Keyboard     KeyboardFunc              // Can be func(*Context) map[string]interface{} or nil
    TemplateData map[string]interface{}    // NEW: Template variables (takes precedence over context)
}
```

#### 2.2 Add Template Detection Utility
```go
func isTemplateMessage(message string) (bool, string) {
    const prefix = "template:"
    if strings.HasPrefix(message, prefix) {
        return true, strings.TrimPrefix(message, prefix)
    }
    return false, ""
}
```

### Phase 3: Enhanced Message Renderer

#### 3.1 Update Message Renderer
**File**: `core/message_renderer.go`
```go
type messageRenderer struct {
    templateManager TemplateManager
}

func (mr *messageRenderer) renderMessage(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
    if config.Message == nil {
        return "", ParseModeNone, nil
    }

    switch msg := config.Message.(type) {
    case string:
        // Check if it's a template
        if isTemplate, templateName := isTemplateMessage(msg); isTemplate {
            return mr.renderTemplateMessage(templateName, config, ctx)
        }
        // Static string message
        return msg, ParseModeNone, nil

    case func(*Context) string:
        // Dynamic message function
        result := msg(ctx)
        // Check if function returned a template
        if isTemplate, templateName := isTemplateMessage(result); isTemplate {
            return mr.renderTemplateMessage(templateName, config, ctx)
        }
        return result, ParseModeNone, nil

    default:
        return "", ParseModeNone, fmt.Errorf("unsupported message type: %T", msg)
    }
}

func (mr *messageRenderer) renderTemplateMessage(templateName string, config *PromptConfig, ctx *Context) (string, ParseMode, error) {
    // Merge template data: context data + explicit template data (template data takes precedence)
    mergedData := mr.mergeDataSources(config.TemplateData, ctx.data)
    
    // Render template
    renderedText, parseMode, err := mr.templateManager.RenderTemplate(templateName, mergedData)
    if err != nil {
        return "", ParseModeNone, fmt.Errorf("template rendering failed for '%s': %w", templateName, err)
    }
    
    return renderedText, parseMode, nil
}

func (mr *messageRenderer) mergeDataSources(templateData, contextData map[string]interface{}) map[string]interface{} {
    merged := make(map[string]interface{})
    
    // First add context data
    for k, v := range contextData {
        merged[k] = v
    }
    
    // Then add template data (takes precedence)
    for k, v := range templateData {
        merged[k] = v
    }
    
    return merged
}
```

### Phase 4: Update Prompt Renderer

#### 4.1 Handle Parse Modes in Prompt Renderer
**File**: `core/prompt_renderer.go`
```go
func (pr *promptRenderer) render(renderCtx *renderContext) error {
    // Render message (now returns parse mode)
    message, parseMode, err := pr.messageRenderer.renderMessage(renderCtx.promptConfig, renderCtx.ctx)
    if err != nil {
        return pr.logFriendlyError("message rendering", renderCtx, err)
    }

    // Store parse mode for context.send()
    if parseMode != ParseModeNone {
        renderCtx.ctx.Set("__render_parse_mode", parseMode)
    }

    // Continue with existing logic...
    image, err := pr.imageHandler.processImage(renderCtx.promptConfig.Image, renderCtx.ctx)
    // ... rest of method unchanged
}
```

### Phase 5: Update Context.send() for Parse Mode Support

#### 5.1 Enhanced Context.send()
**File**: `core/context.go`
```go
func (c *Context) send(text string, keyboard ...interface{}) error {
    msg := tgbotapi.NewMessage(c.ChatID(), text)

    // Check if a parse mode was set during rendering
    if parseMode, exists := c.Get("__render_parse_mode"); exists {
        if pm, ok := parseMode.(ParseMode); ok && pm != ParseModeNone {
            msg.ParseMode = string(pm)
        }
        // Clean up the temporary parse mode
        delete(c.data, "__render_parse_mode")
    }

    // Existing keyboard and menu button logic...
    c.applyAutomaticMenuButton()
    
    // Apply keyboard markup (existing code unchanged)
    if len(keyboard) > 0 && keyboard[0] != nil {
        // ... existing keyboard logic
    } else {
        // ... existing default keyboard logic
    }

    _, err := c.bot.api.Send(msg)
    return err
}
```

### Phase 6: Convenience Methods and Backwards Compatibility

#### 6.1 Add Convenience Methods to Context
**File**: `core/context.go`
```go
// ReplyTemplate renders and sends a template with data
func (c *Context) ReplyTemplate(templateName string, data map[string]interface{}, keyboard ...interface{}) error {
    return c.SendPrompt(&PromptConfig{
        Message:      "template:" + templateName,
        TemplateData: data,
    })
}

// SendPromptWithTemplate is a convenience method for template-based prompts
func (c *Context) SendPromptWithTemplate(templateName string, data map[string]interface{}) error {
    return c.SendPrompt(&PromptConfig{
        Message:      "template:" + templateName,
        TemplateData: data,
    })
}
```

#### 6.2 Update Bot Constructor
**File**: `core/bot.go`
```go
func NewBot(token string, options ...BotOption) (*Bot, error) {
    // ... existing code ...
    
    b := &Bot{
        // Remove templates field
        // ... other fields unchanged
    }
    
    // Initialize template manager in flow manager
    b.flowManager.initialize(b)
    return b, nil
}

// Update template methods to delegate
func (b *Bot) AddTemplate(name, templateText string, parseMode ParseMode) error {
    return defaultTemplateManager.AddTemplate(name, templateText, parseMode)
}

func (b *Bot) HasTemplate(name string) bool {
    return defaultTemplateManager.HasTemplate(name)
}

// ... other template methods updated similarly
```

## 🔄 Migration Guide

### For Existing Code
1. **Template Registration**: No changes needed - `bot.AddTemplate()` still works
2. **Basic SendPrompt**: No changes needed - existing code continues to work
3. **New Template Usage**: Use `"template:templateName"` in Message field

### Example Usage After Implementation
```go
// 1. Register template (unchanged)
bot.AddTemplate("welcome", "Hello {{.Name}}! Welcome to {{.Service}}.", ParseModeHTML)

// 2. Use template with explicit data
ctx.SendPrompt(&PromptConfig{
    Message: "template:welcome",
    TemplateData: map[string]interface{}{
        "Name": "John",
        "Service": "TeleFlow",
    },
})

// 3. Use template with context data auto-binding
ctx.Set("Name", "John")
ctx.Set("Service", "TeleFlow")
ctx.SendPrompt(&PromptConfig{
    Message: "template:welcome",
    // TemplateData not provided - uses context data
})

// 4. Use template with mixed data (TemplateData takes precedence)
ctx.Set("Name", "John")      // Will be overridden
ctx.Set("Service", "TeleFlow")
ctx.SendPrompt(&PromptConfig{
    Message: "template:welcome",
    TemplateData: map[string]interface{}{
        "Name": "Jane", // Takes precedence over context data
    },
})

// 5. Convenience method
ctx.ReplyTemplate("welcome", map[string]interface{}{
    "Name": "John",
    "Service": "TeleFlow",
})
```

## 🧪 Testing Requirements

### Unit Tests Needed
1. **TemplateManager**: Template registration, rendering, data merging
2. **MessageRenderer**: Template detection, rendering with different data sources
3. **Context.send()**: Parse mode application
4. **Data Precedence**: Verify TemplateData overrides Context data

### Integration Tests Needed
1. **End-to-end template rendering** in flows
2. **Parse mode preservation** through the rendering pipeline
3. **Backwards compatibility** with existing template usage

## ✅ Acceptance Criteria

1. ✅ Templates separated from Bot struct
2. ✅ `"template:name"` format supported in MessageSpec
3. ✅ TemplateData field added to PromptConfig
4. ✅ Context data auto-binding implemented
5. ✅ TemplateData takes precedence over Context data
6. ✅ Parse modes properly applied in context.send()
7. ✅ Backwards compatibility maintained
8. ✅ Convenience methods provided
9. ✅ Comprehensive test coverage
10. ✅ Clean separation of concerns achieved

## 📁 File Changes Summary

| File | Changes |
|------|---------|
| `core/template_manager.go` | **NEW** - Standalone template management |
| `core/flow_types.go` | **MODIFY** - Add TemplateData to PromptConfig |
| `core/message_renderer.go` | **MODIFY** - Add template support and parse mode |
| `core/prompt_renderer.go` | **MODIFY** - Handle parse mode from message renderer |
| `core/context.go` | **MODIFY** - Add parse mode support, convenience methods |
| `core/bot.go` | **MODIFY** - Remove template fields, delegate to manager |
| `core/templates.go` | **MODIFY** - Move logic to template manager |

This plan provides a clear roadmap for implementing the template system separation while maintaining backwards compatibility and adding the requested hybrid data binding approach.