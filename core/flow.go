package teleflow

import (
	// For FlowValidatorFunc new signature

	"fmt"
	"log"
	"strconv" // For ChoiceValidator
	"time"
)

// Flow system enables the creation of sophisticated multi-step conversational
// interfaces with validation, branching logic, and state management. Flows
// provide a structured way to guide users through complex interactions while
// maintaining conversation context and handling various input types.
//
// The flow system supports:
//   - Fluent builder interface for step definition and configuration
//   - Input validation with custom validator functions
//   - Conditional branching and dynamic step transitions
//   - Automatic state management and progress tracking
//   - Timeout handling and flow cancellation
//   - Custom step handlers for complex interactions
//   - Step start and input handlers for fine-grained control
//
// Basic Flow Creation:
//
//	flow := teleflow.NewFlow("user_registration").
//		Step("name").
//		OnStart(func(ctx *teleflow.Context) error {
//			return ctx.Reply("What's your name?")
//		}).
//		OnInput(func(ctx *teleflow.Context, input string) error {
//			ctx.Set("user_name", input)
//			return nil
//		}).
//		Step("age").
//		OnStart(func(ctx *teleflow.Context) error {
//			return ctx.Reply("How old are you?")
//		}).
//		OnInput(func(ctx *teleflow.Context, input string) error {
//			ctx.Set("user_age", input)
//			return nil
//		}).
//		OnComplete(func(ctx *teleflow.Context, flowData map[string]interface{}) error {
//			// Handle completed flow
//			name, _ := flowData["user_name"].(string)
//			age, _ := flowData["user_age"].(string)
//			return saveUserData(name, age)
//		}).
//		Build()
//
//	bot.RegisterFlow(flow)
//
// Advanced Flow with Validation:
//
//	flow := teleflow.NewFlow("order_processing").
//		Step("email").
//		OnStart(func(ctx *teleflow.Context) error {
//			return ctx.Reply("Enter your email:")
//		}).
//		WithValidator(func(input string) error {
//			if !isValidEmail(input) {
//				return fmt.Errorf("Please enter a valid email address")
//			}
//			return nil
//		}).
//		OnInput(func(ctx *teleflow.Context, input string) error {
//			ctx.Set("email", input)
//			return nil
//		}).
//		Step("shipping").
//		OnStart(func(ctx *teleflow.Context) error {
//			keyboard := teleflow.NewInlineKeyboard().
//				AddRow(teleflow.NewInlineButton("Standard", "Standard")).
//				AddRow(teleflow.NewInlineButton("Express", "Express")).
//				AddRow(teleflow.NewInlineButton("Overnight", "Overnight"))
//			return ctx.ReplyWithInlineKeyboard("Choose shipping method:", keyboard)
//		}).
//		WithValidator(teleflow.ChoiceValidator([]string{"Standard", "Express", "Overnight"})).
//		OnInput(func(ctx *teleflow.Context, input string) error {
//			ctx.Set("shipping", input)
//			return nil
//		}).
//		Step("address").
//		OnStart(func(ctx *teleflow.Context) error {
//			return ctx.Reply("Enter shipping address:")
//		}).
//		OnInput(func(ctx *teleflow.Context, input string) error {
//			ctx.Set("address", input)
//			return nil
//		}).
//		Build()
//
// Flow Branching with Transitions:
//
//	flow := teleflow.NewFlow("conditional_flow").
//		Step("choice").
//		OnStart(func(ctx *teleflow.Context) error {
//			keyboard := teleflow.NewInlineKeyboard().
//				AddRow(teleflow.NewInlineButton("Express", "express")).
//				AddRow(teleflow.NewInlineButton("Standard", "standard"))
//			return ctx.ReplyWithInlineKeyboard("Select shipping:", keyboard)
//		}).
//		AddTransition("express", "premium_options").
//		AddTransition("standard", "basic_options").
//		Step("premium_options").
//		OnStart(func(ctx *teleflow.Context) error {
//			return ctx.Reply("Select premium options:")
//		}).
//		Step("basic_options").
//		OnStart(func(ctx *teleflow.Context) error {
//			return ctx.Reply("Select basic options:")
//		}).
//		Build()
//
// Flow Control:
//
//	// Start a flow
//	bot.HandleCommand("/register", func(ctx *teleflow.Context, command string, args string) error {
//		return ctx.StartFlow("user_registration")
//	})
//
//	// Cancel current flow
//	bot.HandleCommand("/cancel", func(ctx *teleflow.Context, command string, args string) error {
//		return ctx.CancelFlow()
//	})
//
//	// Check flow status
//	if ctx.IsInFlow() {
//		// Handle flow-specific logic
//	}

