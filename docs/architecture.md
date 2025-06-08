# Architecture Overview

This document provides a comprehensive overview of the `teleflow/core` package architecture, including its key components, their interactions, and the request processing lifecycle.

## Design Philosophy

The `teleflow/core` package is built around the following core principles:

- **Declarative Flow Design**: Define complex conversations using a Step-Prompt-Process API that separates flow structure from business logic
- **Context-Driven Operations**: All operations flow through a rich [`Context`](../core/context.go:12) object that provides unified access to bot capabilities
- **Unified Middleware System**: Consistent middleware processing across all handler types (commands, text, callbacks, flows)
- **Type-Safe APIs**: Comprehensive type safety for all bot interactions with clear error handling
- **Pluggable Architecture**: Extensible design with interfaces for state management, access control, and custom components

## Core Components

### Bot - Central Orchestrator

**File**: [`core/bot.go`](../core/bot.go)

The [`Bot`](../core/bot.go) serves as the central orchestrator of the entire system. It manages:

- **Update Reception**: Receives and routes incoming Telegram updates
- **Handler Registration**: Maintains registries for commands, text handlers, and flows
- **Middleware Stack**: Executes middleware chain for all requests
- **Flow Management**: Coordinates flow execution and state transitions
- **API Integration**: Provides unified access to Telegram Bot API

**Key Responsibilities**:
- Initialize core system components (state manager, callback registry, prompt composer)
- Route updates to appropriate handlers based on type and content
- Manage bot lifecycle (start, stop, graceful shutdown)
- Coordinate between all system components

### Context - Request Lifecycle Manager

**File**: [`core/context.go`](../core/context.go)

The [`Context`](../core/context.go) is the heart of all bot interactions, providing:

- **Request Data**: Access to current Telegram update, user info, and chat details
- **Helper Methods**: Convenient methods for sending messages, templates, and flow control
- **Data Storage**: Two-tier storage system for request-scoped and flow-persistent data
- **Bot Operations**: Direct access to bot capabilities through helper methods

**Data Storage Mechanisms**:
- **Request-scoped**: [`Set()`](../core/context.go)/[`Get()`](../core/context.go) for temporary data during update processing
- **Flow-persistent**: [`SetFlowData()`](../core/context.go)/[`GetFlowData()`](../core/context.go) for data that persists across flow steps

### Flow System - Conversation Management

**Files**: [`core/flow.go`](../core/flow.go), [`core/flow_builder.go`](../core/flow_builder.go), [`core/flow_types.go`](../core/flow_types.go)

The Flow System implements the Step-Prompt-Process API paradigm:

**Components**:
- **[`FlowBuilder`](../core/flow_types.go)**: Fluent API for flow construction
- **[`StepBuilder`](../core/flow_types.go)**: Individual step configuration
- **Flow Runtime**: Execution engine with state management and error handling
- **[`ProcessFunc`](../core/flow_types.go)**: User-defined logic for input processing

**Key Features**:
- **Step Sequencing**: Automatic progression through defined steps
- **Error Handling**: Configurable strategies (cancel, retry, ignore)
- **State Persistence**: Automatic state synchronization across steps
- **Navigation Control**: Dynamic step jumping and flow completion

### PromptComposer & Rendering System

**Files**: [`core/prompt_composer.go`](../core/prompt_composer.go), [`core/message_renderer.go`](../core/message_renderer.go), [`core/image_handler.go`](../core/image_handler.go)

The rendering system orchestrates message composition and delivery:

**[`PromptComposer`](../core/prompt_composer.go)**:
- Coordinates message rendering, image processing, and keyboard attachment
- Handles multiple content types (text, images, templates)
- Manages Telegram API communication

**[`MessageRenderer`](../core/message_renderer.go)**:
- Template processing and variable substitution
- Telegram format validation (Markdown, MarkdownV2, HTML)
- Security escaping and content sanitization

