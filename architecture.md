# Teleflow Architecture Overview

This document provides a high-level overview of the Teleflow Go package architecture. Teleflow is designed to be a powerful and flexible framework for building sophisticated Telegram bots.

## Core Components

The `core/` directory houses the fundamental building blocks of the Teleflow framework:

*   **`bot.go` ([core/bot.go](core/bot.go)): The Bot Engine**
    *   The central `Bot` struct orchestrates all operations.
    *   Manages bot API interactions, handler registration (commands, text, callbacks), middleware application, and flow management.
    *   Initializes and starts the bot, processing incoming updates.

*   **`context.go` ([core/context.go](core/context.go)): Interaction Context**
    *   The `Context` struct is passed to all handlers and middleware.
    *   Encapsulates the incoming Telegram update, provides access to the `Bot` instance, and offers helper methods for replying, sending messages with keyboards, managing flow state, and accessing user/chat information.
    *   Includes `MenuContext` for permission-based UI elements.

*   **`callbacks.go` ([core/callbacks.go](core/callbacks.go)): Callback Handling**
    *   Manages `CallbackHandler` interfaces and a `CallbackRegistry`.
    *   Provides type-safe handling of inline keyboard button interactions, supporting pattern matching and data extraction from callback queries.

*   **`flow.go` ([core/flow.go](core/flow.go)): Conversational Flows**
    *   Implements the `FlowManager`, `Flow`, and `FlowStep` structures for creating multi-step conversational interfaces.
    *   Supports input validation, conditional branching, state management within flows, and timeout handling.
    *   Uses a `FlowBuilder` for a fluent API to define flows.

*   **`keyboards.go` ([core/keyboards.go](core/keyboards.go)): UI Keyboards**
    *   Provides abstractions (`ReplyKeyboard`, `InlineKeyboard`) for creating Telegram reply and inline keyboards.
    *   Offers a fluent API to add buttons, rows, and configure keyboard properties (e.g., resizable, one-time).
    *   Supports special button types like request contact/location and web apps.

*   **`menu_button.go` ([core/menu_button.go](core/menu_button.go)): Bot Menu Button**
    *   Manages the configuration and setting of Telegram's native menu button.
    *   Supports different menu button types: commands, web app, or default.
    *   Handles registration of bot commands with Telegram when using the `commands` type menu button.

*   **`middleware.go` ([core/middleware.go](core/middleware.go)): Request Processing Pipeline**
    *   Defines the `MiddlewareFunc` type and provides a system for intercepting and processing requests before they reach the main handlers.
    *   Middleware can be used for logging, authentication, rate limiting, error recovery, and other cross-cutting concerns.
    *   Includes pre-built middleware like `LoggingMiddleware`, `AuthMiddleware`, `RateLimitMiddleware`, and `RecoveryMiddleware`.

*   **`state.go` ([core/state.go](core/state.go)): State Management**
    *   Defines the `StateManager` interface for persisting user and conversation data.
    *   Includes a default `InMemoryStateManager` for session-based state.
    *   Designed to be extensible with custom storage backends (e.g., Redis, databases).

*   **`templates.go` ([core/templates.go](core/templates.go)): Dynamic Messages**
    *   Implements a template system using Go's `text/template` engine.
    *   Allows for dynamic message content generation with variable substitution, conditional logic, loops, and custom functions.
    *   Supports different `ParseMode` (Markdown, MarkdownV2, HTML, None) and provides helpers for escaping content appropriately.

## High-Level Request Lifecycle

1.  **Update Reception:** The `Bot` receives an update from the Telegram API (e.g., new message, callback query).
2.  **Context Creation:** A `Context` object is created, encapsulating the update and bot instance.
3.  **Middleware Execution:** The request passes through the chain of registered global middleware. Each middleware can process the request, modify the context, or halt further processing.
4.  **Flow Handling:**
    *   The `FlowManager` checks if the user is currently in an active flow.
    *   If in a flow, the `FlowManager` processes the update according to the current flow step's logic, including validation and transitions.
5.  **Handler Resolution:** If not handled by a flow (or if the flow allows global commands):
    *   The `Bot` resolves the appropriate handler based on the update type:
        *   Command handlers (`HandleCommand`) for messages starting with `/`.
        *   Text handlers (`HandleText`) for regular text messages.
        *   Callback handlers (`RegisterCallback` via `CallbackRegistry`) for inline button presses.
6.  **Handler Execution:** The resolved handler function is executed, receiving the `Context` object.
7.  **Interaction & Response:** The handler uses the `Context` to interact with the user, potentially:
    *   Sending replies (text, templates).
    *   Displaying keyboards.
    *   Starting or continuing flows.
    *   Modifying state.
8.  **Error Handling:** Errors returned by handlers or middleware are logged. Recovery middleware can catch panics.

## Key Design Principles

*   **Modularity:** Each core concern (flows, state, keyboards, etc.) is handled by a dedicated component.
*   **Type Safety:** Leverages Go's type system, especially for handlers and callbacks.
*   **Extensibility:** Designed for extension, e.g., custom `StateManager`, `AccessManager`, and `MiddlewareFunc`.
*   **Developer Experience:** Aims for an intuitive API with fluent builders (e.g., `FlowBuilder`, keyboard methods) and comprehensive `Context` helpers.
*   **Clear Separation of Concerns:** UI (keyboards, templates), logic (handlers, flows), and infrastructure (middleware, state) are distinct.