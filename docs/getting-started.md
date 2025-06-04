# Getting Started with Teleflow

Welcome to Teleflow! This guide will help you get your first Telegram bot up and running quickly using our modern, type-safe Go framework. Teleflow is designed to make building interactive and sophisticated bots simple and enjoyable.

## Table of Contents

- [Installation](#installation)
- [Your First Bot: Hello Teleflow](#your-first-bot-hello-teleflow)
- [Core Concepts Overview](#core-concepts-overview)
  - [The Bot Engine](#the-bot-engine)
  - [Handlers: Responding to Users](#handlers-responding-to-users)
  - [Context: The Heart of Interaction](#context-the-heart-of-interaction)
  - [Keyboards: Interactive Buttons](#keyboards-interactive-buttons)
  - [Middleware: Processing Pipeline](#middleware-processing-pipeline)
  - [Flows: Managing Conversations](#flows-managing-conversations)
  - [Menu Button: Persistent Commands](#menu-button-persistent-commands)
  - [Templates: Dynamic Messages](#templates-dynamic-messages)
- [Running the Example](#running-the-example)
- [Next Steps](#next-steps)

## Installation

### Prerequisites

- Go 1.19 or later.
- A Telegram Bot Token obtained from [@BotFather](https://t.me/botfather) on Telegram.

### Get Teleflow

In your Go project directory:
```bash
go mod init your-bot-module-name # e.g., my-teleflow-bot
go get github.com/kslamph/teleflow/core
```

## Your First Bot: Hello Teleflow

Let's create a simple bot that responds to a `/start` command.

Create a `main.go` file:
```go
package main

import (
	"log"
	"os" // For reading the bot token from environment variable

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	// 1. Create a new Bot instance with your token
	// It's good practice to store your token in an environment variable
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable not set")
	}

	bot, err := teleflow.NewBot(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// 2. Register a handler for the /start command
	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		// Use the Context to reply to the user
		return ctx.Reply("🎉 Hello from Teleflow! Your bot is up and running.")
	})

	// 3. Start the bot
	log.Println("Bot starting...")
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
```

## Core Concepts Overview

Teleflow is built around a few key concepts:

### The Bot Engine

The `teleflow.Bot` ([core/bot.go](core/bot.go)) struct is the central piece of your application. It manages the connection to Telegram, processes incoming updates, and routes them to your handlers.

### Handlers: Responding to Users

Handlers are functions that you define to process different types of user interactions.
- **Command Handlers**: Respond to slash commands (e.g., `/help`).
  ```go
  bot.HandleCommand("help", func(ctx *teleflow.Context) error { /* ... */ })
  ```
- **Text Handlers**: Respond to plain text messages or reply keyboard presses.
  ```go
  bot.HandleText("Show Menu", func(ctx *teleflow.Context) error { /* ... */ }) // Specific text
  bot.HandleText("", func(ctx *teleflow.Context) error { /* ... */ })        // Any other text
  ```
- **Callback Handlers**: Respond to inline keyboard button presses.
  ```go
  bot.RegisterCallback(teleflow.SimpleCallback("action_*", func(ctx *teleflow.Context, data string) error { /* ... */ }))
  ```
For a detailed guide on all handler types, see the [Handlers Guide](handlers-guide.md).

### Context: The Heart of Interaction

The `teleflow.Context` ([core/context.go](core/context.go)) is passed to every handler. It provides:
- The incoming Telegram update (`ctx.Update`).
- User and chat information (`ctx.UserID()`, `ctx.ChatID()`).
- Helper methods to reply (`ctx.Reply()`, `ctx.ReplyTemplate()`), edit messages, manage state, and more.

### Keyboards: Interactive Buttons

Engage users with interactive buttons:
- **Reply Keyboards**: Appear below the text input field, good for main navigation.
  ```go
  kb := teleflow.NewReplyKeyboard().AddButton("Option 1").AddRow()
  ctx.Reply("Choose:", kb)
  ```
- **Inline Keyboards**: Attached directly to messages, ideal for contextual actions.
  ```go
  kb := teleflow.NewInlineKeyboard().AddButton("Confirm", "confirm_data").AddRow()
  ctx.Reply("Proceed?", kb)
  ```
Learn more in the [Keyboards Guide](keyboards-guide.md).

### Middleware: Processing Pipeline

Middleware functions intercept requests before they reach your handlers. Use them for logging, authentication, rate limiting, etc.
```go
bot.Use(teleflow.LoggingMiddleware()) // Built-in logger
bot.Use(func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
    return func(ctx *teleflow.Context) error {
        log.Println("Before handler")
        err := next(ctx)
        log.Println("After handler")
        return err
    }
})
```
Explore further in the [Middleware Guide](middleware-guide.md).

### Flows: Managing Conversations

For multi-step interactions like registrations or surveys, Teleflow provides a powerful Flow system.
```go
registrationFlow := teleflow.NewFlow("register").
    Step("name").WithPrompt("What's your name?").NextStep("email").
    Step("email").WithPrompt("What's your email?").
    OnComplete(func(ctx *teleflow.Context) error {
        // Access flow data from ctx.Get("name"), ctx.Get("email")
        return ctx.Reply("Registration complete!")
    }).Build()
bot.RegisterFlow(registrationFlow)
// To start: ctx.StartFlow("register")
```
Dive deep into conversational logic with the [Flow Guide](flow-guide.md).

### Menu Button: Persistent Commands

Configure Telegram's native menu button (often seen as a "/" button in the chat input) for quick access to commands or a web app.
```go
menu := teleflow.NewCommandsMenuButton().
    AddCommand("Start Bot", "/start").
    AddCommand("Get Help", "/help")
bot.WithMenuButton(menu) // Set during bot initialization or later
```
Details can be found in the [Menu Button Guide](menu-button-guide.md).

### Templates: Dynamic Messages

Use Go's `text/template` engine to create dynamic and personalized messages.
```go
bot.MustAddTemplate("greeting", "Hello {{.UserName}}! You have {{.MessageCount}} new messages.", teleflow.ParseModeNone)
// In handler:
data := map[string]interface{}{"UserName": "Alex", "MessageCount": 5}
ctx.ReplyTemplate("greeting", data)
```
Master dynamic content with the [Templates Guide](templates-guide.md).

## Running the Example

1.  Set your bot token as an environment variable:
    ```bash
    export BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"
    ```
2.  Run your `main.go` file:
    ```bash
    go run main.go
    ```
3.  Open Telegram and send `/start` to your bot!

## Next Steps

You've taken your first step with Teleflow! To build more advanced bots, explore these resources:

- **Detailed Feature Guides**:
  - [Handlers Guide](handlers-guide.md)
  - [Keyboards Guide](keyboards-guide.md)
  - [Middleware Guide](middleware-guide.md)
  - [Flow Guide](flow-guide.md) (Conversational Flows)
  - [Menu Button Guide](menu-button-guide.md)
  - [Templates Guide](templates-guide.md) (Dynamic Messages)
  - [API Reference](api-reference.md) (Full package documentation)
- **Examples Directory**: Check out the `examples/` directory in the Teleflow repository for more complete bot implementations demonstrating various features.
- **Teleflow Core Files**: For the deepest understanding, you can explore the source code in the `core/` directory.

Happy bot building!