**[`ImageHandler`](../core/image_handler.go)**:
- Multiple image source support (files, URLs, base64, file paths)
- Format validation and processing
- Integration with message composition

### Keyboard System

**Files**: [`core/keyboards.go`](../core/keyboards.go), [`core/inline_keyboard_builder.go`](../core/inline_keyboard_builder.go), [`core/keyboard_builder.go`](../core/keyboard_builder.go)

The keyboard system provides multiple approaches for creating interactive interfaces:

**Modern Map-based Approach** (Recommended for flows):
```go
map[string]interface{}{
    "✅ Approve": "approve_123",
    "❌ Reject":  "reject_123",
}
```

**Builder Pattern** (For complex scenarios):
- **[`InlineKeyboardBuilder`](../core/inline_keyboard_builder.go)**: Fluent API for inline keyboards
- **[`KeyboardBuilder`](../core/keyboard_builder.go)**: Reply keyboard construction
- **UUID Mapping**: Automatic callback data management

### CallbackRegistry - Button Interaction Handler

**File**: [`core/callbacks.go`](../core/callbacks.go)

The [`CallbackRegistry`](../core/callbacks.go) manages inline keyboard button interactions:

- **Thread-safe Registration**: Concurrent handler registration and lookup
- **Pattern Matching**: Routing callback data to appropriate handlers
- **Flow Integration**: Seamless integration with Step-Prompt-Process API
- **Automatic Cleanup**: Memory management for expired callbacks

### Middleware System

**Files**: [`core/middleware.go`](../core/middleware.go), [`core/middleware_types.go`](../core/middleware_types.go)

Unified middleware processing for all handler types:

**Built-in Middleware**:
- **[`LoggingMiddleware`](../core/middleware.go)**: Request/response logging
- **[`AuthMiddleware`](../core/middleware.go)**: Permission-based access control
- **[`RateLimitMiddleware`](../core/middleware.go)**: Per-user rate limiting
- **[`RecoveryMiddleware`](../core/middleware.go)**: Panic recovery with graceful error handling

**Middleware Chain**: All handlers (commands, text, callbacks, flows) pass through the same middleware stack.

### StateManager - Data Persistence

**File**: [`core/state.go`](../core/state.go)

The [`StateManager`](../core/state.go) provides persistent storage for user data:

**Features**:
- **User State**: Persistent data across different conversations
- **Flow State**: Temporary data during active flow execution
- **Thread-safe**: Concurrent access from multiple users
- **Pluggable**: Interface-based design for custom storage backends

**Default Implementation**: In-memory storage suitable for development and small-scale deployments.

### Template System

**Files**: [`core/templates.go`](../core/templates.go), [`core/template_manager.go`](../core/template_manager.go)

Dynamic message generation with Telegram-specific formatting:

**Features**:
- **Go Template Syntax**: Variable substitution, conditionals, loops
- **Built-in Functions**: `title`, `upper`, `lower`, `escape`, `safe`
- **Multiple Parse Modes**: Support for Markdown, MarkdownV2, HTML
- **Security**: Automatic escaping and validation
- **Data Precedence**: Template data overrides context data

### MenuButton Configuration

**File**: [`core/menu_button.go`](../core/menu_button.go)

Bot menu button configuration for enhanced user experience:

- **Command Integration**: Direct command triggering from menu
- **Web App Support**: Integration with Telegram Web Apps
- **Dynamic Configuration**: Runtime menu button updates

## Component Interactions

### Request Processing Flow

```
Telegram Update → Bot → Context Creation → Middleware Stack → Handler/Flow → Response
                   ↓
              StateManager ← PromptComposer ← CallbackRegistry
```

### Flow Execution Lifecycle

```
StartFlow → FlowBuilder → StepBuilder → PromptComposer → User Input → ProcessFunc
    ↓                                                                      ↓
StateManager ← Context ← MessageRenderer ← Templates ← CallbackRegistry ← NextStep/Retry/Complete
```

### Key Interaction Patterns

