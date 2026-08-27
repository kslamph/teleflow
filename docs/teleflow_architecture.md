# Teleflow Package Architecture

## 1. Introduction

The `teleflow` package provides a simple framework for building Telegram bots in Go. It simplifies the development of conversational flows by offering robust state management, middleware support, and extensive keyboard handling. Key features include:

*   **Type-Safe Flows:** Use Go generics for compile-time safety
*   **Automatic State Management:** No manual session handling needed
*   **Middleware Support:** Standard `Handler` interface for authentication, logging, etc.
*   **Keyboard Management:** Easily create and handle inline keyboards for interactive prompts
*   **Large Callback Data:** Built-in mechanism to pass complex data through buttons
*   **Concurrent Operations:** Designed for concurrent handling of user updates and flow states

This document outlines the core architecture of the `teleflow` package, detailing its main components, their responsibilities, and key interaction patterns.

## 2. Core Components & Architecture

The `teleflow` package is built around a set of interconnected components that work together to manage bot interactions and conversation flows.

```mermaid
graph TD
    A[Telegram API] --> B(Bot);
    B --> C[Router];
    C --> D{Command Match?};
    D -- Yes --> E[Handler];
    D -- No --> F[Default Handler];
    E -- Flow --> G[Flow[T]];
    E -- Func --> H[Func];
    G -- Uses --> I[Context];
    H -- Uses --> I;
    F -- Uses --> I;
    I -- Sends via --> J[*tgbotapi.BotAPI];

    style B fill:#f9f,stroke:#333,stroke-width:2px
    style G fill:#ccf,stroke:#333,stroke-width:2px
    style I fill:#cfc,stroke:#333,stroke-width:2px
```

### `Bot` ([`core/bot.go`](core/bot.go))

*   **Responsibility:** The central orchestrator of the bot. It initializes the connection with the Telegram API, receives incoming updates, and routes them to the appropriate handlers via a `Router`.
*   **Key Types:** `Bot`, `Handler`.
*   **Key Methods:**
    *   `NewBot(token string, handler Handler)`: Creates a new bot instance.
    *   `Start()`: Begins listening for and processing Telegram updates.
    *   `processUpdate(update tgbotapi.Update)`: The core internal method for handling each incoming update.
    *   `RegisterData(data interface{})`: Registers complex data for callbacks.
    *   `GetData(id string)`: Retrieves complex callback data.

### `Router` ([`core/flow.go`](core/flow.go))

*   **Responsibility:** Routes incoming updates to the appropriate handlers based on command names.
*   **Key Types:** `Router`.
*   **Key Methods:**
    *   `NewRouter()`: Creates a new router.
    *   `Handle(command string, handler Handler)`: Registers a handler for a specific command.
    *   `Default(handler Handler)`: Registers a fallback handler.
    *   `HandleUpdate(ctx *Context)`: Processes an update and routes it to the appropriate handler.

### `Flow[T]` ([`core/flow.go`](core/flow.go))

*   **Responsibility:** Represents a complete, defined multi-step conversation with automatic state management. It's parameterized with a generic type `T` which defines the structure of the data that will be collected and maintained throughout the flow.
*   **Key Types:** `Flow[T]`, `Step[T]`, `FlowAction`.
*   **Key Methods:**
    *   `NewFlow[T]()`: Creates a new flow for a given data type.
    *   `AddStep(step Step[T]) (*Flow[T], error)`: Adds a new step to the flow.
    *   `OnComplete(handler func(*Context, T) error)`: Sets a callback for successful flow completion.
    *   `HandleUpdate(ctx *Context)`: Processes an update in the context of an active flow.
    *   `IsUserInFlow(chatID, userID int64) bool`: Checks if a user is currently in this flow.
    *   `CancelFlow(chatID, userID int64)`: Terminates a flow for a user.

### `Step[T]` ([`core/flow.go`](core/flow.go))

