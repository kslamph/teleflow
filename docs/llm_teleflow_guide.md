# Teleflow Package Guide for LLMs

This document explains how to use the Teleflow Go package to build Telegram bots. It's designed to help an LLM understand the core concepts and common usage patterns for accomplishing coding tasks with Teleflow.

## 🎯 Core Goal of Teleflow

Teleflow is a Go framework for building enterprise-grade Telegram bots. Its main goals are to:

1.  **Simplify Conversational Flows**: Provide a structured way to define multi-step interactions with users.
2.  **Automate State Management**: Handle user session data and flow progress automatically.
3.  **Enable Type-Safe Interactions**: Allow complex data structures in callbacks, abstracting Telegram's string limitations.
4.  **Offer Enterprise Features**: Include built-in middleware for common needs like rate limiting and authentication.

## 🧩 Key Components

### 1. `Bot`

*   **Purpose**: The central object representing your Telegram bot. It manages connections, handlers, flows, and middleware.
*   **Initialization**:
    ```go
    import teleflow "github.com/kslamph/teleflow/core"
    import "os"
    import "log"

    // Basic initialization
    bot, err := teleflow.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }
    ```
*   **Configuration Options (`teleflow.BotOption`)**:
    *   `teleflow.WithFlowConfig()`: Customize behavior of flows (e.g., exit commands, global command handling).
    *   `teleflow.WithAccessManager()`: Integrate custom permission logic and automatically apply security middleware.
*   **Starting the Bot**:
    ```go
    bot.Start() // This starts the long polling loop to receive updates.
    ```

### 2. `Context` (`teleflow.Context`)

*   **Purpose**: Represents the current interaction state with a user. It's passed to all handlers and flow processing functions.
*   **Key Functions**:
    *   `ctx.UserID()`: Get the ID of the user.
    *   `ctx.ChatID()`: Get the ID of the chat.
    *   `ctx.SendPromptText(message string)`: Send a simple text message.
    *   `ctx.SendPrompt(config *teleflow.PromptConfig)`: Send a rich message with text, image, keyboard, and template data.
    *   `ctx.StartFlow(flowName string)`: Initiate a conversational flow for the user.
    *   `ctx.SetFlowData(key string, value interface{})`: Store data within the current flow's session for the user.
    *   `ctx.GetFlowData(key string) (interface{}, bool)`: Retrieve data stored in the flow session.
    *   `ctx.Update()`: Access the raw `tgbotapi.Update` object.

### 3. Handlers

*   **Purpose**: Functions that process incoming messages or commands when a user is *not* in an active flow.
*   **Types**:
    *   `bot.HandleCommand(commandName string, handler teleflow.CommandHandlerFunc)`:
        ```go
        // Example: /start command
        bot.HandleCommand("start", func(ctx *teleflow.Context, command string, args string) error {
            return ctx.SendPromptText("Welcome!")
        })
        ```
    *   `bot.HandleText(textToMatch string, handler teleflow.TextHandlerFunc)`: Handles exact text matches.
        ```go
        bot.HandleText("hello", func(ctx *teleflow.Context, text string) error {
            return ctx.SendPromptText("Hi there!")
        })
        ```
    *   `bot.DefaultHandler(handler teleflow.DefaultHandlerFunc)`: A fallback handler if no specific command or text handler matches.

### 4. Flows (`teleflow.Flow`)