1. **Update Reception**:
   1. **Update Reception**:
      - [`Bot`](../core/bot.go) receives Telegram update
      - Creates [`Context`](../core/context.go) with update data
      - Loads user state from [`StateManager`](../core/state.go)
   
   2. **Middleware Processing**:
      - [`Context`](../core/context.go) passes through middleware chain
      - Authentication, rate limiting, logging applied uniformly
      - Request may be rejected before reaching handlers
   
   3. **Handler Execution**:
      - Update routed to appropriate handler (command, text, flow)
      - Handlers use [`Context`](../core/context.go) methods for responses
      - [`PromptComposer`](../core/prompt_composer.go) orchestrates message sending
   
   4. **Flow Processing**:
      - [`Context`](../core/context.go) coordinates between flow steps
      - [`StateManager`](../core/state.go) persists data across steps
      - [`CallbackRegistry`](../core/callbacks.go) handles button interactions
## Request Lifecycle

### 1. Update Reception
### 1. Update Reception
- Telegram sends update to bot webhook/polling endpoint
- [`Bot`](../core/bot.go) receives and parses the update
- Initial validation and routing preparation

### 2. Context Creation
- [`Context`](../core/context.go) object created with update data
- User and chat information extracted
- [`StateManager`](../core/state.go) loads existing user state

### 3. Middleware Execution
- Request passes through registered middleware stack
- Authentication, rate limiting, logging applied
- Request may be rejected at this stage

### 4. Handler Routing
- Update type and content determine handler selection
- Commands routed to command handlers
- Text messages routed to text handlers
- Callback queries routed to [`CallbackRegistry`](../core/callbacks.go)
- Active flows intercept appropriate updates

### 5. Handler Processing
- Selected handler executes with [`Context`](../core/context.go)
- Business logic processes the request
- Response preparation using context helper methods

### 6. Response Generation
- [`PromptComposer`](../core/prompt_composer.go) coordinates response composition
- [`MessageRenderer`](../core/message_renderer.go) processes templates
- [`ImageHandler`](../core/image_handler.go) processes any images
- Keyboards attached via keyboard builders

### 7. Response Delivery
- Composed message sent via Telegram Bot API
- [`StateManager`](../core/state.go) persists any state changes
- Flow state updated if applicable
### 8. Cleanup and Completion
- Request-scoped data cleaned up
- Middleware cleanup (if any)
- Ready for next update

## Key Design Principles

### Separation of Concerns
### Separation of Concerns
Each component has a well-defined responsibility:
- **[`Bot`](../core/bot.go)**: Orchestration and routing
- **[`Context`](../core/context.go)**: Request lifecycle management
- **Flow System**: Conversation logic
- **[`PromptComposer`](../core/prompt_composer.go)**: Message composition
- **[`StateManager`](../core/state.go)**: Data persistence

### Unified Processing Model
All handler types (commands, text, callbacks, flows) share:
- Same [`Context`](../core/context.go) object
- Same middleware stack
- Same helper methods
- Same error handling patterns
### Extensibility
Interfaces enable custom implementations:
- [`StateManager`](../core/state.go) for different storage backends
- [`AccessManager`](../core/middleware.go) for custom authorization
- [`MiddlewareFunc`](../core/middleware_types.go) for custom middleware

### Type Safety
Comprehensive type definitions ensure:
- Compile-time error detection
- Clear API contracts
- IDE support and autocompletion

## Next Steps

To dive deeper into specific aspects of the architecture:

- **[Getting Started](./get_started.md)**: Learn how to build your first bot with practical examples
- **[API Guide](./api_guide.md)**: Detailed reference for all available APIs and methods
- **[README](./README.md)**: Package overview and feature highlights

For hands-on learning, explore the complete examples in the repository:
- [`example/basic-flow/`](../example/basic-flow/): Complete registration bot demonstrating key concepts
- [`example/process-message-actions/`](../example/process-message-actions/): Advanced message handling patterns
- [`example/template/`](../example/template/): Template system showcase with formatting examples