package teleflow

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Flow system enables the creation of sophisticated multi-step conversational
// interfaces using the new Step-Prompt-Process API. This system provides:
//   - Declarative prompt configuration (Message, Image, Keyboard)
//   - Unified input processing with ProcessFunc
//   - Clear flow control with ProcessResult actions
//   - Simplified developer experience with zero learning curve
//   - Built-in error handling with configurable recovery strategies

// errorStrategy defines how to handle runtime errors during flow execution
type errorStrategy int

const (
	errorStrategyCancel errorStrategy = iota // Cancel flow (default)
	errorStrategyRetry                       // Retry current step
	errorStrategyIgnore                      // Continue with flow
)

// ErrorConfig defines error handling behavior for flows
type ErrorConfig struct {
	Action  errorStrategy
	Message string // Custom message or ON_ERROR_SILENT for silent operation
}

// Predefined constant for silent error handling
const ON_ERROR_SILENT = "__SILENT__"

// Convenience constructors for error handling configuration

// OnErrorCancel cancels the flow when an error occurs (default behavior)
func OnErrorCancel(message ...string) *ErrorConfig {
	msg := "❗ A technical error occurred. Flow has been cancelled."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &ErrorConfig{
		Action:  errorStrategyCancel,
		Message: msg,
	}
}

// OnErrorRetry retries the current step when an error occurs
func OnErrorRetry(message ...string) *ErrorConfig {
	msg := "🔄 A technical error occurred. Retrying current step..."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &ErrorConfig{
		Action:  errorStrategyRetry,
		Message: msg,
	}
}

// OnErrorIgnore ignores the error and continues with the flow
func OnErrorIgnore(message ...string) *ErrorConfig {
	msg := "⚠️ A technical error occurred. Continuing with flow..."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &ErrorConfig{
		Action:  errorStrategyIgnore,
		Message: msg,
	}
}

// FlowConfig holds configuration for flow behavior
type FlowConfig struct {
	ExitCommands        []string
	ExitMessage         string
	AllowGlobalCommands bool
	HelpCommands        []string
	OnProcessAction     ProcessMessageAction // How to handle previous messages on button clicks
}

// flowManager manages all flows and user flow states
type flowManager struct {
	flows        map[string]*Flow
	userFlows    map[int64]*userFlowState
	stateManager StateManager
	botConfig    *FlowConfig
	bot          *Bot // Reference to bot for accessing promptComposer
}

// newFlowManager creates a new flow manager
func newFlowManager(stateManager StateManager) *flowManager {
	return &flowManager{
		flows:        make(map[string]*Flow),
		userFlows:    make(map[int64]*userFlowState),
		stateManager: stateManager,
	}
}

func (fm *flowManager) initialize(bot *Bot) {
	fm.bot = bot
	fm.botConfig = &bot.flowConfig
}

// isUserInFlow checks if a user is currently in a flow
func (fm *flowManager) isUserInFlow(userID int64) bool {
	_, exists := fm.userFlows[userID]
	return exists
}

// cancelFlow cancels the current flow for a user
func (fm *flowManager) cancelFlow(userID int64) {
	delete(fm.userFlows, userID)
}

// Flow represents a structured multi-step conversation using the new Step-Prompt-Process API
type Flow struct {
	Name            string
	Steps           map[string]*flowStep // Map for easier lookup by step name
	Order           []string             // Maintains step execution order
	OnComplete      func(*Context) error // Simplified completion handler
	OnError         *ErrorConfig         // Flow-level error handling configuration
	OnProcessAction ProcessMessageAction // How to handle previous messages on button clicks
	Timeout         time.Duration
}

// flowStep represents a single step in a flow using the new API
type flowStep struct {
	Name         string
	PromptConfig *PromptConfig        // New: declarative prompt specification
	ProcessFunc  ProcessFunc          // New: unified input processing
	OnComplete   func(*Context) error // Optional step completion handler
	Timeout      time.Duration
}

// userFlowState tracks a user's current position in a flow
type userFlowState struct {
	FlowName      string
	CurrentStep   string
	Data          map[string]interface{}
	StartedAt     time.Time
	LastActive    time.Time
	LastMessageID int // Track last message ID for deletion options
}

