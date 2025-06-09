# Comprehensive Refactoring Plan for Teleflow Core

**Version:** 1.0
**Date:** 2025-06-08
**Author:** Architect AI

## 1. Goals

*   **Enhanced Maintainability:** Code should be easier to understand, modify, and debug.
*   **Improved Testability:** Components should be unit-testable in isolation, with clear separation of concerns and mockable dependencies.
*   **Clean Code:** Adhere to Go best practices, promote readability, apply SRP (Single Responsibility Principle) where beneficial, and reduce code complexity.
*   **Stable User API Surface:** The public API for library users (e.g., `NewBot`, `HandleCommand`, `RegisterFlow`, `NewFlow`, `Step`) must remain unchanged or become even more intuitive. Internal refactoring should not leak complexity.
*   **Increased Automation (Internal):** Streamline internal processes, such as dependency resolution and context data flow, to reduce boilerplate and potential for errors.
*   **No Backward Compatibility Constraint (Internal):** Internal structures and function signatures can be changed freely to achieve the above goals.

## 2. Overall Architectural Notes & Principles

*   **Dependency Injection (DI):** Heavily utilize constructor-based DI for all major components. This is key for testability and decoupling.
*   **Interfaces:** Define interfaces for collaborations between components. This allows for mock implementations in tests and promotes loose coupling. The existing `BotAPI` in `prompt_composer.go` is a good starting point and will be generalized.
*   **Single Responsibility Principle (SRP):** Evaluate structs and functions to ensure they have a single, well-defined responsibility.
*   **Error Handling:** Standardize error handling. Errors should be descriptive and propagated clearly. Avoid panics for recoverable errors.
*   **Context (`core.Context`):**
    *   The `Context` object should primarily be a request-scoped data carrier and a means to access essential request-specific functionalities (like sending replies).
    *   It should not become a "god object" holding references to all services. Instead, services/handlers should receive their long-lived dependencies via DI.
*   **Configuration:** Configuration for components (like `FlowConfig` for `flowManager`) should be clearly passed during initialization.
*   **Immutability:** Where possible, favor immutability for configuration objects passed around.
*   **Package Structure:** Current single `core` package is acceptable for now, but as the system grows, consider sub-packages if logical boundaries become very distinct (e.g., `core/telegram`, `core/flows`, `core/templating`). For this refactor, we'll keep it as one `core` package.

## 3. Key Component Refactoring Plans

### Task Group 1: Bot API Abstraction & Bot Initialization

**Goal:** Decouple `Bot` and other components from the concrete `tgbotapi.BotAPI` for testability, while keeping `NewBot` user-friendly.

**Notes for Developer:**
*   The `BotAPI` interface in `prompt_composer.go` is a good start. We will rename and generalize it.
*   The user-facing `NewBot(token string, ...)` must remain unchanged.

**Tasks:**

1.  **Task 1.1: Define `TelegramClient` Interface**
    *   **File:** `core/telegram_client.go` (new file)
    *   **Instruction:** Create a `TelegramClient` interface. This interface should include all methods from `tgbotapi.BotAPI` that are *actually used* by any component within the `core` package (e.g., `Send`, `Request`, `GetUpdatesChan`, `GetMe` if `api.Self` is used).
    *   **Rename:** The existing `BotAPI` interface in `prompt_composer.go` will be replaced by this more general `TelegramClient`.
    *   **Example:**
        ```go
        package teleflow
        import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
        type TelegramClient interface {
            Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
            Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
            GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
            GetMe() (tgbotapi.User, error) // If bot.api.Self is used
        }
        ```

2.  **Task 1.2: Refactor `Bot` Struct and Initialization**
    *   **File:** `core/bot.go`
    *   **Instruction:**
        *   Change `Bot.api` field from `*tgbotapi.BotAPI` to `TelegramClient`.
        *   Modify `NewBot(token string, ...)`:
            *   It will still create `tgbotapi.NewBotAPI(token)`.
            *   It will then call a new *unexported* constructor `newBotInternal(client TelegramClient, options ...BotOption) (*Bot, error)`.
        *   The `newBotInternal` function will initialize the `Bot` struct, injecting the `TelegramClient`.
        *   All direct `b.api.*` calls will now use the interface methods.
    *   **Developer Note:** Ensure `b.api.Self` usage is handled, likely by calling `client.GetMe()` in `NewBot` and storing the `tgbotapi.User` in the `Bot` struct if needed, or having the mock provide it.

3.  **Task 1.3: Update `TestableBotAPI` (in `core/bot_test.go`)**
    *   **File:** `core/bot_test.go`
    *   **Instruction:**
        *   Modify `TestableBotAPI` to implement the new `TelegramClient` interface.
        *   Remove the `TestableBot` wrapper struct and its re-implemented `SetBotCommands`.
        *   Update tests to use `newBotInternal` (or an exported test helper if tests are in `_test` package) to inject `TestableBotAPI` into a *real* `Bot` instance.