// FlowStepType represents different types of flow steps
// type FlowStepType int

// const (
// 	// StepTypeText represents a text input step
// 	StepTypeText FlowStepType = iota
// 	// StepTypeChoice represents a choice/button step
// 	StepTypeChoice
// 	// StepTypeConfirmation represents a yes/no confirmation step
// 	StepTypeConfirmation
// 	// StepTypeCustom represents a custom step type
// 	StepTypeCustom
// )

// FlowValidatorFunc defines the function signature for input validation within a flow step.
// It is called with the raw user input string.
//
// Parameters:
//   - input: The raw string input from the user.
//
// Returns:
//   - error: nil if the input is valid. If the input is invalid, it returns an error
//     whose message (error.Error()) will be shown to the user.
//     This error should not be used for internal validator processing errors;
//     such errors should be logged within the validator and a user-friendly
//     validation error message returned instead, or handle them in a way that
//     doesn't expose internal details to the user.
type FlowValidatorFunc func(input string) error

// FlowStepStartHandlerFunc defines the function signature for handlers executed when a flow step begins.
// This handler is called before prompting the user for input for the current step.
// It can be used to send introductory messages, set up initial state for the step, or perform other preparatory actions.
//
// Parameters:
//   - ctx: The context for the current update.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type FlowStepStartHandlerFunc func(ctx *Context) error

// FlowStepInputHandlerFunc defines the function signature for handlers that process user input for a flow step.
// This handler is called after the user provides input for the current step and after any validator
// associated with the step has successfully validated the input.
//
// Parameters:
//   - ctx: The context for the current update.
//   - input: The raw string input from the user. If a validator was configured for the step
//     and ran, this input has been deemed valid by that validator.
//     Transformations of input are no longer a direct return from validators;
//     if needed, they should be handled within this input handler or a middleware,
//     possibly by parsing the `input` string and setting derived values in `ctx`.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type FlowStepInputHandlerFunc func(ctx *Context, input string) error

// FlowCompletionHandlerFunc defines the function signature for handlers executed when a flow successfully completes.
// This handler is called after the last step in the flow has been processed.
//
// Parameters:
//   - ctx: The context for the current update.
//   - flowData: A map containing all the data collected throughout the flow. Keys correspond
//     to step names or keys set by validators/handlers (e.g., "validated_input").
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type FlowCompletionHandlerFunc func(ctx *Context, flowData map[string]interface{}) error

// FlowCancellationHandlerFunc defines the function signature for handlers executed when a flow is cancelled.
// This can happen if the user issues an exit command or if a step explicitly cancels the flow.
//
// Parameters:
//   - ctx: The context for the current update.
//   - flowData: A map containing data collected in the flow up to the point of cancellation.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type FlowCancellationHandlerFunc func(ctx *Context, flowData map[string]interface{}) error

// FlowTransition represents a transition between flow steps
type FlowTransition struct {
	Condition string // Input condition that triggers this transition
	NextStep  string // Name of the next step
}

// FlowConfig holds configuration for flow behavior
type FlowConfig struct {
	ExitCommands        []string
	ExitMessage         string
	AllowGlobalCommands bool
	HelpCommands        []string
}

// FlowManager manages all flows and user flow states
type FlowManager struct {
	flows        map[string]*Flow
	userFlows    map[int64]*UserFlowState
	stateManager StateManager
	botConfig    *FlowConfig
}

// NewFlowManager creates a new flow manager
func NewFlowManager(stateManager StateManager) *FlowManager {
	return &FlowManager{
		flows:        make(map[string]*Flow),
		userFlows:    make(map[int64]*UserFlowState),
		stateManager: stateManager,
	}
}

