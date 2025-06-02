# Teleflow Implementation Plan - Phase 1

## Project Structure

```
teleflow/
├── README.md                    # Project overview and quick start
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── .gitignore                  # Git ignore file
├── LICENSE                     # MIT License
├── core/                       # Core framework packages
│   ├── bot.go                  # Main Bot struct and configuration
│   ├── context.go              # Context struct and helper methods
│   ├── handlers.go             # Handler types and registration
│   ├── keyboards.go            # Keyboard abstractions
│   ├── state.go                # State management interfaces and implementations
│   ├── flow.go                 # Flow management and DSL
│   ├── callbacks.go            # Type-safe callback system
│   └── templates.go            # Message templating
├── middleware/                 # Built-in middleware packages
│   ├── logging.go              # Logging middleware
│   ├── auth.go                 # Authentication middleware
│   ├── ratelimit.go            # Rate limiting middleware
│   └── recovery.go             # Panic recovery middleware
├── ui/                         # UI component system (future extension)
│   └── README.md               # Placeholder for UI components
├── examples/                   # Example implementations
│   ├── basic-bot/              # Simple command-based bot
│   │   └── main.go
│   ├── flow-bot/               # Multi-step conversational bot
│   │   └── main.go
│   ├── middleware-bot/         # Bot with middleware demonstration
│   │   └── main.go
│   └── ui-components/          # UI component examples (future)
│       └── README.md
├── docs/                       # Documentation
│   ├── getting-started.md      # Getting started guide
│   ├── api-reference.md        # Complete API reference
│   ├── flow-guide.md           # Flow system guide
│   ├── middleware-guide.md     # Middleware development guide
│   └── examples-guide.md       # Examples walkthrough
└── tests/                      # Test files (future phases)
    └── README.md
```

## Phase 1: API and Examples Implementation

### Objective
Implement the core API exactly as designed in `newdesign.md` with working examples to validate the API design and developer experience.

### Sub-Tasks

#### Task 1: Project Setup and Structure
**Duration: 1 day**

1.1. **Create Go Module**
   - Initialize `go.mod` with module name `github.com/kslamph/teleflow`
   - Add dependency: `github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1`
   - Set Go version to 1.24 

1.2. **Create Directory Structure**
   - Create all directories as specified in project structure
   - Add placeholder README.md files in empty directories

1.3. **Create Base Files**
   - Create `.gitignore` with Go template
   - Create `LICENSE` file with MIT license
   - Create basic `README.md` with project description

#### Task 2: Core Package Implementation
**Duration: 5 days**

2.1. **Implement `core/bot.go`** (1 day)
   - Copy exact code from `newdesign.md` bot.go section (lines 30-261)
   - Implement all types: `Bot`, `HandlerFunc`, `MiddlewareFunc`, `BotOption`
   - Implement all functions: `NewBot`, `WithMainMenu`, `WithFlowConfig`, `WithExitCommands`, `WithUserPermissions`
   - Implement bot methods: `Use`, `HandleCommand`, `HandleText`, `RegisterCallback`, `RegisterFlow`
   - Implement internal methods: `applyMiddleware`, `wrapWithPermissions`, `processUpdate`, `isGlobalExitCommand`, `resolveGlobalCommand`, `resolveHandler`
   - Implement `Start` method with update processing loop
   - **Requirements**: Must compile without errors, all public APIs must match design exactly

2.2. **Implement `core/context.go`** (1 day)
   - Copy exact code from `newdesign.md` context.go section (lines 264-415)
   - Implement `Context` struct with fields: `Bot`, `Update`, `data`, `userID`, `chatID`
   - Implement all methods: `NewContext`, `UserID`, `ChatID`, `Set`, `Get`, `Reply`, `ReplyTemplate`, `EditOrReply`, `StartFlow`
   - Implement internal methods: `send`, `extractUserID`, `extractChatID`
   - **Requirements**: All context methods must work with both messages and callback queries

2.3. **Implement `core/keyboards.go`** (1 day)
   - Copy exact code from `newdesign.md` keyboards.go section (lines 418-520)
   - Implement types: `ReplyKeyboardButton`, `ReplyKeyboard`, `InlineKeyboardButton`, `InlineKeyboard`, `MenuButtonConfig`, `WebAppInfo`
   - Implement constructor functions: `NewReplyKeyboard`, `NewInlineKeyboard`
   - Implement conversion methods: `toTgbotapi()` for both keyboard types
   - **Requirements**: Must produce valid Telegram keyboard markup