### Task Group 2: Prompt Composition System

**Goal:** Make `PromptComposer` and its sub-handlers (`messageHandler`, `imageHandler`) testable and clearly define their dependencies.

**Notes for Developer:**
*   `PromptComposer` currently takes `BotAPI` (which will become `TelegramClient`), `messageHandler`, `imageHandler`, and `PromptKeyboardHandler`. This is a good DI pattern.

**Tasks:**

1.  **Task 2.1: Refactor `PromptComposer` Initialization**
    *   **File:** `core/prompt_composer.go`
    *   **Instruction:**
        *   The `newPromptComposer` function already takes `BotAPI` (to be `TelegramClient`), `*messageHandler`, `*imageHandler`, and `*PromptKeyboardHandler`. This is good.
        *   Ensure `botAPI BotAPI` field in `PromptComposer` is changed to `botAPI TelegramClient`.
        *   No major structural changes needed here if dependencies are already explicit.

2.  **Task 2.2: Refactor `messageHandler` (`prompt_message_handler.go`)**
    *   **File:** `core/prompt_message_handler.go`
    *   **Instruction:**
        *   `messageHandler` depends on `TemplateManager`. `newMessageRenderer` uses `GetDefaultTemplateManager()`.
        *   Modify `newMessageRenderer` (consider renaming to `newMessageHandler`) to accept `TemplateManager` as a parameter for better testability (constructor injection).
        *   The `Bot` struct will create the `TemplateManager` instance (or use the default) and pass it when creating `messageHandler`.

3.  **Task 2.3: Refactor `imageHandler` (`prompt_image_handler.go`)**
    *   **File:** `core/prompt_image_handler.go`
    *   **Instruction:** `imageHandler` currently has no external dependencies in its struct. Its methods like `processImage` are self-contained or use standard library functions. This is already quite testable. No major refactoring needed unless specific testability issues arise for its private helpers.

4.  **Task 2.4: Refactor `PromptKeyboardHandler` (`prompt_keyboard_handler.go`)**
    *   **File:** `core/prompt_keyboard_handler.go`
    *   **Instruction:** `PromptKeyboardHandler` is largely self-contained, managing `userUUIDMappings`. Its methods like `buildKeyboard` take `Context`. This is generally fine.
        *   **Define Interface:** Create `PromptKeyboardActions` (or similar) interface that `PromptKeyboardHandler` implements. This interface will be used by `flowManager` and potentially `Bot` or `Context` to interact with it.
        *   **Example Interface:**
            ```go
            // In core/prompt_keyboard_handler.go or a new interfaces.go
            type PromptKeyboardActions interface {
                BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error)
                GetCallbackData(userID int64, uuid string) (interface{}, bool)
                CleanupUserMappings(userID int64)
            }
            ```
        *   The `Bot` will hold an instance of `PromptKeyboardActions`.

### Task Group 3: Flow Management System (`flowManager`)

**Goal:** Decouple `flowManager` from concrete `Bot` and other components, making its state logic highly testable.

**Notes for Developer:**
*   `flowManager` is central to flow execution and state. Its dependencies need to be explicit.
*   The `initialize(bot *Bot)` method will be removed in favor of constructor injection.

**Tasks:**

1.  **Task 3.1: Define `FlowManagerDependencies` and Refactor `flowManager` Struct**
    *   **File:** `core/flow.go`
    *   **Instruction:**
        *   `flowManager` should not hold a `*Bot`.
        *   It needs:
            *   `FlowConfig` (passed at construction).
            *   An interface for sending prompts (e.g., `PromptSender`). `PromptComposer` can implement this.
            *   An interface for keyboard data (the `PromptKeyboardActions` from Task 2.4).
            *   An interface for deleting messages/keyboards if `handleMessageAction`'s logic is to remain in `flowManager` (e.g., `MessageEditor`). This could be part of `TelegramClient` or a more specific interface.
        *   **Refactor `flowManager` struct:**
            ```go
            type PromptSender interface { // To be implemented by PromptComposer
                ComposeAndSend(ctx *Context, config *PromptConfig) error
            }
            type MessageCleaner interface { // For deleting messages/keyboards
                DeleteMessage(chatID int64, messageID int) error
                EditMessageReplyMarkup(chatID int64, messageID int, replyMarkup interface{}) error // Example
            }

            type flowManager struct {
                flows     map[string]*Flow
                userFlows map[int64]*userFlowState
                botConfig *FlowConfig // From Bot's config

                promptSender   PromptSender
                keyboardAccess PromptKeyboardActions
                messageCleaner MessageCleaner // To be implemented by a component that uses TelegramClient
            }
            ```
        *   Modify `newFlowManager` to accept these dependencies: `newFlowManager(config *FlowConfig, pSender PromptSender, kAccess PromptKeyboardActions, mCleaner MessageCleaner) *flowManager`.
        *   Remove `fm.initialize(bot *Bot)`.

