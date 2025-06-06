# External Service Integration Patterns for TeleFlow

## Overview

This document defines patterns and best practices for integrating external services (databases, APIs, microservices) within template functions and keyboard functions in the new Step-Prompt-Process API. The goal is to provide clean, performant, and reliable patterns that developers can follow when building complex flows that interact with external systems.

## Core Integration Principles

### 1. Dependency Injection Pattern
Services should be injected into the flow context rather than being globally accessible, enabling testability and service isolation.

### 2. Context Propagation
All external service calls should respect Go context for cancellation, timeouts, and tracing.

### 3. Error Resilience
External service failures should degrade gracefully without breaking the conversational flow.

### 4. Performance Optimization
Caching, batching, and async patterns should be used to optimize external service interactions.

## Service Injection Architecture

### 1. Service Container Pattern

```go
// ServiceContainer holds all external services available to flows
type ServiceContainer struct {
    UserService     UserServiceInterface
    OrderService    OrderServiceInterface
    PaymentService  PaymentServiceInterface
    LocationService LocationServiceInterface
    // Add more services as needed
}

// ServiceContext extends Context with service access
type ServiceContext struct {
    *Context
    Services *ServiceContainer
}

// Enhanced Context creation with services
func NewServiceContext(ctx *Context, services *ServiceContainer) *ServiceContext {
    return &ServiceContext{
        Context:  ctx,
        Services: services,
    }
}

// Bot configuration with services
type BotConfig struct {
    Services *ServiceContainer
    // Other config options
}

func NewBotWithServices(token string, config BotConfig) (*Bot, error) {
    bot, err := NewBot(token)
    if err != nil {
        return nil, err
    }
    
    // Inject services into bot
    bot.serviceContainer = config.Services
    return bot, nil
}
```

### 2. Service Interface Definitions

```go
// Example service interfaces for type safety and testability
type UserServiceInterface interface {
    GetProfile(ctx context.Context, userID string) (*UserProfile, error)
    UpdateProfile(ctx context.Context, userID string, updates UserProfileUpdates) error
    IncrementVisitCount(ctx context.Context, userID string) (int, error)
    GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error)
}

type OrderServiceInterface interface {
    GetOrderHistory(ctx context.Context, userID string, limit int) ([]*Order, error)
    CreateOrder(ctx context.Context, order *OrderRequest) (*Order, error)
    GetOrderStatus(ctx context.Context, orderID string) (*OrderStatus, error)
    CancelOrder(ctx context.Context, orderID string) error
}

type LocationServiceInterface interface {
    ValidateLocation(ctx context.Context, address string) (*LocationResult, error)
    GetShippingOptions(ctx context.Context, coordinates Coordinates) ([]*ShippingOption, error)
    GetNearbyStores(ctx context.Context, coordinates Coordinates, radius float64) ([]*Store, error)
}
```

## Template Function Integration Patterns

### 1. Basic Service Access Pattern

```go
// Template function with service access
func createWelcomeMessageTemplate(services *ServiceContainer) func(*Context) string {
    return func(ctx *Context) string {
        // Create service context with timeout
        serviceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        userID := fmt.Sprintf("%d", ctx.UserID())
        
        // Fetch user profile with error handling
        userProfile, err := services.UserService.GetProfile(serviceCtx, userID)
        if err != nil {
            log.Printf("Error fetching profile for user %s: %v", userID, err)
            return "Welcome! Please tell us a bit about yourself."
        }
        
        // Fetch visit count with fallback
        visitCount, err := services.UserService.IncrementVisitCount(serviceCtx, userID)
        if err != nil {
            log.Printf("Error incrementing visit count for user %s: %v", userID, err)
            visitCount = userProfile.VisitCount // Use cached count
        }
        
        return fmt.Sprintf("Welcome back, %s! This is visit #%d. You have %d loyalty points.",
            userProfile.Name, visitCount, userProfile.LoyaltyPoints)
    }
}

// Usage in flow definition
flow := teleflow.NewFlow("user_onboarding").
    Step("welcome").
        Prompt(createWelcomeMessageTemplate(services), "", nil).
        Process(func(ctx *Context, input string, btn *ButtonClick) ProcessResult {
            // Process logic
            return NextStep()
        })
```

