# TeleFlow API Redesign Proposal

## Executive Summary

Based on user feedback indicating that the current flow system lacks clarity and has a steep learning curve, this proposal outlines a simplified API design that achieves zero learning curve while maintaining powerful functionality.

## Current Pain Points

1. **Unclear separation of concerns**: OnStart vs OnInput handlers create confusion
2. **Complex state management**: Users must understand internal state handling
3. **Boilerplate code**: Keyboard creation, callback handling requires technical knowledge
4. **Steep learning curve**: Current builder pattern is not intuitive

## Proposed Simplified API

### Core Concept: Step-Prompt-Process Pattern

```go
teleflow.Step("step_1").
    Prompt(message, image, keyboard).
    Process(func (input) processResult).
    OnComplete(func(ctx *teleflow.Context, flowData map[string]interface{}))
```

## Detailed Design Specification

### 1. Step Creation
```go
type StepBuilder struct {
    name string
    prompt *PromptConfig
    processor ProcessFunc
    // internal fields hidden from user
}

func Step(name string) *StepBuilder
```

### 2. Prompt Configuration
```go
type PromptConfig struct {
    Message  interface{}  // string or template function
    Image    string      // base64 encoded image (optional)
    Keyboard KeyboardFunc // function returning keyboard config (optional)
}

type KeyboardFunc func(ctx *Context) map[string]interface{}

func (sb *StepBuilder) Prompt(message interface{}, image string, keyboard KeyboardFunc) *StepBuilder
```

**Message Templates Support:**
```go
// Simple string
.Prompt("What's your name?", "", nil)

// Template with context data
.Prompt(func(ctx *Context) string {
    return fmt.Sprintf("Hello %s, what's your age?", ctx.Get("name"))
}, "", nil)

// Template accessing external services (e.g., a database via userService)
// Assumes 'userService' is an instance of your service struct/interface,
// appropriately injected or accessible within the scope where the flow is defined.
.Prompt(func(ctx *Context) string {
    userID, ok := ctx.Get("user_id").(string)
    if !ok {
        // Log this issue appropriately in a real application
        log.Println("User ID not found or not a string in context for prompt generation")
        return "Hello! Please tell us a bit about yourself."
    }

    // Example: Fetching user data from a hypothetical userService
    userProfile, err := userService.GetProfile(ctx, userID) // Pass context for potential cancellation
    if err != nil {
        log.Printf("Error fetching profile for user %s: %v", userID, err)
        // Fallback message if DB call fails
        return "Welcome! We had trouble fetching your details at the moment."
    }
    
    // Example: Incrementing a visit count (write operation during prompt formulation)
    // This demonstrates that complex logic, including writes, can occur here if needed,
    // though typically prompts might focus more on reads.
    updatedVisitCount, err := userService.IncrementVisitCount(ctx, userID)
    if err != nil {
        log.Printf("Error incrementing visit count for user %s: %v", userID, err)
        // Non-critical error, proceed with prompt using potentially stale visit count from profile
        return fmt.Sprintf("Welcome back, %s! You have %d loyalty points.",
            userProfile.Name, userProfile.LoyaltyPoints)
    }

    return fmt.Sprintf("Welcome back, %s! This is visit #%d. You have %d loyalty points.",
        userProfile.Name, updatedVisitCount, userProfile.LoyaltyPoints)
}, "", nil) // Image and Keyboard are nil in this example
```

### 3. Input Processing

The `ProcessFunc` determines the outcome of a step after user input. It returns a `ProcessResult` which dictates the next action and can optionally include a `PromptConfig` to be rendered.

```go
// ProcessResult represents the outcome of processing user input in a flow step.
// It contains the action to take and an optional prompt to render.
type ProcessResult struct {
    Action     FlowAction
    TargetStep string        // Used only when Action is GoToStep
    Prompt     *PromptConfig // Optional prompt to render before executing the action
}

// FlowAction defines the possible flow control actions.
type FlowAction int

const (
    ActionNextStep FlowAction = iota // Continue to the next step in sequence
    ActionGoToStep                   // Jump to a specific named step
    ActionRetry                      // Stay on current step (typically for validation)
    ActionCancelFlow                 // Terminate the flow
    ActionCompleteFlow               // Mark flow as successfully completed
)

// WithPrompt sets a prompt to be rendered before the action is executed.
func (pr ProcessResult) WithPrompt(prompt PromptConfig) ProcessResult {
    pr.Prompt = &prompt
    return pr
}

// --- Action Constructor Functions ---

// NextStep proceeds to the next step in the defined sequence.
func NextStep() ProcessResult {
    return ProcessResult{Action: ActionNextStep}
}

// GoToStep transitions to a specific named step.
func GoToStep(stepName string) ProcessResult {
    return ProcessResult{Action: ActionGoToStep, TargetStep: stepName}
}

// Retry re-prompts the current step. Typically used with WithPrompt() for validation messages.
func Retry() ProcessResult {
    return ProcessResult{Action: ActionRetry}
}

// CancelFlow terminates the current flow.
func CancelFlow() ProcessResult {
    return ProcessResult{Action: ActionCancelFlow}
}

// CompleteFlow marks the flow as successfully completed.
func CompleteFlow() ProcessResult {
    return ProcessResult{Action: ActionCompleteFlow}
}

// ButtonClick holds data from a button interaction.
type ButtonClick struct {
    Caption      string
    CallbackData interface{}
}

// ProcessFunc defines the signature for processing user input.
// It must return one of the action constructor functions (NextStep(), GoToStep(), etc.),
// optionally chained with .WithPrompt(PromptConfig{...}).
type ProcessFunc func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult

func (sb *StepBuilder) Process(processor ProcessFunc) *StepBuilder
```

