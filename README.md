# Teleflow

[![Go Reference](https://pkg.go.dev/badge/github.com/kslamph/teleflow/core.svg)](https://pkg.go.dev/github.com/kslamph/teleflow/core)
[![Go Report Card](https://goreportcard.com/badge/github.com/kslamph/teleflow)](https://goreportcard.com/report/github.com/kslamph/teleflow)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)

A powerful, production-ready Go framework for building sophisticated Telegram bots with intuitive APIs, advanced flow management, and comprehensive middleware support.

## ✨ Key Features

- **🔒 Type-Safe APIs**: Strongly typed interfaces for commands, callbacks, and user interactions.
- **🌊 Advanced Flow Management**: Build complex multi-step conversations with validation, branching logic, and state persistence within flows.
- **🛡️ Comprehensive Middleware**: Built-in middleware for logging, authentication, rate limiting, error recovery, and easy custom extensions.
- **⌨️ Intuitive Keyboard System**: Easy-to-use reply and inline keyboard builders with robust callback handling.
- **💾 Flexible State Management**: In-memory by default, with support for custom persistent state managers.
- **📝 Powerful Template System**: Message templating with Go's `text/template` engine for dynamic content and various parse modes.
- **🎯 Context-Rich Operations**: Rich context objects with helper methods for common bot operations.
- **🚀 Production Ready**: Designed with considerations for scale, error handling, and logging.
- **📖 Comprehensive Documentation**: Extensive guides and examples to get you started and master advanced features.

## 🚀 Quick Start

### Basic Bot Example

```go
package main

import (
	"log"
	"os" // For BOT_TOKEN

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable not set")
	}

	bot, err := teleflow.NewBot(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Handle the /start command
	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		keyboard := teleflow.NewReplyKeyboard().
			AddButton("🏠 Home").AddButton("📊 Stats").AddRow().
			AddButton("⚙️ Settings").AddButton("❓ Help").AddRow().
			Resize() // Optional: make keyboard more compact

		return ctx.Reply("🎉 Welcome to Teleflow!\n\nChoose an option:", keyboard)
	})

	// Handle button presses (text messages)
	bot.HandleText("🏠 Home", func(ctx *teleflow.Context) error {
		return ctx.Reply("You're now at the home screen!")
	})
	// Add other text handlers for "📊 Stats", "⚙️ Settings", "❓ Help"

	// Start the bot
	log.Println("Bot starting...")
	if err := bot.Start(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
}
```

### Flow-Based Conversation Example

```go
package main

import (
	"fmt"
	"log"
	"os"

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable not set")
	}
	bot, err := teleflow.NewBot(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create a user registration flow
	registrationFlow := teleflow.NewFlow("user_registration").
		Step("get_name").
			OnStart(func(ctx *teleflow.Context) error {
				return ctx.Reply("Let's register you! What's your name?")
			}).
			OnInput(func(ctx *teleflow.Context) error {
				ctx.Set("name", ctx.Update.Message.Text) // Save name to flow data
				return nil
			}).
			NextStep("get_age").
		Step("get_age").
			OnStart(func(ctx *teleflow.Context) error {
				name, _ := ctx.Get("name")
				return ctx.Reply(fmt.Sprintf("Nice, %s! How old are you?", name.(string)))
			}).
			OnInput(func(ctx *teleflow.Context) error {
				ctx.Set("age", ctx.Update.Message.Text) // Save age to flow data
				return nil
			}).
		OnComplete(func(ctx *teleflow.Context) error {
			name, _ := ctx.Get("name")
			age, _ := ctx.Get("age")
			// Type assertions are important for data retrieved from flow context
			return ctx.Reply(fmt.Sprintf("Registration complete! Welcome %s, aged %s.",
				name.(string), age.(string)))
		}).
		Build()

	// Register the flow
	bot.RegisterFlow(registrationFlow)

	// Command to start the flow
	bot.HandleCommand("register", func(ctx *teleflow.Context) error {
		if ctx.IsUserInFlow() {
			return ctx.Reply("You are already in a process. Type /cancel to exit first.")
		}
		return ctx.StartFlow("user_registration")
	})

	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		return ctx.Reply("Hello! Type /register to begin the registration process.")
	})
	
	log.Println("Bot starting...")
	if err := bot.Start(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
}
```

## 📦 Installation

### Prerequisites
- Go 1.19 or later
- A Telegram Bot Token (obtain from [@BotFather](https://t.me/botfather))

### Install Teleflow Core Package

```bash
go mod init your-bot-project # If you haven't already
go get github.com/kslamph/teleflow/core
```

### Get Dependencies
After adding imports to your Go files:
```bash
go mod tidy
```

## 📚 Documentation

Our comprehensive documentation provides guides, examples, and API references to help you get the most out of Teleflow.

- **Getting Started**:
    - [🚀 Getting Started Guide](docs/getting-started.md) - Your first steps with Teleflow.
- **Core Concepts & Guides**:
    - [🏛️ Architecture Overview](docs/architecture.md) - Understand the framework's structure.
    - [🧩 Handlers Guide](docs/handlers-guide.md) - Responding to commands, text, and callbacks.
    - [⌨️ Keyboards Guide](docs/keyboards-guide.md) - Building interactive reply and inline keyboards.
    - [🛡️ Middleware Guide](docs/middleware-guide.md) - Using and creating middleware.
    - [🌊 Conversational Flow Guide](docs/flow-guide.md) - Managing multi-step conversations.
    - [⚙️ Menu Button Guide](docs/menu-button-guide.md) - Configuring the bot's menu button.
    - [📝 Templates Guide](docs/templates-guide.md) - Creating dynamic messages.
    - [💾 State Management Guide](docs/state-management-guide.md) - Persisting user and chat data.
- **API**:
    - [📄 API Reference](docs/api-reference.md) - Detailed package and type documentation (use `godoc` for full details).
- **Examples**:
    - Explore practical examples in the [`examples/`](examples/) directory of the Teleflow repository.

## 🎯 Examples

Explore our comprehensive examples showcasing different aspects of the framework, available in the `examples/` directory of the [Teleflow GitHub repository](https://github.com/kslamph/teleflow).

### 🤖 [Basic Bot](examples/basic-bot/)
Simple command-based bot demonstrating:
- Command handlers
- Text message processing
- Basic keyboard interactions
- Context usage

### 🌊 [Flow Bot](examples/flow-bot/)
Advanced conversational bot featuring:
- Multi-step conversation flows
- Input validation and branching
- State persistence within flows
- Complex user interactions

### 🛡️ [Middleware Bot](examples/middleware-bot/)
Demonstrates middleware capabilities:
- Authentication and authorization
- Request logging and monitoring
- Rate limiting
- Error handling and recovery
- Custom middleware development

### Getting Started with Examples

```bash
# Clone the repository
git clone https://github.com/kslamph/teleflow.git
cd teleflow

# Navigate to an example directory, e.g., basic-bot
cd examples/basic-bot

# Set your BOT_TOKEN environment variable
export BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"

# Run the example
go run main.go
```

## 🏗️ Architecture

Teleflow is built with a modular architecture that promotes clean code and extensibility. For a detailed overview, please see the [Architecture Overview](docs/architecture.md) document.

A high-level view of the project structure:
```
teleflow/
├── core/           # Core framework components
│   ├── bot.go      # Main bot implementation
│   ├── context.go  # Request context and helpers
│   ├── flow.go     # Conversation flow system
│   ├── keyboards.go # Keyboard abstractions
│   ├── middleware.go # Middleware system
│   ├── templates.go # Template engine
│   ├── state.go     # State management
│   └── ...
├── examples/       # Example implementations
├── docs/           # Comprehensive documentation
└── ...             # Other project files (tests, etc.)
```

## 🚦 Project Status

**🎯 Production Ready** - Teleflow is designed to be stable and suitable for production use with:
- ✅ Comprehensive test coverage (ongoing)
- ✅ Stable API design
- ✅ Production-grade error handling
- ✅ Performance considerations
- ✅ Extensive documentation

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **🐛 Report Bugs**: Submit detailed bug reports via [GitHub Issues](https://github.com/kslamph/teleflow/issues).
2. **💡 Feature Requests**: Propose new features and enhancements.
3. **📝 Documentation**: Help improve our documentation.
4. **🔧 Code Contributions**: Submit pull requests for bug fixes and features. Please refer to our contribution guidelines (if available) or open an issue to discuss significant changes.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/kslamph/teleflow.git
cd teleflow

# Install dependencies
go mod tidy

# Run tests
go test ./core/... # Assuming tests are primarily in the core package

# Run specific example
# (Ensure BOT_TOKEN is set)
# go run examples/basic-bot/main.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 💬 Support & Community

| Resource | Description |
|----------|-------------|
| 📖 [Documentation](#documentation) | Complete guides and API reference links above |
| 🐛 [GitHub Issues](https://github.com/kslamph/teleflow/issues) | Bug reports and feature requests |
| 💬 [GitHub Discussions](https://github.com/kslamph/teleflow/discussions) | Community questions and discussions |

---

<div align="center">

**⭐ Star this project if you find it useful!**

Made with ❤️ for the Go and Telegram bot development community.

[🚀 Get Started](docs/getting-started.md) · [📖 Documentation](#documentation) · [🎯 Examples](examples/)

</div>