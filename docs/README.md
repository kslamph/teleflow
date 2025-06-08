# Teleflow Core Package

Welcome to the documentation for the `teleflow/core` package!

## Introduction

The `teleflow/core` package is the heart of the Teleflow framework, providing the essential tools and abstractions for building powerful and flexible Telegram bots in Go. It is designed to simplify bot development by offering a structured approach to handling updates, managing conversations, and interacting with the Telegram Bot API.

## Purpose and Key Features

`teleflow/core` aims to provide a robust foundation for bot development with features such as:

*   **Intuitive Bot Creation**: Easy setup and configuration of your Telegram bot with a clean, fluent API.
*   **Powerful Flow System**: Define complex conversation flows with a step-by-step, prompt-based API that handles user interactions seamlessly.
*   **Flexible Middleware**: Extend bot functionality with a unified middleware system for request processing, including built-in logging, authentication, rate limiting, and recovery middleware.
*   **Context-Aware Handlers**: Access request-specific data and helper methods within your handlers through a rich Context API.
*   **State Management**: Built-in support for managing user and flow state with pluggable state storage backends.
*   **Message Templating**: Easily create and manage dynamic messages with template support for Markdown, MarkdownV2, and HTML parsing.
*   **Keyboard Builders**: Conveniently construct inline and reply keyboards with callback handling and UUID mapping.
*   **Image Handling**: Support for various image sources including static files, base64 data, URLs, and file paths.
*   **Access Control**: Built-in permission system with automatic UI management for unauthorized users.
*   **Error Handling**: Comprehensive error handling with configurable strategies (cancel, retry, ignore) for flow steps.
*   **Menu Button Integration**: Support for bot menu buttons with web app and command configurations.
*   **Main Menu Keyboards**: Persistent reply keyboards that appear below the chat input box for quick user actions.

## Architecture Highlights

The `teleflow/core` package is built around several key components:

*   **Bot**: The main orchestrator that handles updates, manages handlers, and coordinates all system components.
*   **Flow System**: A sophisticated conversation management system with flow builders, step definitions, and user state tracking.
*   **Context**: A rich request context that provides access to the Telegram update, bot API, state management, and helper methods.
*   **Middleware Stack**: A unified middleware system that processes all handler types consistently.
*   **Template Engine**: A powerful templating system with support for multiple parse modes and built-in helper functions.
*   **State Management**: Pluggable state storage with an in-memory implementation provided out of the box.

## Dive Deeper

To learn more about using the `teleflow/core` package, explore the following documents:

*   **[Getting Started](./get_started.md)**: Learn how to install Teleflow and build your first bot.
*   **[Architecture Overview](./architecture.md)**: Understand the core components and design principles of the `teleflow/core` package.
*   **[API Guide](./api_guide.md)**: A detailed reference for all available APIs, organized by functionality.

## Quick Example

Here's a glimpse of what building with `teleflow/core` looks like:

```go
package main

import (
    "github.com/kslamph/teleflow/core"
)

func main() {
    bot, err := core.NewBot("YOUR_BOT_TOKEN")
    if err != nil {
        panic(err)
    }

    // Simple command handler
    bot.HandleCommand("start", func(ctx *core.Context) error {
        return ctx.SendPromptText("Hello! Welcome to my bot.")
    })

    // Start a conversation flow
    bot.HandleCommand("survey", func(ctx *core.Context) error {
        return ctx.StartFlow("user_survey")
    })

    bot.Start()
}
```

The `teleflow/core` package makes it easy to build everything from simple command bots to complex conversational applications with rich user interactions.