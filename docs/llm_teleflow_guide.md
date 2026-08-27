# Teleflow Package Guide for LLMs

This document explains how to use the Teleflow Go package to build Telegram bots. It's designed to help an LLM understand the core concepts and common usage patterns for accomplishing coding tasks with Teleflow.

## 🎯 Core Goal of Teleflow

Teleflow is a Go framework for building Telegram bots. Its main goals are to:

1. **Simplify Conversational Flows**: Provide a structured way to define multi-step interactions with users.
2. **Automate State Management**: Handle user session data and flow progress automatically.
3. **Enable Type-Safe Interactions**: Use Go generics for compile-time safety.
4. **Offer Clean API**: Simple, consistent interface for all common bot operations.

## 🧩 Key Components

### 1. `Bot`

* **Purpose**: The central object representing your Telegram bot. It manages connections and routes updates to handlers.
* **Initialization**:

    ```go
    import (
        "log"
        "os"

        teleflow "github.com/kslamph/teleflow/core"
    )

    // Set up the router with your handlers
    router := teleflow.NewRouter()
    router.Handle("start", teleflow.Func(func(ctx *teleflow.Context) error {
        return ctx.ReplyText("Welcome!")
    }))
    // Register more handlers or flows as needed...

    // Create the bot with token and router
    bot, err := teleflow.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"), router)
    if err != nil {
        log.Fatal(err)
    }
    ```

* **Starting the Bot**:

    ```go
    log.Fatal(bot.Start()) // Starts the long polling loop to receive updates.
    ```

### 2. `Context` (`teleflow.Context`)

* **Purpose**: Represents the current interaction state with a user. It's passed to all handlers and flow processing functions.
* **Key Functions**:
  * `ctx.Text()`: Get the text of a message or the data of a callback query.
  * `ctx.UserID()`: Get the ID of the user.
  * `ctx.ChatID()`: Get the ID of the chat.
  * `ctx.ReplyText(message string)`: Send a simple text message.
  * `ctx.ReplyWithKeyboard(text string, markup interface{})`: Send a message with a keyboard.
  * `ctx.ReplyTextWithParseMode(text string, parseMode ParseMode)`: Send a formatted text message.
  * `ctx.ReplyWithKeyboardAndParseMode(text string, markup interface{}, parseMode ParseMode)`: Send a formatted message with a keyboard.
  * `ctx.EditMessageText(messageID int, text string)`: Edit an existing message.
  * `ctx.EditMessageTextWithParseMode(messageID int, text string, parseMode ParseMode)`: Edit an existing message with formatting.
  * `ctx.EditMessageReplyMarkup(messageID int, markup tgbotapi.InlineKeyboardMarkup)`: Edit the keyboard of a message.
  * `ctx.EditMessageMedia(messageID int, media interface{})`: Edit the media (e.g. photo) of a message.
  * `ctx.DeleteMessage(messageID int)`: Delete a message.
  * `ctx.AnswerCallback()`: Acknowledge a callback query; must be called from your handler to stop the client's loading spinner (the framework never auto-answers).
  * `ctx.RegisterCallback(handler)`: Register a one-time handler for a callback query; auto-removed after execution.
  * `ctx.RegisterData(data)`: Store complex data under a unique ID for callback buttons (bypasses the 64-byte callback limit).
  * `ctx.GetData()`: Retrieve and delete data stored by `RegisterData`.
  * `ctx.Set(key, value)` / `ctx.Get(key)`: Store and retrieve request-scoped data.
  * `ctx.Update()`: Access the raw `tgbotapi.Update` object.
  * `ctx.API()`: Access the underlying `tgbotapi.BotAPI` instance.

* **Parse Mode Constants**: Teleflow provides constants for message formatting:
  * `teleflow.ParseModeHTML`: HTML formatting
  * `teleflow.ParseModeMarkdown`: Markdown formatting
  * `teleflow.ParseModeMarkdownV2`: MarkdownV2 formatting (recommended)

### 3. Handlers

* **Purpose**: Functions that process incoming messages or commands.
* **Using Router**:

    ```go
    router := teleflow.NewRouter()

    // Register a simple handler
    router.Handle("start", teleflow.Func(func(ctx *teleflow.Context) error {
        return ctx.ReplyText("Welcome!")
    }))

    // Register a flow
    router.Handle("register", registrationFlow)
    ```

* **Default Handler**:

    ```go
    router.Default(teleflow.Func(func(ctx *teleflow.Context) error {
        return ctx.ReplyText("Unknown command. Use /start to begin.")
    }))
    ```

* **Important — Inline-Button Routing**: Inline-keyboard button presses arrive as `CallbackQuery` updates, **not** commands. `Router.HandleUpdate` only routes `msg.IsCommand()` messages. To route button presses to a `Flow`, use the `router.Default` forwarding pattern:

    ```go
    router.Default(teleflow.Func(func(ctx *teleflow.Context) error {
        if flow.IsUserInFlow(ctx.ChatID(), ctx.UserID()) {
            return flow.HandleUpdate(ctx)
        }
        return ctx.ReplyText("Unknown command. Use /start.")
    }))
    ```

    Alternatively, register a one-time callback handler via `ctx.RegisterCallback(handler)`, which auto-removes after execution.

    **Important — callbacks are never auto-answered.** The framework does not call `AnswerCallback` for you. Your handler (or the flow `Process` that handles the button press) must call `ctx.AnswerCallback()` to stop the Telegram client's loading spinner and acknowledge the press. A stale button whose handler/data has already been consumed will silently fall through to the router, so always call `AnswerCallback()` before doing any work.

### 4. Flows (`teleflow.Flow[T]`)

