# Integration Architecture Analysis: Step-Prompt-Process API

## Overview

This document analyzes how the proposed Step-Prompt-Process API will integrate with the existing TeleFlow architecture, specifically the FlowManager, Context, and Bot systems. The goal is to maintain the sophisticated features of the current system while providing the simplified developer experience outlined in the API redesign proposal.

## Current Architecture Analysis

### Existing Flow System Components

#### 1. FlowManager (core/flow.go)
- **Responsibilities**: Flow registration, user state tracking, step transitions
- **Key Methods**: `StartFlow()`, `HandleUpdate()`, `determineNextStep()`
- **State Management**: `UserFlowState` with current step tracking
- **Integration Points**: Bot.processUpdate(), middleware system

#### 2. Context System (core/context.go)
- **Responsibilities**: Request context, state management, response helpers
- **Key Methods**: `Set()`, `Get()`, `Reply()`, `StartFlow()`
- **Integration**: Bot API access, template rendering, keyboard management

#### 3. Bot System (core/bot.go)
- **Responsibilities**: Update routing, handler management, middleware coordination
- **Flow Integration**: `RegisterFlow()`, flow exit command handling
- **Middleware**: Type-specific middleware for different handler types

## Proposed Integration Architecture

### 1. New Flow Definition Layer

```go
// New simplified API layer
type NewFlowBuilder struct {
    // Wraps existing Flow but with simplified interface
    internalFlow *Flow
    currentStep  *NewStepBuilder
}

type NewStepBuilder struct {
    // Maps to existing FlowStep but with cleaner interface
    internalStep *FlowStep
    promptConfig *PromptConfig
    processFunc  ProcessFunc
}
```

**Integration Strategy:**
- New API acts as a facade over existing `Flow` and `FlowStep` structures
- Maintains full compatibility with existing `FlowManager`
- Translates new API calls to existing internal representations

### 2. FlowManager Integration

#### Current FlowManager.HandleUpdate() Flow:
```
Update → Validate Input → Execute Step Handler → Determine Next Step → Execute Start Handler
```

#### Enhanced FlowManager.HandleUpdate() Flow:
```
Update → Execute ProcessFunc → Handle ProcessResult → Render PromptConfig → Update State
```

**Key Changes:**

1. **Replace OnInput Handler with ProcessFunc**
   ```go
   // Current: OnInput handler processes input
   if err := currentStep.Handler(ctx, inputForHandler); err != nil {
       return true, err
   }
   
   // New: ProcessFunc returns ProcessResult
   result := currentStep.ProcessFunc(ctx, input, buttonClick)
   return fm.handleProcessResult(ctx, result, userState)
   ```

2. **Add ProcessResult Handling**
   ```go
   func (fm *FlowManager) handleProcessResult(ctx *Context, result ProcessResult, userState *UserFlowState) (bool, error) {
       // Render prompt if specified
       if result.Prompt != nil {
           if err := fm.renderPromptConfig(ctx, result.Prompt); err != nil {
               return true, err
           }
       }
       
       // Execute action
       switch result.Action {
       case ActionNextStep:
           return fm.advanceToNextStep(ctx, userState)
       case ActionGoToStep:
           return fm.goToSpecificStep(ctx, userState, result.TargetStep)
       case ActionRetry:
           return true, nil // Stay on current step
       case ActionCompleteFlow:
           return fm.completeFlow(ctx, userState)
       case ActionCancelFlow:
           return fm.cancelFlow(ctx, userState)
       }
   }
   ```

3. **Replace OnStart Handler with PromptConfig Rendering**
   ```go
   // Current: OnStart handler sends initial message
   if newStepConfiguration.StartHandler != nil {
       err := newStepConfiguration.StartHandler(ctx)
   }
   
   // New: Render PromptConfig for step
   if newStepConfiguration.PromptConfig != nil {
       err := fm.renderPromptConfig(ctx, newStepConfiguration.PromptConfig)
   }
   ```

### 3. Context System Integration

