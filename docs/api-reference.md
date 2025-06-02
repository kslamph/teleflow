# Teleflow API Reference

Complete reference for all public types, methods, and functions in the Teleflow framework.

## Table of Contents

- [Core Types](#core-types)
- [Bot Management](#bot-management)
- [Context and Handlers](#context-and-handlers)
- [Keyboards](#keyboards)
- [Flows](#flows)
- [Middleware](#middleware)
- [State Management](#state-management)
- [Callbacks](#callbacks)
- [Templates](#templates)

## Core Types

### Bot

The main application structure that manages your Telegram bot.

```go
type Bot struct {
    // Internal fields - access via methods
}
```

### HandlerFunc

Function signature for all interaction handlers.

```go
type HandlerFunc func(ctx *Context) error
```

**Example:**
```go
func myHandler(ctx *teleflow.Context) error {
    return ctx.Reply("Hello from handler!")
}
```

### MiddlewareFunc

Function signature for middleware that wraps handlers.

```go
type MiddlewareFunc func(next HandlerFunc) HandlerFunc
```

**Example:**
```go
func MyMiddleware() MiddlewareFunc {
    return func(next HandlerFunc) HandlerFunc {
        return func(ctx *Context) error {
            // Pre-processing
            err := next(ctx)
            // Post-processing
            return err
        }
    }
}
```

### BotOption

Functional option for configuring Bot instances.

```go
type BotOption func(*Bot)
```

## Bot Management

### NewBot

Creates a new Bot instance with optional configuration.

```go
func NewBot(token string, options ...BotOption) (*Bot, error)
```

**Parameters:**
- `token` - Your Telegram bot token from [@BotFather](https://t.me/botfather)
- `options` - Variable number of configuration options

**Returns:**
- `*Bot` - Configured bot instance
- `error` - Error if bot creation fails

**Example:**
```go
bot, err := teleflow.NewBot("your-token-here")
if err != nil {
    log.Fatal(err)
}

// With options
bot, err := teleflow.NewBot(token,
    teleflow.WithFlowConfig(teleflow.FlowConfig{
        ExitCommands: []string{"/cancel"},
        ExitMessage: "Cancelled!",
    }),
)
```

### Bot.Start

Begins listening for updates and processing messages.

```go
func (b *Bot) Start() error
```

**Returns:**
- `error` - Error if bot fails to start

**Example:**
```go
log.Println("Starting bot...")
if err := bot.Start(); err != nil {
    log.Fatal("Failed to start:", err)
}
```

### Bot.HandleCommand

Registers a handler for a specific slash command.

```go
func (b *Bot) HandleCommand(command string, handler HandlerFunc, permissions ...string)
```

**Parameters:**
- `command` - Command name without the `/` prefix
- `handler` - Function to handle the command
- `permissions` - Optional permission strings for authorization

**Example:**
```go
bot.HandleCommand("start", func(ctx *teleflow.Context) error {
    return ctx.Reply("Welcome!")
})

// With permissions
bot.HandleCommand("admin", adminHandler, "admin_access")
```

### Bot.HandleText

Registers a handler for text messages and keyboard button presses.

```go
func (b *Bot) HandleText(handler HandlerFunc, permissions ...string)
```

**Parameters:**
- `handler` - Function to handle text input
- `permissions` - Optional permission strings

**Example:**
```go
bot.HandleText(func(ctx *teleflow.Context) error {
    text := ctx.Update.Message.Text
    return ctx.Reply("You wrote: " + text)
})
```

### Bot.Use

Adds middleware to the bot's middleware chain.

```go
func (b *Bot) Use(middleware MiddlewareFunc)
```

**Parameters:**
- `middleware` - Middleware function to add

**Example:**
```go
bot.Use(teleflow.LoggingMiddleware())
bot.Use(teleflow.RateLimitMiddleware(10))
```

### Bot.RegisterCallback

Registers a type-safe callback handler for inline keyboards.

```go
func (b *Bot) RegisterCallback(handler CallbackHandler)
```

**Parameters:**
- `handler` - Callback handler implementing the CallbackHandler interface

### Bot.RegisterFlow

Registers a flow for multi-step conversations.

```go
func (b *Bot) RegisterFlow(flow *Flow)
```

**Parameters:**
- `flow` - Flow instance to register

**Example:**
```go
flow := teleflow.NewFlow("registration").
    Step("name").OnInput(handleNameInput).
    Step("email").OnInput(handleEmailInput).
    Build()

bot.RegisterFlow(flow)
```

## Bot Options

### WithFlowConfig

Sets the flow configuration for the bot.

```go
func WithFlowConfig(config FlowConfig) BotOption
```

**Example:**
```go
bot, err := teleflow.NewBot(token,
    teleflow.WithFlowConfig(teleflow.FlowConfig{
        ExitCommands:        []string{"/cancel", "/exit"},
        ExitMessage:         "Operation cancelled.",
        AllowGlobalCommands: true,
        HelpCommands:        []string{"/help"},
    }),
)
```

### WithUserPermissions

Sets a user permission checker for authorization.

```go
func WithUserPermissions(checker UserPermissionChecker) BotOption
```

**Example:**
```go
type MyPermissionChecker struct{}

func (c *MyPermissionChecker) CanExecute(userID int64, action string) bool {
    // Your permission logic here
    return true
}

func (c *MyPermissionChecker) GetMainMenuForUser(userID int64) *ReplyKeyboard {
    // Return user-specific menu
    return nil
}

bot, err := teleflow.NewBot(token,
    teleflow.WithUserPermissions(&MyPermissionChecker{}),
)
```

## Context and Handlers

### Context

Provides information and helpers for the current interaction.

```go
type Context struct {
    Bot    *Bot
    Update tgbotapi.Update
    // Private fields
}
```

### Context.UserID

Returns the ID of the user who initiated the update.

```go
func (c *Context) UserID() int64
```

**Example:**
```go
func myHandler(ctx *teleflow.Context) error {
    userID := ctx.UserID()
    log.Printf("Processing request from user %d", userID)
    return ctx.Reply("Hello!")
}
```

### Context.ChatID

Returns the ID of the chat where the update originated.

```go
func (c *Context) ChatID() int64
```

### Context.Set

Stores a value in the context's data map.

```go
func (c *Context) Set(key string, value interface{})
```

**Example:**
```go
ctx.Set("user_name", "John")
ctx.Set("step", 1)
```

### Context.Get

Retrieves a value from the context's data map.

```go
func (c *Context) Get(key string) (interface{}, bool)
```

**Returns:**
- `interface{}` - The stored value
- `bool` - Whether the key exists

**Example:**
```go
if name, ok := ctx.Get("user_name"); ok {
    userName := name.(string)
    // Use userName
}
```

### Context.Reply

Sends a text message with optional keyboard.

```go
func (c *Context) Reply(text string, keyboard ...interface{}) error
```

**Parameters:**
- `text` - Message text to send
- `keyboard` - Optional keyboard (ReplyKeyboard, InlineKeyboard, or telegram-bot-api types)

**Example:**
```go
// Simple reply
ctx.Reply("Hello!")

// With inline keyboard
keyboard := teleflow.NewInlineKeyboard()
keyboard.AddButton("Yes", "confirm_yes").AddButton("No", "confirm_no")
ctx.Reply("Confirm action?", keyboard)

// With reply keyboard
keyboard := teleflow.NewReplyKeyboard()
keyboard.AddButton("Option 1").AddButton("Option 2")
ctx.Reply("Choose option:", keyboard)
```

### Context.ReplyTemplate

Sends a message using a template with data.

```go
func (c *Context) ReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error
```

**Parameters:**
- `templateName` - Name of the registered template
- `data` - Data to pass to the template
- `keyboard` - Optional keyboard

**Example:**
```go
// First register template (see Templates section)
data := struct {
    Name string
    Count int
}{
    Name: "John",
    Count: 5,
}
ctx.ReplyTemplate("welcome", data)
```

### Context.EditOrReply

Attempts to edit the current message, falls back to sending a new one.

```go
func (c *Context) EditOrReply(text string, keyboard ...interface{}) error
```

**Example:**
```go
// Will edit if called from callback query, otherwise send new message
ctx.EditOrReply("Updated message", updatedKeyboard)
```

### Context.StartFlow

Initiates a flow for the current user.

```go
func (c *Context) StartFlow(flowName string) error
```

**Parameters:**
- `flowName` - Name of the registered flow to start

**Example:**
```go
func startRegistration(ctx *teleflow.Context) error {
    ctx.Set("start_time", time.Now())
    return ctx.StartFlow("user_registration")
}
```

## Keyboards

### ReplyKeyboard

Persistent keyboard shown below the message input.

```go
type ReplyKeyboard struct {
    Keyboard              [][]ReplyKeyboardButton
    ResizeKeyboard        bool
    OneTimeKeyboard       bool
    InputFieldPlaceholder string
    Selective             bool
}
```

### NewReplyKeyboard

Creates a new reply keyboard.

```go
func NewReplyKeyboard(rows ...[]ReplyKeyboardButton) *ReplyKeyboard
```

**Example:**
```go
keyboard := teleflow.NewReplyKeyboard()
```

### ReplyKeyboard.AddButton

Adds a button to the current row.

```go
func (kb *ReplyKeyboard) AddButton(text string) *ReplyKeyboard
```

**Example:**
```go
keyboard.AddButton("🏠 Home").AddButton("ℹ️ Info")
```

### ReplyKeyboard.AddRow

Starts a new row of buttons.

```go
func (kb *ReplyKeyboard) AddRow() *ReplyKeyboard
```

**Example:**
```go
keyboard.AddButton("Button 1").AddButton("Button 2").AddRow()
keyboard.AddButton("Button 3").AddRow()
```

### ReplyKeyboard.AddRequestContact

Adds a contact request button.

```go
func (kb *ReplyKeyboard) AddRequestContact() *ReplyKeyboard
```

### ReplyKeyboard.AddRequestLocation

Adds a location request button.

```go
func (kb *ReplyKeyboard) AddRequestLocation() *ReplyKeyboard
```

### ReplyKeyboard.Resize

Makes the keyboard compact.

```go
func (kb *ReplyKeyboard) Resize() *ReplyKeyboard
```

### ReplyKeyboard.OneTime

Makes the keyboard hide after one use.

```go
func (kb *ReplyKeyboard) OneTime() *ReplyKeyboard
```

### ReplyKeyboard.Placeholder

Sets placeholder text for the input field.

```go
func (kb *ReplyKeyboard) Placeholder(text string) *ReplyKeyboard
```

**Complete Example:**
```go
keyboard := teleflow.NewReplyKeyboard()
keyboard.AddButton("🏠 Home").AddButton("ℹ️ Info").AddRow()
keyboard.AddButton("📞 Contact").AddRequestContact().AddRow()
keyboard.AddRequestLocation().AddRow()
keyboard.Resize().OneTime().Placeholder("Choose an option...")

return ctx.Reply("Main menu:", keyboard)
```

### InlineKeyboard

Buttons attached to specific messages.

```go
type InlineKeyboard struct {
    InlineKeyboard [][]InlineKeyboardButton
}
```

### NewInlineKeyboard

Creates a new inline keyboard.

```go
func NewInlineKeyboard(rows ...[]InlineKeyboardButton) *InlineKeyboard
```

### InlineKeyboard.AddButton

Adds a callback button to the current row.

```go
func (kb *InlineKeyboard) AddButton(text, data string) *InlineKeyboard
```

**Parameters:**
- `text` - Button label
- `data` - Callback data sent when pressed

### InlineKeyboard.AddURL

Adds a URL button to the current row.

```go
func (kb *InlineKeyboard) AddURL(text, url string) *InlineKeyboard
```

### InlineKeyboard.AddWebApp

Adds a web app button to the current row.

```go
func (kb *InlineKeyboard) AddWebApp(text string, webApp WebAppInfo) *InlineKeyboard
```

### InlineKeyboard.AddRow

Starts a new row of buttons.

```go
func (kb *InlineKeyboard) AddRow() *InlineKeyboard
```

**Complete Example:**
```go
keyboard := teleflow.NewInlineKeyboard()
keyboard.AddButton("✅ Confirm", "confirm_action").AddButton("❌ Cancel", "cancel_action").AddRow()
keyboard.AddURL("📖 Documentation", "https://github.com/kslamph/teleflow").AddRow()
keyboard.AddButton("ℹ️ More Info", "show_info").AddRow()

return ctx.Reply("Choose an action:", keyboard)
```

## Flows

Multi-step conversation management system.

### Flow

Represents a structured multi-step conversation.

```go
type Flow struct {
    Name        string
    Steps       []*FlowStep
    OnComplete  HandlerFunc
    OnCancel    HandlerFunc
    Timeout     time.Duration
}
```

### NewFlow

Creates a new flow builder.

```go
func NewFlow(name string) *FlowBuilder
```

**Example:**
```go
flow := teleflow.NewFlow("user_registration")
```

### FlowBuilder.Step

Adds a step to the flow.

```go
func (fb *FlowBuilder) Step(name string) *FlowStepBuilder
```

### FlowStepBuilder.OnInput

Sets the input handler for the step.

```go
func (fsb *FlowStepBuilder) OnInput(handler HandlerFunc) *FlowStepBuilder
```

### FlowStepBuilder.WithValidator

Adds input validation to the step.

```go
func (fsb *FlowStepBuilder) WithValidator(validator FlowValidatorFunc) *FlowStepBuilder
```

### FlowStepBuilder.NextStep

Sets the next step name.

```go
func (fsb *FlowStepBuilder) NextStep(stepName string) *FlowStepBuilder
```

### FlowStepBuilder.WithTimeout

Sets step timeout.

```go
func (fsb *FlowStepBuilder) WithTimeout(timeout time.Duration) *FlowStepBuilder
```

### FlowBuilder.OnComplete

Sets the completion handler.

```go
func (fb *FlowBuilder) OnComplete(handler HandlerFunc) *FlowBuilder
```

### FlowBuilder.Build

Creates the final flow.

```go
func (fb *FlowBuilder) Build() *Flow
```

**Complete Flow Example:**
```go
flow := teleflow.NewFlow("registration").
    Step("ask_name").
    OnInput(func(ctx *teleflow.Context) error {
        name := ctx.Update.Message.Text
        ctx.Set("name", name)
        return ctx.Reply("Great! Now please enter your email:")
    }).
    Step("ask_email").
    OnInput(func(ctx *teleflow.Context) error {
        email := ctx.Update.Message.Text
        ctx.Set("email", email)
        return ctx.Reply("Perfect! Registration will be completed.")
    }).
    WithValidator(EmailValidator()).
    OnComplete(func(ctx *teleflow.Context) error {
        name, _ := ctx.Get("name")
        email, _ := ctx.Get("email")
        return ctx.Reply(fmt.Sprintf("Welcome %s! Your email %s has been registered.", name, email))
    }).
    Build()

bot.RegisterFlow(flow)
```

### FlowStepType

Enumeration of step types.

```go
const (
    StepTypeText FlowStepType = iota
    StepTypeChoice
    StepTypeConfirmation
    StepTypeCustom
)
```

### Flow Validators

Pre-built validation functions.

#### NumberValidator

Validates numeric input.

```go
func NumberValidator() FlowValidatorFunc
```

**Example:**
```go
flow.Step("ask_age").
    WithValidator(teleflow.NumberValidator()).
    OnInput(handleAgeInput)
```

#### ChoiceValidator

Validates input against allowed choices.

```go
func ChoiceValidator(choices []string) FlowValidatorFunc
```

**Example:**
```go
choices := []string{"Option A", "Option B", "Option C"}
flow.Step("ask_choice").
    WithValidator(teleflow.ChoiceValidator(choices)).
    OnInput(handleChoiceInput)
```

## Middleware

### LoggingMiddleware

Logs all interactions and execution times.

```go
func LoggingMiddleware() MiddlewareFunc
```

**Example:**
```go
bot.Use(teleflow.LoggingMiddleware())
// Output: [123456] Processing command: start
//         [123456] Handler completed in 2.5ms
```

### AuthMiddleware

Checks user authorization.

```go
func AuthMiddleware(checker UserPermissionChecker) MiddlewareFunc
```

**Example:**
```go
bot.Use(teleflow.AuthMiddleware(myPermissionChecker))
```

### RateLimitMiddleware

Implements rate limiting per user.

```go
func RateLimitMiddleware(requestsPerMinute int) MiddlewareFunc
```

**Example:**
```go
// Allow 10 requests per minute per user
bot.Use(teleflow.RateLimitMiddleware(10))
```

### RecoveryMiddleware

Recovers from panics and logs them.

```go
func RecoveryMiddleware() MiddlewareFunc
```

**Example:**
```go
bot.Use(teleflow.RecoveryMiddleware())
```

### Custom Middleware

Create your own middleware:

```go
func TimingMiddleware() teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            start := time.Now()
            err := next(ctx)
            duration := time.Since(start)
            
            if duration > 1*time.Second {
                log.Printf("Slow handler: %v", duration)
            }
            
            return err
        }
    }
}

bot.Use(TimingMiddleware())
```

## State Management

### StateManager

Interface for managing user state.

```go
type StateManager interface {
    SetState(userID int64, state interface{}) error
    GetState(userID int64) (interface{}, error)
    DeleteState(userID int64) error
}
```

### NewInMemoryStateManager

Creates an in-memory state manager.

```go
func NewInMemoryStateManager() StateManager
```

**Example:**
```go
stateManager := teleflow.NewInMemoryStateManager()
// Used internally by Bot and FlowManager
```

## Callbacks

### CallbackHandler

Interface for handling inline keyboard callbacks.

```go
type CallbackHandler interface {
    CallbackData() string
    Handle(ctx *Context, data string) error
}
```

**Example Implementation:**
```go
type ConfirmHandler struct{}

func (h *ConfirmHandler) CallbackData() string {
    return "confirm_"
}

func (h *ConfirmHandler) Handle(ctx *teleflow.Context, data string) error {
    switch data {
    case "confirm_yes":
        return ctx.EditOrReply("✅ Confirmed!")
    case "confirm_no":
        return ctx.EditOrReply("❌ Cancelled!")
    default:
        return ctx.EditOrReply("Unknown option")
    }
}

// Register the handler
bot.RegisterCallback(&ConfirmHandler{})
```

## Templates

### Template Registration

Register templates for reusable messages.

```go
// Templates are registered on the bot instance
// Access via bot.templates (internal use)
```

**Example:**
```go
// In your bot initialization
templateText := `Hello {{.Name}}! You have {{.Count}} new messages.`
bot.templates.New("welcome").Parse(templateText)

// Usage in handler
data := struct {
    Name  string
    Count int
}{
    Name:  "John",
    Count: 5,
}
ctx.ReplyTemplate("welcome", data)
```

## Error Handling

### Best Practices

Always handle errors in your handlers:

```go
bot.HandleCommand("example", func(ctx *teleflow.Context) error {
    // Your logic here
    if err := someOperation(); err != nil {
        log.Printf("Operation failed: %v", err)
        return ctx.Reply("❌ Something went wrong. Please try again.")
    }
    
    return ctx.Reply("✅ Operation completed successfully!")
})
```

### Common Error Patterns

```go
// Network errors
if err := ctx.Reply("Message"); err != nil {
    log.Printf("Failed to send message: %v", err)
    // Don't return the error to avoid user-facing error messages
    return nil
}

// Validation errors
if input == "" {
    return ctx.Reply("❌ Input cannot be empty")
}

// Flow errors
if err := ctx.StartFlow("nonexistent"); err != nil {
    log.Printf("Flow error: %v", err)
    return ctx.Reply("❌ Service temporarily unavailable")
}
```

## Configuration Types

### FlowConfig

Configuration for flow behavior.

```go
type FlowConfig struct {
    ExitCommands        []string // Commands that exit flows (e.g., "/cancel")
    ExitMessage         string   // Message shown when flow is cancelled
    AllowGlobalCommands bool     // Allow global commands during flows
    HelpCommands        []string // Help commands allowed during flows
}
```

### UserPermissionChecker

Interface for user authorization.

```go
type UserPermissionChecker interface {
    CanExecute(userID int64, action string) bool
    GetMainMenuForUser(userID int64) *ReplyKeyboard
}
```

This completes the comprehensive API reference. Each section provides detailed information about the types, methods, and usage patterns available in the Teleflow framework.