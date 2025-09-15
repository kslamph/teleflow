# GoDoc Documentation Summary for Teleflow Package

This document summarizes the comprehensive GoDoc comments in the Teleflow package to provide thorough API documentation.

## Package Overview

### Main Package Documentation (`core/doc.go`)
- Created comprehensive package-level documentation explaining the framework's purpose
- Included quick start examples for basic bot creation
- Documented flow management capabilities with examples
- Showed how to use the new simplified API

## Core Components Documented

### 1. Bot Management (`core/bot.go`)

**Types:**
- `Handler` - Interface for handling Telegram updates
- `Bot` - Main bot instance with full documentation

**Functions:**
- `NewBot()` - Bot creation with examples
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

### 1. Comprehensive Function Documentation
- Purpose and behavior explanation
- Parameter descriptions
- Return value explanations
- Usage examples for complex functions
- Cross-references to related functions

### 2. Type Documentation
- Clear purpose statements
- Field descriptions with types and purposes
- Usage context explanations
- Relationship to other types

### 3. Constant Documentation
- Value meanings and use cases
- When to use each constant
- Behavior differences between options

### 4. Example Integration
- Practical usage examples
- Complete code snippets
- Common patterns and best practices
- Error handling examples

### 5. Cross-Reference Documentation
- Links between related functions and types
- Workflow explanations
- Integration patterns

## Benefits of Added Documentation

1. **Developer Experience**: Clear understanding of API usage
2. **Maintainability**: Easier codebase navigation and modification
3. **Onboarding**: Faster learning curve for new developers
4. **API Discoverability**: Better IDE support and code completion
5. **Best Practices**: Examples showing recommended usage patterns
6. **Error Prevention**: Clear parameter requirements and constraints

## Tools Integration

The comprehensive GoDoc comments enable:
- `go doc` command usage for API exploration
- IDE hover documentation
- Online documentation generation
- API reference websites
- Code completion improvements
- Static analysis tool integration