**Behavior of `WithPrompt()`:**
If a `ProcessResult` is returned with `WithPrompt(promptConfig)`, the specified `promptConfig` is *always* rendered to the user.
- For `NextStep()`, `GoToStep()`, `CompleteFlow()`, and `CancelFlow()`: The prompt is rendered, and then the respective action is executed. This allows for a final message or a message before transitioning.
- For `Retry()`: The prompt is rendered, and the flow remains on the current step, awaiting new input. This is the primary mechanism for validation feedback.
If `WithPrompt()` is not called, no additional prompt is rendered by the action itself (the standard prompt for the next step will render as usual if transitioning).
**Process Function Examples:**
```go
// Simple validation and storage
.Process(func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
    if buttonClick != nil {
        // Handle button click
        ctx.Set("shipping_method", buttonClick.CallbackData)
        return NextStep()
    }
    
    // Handle text input
    if len(input) < 2 {
        // Re-prompt the current step with a validation message
        return Retry().WithPrompt(PromptConfig{
            Message: "Your name is too short (minimum 2 characters). Please enter your full name:",
        })
    }
    
    ctx.Set("user_name", input)
    return NextStep()
})
})
// Conditional branching
.Process(func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
    if buttonClick != nil {
        switch buttonClick.CallbackData.(string) {
        case "premium":
            return GoToStep("premium_options")
        case "basic":
            return GoToStep("basic_options")
        case "cancel":
            return CancelFlow()
        }
    }
    // Default if no button click matched or if it's text input not handled above
    return NextStep()
})
```

### 4. Enhanced Features

#### A. Context Management
```go
// Built-in context operations available in Process functions
ctx.Set(key, value)
ctx.Get(key)
// ctx.Reply(message) // Removed: Use PromptConfig return from ProcessFunc for messages
// ctx.EditLastMessage(message)
// ctx.SendImage(base64Image) // Removed: Image is part of PromptConfig, rendered by the engine.
```

#### B. Dynamic Keyboard Generation
```go
.Prompt("Choose shipping method:", "", func(ctx *Context) map[string]interface{} {
    // Access external service
    methods := shippingService.GetAvailableMethods(ctx.Get("location"))
    
    keyboard := make(map[string]interface{})
    for _, method := range methods {
        keyboard[method.Name] = map[string]interface{}{
            "callback_data": method.ID,
            "price": method.Price,
        }
    }
    return keyboard
})

**Note on Prompt Rendering:** The Teleflow engine is responsible for rendering the `PromptConfig`. When a step begins, or when a `ProcessFunc` returns a `PromptConfig` (for a retry/re-prompt), the engine will:
1. Evaluate the `Message` field: If it's a function, execute it with the current `Context` to get the string. If it's already a string, use it directly. This message string can be processed by a templating engine similar to `core/templates.go` if desired, or used as plain text.
2. Send the `Image` (if provided in `PromptConfig`) as a base64 encoded image.
3. Construct and send the `Keyboard` (if `KeyboardFunc` is provided and returns a valid keyboard map).
This ensures that the developer only needs to define *what* to show in `PromptConfig`, and the package handles *how* to show it.
```

#### C. Transitioning to Specific Steps

To transition to a specific step by name from within a `ProcessFunc`, use the `GoToStep()` action constructor:

```go
.Process(func(ctx *Context, input string, btn *ButtonClick) ProcessResult {
    if input == "admin" {
        return GoToStep("admin_dashboard_step")
    }
    return NextStep()
})
```
The `GoToStep("target_step_name")` function returns a `ProcessResult` that the flow engine interprets.

## Implementation Strategy

### Phase 1: Core API
1. Create new simplified types (`StepBuilder`, `PromptConfig`, `ProcessResult`)
2. Implement basic Step-Prompt-Process chain
3. Template system for dynamic messages
4. ProcessResult handling with flow control

