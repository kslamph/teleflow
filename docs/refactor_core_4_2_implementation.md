# Context Dependency Refactoring Implementation (Subtask 4.2)

**Date:** 2025-06-08  
**Objective:** Replace `Context.bot *Bot` with more specific dependencies via interfaces to improve testability and decouple components.

## Summary of Changes

### 1. New Interface Definition

**File:** `core/interfaces.go`

Added `ContextFlowOperations` interface to define flow-related operations that `Context` needs:

```go
// ContextFlowOperations defines methods for interacting with user flows from the context.
type ContextFlowOperations interface {
    // SetUserFlowData sets flow-specific data for a user.
    setUserFlowData(userID int64, key string, value interface{}) error
    // GetUserFlowData retrieves flow-specific data for a user.
    getUserFlowData(userID int64, key string) (interface{}, bool)
    // StartFlow starts a flow for a user.
    startFlow(userID int64, flowName string, ctx *Context) error
    // IsUserInFlow checks if a user is currently in a flow.
    isUserInFlow(userID int64) bool
    // CancelFlow cancels the current flow for a user.
    cancelFlow(userID int64)
}
```

### 2. Context Struct Refactoring

**File:** `core/context.go`

#### Before:
```go
type Context struct {
    bot    *Bot
    update tgbotapi.Update
    data   map[string]interface{}
    // ... other fields
}
```

#### After:
```go
type Context struct {
    telegramClient     TelegramClient
    templateManager    TemplateManager
    flowOps            ContextFlowOperations
    promptSender       PromptSender
    accessManager      AccessManager
    
    update tgbotapi.Update
    data   map[string]interface{}
    // ... other fields
}
```

### 3. Constructor Update

**File:** `core/context.go`

#### Before:
```go
func newContext(bot *Bot, update tgbotapi.Update) *Context
```

#### After:
```go
func newContext(
    update tgbotapi.Update,
    client TelegramClient,
    tm TemplateManager,
    fo ContextFlowOperations,
    ps PromptSender,
    am AccessManager,
) *Context
```

### 4. Method Updates

All `Context` methods were updated to use the new interface dependencies:

#### Flow Operations:
- `SetFlowData()`: `c.bot.flowManager.setUserFlowData()` → `c.flowOps.setUserFlowData()`
- `GetFlowData()`: `c.bot.flowManager.getUserFlowData()` → `c.flowOps.getUserFlowData()`
- `StartFlow()`: `c.bot.flowManager.startFlow()` → `c.flowOps.startFlow()`
- `isUserInFlow()`: `c.bot.flowManager.isUserInFlow()` → `c.flowOps.isUserInFlow()`
- `CancelFlow()`: `c.bot.flowManager.cancelFlow()` → `c.flowOps.cancelFlow()`

#### Prompt Operations:
- `SendPrompt()`: `c.bot.promptComposer.ComposeAndSend()` → `c.promptSender.ComposeAndSend()`

#### Template Operations:
- `AddTemplate()`: `defaultTemplateManager.AddTemplate()` → `c.templateManager.AddTemplate()`
- `GetTemplateInfo()`: `defaultTemplateManager.GetTemplateInfo()` → `c.templateManager.GetTemplateInfo()`
- `ListTemplates()`: `defaultTemplateManager.ListTemplates()` → `c.templateManager.ListTemplates()`
- `HasTemplate()`: `defaultTemplateManager.HasTemplate()` → `c.templateManager.HasTemplate()`
- `RenderTemplate()`: `defaultTemplateManager.RenderTemplate()` → `c.templateManager.RenderTemplate()`
- `TemplateManager()`: `defaultTemplateManager` → `c.templateManager`

#### Telegram API Operations:
- `answerCallbackQuery()`: `c.bot.api.Request()` → `c.telegramClient.Request()`
- `sendSimpleText()`: `c.bot.api.Send()` → `c.telegramClient.Send()`

#### Access Management:
- `getPermissionContext()`: `c.bot.accessManager` → `c.accessManager`

### 5. Bot Integration Update

**File:** `core/bot.go`

Updated `Bot.processUpdate()` to provide the appropriate implementations:

#### Before:
```go
ctx := newContext(b, update)
```

#### After:
```go
ctx := newContext(update, b.api, b.templateManager, b.flowManager, b.promptComposer, b.accessManager)
```

## Interface Implementations

The refactoring leverages existing implementations:

1. **`TelegramClient`**: Implemented by `*tgbotapi.BotAPI` (existing interface)
2. **`TemplateManager`**: Implemented by `*templateManager` (existing interface)
3. **`ContextFlowOperations`**: Implemented by `*flowManager` (new interface, existing implementation)
4. **`PromptSender`**: Implemented by `*PromptComposer` (existing interface)
5. **`AccessManager`**: Implemented by user-provided access managers (existing interface)

## Benefits Achieved

1. **Improved Testability**: `Context` can now be unit tested with mock implementations of each interface
2. **Better Separation of Concerns**: `Context` only depends on the specific operations it needs
3. **Reduced Coupling**: `Context` no longer depends on the entire `Bot` struct
4. **Dependency Injection**: Clear, explicit dependencies make the code more maintainable
5. **Interface-based Design**: Makes the system more flexible and extensible

## Validation

- ✅ All code compiles without errors
- ✅ All existing tests pass
- ✅ golangci-lint passes without issues
- ✅ No breaking changes to public API
- ✅ Backward compatibility maintained for library users

## Next Steps

This refactoring aligns with Task Group 4 from the comprehensive refactoring plan and enables:

1. Better unit testing of `Context` methods in isolation
2. Easier mocking of dependencies for handler testing
3. Cleaner dependency management throughout the system
4. Foundation for further refactoring of related components

The refactoring successfully decouples `Context` from `Bot` while maintaining all existing functionality and improving the overall architecture of the teleflow library.