// registerFlow registers a flow with the manager
func (fm *flowManager) registerFlow(flow *Flow) {
	fm.flows[flow.Name] = flow
}

// startFlow starts a flow for a user
func (fm *flowManager) startFlow(userID int64, flowName string, ctx *Context) error {
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

	userState := &userFlowState{
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
func (fm *flowManager) renderStepPrompt(ctx *Context, flow *Flow, stepName string, userState *userFlowState) error {
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

	// Use PromptComposer instead of promptRenderer
	if fm.bot == nil || fm.bot.promptComposer == nil {
		return fmt.Errorf("PromptComposer not initialized - Bot not properly set")
	}

	// Attempt to render with error handling
	err := fm.bot.promptComposer.ComposeAndSend(ctx, step.PromptConfig)
	if err != nil {
		return fm.handleRenderError(ctx, err, flow, stepName, userState)
	}

	return nil
}

// HandleUpdate processes an update for a user in a flow using the new API
func (fm *flowManager) HandleUpdate(ctx *Context) (bool, error) {
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

		// Handle previous message actions based on flow configuration
		// Use the message ID from the callback query (the message that contains the clicked button)
		var messageIDToDelete int
		if ctx.update.CallbackQuery != nil && ctx.update.CallbackQuery.Message != nil {
			messageIDToDelete = ctx.update.CallbackQuery.Message.MessageID
		}

		if messageIDToDelete > 0 {
			if err := fm.handleMessageAction(ctx, flow, messageIDToDelete); err != nil {
				log.Printf("Error handling message action for UserID %d: %v", ctx.UserID(), err)
				// Continue processing even if message deletion fails
			}
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
func (fm *flowManager) extractInputData(ctx *Context) (string, *ButtonClick) {
	var input string
	var buttonClick *ButtonClick

	if ctx.update.Message != nil {
		input = ctx.update.Message.Text
	} else if ctx.update.CallbackQuery != nil {
		input = ctx.update.CallbackQuery.Data // This is the UUID string
		var originalData interface{} = input  // Default to UUID if not found

		// Use PromptKeyboardHandler (assuming fm has access to it, perhaps via Bot)
		if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
			if mappedData, found := pkh.GetCallbackData(ctx.UserID(), input); found {
				originalData = mappedData
			}
		}

		buttonClick = &ButtonClick{
			Data:     originalData, // Now contains the original interface{} data
			Text:     ctx.update.CallbackQuery.Message.Text,
			UserID:   ctx.UserID(),
			ChatID:   ctx.update.CallbackQuery.Message.Chat.ID,
			Metadata: make(map[string]interface{}),
		}
	}

	return input, buttonClick
}

// handleProcessResult processes the result from a ProcessFunc
func (fm *flowManager) handleProcessResult(ctx *Context, result ProcessResult, userState *userFlowState, flow *Flow) (bool, error) {
	// Show custom prompt if specified (informational only, no keyboard)
	if result.Prompt != nil {
		if err := fm.renderInformationalPrompt(ctx, result.Prompt); err != nil {
			// Apply error handling to informational prompts as well
			return true, fm.handleRenderError(ctx, err, flow, userState.CurrentStep, userState)
		}
	}

	// Execute action
	switch result.Action {
	case actionNextStep:
		return fm.advanceToNextStep(ctx, userState, flow)

	case actionGoToStep:
		return fm.goToSpecificStep(ctx, userState, flow, result.TargetStep)

	case actionRetryStep:
		// Stay on current step, optionally re-render prompt
		if result.Prompt == nil {
			currentStep := flow.Steps[userState.CurrentStep]
			if currentStep != nil && currentStep.PromptConfig != nil {
				return true, fm.renderStepPrompt(ctx, flow, userState.CurrentStep, userState)
			}
		}
		return true, nil

	case actionCompleteFlow:
		return fm.completeFlow(ctx, flow)

	case actionCancelFlow:
		return fm.cancelFlowAction(ctx)

	default:
		return true, fmt.Errorf("unknown ProcessAction: %d", result.Action)
	}
}

// renderInformationalPrompt renders a prompt without keyboard for informational messages
func (fm *flowManager) renderInformationalPrompt(ctx *Context, config *PromptConfig) error {
	if fm.bot == nil || fm.bot.promptComposer == nil {
		return fmt.Errorf("PromptComposer not initialized - Bot not properly set")
	}

	// Create informational prompt without keyboard
	infoPrompt := &PromptConfig{
		Message: config.Message,
		Image:   config.Image,
		// Keyboard is intentionally omitted for informational messages
	}

	return fm.bot.promptComposer.ComposeAndSend(ctx, infoPrompt)
}

// advanceToNextStep moves to the next step in sequence
func (fm *flowManager) advanceToNextStep(ctx *Context, userState *userFlowState, flow *Flow) (bool, error) {
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
		return fm.completeFlow(ctx, flow)
	}

	// Move to next step
	nextStepName := flow.Order[currentIndex+1]
	userState.CurrentStep = nextStepName

	// Render the next step's prompt
	return true, fm.renderStepPrompt(ctx, flow, nextStepName, userState)
}

// goToSpecificStep jumps to a named step
func (fm *flowManager) goToSpecificStep(ctx *Context, userState *userFlowState, flow *Flow, targetStep string) (bool, error) {
	if _, exists := flow.Steps[targetStep]; !exists {
		return true, fmt.Errorf("target step %s not found in flow", targetStep)
	}

	userState.CurrentStep = targetStep
	return true, fm.renderStepPrompt(ctx, flow, targetStep, userState)
}

// completeFlow handles flow completion
func (fm *flowManager) completeFlow(ctx *Context, flow *Flow) (bool, error) {
	// Execute completion handler if it exists
	if flow.OnComplete != nil {
		if err := flow.OnComplete(ctx); err != nil {
			delete(fm.userFlows, ctx.UserID())
			// Clean up UUID mappings when flow is cancelled due to error
			if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
				pkh.CleanupUserMappings(ctx.UserID())
			}
			return true, err
		}
	}

	// Clean up UUID mappings when flow completes
	if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
		pkh.CleanupUserMappings(ctx.UserID())
	}

	// Remove user from flow
	delete(fm.userFlows, ctx.UserID())
	return true, nil
}