### 2. Caching Pattern for Template Functions

```go
// CachedTemplateFunc provides caching for expensive template operations
type CachedTemplateFunc struct {
    cache    map[string]CacheEntry
    mu       sync.RWMutex
    ttl      time.Duration
    fetchFn  func(ctx context.Context, key string) (string, error)
}

type CacheEntry struct {
    Value     string
    Timestamp time.Time
}

func NewCachedTemplateFunc(ttl time.Duration, fetchFn func(context.Context, string) (string, error)) *CachedTemplateFunc {
    return &CachedTemplateFunc{
        cache:   make(map[string]CacheEntry),
        ttl:     ttl,
        fetchFn: fetchFn,
    }
}

func (ctf *CachedTemplateFunc) Get(ctx *Context, cacheKey string) string {
    // Check cache first
    ctf.mu.RLock()
    if entry, exists := ctf.cache[cacheKey]; exists {
        if time.Since(entry.Timestamp) < ctf.ttl {
            ctf.mu.RUnlock()
            return entry.Value
        }
    }
    ctf.mu.RUnlock()
    
    // Fetch from service
    serviceCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    value, err := ctf.fetchFn(serviceCtx, cacheKey)
    if err != nil {
        log.Printf("Cache fetch failed for key %s: %v", cacheKey, err)
        return "Loading..." // Fallback value
    }
    
    // Update cache
    ctf.mu.Lock()
    ctf.cache[cacheKey] = CacheEntry{
        Value:     value,
        Timestamp: time.Now(),
    }
    ctf.mu.Unlock()
    
    return value
}

// Usage example
func createDynamicStatsTemplate(services *ServiceContainer) func(*Context) string {
    // Create cached function for expensive stats calculation
    statsCache := NewCachedTemplateFunc(1*time.Minute, func(ctx context.Context, userID string) (string, error) {
        stats, err := services.AnalyticsService.GetUserStats(ctx, userID)
        if err != nil {
            return "", err
        }
        return fmt.Sprintf("📊 Your Stats:\n• Orders: %d\n• Savings: $%.2f\n• Rank: %s", 
            stats.OrderCount, stats.TotalSavings, stats.UserRank), nil
    })
    
    return func(ctx *Context) string {
        userID := fmt.Sprintf("%d", ctx.UserID())
        return statsCache.Get(ctx, userID)
    }
}
```

### 3. Async Template Pattern

```go
// AsyncTemplateResult holds async operation results
type AsyncTemplateResult struct {
    Content   string
    Ready     chan bool
    Error     error
    mutex     sync.Mutex
}

// AsyncTemplateManager manages background template operations
type AsyncTemplateManager struct {
    operations map[string]*AsyncTemplateResult
    mu         sync.RWMutex
}

func NewAsyncTemplateManager() *AsyncTemplateManager {
    return &AsyncTemplateManager{
        operations: make(map[string]*AsyncTemplateResult),
    }
}

func (atm *AsyncTemplateManager) StartOperation(key string, operation func() (string, error)) *AsyncTemplateResult {
    atm.mu.Lock()
    defer atm.mu.Unlock()
    
    // Check if operation already exists
    if result, exists := atm.operations[key]; exists {
        return result
    }
    
    // Create new async operation
    result := &AsyncTemplateResult{
        Ready: make(chan bool, 1),
    }
    atm.operations[key] = result
    
    // Start operation in background
    go func() {
        defer func() {
            result.Ready <- true
            close(result.Ready)
        }()
        
        content, err := operation()
        result.mutex.Lock()
        result.Content = content
        result.Error = err
        result.mutex.Unlock()
    }()
    
    return result
}

// Usage for slow operations like AI-generated content
func createAIGeneratedTemplate(services *ServiceContainer, asyncManager *AsyncTemplateManager) func(*Context) string {
    return func(ctx *Context) string {
        userID := fmt.Sprintf("%d", ctx.UserID())
        operationKey := fmt.Sprintf("ai_content_%s", userID)
        
        // Start or get existing async operation
        result := asyncManager.StartOperation(operationKey, func() (string, error) {
            // Simulate AI content generation (expensive operation)
            serviceCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()
            
            profile, err := services.UserService.GetProfile(serviceCtx, userID)
            if err != nil {
                return "", err
            }
            
            return services.AIService.GeneratePersonalizedContent(serviceCtx, profile)
        })
        
        // Check if result is ready (non-blocking)
        select {
        case <-result.Ready:
            result.mutex.Lock()
            defer result.mutex.Unlock()
            if result.Error != nil {
                return "Sorry, we're having trouble generating personalized content. Please try again later."
            }
            return result.Content
        default:
            // Operation still running, return placeholder
            return "🤖 Generating personalized content for you... Please wait a moment."
        }
    }
}
```