*   **Purpose**: Define multi-step conversational interactions. Teleflow manages the state and transitions between steps.
*   **Building a Flow**:
    ```go
    registrationFlow, err := teleflow.NewFlow("user_registration"). // Unique name for the flow
        Step("ask_name"). // Define a step
        Prompt("What's your name?"). // Message to send when this step starts
        Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
            // 'input' is the user's text response.
            // 'buttonClick' is non-nil if the user clicked an inline keyboard button.
            if input == "" {
                return teleflow.Retry().WithPrompt("Please provide a name.") // Ask the same step again
            }
            ctx.SetFlowData("userName", input) // Store data for this flow instance
            return teleflow.NextStep()         // Proceed to the next defined step
        }).

        Step("ask_email").
        Prompt(func(ctx *teleflow.Context) string { // Prompt can be a dynamic function
            name, _ := ctx.GetFlowData("userName")
            return fmt.Sprintf("Nice to meet you, %s! What's your email?", name)
        }).
        Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
            if !isValidEmail(input) { // Assume isValidEmail is defined elsewhere
                return teleflow.Retry().WithPrompt("Invalid email. Please try again.")
            }
            ctx.SetFlowData("userEmail", input)
            return teleflow.CompleteFlow() // End the flow successfully
        }).

        OnComplete(func(ctx *teleflow.Context) error { // Executed after CompleteFlow()
            name, _ := ctx.GetFlowData("userName")
            email, _ := ctx.GetFlowData("userEmail")
            // Send a summary or save data
            return ctx.SendPromptText(fmt.Sprintf("Registered: %s, %s", name, email))
        }).
        Build()

    if err != nil {
        log.Fatal(err)
    }
    ```
*   **Registering a Flow**:
    ```go
    bot.RegisterFlow(registrationFlow)
    ```
*   **Starting a Flow (e.g., from a command handler)**:
    ```go
    bot.HandleCommand("register", func(ctx *teleflow.Context, command string, args string) error {
        return ctx.StartFlow("user_registration") // Use the flow's unique name
    })
    ```
*   **Flow Control in `Process` function**:
    *   `teleflow.NextStep()`: Move to the next step in sequence.
    *   `teleflow.GoToStep(stepName string)`: Jump to a specific step.
    *   `teleflow.Retry()`: Re-prompt the current step. Can be chained with `.WithPrompt()` for a custom retry message.
    *   `teleflow.CompleteFlow()`: Successfully end the flow and trigger `OnComplete`.
    *   `teleflow.CancelFlow()`: Abort the flow.

### 5. Prompts (`teleflow.PromptConfig`)

*   **Purpose**: Define the message to be sent to the user. Can include text, images, keyboards, and use templates.
*   **Sending a Rich Prompt**:
    ```go
    // Assume imageBytes is []byte content of an image
    return ctx.SendPrompt(&teleflow.PromptConfig{
        Message:      "template:welcome_message", // Can be literal text or "template:template_name"
        Image:        imageBytes,                 // Optional image
        TemplateData: map[string]interface{}{    // Data for the template
            "UserName": "John Doe",
        },
        Keyboard: teleflow.NewPromptKeyboard(). // Optional inline keyboard
            ButtonCallback("Option 1", "callback_data_1").
            ButtonURL("Visit Site", "https://example.com").
            Build(),
        ParseMode: teleflow.ParseModeMarkdownV2, // Optional, defaults based on template or global config
    })
    ```

### 6. Keyboards

*   **Inline Keyboards (`teleflow.PromptKeyboardBuilder`)**: Attached to messages.
    ```go
    keyboard := teleflow.NewPromptKeyboard().
        ButtonCallback("✅ Approve", approveCallbackData). // approveCallbackData can be a struct or map
        ButtonCallback("❌ Reject", "reject_action").
        Row(). // Start a new row of buttons
        ButtonURL("More Info", "https://example.com/info").
        Build() // Returns *tgbotapi.InlineKeyboardMarkup

    // Use in PromptConfig:
    // Keyboard: keyboard,
    ```
    *   `ButtonCallback(text string, data interface{})`: `data` can be a simple string, or a struct/map. Teleflow handles serialization and deserialization. The `data` is available in `buttonClick.Data` in the `Process` function.
*   **Reply Keyboards (`teleflow.ReplyKeyboard`)**: Replace the user's standard keyboard. Often used for main menus via `AccessManager`.

### 7. Templates

*   **Purpose**: Define reusable message formats with placeholders. Supports Go's `text/template` syntax.
*   **Registration**:
    ```go
    teleflow.AddTemplate("welcome_message", "Hello {{.UserName}}! Welcome to our bot.", teleflow.ParseModeMarkdownV2)
    ```
*   **Usage**:
    *   In `PromptConfig.Message`: `"template:welcome_message"`
    *   With `PromptConfig.TemplateData`: `map[string]interface{}{"UserName": "Jane"}`
    *   Directly: `ctx.SendPromptWithTemplate("template_name", dataMap)`