2.  **Task 3.2: Update `Bot` to Initialize `flowManager` with Dependencies**
    *   **File:** `core/bot.go`
    *   **Instruction:** In `newBotInternal`, create and pass the necessary implementations to `newFlowManager`.
        *   `PromptComposer` will implement `PromptSender`.
        *   `PromptKeyboardHandler` will implement `PromptKeyboardActions`.
        *   A new small component or methods on `Bot` itself (using its `TelegramClient`) can implement `MessageCleaner`.

3.  **Task 3.3: Adapt `flowManager` Methods**
    *   **File:** `core/flow.go`
    *   **Instruction:**
        *   Replace `fm.bot.promptComposer.composeAndSend` with `fm.promptSender.ComposeAndSend`.
        *   Replace `ctx.bot.GetPromptKeyboardHandler().*` calls with `fm.keyboardAccess.*`. This implies `Context` might not need direct access to `PromptKeyboardHandler` if `flowManager` handles all keyboard data extraction via its own injected dependency.
        *   Adapt `handleMessageAction`, `deletePreviousMessage`, `deletePreviousKeyboard` to use `fm.messageCleaner`.
        *   Review `extractInputData`: if it needs `keyboardAccess`, this dependency should be available.

4.  **Task 3.4: Review `setUserFlowData` and `getUserFlowData`**
    *   **File:** `core/flow.go` (lines 524-550)
    *   **Instruction:** These methods are currently on `flowManager` and are fine. `Context` calls them via `c.bot.flowManager.*`. This coupling from `Context` to `flowManager` is acceptable for flow data management but ensure `Context` gets `flowManager` in a clean way (see Task Group 4).

### Task Group 4: Context (`core.Context`)

**Goal:** Clarify `Context`'s role as a request-scoped data carrier and API for common request operations, without it becoming a service locator.

**Notes for Developer:**
*   `Context` currently holds `*Bot`, which gives it access to everything. This should be refined.
*   `Context` methods for sending messages (`sendSimpleText`, `answerCallbackQuery`) directly use `c.bot.api`. This is good if `c.bot.api` is the `TelegramClient` interface.

**Tasks:**

1.  **Task 4.1: Refine `Context` Dependencies**
    *   **File:** `core/context.go`
    *   **Instruction:**
        *   `Context` should still hold a reference to the `Bot` instance (`bot *Bot`), but its methods should prefer using well-defined interfaces on `Bot` or passed-in services if functionality is delegated.
        *   For sending messages: `c.bot.api.Send()` (where `api` is `TelegramClient`) is fine.
        *   For flow operations (`StartFlow`, `CancelFlow`, `SetFlowData`, `GetFlowData`): `c.bot.flowManager.*` is acceptable if `flowManager` itself is well-defined. Consider if `Bot` should expose these as methods, e.g., `c.bot.StartUserFlow(...)`.
        *   For `SendPrompt`: `c.bot.promptComposer.composeAndSend` is fine if `promptComposer` is a field on `Bot`.
        *   For template management: `defaultTemplateManager` is a global. Consider if `Context` should get `TemplateManager` via `c.bot.TemplateManager()` for consistency, or if the global is acceptable for this specific feature. For better testability of handlers using templates via context, injecting it via `Bot` is cleaner.

