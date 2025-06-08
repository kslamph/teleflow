# Getting Started with Teleflow

Welcome to Teleflow! This guide will help you build your first Telegram bot using the `teleflow/core` package. We'll cover installation, basic bot setup, and create both simple command handlers and conversation flows.

## Prerequisites

- Go 1.24 or later
- A Telegram Bot Token (get one from [@BotFather](https://t.me/botfather))

## Installation

To get started with Teleflow, add it to your Go project:

```bash
go mod init your-bot-project
go get github.com/kslamph/teleflow/core
```

## Getting Your Bot Token

1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Send `/newbot` and follow the instructions
3. Choose a name and username for your bot
4. Copy the bot token provided by BotFather

## Your First Bot: Echo Bot

Let's create a simple bot that responds to commands and echoes messages:

```go
package main

import (
	"log"
	"os"

	"github.com/kslamph/teleflow/core"
)

func main() {
	// Get bot token from environment variable
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	// Create a new bot instance
	bot, err := core.NewBot(botToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Handle the /start command
	// Handle the /start command
	bot.HandleCommand("start", func(ctx *core.Context) error {
		return ctx.SendPromptText("👋 Hello! I'm your new Teleflow bot. Send me any message and I'll echo it back!")
	})

	// Handle the /help command
	bot.HandleCommand("help", func(ctx *core.Context) error {
		return ctx.SendPromptText("🤖 Available commands:\n/start - Start the bot\n/help - Show this help\n\nSend me any text and I'll echo it back!")
	})

	// Echo all text messages
	bot.DefaultHandler(func(ctx *core.Context, text string) error {
		return ctx.SendPromptText("You said: " + text)
	})
	// Start the bot
	log.Println("🚀 Bot starting...")
	log.Fatal(bot.Start())
}
```

### Running Your Bot

1. Save the code to `main.go`
2. Set your bot token:
   ```bash
   export TELEGRAM_BOT_TOKEN="your_bot_token_here"
   ```
3. Run the bot:
   ```bash
   go run main.go
   ```

## Main Menu Keyboard Setup

Add a persistent keyboard below the chat input box for quick user actions:

```go
// Create keyboard with action buttons
mainMenuKeyboard := core.BuildReplyKeyboard([]string{"📝 Register", "🏠 Home", "⚙️ Settings", "❓ Help"}, 3).Resize()

// Implement AccessManager for keyboard management
type MyAccessManager struct {
	mainMenu *core.ReplyKeyboard
}

func (m *MyAccessManager) CheckPermission(ctx *core.PermissionContext) error {
	return nil // Allow all actions
}

func (m *MyAccessManager) GetReplyKeyboard(ctx *core.PermissionContext) *core.ReplyKeyboard {
	return m.mainMenu
}

// Create bot with AccessManager
accessManager := &MyAccessManager{mainMenu: mainMenuKeyboard}
bot, err := core.NewBot(token, core.WithAccessManager(accessManager))

// Handle keyboard button presses
bot.HandleText("📝 Register", func(ctx *core.Context, text string) error {
	return ctx.StartFlow("user_registration")
})
```

## Conversation Flows: Registration Bot

Now let's create a more sophisticated bot using Teleflow's powerful Step-Prompt-Process API for conversation flows:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kslamph/teleflow/core"
)

func main() {
	// Get bot token
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	// Create bot with flow configuration
	bot, err := core.NewBot(botToken,
		core.WithFlowConfig(core.FlowConfig{
			ExitCommands:        []string{"/cancel", "/exit"},
			ExitMessage:         "❌ Operation cancelled.",
			AllowGlobalCommands: false,
		}),
	)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Create a user registration flow
	registrationFlow, err := core.NewFlow("user_registration").
		// Handle errors by cancelling the flow
		OnError(core.OnErrorCancel()).
		
		// Step 1: Ask for name
		Step("name").
		Prompt("👋 Welcome! What's your name?").
		Process(func(ctx *core.Context, input string, buttonClick *core.ButtonClick) core.ProcessResult {
			if input == "" {
				return core.Retry().WithPrompt("Please enter your name:")
			}
			
			// Store the name for later use
			ctx.SetFlowData("user_name", input)
			return core.NextStep().WithPrompt("✅ Nice to meet you, " + input + "!")
		}).
		
		// Step 2: Ask for age
		Step("age").
		Prompt(func(ctx *core.Context) string {
			name, _ := ctx.GetFlowData("user_name")
			return fmt.Sprintf("How old are you, %s?", name)
		}).
		Process(func(ctx *core.Context, input string, buttonClick *core.ButtonClick) core.ProcessResult {
			// Simple validation
			if len(input) == 0 || len(input) > 3 {
				return core.Retry().WithPrompt("Please enter a valid age (1-3 digits):")
			}
			
			ctx.SetFlowData("user_age", input)
			return core.NextStep().WithPrompt("✅ Age recorded!")
		}).
		
		// Step 3: Confirmation with inline keyboard
		Step("confirmation").
		Prompt(func(ctx *core.Context) string {
			name, _ := ctx.GetFlowData("user_name")
			age, _ := ctx.GetFlowData("user_age")
			return fmt.Sprintf("Please confirm:\n👤 Name: %s\n🎂 Age: %s\n\nIs this correct?", name, age)
		}).
		WithInlineKeyboard(func(ctx *core.Context) *core.InlineKeyboardBuilder {
			return core.NewInlineKeyboard().
				ButtonCallback("✅ Yes, correct", "confirm").
				ButtonCallback("❌ Start over", "restart")
		}).
		Process(func(ctx *core.Context, input string, buttonClick *core.ButtonClick) core.ProcessResult {
			switch input {
			case "confirm":
				return core.CompleteFlow().WithPrompt("🎉 Registration complete!")
			case "restart":
				return core.GoToStep("name").WithPrompt("🔄 Let's start over...")
			default:
				return core.Retry().WithPrompt("Please click one of the buttons above.")
			}
		}).
		
		// Handle flow completion
		OnComplete(func(ctx *core.Context) error {
			name, _ := ctx.GetFlowData("user_name")
			age, _ := ctx.GetFlowData("user_age")
			
			message := fmt.Sprintf("🎉 Welcome to our service!\n\n👤 Name: %s\n🎂 Age: %s\n\nYou're all set!", name, age)
			return ctx.SendPromptText(message)
		}).
		Build()

	if err != nil {
		log.Fatal("Failed to build flow:", err)
	}

	// Register the flow with the bot
	bot.RegisterFlow(registrationFlow)

	// Command handlers
	// Command handlers
	bot.HandleCommand("start", func(ctx *core.Context) error {
		return ctx.StartFlow("user_registration")
	})

	bot.HandleCommand("register", func(ctx *core.Context) error {
		return ctx.StartFlow("user_registration")
	})

	bot.HandleCommand("help", func(ctx *core.Context) error {
		return ctx.SendPromptText("🤖 Available commands:\n/start - Begin registration\n/register - Start registration flow\n/cancel - Cancel current operation\n/help - Show this help")
	})
	// Set bot commands for the menu
	if err := bot.SetBotCommands(map[string]string{
		"start":    "Begin user registration",
		"register": "Start registration flow", 
		"cancel":   "Cancel current operation",
		"help":     "Show help information",
	}); err != nil {
		log.Printf("Warning: Failed to set bot commands: %v", err)
	}

	// Start the bot
	log.Println("🚀 Registration bot starting...")
	log.Fatal(bot.Start())
}
```

## Key Concepts Explained

### 1. Bot Creation
```go
bot, err := core.NewBot(botToken, options...)
```
The [`NewBot()`](core/bot.go:15) function creates a new bot instance with your token. You can pass additional options like flow configuration.

### 2. Command Handlers
```go
bot.HandleCommand("start", func(ctx *core.Context) error {
    return ctx.SendPromptText("Hello!")
})
```
Command handlers respond to `/command` messages. The [`Context`](core/context.go:12) provides access to user info, chat details, and helper methods.

### 3. Text Handlers
```go
bot.HandleText("keyword", handlerFunc)  // Specific text
bot.DefaultHandler(handlerFunc)         // Any text message
```
Text handlers process regular messages from users.

### 4. Conversation Flows
The Step-Prompt-Process API allows you to create sophisticated conversation flows:

- **[`Step()`](core/flow_builder.go)**: Define a conversation step
- **[`Prompt()`](core/prompt_composer.go)**: Set the message to show users
- **[`Process()`](core/flow_types.go)**: Handle user responses and decide next action
- **[`WithInlineKeyboard()`](core/inline_keyboard_builder.go)**: Add interactive buttons

### 5. Flow Control
Within process functions, you can:
- [`NextStep()`](core/flow_types.go): Move to the next step
- [`Retry()`](core/flow_types.go): Ask for input again
- [`GoToStep("stepName")`](core/flow_types.go): Jump to a specific step
- [`CompleteFlow()`](core/flow_types.go): Finish the conversation

### 6. Data Storage
- [`ctx.SetFlowData("key", value)`](core/context.go): Store data for the current step
- [`ctx.GetFlowData("key")`](core/context.go): Retrieve stored data
- Data persists throughout the entire flow

## Environment Setup

For production bots, always use environment variables for your bot token:

```bash
# Linux/macOS
export TELEGRAM_BOT_TOKEN="your_bot_token_here"

# Windows
set TELEGRAM_BOT_TOKEN=your_bot_token_here
```

## Error Handling

Teleflow provides built-in error handling for flows:

```go
flow := core.NewFlow("example").
    OnError(core.OnErrorCancel()).  // Cancel flow on errors
    // ... rest of flow
```

Options include:
- `OnErrorCancel()`: Cancel the flow
- `OnErrorRetry()`: Retry the current step
- Custom error handlers

## Advanced Features

Teleflow supports many advanced features:

- **Middleware**: Authentication, rate limiting, logging
- **Templates**: Dynamic message generation with data
- **Image Support**: Send photos with messages
- **State Management**: Persistent data across bot restarts
- **Access Control**: Permission-based command access
- **Keyboard Builders**: Complex interactive keyboards

## Next Steps

Now that you have your first bot running, explore these resources to learn more:

- **[Architecture Overview](./architecture.md)**: Understand Teleflow's design and components
- **[API Guide](./api_guide.md)**: Detailed reference for all available APIs
- **[README](./README.md)**: Package overview and quick examples

## Example Projects

Check out the complete examples in the repository:

- [`example/basic-flow/`](../example/basic-flow/): Complete registration bot with images and keyboards
- [`example/process-message-actions/`](../example/process-message-actions/): Advanced message handling
- [`example/template/`](../example/template/): Template system showcase

## Troubleshooting

**Bot not responding?**
- Verify your bot token is correct
- Check that the bot is not already running elsewhere
- Ensure your bot has the necessary permissions

**Flow not working?**
- Make sure you've registered the flow with `bot.RegisterFlow()`
- Check that step names are unique within each flow
- Verify process functions return appropriate `ProcessResult` values

**Need help?**
- Check the [examples](../example/) directory for working code
- Review the [API documentation](./api_guide.md) for detailed function references

Happy bot building with Teleflow! 🤖✨