package teleflow

import (
	"fmt"
	"log"
	"time"
)

// Flow system enables the creation of sophisticated multi-step conversational
// interfaces with validation, branching logic, and state management. Flows
// provide a structured way to guide users through complex interactions while
// maintaining conversation context and handling various input types.
//
// The flow system supports:
//   - Multiple step types (text input, choices, confirmations, custom)
//   - Input validation with custom validator functions
//   - Conditional branching and dynamic step transitions
//   - Automatic state management and progress tracking
//   - Timeout handling and flow cancellation
//   - Custom step handlers for complex interactions
//
// Basic Flow Creation:
//
//	flow := teleflow.NewFlow("user_registration").
//		AddStep("name", teleflow.StepTypeText, "What's your name?").
//		AddStep("age", teleflow.StepTypeText, "How old are you?").
//		AddStep("confirm", teleflow.StepTypeConfirmation, "Is this information correct?")
//
//	bot.RegisterFlow(flow, func(ctx *teleflow.Context, result map[string]string) error {
//		// Handle completed flow
//		return saveUserData(result["name"], result["age"])
//	})
//
// Advanced Flow with Validation:
//
//	flow := teleflow.NewFlow("order_processing").
//		AddStepWithValidator("email", teleflow.StepTypeText, "Enter your email:",
//			func(input string) (bool, string) {
//				if !isValidEmail(input) {
//					return false, "Please enter a valid email address"
//				}
//				return true, ""
//			}).
//		AddChoiceStep("shipping", "Choose shipping method:",
//			[]string{"Standard", "Express", "Overnight"}).
//		AddStep("address", teleflow.StepTypeText, "Enter shipping address:")
//
// Flow Branching:
//
//	flow.AddConditionalStep("premium_options", func(ctx *teleflow.Context, answers map[string]string) bool {
//		return answers["shipping"] == "Express" || answers["shipping"] == "Overnight"
//	}, teleflow.StepTypeChoice, "Select premium options:")
//
// Flow Control:
//
//	// Start a flow
//	bot.HandleCommand("/register", func(ctx *teleflow.Context) error {
//		return bot.StartFlow(ctx, "user_registration")
//	})
//
//	// Cancel current flow
//	bot.HandleCommand("/cancel", func(ctx *teleflow.Context) error {
//		return bot.CancelFlow(ctx)
//	})
//
//	// Check flow status
//	if bot.IsInFlow(ctx) {
//		currentStep := bot.GetCurrentFlowStep(ctx)
//		// Handle flow-specific logic
//	}

// FlowStepType represents different types of flow steps
type FlowStepType int

const (
	// StepTypeText represents a text input step
	StepTypeText FlowStepType = iota
	// StepTypeChoice represents a choice/button step
	StepTypeChoice
	// StepTypeConfirmation represents a yes/no confirmation step
	StepTypeConfirmation
	// StepTypeCustom represents a custom step type
	StepTypeCustom
)

// FlowValidatorFunc defines the function signature for input validation
type FlowValidatorFunc func(input string) (bool, string)

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
	OnComplete  HandlerFunc
	OnCancel    HandlerFunc
	Timeout     time.Duration
}