// SetBotConfig sets the bot configuration for the flow manager
func (fm *FlowManager) SetBotConfig(config FlowConfig) {
	fm.botConfig = &config
}

// IsUserInFlow checks if a user is currently in a flow
func (fm *FlowManager) IsUserInFlow(userID int64) bool {
	_, exists := fm.userFlows[userID]
	return exists
}

// CancelFlow cancels the current flow for a user
func (fm *FlowManager) CancelFlow(userID int64) {

	delete(fm.userFlows, userID)

}

// Flow represents a structured multi-step conversation
type Flow struct {
	Name        string
	Steps       []*FlowStep
	stepMap     map[string]*FlowStep
	transitions map[string][]string
	OnComplete  FlowCompletionHandlerFunc
	OnCancel    FlowCancellationHandlerFunc
	Timeout     time.Duration
}

// FlowStep represents a single step in a flow
type FlowStep struct {
	Name               string
	StartHandler       FlowStepStartHandlerFunc // Called when entering the step
	Handler            FlowStepInputHandlerFunc // Called when receiving input
	Validator          FlowValidatorFunc
	NextStep           string
	Transitions        map[string]string // input -> next step
	Timeout            time.Duration
	StayOnInvalidInput bool // Stay in step (true) or cancel flow (false) on invalid input
	// StepType            FlowStepType
	InvalidInputMessage string
}

// UserFlowState tracks a user's current position in a flow
type UserFlowState struct {
	FlowName    string
	CurrentStep string
	Data        map[string]interface{}
	StartedAt   time.Time
	LastActive  time.Time
}

// FlowBuilder provides a fluent interface for building flows
type FlowBuilder struct {
	flow *Flow
}

// NewFlow creates a new flow with the given name
func NewFlow(name string) *FlowBuilder {
	return &FlowBuilder{
		flow: &Flow{
			Name:        name,
			Steps:       []*FlowStep{},
			stepMap:     make(map[string]*FlowStep),
			transitions: make(map[string][]string),
			Timeout:     30 * time.Minute,
		},
	}
}

// Step creates and returns a new step builder
func (fb *FlowBuilder) Step(name string) *FlowStepBuilder {
	step := &FlowStep{
		Name:               name,
		Transitions:        make(map[string]string),
		Timeout:            5 * time.Minute,
		StayOnInvalidInput: true, // Default: stay on invalid input for better UX
		// StepType:           StepTypeText,
	}

	fb.flow.Steps = append(fb.flow.Steps, step)
	fb.flow.stepMap[name] = step

	// Auto-link to previous step if this is not the first
	if len(fb.flow.Steps) > 1 {
		prevStep := fb.flow.Steps[len(fb.flow.Steps)-2]
		if prevStep.NextStep == "" {
			prevStep.NextStep = name
		}
	}

	return &FlowStepBuilder{
		flowBuilder: fb,
		step:        step,
	}
}

// OnComplete sets the FlowCompletionHandlerFunc that is called when the flow successfully
// finishes all its steps.
//
// Parameters:
//   - handler: The FlowCompletionHandlerFunc to execute upon flow completion.
func (fb *FlowBuilder) OnComplete(handler FlowCompletionHandlerFunc) *FlowBuilder {
	fb.flow.OnComplete = handler
	return fb
}

// OnCancel sets the FlowCancellationHandlerFunc that is called if the flow is cancelled
// before completion (e.g., by an exit command or programmatically).
//
// Parameters:
//   - handler: The FlowCancellationHandlerFunc to execute upon flow cancellation.
func (fb *FlowBuilder) OnCancel(handler FlowCancellationHandlerFunc) *FlowBuilder {
	fb.flow.OnCancel = handler
	return fb
}

// Build creates the final flow
func (fb *FlowBuilder) Build() *Flow {
	return fb.flow
}

// FlowStepBuilder provides a fluent interface for building flow steps
type FlowStepBuilder struct {
	flowBuilder *FlowBuilder
	step        *FlowStep
}

// WithValidator sets a FlowValidatorFunc for the current step.
// This function will be called to validate user input for this step before the OnInput handler.
//
// Parameters:
//   - validator: The FlowValidatorFunc to use for input validation.
func (fsb *FlowStepBuilder) WithValidator(validator FlowValidatorFunc) *FlowStepBuilder {
	fsb.step.Validator = validator
	return fsb
}

