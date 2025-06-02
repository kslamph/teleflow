# Flow Guide

A comprehensive guide to building multi-step conversations with Teleflow's flow system.

## Table of Contents

- [What are Flows?](#what-are-flows)
- [Flow Concepts](#flow-concepts)
- [Building Your First Flow](#building-your-first-flow)
- [Flow Builder DSL](#flow-builder-dsl)
- [Step Types and Validation](#step-types-and-validation)
- [Advanced Flow Patterns](#advanced-flow-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## What are Flows?

Flows are Teleflow's powerful system for managing multi-step conversations with users. Instead of handling each message individually, flows allow you to create structured, stateful interactions that guide users through complex processes like registration, surveys, ordering, or configuration.

### Why Use Flows?

- **Stateful Conversations**: Maintain context across multiple messages
- **User Guidance**: Guide users through complex multi-step processes
- **Input Validation**: Validate user input at each step
- **Flexible Transitions**: Create dynamic conversation paths
- **Automatic State Management**: Framework handles user state automatically

### When to Use Flows vs Regular Handlers

**Use Flows for:**
- User registration/onboarding
- Multi-step forms or surveys
- Configuration wizards
- Order placement processes
- Data collection workflows

**Use Regular Handlers for:**
- Simple commands (`/help`, `/status`)
- Single-response interactions
- Immediate actions
- Menu navigation

## Flow Concepts

### Core Components

#### Flow
A complete multi-step conversation with a name, steps, and completion handlers.

#### Steps
Individual stages in the conversation, each with its own handler and validation.

#### Transitions
Rules that determine how to move between steps based on user input.

#### Validation
Input validation functions that ensure user data meets requirements.

#### State
User data that persists throughout the flow execution.

### Flow Lifecycle

1. **Initiation**: Flow starts via [`ctx.StartFlow()`](../core/context.go:99)
2. **Step Execution**: Each step processes user input
3. **Validation**: Input is validated if validator is configured
4. **Transition**: Framework determines next step
5. **Completion**: Flow ends and completion handler runs

## Building Your First Flow

Let's create a simple user registration flow:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "regexp"
    "strings"
    teleflow "github.com/kslamph/teleflow/core"
)

func main() {
    bot, err := teleflow.NewBot(os.Getenv("BOT_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    // Create and register the registration flow
    registrationFlow := createRegistrationFlow()
    bot.RegisterFlow(registrationFlow)

    // Command to start registration
    bot.HandleCommand("register", func(ctx *teleflow.Context) error {
        return ctx.StartFlow("registration")
    })

    log.Println("Bot starting...")
    bot.Start()
}

func createRegistrationFlow() *teleflow.Flow {
    return teleflow.NewFlow("registration").
        Step("ask_name").
        OnInput(func(ctx *teleflow.Context) error {
            name := ctx.Update.Message.Text
            ctx.Set("name", name)
            return ctx.Reply("👤 Great! Now please enter your email address:")
        }).
        Step("ask_email").
        OnInput(func(ctx *teleflow.Context) error {
            email := ctx.Update.Message.Text
            ctx.Set("email", email)
            return ctx.Reply("📞 Finally, please enter your phone number:")
        }).
        WithValidator(emailValidator()).
        Step("ask_phone").
        OnInput(func(ctx *teleflow.Context) error {
            phone := ctx.Update.Message.Text
            ctx.Set("phone", phone)
            return ctx.Reply("✅ Registration completed! Saving your information...")
        }).
        WithValidator(phoneValidator()).
        OnComplete(func(ctx *teleflow.Context) error {
            name, _ := ctx.Get("name")
            email, _ := ctx.Get("email")
            phone, _ := ctx.Get("phone")
            
            response := fmt.Sprintf(
                "🎉 Registration Successful!\n\n"+
                "👤 Name: %s\n"+
                "📧 Email: %s\n"+
                "📞 Phone: %s\n\n"+
                "Welcome to our service!",
                name, email, phone,
            )
            return ctx.Reply(response)
        }).
        Build()
}

func emailValidator() teleflow.FlowValidatorFunc {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return func(input string) (bool, string) {
        if !emailRegex.MatchString(input) {
            return false, "❌ Please enter a valid email address (e.g., user@example.com)"
        }
        return true, ""
    }
}

func phoneValidator() teleflow.FlowValidatorFunc {
    return func(input string) (bool, string) {
        // Remove spaces and dashes
        phone := strings.ReplaceAll(strings.ReplaceAll(input, " ", ""), "-", "")
        if len(phone) < 10 {
            return false, "❌ Please enter a valid phone number (at least 10 digits)"
        }
        return true, ""
    }
}
```

## Flow Builder DSL

The Flow Builder provides a fluent interface for creating flows:

### Basic Structure

```go
flow := teleflow.NewFlow("flow_name").
    Step("step1").OnInput(handler1).
    Step("step2").OnInput(handler2).
    OnComplete(completionHandler).
    Build()
```

### Step Configuration

Each step can be configured with various options:

```go
step := flow.Step("step_name").
    OnInput(inputHandler).                          // Required: handles user input
    WithValidator(validatorFunc).                   // Optional: validates input
    NextStep("next_step_name").                    // Optional: explicit next step
    WithTimeout(5 * time.Minute).                  // Optional: step timeout
    WithStepType(teleflow.StepTypeChoice).         // Optional: step type
    StayOnInvalidInput()                           // Optional: behavior on invalid input
```

### Flow Configuration

```go
flow := teleflow.NewFlow("name").
    // ... steps ...
    OnComplete(completionHandler).                  // Optional: completion handler
    OnCancel(cancellationHandler).                 // Optional: cancellation handler
    Build()
```

## Step Types and Validation

### Step Types

Define the expected input type for better validation:

```go
const (
    StepTypeText         // Free text input (default)
    StepTypeChoice       // Choice from predefined options
    StepTypeConfirmation // Yes/No confirmation
    StepTypeCustom       // Custom validation logic
)
```

**Example:**
```go
step.WithStepType(teleflow.StepTypeChoice).
     WithValidator(teleflow.ChoiceValidator([]string{"Option A", "Option B", "Option C"}))
```

### Built-in Validators

#### NumberValidator
Validates numeric input:

```go
step.WithValidator(teleflow.NumberValidator())
```

#### ChoiceValidator
Validates input against allowed choices:

```go
choices := []string{"Small", "Medium", "Large"}
step.WithValidator(teleflow.ChoiceValidator(choices))
```

### Custom Validators

Create your own validation logic:

```go
func ageValidator() teleflow.FlowValidatorFunc {
    return func(input string) (bool, string) {
        age, err := strconv.Atoi(input)
        if err != nil {
            return false, "Please enter a valid number"
        }
        if age < 18 {
            return false, "You must be at least 18 years old"
        }
        if age > 120 {
            return false, "Please enter a realistic age"
        }
        return true, ""
    }
}

step.WithValidator(ageValidator())
```

### Complex Validation Example

```go
func creditCardValidator() teleflow.FlowValidatorFunc {
    return func(input string) (bool, string) {
        // Remove spaces and dashes
        card := strings.ReplaceAll(strings.ReplaceAll(input, " ", ""), "-", "")
        
        // Check length
        if len(card) < 13 || len(card) > 19 {
            return false, "Credit card number must be 13-19 digits"
        }
        
        // Check if all digits
        for _, char := range card {
            if char < '0' || char > '9' {
                return false, "Credit card number must contain only digits"
            }
        }
        
        // Luhn algorithm check
        if !luhnCheck(card) {
            return false, "Invalid credit card number"
        }
        
        return true, ""
    }
}
```

## Advanced Flow Patterns

### Conditional Branching

Create flows that branch based on user input:

```go
surveyFlow := teleflow.NewFlow("survey").
    Step("ask_experience").
    OnInput(func(ctx *teleflow.Context) error {
        experience := ctx.Update.Message.Text
        ctx.Set("experience", experience)
        
        if experience == "Beginner" {
            return ctx.Reply("Let's start with basic questions...")
        } else {
            return ctx.Reply("Great! Let's dive into advanced topics...")
        }
    }).
    WithValidator(teleflow.ChoiceValidator([]string{"Beginner", "Intermediate", "Advanced"})).
    Step("beginner_questions").
    OnInput(handleBeginnerQuestions).
    Step("advanced_questions").
    OnInput(handleAdvancedQuestions).
    Build()
```

### Dynamic Step Transitions

Use the Transitions map for complex routing:

```go
step := flow.Step("menu").
    OnInput(func(ctx *teleflow.Context) error {
        choice := ctx.Update.Message.Text
        ctx.Set("menu_choice", choice)
        return ctx.Reply("Processing your choice...")
    })

// Configure transitions in the step
step.step.Transitions = map[string]string{
    "Profile":  "edit_profile",
    "Settings": "edit_settings", 
    "Help":     "show_help",
    "Exit":     "", // Empty string means complete flow
}
```

### Keyboard Integration

Combine flows with keyboards for better UX:

```go
step.OnInput(func(ctx *teleflow.Context) error {
    keyboard := teleflow.NewReplyKeyboard()
    keyboard.AddButton("Small").AddButton("Medium").AddButton("Large").AddRow()
    keyboard.AddButton("Cancel Order").AddRow()
    keyboard.Resize()
    
    return ctx.Reply("Please select your size:", keyboard)
}).
WithValidator(teleflow.ChoiceValidator([]string{"Small", "Medium", "Large", "Cancel Order"}))
```

### Flow Nesting and Composition

Start sub-flows from within flows:

```go
mainFlow := teleflow.NewFlow("checkout").
    Step("confirm_order").
    OnInput(func(ctx *teleflow.Context) error {
        if ctx.Update.Message.Text == "Change Address" {
            // Start address flow, will return here when complete
            return ctx.StartFlow("address_entry")
        }
        // Continue with main flow
        return ctx.Reply("Order confirmed!")
    }).
    Build()

addressFlow := teleflow.NewFlow("address_entry").
    Step("ask_street").OnInput(handleStreet).
    Step("ask_city").OnInput(handleCity).
    OnComplete(func(ctx *teleflow.Context) error {
        // Return to main flow
        return ctx.Reply("Address updated! Returning to checkout...")
    }).
    Build()
```

### Data Persistence

Store and retrieve complex data:

```go
type UserProfile struct {
    Name     string
    Email    string
    Preferences map[string]string
}

step.OnInput(func(ctx *teleflow.Context) error {
    // Retrieve existing profile
    var profile UserProfile
    if data, ok := ctx.Get("profile"); ok {
        profile = data.(UserProfile)
    } else {
        profile = UserProfile{Preferences: make(map[string]string)}
    }
    
    // Update profile
    profile.Name = ctx.Update.Message.Text
    ctx.Set("profile", profile)
    
    return ctx.Reply("Profile updated!")
})
```

## Best Practices

### Flow Design

1. **Keep Steps Focused**: Each step should have one clear purpose
2. **Provide Clear Instructions**: Tell users exactly what input you expect
3. **Show Progress**: Let users know where they are in the process
4. **Allow Cancellation**: Always provide an exit option
5. **Handle Errors Gracefully**: Provide helpful error messages

### User Experience

#### Progress Indicators
```go
func showProgress(ctx *teleflow.Context, current, total int) error {
    progress := fmt.Sprintf("Step %d of %d", current, total)
    // Include progress in your message
    return ctx.Reply(fmt.Sprintf("📝 %s\n\nPlease enter your email:", progress))
}
```

#### Exit Options
```go
// Configure global exit commands
bot, err := teleflow.NewBot(token,
    teleflow.WithFlowConfig(teleflow.FlowConfig{
        ExitCommands: []string{"/cancel", "/exit", "/stop"},
        ExitMessage:  "❌ Operation cancelled. Type /help for available commands.",
    }),
)
```

#### Input Hints
```go
step.OnInput(func(ctx *teleflow.Context) error {
    keyboard := teleflow.NewReplyKeyboard()
    keyboard.AddButton("john@example.com").AddRow() // Example email
    keyboard.Placeholder("Enter your email address")
    
    return ctx.Reply("📧 Please enter your email address:", keyboard)
})
```

### Error Handling

#### Validation Messages
```go
func emailValidator() teleflow.FlowValidatorFunc {
    return func(input string) (bool, string) {
        if input == "" {
            return false, "❌ Email cannot be empty.\n💡 Please enter your email address."
        }
        if !isValidEmail(input) {
            return false, "❌ Invalid email format.\n💡 Please use format: user@example.com"
        }
        return true, ""
    }
}
```

#### Timeout Handling
```go
step.WithTimeout(2 * time.Minute).
     OnInput(func(ctx *teleflow.Context) error {
         // Handle timeout in your flow logic
         return ctx.Reply("⏰ Please respond within 2 minutes to continue.")
     })
```

### Performance Considerations

1. **Minimize State Size**: Only store necessary data in context
2. **Clean Up State**: Remove unnecessary data as flow progresses
3. **Set Appropriate Timeouts**: Prevent flows from hanging indefinitely
4. **Validate Early**: Catch invalid input as soon as possible

### Testing Flows

#### Unit Testing Steps
```go
func TestRegistrationStep(t *testing.T) {
    // Create mock context
    ctx := &teleflow.Context{}
    ctx.Set("name", "John")
    
    // Test step handler
    err := emailStepHandler(ctx)
    
    // Assert results
    assert.NoError(t, err)
    email, ok := ctx.Get("email")
    assert.True(t, ok)
    assert.Equal(t, "john@example.com", email)
}
```

#### Integration Testing
```go
func TestCompleteFlow(t *testing.T) {
    // Create test bot
    bot := createTestBot()
    
    // Register flow
    flow := createRegistrationFlow()
    bot.RegisterFlow(flow)
    
    // Simulate user interaction
    simulateUserInput(bot, "/register")
    simulateUserInput(bot, "John Doe")
    simulateUserInput(bot, "john@example.com")
    
    // Assert flow completion
    assertFlowCompleted(t, bot)
}
```

## Troubleshooting

### Common Issues

#### Flow Not Starting
```go
// ❌ Wrong: Flow not registered
ctx.StartFlow("unregistered_flow") // Will fail

// ✅ Correct: Register flow first
bot.RegisterFlow(myFlow)
ctx.StartFlow("my_flow")
```

#### Steps Not Linking
```go
// ❌ Wrong: Missing NextStep or Build
flow.Step("step1").OnInput(handler1).
     Step("step2").OnInput(handler2) // Missing Build()

// ✅ Correct: Proper step linking
flow.Step("step1").OnInput(handler1).
     Step("step2").OnInput(handler2).
     Build()
```

#### Validation Not Working
```go
// ❌ Wrong: Validator after NextStep
step.NextStep("next").WithValidator(validator)

// ✅ Correct: Validator before NextStep
step.WithValidator(validator).NextStep("next")
```

#### State Not Persisting
```go
// ❌ Wrong: Not storing in context
func handler(ctx *teleflow.Context) error {
    name := ctx.Update.Message.Text
    // name is lost after this handler
    return ctx.Reply("Got it!")
}

// ✅ Correct: Store in context
func handler(ctx *teleflow.Context) error {
    name := ctx.Update.Message.Text
    ctx.Set("name", name) // Persists across steps
    return ctx.Reply("Got it!")
}
```

### Debugging Tips

#### Enable Flow Logging
```go
bot.Use(teleflow.LoggingMiddleware()) // Logs all flow transitions
```

#### Check Flow State
```go
func debugHandler(ctx *teleflow.Context) error {
    // Log current state
    log.Printf("User %d state: %+v", ctx.UserID(), ctx.data)
    return ctx.Reply("Debug info logged")
}
```

#### Validate Flow Structure
```go
func validateFlow(flow *teleflow.Flow) error {
    for _, step := range flow.Steps {
        if step.Handler == nil {
            return fmt.Errorf("step %s missing handler", step.Name)
        }
    }
    return nil
}
```

### Performance Monitoring

```go
func flowTimingMiddleware() teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            start := time.Now()
            err := next(ctx)
            duration := time.Since(start)
            
            if duration > 5*time.Second {
                log.Printf("Slow flow step for user %d: %v", ctx.UserID(), duration)
            }
            
            return err
        }
    }
}
```

## Complete Example: E-commerce Order Flow

Here's a comprehensive example showing many advanced concepts:

```go
func createOrderFlow() *teleflow.Flow {
    return teleflow.NewFlow("place_order").
        Step("select_category").
        OnInput(func(ctx *teleflow.Context) error {
            category := ctx.Update.Message.Text
            ctx.Set("category", category)
            
            keyboard := teleflow.NewReplyKeyboard()
            switch category {
            case "Electronics":
                keyboard.AddButton("Phone").AddButton("Laptop").AddButton("Tablet")
            case "Clothing":
                keyboard.AddButton("T-Shirt").AddButton("Jeans").AddButton("Shoes")
            case "Books":
                keyboard.AddButton("Fiction").AddButton("Non-Fiction").AddButton("Technical")
            }
            keyboard.AddRow().AddButton("🔙 Back").Resize()
            
            return ctx.Reply(fmt.Sprintf("Great! What %s item would you like?", category), keyboard)
        }).
        WithValidator(teleflow.ChoiceValidator([]string{"Electronics", "Clothing", "Books"})).
        
        Step("select_item").
        OnInput(func(ctx *teleflow.Context) error {
            item := ctx.Update.Message.Text
            if item == "🔙 Back" {
                return ctx.Reply("Please select a category:", getCategoryKeyboard())
            }
            
            ctx.Set("item", item)
            return ctx.Reply("How many would you like? (1-10)")
        }).
        
        Step("select_quantity").
        OnInput(func(ctx *teleflow.Context) error {
            quantity := ctx.Update.Message.Text
            ctx.Set("quantity", quantity)
            
            // Calculate total
            price := getItemPrice(ctx.Get("item").(string))
            qty, _ := strconv.Atoi(quantity)
            total := price * float64(qty)
            ctx.Set("total", total)
            
            keyboard := teleflow.NewInlineKeyboard()
            keyboard.AddButton("✅ Confirm Order", "confirm_order")
            keyboard.AddButton("❌ Cancel", "cancel_order")
            
            summary := fmt.Sprintf(
                "📋 Order Summary:\n"+
                "🏷️ Category: %s\n"+
                "📦 Item: %s\n"+
                "🔢 Quantity: %s\n"+
                "💰 Total: $%.2f\n\n"+
                "Confirm your order?",
                ctx.Get("category"), ctx.Get("item"), quantity, total,
            )
            
            return ctx.Reply(summary, keyboard)
        }).
        WithValidator(teleflow.NumberValidator()).
        
        OnComplete(func(ctx *teleflow.Context) error {
            return ctx.Reply("🎉 Order placed successfully! You'll receive a confirmation email shortly.")
        }).
        
        OnCancel(func(ctx *teleflow.Context) error {
            return ctx.Reply("❌ Order cancelled. Feel free to browse again anytime!")
        }).
        
        Build()
}
```

This comprehensive guide covers everything you need to know about building effective flows with Teleflow. Start with simple flows and gradually incorporate advanced patterns as your bot's complexity grows.