// FlowStep represents a single step in a flow
type FlowStep struct {
	Name                string
	StartHandler        HandlerFunc // Called when entering the step
	Handler             HandlerFunc // Called when receiving input
	Validator           FlowValidatorFunc
	NextStep            string
	Transitions         map[string]string // input -> next step
	Timeout             time.Duration
	StayOnInvalidInput  bool // Stay in step (true) or cancel flow (false) on invalid input
	StepType            FlowStepType
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
		StepType:           StepTypeText,
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

// OnComplete sets the completion handler
func (fb *FlowBuilder) OnComplete(handler HandlerFunc) *FlowBuilder {
	fb.flow.OnComplete = handler
	return fb
}

// OnCancel sets the cancellation handler
func (fb *FlowBuilder) OnCancel(handler HandlerFunc) *FlowBuilder {
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

// WithValidator adds input validation to the step
func (fsb *FlowStepBuilder) WithValidator(validator FlowValidatorFunc) *FlowStepBuilder {
	fsb.step.Validator = validator
	return fsb
}

// NextStep sets the default next step
func (fsb *FlowStepBuilder) NextStep(stepName string) *FlowStepBuilder {
	fsb.step.NextStep = stepName
	return fsb
}

// OnStart sets the step start handler (called when entering the step)
func (fsb *FlowStepBuilder) OnStart(handler HandlerFunc) *FlowStepBuilder {
	fsb.step.StartHandler = handler
	return fsb
}

// OnInput sets the step input handler
func (fsb *FlowStepBuilder) OnInput(handler HandlerFunc) *FlowStepBuilder {
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
func (fsb *FlowStepBuilder) WithStepType(stepType FlowStepType) *FlowStepBuilder {
	fsb.step.StepType = stepType
	return fsb
}

// AddTransition adds an input-based transition to another step
func (fsb *FlowStepBuilder) AddTransition(input, nextStep string) *FlowStepBuilder {
	fsb.step.Transitions[input] = nextStep
	return fsb
}

// Step continues building the flow with a new step
func (fsb *FlowStepBuilder) Step(name string) *FlowStepBuilder {
	return fsb.flowBuilder.Step(name)
}

// OnComplete sets the completion handler
func (fsb *FlowStepBuilder) OnComplete(handler HandlerFunc) *FlowBuilder {
	return fsb.flowBuilder.OnComplete(handler)
}

// OnCancel sets the cancellation handler
func (fsb *FlowStepBuilder) OnCancel(handler HandlerFunc) *FlowBuilder {
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
	if currentStep.Validator != nil {
		var inputToValidate string
		if ctx.Update.Message != nil {
			inputToValidate = ctx.Update.Message.Text
		} else if ctx.Update.CallbackQuery != nil {
			inputToValidate = ctx.Update.CallbackQuery.Data
		}

		if inputToValidate != "" {
			if valid, helpText := currentStep.Validator(inputToValidate); !valid {
				exitHint := ""
				if fm.botConfig != nil && len(fm.botConfig.ExitCommands) > 0 {
					exitHint = fmt.Sprintf("\n\nType '%s' to cancel.", fm.botConfig.ExitCommands[0])
				}
				return true, ctx.Reply(fmt.Sprintf("❌ Invalid input.\n\n%s%s", helpText, exitHint))
			}
		}
	}

	// Execute step handler
	if currentStep.Handler != nil {
		if err := currentStep.Handler(ctx); err != nil {
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
			return true, flow.OnComplete(ctx)
		}
		return true, nil
	} else if nextStep == "_cancel_" {
		// Flow cancelled
		delete(fm.userFlows, ctx.UserID())
		if flow.OnCancel != nil {
			return true, flow.OnCancel(ctx)
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
		newStep := flow.stepMap[nextStep]
		if newStep != nil && newStep.StartHandler != nil {
			// Update context with the new state data
			for key, value := range userState.Data {
				ctx.Set(key, value)
			}
			return true, newStep.StartHandler(ctx)
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
	return func(input string) (bool, string) {
		if input == "" {
			return false, "Please enter a number."
		}

		// Simple numeric validation - check if input contains only digits, decimal point, and optional minus sign
		for i, char := range input {
			if char >= '0' && char <= '9' {
				continue
			}
			if char == '.' {
				continue
			}
			if char == '-' && i == 0 {
				continue
			}
			return false, "Please enter a valid number."
		}

		return true, ""
	}
}

// ChoiceValidator validates choice input against allowed options
func ChoiceValidator(choices []string) FlowValidatorFunc {
	return func(input string) (bool, string) {
		if input == "" {
			return false, fmt.Sprintf("Please choose one of: %v", choices)
		}

		for _, choice := range choices {
			if input == choice {
				return true, ""
			}
		}

		return false, fmt.Sprintf("Please choose one of: %v", choices)
	}
}