2.  **Task 4.2: Review `Context.data`**
    *   **File:** `core/context.go`
    *   **Instruction:** The `data map[string]interface{}` for request-scoped values is good. Ensure its usage is clear (e.g., primarily for middleware to pass data to handlers, or for `ProcessFunc` to store temporary step data before it's copied to `userFlowState.Data`).

### Task Group 5: Template Management (`TemplateManager`)

**Goal:** Ensure `TemplateManager` is testable and its usage is clean.

**Notes for Developer:**
*   `TemplateManager` uses a global `defaultTemplateManager`. This can make testing components that use it (like `messageHandler` or `Context` methods) harder if they directly access the global.

**Tasks:**

1.  **Task 5.1: Make `TemplateManager` Instantiable and Injectable**
    *   **File:** `core/template_manager.go`, `core/templates.go`
    *   **Instruction:**
        *   `newTemplateManager()` is good.
        *   `GetDefaultTemplateManager()` provides the global.
        *   Components that need `TemplateManager` (like `messageHandler`) should receive it via DI (as per Task 2.2).
        *   The `Bot` struct can hold a `TemplateManager` instance (either the default or a custom one if a `BotOption` allows it) and provide it to components that need it.
        *   Global functions in `templates.go` (`AddTemplate`, `GetTemplateInfo`, etc.) that use `defaultTemplateManager` are convenient for users but make direct testing of these functions harder. For internal use, prefer injected instances.

### Task Group 6: Builders (FlowBuilder, PromptKeyboardBuilder)

**Goal:** Ensure builders are robust and their produced objects are consistent.

**Notes for Developer:**
*   `FlowBuilder` (`core/flow_builder.go`) and `PromptKeyboardBuilder` (`core/prompt_keyboard_builder.go`) are user-facing APIs for defining flows and keyboards. Their external API should remain intuitive.
*   Internal validation within `Build()` methods is good.

**Tasks:**

1.  **Task 6.1: Review `FlowBuilder`**
    *   **File:** `core/flow_builder.go`, `core/flow_types.go`
    *   **Instruction:** The current builder pattern seems reasonable. Ensure all configurations set on the builder are correctly transferred to the `Flow` object. Validate step consistency (e.g., all steps in `Order` exist in `Steps` map).

2.  **Task 6.2: Review `PromptKeyboardBuilder`**
    *   **File:** `core/prompt_keyboard_builder.go`
    *   **Instruction:** This builder is also quite solid. The UUID mapping is internal and handled by `PromptKeyboardHandler`. Ensure `validateBuilder` is comprehensive.

### Task Group 7: Middleware

**Goal:** Ensure middleware is easy to write and integrate, and testable.

**Notes for Developer:**
*   `MiddlewareFunc` type is standard.
*   Provided middlewares (`LoggingMiddleware`, `AuthMiddleware`, etc.) use `Context` methods.

**Tasks:**

1.  **Task 7.1: Review Middleware Interactions with `Context`**
    *   **File:** `core/middleware.go`
    *   **Instruction:** Ensure that `Context` methods used by middleware (e.g., `ctx.sendSimpleText`, `ctx.getPermissionContext`, `ctx.UserID`) are stable and testable after `Context` and `Bot` refactoring. If `AuthMiddleware` needs `AccessManager`, it's passed during `AuthMiddleware(accessManager)` creation, which is good.

## 4. General Code Quality Tasks (Apply across all files)

1.  **Task G1: Consistent Naming:** Ensure names for variables, functions, methods, interfaces, and structs are clear, descriptive, and follow Go conventions.
2.  **Task G2: Godoc Comments:** Add or improve Godoc comments for all exported types, functions, and methods. Explain purpose, parameters, and return values.
3.  **Task G3: Error Wrapping:** Use `fmt.Errorf` with `%w` to wrap errors where appropriate to provide context without losing the original error.
4.  **Task G4: Reduce Cyclomatic Complexity:** Identify and refactor functions/methods with high cyclomatic complexity.
5.  **Task G5: Unit Tests:** Write comprehensive unit tests for all refactored components, focusing on testing logic in isolation using mocked dependencies. Aim for high test coverage.
6.  **Task G6: Remove Dead/Unused Code:** Identify and remove any code that is no longer used after refactoring.
7.  **Task G7: Logging:** Standardize logging. Use structured logging if possible. Distinguish between debug, info, and error logs.

## 5. Implementation Order Suggestion

1.  Start with **Task Group 1 (Bot API Abstraction & Bot Initialization)** as this is foundational.
2.  Move to **Task Group 2 (Prompt Composition System)**, ensuring `PromptComposer` and its handlers use the new `TelegramClient` and DI for `TemplateManager`. Define `PromptKeyboardActions`.
3.  Tackle **Task Group 5 (Template Management)** to make `TemplateManager` injectable.
4.  Then, refactor **Task Group 3 (Flow Management System)**, as it depends on interfaces from prompt composition and potentially the `TelegramClient` for message cleaning.
5.  Refine **Task Group 4 (Context)** to align with the new DI patterns.
6.  Review **Task Group 6 (Builders)** and **Task Group 7 (Middleware)** for any necessary adjustments based on other changes.
7.  Apply **Task Group G (General Code Quality)** throughout the process and as a final pass.

## 6. Notes for the Developer Implementing This Plan

*   **Iterate:** This is a comprehensive plan. Implement and test in small, manageable chunks (per task or sub-task).
*   **Test Driven:** Consider writing tests before or alongside the refactoring of each component to verify behavior.
*   **Communication:** If any part of the plan is unclear or a better approach is found during implementation, discuss it.
*   **User API First:** Always keep the public API (`NewBot`, `HandleCommand`, etc.) in mind. Changes should simplify or maintain its current ease of use.
*   **Focus on Decoupling:** The primary goal of internal changes is to decouple components for better maintainability and testability. Interfaces and DI are your main tools.

This plan provides a roadmap. The developer should use their judgment to adapt and refine as they delve into the code.