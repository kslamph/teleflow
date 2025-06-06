package teleflow

import (
	"fmt"
	"time"
)

// Flow system enables the creation of sophisticated multi-step conversational
// interfaces using the new Step-Prompt-Process API. This system provides:
//   - Declarative prompt configuration (Message, Image, Keyboard)
//   - Unified input processing with ProcessFunc
//   - Clear flow control with ProcessResult actions
//   - Simplified developer experience with zero learning curve

// FlowConfig holds configuration for flow behavior
type FlowConfig struct {
	ExitCommands        []string
	ExitMessage         string
	AllowGlobalCommands bool
	HelpCommands        []string
}

// FlowManager manages all flows and user flow states
type FlowManager struct {
	flows          map[string]*Flow
	userFlows      map[int64]*UserFlowState
	stateManager   StateManager
	botConfig      *FlowConfig
	promptRenderer *PromptRenderer
}

// NewFlowManager creates a new flow manager
func NewFlowManager(stateManager StateManager) *FlowManager {
	return &FlowManager{
		flows:        make(map[string]*Flow),
		userFlows:    make(map[int64]*UserFlowState),
		stateManager: stateManager,
	}
}

// SetBot sets the bot instance and initializes the prompt renderer
func (fm *FlowManager) SetBot(bot *Bot) {
	fm.promptRenderer = NewPromptRenderer(bot)
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

// Flow represents a structured multi-step conversation using the new Step-Prompt-Process API
type Flow struct {
	Name       string
	Steps      map[string]*FlowStep // Map for easier lookup by step name
	Order      []string             // Maintains step execution order
	OnComplete func(*Context) error // Simplified completion handler
	Timeout    time.Duration
}

// FlowStep represents a single step in a flow using the new API
type FlowStep struct {
	Name         string
	PromptConfig *PromptConfig        // New: declarative prompt specification
	ProcessFunc  ProcessFunc          // New: unified input processing
	OnComplete   func(*Context) error // Optional step completion handler
	Timeout      time.Duration
}

// UserFlowState tracks a user's current position in a flow
type UserFlowState struct {
	FlowName    string
	CurrentStep string
	Data        map[string]interface{}
	StartedAt   time.Time
	LastActive  time.Time
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

	if len(flow.Order) == 0 {
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
		CurrentStep: flow.Order[0], // First step in order
		Data:        initialData,
		StartedAt:   time.Now(),
		LastActive:  time.Now(),
	}

	fm.userFlows[userID] = userState

	// Render the first step's prompt
	if ctx != nil {
		return fm.renderStepPrompt(ctx, flow, flow.Order[0], userState)
	}

	return nil
}

// renderStepPrompt renders the prompt for a given step
func (fm *FlowManager) renderStepPrompt(ctx *Context, flow *Flow, stepName string, userState *UserFlowState) error {
	step := flow.Steps[stepName]
	if step == nil {
		return fmt.Errorf("step %s not found", stepName)
	}

	if step.PromptConfig == nil {
		return fmt.Errorf("step %s has no prompt configuration", stepName)
	}

	// Load user state data into context for prompt rendering
	for key, value := range userState.Data {
		ctx.Set(key, value)
	}

	// Create render context with proper step and flow information
	if fm.promptRenderer == nil {
		return fmt.Errorf("PromptRenderer not initialized - call SetBot() on FlowManager")
	}

	renderCtx := &RenderContext{
		ctx:          ctx,
		promptConfig: step.PromptConfig,
		stepName:     stepName,
		flowName:     flow.Name,
	}

	return fm.promptRenderer.Render(renderCtx)
}

// HandleUpdate processes an update for a user in a flow using the new API
func (fm *FlowManager) HandleUpdate(ctx *Context) (bool, error) {
	userID := ctx.UserID()
	userState, exists := fm.userFlows[userID]
	if !exists {
		return false, nil // User not in a flow
	}

	flow := fm.flows[userState.FlowName]
	if flow == nil {
		delete(fm.userFlows, userID)
		return false, fmt.Errorf("flow %s not found", userState.FlowName)
	}

	currentStep := flow.Steps[userState.CurrentStep]
	if currentStep == nil {
		delete(fm.userFlows, userID)
		return false, fmt.Errorf("step %s not found", userState.CurrentStep)
	}

	// Update last active time
	userState.LastActive = time.Now()

	// Extract input and button click data
	input, buttonClick := fm.extractInputData(ctx)

	// Load user state data into context
	for key, value := range userState.Data {
		ctx.Set(key, value)
	}

	// Execute ProcessFunc
	if currentStep.ProcessFunc == nil {
		return true, fmt.Errorf("step %s has no process function", userState.CurrentStep)
	}

	result := currentStep.ProcessFunc(ctx, input, buttonClick)

	// Answer callback query if this was a button click to dismiss loading indicator
	if buttonClick != nil {
		if err := ctx.answerCallbackQuery(""); err != nil {
			// Log the error but don't fail the flow processing
			// The user experience continues even if callback answering fails
			_ = err // Acknowledge error but continue
		}
	}

	// Update user state data from context
	for key, value := range ctx.data {
		userState.Data[key] = value
	}

	// Handle ProcessResult
	return fm.handleProcessResult(ctx, result, userState, flow)
}

// extractInputData extracts input text and button click information from the update
func (fm *FlowManager) extractInputData(ctx *Context) (string, *ButtonClick) {
	var input string
	var buttonClick *ButtonClick

	if ctx.Update.Message != nil {
		input = ctx.Update.Message.Text
	} else if ctx.Update.CallbackQuery != nil {
		input = ctx.Update.CallbackQuery.Data
		buttonClick = &ButtonClick{
			Data:     ctx.Update.CallbackQuery.Data,
			Text:     ctx.Update.CallbackQuery.Message.Text,
			UserID:   ctx.UserID(),
			ChatID:   ctx.Update.CallbackQuery.Message.Chat.ID,
			Metadata: make(map[string]interface{}),
		}
	}

	return input, buttonClick
}

// handleProcessResult processes the result from a ProcessFunc
func (fm *FlowManager) handleProcessResult(ctx *Context, result ProcessResult, userState *UserFlowState, flow *Flow) (bool, error) {
	// Show custom prompt if specified (informational only, no keyboard)
	if result.Prompt != nil {
		if err := fm.renderInformationalPrompt(ctx, result.Prompt); err != nil {
			return true, fmt.Errorf("failed to render ProcessResult prompt: %w", err)
		}
	}

	// Execute action
	switch result.Action {
	case ActionNextStep:
		return fm.advanceToNextStep(ctx, userState, flow)

	case ActionGoToStep:
		return fm.goToSpecificStep(ctx, userState, flow, result.TargetStep)

	case ActionRetry:
		// Stay on current step, optionally re-render prompt
		if result.Prompt == nil {
			currentStep := flow.Steps[userState.CurrentStep]
			if currentStep != nil && currentStep.PromptConfig != nil {
				return true, fm.renderStepPrompt(ctx, flow, userState.CurrentStep, userState)
			}
		}
		return true, nil

	case ActionCompleteFlow:
		return fm.completeFlow(ctx, userState, flow)

	case ActionCancelFlow:
		return fm.cancelFlow(ctx, userState, flow)

	default:
		return true, fmt.Errorf("unknown ProcessAction: %d", result.Action)
	}
}

// renderInformationalPrompt renders a prompt without keyboard for informational messages
func (fm *FlowManager) renderInformationalPrompt(ctx *Context, config *PromptConfig) error {
	if fm.promptRenderer == nil {
		return fmt.Errorf("PromptRenderer not initialized - call SetBot() on FlowManager")
	}

	// Create informational prompt without keyboard
	infoPrompt := &PromptConfig{
		Message: config.Message,
		Image:   config.Image,
		// Keyboard is intentionally omitted for informational messages
	}

	renderCtx := &RenderContext{
		ctx:          ctx,
		promptConfig: infoPrompt,
		stepName:     "info",
		flowName:     "system",
	}

	return fm.promptRenderer.Render(renderCtx)
}

// advanceToNextStep moves to the next step in sequence
func (fm *FlowManager) advanceToNextStep(ctx *Context, userState *UserFlowState, flow *Flow) (bool, error) {
	// Find current step index
	currentIndex := -1
	for i, stepName := range flow.Order {
		if stepName == userState.CurrentStep {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return true, fmt.Errorf("current step %s not found in flow order", userState.CurrentStep)
	}

	// Check if there's a next step
	if currentIndex+1 >= len(flow.Order) {
		// No more steps, complete the flow
		return fm.completeFlow(ctx, userState, flow)
	}

	// Move to next step
	nextStepName := flow.Order[currentIndex+1]
	userState.CurrentStep = nextStepName

	// Render the next step's prompt
	return true, fm.renderStepPrompt(ctx, flow, nextStepName, userState)
}

// goToSpecificStep jumps to a named step
func (fm *FlowManager) goToSpecificStep(ctx *Context, userState *UserFlowState, flow *Flow, targetStep string) (bool, error) {
	if _, exists := flow.Steps[targetStep]; !exists {
		return true, fmt.Errorf("target step %s not found in flow", targetStep)
	}

	userState.CurrentStep = targetStep
	return true, fm.renderStepPrompt(ctx, flow, targetStep, userState)
}

// completeFlow handles flow completion
func (fm *FlowManager) completeFlow(ctx *Context, userState *UserFlowState, flow *Flow) (bool, error) {
	// Execute completion handler if it exists
	if flow.OnComplete != nil {
		if err := flow.OnComplete(ctx); err != nil {
			delete(fm.userFlows, ctx.UserID())
			return true, err
		}
	}

	// Remove user from flow
	delete(fm.userFlows, ctx.UserID())
	return true, nil
}

// cancelFlow handles flow cancellation
func (fm *FlowManager) cancelFlow(ctx *Context, userState *UserFlowState, flow *Flow) (bool, error) {
	// Remove user from flow
	delete(fm.userFlows, ctx.UserID())
	return true, nil
}