### Phase 2: Enhanced Features  
1. Dynamic keyboard generation
2. Base64 image support
3. Context manipulation functions
4. Button click handling with callback data

## Benefits of New Design

### 1. Zero Learning Curve
- Clear separation: Prompt shows, Process handles
- Intuitive flow: Step → Prompt → Process → Result
- No need to understand internal state management

### 2. Powerful Yet Simple
- Template system handles dynamic content
- ProcessResult covers all flow control needs
- Built-in context operations eliminate boilerplate

### 3. Flexible Architecture
- Dynamic keyboard generation from external services
- Conditional branching through ProcessResult
- Template functions can access any data source

### 4. Developer Experience
- IDE autocompletion for all operations
- Type safety with ProcessResult
- Clear error messages and validation

## Example: Complete Flow

```go
flow := teleflow.NewFlow("delivery_onboarding").
    Step("welcome").
        Prompt("Welcome! What's your name?", "", nil).
        Process(func(ctx *Context, input string, btn *ButtonClick) ProcessResult {
            if len(input) < 2 {
                return Retry().WithPrompt(PromptConfig{
                    Message: "Your name is too short. Please enter a valid name (at least 2 characters):",
                })
            }
            ctx.Set("name", input)
            return NextStep()
        }).
    
    Step("get_location").
        Prompt(func(ctx *Context) string {
            return fmt.Sprintf("Hi %s! Please enter your delivery location (address or landmark):", ctx.Get("name"))
        }, "", nil).
        Process(func(ctx *Context, input string, btn *ButtonClick) ProcessResult {
            if len(input) < 5 {
                return Retry().WithPrompt(PromptConfig{
                    Message: "Please provide a more detailed location (at least 5 characters):",
                })
            }
            
            // Validate location using external service
            location, isServiceable, err := locationService.ValidateLocation(ctx, input)
            if err != nil {
                return Retry().WithPrompt(PromptConfig{
                    Message: "Sorry, we couldn't verify that location. Please try again:",
                })
            }
            
            if !isServiceable {
                return CancelFlow().WithPrompt(PromptConfig{
                    Message: "Sorry, we don't currently deliver to that location. Please try us again when we expand to your area!",
                })
            }
            
            // Store validated location data
            ctx.Set("location", location.Address)
            ctx.Set("location_coords", location.Coordinates)
            ctx.Set("location_map_image", location.MapImageBase64)
            return NextStep()
        }).
    
    Step("confirm_location").
        Prompt(func(ctx *Context) string {
            return fmt.Sprintf("Is this your correct delivery location?\n\n📍 %s", ctx.Get("location"))
        }, func(ctx *Context) string {
            // Return base64 encoded map image
            return ctx.Get("location_map_image").(string)
        }, func(ctx *Context) map[string]interface{} {
            return map[string]interface{}{
                "✅ Yes, that's correct": "yes",
                "❌ No, let me re-enter": "no",
            }
        }).
        Process(func(ctx *Context, input string, btn *ButtonClick) ProcessResult {
            if btn != nil {
                switch btn.CallbackData {
                case "yes":
                    return NextStep()
                case "no":
                    return GoToStep("get_location")
                }
            }
            return Retry().WithPrompt(PromptConfig{
                Message: "Please use the buttons to confirm your location.",
            })
        }).
    
    Step("shipping_preference").
        Prompt("Great! Now choose your preferred delivery option:", "", func(ctx *Context) map[string]interface{} {
            // Dynamic shipping options based on location
            options := shippingService.GetAvailableOptions(ctx.Get("location_coords"))
            keyboard := make(map[string]interface{})
            for _, option := range options {
                keyboard[fmt.Sprintf("%s - $%.2f", option.Name, option.Price)] = option.ID
            }
            return keyboard
        }).
        Process(func(ctx *Context, input string, btn *ButtonClick) ProcessResult {
            if btn != nil {
                ctx.Set("shipping_option", btn.CallbackData)
                return CompleteFlow().WithPrompt(PromptConfig{
                    Message: "Perfect! Your delivery preferences have been saved. We'll contact you soon!",
                })
            }
            return Retry().WithPrompt(PromptConfig{
                Message: "Please select a delivery option from the menu above.",
            })
        }).
    
    OnComplete(func(ctx *Context, flowData map[string]interface{}) error {
        // Save delivery setup
        return deliveryService.CreateDeliveryProfile(flowData)
    }).
    
    Build()
```

## Recommendation

This redesign achieves the stated goal of zero learning curve while maintaining all the power of the current system. The clear Prompt/Process separation eliminates confusion, and the ProcessResult provides intuitive flow control.

The template system and dynamic keyboard generation hide complexity while enabling powerful integrations with external services, making this API both beginner-friendly and enterprise-ready.