2.4. **Implement `core/callbacks.go`** (1 day)
   - Copy exact code from `newdesign.md` callbacks.go section (lines 432-515)
   - Implement types: `CallbackHandler`, `CallbackRegistry`, `simpleCallbackHandler`, `ActionCallback`
   - Implement methods: `NewCallbackRegistry`, `Register`, `Handle`, `matchPattern`
   - Implement helper functions: `SimpleCallback`
   - **Requirements**: Pattern matching must work with exact matches and wildcard suffixes

2.5. **Implement `core/state.go`** (1 day)
   - Copy exact code from `newdesign.md` state.go section (lines 522-580)
   - Implement types: `StateManager` interface, `InMemoryStateManager`
   - Implement all methods: `NewInMemoryStateManager`, `SetState`, `GetState`, `ClearState`
   - **Requirements**: Must be thread-safe, data must persist across handler calls

#### Task 3: Flow System Implementation
**Duration: 3 days**

3.1. **Implement `core/flow.go` - Core Types** (1 day)
   - Copy exact code from `newdesign.md` flow.go section (lines 518-848)
   - Implement types: `FlowManager`, `Flow`, `FlowStep`, `UserFlowState`, `FlowBuilder`, `FlowStepBuilder`
   - Implement constructor functions: `NewFlowManager`, `NewFlow`
   - **Requirements**: All structs must have correct field types and relationships

3.2. **Implement `core/flow.go` - Flow Builder DSL** (1 day)
   - Implement FlowBuilder methods: `Step`, `OnComplete`, `OnCancel`, `Build`
   - Implement FlowStepBuilder methods: `WithValidator`, `NextStep`, `OnInput`, `WithTimeout`, `StayOnInvalidInput`
   - Implement fluent interface chain methods
   - **Requirements**: Builder pattern must allow method chaining, auto-linking of steps must work

3.3. **Implement `core/flow.go` - Flow Execution** (1 day)
   - Implement FlowManager methods: `SetBotConfig`, `IsUserInFlow`, `CancelFlow`, `RegisterFlow`, `StartFlow`, `HandleUpdate`
   - Implement flow execution logic: `determineNextStep`
   - Implement validator helper functions: `NumberValidator`, `ChoiceValidator`
   - **Requirements**: Flow execution must handle all transition types, validation must work with custom error messages

#### Task 4: Template System Implementation
**Duration: 1 day**

4.1. **Implement `core/templates.go`**
   - Implement template registration in Bot struct (already included in bot.go)
   - Implement `AddTemplate` method functionality
   - Implement `ReplyTemplate` context method functionality  
   - **Requirements**: Must use Go's `text/template` package, template execution must handle data properly

#### Task 5: Middleware System Implementation
**Duration: 2 days**

5.1. **Implement `middleware/logging.go`** (0.5 day)
   - Copy exact code from `newdesign.md` middleware.go LoggingMiddleware section
   - Implement `LoggingMiddleware()` function that returns `MiddlewareFunc`
   - Log format: `[userID] Processing updateType` and `[userID] Handler completed/failed in duration`
   - **Requirements**: Must log all updates and execution times, errors must be logged with details

5.2. **Implement `middleware/auth.go`** (0.5 day)
   - Copy exact code from `newdesign.md` middleware.go AuthMiddleware section
   - Implement `AuthMiddleware(checker UserPermissionChecker)` function
   - **Requirements**: Must check basic_access permission, unauthorized users get clear error message

5.3. **Implement `middleware/ratelimit.go`** (0.5 day)
   - Copy exact code from `newdesign.md` middleware.go RateLimitMiddleware section
   - Implement `RateLimitMiddleware(requestsPerMinute int)` function
   - **Requirements**: Must track per-user request times, rate limited users get clear message

5.4. **Implement `middleware/recovery.go`** (0.5 day)
   - Copy exact code from `newdesign.md` middleware.go RecoveryMiddleware section
   - Implement `RecoveryMiddleware()` function with panic recovery
   - **Requirements**: Must catch panics, log them, and send user-friendly error message