## Keyboard Function Integration Patterns

### 1. Dynamic Service-Based Keyboards

```go
// Dynamic keyboard generation from service data
func createShippingOptionsKeyboard(services *ServiceContainer) KeyboardFunc {
    return func(ctx *Context) map[string]interface{} {
        // Get user's location from context
        locationData, ok := ctx.Get("location_coords")
        if !ok {
            return map[string]interface{}{
                "❌ Location not set": "no_location",
            }
        }
        
        coordinates, ok := locationData.(Coordinates)
        if !ok {
            return map[string]interface{}{
                "❌ Invalid location": "invalid_location",
            }
        }
        
        // Fetch shipping options from service
        serviceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        options, err := services.LocationService.GetShippingOptions(serviceCtx, coordinates)
        if err != nil {
            log.Printf("Error fetching shipping options: %v", err)
            return map[string]interface{}{
                "❌ Unable to load options": "error_loading",
                "🔄 Try again": "retry_shipping",
            }
        }
        
        // Build keyboard from options
        keyboard := make(map[string]interface{})
        for _, option := range options {
            buttonText := fmt.Sprintf("%s - $%.2f (%s)", option.Name, option.Price, option.EstimatedTime)
            keyboard[buttonText] = option.ID
        }
        
        return keyboard
    }
}
```

### 2. Paginated Keyboard Pattern

```go
// PaginatedKeyboard manages large datasets with pagination
type PaginatedKeyboard struct {
    services    *ServiceContainer
    pageSize    int
    currentPage int
    totalItems  int
    cacheKey    string
}

func createPaginatedOrderHistoryKeyboard(services *ServiceContainer) KeyboardFunc {
    return func(ctx *Context) map[string]interface{} {
        userID := fmt.Sprintf("%d", ctx.UserID())
        
        // Get current page from context (default to 0)
        currentPage := 0
        if pageData, ok := ctx.Get("order_history_page"); ok {
            if page, ok := pageData.(int); ok {
                currentPage = page
            }
        }
        
        // Fetch paginated orders
        serviceCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        
        pageSize := 5
        orders, err := services.OrderService.GetOrderHistory(serviceCtx, userID, pageSize, currentPage*pageSize)
        if err != nil {
            return map[string]interface{}{
                "❌ Error loading orders": "error_orders",
            }
        }
        
        keyboard := make(map[string]interface{})
        
        // Add order items
        for _, order := range orders {
            buttonText := fmt.Sprintf("Order #%s - $%.2f (%s)", 
                order.ID[:8], order.Total, order.Status)
            keyboard[buttonText] = fmt.Sprintf("order_%s", order.ID)
        }
        
        // Add pagination controls
        if currentPage > 0 {
            keyboard["⬅️ Previous"] = fmt.Sprintf("page_%d", currentPage-1)
        }
        
        if len(orders) == pageSize {
            keyboard["➡️ Next"] = fmt.Sprintf("page_%d", currentPage+1)
        }
        
        return keyboard
    }
}
```

