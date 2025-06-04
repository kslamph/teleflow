# Teleflow Handlers Guide

Handlers are the backbone of your Teleflow bot's interactivity. They are Go functions that you define to process incoming Telegram updates, such as commands, text messages, and inline keyboard button presses. This guide explores the different types of handlers available in Teleflow, their specific signatures for improved type safety, and how to use them effectively, including the refactored type-specific middleware system.

## Table of Contents

- [Introduction to Handlers](#introduction-to-handlers)
  - [The `teleflow.Context`](#the-teleflowcontext)
- [1. Command Handlers (`teleflow.CommandHandlerFunc`)](#1-command-handlers-teleflowcommandhandlerfunc)
  - [Registering Command Handlers](#registering-command-handlers)
  - [Accessing Command Data (Direct Parameters)](#accessing-command-data-direct-parameters)
  - [Example](#example)
- [2. Text Handlers](#2-text-handlers)
  - [Specific Text Handlers (`teleflow.TextHandlerFunc`)](#specific-text-handlers-teleflowtexthandlerfunc)
    - [Registering Specific Text Handlers](#registering-specific-text-handlers)
  - [Default Text Handler (`teleflow.DefaultTextHandlerFunc`)](#default-text-handler-teleflowdefaulttexthandlerfunc)
    - [Registering the Default Text Handler](#registering-the-default-text-handler)
  - [Accessing Text Data (Direct Parameters)](#accessing-text-data-direct-parameters)
  - [Example](#example-1)
  - [Use Case: Handling Reply Keyboard Buttons](#use-case-handling-reply-keyboard-buttons)
- [3. Callback Handlers (for Inline Keyboards)](#3-callback-handlers-for-inline-keyboards)
  - [Registering Callback Handlers](#registering-callback-handlers)
  - [The `CallbackHandler` Interface](#the-callbackhandler-interface)
  - [Using `SimpleCallback`](#using-simplecallback)
  - [Pattern Matching](#pattern-matching)
  - [Example](#example-2)
- [4. Flow Step Handlers](#4-flow-step-handlers)
  - [`teleflow.FlowStepStartHandlerFunc`](#teleflowflowstepstarthandlerfunc)
  - [`teleflow.FlowValidatorFunc`](#teleflowflowvalidatorfunc)
  - [`teleflow.FlowStepInputHandlerFunc`](#teleflowflowstepinputhandlerfunc)
  - [`teleflow.FlowCompletionHandlerFunc`](#teleflowflowcompletionhandlerfunc)
  - [`teleflow.FlowCancellationHandlerFunc`](#teleflowflowcancellationhandlerfunc)
- [5. Middleware](#5-middleware)
  - [Middleware Function Signatures](#middleware-function-signatures)
  - [Registering Type-Specific Middleware](#registering-type-specific-middleware)
  - [Middleware Examples](#middleware-examples)
  - [Middleware for Flow Step Handlers](#middleware-for-flow-step-handlers)
- [6. Handler Execution Order](#6-handler-execution-order)
- [7. Best Practices for Handlers](#7-best-practices-for-handlers)
- [8. Next Steps](#8-next-steps)

## Introduction to Handlers

In Teleflow, while historically a handler was often a generic function like `func(ctx *teleflow.Context) error`, the framework has shifted towards **specific handler types** (e.g., `teleflow.CommandHandlerFunc`, `teleflow.TextHandlerFunc`). This approach significantly improves type safety and clarity by providing dedicated signatures that directly include relevant input data (like the command text or message content) as parameters. These specific handlers are executed when a corresponding condition is met, such as a user sending a particular command or text message.

### The `teleflow.Context`

Even with specific handler types passing primary input data directly as parameters, every handler still receives a `*teleflow.Context` ([`core/context.go`](../core/context.go):1) as its first parameter. This `Context` object remains crucial as it:
- Provides access to the underlying Telegram update (`ctx.Update`) if needed for less common data.
- Offers information about the user and chat (`ctx.UserID()`, `ctx.ChatID()`).
- Contains essential helper methods for sending replies (`ctx.Reply()`, `ctx.ReplyTemplate()`), editing messages, managing state (`ctx.SetState()`, `ctx.GetState()`), starting and managing [Flows](flow-guide.md), accessing the bot instance (`ctx.Bot`), and more.
- Is the primary means for interacting with the Telegram API from within a handler.

## 1. Command Handlers (`teleflow.CommandHandlerFunc`)

Command handlers are used to respond to messages that start with a slash (`/`), known as bot commands (e.g., `/start`, `/help`, `/settings`).

The specific type for these handlers is `teleflow.CommandHandlerFunc`, defined as:
`func(ctx *teleflow.Context, command string, args string) error`

- `ctx`: The `teleflow.Context` for the current update.
- `command`: The command name itself (e.g., "start", "help"), without the leading slash.
- `args`: The string of arguments that follow the command. If no arguments are provided, this will be an empty string.

### Registering Command Handlers

You register a command handler using `bot.HandleCommand(commandName string, handler teleflow.CommandHandlerFunc)`:
```go
import (
    teleflow "github.com/kslamph/teleflow/core" // Assuming this import
    "log"
)

// ... in your bot setup ...
bot.HandleCommand("start", func(ctx *teleflow.Context, command string, args string) error {
    // 'command' will be "start", 'args' will be empty if user sent just /start
    log.Printf("Command: /%s, Args: '%s'", command, args)
    return ctx.Reply("Welcome to the bot!")
})

bot.HandleCommand("help", func(ctx *teleflow.Context, command string, args string) error {
    return ctx.Reply("Here's how to use the bot...")
})
```
The `commandName` should be the command without the leading slash (e.g., "start", not "/start").

### Accessing Command Data (Direct Parameters)

With `teleflow.CommandHandlerFunc`, the command and its arguments are directly passed to your handler function:
- The `command` parameter gives you the command name.
- The `args` parameter gives you the arguments string.

You no longer need to access them via `ctx.Update.Message.Command()` or `ctx.Update.Message.CommandArguments()` for typical use cases, though that data is still available through `ctx.Update` if needed for very specific scenarios.

### Example
```go
bot.HandleCommand("greet", func(ctx *teleflow.Context, command string, args string) error {
    // 'command' will be "greet"
    // 'args' will be "John" if user sends "/greet John"
    if args == "" {
        return ctx.Reply("Hello there! You can also try `/greet YourName`.")
    }
    return ctx.Reply("Hello, " + args + "!")
})
// User sends: /greet John
// Bot replies: Hello, John!

// User sends: /greet
// Bot replies: Hello there! You can also try `/greet YourName`.
```

## 2. Text Handlers

Text handlers are used to respond to regular text messages sent by users or text sent when a user presses a button on a [Reply Keyboard](keyboards-guide.md#reply-keyboards).

### Specific Text Handlers (`teleflow.TextHandlerFunc`)

For messages that exactly match a specific string, you use `teleflow.TextHandlerFunc`, defined as:
`func(ctx *teleflow.Context, text string) error`

- `ctx`: The `teleflow.Context` for the current update.
- `text`: The exact text of the incoming message that matched `textToMatch`.

#### Registering Specific Text Handlers

You register these handlers using `bot.HandleText(textToMatch string, handler teleflow.TextHandlerFunc)`:
```go
// import teleflow "github.com/kslamph/teleflow/core" // Assuming this import from above

bot.HandleText("Show Menu", func(ctx *teleflow.Context, text string) error {
    // 'text' will be "Show Menu"
    // Logic to show a menu
    return ctx.Reply("Here is the menu...")
})
```
The handler triggers if the incoming message text exactly matches `textToMatch`.

### Default Text Handler (`teleflow.DefaultTextHandlerFunc`)

For any text message that doesn't match a command or a specific text handler, you can set a default text handler. This uses `teleflow.DefaultTextHandlerFunc`, defined as:
`func(ctx *teleflow.Context, fullMessageText string) error`

- `ctx`: The `teleflow.Context` for the current update.
- `fullMessageText`: The complete text of the incoming message.

#### Registering the Default Text Handler

You register the default text handler using `bot.SetDefaultTextHandler(handler teleflow.DefaultTextHandlerFunc)`:
```go
bot.SetDefaultTextHandler(func(ctx *teleflow.Context, fullMessageText string) error {
    // 'fullMessageText' contains whatever the user sent
    return ctx.Reply("I didn't understand that ('" + fullMessageText + "'). Try /help.")
})
```
**Note:** The default text handler will only be triggered if no command handler or specific text handler (registered with `bot.HandleText`) matches.

### Accessing Text Data (Direct Parameters)

- For `teleflow.TextHandlerFunc`, the matching `text` is passed directly.
- For `teleflow.DefaultTextHandlerFunc`, the `fullMessageText` is passed directly.

You generally no longer need to access `ctx.Update.Message.Text` for these common cases, though it remains available.

### Example
```go
// Specific handler for "hello"
bot.HandleText("hello", func(ctx *teleflow.Context, text string) error {
    // 'text' will be "hello"
    return ctx.Reply("Hi there!")
})

// Default handler for any other text
bot.SetDefaultTextHandler(func(ctx *teleflow.Context, fullMessageText string) error {
    return ctx.Reply("You said: '" + fullMessageText + "'. If you need help, type /help.")
})
```

### Use Case: Handling Reply Keyboard Buttons

Reply keyboard buttons, when pressed, send their text as a regular message. You can use specific text handlers (`bot.HandleText`) to process these button presses:
```go
// Assuming a reply keyboard with a "Profile" button
bot.HandleText("Profile", func(ctx *teleflow.Context, text string) error {
    // 'text' will be "Profile"
    // Show user profile
    return ctx.Reply("Here is your profile...")
})
```

## 3. Callback Handlers (for Inline Keyboards)

Callback handlers are designed to process interactions with [Inline Keyboard](keyboards-guide.md#inline-keyboards) buttons. When an inline button is pressed, Telegram sends a `CallbackQuery` update to your bot, containing specific `data` associated with that button.

### Registering Callback Handlers

You register callback handlers using `bot.RegisterCallback(handler teleflow.CallbackHandler)`. Teleflow uses the `teleflow.CallbackHandler` interface for this.

### The `CallbackHandler` Interface

The `teleflow.CallbackHandler` interface is defined as:
```go
type CallbackHandler interface {
    Pattern() string // The pattern to match against callback data
    Handle(ctx *Context, fullCallbackData string, extractedData string) error // The function to execute
}
```
- `Pattern()`: Returns a string that defines which callback data this handler should process. It can be an exact string or include a wildcard (`*`) at the end of a prefix.
- `Handle(ctx *Context, fullCallbackData string, extractedData string) error`: This method is called when a callback query's data matches the pattern.
    - `ctx`: The `teleflow.Context` for the current update.
    - `fullCallbackData`: The complete, original data string from the callback query.
    - `extractedData`: If a wildcard (`*`) was used in the pattern (e.g., `"prefix_*" `), this parameter contains the part of the callback data that matched the wildcard. If it was an exact match (no wildcard), `extractedData` will be an empty string.

### Using `SimpleCallback`

For many common cases, you can use the `teleflow.SimpleCallback` helper function to easily create a `CallbackHandler`. Its function argument now also receives both `fullCallbackData` and `extractedData`:
`func(ctx *teleflow.Context, fullCallbackData string, extractedData string) error`

```go
// import teleflow "github.com/kslamph/teleflow/core" // Assuming this import from above
// import "log" // Assuming this import from above

// Handler for exact callback data "confirm_action"
bot.RegisterCallback(teleflow.SimpleCallback("confirm_action", func(ctx *teleflow.Context, fullData string, extractedData string) error {
    // fullData will be "confirm_action"
    // extractedData will be "" (empty string) as it's an exact match
    log.Printf("Confirm action called. Full data: %s, UserID: %d", fullData, ctx.UserID())
    return ctx.EditOrReply("Action confirmed!")
}))

// Handler for callback data starting with "item_"
bot.RegisterCallback(teleflow.SimpleCallback("item_*", func(ctx *teleflow.Context, fullData string, itemID string) error {
    // If callback data was "item_123":
    // fullData will be "item_123"
    // itemID (extractedData) will be "123"
    log.Printf("Item action. Full data: %s, Item ID: %s, UserID: %d", fullData, itemID, ctx.UserID())
    return ctx.EditOrReply("You selected item: " + itemID)
}))
```

### Pattern Matching

- **Exact Match**: If `Pattern()` returns `"my_action"`, it will only handle callback data that is exactly `"my_action"`. `extractedData` in the handler will be empty.
- **Wildcard Suffix**: If `Pattern()` returns `"prefix_*"`, it will handle any callback data starting with `"prefix_"`. The part of the callback data *after* the prefix is passed as the `extractedData` argument to the `Handle` method. `fullCallbackData` will contain the entire string (e.g., `"prefix_value"`).

### Example

Consider an inline keyboard with two buttons:
- Button 1: Text "View Details", Callback Data "view_product_101"
- Button 2: Text "Delete", Callback Data "delete_product_101"

```go
// Handler for viewing product details
bot.RegisterCallback(teleflow.SimpleCallback("view_product_*", func(ctx *teleflow.Context, fullData string, productID string) error {
    // If "view_product_101" was pressed:
    // fullData will be "view_product_101"
    // productID will be "101"
    message := "Showing details for product " + productID + " (Full data: " + fullData + ")"

    // Answer the callback query to remove the "loading" state on the button
    // This is good practice if not immediately editing the message.
    // Note: ctx.EditOrReply also answers the query if successful.
    ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Loading details...")
    return ctx.EditOrReply(message)
}))

// Handler for deleting a product
bot.RegisterCallback(teleflow.SimpleCallback("delete_product_*", func(ctx *teleflow.Context, fullData string, productID string) error {
    // If "delete_product_101" was pressed:
    // fullData will be "delete_product_101"
    // productID will be "101"
    // Perform deletion logic...
    log.Printf("Attempting to delete product %s (Full data: %s), UserID: %d", productID, fullData, ctx.UserID())

    ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Product deleted") // Explicit answer
    return ctx.EditOrReply("Product " + productID + " has been deleted.")
}))
```
**Important**: Always answer callback queries using `ctx.Bot.AnswerCallbackQuery(callbackQueryID, optionalText)` or by methods like `ctx.EditOrReply` that do so implicitly. This acknowledges the button press to Telegram and removes the loading spinner on the button. `ctx.EditOrReply` and `ctx.EditOrReplyTemplate` handle this automatically if they successfully edit the message.

## 4. Flow Step Handlers

Teleflow's [Flows](flow-guide.md) allow you to build multi-step conversations. Each step in a flow can have specific handlers for different events within that step's lifecycle. These handlers provide fine-grained control over user interaction during a flow.

All flow step handlers are configured using methods on the `FlowStepBuilder` when defining a flow step.

### `teleflow.FlowStepStartHandlerFunc`

Signature: `func(ctx *teleflow.Context) error`

- **Purpose**: This handler is executed when a user first enters a flow step. It's typically used to send the initial message for that step, like asking a question or presenting options.
- **Example Usage with `FlowBuilder`**:
  ```go
  // import teleflow "github.com/kslamph/teleflow/core"
  // import "strconv" // For examples below

  // Assuming 'flow' is a *teleflow.FlowBuilder instance
  flow.NewStep("ask_name").
      SetStartHandler(func(ctx *teleflow.Context) error {
          return ctx.Reply("What is your name?")
      })
  ```

### `teleflow.FlowValidatorFunc`

Signature: `func(input string) (isValid bool, message string, validatedInput interface{}, err error)`

- **Purpose**: This handler is called *before* the `FlowStepInputHandlerFunc` to validate the user's text input or callback data.
- **Parameters**:
    - `input`: The raw text or callback data string received from the user.
- **Return Values**:
    - `isValid bool`: `true` if the input is valid, `false` otherwise.
    - `message string`: If `isValid` is `false`, this message is sent back to the user explaining why their input was rejected.
    - `validatedInput interface{}`: If `isValid` is `true`, this can be the parsed, typed, or transformed version of the input (e.g., an integer if a number was expected). This value is then made available to the `FlowStepInputHandlerFunc`. Can be `nil` if no specific transformation is needed beyond validation.
    - `err error`: For reporting internal errors within the validator itself (e.g., database connection issue during validation). This error is logged by Teleflow but not typically shown to the user directly.
- **Example Usage with `FlowBuilder`**:
  ```go
  flow.NewStep("ask_age").
      SetStartHandler(func(ctx *teleflow.Context) error {
          return ctx.Reply("How old are you?")
      }).
      SetValidator(func(input string) (bool, string, interface{}, error) {
          age, err := strconv.Atoi(input)
          if err != nil {
              return false, "Please enter a valid number for your age.", nil, nil
          }
          if age <= 0 || age > 120 { // Age validation
              return false, "Please enter a realistic age (1-120).", nil, nil
          }
          return true, "", age, nil // age (int) is the validatedInput
      })
  ```

### `teleflow.FlowStepInputHandlerFunc`

Signature: `func(ctx *teleflow.Context, input string) error`

- **Purpose**: This handler is executed after the user's input (text message or callback data from an inline keyboard associated with the flow step) has been successfully validated by a `FlowValidatorFunc` (if one is defined for the step).
- **Parameters**:
    - `ctx`: The `teleflow.Context`.
    - `input`: The raw input string (text or callback data) that was validated.
- **Accessing Validated Input**: If a `FlowValidatorFunc` returned a non-nil `validatedInput`, it can be accessed within the `FlowStepInputHandlerFunc` using `ctx.Get(teleflow.ValidatedInputKey)` (or `ctx.Get("validated_input")` if using the string literal, though the constant `teleflow.ValidatedInputKey` is preferred). You'll need to type-assert it to its expected type.
- **Example Usage with `FlowBuilder`**:
  ```go
  flow.NewStep("ask_age").
      // ... SetStartHandler and SetValidator as above ...
      SetInputHandler(func(ctx *teleflow.Context, rawInput string) error {
          // rawInput is the original string, e.g., "25"
          validatedAge, ok := ctx.Get(teleflow.ValidatedInputKey).(int)
          if !ok {
              // This should ideally not happen if validator worked correctly and returned an int
              log.Printf("Error: validated_input was not an int or not found for user %d", ctx.UserID())
              return ctx.Reply("Something went wrong processing your age. Please try again.")
          }
          // Store the validated age in flow data
          ctx.FlowData()["age"] = validatedAge
          ctx.ReplyTemplate("Thanks! I've recorded your age as {{.Flow.age}}.", nil)
          // return ctx.NextStep() // or ctx.EndFlow() or another step name
          return ctx.EndFlow() // Example: end flow after collecting age
      })
  ```

### `teleflow.FlowCompletionHandlerFunc`

Signature: `func(ctx *teleflow.Context, flowData map[string]interface{}) error`

- **Purpose**: This handler is executed when the entire flow successfully completes (e.g., after `ctx.EndFlow()` is called from a step's input handler or if the flow naturally reaches its end).
- **Parameters**:
    - `ctx`: The `teleflow.Context`.
    - `flowData`: A map containing all the data collected and stored throughout the flow using `ctx.FlowData()[key] = value`.
- **Example Usage with `FlowBuilder`**:
  ```go
  // Assuming 'myFlow' is a *teleflow.FlowBuilder instance
  myFlow := teleflow.NewFlow("user_onboarding").
      // ... define steps (e.g., "ask_name", "ask_age") ...
      SetCompletionHandler(func(ctx *teleflow.Context, data map[string]interface{}) error {
          name, _ := data["name"].(string) // Add type assertion and error/existence check
          age, _ := data["age"].(int)     // Add type assertion and error/existence check
          log.Printf("Flow 'user_onboarding' completed for user %d. Name: %s, Age: %d", ctx.UserID(), name, age)
          return ctx.Reply("Thanks for completing the onboarding!")
      })
  // bot.RegisterFlow(myFlow) // Register the flow with the bot
  ```

### `teleflow.FlowCancellationHandlerFunc`

Signature: `func(ctx *teleflow.Context, flowData map[string]interface{}) error`

- **Purpose**: This handler is executed if the flow is explicitly cancelled by the user (e.g., via a global cancel command like `/cancel` if configured) or programmatically via `ctx.CancelFlow()`.
- **Parameters**:
    - `ctx`: The `teleflow.Context`.
    - `flowData`: A map containing any data collected up to the point of cancellation.
- **Example Usage with `FlowBuilder`**:
  ```go
  myFlow.SetCancellationHandler(func(ctx *teleflow.Context, data map[string]interface{}) error {
      log.Printf("Flow 'user_onboarding' cancelled by user %d. Collected data: %v", ctx.UserID(), data)
      return ctx.Reply("Your onboarding process has been cancelled.")
  })
  ```

These specific flow step handlers provide a structured way to manage complex, multi-step interactions within your Teleflow bot. Refer to the [Flow Guide](flow-guide.md) for more comprehensive examples of building flows.

## 5. Middleware

Middleware in Teleflow allows you to execute code before or after your main handlers. This is useful for cross-cutting concerns like logging, authentication, rate limiting, or modifying the context.

With the introduction of specific handler types, middleware has also become **type-specific**, allowing you to apply middleware selectively to different categories of handlers.

### Middleware Function Signatures

Teleflow defines specific middleware function types for each handler category:

- **`teleflow.CommandMiddlewareFunc`**:
  `func(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc`
  Wraps a `CommandHandlerFunc`.

- **`teleflow.TextMiddlewareFunc`**:
  `func(next teleflow.TextHandlerFunc) teleflow.TextHandlerFunc`
  Wraps a `TextHandlerFunc` for specific text matches.

- **`teleflow.DefaultTextMiddlewareFunc`**:
  `func(next teleflow.DefaultTextHandlerFunc) teleflow.DefaultTextHandlerFunc`
  Wraps the `DefaultTextHandlerFunc`.

- **`teleflow.CallbackMiddlewareFunc`**:
  `func(next teleflow.CallbackHandlerFunc) teleflow.CallbackHandlerFunc`
  Note: `teleflow.CallbackHandlerFunc` is the underlying function type `func(ctx *Context, fullCallbackData string, extractedData string) error` used by `SimpleCallback`. Middleware for `CallbackHandler` (the interface) applies to the `Handle` method's effective function.

### Registering Type-Specific Middleware

You register middleware using category-specific methods on the `Bot` object:

- `bot.UseCommandMiddleware(m ...teleflow.CommandMiddlewareFunc)`
- `bot.UseTextMiddleware(m ...teleflow.TextMiddlewareFunc)` (for specific text handlers registered with `bot.HandleText`)
- `bot.UseDefaultTextMiddleware(m ...teleflow.DefaultTextMiddlewareFunc)` (for the default text handler registered with `bot.SetDefaultTextHandler`)
- `bot.UseCallbackMiddleware(m ...teleflow.CallbackMiddlewareFunc)` (for callback handlers registered with `bot.RegisterCallback`)

Middleware functions are executed in the order they are added for a given category.

### Middleware Examples

#### Command Middleware Example (Logging)
```go
// import teleflow "github.com/kslamph/teleflow/core"
// import "log"

func CommandLoggerMiddleware(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc {
    return func(ctx *teleflow.Context, command string, args string) error {
        log.Printf("Command Middleware: Before /%s, Args: '%s', UserID: %d", command, args, ctx.UserID())
        err := next(ctx, command, args) // Call the next middleware or the actual handler
        if err != nil {
            log.Printf("Command Middleware: Error in handler for /%s: %v", command, err)
        }
        log.Printf("Command Middleware: After /%s", command)
        return err
    }
}

// bot.UseCommandMiddleware(CommandLoggerMiddleware) // In your bot setup
```

#### Text Middleware Example (Authentication Check)
```go
// import "log" // Assuming log is imported

func isAuthenticated(userID int64) bool {
    // Replace with your actual authentication logic
    // For example, check if userID exists in a database of authenticated users
    log.Printf("Checking authentication for UserID: %d", userID)
    return true // Placeholder
}

func TextAuthMiddleware(next teleflow.TextHandlerFunc) teleflow.TextHandlerFunc {
    return func(ctx *teleflow.Context, text string) error {
        if !isAuthenticated(ctx.UserID()) {
            return ctx.Reply("You need to be logged in to say that.")
        }
        log.Printf("Text Middleware: User %d authenticated for text: '%s'", ctx.UserID(), text)
        return next(ctx, text)
    }
}

// bot.UseTextMiddleware(TextAuthMiddleware)
// This middleware will apply to handlers registered with bot.HandleText("some specific text", ...)
```

#### Default Text Middleware Example
```go
func DefaultTextLoggerMiddleware(next teleflow.DefaultTextHandlerFunc) teleflow.DefaultTextHandlerFunc {
    return func(ctx *teleflow.Context, fullMessageText string) error {
        log.Printf("Default Text Middleware: Processing: '%s', UserID: %d", fullMessageText, ctx.UserID())
        return next(ctx, fullMessageText)
    }
}

// bot.UseDefaultTextMiddleware(DefaultTextLoggerMiddleware)
```

#### Callback Middleware Example (Data Preprocessing)
```go
// Assuming teleflow.CallbackHandlerFunc is:
// type CallbackHandlerFunc func(ctx *Context, fullCallbackData string, extractedData string) error
// import "strings" // For example

func CallbackDataEnhancerMiddleware(next teleflow.CallbackHandlerFunc) teleflow.CallbackHandlerFunc {
    return func(ctx *teleflow.Context, fullData string, extractedData string) error {
        // Example: Add something to the context based on callback data
        if strings.HasPrefix(fullData, "admin_") {
            ctx.Set("is_admin_callback", true)
            log.Printf("Callback Middleware: Admin callback detected for UserID: %d", ctx.UserID())
        }
        log.Printf("Callback Middleware: Before. Full: %s, Extracted: %s, UserID: %d", fullData, extractedData, ctx.UserID())
        err := next(ctx, fullData, extractedData)
        log.Printf("Callback Middleware: After. Full: %s", fullData)
        return err
    }
}

// bot.UseCallbackMiddleware(CallbackDataEnhancerMiddleware)
```

### Middleware for Flow Step Handlers

Currently, global middleware (Command, Text, Callback) might intercept inputs before they reach a flow if the input matches their criteria (e.g., a command message that is not a flow-specific command). Middleware specifically targeting individual `FlowStepStartHandlerFunc`, `FlowStepInputHandlerFunc`, etc., within a flow step is a potential area for future enhancement if more granular control is needed directly at the step level beyond what flow validators offer. For now, common logic within flow steps should be handled by shared functions or within the handlers/validators themselves.

For more general concepts on middleware, you can still refer to the [Middleware Guide](middleware-guide.md), but the registration methods and type-specific signatures described here are the current API.

## 6. Handler Execution Order

Teleflow processes incoming updates and resolves handlers in a specific order. Middleware (if registered for the corresponding type) is applied before its respective handler is invoked.

1.  **Global Flow Exit Commands**: If the user is in an active [Flow](flow-guide.md) and sends a command configured as a global exit command for flows (e.g., `/cancel` via `bot.SetGlobalFlowExitCommands()`), the flow is exited. The command handler for this exit command (if any) will then be processed according to rule #3.
2.  **Flow Handling**: If the user is in an active flow, the `FlowManager` attempts to handle the update (text or callback query). This involves checking flow-specific commands, then validators and input handlers for the current step. If the flow handles the update, no further non-flow handlers (command, text, callback) are typically processed for that update.
3.  **Command Handlers (`teleflow.CommandHandlerFunc`)**: If the update is a message that starts with `/` and is not handled by an active flow (or is a global flow exit command after flow exit), registered command handlers are checked. `CommandMiddlewareFunc` runs before the handler.
4.  **Specific Text Handlers (`teleflow.TextHandlerFunc`)**: If the update is a text message not handled by a flow or as a command, specific text handlers (registered with `bot.HandleText(textToMatch, ...)`) are checked. `TextMiddlewareFunc` runs before the handler.
5.  **Default Text Handler (`teleflow.DefaultTextHandlerFunc`)**: If the update is a text message and no flow, command, or specific text handler matches, the default text handler (registered with `bot.SetDefaultTextHandler(...)`) is executed. `DefaultTextMiddlewareFunc` runs before the handler.
6.  **Callback Handlers (`teleflow.CallbackHandler` interface)**: If the update is a `CallbackQuery` (from an inline button press) and not handled by an active flow, registered callback handlers are checked based on their patterns. `CallbackMiddlewareFunc` runs before the handler's `Handle` method (or the function wrapped by `SimpleCallback`).

It's important to understand this order to predict how your bot will respond to different types of user input, especially when using a combination of flows, commands, and text handlers.

## 7. Best Practices for Handlers

- **Utilize Typed Parameters**: Leverage the specific handler signatures (e.g., `CommandHandlerFunc`, `TextHandlerFunc`) that provide input data (like command arguments or message text) directly as typed parameters. This improves code clarity and type safety.
- **Keep Handlers Concise**: Aim for handlers that perform a specific, well-defined task. For complex logic, break it down into smaller helper functions or consider using [Flows](flow-guide.md) for multi-step interactions.
- **Leverage `teleflow.Context`**: While primary input data often comes via direct parameters, the `ctx` object remains essential for sending replies, managing state, accessing bot/user/chat info, and other interactions.
- **Error Handling**: Always handle errors returned by `Context` methods (e.g., `ctx.Reply`) and other operations. Return errors from your handler if something goes wrong; Teleflow's core loop or your middleware can then log them.
- **Type-Specific Middleware**: When using middleware, choose the correct registration method (e.g., `bot.UseCommandMiddleware`, `bot.UseTextMiddleware`) corresponding to the handler category you want to target.
- **Idempotency for Callbacks**: Design `CallbackHandler`s to be idempotent where possible, as Telegram might occasionally send duplicate callback queries. This means the handler can be safely executed multiple times with the same data without unintended side effects.
- **Answer Callback Queries**: For `CallbackHandler`s, always ensure the callback query is answered (e.g., via `ctx.Bot.AnswerCallbackQuery`, `ctx.EditOrReply`, or `ctx.ReplyWithMarkup`). This provides feedback to the user (removes the loading spinner on the button) and acknowledges the interaction to Telegram.
- **Clear Flow Logic**: When using Flow Step Handlers, ensure each handler (`Start`, `Validator`, `Input`, `Completion`, `Cancellation`) has a clear purpose. Use validators effectively to guide users and ensure data quality before it reaches your input handlers. Access validated input via `ctx.Get(teleflow.ValidatedInputKey)`.

## 8. Next Steps

With a solid understanding of Teleflow's typed handlers and middleware system, you're ready to explore more features:
- [Keyboards Guide](keyboards-guide.md): Create interactive reply and inline keyboards.
- [Flow Guide](flow-guide.md): Build complex, multi-step conversations using flows and their specific step handlers.
- [Middleware Guide](middleware-guide.md): For more general concepts on middleware, though this guide covers the current type-specific API.
- [Templates Guide](templates-guide.md): Craft dynamic messages with ease.
- [State Management Guide](state-management-guide.md): Learn how to manage user and chat state.
- [API Reference](api-reference.md): For a detailed look at all available functions and types.