// NextStep sets the default next step
func (fsb *FlowStepBuilder) NextStep(stepName string) *FlowStepBuilder {
	fsb.step.NextStep = stepName
	return fsb
}

// OnStart sets the FlowStepStartHandlerFunc for the current step.
// This handler is executed when the flow transitions into this step, before prompting the user for input.
//
// Parameters:
//   - handler: The FlowStepStartHandlerFunc to execute when the step starts.
func (fsb *FlowStepBuilder) OnStart(handler FlowStepStartHandlerFunc) *FlowStepBuilder {
	fsb.step.StartHandler = handler
	return fsb
}

// OnInput sets the FlowStepInputHandlerFunc for the current step.
// This handler is executed after the user provides input and (if configured) after the input
// has been successfully validated by a FlowValidatorFunc.
//
// Parameters:
//   - handler: The FlowStepInputHandlerFunc to process user input for this step.
func (fsb *FlowStepBuilder) OnInput(handler FlowStepInputHandlerFunc) *FlowStepBuilder {
	fsb.step.Handler = handler
	return fsb
}

// WithTimeout sets a timeout for this step
func (fsb *FlowStepBuilder) WithTimeout(timeout time.Duration) *FlowStepBuilder {
	fsb.step.Timeout = timeout
	return fsb
}

// StayOnInvalidInput configures behavior on invalid input
func (fsb *FlowStepBuilder) StayOnInvalidInput() *FlowStepBuilder {
	fsb.step.StayOnInvalidInput = true
	return fsb
}

// WithStepType sets the type of the step
// func (fsb *FlowStepBuilder) WithStepType(stepType FlowStepType) *FlowStepBuilder {
// 	fsb.step.StepType = stepType
// 	return fsb
// }

// AddTransition adds an input-based transition to another step
func (fsb *FlowStepBuilder) AddTransition(input, nextStep string) *FlowStepBuilder {
	fsb.step.Transitions[input] = nextStep
	return fsb
}

// Step continues building the flow with a new step
func (fsb *FlowStepBuilder) Step(name string) *FlowStepBuilder {
	return fsb.flowBuilder.Step(name)
}

// OnComplete sets the FlowCompletionHandlerFunc for the entire flow.
// This is a convenience method that calls the underlying FlowBuilder's OnComplete.
//
// Parameters:
//   - handler: The FlowCompletionHandlerFunc to execute upon flow completion.
func (fsb *FlowStepBuilder) OnComplete(handler FlowCompletionHandlerFunc) *FlowBuilder {
	return fsb.flowBuilder.OnComplete(handler)
}

// OnCancel sets the FlowCancellationHandlerFunc for the entire flow.
// This is a convenience method that calls the underlying FlowBuilder's OnCancel.
//
// Parameters:
//   - handler: The FlowCancellationHandlerFunc to execute upon flow cancellation.
func (fsb *FlowStepBuilder) OnCancel(handler FlowCancellationHandlerFunc) *FlowBuilder {
	return fsb.flowBuilder.OnCancel(handler)
}

// Build creates the final flow
func (fsb *FlowStepBuilder) Build() *Flow {
	return fsb.flowBuilder.Build()
}

// RegisterFlow registers a flow with the manager
func (fm *FlowManager) RegisterFlow(flow *Flow) {
	fm.flows[flow.Name] = flow
}

// StartFlow starts a flow for a user
func (fm *FlowManager) StartFlow(userID int64, flowName string, ctx *Context) error {
	flow, exists := fm.flows[flowName]
	if !exists {
		return fmt.Errorf("flow %s not found", flowName)
	}

	if len(flow.Steps) == 0 {
		return fmt.Errorf("flow %s has no steps", flowName)
	}

	// Initialize data from context if available
	initialData := make(map[string]interface{})
	if ctx != nil {
		for key, value := range ctx.data {
			initialData[key] = value
		}
	}

	userState := &UserFlowState{
		FlowName:    flowName,
		CurrentStep: flow.Steps[0].Name,
		Data:        initialData,
		StartedAt:   time.Now(),
		LastActive:  time.Now(),
	}

	fm.userFlows[userID] = userState

	// Execute the start handler of the first step if it exists
	firstStep := flow.Steps[0]
	if firstStep.StartHandler != nil && ctx != nil {
		// Store user state data in context
		for key, value := range userState.Data {
			ctx.Set(key, value)
		}
		return firstStep.StartHandler(ctx)
	}

	return nil
}

