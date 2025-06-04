# Teleflow Improvement Recommendations

This document outlines potential areas for improvement in the Teleflow Go package, focusing on enhancing developer experience, reducing boilerplate, and clarifying design patterns.

## 1. Flow Step UI Automation

*   **Current:** For `FlowStepTypeChoice` and `FlowStepTypeConfirmation`, developers need to implement a `StartHandler` that manually creates and sends the appropriate Telegram keyboard (inline or reply) using `ctx.Reply()` or similar methods.
*   **Recommendation:** Enhance `FlowStep` to optionally include prompt text and choice definitions directly. The `FlowManager` could then automatically generate and send the necessary keyboard when such a step is entered.
    *   **Example:**
        ```go
        // Hypothetical enhanced FlowStep definition
        flow.Step("confirm_action").
            WithType(teleflow.StepTypeConfirmation).
            WithPrompt("Are you sure you want to proceed?").
            WithChoices([]string{"Yes, proceed", "No, cancel"}). // Or specific Yes/No button text
            WithKeyboardType(teleflow.KeyboardTypeInline) // or KeyboardTypeReply
        ```
    *   This would significantly reduce boilerplate for common interactive flow steps. The `FlowManager` or the step's internal logic would handle sending the message with the keyboard.

## 2. Structured User-Facing Errors

*   **Current:** Errors returned from `HandlerFunc` or middleware are often logged, and a generic "An error occurred" message might be sent to the user by the core `processUpdate` logic or by middleware like `RecoveryMiddleware`. `AuthMiddleware` replies with the error message from `AccessManager.CheckPermission`.
*   **Recommendation:** Introduce a specific error type, e.g., `teleflow.UserError`, that handlers or middleware can return.
    *   If `UserError` is returned, its message should be sent directly to the user.
    *   Other error types would still result in a generic "An error occurred" message (or be handled by `RecoveryMiddleware`).
    *   This would give developers finer-grained control over user-facing error messages without needing to call `ctx.Reply("custom error")` and then `return nil` to signify a "handled" error.
    *   `AccessManager.CheckPermission` could be designed to return `UserError` for permission denied messages.

## 3. Enhanced Bot Configuration Options

*   **Current:** `NewBot` initializes some components like `StateManager` to `NewInMemoryStateManager()` by default. Customizing these requires modifying bot fields after creation.
*   **Recommendation:**
    *   Introduce a `WithStateManager(StateManager)` `BotOption` to allow setting a custom state manager during `NewBot` initialization, similar to `WithAccessManager`. This would make `bot.stateManager` and `bot.flowManager.stateManager` consistent.
    *   Make the list of "allowed global commands during flows" more configurable. Currently, only `HelpCommands` (defined in `FlowConfig`) are explicitly checked in `resolveGlobalCommand`. This could be expanded in `FlowConfig` or via a new `BotOption` to allow developers to specify other commands that should bypass active flow progression.

## 4. Flow Data Passing and Initial Data

*   **Current:** `FlowManager.StartFlow` now accepts `ctx *Context`, and initial data can be populated from `ctx.data`. Data within a flow is managed via `UserFlowState.Data` and accessed/modified via `ctx.Set/Get`.
*   **Recommendation (Documentation & Minor API Refinement):**
    *   Clearly document the pattern for passing initial data to a flow using the context.
    *   Consider if a more explicit `initialData map[string]interface{}` parameter to `StartFlow` (in addition to `ctx`) could be beneficial for clarity in some use cases, though the current context-based approach is flexible. The `FlowManager` would then merge this with `ctx.data`.
    *   Ensure robust documentation on how data is persisted and accessible across flow steps.

## 5. AccessManager Granularity and Patterns

*   **Current:** `AccessManager.CheckPermission(ctx *PermissionContext)` is flexible but can lead to large switch statements in implementations if many distinct permissions exist. `PermissionContext` includes `Command`, `Arguments`, etc.
*   **Recommendation (Pattern/Documentation Focus):**
    *   Document best practices for implementing `AccessManager`. Encourage using `PermissionContext.Command` as a primary scope.
    *   Suggest patterns for handling more granular permissions, perhaps by defining action strings/enums that can be passed in `ctx.data` and checked by the `AccessManager`, or by using `PermissionContext.Arguments`.
    *   While more specific methods on `AccessManager` (e.g., `CanAccessCommand`) could be added, this might make the interface too broad. Focusing on patterns for using the existing `PermissionContext` effectively might be better.

## 6. Template System Clarity

*   **Current:** The template system is powerful, with parse-mode-specific functions injected at execution time by cloning templates.
*   **Recommendation (Documentation Focus):**
    *   Clearly document how `ParseMode` affects template rendering and available functions (especially the `escape` function's behavior).
    *   Provide comprehensive examples of using templates with different parse modes and custom functions.
    *   The internal complexity of cloning templates for function injection is an implementation detail that works; the focus should be on clear usage documentation.

## 7. Documentation for Handler Types

*   **Current:** The framework uses `HandlerFunc`, `CallbackHandler` (interface), and flow step handlers (`StartHandler`, `Handler` which are `HandlerFunc`).
*   **Recommendation (Documentation Focus):**
    *   Create a dedicated documentation section explaining the different types of handlers, their signatures, when they are used, and how they receive data (e.g., `CallbackHandler.Handle` gets extracted `data`, `HandlerFunc` in flows might use `ctx.Update.Message.Text`).
    *   Clarify the role of the default text handler (`HandleText("", handler)`).

This list provides a starting point for future enhancements to make Teleflow even more robust and developer-friendly.