#### Task 6: Examples Implementation
**Duration: 3 days**

6.1. **Implement `examples/basic-bot/main.go`** (1 day)
   - Create simple bot with commands: `/start`, `/help`, `/ping`
   - Use reply keyboards with buttons: "🏠 Home", "ℹ️ Info", "❓ Help"
   - Demonstrate basic `HandleCommand` and `HandleText` usage
   - Use logging middleware
   - **Requirements**: Must demonstrate basic API usage, bot must respond to all defined commands/buttons

6.2. **Implement `examples/flow-bot/main.go`** (1 day)
   - Copy and adapt transfer flow example from `newdesign.md` (lines 1050-1100)
   - Implement complete transfer flow: amount → recipient → confirm
   - Use number validator for amount step
   - Demonstrate flow cancellation with `/cancel`
   - Use multiple middleware: logging, auth, recovery
   - **Requirements**: Flow must work end-to-end, validation must work, cancellation must work

6.3. **Implement `examples/middleware-bot/main.go`** (1 day)
   - Create bot demonstrating all middleware types
   - Implement custom permission checker with admin/user roles
   - Demonstrate rate limiting with low limit (2 requests per minute)
   - Add admin-only command that shows middleware effects
   - **Requirements**: All middleware must be visible in action, rate limiting must trigger, auth must block unauthorized users

#### Task 7: Documentation
**Duration: 2 days**

7.1. **Implement `docs/getting-started.md`** (0.5 day)
   - Installation instructions
   - "Hello World" bot in 10 lines
   - Basic concepts explanation: handlers, context, keyboards
   - Link to examples
   - **Requirements**: Must be complete enough for new users to get started in 5 minutes

7.2. **Implement `docs/api-reference.md`** (1 day)
   - Document all public types and methods from core package
   - Include code examples for each major feature
   - Document all middleware functions
   - **Requirements**: Must be complete API documentation with examples

7.3. **Implement `docs/flow-guide.md`** (0.5 day)
   - Explain flow concepts: steps, transitions, validation
   - Flow builder DSL examples
   - Best practices for flow design
   - **Requirements**: Must explain flow system thoroughly with practical examples

#### Task 8: Project Polish
**Duration: 1 day**

8.1. **Complete README.md**
   - Project description and features
   - Quick start example
   - Links to documentation and examples
   - Installation instructions
   - **Requirements**: Must represent the project professionally

8.2. **Add Package Documentation**
   - Add package-level comments to all core packages
   - Add example usage in package comments
   - **Requirements**: Must follow Go documentation conventions

### Acceptance Criteria

#### Functional Requirements
1. **All examples must run without errors** when provided with valid Telegram bot token
2. **Basic bot example** must respond to all commands and keyboard buttons
3. **Flow bot example** must complete transfer flow end-to-end with validation
4. **Middleware bot example** must demonstrate all middleware working
5. **All APIs must match the design exactly** - no deviations from `newdesign.md`

#### Quality Requirements
1. **Code must compile** without warnings with Go 1.19+
2. **No external dependencies** beyond `telegram-bot-api` package
3. **Thread-safe state management** must work correctly
4. **Memory leaks prevention** - proper cleanup of flow states
5. **Error handling** must be consistent and user-friendly

#### Documentation Requirements
1. **Getting started guide** must enable new users to create working bot in 5 minutes
2. **API reference** must document every public function with examples
3. **Code comments** must follow Go documentation conventions
4. **Examples** must be thoroughly commented and self-explanatory

### Definition of Done

Phase 1 is complete when:
1. ✅ All code files implement exact specifications from `newdesign.md`
2. ✅ All three example bots run successfully and demonstrate intended functionality
3. ✅ Documentation is complete and accurate
4. ✅ Project structure matches specification exactly
5. ✅ No compilation errors or warnings
6. ✅ All acceptance criteria are met

### Deliverables
1. Complete `teleflow/` project directory with all specified files
2. Three working example bots
3. Complete documentation set
4. Verification that all APIs work as designed

This phase focuses exclusively on **API validation** and **developer experience**. The implementation must be exact - no interpretation or "improvements" to the design. Any issues discovered should be documented for potential design revisions rather than changed during implementation.