// handleRenderError handles errors that occur during prompt rendering
func (fm *flowManager) handleRenderError(ctx *Context, renderErr error, flow *Flow, stepName string, userState *userFlowState) error {
	// Always log the technical error with full context
	fm.logRenderError(renderErr, stepName, flow.Name, ctx.UserID())

	// Determine error handling strategy
	action := errorStrategyCancel // default behavior
	message := "❗ A technical error occurred. Flow has been cancelled."

	if flow.OnError != nil {
		action = flow.OnError.Action
		if flow.OnError.Message != "" {
			message = flow.OnError.Message
		}
	}

	// Log the action being taken
	log.Printf("[FLOW_ERROR_ACTION] Flow: %s, Step: %s, User: %d, Action: %s",
		flow.Name, stepName, ctx.UserID(), fm.getActionName(action))

	// Execute the configured action
	switch action {
	case errorStrategyCancel:
		fm.notifyUserIfNeeded(ctx, message)
		delete(fm.userFlows, ctx.UserID())
		return nil

	case errorStrategyRetry:
		fm.notifyUserIfNeeded(ctx, message)
		// Stay on current step - next update will retry the render
		return nil

	case errorStrategyIgnore:
		fm.notifyUserIfNeeded(ctx, message)
		// Try to render the step again without the problematic image
		step := flow.Steps[stepName]
		if step != nil && step.PromptConfig != nil {
			// Create a fallback prompt without image to avoid repeated errors
			fallbackPrompt := &PromptConfig{
				Message:  step.PromptConfig.Message,
				Keyboard: step.PromptConfig.Keyboard,
				// Image is intentionally omitted to avoid repeated render errors
			}

			// Try to render without image using PromptComposer - if this fails, we'll advance to next step
			if err := fm.bot.promptComposer.ComposeAndSend(ctx, fallbackPrompt); err != nil {
				// If even the fallback fails, advance to next step
				_, err := fm.advanceToNextStep(ctx, userState, flow)
				return err
			}
		}
		return nil

	default:
		// Fallback to cancel if unknown action
		fm.notifyUserIfNeeded(ctx, "❗ A technical error occurred. Flow has been cancelled.")
		delete(fm.userFlows, ctx.UserID())
		return nil
	}
}

