# Teleflow API Reference

This document provides a reference for the Teleflow Go package API.

**Note:** This is a foundational API reference. For a complete and detailed API documentation, it is recommended to use `godoc` or a similar Go documentation tool to explore the `core` package.

Example:
```bash
godoc -http=:6060
```
Then navigate to `http://localhost:6060/pkg/github.com/kslamph/teleflow/core/` in your browser.

## Core Package: `teleflow`

The `core` package contains all the fundamental types and functions for building Telegram bots with Teleflow.

### Key Types

*   **`Bot`**: The main struct representing your Telegram bot.
    *   `NewBot(token string, options ...BotOption) (*Bot, error)`: Creates a new bot instance.
    *   `Start() error`: Starts the bot and begins processing updates.
    *   `HandleCommand(command string, handler HandlerFunc)`: Registers a handler for a slash command.
    *   `HandleText(text string, handler HandlerFunc)`: Registers a handler for text messages.
    *   `RegisterCallback(handler CallbackHandler)`: Registers a handler for inline keyboard callbacks.
    *   `RegisterFlow(flow *Flow)`: Registers a conversational flow.
    *   `Use(middleware MiddlewareFunc)`: Adds middleware to the bot's processing pipeline.
    *   `WithMenuButton(config *MenuButtonConfig)`: Sets the bot's menu button.
    *   `WithFlowConfig(config FlowConfig)`: Configures global flow behavior.
    *   `WithAccessManager(accessManager AccessManager)`: Sets a custom access manager.
    *   `AddTemplate(name, templateText string, parseMode ParseMode) error`: Adds a message template.
    *   See [`core/bot.go`](../core/bot.go) for all methods.

*   **`Context`**: Passed to all handlers, providing information about the current interaction and helper methods.
    *   `UserID() int64`: Returns the ID of the user.
    *   `ChatID() int64`: Returns the ID of the chat.
    *   `Reply(text string, keyboard ...interface{}) error`: Sends a reply message.
    *   `ReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error`: Sends a reply using a template.
    *   `EditOrReply(text string, keyboard ...interface{}) error`: Edits the current message or sends a new one.
    *   `StartFlow(flowName string) error`: Starts a conversational flow.
    *   `Set(key string, value interface{})`: Stores data in the context (request-scoped).
    *   `Get(key string) (interface{}, bool)`: Retrieves data from the context.
    *   See [`core/context.go`](../core/context.go) for all methods.

*   **`HandlerFunc`**: `func(ctx *Context) error` - The signature for command and text handlers.

*   **`MiddlewareFunc`**: `func(next HandlerFunc) HandlerFunc` - The signature for middleware.

*   **`Flow`**: Represents a multi-step conversation.
    *   `NewFlow(name string) *FlowBuilder`: Starts building a new flow.
    *   See [`core/flow.go`](../core/flow.go) for `FlowBuilder` methods.

*   **`ReplyKeyboard`**: For creating reply keyboards.
    *   `NewReplyKeyboard() *ReplyKeyboard`
    *   `AddButton(text string) *ReplyKeyboard`
    *   `AddRow() *ReplyKeyboard`
    *   See [`core/keyboards.go`](../core/keyboards.go) for all methods.

*   **`InlineKeyboard`**: For creating inline keyboards.
    *   `NewInlineKeyboard() *InlineKeyboard`
    *   `AddButton(text, callbackData string) *InlineKeyboard`
    *   `AddURL(text, url string) *InlineKeyboard`
    *   `AddRow() *InlineKeyboard`
    *   See [`core/keyboards.go`](../core/keyboards.go) for all methods.

*   **`MenuButtonConfig`**: For configuring the bot's menu button.
    *   `NewCommandsMenuButton() *MenuButtonConfig`
    *   `NewWebAppMenuButton(text, url string) *MenuButtonConfig`
    *   `AddCommand(text, command string) *MenuButtonConfig`
    *   See [`core/menu_button.go`](../core/menu_button.go) and [`core/keyboards.go`](../core/keyboards.go) for related types.

*   **`StateManager`**: Interface for state persistence.
    *   `NewInMemoryStateManager() StateManager`
    *   See [`core/state.go`](../core/state.go).

*   **`AccessManager`**: Interface for permission checking and dynamic UI.
    *   See [`core/bot.go`](../core/bot.go).

### Helper Functions and Constants

The `core` package also includes various helper functions (e.g., for creating specific middleware, validators) and constants (e.g., `ParseModeMarkdownV2`, `StepTypeText`). These are best explored using `godoc` or by reading the respective source files.

## Further Exploration

For detailed information on each type, method, and function, please refer to:
- The individual source files in the `core/` directory.
- The auto-generated documentation using `godoc`.
- The specific feature guides (e.g., [Flow Guide](flow-guide.md), [Keyboards Guide](keyboards-guide.md)).