*   **Responsibility:** Represents a single stage or interaction point within a `Flow[T]`. Each step defines what prompt to show the user and how to process their response.
*   **Key Types:** `Step[T]`.
*   **Key Fields:**
    *   `Name`: Unique identifier for the step.
    *   `Prompt`: A function that sends a message to the user.
    *   `Process`: A function that takes user input and decides the next action.

### `Context` ([`core/context.go`](core/context.go))

*   **Responsibility:** An object passed to all handlers. It encapsulates the incoming Telegram update and provides convenient methods for sending messages and managing data. Each `Context` is specific to a single update.
*   **Key Types:** `Context`.
*   **Key Methods:**
    *   `UserID()`, `ChatID()`: Get user and chat identifiers.
    *   `Text()`: Get text from message or callback data.
    *   `Set(key string, value interface{})`, `Get(key string)`: Manage request-scoped data.
    *   `ReplyText(text string)`: Sends a simple text message.
    *   `ReplyWithKeyboard(text string, markup interface{})`: Sends a message with a keyboard.
    *   `RegisterCallback(handler func(*Context) error)`: Registers a one-time callback handler.
    *   `RegisterData(data interface{})`: Registers complex data for callbacks.
    *   `GetData()`: Retrieves complex callback data.
    *   `AnswerCallback()`: Responds to a Telegram callback query.
    *   `EditMessageText(messageID int, text string)`: Edits an existing message.
    *   `DeleteMessage(messageID int)`: Deletes a message.

### `InlineKeyboard` ([`core/flow.go`](core/flow.go))

*   **Responsibility:** Provides a fluent API to define and construct inline keyboards.
*   **Key Types:** `InlineKeyboard`.
*   **Key Methods:**
    *   `NewInlineKeyboard()`: Creates a new keyboard builder.
    *   `AddRow(buttons ...tgbotapi.InlineKeyboardButton)`: Adds a new row of buttons.
    *   `Build()`: Finalizes the keyboard and returns a `tgbotapi.InlineKeyboardMarkup`.

### Handler ([`core/flow.go`](core/flow.go))

*   **Responsibility:** Defines the interface for handling Telegram updates. Both `Flow[T]` and `Func` implement this interface.
*   **Key Types:** `Handler`.
*   **Key Methods:**
    *   `HandleUpdate(ctx *Context)`: Processes an update.

### `Func` ([`core/flow.go`](core/flow.go))

*   **Responsibility:** An adapter to allow the use of ordinary functions as Handlers.
*   **Key Types:** `Func`.

## 3. Key Interaction Flows

### Bot Initialization and Update Handling

1.  **Initialization:**
    *   `teleflow.NewBot(token, handler)` is called with a `Handler` (typically a `Router`).
    *   A `*tgbotapi.BotAPI` client is created internally.
    *   A `Router` is provided to handle command routing.
2.  **Starting the Bot:**
    *   `bot.Start()` is called.
    *   The bot begins polling Telegram for updates via `bot.api.GetUpdatesChan()`.
3.  **Processing an Update (`Bot.processUpdate`)**:
    This is a critical method that orchestrates how each incoming update is handled.

    ```mermaid
    graph TD
        Start((Update Received)) --> CreateCtx[Create Context];
        CreateCtx --> IsCallback{Callback Query?};
        IsCallback -- Yes --> CheckReg{Registered Callback?};
        CheckReg -- Yes --> RunHandler[Run One-Time Handler];
        CheckReg -- No --> RouterCB[Forward to Router];
        IsCallback -- No --> RouterDirect[Forward to Router];
        RunHandler --> End((End));
        RouterCB --> IsCommand{Is Command?};
        RouterDirect --> IsCommand;
        IsCommand -- Yes --> RouteCommand[Route via Router];
        IsCommand -- No --> DefaultHandler[Use Default Handler];
        RouteCommand --> Handler[Handler.HandleUpdate];
        DefaultHandler --> Handler;
        Handler --> End;
    ```
    *   **Context Creation:** A new `Context` object is created for the update.
    *   **Callback Handling:** If the update is a `CallbackQuery`, the bot checks for a registered one-time callback handler (via `ctx.RegisterCallback`). If found, it runs that handler. If not found, the callback is forwarded to the `Router`. The framework does NOT auto-answer callbacks; the handler must call `ctx.AnswerCallback()` to stop the client loading spinner.
    *   **Command Routing:** If it's a command, the `Router` routes it to the appropriate handler.
    *   **Default Handling:** If it's not a command, the default handler is used.