### 8. Middleware (`teleflow.MiddlewareFunc`)

*   **Purpose**: Intercept and process updates before they reach handlers. Useful for logging, authentication, rate limiting, etc.
*   **Structure**:
    ```go
    func MyMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            // Code before handler
            log.Printf("User %d accessing: %s", ctx.UserID(), ctx.Update().Message.Text)

            err := next(ctx) // Call the next middleware or the actual handler

            // Code after handler
            return err
        }
    }
    ```
*   **Applying Middleware**:
    ```go
    bot.UseMiddleware(MyMiddleware)
    bot.UseMiddleware(teleflow.RateLimitMiddleware(10, time.Minute*1)) // Built-in
    ```
    If an `AccessManager` is provided via `WithAccessManager`, an `AuthMiddleware` is automatically added.

## ✨ Key Teleflow Features for LLMs

1.  **Automatic State Management**:
    *   LLMs don't need to manually track which step a user is in a flow. Teleflow does this.
    *   `ctx.SetFlowData()` and `ctx.GetFlowData()` provide a simple key-value store scoped to the user's current flow session.

2.  **Type-Safe Callbacks**:
    *   When creating inline keyboard buttons with `ButtonCallback("Text", data)`, `data` can be a Go struct or map.
    *   In the flow's `Process` function, `buttonClick.Data` will be the deserialized struct/map. This avoids manual JSON parsing of callback strings.
    *   Example:
        ```go
        type UserAction struct {
            Action string `json:"action"`
            ItemID int    `json:"item_id"`
        }
        // ... in keyboard setup
        ButtonCallback("Delete Item 5", UserAction{Action: "delete", ItemID: 5})
        // ... in Process function
        if buttonClick != nil {
            if actionData, ok := buttonClick.Data.(UserAction); ok {
                // actionData is now a UserAction struct
                if actionData.Action == "delete" { /* ... */ }
            }
        }
        ```

3.  **Fluent Flow Definition**:
    *   The chained `.Step().Prompt().Process()` syntax makes defining conversational logic intuitive.

4.  **Clear Separation of Concerns**:
    *   Command/text handlers for general interactions.
    *   Flows for structured conversations.
    *   Middleware for cross-cutting concerns.

## 📝 Common Tasks for LLMs using Teleflow

*   **Creating a new command handler**:
    *   Use `bot.HandleCommand("mycommand", ...)`
    *   Inside the handler, use `ctx.SendPromptText()` or `ctx.SendPrompt()` or `ctx.StartFlow()`.

*   **Defining a new conversational flow**:
    *   Use `teleflow.NewFlow("flow_name").Step(...).Process(...).Build()`.
    *   Register it with `bot.RegisterFlow(myFlow)`.
    *   Start it using `ctx.StartFlow("flow_name")`.

*   **Asking a question and getting a response in a flow**:
    *   Define a `Step()`.
    *   Use `.Prompt("Your question?")`.
    *   In `.Process(func(ctx *teleflow.Context, input string, ...))`, the `input` variable will contain the user's text reply.

*   **Adding buttons to a message in a flow**:
    *   In a `Step()`, use `.WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder { ... })`.
    *   Build the keyboard using `teleflow.NewPromptKeyboard().ButtonCallback(...).Build()`.
    *   In `.Process(func(ctx *teleflow.Context, ..., buttonClick *teleflow.ButtonClick))`, check if `buttonClick != nil` and access `buttonClick.Data`.

*   **Storing and retrieving data during a flow**:
    *   `ctx.SetFlowData("myKey", myValue)`
    *   `value, ok := ctx.GetFlowData("myKey")`

*   **Using templates for messages**:
    *   `teleflow.AddTemplate("my_template", "Content with {{.Placeholder}}", teleflow.ParseModeMarkdownV2)`
    *   In `PromptConfig`, set `Message: "template:my_template"` and provide `TemplateData`.

This guide should provide a solid foundation for an LLM to understand and generate Go code using the Teleflow package. Refer to the `README.md` and specific Go files in the `core` directory for more detailed examples and advanced features.