#### Enhanced Context Methods:
```go
// Add new method for PromptConfig rendering
func (c *Context) RenderPromptConfig(config *PromptConfig) error {
    message, err := c.evaluatePromptMessage(config)
    if err != nil {
        return err
    }
    
    keyboard, err := c.evaluatePromptKeyboard(config)
    if err != nil {
        return err
    }
    
    if config.Image != "" {
        return c.sendMessageWithImage(message, config.Image, keyboard)
    }
    
    return c.send(message, keyboard)
}

func (c *Context) evaluatePromptMessage(config *PromptConfig) (string, error) {
    switch msg := config.Message.(type) {
    case string:
        return msg, nil
    case func(*Context) string:
        return msg(c), nil
    default:
        return "", fmt.Errorf("invalid message type in PromptConfig")
    }
}

func (c *Context) evaluatePromptKeyboard(config *PromptConfig) (interface{}, error) {
    if config.Keyboard == nil {
        return nil, nil
    }
    
    keyboardData := config.Keyboard(c)
    return c.convertToTelegramKeyboard(keyboardData), nil
}
```

### 4. Bot System Integration

#### Flow Registration Enhancement:
```go
// Current registration
func (b *Bot) RegisterFlow(flow *Flow) {
    b.flowManager.RegisterFlow(flow)
}

// Enhanced registration - no change needed
// New API builds compatible Flow objects internally
```

#### Update Processing Enhancement:
```go
// No changes needed to Bot.processUpdate()
// FlowManager.HandleUpdate() handles the new API internally
```

## Implementation Phases

### Phase 1: Core Types and Interfaces
1. **Add New Types** (`StepBuilder`, `PromptConfig`, `ProcessResult`)
2. **Extend FlowStep** to support new fields
3. **Add ProcessResult handling** to FlowManager

### Phase 2: PromptConfig Rendering Engine
1. **Implement `FlowManager.renderPromptConfig()`**
2. **Add Context methods** for message/keyboard evaluation
3. **Integrate with existing template system**

### Phase 3: ProcessFunc Integration
1. **Replace OnInput flow** with ProcessFunc execution
2. **Implement ProcessResult action handling**
3. **Add ButtonClick data extraction**

### Phase 4: Builder API Implementation
1. **Implement NewFlowBuilder facade**
2. **Create StepBuilder with fluent interface**
3. **Add conversion to internal Flow representation**

## Compatibility Matrix

| Current Feature | New API Equivalent | Integration Method |
|----------------|-------------------|-------------------|
| `OnStart` handler | `PromptConfig` rendering | Replace with automatic rendering |
| `OnInput` handler | `ProcessFunc` | Replace with ProcessResult handling |
| `FlowValidatorFunc` | `ProcessFunc` validation | Integrate validation into ProcessFunc |
| Step transitions | `ProcessResult.Action` | Map actions to existing transition logic |
| Flow completion | `ActionCompleteFlow` | Use existing OnComplete handler |
| Flow cancellation | `ActionCancelFlow` | Use existing OnCancel handler |
| Context data flow | `ctx.Set()`/`ctx.Get()` | No changes needed |
| Middleware | Type-specific middleware | Adapt to ProcessFunc calls |

## Architecture Benefits

1. **Backward Compatibility**: Existing FlowManager and Context systems unchanged
2. **Incremental Adoption**: New API can coexist with current implementation
3. **Feature Preservation**: All current features (middleware, state, callbacks) retained
4. **Performance**: No additional overhead - facade pattern only
5. **Testability**: Existing test infrastructure remains valid

## Next Steps

1. **Validate this integration approach** with stakeholders
2. **Create detailed interface specifications** for new types
3. **Design PromptConfig rendering engine** architecture
4. **Specify ProcessFunc execution context** and error handling
5. **Plan implementation timeline** and testing strategy

## Questions for Further Clarification

1. **Image Handling**: Should base64 images be converted to file uploads for better performance?
2. **Template Integration**: How should PromptConfig message functions integrate with existing template system?
3. **Keyboard Generation**: Should we cache keyboard function results or execute on each render?
4. **Error Handling**: How should ProcessFunc errors be distinguished from validation errors?