### Command Handling

1.  Update arrives at `Bot.processUpdate`.
2.  `Context` is created.
3.  `Router.HandleUpdate(ctx)` is called.
4.  The `Router` checks if the message is a command:
    *   If yes, it looks up the command in its routes map.
    *   The corresponding `Handler` is executed.
5.  If not a command, the default handler is executed.

### Inline-Button Routing

Inline-keyboard button presses arrive as `CallbackQuery` updates, **not** as commands. `Router.HandleUpdate` only routes `msg.IsCommand()` messages, so buttons will never match a `router.Handle` route. Instead, buttons reach a `Flow` via:

1. **Default handler forwarding** (used by all examples):
   ```go
   router.Default(teleflow.Func(func(ctx *teleflow.Context) error {
       if flow.IsUserInFlow(ctx.ChatID(), ctx.UserID()) {
           return flow.HandleUpdate(ctx)
       }
       return ctx.ReplyText("Unknown command. Use /start.")
   }))
   ```
2. **One-time callback handler** — `ctx.RegisterCallback(handler)`, which runs once then auto-removes.

### Flow Lifecycle

1.  **Flow Definition:**
    *   Developer defines a data struct `T`.
    *   Creates a flow with `teleflow.NewFlow[T]()`.
    *   Defines `teleflow.Step[T]` with `Prompt` and `Process` functions.
    *   Adds steps with `flow.AddStep()`.
2.  **Starting a Flow:**
    *   A handler calls `router.Handle("command", flow)`.
    *   When a user sends the command, the flow's `HandleUpdate` method is called.
    *   If the user is not already in the flow, a new state is created for them.
    *   The first step's `Prompt` function is executed.
3.  **Processing a Flow Step:**
    *   User sends a message or clicks an inline button.
    *   Update arrives at `Bot.processUpdate`.
    *   `Flow.HandleUpdate(ctx)` is called.
    *   The flow retrieves the user's state.
    *   It calls the current step's `Process` function.
    *   The `Process` function executes custom logic and returns a `FlowAction`:
        *   `teleflow.Next()`: Advances to the next step.
        *   `teleflow.Retry()`: Re-prompts the current step.
        *   `teleflow.End()`: Completes the flow.
        *   `teleflow.GoTo("step_name")`: Jumps to a specific step.
    *   The flow handles the `FlowAction` and either renders the next step's prompt or completes the flow.
4.  **Flow Completion:**
    *   Triggered by `teleflow.End()` from a `Process` function or by reaching the end of steps.
    *   If `flow.OnComplete` handler is set, it's executed.
    *   The user's state is deleted.

### Message Sending

1.  Handler or flow step calls `ctx.ReplyText()` or `ctx.ReplyWithKeyboard()`.
2.  These methods construct the appropriate `tgbotapi.Chattable` and send it via `bot.api.Send()`.

## 4. Important Interfaces

*   **`Handler` ([`core/flow.go`](core/flow.go)):** Defines the contract for handling updates. Implemented by `Flow[T]` and `Func`.

## 5. Extensibility

*   **Middleware:** Add custom middleware by implementing the `Handler` interface.
*   **Custom Handlers:** Define various handlers for bot interactions.
*   **Flow Logic:** The `Process` function within each flow step allows for arbitrary Go code to handle user input and make decisions.