# Teleflow Core API Guide

## Introduction

This guide provides a comprehensive overview of the `teleflow/core` API, organized by functionality. The teleflow framework offers a powerful, type-safe Go framework for building sophisticated Telegram bots with intuitive APIs and advanced flow management.

For detailed implementation specifics, refer to the [GoDoc documentation](https://pkg.go.dev/github.com/kslamph/teleflow/core). This guide serves as a curated, high-level entry point to help you understand the overall architecture and find the right APIs for your needs.

## 1. Bot Initialization & Configuration

### Core Bot Creation

**[`NewBot(token string, options ...BotOption) (*Bot, error)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#NewBot)**
- Creates a new Bot instance with Telegram bot token
- Accepts variadic `BotOption` functions for configuration
- Sets up default flow configuration and initializes all core components

**[`BotOption`](https://pkg.go.dev/github.com/kslamph/teleflow/core#BotOption)**
- Functional option type for Bot configuration during initialization
- Enables flexible, chainable bot setup

### Configuration Options

**[`WithFlowConfig(config FlowConfig) BotOption`](https://pkg.go.dev/github.com/kslamph/teleflow/core#WithFlowConfig)**
- Customizes flow behavior (exit commands, global commands, message processing)
- Controls conversational flow lifecycle and user experience

**[`WithAccessManager(accessManager AccessManager) BotOption`](https://pkg.go.dev/github.com/kslamph/teleflow/core#WithAccessManager)**
- Sets up permission checking and automatic UI management
- Automatically applies optimal middleware stack (rate limiting + authentication)

**[`WithMenuButton(config *MenuButtonConfig) BotOption`](https://pkg.go.dev/github.com/kslamph/teleflow/core#WithMenuButton)**
- Configures the persistent menu button next to text input field
- Supports web_app and default types

### Bot Management

**[`Bot.Start() error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.Start)**
- Starts the bot's update processing loop
- Begins polling for and handling Telegram updates

**[`Bot.SetBotCommands(commands map[string]string) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.SetBotCommands)**
- Sets bot commands that appear in Telegram's command menu
- Configures the "/" command interface for users

## 2. Handling Updates

### Command Handlers

**[`CommandHandlerFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#CommandHandlerFunc)**
- Function signature for command handlers: `func(ctx *Context) error`
- Command and arguments are accessible through the Context object

**[`Bot.HandleCommand(commandName string, handler CommandHandlerFunc)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.HandleCommand)**
- Registers handlers for specific commands (without leading slash)
- Automatically applies all registered middleware

### Text Message Handlers

**[`TextHandlerFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#TextHandlerFunc)**
- Function signature for exact text matching: `func(ctx *Context, text string) error`
- Case-sensitive exact string matching

**[`Bot.HandleText(textToMatch string, handler TextHandlerFunc)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.HandleText)**
- Registers handlers for exact text matches
- Useful for menu button responses and keywords

**[`DefaultHandlerFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#DefaultHandlerFunc)**
- Function signature for fallback text handler: `func(ctx *Context, text string) error`
- Handles any text that doesn't match other handlers

**[`Bot.DefaultHandler(handler DefaultHandlerFunc)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.DefaultHandler)**
- Registers fallback handler for unmatched text messages
- Only one default handler can be registered

### Core Handler Types

**[`HandlerFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#HandlerFunc)**
- Basic handler signature: `func(ctx *Context) error`
- Used internally by middleware system and as common interface

## 3. The Context Object

### Core Context Operations

**[`Context`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context)**
- Central object for all bot operations and state management
- Provides access to user info, chat details, and bot capabilities

**[`Context.UserID() int64`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.UserID)**
- Returns Telegram user ID for the current update

**[`Context.ChatID() int64`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.ChatID)**
- Returns chat ID where the update originated

**[`Context.IsGroup() bool`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.IsGroup)**
- Checks if current chat is a group or supergroup

**[`Context.IsChannel() bool`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.IsChannel)**
- Checks if current chat is a channel

### Request-Scoped Data

**[`Context.Set(key string, value interface{})`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.Set)**
- Stores temporary data for current update processing
- Data exists only for the duration of current request

**[`Context.Get(key string) (interface{}, bool)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.Get)**
- Retrieves temporary data stored during current update
- Returns value and existence flag

### Flow-Persistent Data

**[`Context.SetFlowData(key string, value interface{}) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.SetFlowData)**
- Stores data that persists across flow steps
- Only available when user is in an active flow

**[`Context.GetFlowData(key string) (interface{}, bool)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.GetFlowData)**
- Retrieves persistent flow data
- Returns nil if user not in flow or key doesn't exist

### Response Methods

**[`Context.Reply(text string, keyboardMarkup ...interface{}) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.Reply)**
- Sends text message with optional keyboard
- Automatically applies AccessManager keyboards if configured

**[`Context.ReplyWithParseMode(text string, parseMode ParseMode, keyboardMarkup ...interface{}) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.ReplyWithParseMode)**
- Sends formatted message with specific parse mode (Markdown, MarkdownV2, HTML)

**[`Context.ReplyTemplate(templateName string, data map[string]interface{}, keyboard ...interface{}) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.ReplyTemplate)**
- Renders and sends message using template system
- Supports optional keyboard attachment

**[`Context.SendPrompt(prompt *PromptConfig) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.SendPrompt)**
- Sends informational messages using prompt composition
- Supports templates and images without keyboards

**[`Context.SendPromptWithTemplate(templateName string, data map[string]interface{}) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.SendPromptWithTemplate)**
- Convenience method for template-based informational messages

## 4. Conversation Flows (Step-Prompt-Process API)

### Flow Creation

**[`NewFlow(name string) *FlowBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#NewFlow)**
- Creates new flow builder with specified name
- Entry point for fluent flow construction

**[`FlowBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder)**
- Fluent API for constructing flows with steps and lifecycle management
- Accumulates configuration before building final Flow

### Step Definition

**[`FlowBuilder.Step(name string) *StepBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder.Step)**
- Adds new step to flow with unique name
- Returns StepBuilder for step configuration

**[`StepBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#StepBuilder)**
- Builder for individual step configuration
- Connects prompts with processing logic

### Prompt Configuration

**[`StepBuilder.Prompt(message MessageSpec) *PromptBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#StepBuilder.Prompt)**
- Starts prompt configuration with message content
- Supports static strings, functions, and template references

**[`PromptConfig`](https://pkg.go.dev/github.com/kslamph/teleflow/core#PromptConfig)**
- Declarative prompt specification with message, image, keyboard, and template data
- Processed by PromptComposer during step rendering

**[`MessageSpec`](https://pkg.go.dev/github.com/kslamph/teleflow/core#MessageSpec)**
- Flexible message type supporting strings, functions, and templates
- Enables dynamic content based on context

**[`ImageSpec`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ImageSpec)**
- Flexible image type supporting URLs, file paths, base64, and functions
- Enables dynamic image content

### Prompt Building

**[`PromptBuilder.WithImage(image ImageSpec) *PromptBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#PromptBuilder.WithImage)**
- Adds image to prompt (URL, file path, base64, or function)

**[`PromptBuilder.WithInlineKeyboard(keyboard KeyboardFunc) *PromptBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#PromptBuilder.WithInlineKeyboard)**
- Adds interactive inline keyboard to prompt

**[`PromptBuilder.WithTemplateData(data map[string]interface{}) *PromptBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#PromptBuilder.WithTemplateData)**
- Sets template variables for prompt rendering

**[`KeyboardFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#KeyboardFunc)**
- Function signature for generating interactive keyboards: `func(ctx *Context) *InlineKeyboardBuilder`

### Processing Input

**[`PromptBuilder.Process(processFunc ProcessFunc) *StepBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#PromptBuilder.Process)**
- Sets processing function for user input
- Completes step definition

**[`ProcessFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ProcessFunc)**
- Core processing function: `func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult`
- Handles both text input and button clicks

**[`ButtonClick`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ButtonClick)**
- Information about button click events with original callback data preserved
- Contains Data, Text, UserID, ChatID, and optional Metadata

### Flow Control

**[`ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ProcessResult)**
- Specifies outcome of ProcessFunc execution and flow navigation
- Combines action with optional user prompt

**Flow Navigation Functions:**
- **[`NextStep() ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#NextStep)** - Proceed to next sequential step
- **[`GoToStep(stepName string) ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#GoToStep)** - Jump to specific step
- **[`Retry() ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Retry)** - Retry current step
- **[`CompleteFlow() ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#CompleteFlow)** - Mark flow as completed
- **[`CancelFlow() ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#CancelFlow)** - Cancel flow immediately

**Result Enhancement:**
- **[`ProcessResult.WithPrompt(prompt MessageSpec) ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ProcessResult.WithPrompt)** - Add informational prompt
- **[`ProcessResult.WithImage(image ImageSpec) ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ProcessResult.WithImage)** - Add image to result prompt
- **[`ProcessResult.WithTemplateData(data map[string]interface{}) ProcessResult`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ProcessResult.WithTemplateData)** - Add template data

### Flow Lifecycle

**[`FlowBuilder.OnComplete(handler func(*Context) error) *FlowBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder.OnComplete)**
- Sets completion handler for entire flow

**[`FlowBuilder.OnError(config *ErrorConfig) *FlowBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder.OnError)**
- Sets error handling configuration for flow

**[`FlowBuilder.WithTimeout(duration time.Duration) *FlowBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder.WithTimeout)**
- Sets maximum duration for flow execution

### Message Processing Behavior

**[`ProcessMessageAction`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ProcessMessageAction)**
- Controls how previous messages are handled during button click processing
- Options: `ProcessKeepMessage`, `ProcessDeleteMessage`, `ProcessDeleteKeyboard`

**[`FlowBuilder.OnProcessDeleteMessage() *FlowBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder.OnProcessDeleteMessage)**
- Delete entire previous messages on button clicks

**[`FlowBuilder.OnProcessDeleteKeyboard() *FlowBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder.OnProcessDeleteKeyboard)**
- Remove only keyboards from previous messages

### Flow Management

**[`FlowBuilder.Build() (*Flow, error)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#FlowBuilder.Build)**
- Creates final Flow object with validation

**[`Bot.RegisterFlow(flow *Flow)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.RegisterFlow)**
- Registers flow with bot's flow management system

**[`Context.StartFlow(flowName string) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.StartFlow)**
- Initiates flow for current user

**[`Context.IsUserInFlow() bool`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.IsUserInFlow)**
- Checks if user is actively in a flow

**[`Context.CancelFlow()`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Context.CancelFlow)**
- Cancels current flow for user

## 5. Keyboards

### Reply Keyboards

**[`ReplyKeyboard`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ReplyKeyboard)**
- Custom reply keyboard with buttons arranged in rows
- Appears below message input field

**[`BuildReplyKeyboard(buttons []string, buttonsPerRow int) *ReplyKeyboard`](https://pkg.go.dev/github.com/kslamph/teleflow/core#BuildReplyKeyboard)**
- Creates reply keyboard with automatic button distribution

**Reply Keyboard Configuration:**
- **[`ReplyKeyboard.Resize() *ReplyKeyboard`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ReplyKeyboard.Resize)** - Enable automatic keyboard resizing
- **[`ReplyKeyboard.OneTime() *ReplyKeyboard`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ReplyKeyboard.OneTime)** - Hide keyboard after use
- **[`ReplyKeyboard.Placeholder(text string) *ReplyKeyboard`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ReplyKeyboard.Placeholder)** - Set input field placeholder

**[`ReplyKeyboard.ToTgbotapi() tgbotapi.ReplyKeyboardMarkup`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ReplyKeyboard.ToTgbotapi)**
- Converts to telegram-bot-api format

### Inline Keyboards

**[`InlineKeyboard`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboard)**
- Inline keyboard with buttons arranged in rows
- Appears directly below messages

**[`NewInlineKeyboard() *InlineKeyboardBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#NewInlineKeyboard)**
- Creates new inline keyboard builder

**[`InlineKeyboardBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboardBuilder)**
- Fluent API for building inline keyboards with UUID management

**Inline Keyboard Building:**
- **[`InlineKeyboardBuilder.ButtonCallback(text string, data interface{}) *InlineKeyboardBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboardBuilder.ButtonCallback)** - Add callback button with automatic UUID generation
- **[`InlineKeyboardBuilder.ButtonUrl(text string, url string) *InlineKeyboardBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboardBuilder.ButtonUrl)** - Add URL button
- **[`InlineKeyboardBuilder.Row() *InlineKeyboardBuilder`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboardBuilder.Row)** - Start new row of buttons

**[`InlineKeyboardBuilder.Build() tgbotapi.InlineKeyboardMarkup`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboardBuilder.Build)**
- Finalizes keyboard construction

**[`InlineKeyboardBuilder.GetUUIDMapping() map[string]interface{}`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboardBuilder.GetUUIDMapping)**
- Returns UUID-to-data mapping for callback resolution

### Keyboard Types

**[`ReplyKeyboardButton`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ReplyKeyboardButton)**
- Individual button in reply keyboard with contact/location request support

**[`InlineKeyboardButton`](https://pkg.go.dev/github.com/kslamph/teleflow/core#InlineKeyboardButton)**
- Individual button in inline keyboard with callback, URL, and inline query support

## 6. Middleware

### Middleware System

**[`MiddlewareFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#MiddlewareFunc)**
- Unified middleware function signature: `func(next HandlerFunc) HandlerFunc`
- Intercepts all handler types (commands, text, callbacks, flows)

**[`Bot.UseMiddleware(m MiddlewareFunc)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.UseMiddleware)**
- Adds middleware to bot's middleware stack
- Executes in reverse order of registration (LIFO)

### Built-in Middleware

**[`LoggingMiddleware() MiddlewareFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#LoggingMiddleware)**
- Comprehensive request/response logging with configurable levels
- Logs update types, user IDs, execution times, and errors

**[`AuthMiddleware(accessManager AccessManager) MiddlewareFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#AuthMiddleware)**
- Permission-based access control using AccessManager interface
- Automatic context extraction and graceful error handling

**[`RateLimitMiddleware(requestsPerMinute int) MiddlewareFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#RateLimitMiddleware)**
- Per-user rate limiting to prevent spam and abuse
- Thread-safe with configurable request limits

**[`RecoveryMiddleware() MiddlewareFunc`](https://pkg.go.dev/github.com/kslamph/teleflow/core#RecoveryMiddleware)**
- Panic recovery with graceful error handling
- Prevents bot crashes from unexpected errors

### Access Management

**[`AccessManager`](https://pkg.go.dev/github.com/kslamph/teleflow/core#AccessManager)**
- Interface for authorization and automatic UI management
- Provides context-aware permission checking and keyboard generation

**[`PermissionContext`](https://pkg.go.dev/github.com/kslamph/teleflow/core#PermissionContext)**
- Rich context for permission checking with user, chat, command, and update information

## 7. State Management

### State Manager Interface

**[`StateManager`](https://pkg.go.dev/github.com/kslamph/teleflow/core#StateManager)**
- Interface for managing persistent user state data
- Storage-agnostic design supporting various backends

**State Manager Methods:**
- **[`SetState(userID int64, key string, value interface{}) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#StateManager.SetState)** - Store user data
- **[`GetState(userID int64, key string) (interface{}, bool)`](https://pkg.go.dev/github.com/kslamph/teleflow/core#StateManager.GetState)** - Retrieve user data
- **[`ClearState(userID int64) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#StateManager.ClearState)** - Remove all user data

### In-Memory Implementation

**[`NewInMemoryStateManager() StateManager`](https://pkg.go.dev/github.com/kslamph/teleflow/core#NewInMemoryStateManager)**
- Creates thread-safe in-memory state manager
- Suitable for development and small-scale deployments

## 8. Templating

### Template System

**[`ParseMode`](https://pkg.go.dev/github.com/kslamph/teleflow/core#ParseMode)**
- Telegram message parsing modes: `ParseModeNone`, `ParseModeMarkdown`, `ParseModeMarkdownV2`, `ParseModeHTML`

**[`Bot.AddTemplate(name, templateText string, parseMode ParseMode) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.AddTemplate)**
- Registers template with validation and built-in functions

**[`Bot.GetTemplateInfo(name string) *TemplateInfo`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.GetTemplateInfo)**
- Retrieves comprehensive template metadata

**[`Bot.ListTemplates() []string`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.ListTemplates)**
- Returns all registered template names

**[`Bot.HasTemplate(name string) bool`](https://pkg.go.dev/github.com/kslamph/teleflow/core#Bot.HasTemplate)**
- Checks template existence

### Template Features

**[`TemplateInfo`](https://pkg.go.dev/github.com/kslamph/teleflow/core#TemplateInfo)**
- Comprehensive template metadata with name, parse mode, and compiled template

**Template Functions:**
- `title`, `upper`, `lower` - Text case conversion
- `escape` - Auto-escape based on ParseMode
- `safe` - Bypass escaping (use with caution)

**Template Usage:**
- Variable substitution: `{{.VariableName}}`
- Conditionals: `{{if .Condition}}...{{end}}`
- Loops: `{{range .Items}}...{{end}}`
- Template references: `"template:template_name"`

## 9. Access Management

### Access Manager Interface

**[`AccessManager`](https://pkg.go.dev/github.com/kslamph/teleflow/core#AccessManager)**
- Provides authorization and automatic UI management capabilities
- Enables context-aware permission checking and keyboard generation

**Access Manager Methods:**
- **[`CheckPermission(ctx *PermissionContext) error`](https://pkg.go.dev/github.com/kslamph/teleflow/core#AccessManager.CheckPermission)** - Check user permissions
- **[`GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard`](https://pkg.go.dev/github.com/kslamph/teleflow/core#AccessManager.GetReplyKeyboard)** - Generate context-aware keyboards

**[`PermissionContext`](https://pkg.go.dev/github.com/kslamph/teleflow/core#PermissionContext)**
- Rich context containing UserID, ChatID, Command, Arguments, IsGroup, IsChannel, MessageID, and full Update

## 10. Menu Button

### Menu Button Configuration

**[`MenuButtonConfig`](https://pkg.go.dev/github.com/kslamph/teleflow/core#MenuButtonConfig)**
- Configuration for Telegram's native menu button
- Supports commands and default types

**[`NewCommandsMenuButton() *MenuButtonConfig`](https://pkg.go.dev/github.com/kslamph/teleflow/core#NewCommandsMenuButton)**
- Creates configuration for bot commands menu button

**[`NewDefaultMenuButton() *MenuButtonConfig`](https://pkg.go.dev/github.com/kslamph/teleflow/core#NewDefaultMenuButton)**
- Creates configuration for Telegram's default menu button

**[`MenuButtonConfig.AddCommand(text, command string) *MenuButtonConfig`](https://pkg.go.dev/github.com/kslamph/teleflow/core#MenuButtonConfig.AddCommand)**
- Adds command entry with fluent chaining support

## Next Steps and References

- **[README.md](README.md)** - Project overview and quick start
- **[get_started.md](get_started.md)** - Detailed tutorial and examples  
- **[architecture.md](architecture.md)** - System design and architectural decisions
- **[GoDoc](https://pkg.go.dev/github.com/kslamph/teleflow/core)** - Complete API reference with implementation details

This API guide provides the foundation for building sophisticated Telegram bots with teleflow. The Step-Prompt-Process API enables declarative flow definitions, while the comprehensive middleware system and template support provide the tools needed for production-ready bot applications.