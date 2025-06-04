# Teleflow Conversational Flow Guide

Teleflow's conversational flow system is a powerful feature for creating structured, multi-step interactions with users. It allows you to guide users through processes like registration, surveys, data collection, or complex command sequences while managing state and handling various inputs.

## Table of Contents

- [What are Flows?](#what-are-flows)
- [Core Components of a Flow](#core-components-of-a-flow)
  - [`Flow`](#flow)
  - [`FlowStep`](#flowstep)
  - [`FlowManager`](#flowmanager)
  - [`UserFlowState`](#userflowstate)
- [Building a Flow: The `FlowBuilder`](#building-a-flow-the-flowbuilder)
  - [Creating a New Flow](#creating-a-new-flow)
  - [Defining Steps](#defining-steps)
    - [Basic Text Steps](#basic-text-steps)
    - [Step Types (`StepType*`)](#step-types-steptype)
    - [Step Prompts (Implicit)](#step-prompts-implicit)
  - [Configuring Steps (`FlowStepBuilder`)](#configuring-steps-flowstepbuilder)
    - [`OnStart(handler teleflow.FlowStepStartHandlerFunc)`](#onstarthandler-teleflowflowstepstarthandlerfunc)
    - [`OnInput(handler teleflow.FlowStepInputHandlerFunc)`](#oninputhandler-teleflowflowstepinputhandlerfunc)
    - [`WithValidator(validator teleflow.FlowValidatorFunc)`](#withvalidatorvalidator-teleflowflowvalidatorfunc)
    - [`NextStep(stepName string)`](#nextstepstepname-string)
    - [`AddTransition(input, nextStep string)`](#addtransitioninput-nextstep-string)
    - [`WithStepType(stepType FlowStepType)`](#withsteptypesteptype-flowsteptype)
    - [`StayOnInvalidInput()`](#stayoninvalidinput)
    - [`WithTimeout(timeout time.Duration)`](#withtimeouttimeout-timeduration)
  - [Setting Flow Completion and Cancellation Handlers](#setting-flow-completion-and-cancellation-handlers)
    - [`OnComplete(handler teleflow.FlowCompletionHandlerFunc)`](#oncompletehandler-teleflowflowcompletionhandlerfunc)
    - [`OnCancel(handler teleflow.FlowCancellationHandlerFunc)`](#oncancelhandler-teleflowflowcancellationhandlerfunc)
  - [Building the Flow](#building-the-flow)
- [Registering a Flow](#registering-a-flow)
- [Starting a Flow](#starting-a-flow)
- [How Flows Process Updates](#how-flows-process-updates)
  - [Input Handling](#input-handling)
  - [Validation](#validation)
  - [Transitions](#transitions)
  - [Data Management](#data-management)
- [Managing Flow State](#managing-flow-state)
  - [Accessing Flow Data in Handlers](#accessing-flow-data-in-handlers)
  - [Persisting Data Beyond the Flow](#persisting-data-beyond-the-flow)
- [Cancelling a Flow](#cancelling-a-flow)
  - [Programmatic Cancellation](#programmatic-cancellation)
  - [User-Initiated Cancellation (Exit Commands)](#user-initiated-cancellation-exit-commands)
- [Flow Configuration (`FlowConfig`)](#flow-configuration-flowconfig)
- [Advanced Flow Concepts](#advanced-flow-concepts)
  - [Step Types and UI Automation (Future Enhancement Idea)](#step-types-and-ui-automation-future-enhancement-idea)
  - [Conditional Logic within Steps](#conditional-logic-within-steps)
  - [Reusing Flows](#reusing-flows)
- [Example: User Registration Flow](#example-user-registration-flow)
- [Best Practices for Flows](#best-practices-for-flows)
- [Next Steps](#next-steps)

## What are Flows?
A flow represents a guided conversation with a user, broken down into a sequence of steps. Each step can ask for input, present choices, validate data, and decide which step comes next. Flows are ideal for:
- User onboarding and registration
- Collecting structured information (e.g., orders, feedback)
- Interactive quizzes or surveys
- Complex command sequences that require multiple inputs

## Core Components of a Flow

Refer to `core/flow.go` for the source.

### `Flow`
The main structure representing a conversational flow. It contains:
- `Name`: A unique name for the flow.
- `Steps`: An ordered slice of `FlowStep` pointers.
- `OnComplete`: A `teleflow.FlowCompletionHandlerFunc` called when the flow successfully finishes. Signature: `func(ctx *teleflow.Context, flowData map[string]interface{}) error`.
- `OnCancel`: A `teleflow.FlowCancellationHandlerFunc` called if the flow is cancelled. Signature: `func(ctx *teleflow.Context, flowData map[string]interface{}) error`.
- `Timeout`: A `time.Duration` for the entire flow (not yet fully implemented for automatic timeout).

### `FlowStep`
Represents a single step within a flow. Key attributes:
- `Name`: A unique name for the step within the flow.
- `StartHandler`: A `teleflow.FlowStepStartHandlerFunc` executed when the user *enters* this step. Use this to send prompts or initial messages for the step. Its signature is `func(ctx *teleflow.Context) error`.
- `Handler`: A `teleflow.FlowStepInputHandlerFunc` executed when the user *provides input* while on this step (this is the handler set by `OnInput`). Its signature is `func(ctx *teleflow.Context, input string) error`.
- `Validator`: A `teleflow.FlowValidatorFunc` to validate user input for this step. Its signature is `func(input string) (isValid bool, message string, validatedInput interface{}, err error)`.
- `NextStep`: The name of the default next step if no other transition matches.
- `Transitions`: A map `map[string]string` where keys are user inputs and values are the names of the next step to transition to.
- `StepType`: A `FlowStepType` (e.g., `StepTypeText`, `StepTypeChoice`). Currently, this is more for informational purposes; UI automation based on type is a potential future enhancement.
- `InvalidInputMessage`: A message to send if validation fails and `StayOnInvalidInput` is true.

### `FlowManager`
Manages all registered flows and tracks the state of users currently engaged in flows.
- `RegisterFlow(flow *Flow)`
- `StartFlow(userID int64, flowName string, ctx *Context) error`
- `HandleUpdate(ctx *Context) (bool, error)`: Processes an update for a user in a flow.
- `CancelFlow(userID int64)`
- `IsUserInFlow(userID int64) bool`

### `UserFlowState`
Tracks a user's current position and data within a flow.
- `FlowName`: Name of the active flow.
- `CurrentStep`: Name of the current step the user is on.
- `Data`: A `map[string]interface{}` to store data collected during the flow.
- `StartedAt`, `LastActive`: Timestamps for tracking.

## Building a Flow: The `FlowBuilder`

Teleflow provides a fluent `FlowBuilder` API to define flows.

### Creating a New Flow
Start with `teleflow.NewFlow(name string)`:
```go
import teleflow "github.com/kslamph/teleflow/core"

feedbackFlow := teleflow.NewFlow("user_feedback")
```

### Defining Steps
Use the `Step(name string)` method on the `FlowBuilder` or `FlowStepBuilder` to add steps.
```go
feedbackFlow.
    Step("ask_rating"). // Returns a FlowStepBuilder
    // ... configure this step ...
    Step("ask_comments"). // Adds another step, auto-linking from "ask_rating" if NextStep wasn't set
    // ... configure this step ...
```
If `NextStep` is not explicitly set on a step, the builder automatically links it to the subsequently defined step.

#### Basic Text Steps
By default, steps are of `StepTypeText`. You'll typically use the `OnStart` handler to prompt the user and the `OnInput` handler (or just rely on `NextStep`) to process their text response.

The `FlowStepBuilder` has a `WithPrompt(message string)` method. **Currently, this method is a placeholder and does not automatically send the prompt.** You must use `OnStart` to send messages.
```go
// Current way to prompt:
.Step("get_name").
    OnStart(func(ctx *teleflow.Context) error {
        return ctx.Reply("What's your name?")
    })
```

#### Step Types (`StepType*`)
Constants like `StepTypeText`, `StepTypeChoice`, `StepTypeConfirmation` are available. While they don't automate UI generation yet, they can be used for your own logic within step handlers.

#### Step Prompts (Implicit)
While `WithPrompt` is a placeholder, the common pattern is to send the prompt message within the `OnStart` handler of a step.

### Configuring Steps (`FlowStepBuilder`)
The `FlowStepBuilder` (returned by `flowBuilder.Step()`) offers methods to configure each step:

#### `OnStart(handler teleflow.FlowStepStartHandlerFunc)`
Sets a handler (`func(ctx *teleflow.Context) error`) to be executed when the user enters this step. Ideal for sending the step's question or instructions.
```go
.OnStart(func(ctx *teleflow.Context) error {
    return ctx.Reply("Please enter your email address:")
})
```

#### `OnInput(handler teleflow.FlowStepInputHandlerFunc)`
Sets a handler (`func(ctx *teleflow.Context, input string) error`) to be executed when the user provides input for this step. This handler is called *after* successful validation and *before* transitioning to the next step.
The `input string` parameter to this handler is the raw text or callback data provided by the user if no `FlowValidatorFunc` was set or if the validator returned `nil` for `validatedInput`.
If a `FlowValidatorFunc` was set and returned a non-nil `validatedInput` (e.g., a parsed number, a custom struct), you can access this richer, potentially type-converted, object using `ctx.Get("validated_input")`. The `input string` parameter will still be the original user input in this case.
You can process the input (either the string parameter or the `validated_input` from context) here, save data using `ctx.Set()`, etc.
```go
.OnInput(func(ctx *teleflow.Context) error {
.OnInput(func(ctx *teleflow.Context, input string) error {
    // 'input' is the raw user input string.
    // If a validator provided a richer object, access it:
    var validatedData interface{}
    if vInput, ok := ctx.Get("validated_input"); ok {
        validatedData = vInput
        // Example: if validator returned an int
        // if age, ok := validatedData.(int); ok { ... }
    } else {
        // Fallback to raw input if no validated_input or if it's not needed
        validatedData = input
    }
    ctx.Set("user_email", validatedData) // Save to flow data
    log.Printf("User input: %s, Validated data: %v", input, validatedData)
    return nil
})
```

#### `WithValidator(validator teleflow.FlowValidatorFunc)`
Sets an input validator for the step. The `teleflow.FlowValidatorFunc` has the signature:
`func(input string) (isValid bool, message string, validatedInput interface{}, err error)`
It returns:
- `isValid bool`: `true` if the input is valid, `false` otherwise.
- `message string`: A message to be sent to the user, especially if validation fails (e.g., "Please enter a valid email.").
- `validatedInput interface{}`: The original input, potentially transformed or type-converted. This value can be accessed in the subsequent `FlowStepInputHandlerFunc`. If validation fails, this is typically `nil`.
- `err error`: An optional error object if something went wrong during validation beyond simple invalid input (e.g., a database lookup failed). If `err` is not `nil`, the flow processing for the step typically stops, and the error might be logged.
```go
import "net/mail"

func emailValidator(input string) (bool, string, interface{}, error) {
    addr, err := mail.ParseAddress(input)
    if err != nil {
        return false, "Please enter a valid email address (e.g., user@example.com).", nil, nil
    }
    // Return the parsed address as validatedInput for potential use in OnInput
    return true, "", addr, nil
}

// In OnInput, you could then access the mail.Address struct:
// if validatedAddr, ok := ctx.Get("validated_input").(*mail.Address); ok {
//     log.Printf("Validated email address object: %s", validatedAddr.Address)
// }

// ...
.WithValidator(emailValidator)
```

#### `NextStep(stepName string)`
Sets the default next step to transition to after this step is completed (and input is valid).
```go
.Step("get_name").NextStep("get_email")
.Step("get_email") // ...
```

#### `AddTransition(input, nextStep string)`
Defines a conditional transition. If the user's input for the current step exactly matches `input`, the flow transitions to `nextStep`, overriding the default `NextStep`.
```go
.Step("confirm_choice").
    OnStart(func(ctx *teleflow.Context) error {
        // Send a message with "Yes" and "No" buttons (e.g., via inline keyboard)
        // Callback data for buttons could be "yes_confirm" and "no_confirm"
        // For simplicity, let's assume text input "Yes" or "No"
        return ctx.Reply("Are you sure? (Type Yes or No)")
    }).
    AddTransition("Yes", "final_step").
    AddTransition("No", "cancelled_step").
    NextStep("confirm_choice") // Stay on step if input is neither "Yes" nor "No"
```

#### `WithStepType(stepType FlowStepType)`
Sets the `FlowStepType` (e.g., `teleflow.StepTypeConfirmation`).
```go
.WithStepType(teleflow.StepTypeConfirmation)
```

#### `StayOnInvalidInput()`
If set, and input validation fails, the user remains on the current step. The `InvalidInputMessage` (if set on the step) or the message from the validator will be sent. If not set (default is `true`), invalid input might lead to flow cancellation or other behavior depending on the broader flow logic.

#### `WithTimeout(timeout time.Duration)`
Sets a timeout specific to this step (not yet fully implemented for automatic timeout enforcement by the manager).

### Setting Flow Completion and Cancellation Handlers

#### `OnComplete(handler teleflow.FlowCompletionHandlerFunc)`
Sets a handler (`func(ctx *teleflow.Context, flowData map[string]interface{}) error`) that is executed when the entire flow completes successfully.
The `flowData` parameter is a map containing all data collected during the flow (i.e., items set using `ctx.Set("key", value)` in various step handlers).
```go
myFlow.OnComplete(func(ctx *teleflow.Context, flowData map[string]interface{}) error {
    name, _ := flowData["user_name"].(string) // Type assertion might be needed
    email, _ := flowData["user_email"].(string)
    log.Printf("Flow completed. Name: %s, Email: %s. All data: %v", name, email, flowData)
    return ctx.Reply("Thank you for completing the registration!")
})
```

#### `OnCancel(handler teleflow.FlowCancellationHandlerFunc)`
Sets a handler (`func(ctx *teleflow.Context, flowData map[string]interface{}) error`) that is executed if the flow is cancelled.
The `flowData` parameter contains all data collected up to the point of cancellation.
```go
myFlow.OnCancel(func(ctx *teleflow.Context, flowData map[string]interface{}) error {
    log.Printf("Flow cancelled by user %d. Data collected: %v", ctx.UserID(), flowData)
    // FlowConfig.ExitMessage handles user-facing messages for command-based cancellations.
    // This handler is for cleanup, logging, or custom messages for programmatic cancellations.
    return nil
})
```

### Building the Flow
Finally, call `Build()` on the `FlowBuilder` (or `FlowStepBuilder`) to get the `*teleflow.Flow` object.
```go
finalFlow := myFlow.Build()
```

## Registering a Flow
Once built, register the flow with the bot instance:
```go
bot.RegisterFlow(finalFlow)
```

## Starting a Flow
Typically, you start a flow from within another handler (e.g., a command handler):
```go
bot.HandleCommand("register", func(ctx *teleflow.Context) error {
    if ctx.IsUserInFlow() {
        return ctx.Reply("You are already in a process. Type /cancel to exit.")
    }
    // You can pass initial data to the flow via the context
    ctx.Set("registration_source", "command")
    return ctx.StartFlow("user_registration") // Use the flow's name
})
```
`ctx.StartFlow()` initiates the flow for the user, taking them to the `OnStart` handler of the first step.

## How Flows Process Updates

When a user in a flow sends an update:
1. The `FlowManager` retrieves the user's current flow state.
2. The `StartHandler` of the current step (if not already processed for this entry) or the `Handler` (for input) is invoked.
3. **Input Handling**: The user's input (text, callback data) is made available in `ctx.Update`.
4. **Validation**: If a `teleflow.FlowValidatorFunc` is set for the current step, it's executed with the user's input.
   - The validator returns `isValid bool`, `message string`, `validatedInput interface{}`, and `err error`.
   - If `isValid` is `false`, the `message` from the validator is typically sent to the user as a reply, and the user stays on the current step (assuming `StayOnInvalidInput` behavior, which is common).
   - If `err` is not `nil`, an error has occurred during validation, and the flow processing for this input usually stops. The error should be logged.
   - The `validatedInput` (which can be the original input, a transformed version, or a richer type) is stored in the context under the key `"validated_input"` if `isValid` is `true` and `validatedInput` is not `nil`.
5. **`OnInput` Handler**: If the step has a `teleflow.FlowStepInputHandlerFunc` (set via `OnInput`), it's called *after* successful validation (`isValid` was true and `err` was nil).
   - The handler receives the original user input as its `input string` parameter.
   - It can also access the `validatedInput` returned by the validator (if any) using `ctx.Get("validated_input")`. This is useful if the validator performed type conversion or returned a more complex object.
   - This is where you typically process the input (either the raw `input` string or the richer `validated_input`) and store any necessary data using `ctx.Set("my_data_key", value)`.
6. **Transitions**:
   - The `FlowManager` checks `Transitions` on the current step against the user's input.
   - If a match is found, the user moves to the specified next step.
   - Otherwise, the default `NextStep` for the current step is used.
   - If no `NextStep` and no matching transition, the flow is considered complete.
7. **Data Management**: Data set via `ctx.Set("key", value)` during any step handler (`OnStart`, `OnInput`) is persisted in the `UserFlowState.Data` map. This data is available in subsequent step handlers (via `ctx.Get("key")`) and is passed as the `flowData map[string]interface{}` parameter to the flow's `OnComplete` and `OnCancel` handlers.

## Managing Flow State

### Accessing Flow Data in Handlers
Within any step handler (`teleflow.FlowStepStartHandlerFunc`, `teleflow.FlowStepInputHandlerFunc`) you can access data collected in previous steps using `ctx.Get(key)`. The flow's `teleflow.FlowCompletionHandlerFunc` and `teleflow.FlowCancellationHandlerFunc` receive all collected data directly as a `flowData map[string]interface{}` parameter.
```go
.OnInput(func(ctx *teleflow.Context) error {
    // Assuming "user_name" was collected in a previous step
    if name, ok := ctx.Get("user_name"); ok {
        log.Printf("Current name in flow: %s", name.(string))
    }
    // Collect current step's data
    ctx.Set("current_step_data", ctx.Update.Message.Text)
    return nil
})
```

### Persisting Data Beyond the Flow
The `UserFlowState.Data` is typically cleared when the flow ends. In the `OnComplete` handler, you should explicitly save any data you need to persist long-term (e.g., to a database, or using `bot.stateManager` for more persistent user-specific state outside of flows).

## Cancelling a Flow

### Programmatic Cancellation
From within any handler, you can cancel the current flow for the user:
```go
ctx.CancelFlow()
// Optionally, send a confirmation message
// ctx.Reply("Your current operation has been cancelled.")
```
This will trigger the flow's `OnCancel` handler, if defined.

### User-Initiated Cancellation (Exit Commands)
Teleflow allows configuring global "exit commands" (e.g., "/cancel", "/quit"). If a user in a flow sends one of these commands:
1. The flow is automatically cancelled.
2. The `FlowConfig.ExitMessage` is sent to the user.
3. The flow's `OnCancel` handler is triggered.

This is configured via `teleflow.WithFlowConfig` or specific `BotOption`s.

## Flow Configuration (`FlowConfig`)
You can customize flow behavior globally using `FlowConfig` when creating your bot or via `BotOption`s:
```go
bot, err := teleflow.NewBot(token,
    teleflow.WithFlowConfig(teleflow.FlowConfig{
        ExitCommands:        []string{"/stop", "/abort"},
        ExitMessage:         "The process has been stopped.",
        AllowGlobalCommands: true, // Allows some global commands (like /help) during flows
        HelpCommands:        []string{"/flowhelp"},
    }),
)
```
- `ExitCommands`: Commands that will cancel any active flow.
- `ExitMessage`: Message sent when a flow is cancelled by an exit command.
- `AllowGlobalCommands`: If true, allows commands specified in `HelpCommands` (and potentially others based on `resolveGlobalCommand` logic in `bot.go`) to be processed even if the user is in a flow, without interrupting the flow state unless the command itself cancels the flow.
- `HelpCommands`: Specific commands that are allowed if `AllowGlobalCommands` is true.

## Advanced Flow Concepts

### Step Types and UI Automation (Future Enhancement Idea)
As noted in `improvement-recommendations.md`, `StepTypeChoice` and `StepTypeConfirmation` currently don't automatically generate UI (keyboards). This is a potential area for future enhancement. For now, you would implement the UI (e.g., sending a keyboard) within the `OnStart` handler of such steps and use `AddTransition` or `OnInput` to handle the response.

### Conditional Logic within Steps
While `AddTransition` provides input-based branching, more complex conditional logic (e.g., based on previously collected data) can be implemented within `OnInput` or `OnStart` handlers. These handlers can then programmatically decide the next step by setting a specific value in `ctx.data` that a subsequent `OnInput` or a custom transition logic (not directly supported by `AddTransition` for dynamic values) could use, or by directly calling `ctx.StartFlow()` to a different sub-flow or restarting a step (though direct step jumping isn't a built-in feature, you'd manage it by structuring your flow transitions).

### Reusing Flows
Flows are defined by their `*teleflow.Flow` struct. You can potentially reuse flow definitions if needed, though each active flow instance for a user is a distinct `UserFlowState`.

## Example: User Registration Flow
```go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	teleflow "github.com/kslamph/teleflow/core"
)

// Validator for age
func ageValidator(input string) (bool, string, interface{}, error) {
	age, err := strconv.Atoi(input)
	if err != nil {
		return false, "Please enter a valid number for your age.", nil, nil
	}
	if age <= 0 || age > 120 {
		return false, "Please enter a realistic age.", nil, nil
	}
	// Return the parsed int as validatedInput
	return true, "", age, nil
}

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	bot, err := teleflow.NewBot(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Use(teleflow.LoggingMiddleware())
	bot.Use(teleflow.RecoveryMiddleware())

	// Define the registration flow
	registrationFlow := teleflow.NewFlow("user_registration").
		Step("get_name"). // Step 1: Get name
		OnStart(func(ctx *teleflow.Context) error { // FlowStepStartHandlerFunc
			return ctx.Reply("Welcome to registration! What's your full name?")
		}).
		OnInput(func(ctx *teleflow.Context, input string) error { // FlowStepInputHandlerFunc
			ctx.Set("reg_name", input) // Store name
			return nil
		}).
		NextStep("get_age"). // Default next step

		Step("get_age"). // Step 2: Get age
		OnStart(func(ctx *teleflow.Context) error { // FlowStepStartHandlerFunc
			name, _ := ctx.Get("reg_name")
			return ctx.Reply(fmt.Sprintf("Nice to meet you, %s! How old are you?", name.(string)))
		}).
		WithValidator(ageValidator). // FlowValidatorFunc
		OnInput(func(ctx *teleflow.Context, input string) error { // FlowStepInputHandlerFunc
			// The 'input' string is the raw text.
			// The validated age (int) is available from the validator via context.
			validatedAge, ok := ctx.Get("validated_input").(int)
			if !ok {
				// Should not happen if validator succeeded and returned an int
				log.Println("Error: validated_input for age is not an int")
				// Fallback to raw input, though it might not be ideal
				ctx.Set("reg_age_str", input)
				return fmt.Errorf("age validation issue, stored raw input")
			}
			ctx.Set("reg_age", validatedAge) // Store validated age (int)
			return nil
		}).
		NextStep("confirm_details"). // Default next step

		Step("confirm_details"). // Step 3: Confirmation
		WithStepType(teleflow.StepTypeConfirmation). // Informational
		OnStart(func(ctx *teleflow.Context) error { // FlowStepStartHandlerFunc
			name, _ := ctx.Get("reg_name").(string)
			age, _ := ctx.Get("reg_age").(int) // Age is now stored as int

			kb := teleflow.NewInlineKeyboard().
				AddButton("✅ Yes, looks good!", "reg_confirm_yes").
				AddButton("❌ No, start over", "reg_confirm_no").AddRow()

			return ctx.Reply(
				fmt.Sprintf("Please confirm your details:\nName: %s\nAge: %d", name, age), // Use %d for int
				kb,
			)
		}).
		// Transitions based on inline keyboard callback data
		AddTransition("reg_confirm_yes", "registration_complete_step"). // Custom "final" step name
		AddTransition("reg_confirm_no", "get_name"). // Go back to the first step
		// No default NextStep here, relies on transitions. Or could loop to self.

		Step("registration_complete_step"). // A dummy step to signify completion before OnComplete
		OnStart(func(ctx *teleflow.Context) error { // FlowStepStartHandlerFunc
			// This step's OnStart will be called, then the flow's OnComplete.
			// It's a good place for a "processing" message before the final OnComplete.
			return ctx.Reply("Thanks! Processing your registration...")
		}).
		// No NextStep, so flow will complete after this.

		OnComplete(func(ctx *teleflow.Context, flowData map[string]interface{}) error { // FlowCompletionHandlerFunc
			name, _ := flowData["reg_name"].(string)
			age, _ := flowData["reg_age"].(int) // Age is int
			log.Printf("User %d completed registration. Name: %s, Age: %d. Data: %v", ctx.UserID(), name, age, flowData)
			// Here you would typically save to a database
			return ctx.Reply(fmt.Sprintf("Registration successful for %s, age %d! Welcome!", name, age))
		}).
		OnCancel(func(ctx *teleflow.Context, flowData map[string]interface{}) error { // FlowCancellationHandlerFunc
			log.Printf("User %d cancelled registration. Data collected: %v", ctx.UserID(), flowData)
			// ExitMessage from FlowConfig will be sent automatically if cancelled by command.
			// This handler is for additional cleanup or custom messages if needed.
			return nil
		}).
		Build()

	bot.RegisterFlow(registrationFlow)

	// Handler to start the registration flow
	bot.HandleCommand("register", func(ctx *teleflow.Context) error {
		if ctx.IsUserInFlow() {
			return ctx.Reply("You're already in a process. Type /cancel to exit first.")
		}
		log.Printf("User %d starting registration flow.", ctx.UserID())
		return ctx.StartFlow("user_registration")
	})

	// Callback handlers for the confirmation step's inline keyboard
	// These are technically global callback handlers but are designed for this flow.
	// The flow's AddTransition will pick up these callback data.
	bot.RegisterCallback(teleflow.SimpleCallback("reg_confirm_yes", func(ctx *teleflow.Context, data string) error {
		// The flow's HandleUpdate will process this callback data via AddTransition.
		// We just need to answer the callback query.
		ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "")
		// The flow will then move to "registration_complete_step"
		// and its OnStart will trigger, followed by the flow's OnComplete.
		return nil 
	}))
	bot.RegisterCallback(teleflow.SimpleCallback("reg_confirm_no", func(ctx *teleflow.Context, data string) error {
		ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Restarting...")
		// The flow will transition back to "get_name"
		// Its OnStart will be called.
		// We might want to edit the original confirmation message.
		return ctx.EditOrReply("Okay, let's start over. What's your full name?")
	}))


	log.Println("Bot starting...")
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
```

## Best Practices for Flows
- **Clear Prompts**: Ensure each step clearly asks the user for the required input.
- **User-Friendly Validation**: Provide helpful error messages if validation fails.
- **Easy Cancellation**: Allow users to exit flows easily (e.g., via `/cancel`).
- **State Management**: Be mindful of what data you store in `UserFlowState.Data` and clear or persist it appropriately in `OnComplete` or `OnCancel`.
- **Modularity**: Break down very long or complex processes into smaller, manageable flows if possible.
- **Test Thoroughly**: Test all paths through your flow, including validation failures and cancellations.

## Next Steps
- [Handlers Guide](handlers-guide.md): For general interaction handling.
- [Keyboards Guide](keyboards-guide.md): To design the UI for your flow steps.
- [State Management (coming soon)](): For persisting flow data or user preferences beyond a single flow.
- [API Reference](api-reference.md): For detailed information on `Flow`, `FlowStep`, `FlowManager`, and related types.