# GoDoc Documentation Summary for Teleflow Package

This document summarizes all the comprehensive GoDoc comments added to the Teleflow package to provide thorough API documentation.

## Package Overview

### Main Package Documentation (`core/doc.go`)
- Created comprehensive package-level documentation explaining the framework's purpose
- Included quick start examples for basic bot creation
- Documented flow management capabilities with examples
- Provided template usage examples
- Explained middleware integration
- Showed error handling strategies
- Demonstrated keyboard creation

## Core Components Documented

### 1. Bot Management (`core/bot.go`)

**Types:**
- `HandlerFunc` - Generic handler for processing updates
- `CommandHandlerFunc` - Handler for Telegram commands
- `TextHandlerFunc` - Handler for specific text messages
- `DefaultHandlerFunc` - Fallback handler for unmatched messages
- `BotOption` - Configuration option for bot customization
- `PermissionContext` - Context for access control decisions
- `AccessManager` - Interface for controlling user access
- `Bot` - Main bot instance with full documentation

**Functions:**
- `NewBot()` - Bot creation with examples
- `WithFlowConfig()` - Flow configuration option
- `WithAccessManager()` - Access control configuration
- `UseMiddleware()` - Middleware registration
- `HandleCommand()` - Command handler registration
- `HandleText()` - Text handler registration
- `DefaultHandler()` - Default handler registration
- `RegisterFlow()` - Flow registration
- `SetBotCommands()` - Telegram command menu configuration
- `DeleteMessage()` - Message deletion
- `EditMessageReplyMarkup()` - Keyboard editing
- `Start()` - Bot event loop

### 2. Context Management (`core/context.go`)

**Types:**
- `Context` - Rich interface for handling updates with comprehensive field documentation

**Functions:**
- `UserID()` - User identification
- `ChatID()` - Chat identification
- `Set()/Get()` - Context data storage
- `SetFlowData()/GetFlowData()` - Flow-specific data management
- `StartFlow()` - Flow initiation
- `CancelFlow()` - Flow cancellation
- `SendPrompt()` - Rich message sending with examples
- `SendPromptText()` - Simple text sending
- `SendPromptWithTemplate()` - Template-based messaging
- Template management methods with examples
- `IsGroup()/IsChannel()` - Chat type detection
- `SetPendingReplyKeyboard()` - Keyboard attachment

### 3. Flow System (`core/flow_types.go`, `core/flow.go`, `core/flow_builder.go`)

**Flow Types:**
- `FlowBuilder` - Fluent interface for flow construction
- `StepBuilder` - Individual step configuration
- `PromptConfig` - Prompt configuration with all options
- `ProcessFunc` - User input processing function
- `ButtonClick` - Button interaction data
- `ProcessResult` - Flow control result with helper methods
- `ButtonClickAction` - Message action after button clicks

**Flow Management:**
- `ErrorConfig` - Error handling configuration
- `FlowConfig` - Global flow behavior configuration
- `flowManager` - Internal flow state management

**Functions:**
- `NewFlow()` - Flow creation with comprehensive examples
- `Step()` - Step addition with validation
- `OnComplete()` - Completion handler setup
- `OnError()` - Error handling configuration
- `WithTimeout()` - Flow timeout settings
- `OnButtonClick()` - Button click behavior
- `Build()` - Flow finalization and validation
- `Prompt()` - Step prompt configuration
- `WithTemplateData()` - Template data addition
- `WithImage()` - Image attachment
- `WithPromptKeyboard()` - Keyboard attachment
- `Process()` - Input processing setup

**Flow Control Functions:**
- `NextStep()` - Advance to next step
- `GoToStep()` - Jump to specific step
- `Retry()` - Repeat current step
- `CompleteFlow()` - Finish flow successfully
- `CancelFlow()` - Terminate flow
- `OnErrorCancel()` - Cancel on error strategy
- `OnErrorRetry()` - Retry on error strategy
- `OnErrorIgnore()` - Ignore error strategy

### 4. Keyboard System (`core/keyboards.go`)

**Types:**
- `ReplyKeyboardButton` - Individual keyboard button
- `ReplyKeyboard` - Complete keyboard structure
- `ReplyKeyboardBuilder` - Fluent keyboard construction

**Functions:**
- `BuildReplyKeyboard()` - Quick keyboard creation
- `NewReplyKeyboard()` - Builder initialization
- `AddButton()` - Standard button addition
- `AddContactButton()` - Contact request button
- `AddLocationButton()` - Location request button
- `Row()` - Row management
- `Resize()` - Keyboard sizing
- `OneTime()` - One-time keyboard behavior
- `Placeholder()` - Input field placeholder
- `Selective()` - Selective display
- `Build()` - Keyboard finalization

### 5. Template System (`core/templates.go`)

**Types:**
- `ParseMode` - Telegram formatting modes with detailed explanations
- `TemplateInfo` - Template metadata structure

**Functions:**
- `AddTemplate()` - Global template registration
- `GetTemplateInfo()` - Template information retrieval
- `ListTemplates()` - Template enumeration
- `HasTemplate()` - Template existence check
- `validateParseMode()` - Parse mode validation
- `validateTemplateIntegrity()` - Template compatibility check

### 6. Middleware System (`core/middleware_types.go`)

**Types:**
- `MiddlewareFunc` - Middleware function signature with examples

### 7. Interfaces (`core/interfaces.go`)

**Documented Interfaces:**
- `PromptSender` - Message composition and sending
- `MessageCleaner` - Message management operations
- `ContextFlowOperations` - Flow interaction methods
- `TelegramClient` - Telegram API abstraction

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