// HandleUpdate processes an update for a user in a flow
func (fm *FlowManager) HandleUpdate(ctx *Context) (bool, error) {
	userState, exists := fm.userFlows[ctx.UserID()]
	if !exists {
		return false, nil
	}

	flow := fm.flows[userState.FlowName]
	if flow == nil {
		delete(fm.userFlows, ctx.UserID())
		return false, fmt.Errorf("flow %s not found", userState.FlowName)
	}

	currentStep := flow.stepMap[userState.CurrentStep]
	if currentStep == nil {
		delete(fm.userFlows, ctx.UserID())
		return false, fmt.Errorf("step %s not found in flow %s", userState.CurrentStep, userState.FlowName)
	}

	// Update last active time
	userState.LastActive = time.Now()

	// Store user state data in context
	for key, value := range userState.Data {
		ctx.Set(key, value)
	}

	// Validate input if validator exists
	// var validatedInput interface{} // To store the validated input for the handler
	if currentStep.Validator != nil {
		var inputToValidate string
		if ctx.Update.Message != nil {
			inputToValidate = ctx.Update.Message.Text
		} else if ctx.Update.CallbackQuery != nil {
			// For callbacks, data might not be suitable for direct validation by a text validator.
			// This assumes validators are primarily for text or simple callback data.
			inputToValidate = ctx.Update.CallbackQuery.Data
		}

		if inputToValidate != "" {
			validationErr := currentStep.Validator(inputToValidate)

			if validationErr != nil {
				exitHint := ""
				if fm.botConfig != nil && len(fm.botConfig.ExitCommands) > 0 {
					exitHint = fmt.Sprintf("\n\nType '%s' to cancel.", fm.botConfig.ExitCommands[0])
				}
				replyMsg := validationErr.Error()
				if replyMsg == "" {
					replyMsg = "Invalid input." // Default message if validator returns empty
				}
				return true, ctx.Reply(fmt.Sprintf("❌ %s%s", replyMsg, exitHint))
			}

			ctx.Set("validated_input", inputToValidate) // Make it available in context
		}
	}

	// Execute step input handler (OnInput)
	if currentStep.Handler != nil {
		var inputForHandler string
		// Prefer validated input if available
		if valInput, ok := ctx.Get("validated_input"); ok {
			// The handler expects a string, so we need to decide how to pass validatedInput.
			// If validatedInput is not a string, this might be an issue or require type assertion/conversion.
			// For now, assuming validatedInput can be reasonably converted or the handler is flexible.
			// This is a simplification; a more robust system might involve typed validated input.
			if strVal, okStr := valInput.(string); okStr {
				inputForHandler = strVal
			} else {
				// If validated_input is not a string, pass the original input or handle error
				// For now, let's fall back to original input if validated_input isn't a string.
				// This part might need refinement based on how FlowStepInputHandlerFunc is used with validated types.
				if ctx.Update.Message != nil {
					inputForHandler = ctx.Update.Message.Text
				} else if ctx.Update.CallbackQuery != nil {
					inputForHandler = ctx.Update.CallbackQuery.Data
				}
				log.Printf("Warning: validated_input for step %s was not a string, using original input for handler.", currentStep.Name)
			}
		} else { // No validated input, use raw input
			if ctx.Update.Message != nil {
				inputForHandler = ctx.Update.Message.Text
			} else if ctx.Update.CallbackQuery != nil {
				inputForHandler = ctx.Update.CallbackQuery.Data
			}
		}

		if err := currentStep.Handler(ctx, inputForHandler); err != nil {
			return true, err
		}
	}

	// For callback queries, also execute the registered callback handler
	if ctx.Update.CallbackQuery != nil {
		callbackHandler := ctx.Bot.callbackRegistry.Handle(ctx.Update.CallbackQuery.Data)
		if callbackHandler != nil {
			if err := callbackHandler(ctx); err != nil {
				return true, err
			}
		}
	}

	// Determine next step
	var input string
	if ctx.Update.Message != nil {
		input = ctx.Update.Message.Text
	} else if ctx.Update.CallbackQuery != nil {
		input = ctx.Update.CallbackQuery.Data
	}
	nextStep := fm.determineNextStep(userState, input)

	if nextStep == "" {
		// Flow completed
		delete(fm.userFlows, ctx.UserID())
		if flow.OnComplete != nil {
			// Pass the accumulated flow data to the OnComplete handler
			return true, flow.OnComplete(ctx, userState.Data)
		}
		return true, nil
	} else if nextStep == "_cancel_" {
		// Flow cancelled
		delete(fm.userFlows, ctx.UserID())
		if flow.OnCancel != nil {
			// Pass the accumulated flow data to the OnCancel handler
			return true, flow.OnCancel(ctx, userState.Data)
		}
		return true, nil
	} else if nextStep == currentStep.Name {
		// Staying in current step due to invalid input
		if currentStep.InvalidInputMessage != "" {
			exitHint := ""
			if fm.botConfig != nil && len(fm.botConfig.ExitCommands) > 0 {
				exitHint = fmt.Sprintf(" Type '%s' to cancel.", fm.botConfig.ExitCommands[0])
			}
			if ctx.Reply(currentStep.InvalidInputMessage+exitHint) != nil {
				log.Printf("Failed to send invalid input message for step %s in flow %s", currentStep.Name, flow.Name)
			}
		}
		// Save context data back to user state but don't advance
		for key, value := range ctx.data {
			userState.Data[key] = value
		}
		return true, nil
	} else {
		// Move to next step
		userState.CurrentStep = nextStep

		// Save context data back to user state
		for key, value := range ctx.data {
			userState.Data[key] = value
		}

		// Execute start handler for the new step if it exists
		newStepConfiguration := flow.stepMap[nextStep]
		if newStepConfiguration != nil && newStepConfiguration.StartHandler != nil {
			// Update context with the potentially modified user state data before calling StartHandler
			for key, value := range userState.Data { // userState.Data might have been updated by the previous step's handler
				ctx.Set(key, value)
			}
			if err := newStepConfiguration.StartHandler(ctx); err != nil {
				return true, err
			}
		}

		return true, nil
	}
}