### 3. Real-time Data Keyboard

```go
// Real-time keyboard with WebSocket or polling updates
type RealTimeKeyboard struct {
    services     *ServiceContainer
    wsConnection *websocket.Conn
    dataChannel  chan KeyboardUpdate
    lastUpdate   time.Time
}

type KeyboardUpdate struct {
    Data      map[string]interface{}
    Timestamp time.Time
    Error     error
}

func createLiveStockKeyboard(services *ServiceContainer) KeyboardFunc {
    return func(ctx *Context) map[string]interface{} {
        // For real-time data, we might want to implement polling or WebSocket updates
        // For now, we'll use a short cache with frequent updates
        
        serviceCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        
        stocks, err := services.StockService.GetAvailableStock(serviceCtx)
        if err != nil {
            return map[string]interface{}{
                "❌ Unable to load stock": "error_stock",
                "🔄 Refresh": "refresh_stock",
            }
        }
        
        keyboard := make(map[string]interface{})
        
        for _, stock := range stocks {
            statusIcon := "✅"
            if stock.Quantity < 5 {
                statusIcon = "⚠️"
            } else if stock.Quantity == 0 {
                statusIcon = "❌"
                continue // Don't show out of stock items
            }
            
            buttonText := fmt.Sprintf("%s %s - %d left ($%.2f)", 
                statusIcon, stock.Name, stock.Quantity, stock.Price)
            keyboard[buttonText] = stock.ID
        }
        
        // Add refresh option with timestamp
        keyboard["🔄 Refresh Stock"] = fmt.Sprintf("refresh_%d", time.Now().Unix())
        
        return keyboard
    }
}
```

## Error Handling and Resilience Patterns

### 1. Circuit Breaker Pattern

```go
// ServiceCircuitBreaker protects against cascading failures
type ServiceCircuitBreaker struct {
    failureCount    int
    lastFailureTime time.Time
    threshold       int
    timeout         time.Duration
    state           CircuitState
    mu              sync.RWMutex
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

func (cb *ServiceCircuitBreaker) Call(operation func() (interface{}, error)) (interface{}, error) {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()
    
    if state == CircuitOpen {
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.mu.Lock()
            cb.state = CircuitHalfOpen
            cb.mu.Unlock()
        } else {
            return nil, fmt.Errorf("circuit breaker is open")
        }
    }
    
    result, err := operation()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()
        
        if cb.failureCount >= cb.threshold {
            cb.state = CircuitOpen
        }
        
        return nil, err
    }
    
    // Success - reset circuit breaker
    cb.failureCount = 0
    cb.state = CircuitClosed
    return result, nil
}
```

### 2. Fallback Content Pattern

```go
// FallbackTemplateFunc provides fallback content when services fail
type FallbackTemplateFunc struct {
    primaryFn   func(*Context) string
    fallbackFn  func(*Context) string
    maxRetries  int
    retryDelay  time.Duration
}

func NewFallbackTemplateFunc(primary, fallback func(*Context) string) *FallbackTemplateFunc {
    return &FallbackTemplateFunc{
        primaryFn:   primary,
        fallbackFn:  fallback,
        maxRetries:  3,
        retryDelay:  1 * time.Second,
    }
}

func (ftf *FallbackTemplateFunc) Execute(ctx *Context) string {
    for attempt := 0; attempt < ftf.maxRetries; attempt++ {
        // Try primary function with timeout
        result := make(chan string, 1)
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    result <- ""
                }
            }()
            result <- ftf.primaryFn(ctx)
        }()
        
        select {
        case content := <-result:
            if content != "" {
                return content
            }
        case <-time.After(5 * time.Second):
            log.Printf("Template function timeout on attempt %d", attempt+1)
        }
        
        if attempt < ftf.maxRetries-1 {
            time.Sleep(ftf.retryDelay)
        }
    }
    
    // All retries failed, use fallback
    log.Printf("Primary template function failed, using fallback")
    return ftf.fallbackFn(ctx)
}
```

