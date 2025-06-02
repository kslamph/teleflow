# Getting Started with Teleflow

A modern, type-safe Telegram bot framework for Go that makes building interactive bots simple and enjoyable.

## Table of Contents

- [Installation](#installation)
- [Hello World Bot](#hello-world-bot)
- [Basic Concepts](#basic-concepts)
- [Examples](#examples)
- [Next Steps](#next-steps)

## Installation

### Prerequisites

- Go 1.19 or later
- A Telegram Bot Token from [@BotFather](https://t.me/botfather)

### Install Teleflow

```bash
go mod init your-bot-name
go get github.com/kslamph/teleflow/core
```

### Get Your Bot Token

1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Create a new bot with `/newbot`
3. Follow the instructions to get your bot token
4. Save the token - you'll need it for your bot

## Hello World Bot

Create a simple bot in just 10 lines of code:

```go
package main

import (
    "log"
    "os"
    teleflow "github.com/kslamph/teleflow/core"
)

func main() {
    // Create bot with your token
    bot, err := teleflow.NewBot(os.Getenv("BOT_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    // Handle /start command
    bot.HandleCommand("start", func(ctx *teleflow.Context) error {
        return ctx.Reply("🎉 Hello World! Your Teleflow bot is working!")
    })

    // Start the bot
    log.Println("Bot starting...")
    bot.Start()
}
```

### Running Your Bot

```bash
export BOT_TOKEN="your-bot-token-here"
go run main.go
```

That's it! Your bot will respond to `/start` with a greeting message.

## Basic Concepts

### Handlers

Handlers are functions that process user interactions. Teleflow provides several types:

#### Command Handlers
Handle slash commands like `/start`, `/help`:

```go
bot.HandleCommand("help", func(ctx *teleflow.Context) error {
    return ctx.Reply("Need help? Type /start to begin!")
})
```

#### Text Handlers
Handle text messages and keyboard button presses:

```go
bot.HandleText(func(ctx *teleflow.Context) error {
    text := ctx.Update.Message.Text
    return ctx.Reply("You wrote: " + text)
})
```

#### Callback Handlers
Handle inline keyboard button presses:

```go
bot.RegisterCallback(&MyCallbackHandler{})
```

### Context

The [`Context`](../core/context.go:11) provides access to the current update and helper methods:

```go
func myHandler(ctx *teleflow.Context) error {
    userID := ctx.UserID()           // Get user ID
    chatID := ctx.ChatID()           // Get chat ID
    ctx.Set("key", "value")          // Store data
    value, ok := ctx.Get("key")      // Retrieve data
    return ctx.Reply("Response")     // Send reply
}
```

### Keyboards

Create interactive keyboards for better user experience:

#### Reply Keyboards
Persistent buttons shown below the message input:

```go
keyboard := teleflow.NewReplyKeyboard()
keyboard.AddButton("🏠 Home").AddButton("ℹ️ Info").AddRow()
keyboard.AddButton("❓ Help").AddRow()
keyboard.Resize() // Make keyboard compact

return ctx.Reply("Choose an option:", keyboard)
```

#### Inline Keyboards
Buttons attached to specific messages:

```go
keyboard := teleflow.NewInlineKeyboard()
keyboard.AddButton("Yes", "confirm_yes").AddButton("No", "confirm_no").AddRow()
keyboard.AddURL("Visit Website", "https://example.com").AddRow()

return ctx.Reply("Please confirm:", keyboard)
```

### Middleware

Add cross-cutting functionality like logging, authentication, or rate limiting:

```go
// Add logging to track all interactions
bot.Use(teleflow.LoggingMiddleware())

// Add rate limiting (10 requests per minute)
bot.Use(teleflow.RateLimitMiddleware(10))

// Add custom middleware
bot.Use(func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
    return func(ctx *teleflow.Context) error {
        // Do something before handler
        err := next(ctx)
        // Do something after handler
        return err
    }
})
```

### Bot Configuration

Configure your bot with functional options:

```go
bot, err := teleflow.NewBot(token,
    teleflow.WithFlowConfig(teleflow.FlowConfig{
        ExitCommands: []string{"/cancel", "/exit"},
        ExitMessage:  "Operation cancelled.",
    }),
)
```

## Examples

### Complete Bot with Keyboard

```go
package main

import (
    "log"
    "os"
    teleflow "github.com/kslamph/teleflow/core"
)

func main() {
    bot, err := teleflow.NewBot(os.Getenv("BOT_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    // Add middleware
    bot.Use(teleflow.LoggingMiddleware())

    // Welcome command with keyboard
    bot.HandleCommand("start", func(ctx *teleflow.Context) error {
        keyboard := teleflow.NewReplyKeyboard()
        keyboard.AddButton("🎯 Features").AddButton("📚 Help").AddRow()
        keyboard.AddButton("ℹ️ About").AddRow()
        keyboard.Resize()

        return ctx.Reply("Welcome! Choose an option:", keyboard)
    })

    // Handle keyboard buttons
    bot.HandleText(func(ctx *teleflow.Context) error {
        text := ctx.Update.Message.Text
        
        switch text {
        case "🎯 Features":
            return ctx.Reply("✨ Bot features:\n• Interactive keyboards\n• Command handling\n• Middleware support")
        case "📚 Help":
            return ctx.Reply("💡 Type /start to see the main menu")
        case "ℹ️ About":
            return ctx.Reply("🤖 Built with Teleflow framework")
        default:
            return ctx.Reply("Use the keyboard buttons or type /start")
        }
    })

    log.Println("Bot starting...")
    bot.Start()
}
```

### Using Inline Keyboards

```go
bot.HandleCommand("menu", func(ctx *teleflow.Context) error {
    keyboard := teleflow.NewInlineKeyboard()
    keyboard.AddButton("Option 1", "option_1").AddButton("Option 2", "option_2").AddRow()
    keyboard.AddURL("Documentation", "https://github.com/kslamph/teleflow").AddRow()
    
    return ctx.Reply("Choose an option:", keyboard)
})

// Handle inline button presses (requires callback handler implementation)
```

## Next Steps

Now that you understand the basics, explore these advanced features:

1. **[API Reference](api-reference.md)** - Complete documentation of all types and methods
2. **[Flow Guide](flow-guide.md)** - Learn about multi-step conversations and flow management
3. **[Examples](../examples/)** - Check out complete example bots:
   - [`basic-bot`](../examples/basic-bot/) - Simple bot with keyboards
   - [`flow-bot`](../examples/flow-bot/) - Multi-step conversation flows
   - [`middleware-bot`](../examples/middleware-bot/) - Advanced middleware usage

### Common Patterns

- **Environment Configuration**: Use environment variables for tokens and settings
- **Error Handling**: Always handle errors in your handlers
- **Logging**: Use [`LoggingMiddleware()`](../core/middleware.go:10) to track interactions
- **User Experience**: Provide keyboards and clear instructions
- **State Management**: Use flows for multi-step interactions

### Troubleshooting

**Bot not responding?**
- Check your bot token is correct
- Ensure the bot is started with `bot.Start()`
- Check logs for error messages

**Commands not working?**
- Commands must start with `/`
- Register handlers before calling `bot.Start()`
- Use [`HandleCommand()`](../core/bot.go:119) for slash commands

**Keyboards not showing?**
- Make sure to pass the keyboard as second parameter to `ctx.Reply()`
- Call `.AddRow()` to finalize button rows
- Use `.Resize()` for compact keyboards

Ready to build something amazing? Start with the examples and refer to the [API Reference](api-reference.md) for detailed documentation of all features.