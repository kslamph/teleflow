# Teleflow

[![Go Reference](https://pkg.go.dev/badge/github.com/kslamph/teleflow.svg)](https://pkg.go.dev/github.com/kslamph/teleflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/kslamph/teleflow)](https://goreportcard.com/report/github.com/kslamph/teleflow)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)

A powerful, production-ready Go framework for building sophisticated Telegram bots with intuitive APIs, advanced flow management, and comprehensive middleware support.

## ✨ Key Features

- **🔒 Type-Safe APIs**: Strongly typed interfaces for commands, callbacks, and user interactions
- **🌊 Advanced Flow Management**: Build complex multi-step conversations with validation, branching logic, and state persistence
- **🛡️ Comprehensive Middleware**: Built-in middleware for logging, authentication, rate limiting, error recovery, and custom extensions
- **⌨️ Intuitive Keyboard System**: Easy-to-use reply and inline keyboard builders with callback handling
- **💾 Persistent State Management**: User state persistence across conversation flows and bot restarts
- **📝 Powerful Template System**: Message templating with Go's template engine for dynamic content
- **🎯 Context-Rich Operations**: Rich context objects with helper methods for common bot operations
- **🚀 Production Ready**: Built for scale with proper error handling, logging, and monitoring support
- **📖 Comprehensive Documentation**: Extensive documentation with practical examples and guides

## 🚀 Quick Start

### Basic Bot Example

```go
package main

import (
    "log"
    "github.com/kslamph/teleflow/core"
)

func main() {
    // Create a new bot instance
    bot := core.NewBot("YOUR_BOT_TOKEN")
    
    // Handle the /start command
    bot.HandleCommand("/start", func(ctx *core.Context) error {
        keyboard := core.NewReplyKeyboard().
            AddRow("🏠 Home", "📊 Stats").
            AddRow("⚙️ Settings", "❓ Help")
        
        return ctx.ReplyWithKeyboard("🎉 Welcome to Teleflow!\n\nChoose an option:", keyboard)
    })
    
    // Handle button presses
    bot.HandleText("🏠 Home", func(ctx *core.Context) error {
        return ctx.Reply("You're now at the home screen!")
    })
    
    // Start the bot
    log.Fatal(bot.Start())
}
```

### Flow-Based Conversation Example

```go
// Create a user registration flow
flow := core.NewFlow("user_registration").
    AddStep("name", core.StepTypeText, "What's your name?").
    AddStep("age", core.StepTypeText, "How old are you?").
    AddStep("confirm", core.StepTypeConfirmation, "Confirm registration?")

// Register the flow
bot.RegisterFlow(flow, func(ctx *core.Context, result map[string]string) error {
    return ctx.Reply(fmt.Sprintf("Welcome %s! You are %s years old.",
        result["name"], result["age"]))
})

// Start the flow
bot.HandleCommand("/register", func(ctx *core.Context) error {
    return bot.StartFlow(ctx, "user_registration")
})
```

## 📦 Installation

### Prerequisites
- Go 1.19 or later
- A Telegram Bot Token (obtain from [@BotFather](https://t.me/botfather))

### Install Teleflow

```bash
go mod init your-bot-project
go get github.com/kslamph/teleflow
```

### Get Dependencies

```bash
go mod tidy
```

## 📚 Documentation

| Resource | Description |
|----------|-------------|
| [Getting Started Guide](docs/getting-started.md) | Complete setup and basic usage tutorial |
| [API Reference](docs/api-reference.md) | Comprehensive API documentation |
| [Flow System Guide](docs/flow-guide.md) | Advanced conversation flow patterns |
| [Middleware Development](docs/middleware-guide.md) | Creating custom middleware |

## 🎯 Examples

Explore our comprehensive examples showcasing different aspects of the framework:

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
- State persistence
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

# Run the basic bot example
cd examples/basic-bot
go run main.go

# Run the flow bot example
cd ../flow-bot
go run main.go

# Run the middleware bot example
cd ../middleware-bot
go run main.go
```

## 🏗️ Architecture

Teleflow is built with a modular architecture that promotes clean code and extensibility:

```
teleflow/
├── core/           # Core framework components
│   ├── bot.go      # Main bot implementation
│   ├── context.go  # Request context and helpers
│   ├── flow.go     # Conversation flow system
│   ├── keyboards.go # Keyboard abstractions
│   ├── middleware.go # Middleware system
│   └── ...
├── examples/       # Example implementations
├── docs/          # Comprehensive documentation
└── tests/         # Test suites
```

## 🚦 Project Status

**🎯 Production Ready** - Teleflow is stable and ready for production use with:
- ✅ Comprehensive test coverage
- ✅ Stable API design
- ✅ Production-grade error handling
- ✅ Performance optimizations
- ✅ Extensive documentation

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **🐛 Report Bugs**: Submit detailed bug reports via [GitHub Issues](https://github.com/kslamph/teleflow/issues)
2. **💡 Feature Requests**: Propose new features and enhancements
3. **📝 Documentation**: Help improve our documentation
4. **🔧 Code Contributions**: Submit pull requests for bug fixes and features

### Development Setup

```bash
# Clone the repository
git clone https://github.com/kslamph/teleflow.git
cd teleflow

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run specific example
go run examples/basic-bot/main.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 💬 Support & Community

| Resource | Description |
|----------|-------------|
| 📖 [Documentation](docs/) | Complete guides and API reference |
| 🐛 [GitHub Issues](https://github.com/kslamph/teleflow/issues) | Bug reports and feature requests |
| 💬 [Discussions](https://github.com/kslamph/teleflow/discussions) | Community questions and discussions |
| 📧 [Email Support](https://github.com/kslamph/teleflow/issues/new) | Create an issue for direct support |

---

<div align="center">

**⭐ Star this project if you find it useful!**

Made with ❤️ for the Go and Telegram bot development community.

[🚀 Get Started](docs/getting-started.md) · [📖 Documentation](docs/) · [🎯 Examples](examples/)

</div>