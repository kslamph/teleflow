# Teleflow - Simple Telegram Bot Framework

[![Go Reference](https://pkg.go.dev/badge/github.com/kslamph/teleflow.svg)](https://pkg.go.dev/github.com/kslamph/teleflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/kslamph/teleflow)](https://goreportcard.com/report/github.com/kslamph/teleflow)

> **Quick Start for Go Developers**: Simple, type-safe Telegram bot framework with fluent flows and automatic state management.

## 🚀 Quick Overview

Teleflow is a Telegram bot framework designed for simplicity and ease of use:

- **Type-Safe Flows**: Use Go generics (`teleflow.Flow[T]`) for compile-time safety
- **Automatic State Management**: No manual session handling needed
- **Non-Linear Conversations**: Named steps and `GoTo` actions for complex flows
- **Middleware Support**: Standard `Handler` interface for authentication, logging, etc.
- **Large Callback Data**: Built-in mechanism to pass complex data through buttons
- **Clean API**: Simple, consistent interface for all common bot operations

## 📚 Documentation

For comprehensive guides and architectural details, see our documentation:

- **[Architecture Guide](docs/teleflow_architecture.md)** - Detailed system architecture, component interactions, and design patterns
- **[LLM Developer Guide](docs/llm_teleflow_guide.md)** - Comprehensive guide for understanding core concepts and usage patterns
- **[GoDoc API Reference](docs/godoc_summary.md)** - Complete API documentation summary with examples

## 🏗️ Basic Bot Setup

### Minimal Bot
```go
import teleflow "github.com/kslamph/teleflow/core"

// Set up router with handlers
router := teleflow.NewRouter()
router.Handle("start", teleflow.Func(func(ctx *teleflow.Context) error {
    return ctx.ReplyText("Welcome to Teleflow!")
}))

// Create bot with token and router
bot, err := teleflow.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"), router)
if err != nil {
    log.Fatal(err)
}

bot.Start()
```

### Simple Flow Bot
```go
// 1. Define a struct for your flow's data.
type UserProfile struct {
    Name string
    Age  int
}

// 2. Create a new Flow with your data type.
profileFlow := teleflow.NewFlow[UserProfile]()

// 3. Define the steps of your conversation.
nameStep := teleflow.Step[UserProfile]{
    Name: "Name",
    Prompt: func(ctx *teleflow.Context, data *UserProfile) error {
        return ctx.ReplyText("What is your name?")
    },
    Process: func(ctx *teleflow.Context, data *UserProfile) (teleflow.FlowAction, error) {
        data.Name = ctx.Text()
        return teleflow.Next(), nil
    },
}

ageStep := teleflow.Step[UserProfile]{
    Name: "Age",
    Prompt: func(ctx *teleflow.Context, data *UserProfile) error {
        return ctx.ReplyText(fmt.Sprintf("Nice to meet you, %s! How old are you?", data.Name))
    },
    Process: func(ctx *teleflow.Context, data *UserProfile) (teleflow.FlowAction, error) {
        age, err := strconv.Atoi(ctx.Text())
        if err != nil {
            ctx.ReplyText("Please enter a valid number.")
            return teleflow.Retry(), nil
        }
        data.Age = age
        return teleflow.End(), nil // End the flow.
    },
}

// 4. Add steps to the flow and set a completion handler.
profileFlow.AddStep(nameStep).AddStep(ageStep)
profileFlow.OnComplete(func(ctx *teleflow.Context, data UserProfile) error {
    return ctx.ReplyText(fmt.Sprintf("Thanks, %s! Your profile is saved.", data.Name))
})

// 5. Set up the router and bot.
router := teleflow.NewRouter()
router.Handle("profile", profileFlow)

bot, err := teleflow.NewBot(token, router)
if err != nil {
    log.Fatal(err)
}

bot.Start()
```

## 📡 Advanced Messaging

### Simple Messaging
```go
// Simple text only
ctx.ReplyText("Hello World")
```

### Rich Messaging with Keyboards
```go
// Rich message with inline keyboard
keyboard := teleflow.NewInlineKeyboard().AddRow(
    teleflow.ButtonCallback("Option 1", "opt1"),
    teleflow.ButtonCallback("Option 2", "opt2"),
).Build()

ctx.ReplyWithKeyboard("Choose an option:", keyboard)
```

### Formatted Messaging with Parse Modes

Teleflow supports sending formatted messages using HTML, Markdown, or MarkdownV2 parse modes:

```go
// HTML formatting
ctx.ReplyTextWithParseMode("<b>Bold text</b> and <i>italic text</i>", teleflow.ParseModeHTML)

// MarkdownV2 formatting
ctx.ReplyTextWithParseMode("*Bold text* and _italic text_", teleflow.ParseModeMarkdownV2)

// With keyboard and formatting
keyboard := teleflow.NewInlineKeyboard().AddRow(
    teleflow.ButtonCallback("Option 1", "opt1"),
    teleflow.ButtonCallback("Option 2", "opt2"),
).Build()

ctx.ReplyWithKeyboardAndParseMode("*Choose an option:*", keyboard, teleflow.ParseModeMarkdownV2)

// Editing messages with formatting
ctx.EditMessageTextWithParseMode(messageID, "<u>Updated</u> <s>text</s>", teleflow.ParseModeHTML)
```

## 🌊 Flow Architecture

### Simple Flow with OnComplete
```go
// Define your data structure
type RegistrationData struct {
    Name string
    Age  int
}

// Create a new flow
registrationFlow := teleflow.NewFlow[RegistrationData]()

// Define steps
welcomeStep := teleflow.Step[RegistrationData]{
    Name: "Welcome",
    Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
        return ctx.ReplyText("Welcome! What is your name?")
    },
    Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
        data.Name = ctx.Text()
        return teleflow.Next(), nil
    },
}

ageStep := teleflow.Step[RegistrationData]{
    Name: "Age",
    Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
        return ctx.ReplyText(fmt.Sprintf("Nice to meet you, %s! How old are you?", data.Name))
    },
    Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
        age, err := strconv.Atoi(ctx.Text())
        if err != nil {
            ctx.ReplyText("Please enter a valid number.")
            return teleflow.Retry(), nil
        }
        data.Age = age
        return teleflow.End(), nil
    },
}

// Add steps to the flow
registrationFlow.AddStep(welcomeStep).AddStep(ageStep)

// Set completion handler
registrationFlow.OnComplete(func(ctx *teleflow.Context, data RegistrationData) error {
    return ctx.ReplyText(fmt.Sprintf("Thanks, %s! Registration complete. Age: %d", data.Name, data.Age))
})

// Register with router
router := teleflow.NewRouter()
router.Handle("register", registrationFlow)
```

## 🎛️ Flow Control

Teleflow provides several flow control actions:

- `teleflow.Next()` - Continue to next step
- `teleflow.Retry()` - Re-prompt current step
- `teleflow.End()` - Finish flow, trigger OnComplete
- `teleflow.GoTo("step_name")` - Jump to specific named step

### Non-Linear Flow Example
```go
confirmationStep := teleflow.Step[RegistrationData]{
    Name: "Confirmation",
    Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
        keyboard := teleflow.NewInlineKeyboard().AddRow(
            teleflow.ButtonCallback("✅ Yes, Correct", "confirm"),
            teleflow.ButtonCallback("❌ No, Restart", "restart"),
        ).Build()

        promptText := fmt.Sprintf(
            "Please confirm your details:\nName: %s\nAge: %d",
            data.Name, data.Age,
        )
        return ctx.ReplyWithKeyboard(promptText, keyboard)
    },
    Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
        switch ctx.Text() {
        case "confirm":
            return teleflow.End(), nil
        case "restart":
            // Reset data and go back to the beginning
            *data = RegistrationData{}
            return teleflow.GoTo("Welcome"), nil
        default:
            ctx.ReplyText("Please use the buttons to confirm or restart.")
            return teleflow.Retry(), nil
        }
    },
}
```

## 🔧 Advanced Features

### Complex Inline Data Handling

Teleflow transparently handles complex data structures in callback buttons:

```go
// Define structured callback data
type TransactionData struct {
    Action     string  `json:"action"`
    Amount     float64 `json:"amount"`
    AccountID  string  `json:"account_id"`
}

// In a flow step
step := teleflow.Step[MyData]{
    Name: "Transaction",
    Prompt: func(ctx *teleflow.Context, data *MyData) error {
        // Create complex callback data
        approveData := TransactionData{
            Action:    "approve",
            Amount:    1500.50,
            AccountID: "acc_12345",
        }

        keyboard := teleflow.NewInlineKeyboard().AddRow(
            teleflow.ButtonCallback("✅ Approve $1,500.50", ctx.RegisterData(approveData)),
        ).Build()

        return ctx.ReplyWithKeyboard("Review this transaction:", keyboard)
    },
    Process: func(ctx *teleflow.Context, data *MyData) (teleflow.FlowAction, error) {
        // Retrieve the complex data
        retrievedData, ok := ctx.GetData()
        if !ok {
            return teleflow.Retry(), nil
        }

        // Type assertion to get our structured data
        txData, ok := retrievedData.(TransactionData)
        if !ok {
            return teleflow.Retry(), nil
        }

        // Process based on the data
        switch txData.Action {
        case "approve":
            // Handle approval
            return teleflow.End(), nil
        }

        return teleflow.Retry(), nil
    },
}
```

### Middleware Support

Teleflow supports middleware through the standard `Handler` interface:

```go
// Auth middleware example
func AuthMiddleware(next teleflow.Handler) teleflow.Handler {
    return teleflow.Func(func(ctx *teleflow.Context) error {
        // Check if user is authorized
        if !isUserAuthorized(ctx.UserID()) {
            return ctx.ReplyText("You are not authorized to use this command.")
        }

        // Continue with the next handler
        return next.HandleUpdate(ctx)
    })
}

// Apply middleware when registering handlers
router.Handle("admin", AuthMiddleware(adminFlow))
```

## 📈 Testing Strategies

```go
// Test flow behavior
func TestUserRegistrationFlow(t *testing.T) {
    // Create test flow
    flow := teleflow.NewFlow[TestData]()

    // Add test steps
    flow.AddStep(teleflow.Step[TestData]{
        Name: "TestStep",
        Prompt: func(ctx *teleflow.Context, data *TestData) error {
            return ctx.ReplyText("Test prompt")
        },
        Process: func(ctx *teleflow.Context, data *TestData) (teleflow.FlowAction, error) {
            data.Value = ctx.Text()
            return teleflow.End(), nil
        },
    })

    // Test flow execution
    // ... testing logic
}
```

## 🔥 Quick Reference

### Flow Control
- `teleflow.Next()` - Continue to next step
- `teleflow.Retry()` - Re-prompt current step
- `teleflow.End()` - Finish flow, trigger OnComplete
- `teleflow.GoTo("step_name")` - Jump to specific named step

### Message Actions
- `ctx.ReplyText(msg)` - Simple text message
- `ctx.ReplyWithKeyboard(text, keyboard)` - Message with inline keyboard
- `ctx.ReplyTextWithParseMode(text, parseMode)` - Text message with formatting
- `ctx.ReplyWithKeyboardAndParseMode(text, keyboard, parseMode)` - Message with keyboard and formatting
- `ctx.EditMessageText(messageID, text)` - Edit existing message
- `ctx.EditMessageTextWithParseMode(messageID, text, parseMode)` - Edit existing message with formatting
- `ctx.AnswerCallback()` - Acknowledge callback query

### Data Management
- `ctx.Set(key, value)` - Store request-scoped data
- `ctx.Get(key)` - Retrieve request data
- `ctx.RegisterData(data)` - Register complex data for callbacks
- `ctx.GetData()` - Retrieve complex callback data

This guide provides the foundation for building Telegram bots with Teleflow. The framework handles the complexity while giving you the power to build sophisticated conversational experiences.