// determineNextStep determines the next step based on current state and input
func (fm *FlowManager) determineNextStep(userFlow *UserFlowState, input string) string {
	flow := fm.flows[userFlow.FlowName]
	if flow == nil {
		return "_cancel_"
	}

	currentStep := flow.stepMap[userFlow.CurrentStep]
	if currentStep == nil {
		return "_cancel_"
	}

	// 1. Check for input-based transitions
	if input != "" {
		if nextStep, exists := currentStep.Transitions[input]; exists {
			return nextStep
		}
	}

	// 2. Use default next step if available
	if currentStep.NextStep != "" {
		return currentStep.NextStep
	}

	// 3. For unexpected input, check StayOnInvalidInput behavior
	if currentStep.StayOnInvalidInput {
		return currentStep.Name // Stay in current step
	} else {
		return "_cancel_" // Cancel flow on invalid input
	}
}

// NumberValidator validates numeric input
func NumberValidator() FlowValidatorFunc {
	return func(input string) error {
		if input == "" {
			return fmt.Errorf("Please enter a number.")
		}
		_, parseErr := strconv.Atoi(input)
		if parseErr != nil {
			// Check if it's a float
			_, floatParseErr := strconv.ParseFloat(input, 64)
			if floatParseErr != nil {
				return fmt.Errorf("Please enter a valid number (integer or decimal).")
			}
			return nil
		}
		return nil
	}
}

// ChoiceValidator validates choice input against allowed options
func ChoiceValidator(choices []string) FlowValidatorFunc {
	return func(input string) error {
		if input == "" {
			return fmt.Errorf("Please choose one of: %v", choices)
		}

		for _, choice := range choices {
			if input == choice {
				return nil
			}
		}

		return fmt.Errorf("Please choose one of: %v", choices)
	}
}
