# Teleflow Handlers Guide

Handlers are the backbone of your Teleflow bot's interactivity. They are Go functions that you define to process incoming Telegram updates, such as commands, text messages, and inline keyboard button presses. This guide explores the different types of handlers available in Teleflow and how to use them effectively.

## Table of Contents

- [Introduction to Handlers](#introduction-to-handlers)
  - [The `teleflow.Context`](#the-teleflowcontext)
- [1. Command Handlers](#1-command-handlers)
  - [Registering Command Handlers](#registering-command-handlers)
  - [Accessing Command Data](#accessing-command-data)
  - [Example](#example)
- [2. Text Handlers](#2-text-handlers)
  - [Registering Text Handlers](#registering-text-handlers)
  - [Accessing Text Data](#accessing-text-data)
  - [Example](#example-1)
  - [Use Case: Handling Reply Keyboard Buttons](#use-case-handling-reply-keyboard-buttons)
- [3. Callback Handlers (for Inline Keyboards)](#3-callback-handlers-for-inline-keyboards)
  - [Registering Callback Handlers](#registering-callback-handlers)
  - [The `CallbackHandler` Interface](#the-callbackhandler-interface)
  - [Using `SimpleCallback`](#using-simplecallback)
  - [Pattern Matching](#pattern-matching)
  - [Example](#example-2)
- [Handler Execution Order](#handler-execution-order)
- [Best Practices for Handlers](#best-practices-for-handlers)
- [Next Steps](#next-steps)

## Introduction to Handlers

In Teleflow, a handler is typically a function with the signature `func(ctx *teleflow.Context) error`. This function is executed when a specific condition is met, like a user sending a particular command or text.

### The `teleflow.Context`

Every handler receives a `*teleflow.Context` ([core/context.go](../core/context.go)) as its parameter. This `Context` object is crucial as it:
- Provides access to the incoming Telegram update (`ctx.Update`).
- Offers information about the user and chat (`ctx.UserID()`, `ctx.ChatID()`).
- Contains helper methods for sending replies (`ctx.Reply()`, `ctx.ReplyTemplate()`), editing messages, managing state, starting flows, and more.

All interactions with the Telegram API from within a handler should ideally go through the `Context` object.

## 1. Command Handlers

Command handlers are used to respond to messages that start with a slash (`/`), known as bot commands (e.g., `/start`, `/help`, `/settings`).

### Registering Command Handlers

You register a command handler using `bot.HandleCommand(commandName string, handler teleflow.HandlerFunc)`:
```go
bot.HandleCommand("start", func(ctx *teleflow.Context) error {
    return ctx.Reply("Welcome to the bot!")
})

bot.HandleCommand("help", func(ctx *teleflow.Context) error {
    return ctx.Reply("Here's how to use the bot...")
})
```
The `commandName` should be the command without the leading slash (e.g., "start", not "/start").

### Accessing Command Data

Within a command handler, you can access:
- The command itself: `ctx.Update.Message.Command()` (returns "start", "help", etc.)
- Arguments passed to the command: `ctx.Update.Message.CommandArguments()` (returns the string after the command)

### Example
```go
bot.HandleCommand("greet", func(ctx *teleflow.Context) error {
    args := ctx.Update.Message.CommandArguments()
    if args == "" {
        return ctx.Reply("Hello there!")
    }
    return ctx.Reply("Hello, " + args + "!")
})
// User sends: /greet John
// Bot replies: Hello, John!
```

## 2. Text Handlers

Text handlers are used to respond to regular text messages sent by users or text sent when a user presses a button on a [Reply Keyboard](keyboards-guide.md#reply-keyboards).

### Registering Text Handlers

You register text handlers using `bot.HandleText(textToMatch string, handler teleflow.HandlerFunc)`:

- **Specific Text Match**: The handler triggers if the incoming message text exactly matches `textToMatch`.
  ```go
  bot.HandleText("Show Menu", func(ctx *teleflow.Context) error {
      // Logic to show a menu
      return ctx.Reply("Here is the menu...")
  })
  ```

- **Default Text Handler (Catch-all)**: If `textToMatch` is an empty string (`""`), the handler acts as a default or catch-all for any text message that doesn't match a more specific text handler or command.
  ```go
  bot.HandleText("", func(ctx *teleflow.Context) error {
      return ctx.Reply("I didn't understand that. Try /help.")
  })
  ```
  **Note:** The default text handler will only be triggered if no command handler or specific text handler matches.

### Accessing Text Data

Within a text handler, the incoming message text can be accessed via `ctx.Update.Message.Text`.

### Example
```go
// Specific handler for "hello"
bot.HandleText("hello", func(ctx *teleflow.Context) error {
    return ctx.Reply("Hi there!")
})

// Default handler for any other text
bot.HandleText("", func(ctx *teleflow.Context) error {
    userText := ctx.Update.Message.Text
    return ctx.Reply("You said: " + userText)
})
```

### Use Case: Handling Reply Keyboard Buttons

Reply keyboard buttons, when pressed, send their text as a regular message. You can use specific text handlers to process these button presses:
```go
// Assuming a reply keyboard with a "Profile" button
bot.HandleText("Profile", func(ctx *teleflow.Context) error {
    // Show user profile
    return ctx.Reply("Here is your profile...")
})
```

## 3. Callback Handlers (for Inline Keyboards)

Callback handlers are designed to process interactions with [Inline Keyboard](keyboards-guide.md#inline-keyboards) buttons. When an inline button is pressed, Telegram sends a `CallbackQuery` update to your bot, containing specific `data` associated with that button.

### Registering Callback Handlers

You register callback handlers using `bot.RegisterCallback(handler teleflow.CallbackHandler)`. Teleflow uses an interface, `teleflow.CallbackHandler`, for this.

### The `CallbackHandler` Interface
```go
type CallbackHandler interface {
    Pattern() string // The pattern to match against callback data
    Handle(ctx *Context, data string) error // The function to execute
}
```
- `Pattern()`: Returns a string that defines which callback data this handler should process. It can be an exact string or include a wildcard (`*`).
- `Handle(ctx *Context, data string) error`: This method is called when a callback query's data matches the pattern. The `data` argument here is the *extracted part* of the callback data if a wildcard was used, or the full callback data if it was an exact match (or an empty string if the pattern was an exact match to the callback data).

### Using `SimpleCallback`

For many common cases, you can use the `teleflow.SimpleCallback` helper function to easily create a `CallbackHandler`:
```go
import teleflow "github.com/kslamph/teleflow/core"

// Handler for callback data "confirm_action"
bot.RegisterCallback(teleflow.SimpleCallback("confirm_action", func(ctx *teleflow.Context, extractedData string) error {
    // extractedData will be "" in this case as it's an exact match
    return ctx.EditOrReply("Action confirmed!")
}))

// Handler for callback data starting with "item_"
bot.RegisterCallback(teleflow.SimpleCallback("item_*", func(ctx *teleflow.Context, itemID string) error {
    // If callback data was "item_123", itemID will be "123"
    return ctx.EditOrReply("You selected item: " + itemID)
}))
```

### Pattern Matching

- **Exact Match**: If `Pattern()` returns `"my_action"`, it will only handle callback data that is exactly `"my_action"`.
- **Wildcard Suffix**: If `Pattern()` returns `"prefix_*"`, it will handle any callback data starting with `"prefix_"`. The part of the callback data *after* the prefix is passed as the `data` argument to the `Handle` method.

### Example

Consider an inline keyboard with two buttons:
- Button 1: Text "View Details", Callback Data "view_product_101"
- Button 2: Text "Delete", Callback Data "delete_product_101"

```go
// Handler for viewing product details
bot.RegisterCallback(teleflow.SimpleCallback("view_product_*", func(ctx *teleflow.Context, productID string) error {
    // productID will be "101" if "view_product_101" was pressed
    message := "Showing details for product " + productID
    // Answer the callback query to remove the "loading" state on the button
    ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Loading details...")
    return ctx.EditOrReply(message)
}))

// Handler for deleting a product
bot.RegisterCallback(teleflow.SimpleCallback("delete_product_*", func(ctx *teleflow.Context, productID string) error {
    // productID will be "101"
    // Perform deletion logic...
    ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Product deleted")
    return ctx.EditOrReply("Product " + productID + " has been deleted.")
}))
```
**Important**: Always answer callback queries using `ctx.Bot.AnswerCallbackQuery(callbackQueryID, optionalText)` or by editing the message. This acknowledges the button press to Telegram and removes the loading spinner on the button. `ctx.EditOrReply` and `ctx.EditOrReplyTemplate` handle this automatically if they successfully edit the message.

## Handler Execution Order

Teleflow processes incoming updates and resolves handlers in a specific order:
1.  **Global Flow Exit Commands**: If the user is in a flow and sends a configured global exit command (e.g., `/cancel`), the flow is exited.
2.  **Flow Handling**: If the user is in an active [Flow](flow-guide.md), the `FlowManager` attempts to handle the update. If the flow handles it, no further handlers are typically processed.
3.  **Command Handlers**: If not handled by a flow, and the message is a command, registered command handlers are checked.
4.  **Specific Text Handlers**: If not a command, specific text handlers are checked against the message text.
5.  **Default Text Handler**: If no specific text handler matches, the default text handler (registered with `""`) is executed.
6.  **Callback Handlers**: If the update is a `CallbackQuery` (from an inline button), registered callback handlers are checked based on their patterns.

Middleware is applied before the respective handler (command, text, or callback) is invoked.

## Best Practices for Handlers

- **Keep Handlers Concise**: Aim for handlers that perform a specific task. For complex logic, break it down into smaller functions or consider using Flows.
- **Utilize `Context`**: Leverage the `ctx` object for all interactions and data access.
- **Error Handling**: Always handle errors returned by `Context` methods (e.g., `ctx.Reply`). Return errors from your handler if an operation fails, allowing middleware or the bot's core loop to log it.
- **Idempotency**: Where applicable (especially for callbacks that might be retried by Telegram), design handlers to be idempotent (safe to execute multiple times with the same effect).
- **Answer Callbacks**: For `CallbackHandler`s, ensure you answer the callback query to provide feedback to the user and Telegram.

## Next Steps

With a solid understanding of handlers, you're ready to explore more Teleflow features:
- [Keyboards Guide](keyboards-guide.md): Create interactive buttons.
- [Middleware Guide](middleware-guide.md): Add cross-cutting concerns.
- [Flow Guide](flow-guide.md): Build multi-step conversations.
- [Templates Guide](templates-guide.md): Craft dynamic messages.
- [API Reference](api-reference.md): For a detailed look at all available functions and types.