* **Purpose**: Define multi-step conversational interactions with automatic state management.
* **Building a Flow**:

    ```go
    // 1. Define a struct for your flow's data.
    type RegistrationData struct {
        Name string
        Age  int
    }

    // 2. Create a new Flow with your data type.
    registrationFlow := teleflow.NewFlow[RegistrationData]()

    // 3. Define steps with clean, type-safe functions.
    //    (Requires imports: "fmt", "strconv", "log", teleflow "github.com/kslamph/teleflow/core")
    welcomeStep := teleflow.Step[RegistrationData]{
        Name: "Welcome",
        Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
            return ctx.ReplyText("Welcome! What is your name?")
        },
        Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
            data.Name = ctx.Text()
            return teleflow.Next(), nil
        },
    }

    ageStep := teleflow.Step[RegistrationData]{
        Name: "Age",
        Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
            return ctx.ReplyText(fmt.Sprintf("Nice to meet you, %s! How old are you?", data.Name))
        },
        Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
            age, err := strconv.Atoi(ctx.Text())
            if err != nil {
                ctx.ReplyText("Please enter a valid number.")
                return teleflow.Retry(), nil
            }
            data.Age = age
            return teleflow.End(), nil
        },
    }

    // 4. Add steps to the flow.
    if _, err := registrationFlow.AddStep(welcomeStep); err != nil { log.Fatal(err) }
    if _, err := registrationFlow.AddStep(ageStep); err != nil { log.Fatal(err) }

    // 5. Set a completion handler.
    registrationFlow.OnComplete(func(ctx *teleflow.Context, data RegistrationData) error {
        // The final `data` is type-safe.
        return ctx.ReplyText(fmt.Sprintf("Thanks, %s! Registration complete. Age: %d", data.Name, data.Age))
    })
    ```

* **Flow Control Functions**:
  * `teleflow.Next()`: Move to the next step in sequence.
  * `teleflow.Retry()`: Re-prompt the current step.
  * `teleflow.End()`: Successfully end the flow and trigger `OnComplete`.
  * `teleflow.GoTo("step_name")`: Jump to a specific named step.

### 5. Keyboards

* **Inline Keyboards (`teleflow.InlineKeyboard`)**: Attached to messages.

    ```go
    keyboard := teleflow.NewInlineKeyboard().AddRow(
        teleflow.ButtonCallback("Option 1", "opt1"),
        teleflow.ButtonCallback("Option 2", "opt2"),
    ).Build()

    ctx.ReplyWithKeyboard("Choose an option:", keyboard)
    ```

* **Complex Callback Data**:

    ```go
    // Register complex data for callbacks
    complexData := MyStruct{Field: "value"}
    dataID := ctx.RegisterData(complexData)

    keyboard := teleflow.NewInlineKeyboard().AddRow(
        teleflow.ButtonCallback("Click me", dataID),
    ).Build()

    // Later, retrieve the data
    retrievedData, ok := ctx.GetData()
    if ok {
        myData := retrievedData.(MyStruct)
        // Use myData
    }
    ```

### 6. Middleware

* **Purpose**: Intercept and process updates before they reach handlers.
* **Structure**:

    ```go
    func LoggingMiddleware(next teleflow.Handler) teleflow.Handler {
        return teleflow.Func(func(ctx *teleflow.Context) error {
            // Code before handler
            log.Printf("User %d accessing: %s", ctx.UserID(), ctx.Text())

            err := next.HandleUpdate(ctx) // Call the next handler

            // Code after handler
            return err
        })
    }
    ```

* **Applying Middleware**:

    ```go
    // Apply middleware when registering handlers
    router.Handle("admin", LoggingMiddleware(adminFlow))
    ```

## ✨ Key Teleflow Features for LLMs

1. **Type-Safe Flows**:
    * Use Go generics (`teleflow.Flow[T]`) for compile-time safety.
    * No more type assertions for flow data.

2. **Automatic State Management**:
    * Teleflow automatically manages user state within a flow.
    * Each user has their own isolated data instance.

3. **Non-Linear Conversations**:
    * Named steps and `GoTo` actions enable complex flow control.
    * Easy to create loops, branches, and restartable flows.

4. **Large Callback Data**:
    * Overcome Telegram's 64-byte callback data limit with `ctx.RegisterData()` and `ctx.GetData()`.

## 📝 Common Tasks for LLMs using Teleflow

* **Creating a new command handler**:
  * Use `router.Handle("command", teleflow.Func(...))`
  * Inside the handler, use `ctx.ReplyText()` or `ctx.ReplyWithKeyboard()`.

* **Defining a new conversational flow**:
  * Define a data struct `T`.
  * Create a flow with `teleflow.NewFlow[T]()`.
  * Define `teleflow.Step[T]` with `Prompt` and `Process` functions.
  * Add steps with `flow.AddStep()`.
  * Register with `router.Handle("command", flow)`.

* **Asking a question and getting a response in a flow**:
  * Define a `Step` with a `Prompt` function that sends a message.
  * In the `Process` function, use `ctx.Text()` to get the user's response.

* **Adding buttons to a message**:
  * In a `Prompt` or handler, create a keyboard with `teleflow.NewInlineKeyboard()`.
  * Add buttons with `ButtonCallback()`.
  * Send with `ctx.ReplyWithKeyboard()`.

* **Storing and retrieving data during a flow**:
  * Use the `data *T` parameter in `Prompt` and `Process` functions.
  * Modify the struct fields directly.

* **Handling complex callback data**:
  * Use `ctx.RegisterData(complexData)` to store data.
  * Use `ctx.GetData()` to retrieve data in the next step.

This guide should provide a solid foundation for an LLM to understand and generate Go code using the Teleflow package. Refer to the `README.md` and specific Go files in the `core` directory for more detailed examples and advanced features.
