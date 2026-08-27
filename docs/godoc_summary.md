# GoDoc Documentation Summary for Teleflow Package

This document summarizes the comprehensive GoDoc comments in the Teleflow package to provide thorough API documentation.

## Package Overview

### Main Package Documentation (`core/doc.go`)
- Package-level documentation explaining the framework's purpose
- Quick start example for basic bot creation
- Flow management capabilities with example code

## Core Components Documented

### 1. Bot Management (`core/bot.go`)

**Types:**
- `Handler` - Interface for handling Telegram updates
- `Bot` - Main bot instance with full documentation

**Functions:**
- `NewBot()` - Bot creation
- `RegisterData()` - Register complex data for callbacks
- `GetData()` - Retrieve complex callback data
- `Start()` - Bot event loop
- `SetBotCommands()` - Telegram command menu configuration

### 2. Context Management (`core/context.go`)

**Types:**
- `Context` - Toolkit for handling a single Telegram update

**Constants:**
- `ParseModeHTML` - HTML parse mode for message formatting
- `ParseModeMarkdown` - Markdown parse mode for message formatting
- `ParseModeMarkdownV2` - MarkdownV2 parse mode for message formatting

**Functions:**
- `Update()` - Access raw Telegram update
- `API()` - Access underlying Bot API instance
- `UserID()` - User identification
- `ChatID()` - Chat identification
- `Text()` - Get text from message or callback data
- `Set()/Get()` - Context data storage
- `RegisterCallback()` - Register one-time callback handlers
- `RegisterData()` - Register complex data for callbacks
- `GetData()` - Retrieve complex callback data
- `ReplyText()` - Send simple text message
- `ReplyWithKeyboard()` - Send message with keyboard
- `ReplyTextWithParseMode()` - Send formatted text message
- `ReplyWithKeyboardAndParseMode()` - Send formatted message with keyboard
- `DeleteMessage()` - Delete a message
- `EditMessageText()` - Edit message text
- `EditMessageTextWithParseMode()` - Edit message text with formatting
- `EditMessageReplyMarkup()` - Edit message keyboard
- `EditMessageMedia()` - Edit message media (e.g. photo)
- `AnswerCallback()` - Acknowledge callback query

### 3. Flow System (`core/flow.go`)

**Flow Types:**
- `Flow[T]` - Generic flow with type-safe state management
- `Step[T]` - Individual step with prompt and process functions
- `FlowAction` - Action to control flow progression

**Flow Management:**
- `NewFlow()` - Create new flow
- `AddStep()` - Add step to flow
- `OnComplete()` - Set completion handler
- `IsUserInFlow()` - Check if user is in flow
- `CancelFlow()` - Cancel user's flow
- `HandleUpdate()` - Process update in flow

**Functions:**
- `Next()` - Advance to next step
- `Retry()` - Repeat current step
- `End()` - Finish flow successfully
- `GoTo()` - Jump to specific named step

### 4. Router System (`core/flow.go`)

**Types:**
- `Router` - Command-based multiplexer for handlers

**Functions:**
- `NewRouter()` - Create new router
- `Handle()` - Register handler for command
- `Default()` - Set default handler
- `HandleUpdate()` - Process update in router

### 5. Keyboard System (`core/flow.go`)

**Types:**
- `InlineKeyboard` - Builder for inline keyboards

**Functions:**
- `NewInlineKeyboard()` - Create new keyboard builder
- `AddRow()` - Add row of buttons
- `Build()` - Finalize keyboard
- `ButtonCallback()` - Create callback button

### 6. Handler Adapters (`core/flow.go`)

**Types:**
- `Func` - Adapter to use functions as handlers

## Documentation Standards Applied

### 1. Function Documentation
- Purpose and behavior explanation for each function/method
- Parameter descriptions where the name alone isn't sufficient
- Return value explanations for non-obvious returns
- Notes on thread safety, concurrency, and one-time semantics where relevant

### 2. Type Documentation
- Clear purpose statements for each exported type
- Field descriptions explaining what each field controls
- Relationship to other types (e.g., `Step[T]` as a component of `Flow[T]`)

### 3. Constant Documentation
- Value meanings and when to use each constant
- Behavior differences between options (e.g., parse modes)

### 4. Package-Level Example
- `core/doc.go` includes a complete, runnable example demonstrating flow creation, step definition, and bot wiring
- This serves as the primary usage reference for `go doc` and pkg.go.dev viewers

## Benefits of Added Documentation

1. **Developer Experience**: Purpose and behavior descriptions help users understand API usage without reading implementation
2. **Maintainability**: Clear doc comments make the codebase easier to navigate and modify
3. **Onboarding**: Package-level example in `doc.go` provides a working reference for new users
4. **API Discoverability**: Structured doc comments improve `go doc` output and pkg.go.dev rendering
5. **Constraint Communication**: Doc comments document one-time callback semantics, thread-safety expectations, and state-management behavior

## Tools Integration

The GoDoc comments enable:
- `go doc` command usage for API exploration
- pkg.go.dev rendering of type/method documentation
- IDE hover documentation and go-to-definition support