## Testing Patterns

### 1. Service Mocking

```go
// Mock services for testing
type MockUserService struct {
    profiles map[string]*UserProfile
    errors   map[string]error
}

func (m *MockUserService) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
    if err, exists := m.errors[userID]; exists {
        return nil, err
    }
    
    if profile, exists := m.profiles[userID]; exists {
        return profile, nil
    }
    
    return nil, fmt.Errorf("user not found: %s", userID)
}

func (m *MockUserService) SetProfile(userID string, profile *UserProfile) {
    m.profiles[userID] = profile
}

func (m *MockUserService) SetError(userID string, err error) {
    m.errors[userID] = err
}

// Test template function with mock service
func TestTemplateWithMockService(t *testing.T) {
    mockService := &MockUserService{
        profiles: make(map[string]*UserProfile),
        errors:   make(map[string]error),
    }
    
    // Set up test data
    mockService.SetProfile("123", &UserProfile{
        Name:   "Test User",
        Points: 100,
    })
    
    services := &ServiceContainer{
        UserService: mockService,
    }
    
    templateFunc := createWelcomeMessageTemplate(services)
    
    // Create test context
    ctx := &Context{
        userID: 123,
        data:   make(map[string]interface{}),
    }
    
    result := templateFunc(ctx)
    expected := "Welcome back, Test User! This is visit #1. You have 100 loyalty points."
    
    if result != expected {
        t.Errorf("Expected %q, got %q", expected, result)
    }
}
```

### 2. Integration Testing

```go
// Integration test with real services (using test databases)
func TestFlowWithRealServices(t *testing.T) {
    // Set up test database
    testDB := setupTestDatabase(t)
    defer cleanupTestDatabase(t, testDB)
    
    // Create real services with test database
    services := &ServiceContainer{
        UserService:     NewUserService(testDB),
        OrderService:    NewOrderService(testDB),
        LocationService: NewLocationService(testConfig),
    }
    
    // Create bot with services
    bot, err := NewBotWithServices("test_token", BotConfig{
        Services: services,
    })
    if err != nil {
        t.Fatalf("Failed to create bot: %v", err)
    }
    
    // Test flow with real service interactions
    flow := createDeliveryOnboardingFlow(services)
    bot.RegisterFlow(flow)
    
    // Simulate user interaction
    ctx := createTestContext(t, bot, 12345)
    
    err = ctx.StartFlow("delivery_onboarding")
    if err != nil {
        t.Fatalf("Failed to start flow: %v", err)
    }
    
    // Continue with flow testing...
}
```

## Best Practices Summary

### 1. Service Design
- **Use interfaces** for all services to enable testing and flexibility
- **Implement timeouts** for all external service calls
- **Design for failure** with circuit breakers and fallbacks
- **Cache frequently accessed data** with appropriate TTL

### 2. Template Functions
- **Keep functions pure** when possible (deterministic output for same input)
- **Handle errors gracefully** with fallback content
- **Use context cancellation** for long-running operations
- **Implement caching** for expensive operations

### 3. Keyboard Functions
- **Limit keyboard size** to maintain usability
- **Handle empty results** gracefully
- **Use pagination** for large datasets
- **Provide refresh options** for dynamic content

### 4. Performance
- **Implement connection pooling** for database services
- **Use async operations** for non-critical data
- **Cache template and keyboard results** when appropriate
- **Monitor service performance** and set appropriate timeouts

### 5. Security
- **Validate all external data** before use in templates
- **Escape user input** in templates
- **Use parameterized queries** for database operations
- **Implement rate limiting** for external API calls

This comprehensive pattern library provides developers with proven approaches for integrating external services while maintaining performance, reliability, and user experience in TeleFlow applications.