// logRenderError logs detailed information about rendering errors
func (fm *flowManager) logRenderError(err error, stepName, flowName string, userID int64) {
	log.Printf("[FLOW_RENDER_ERROR] Flow: %s, Step: %s, User: %d, Error: %v",
		flowName, stepName, userID, err)
}

// notifyUserIfNeeded sends a notification to the user unless silent mode is enabled
func (fm *flowManager) notifyUserIfNeeded(ctx *Context, message string) {
	if message == ON_ERROR_SILENT {
		return
	}

	// Clear any parse mode from template rendering before sending error notification
	// This prevents MarkdownV2 parsing errors when sending plain text error messages
	ctx.Set("__render_parse_mode", ParseModeNone)

	// Use a simple text reply for error notifications to avoid additional render errors
	if err := ctx.Reply(message); err != nil {
		log.Printf("[FLOW_ERROR_NOTIFY_FAILED] Failed to notify user %d: %v", ctx.UserID(), err)
	}
}

// getActionName returns a human-readable name for the error action (for logging)
func (fm *flowManager) getActionName(action errorStrategy) string {
	switch action {
	case errorStrategyCancel:
		return "CANCEL"
	case errorStrategyRetry:
		return "RETRY"
	case errorStrategyIgnore:
		return "IGNORE"
	default:
		return "UNKNOWN"
	}
}

// cancelFlowAction handles flow cancellation
func (fm *flowManager) cancelFlowAction(ctx *Context) (bool, error) {
	// Clean up UUID mappings when flow is cancelled
	if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
		pkh.CleanupUserMappings(ctx.UserID())
	}

	// Remove user from flow
	delete(fm.userFlows, ctx.UserID())
	return true, nil
}

// handleMessageAction handles deletion of a specific message based on flow configuration
func (fm *flowManager) handleMessageAction(ctx *Context, flow *Flow, messageID int) error {
	switch flow.OnProcessAction {
	case ProcessDeleteMessage:
		return fm.deletePreviousMessage(ctx, messageID)
	case ProcessDeleteKeyboard:
		return fm.deletePreviousKeyboard(ctx, messageID)
	case ProcessKeepMessage:
		// Do nothing - keep messages untouched
		return nil
	default:
		// Default behavior - keep messages untouched
		return nil
	}
}

// deletePreviousMessage completely deletes the previous message
func (fm *flowManager) deletePreviousMessage(ctx *Context, messageID int) error {
	deleteConfig := tgbotapi.DeleteMessageConfig{
		ChatID:    ctx.ChatID(),
		MessageID: messageID,
	}

	_, err := ctx.bot.api.Request(deleteConfig)
	return err
}

// deletePreviousKeyboard removes only the keyboard from the previous message
func (fm *flowManager) deletePreviousKeyboard(ctx *Context, messageID int) error {
	editConfig := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    ctx.ChatID(),
			MessageID: messageID,
		},
		// ReplyMarkup field is omitted to remove keyboard
	}

	_, err := ctx.bot.api.Request(editConfig)
	return err
}

// setUserFlowData sets flow-specific data for a user
func (fm *flowManager) setUserFlowData(userID int64, key string, value interface{}) error {
	userState, exists := fm.userFlows[userID]
	if !exists {
		return fmt.Errorf("user %d not in a flow", userID)
	}

	if userState.Data == nil {
		userState.Data = make(map[string]interface{})
	}

	userState.Data[key] = value
	return nil
}

// getUserFlowData gets flow-specific data for a user
func (fm *flowManager) getUserFlowData(userID int64, key string) (interface{}, bool) {
	userState, exists := fm.userFlows[userID]
	if !exists {
		return nil, false
	}

	if userState.Data == nil {
		return nil, false
	}

	value, ok := userState.Data[key]
	return value, ok
}
