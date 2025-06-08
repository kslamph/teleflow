package teleflow

import (
	"fmt"
	"log"
	"maps"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Package core/flow.go implements the Step-Prompt-Process API runtime for multi-step conversational flows.
//
// The Flow System provides a declarative approach to building sophisticated conversational interfaces
// with built-in error handling, state management, and lifecycle control. The runtime manages flow
// execution, user state tracking, prompt rendering, and input processing through a unified API.
//
// Key Runtime Components:
//   - Flow execution engine with step sequencing and navigation
//   - User state management with automatic context synchronization
//   - Error handling with configurable recovery strategies (cancel, retry, ignore)
//   - Message action processing for clean UI interactions
//   - Automatic cleanup of keyboard mappings and user sessions
//
// Example Usage:
//   // Error handling configuration
//   flow := NewFlow("example").
//     OnError(OnErrorRetry("Please try again")).
//     Step("input").Prompt("Enter name:").Process(func(ctx *Context, input string, click *ButtonClick) ProcessResult {
//       return NextStep()
//     }).Build()
//
//   bot.RegisterFlow(flow)
//   ctx.StartFlow("example")

// errorStrategy defines how runtime errors are handled during flow execution.
// The strategy determines whether to cancel the flow, retry the current step,
// or ignore the error and continue processing.
type errorStrategy int

const (
	errorStrategyCancel errorStrategy = iota // Cancel flow immediately (default behavior)
	errorStrategyRetry                       // Retry current step, re-render prompt
	errorStrategyIgnore                      // Continue execution, skip problematic step
)

// ErrorConfig defines comprehensive error handling behavior for flows.
// It specifies both the recovery strategy and user notification message.
// Used for flow-level error configuration through OnError() methods.
type ErrorConfig struct {
	Action  errorStrategy // Recovery strategy to apply when errors occur
	Message string        // User notification message, or ON_ERROR_SILENT for silent handling
}

// ON_ERROR_SILENT is a special constant for silent error handling.
// When used as ErrorConfig.Message, no user notification is sent during error recovery.
const ON_ERROR_SILENT = "__SILENT__"

// OnErrorCancel creates an ErrorConfig that cancels the flow when rendering errors occur.
// This is the default error handling strategy that provides safe failure behavior.
//
// Parameters:
//   - message: Optional custom error message. If empty, uses default cancellation message.
//
// Returns:
//   - *ErrorConfig configured for flow cancellation with user notification.
//
// Example:
//
//	flow := NewFlow("example").OnError(OnErrorCancel("Custom error occurred")).Build()
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

// OnErrorRetry creates an ErrorConfig that retries the current step when rendering errors occur.
// The step's prompt is re-rendered, giving the user another opportunity to interact with it.
// Useful for temporary issues like network connectivity or transient rendering problems.
//
// Parameters:
//   - message: Optional custom retry message. If empty, uses default retry message.
//
// Returns:
//   - *ErrorConfig configured for step retry with user notification.
//
// Example:
//
//	flow := NewFlow("example").OnError(OnErrorRetry("Retrying...")).Build()
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

// OnErrorIgnore creates an ErrorConfig that ignores rendering errors and continues flow execution.
// When errors occur, a fallback prompt (without problematic elements) is rendered, or the flow
// advances to the next step. Useful for non-critical rendering issues where flow continuation is preferred.
//
// Parameters:
//   - message: Optional custom warning message. If empty, uses default warning message.
//
// Returns:
//   - *ErrorConfig configured to ignore errors with optional user notification.
//
// Example:
//
//	flow := NewFlow("example").OnError(OnErrorIgnore("Minor issue, continuing...")).Build()
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

// FlowConfig holds global configuration for flow behavior across the entire bot.
// It defines bot-wide settings that affect all flows, including exit commands,
// global command handling, and default message processing behavior.
type FlowConfig struct {
	ExitCommands        []string             // Commands that cancel any active flow (e.g., "/cancel", "/exit")
	ExitMessage         string               // Message shown when user exits a flow via command
	AllowGlobalCommands bool                 // Whether global commands are processed during flows
	HelpCommands        []string             // Commands that show help without affecting flow state
	OnProcessAction     ProcessMessageAction // Default behavior for handling previous messages on button clicks
}

// flowManager manages all registered flows and tracks user flow states across the bot.
// It serves as the central orchestrator for flow execution, handling registration,
// lifecycle management, user state tracking, and error recovery. The manager coordinates
// between flows, user sessions, and the bot's prompt rendering system.
type flowManager struct {
	flows        map[string]*Flow         // Registry of all available flows by name
	userFlows    map[int64]*userFlowState // Active user sessions mapped by user ID
	stateManager StateManager             // Persistent state storage interface
	botConfig    *FlowConfig              // Global flow configuration settings
	bot          *Bot                     // Reference to parent bot for accessing promptComposer and other services
}

// newFlowManager creates a new flow manager with the specified state manager.
// Initializes empty registries for flows and user sessions, preparing the manager
// for flow registration and user interaction handling.
//
// Parameters:
//   - stateManager: StateManager interface for persistent state storage
//
// Returns:
//   - *flowManager: New flow manager instance ready for initialization
func newFlowManager(stateManager StateManager) *flowManager {
	return &flowManager{
		flows:        make(map[string]*Flow),
		userFlows:    make(map[int64]*userFlowState),
		stateManager: stateManager,
	}
}

// initialize completes flow manager setup by linking it to the parent bot.
// This method provides access to the bot's prompt composer and configuration,
// enabling full flow execution capabilities.
//
// Parameters:
//   - bot: Parent Bot instance providing prompt composer and configuration
func (fm *flowManager) initialize(bot *Bot) {
	fm.bot = bot
	fm.botConfig = &bot.flowConfig
}

// isUserInFlow checks whether a specific user is currently participating in any active flow.
// Used to determine if incoming updates should be processed by the flow system.
//
// Parameters:
//   - userID: Telegram user ID to check
//
// Returns:
//   - bool: true if user has an active flow session, false otherwise
func (fm *flowManager) isUserInFlow(userID int64) bool {
	_, exists := fm.userFlows[userID]
	return exists
}

// cancelFlow immediately cancels and removes any active flow session for the specified user.
// This cleans up user state but does not send notifications or handle cleanup callbacks.
// Use for emergency cancellation or when user state becomes invalid.
//
// Parameters:
//   - userID: Telegram user ID whose flow should be cancelled
func (fm *flowManager) cancelFlow(userID int64) {
	delete(fm.userFlows, userID)
}

// Flow represents a complete multi-step conversational flow built with the Step-Prompt-Process API.
// Each flow encapsulates a series of interactive steps that guide users through a structured process,
// with comprehensive error handling, lifecycle management, and customizable behavior for message processing.
type Flow struct {
	Name            string               // Unique identifier for the flow, used for registration and starting
	Steps           map[string]*flowStep // Map of all steps indexed by name for efficient lookup
	Order           []string             // Ordered list of step names defining execution sequence
	OnComplete      func(*Context) error // Optional handler executed when flow completes successfully
	OnError         *ErrorConfig         // Flow-level error handling strategy for rendering failures
	OnProcessAction ProcessMessageAction // Defines how previous messages are handled during button interactions
	Timeout         time.Duration        // Maximum duration before flow auto-cancellation
}

// flowStep represents a single interactive step within a flow using the Step-Prompt-Process API.
// Each step defines what the user sees (prompt), how input is processed (function), and optional
// completion handling. Steps are executed sequentially according to the flow's Order.
type flowStep struct {
	Name         string        // Unique step identifier within the flow
	PromptConfig *PromptConfig // Declarative specification of what to display (message, image, keyboard)
	ProcessFunc  ProcessFunc   // Function that processes user input and determines next action
}

// userFlowState tracks an individual user's current position and data within an active flow.
// This state is maintained throughout the flow execution and automatically synchronized
// with the Context for seamless data access in prompts and process functions.
type userFlowState struct {
	FlowName      string                 // Name of the currently active flow
	CurrentStep   string                 // Name of the current step being executed
	Data          map[string]interface{} // User-specific data accumulated during flow execution
	StartedAt     time.Time              // Timestamp when the flow was initiated
	LastActive    time.Time              // Timestamp of the most recent user interaction
	LastMessageID int                    // Message ID of the last sent message for cleanup operations
}

// registerFlow registers a flow with the manager, making it available for user interaction.
// The flow must have a unique name and be fully configured before registration.
// Once registered, the flow can be started using Context.StartFlow() or bot.StartFlow().
//
// Parameters:
//   - flow: Fully configured Flow instance to register
func (fm *flowManager) registerFlow(flow *Flow) {
	fm.flows[flow.Name] = flow
}

// startFlow initiates a new flow session for the specified user with optional initial context data.
// Creates user state, validates flow configuration, and renders the first step's prompt.
// If context data is provided, it's copied into the flow's initial state for use in prompts and processing.
//
// Parameters:
//   - userID: Telegram user ID to start the flow for
//   - flowName: Name of the registered flow to start
//   - ctx: Context containing optional initial data and providing rendering capabilities
//
// Returns:
//   - error: any error that occurred during flow initialization or first step rendering
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

// renderStepPrompt renders the declarative prompt configuration for a specific step.
// Loads user state data into the context, then uses the bot's PromptComposer to render
// the step's message, image, and keyboard. Applies comprehensive error handling based
// on the flow's OnError configuration if rendering fails.
//
// Parameters:
//   - ctx: Context for rendering and user interaction
//   - flow: Flow containing the step to render
//   - stepName: Name of the step whose prompt should be rendered
//   - userState: Current user state for data synchronization
//
// Returns:
//   - error: any error that occurred during prompt rendering, after applying error handling strategies
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
	err := fm.bot.promptComposer.composeAndSend(ctx, step.PromptConfig)
	if err != nil {
		return fm.handleRenderError(ctx, err, flow, stepName, userState)
	}

	return nil
}

// HandleUpdate is the central method that processes incoming Telegram updates for users in active flows.
// This method orchestrates the complete flow execution cycle: input extraction, state management,
// step processing, result handling, and error recovery. It returns whether the update was handled
// by the flow system and any errors that occurred during processing.
//
// Processing Flow:
//  1. Validates user has active flow session
//  2. Extracts input data (text messages or button clicks)
//  3. Synchronizes user state data with Context
//  4. Executes current step's ProcessFunc
//  5. Handles button click callbacks and message cleanup
//  6. Processes returned ProcessResult for flow navigation
//  7. Applies error handling strategies if issues occur
//
// Parameters:
//   - ctx: Context containing the Telegram update and user session
//
// Returns:
//   - bool: true if update was processed by flow system, false if user not in flow
//   - error: any error that occurred during processing
//
// Example Usage:
//
//	handled, err := flowManager.HandleUpdate(ctx)
//	if handled {
//	  return // Update was processed by flow system
//	}
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
	maps.Copy(userState.Data, ctx.data)

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

	return fm.bot.promptComposer.composeAndSend(ctx, infoPrompt)
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
	} else {
		log.Printf("[FLOW_COMPLETE] Flow %s called for user %d without completion handler", flow.Name, ctx.UserID())
		return true, fmt.Errorf("no completion handler defined for flow %s", flow.Name)
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
	config := &ErrorConfig{
		Action:  errorStrategyCancel,
		Message: "❗ A technical error occurred. Flow has been cancelled.",
	}

	if flow.OnError != nil {
		action = flow.OnError.Action
		config = flow.OnError
	}

	// Log the action being taken
	log.Printf("[FLOW_ERROR_ACTION] Flow: %s, Step: %s, User: %d, Action: %s",
		flow.Name, stepName, ctx.UserID(), fm.getActionName(action))

	// Dispatch to appropriate handler based on strategy
	switch action {
	case errorStrategyCancel:
		fm.handleErrorStrategyCancel(ctx, config)
		return nil

	case errorStrategyRetry:
		fm.handleErrorStrategyRetry(ctx, config)
		return nil

	case errorStrategyIgnore:
		step := flow.Steps[stepName]
		var originalPrompt *PromptConfig
		if step != nil {
			originalPrompt = step.PromptConfig
		}
		return fm.handleErrorStrategyIgnore(ctx, config, originalPrompt, userState, flow)

	default:
		// Fallback to cancel if unknown action
		fm.handleErrorStrategyCancel(ctx, &ErrorConfig{
			Action:  errorStrategyCancel,
			Message: "❗ A technical error occurred. Flow has been cancelled.",
		})
		return nil
	}
}

// handleErrorStrategyCancel handles the cancel error strategy
func (fm *flowManager) handleErrorStrategyCancel(ctx *Context, config *ErrorConfig) {
	// Log the cancellation

	fm.notifyUserIfNeeded(ctx, config.Message)
	delete(fm.userFlows, ctx.UserID())
}

// handleErrorStrategyRetry handles the retry error strategy
func (fm *flowManager) handleErrorStrategyRetry(ctx *Context, config *ErrorConfig) {
	fm.notifyUserIfNeeded(ctx, config.Message)
	// Stay on current step - next update will retry the render
}

// handleErrorStrategyIgnore handles the ignore error strategy
func (fm *flowManager) handleErrorStrategyIgnore(ctx *Context, config *ErrorConfig, originalPrompt *PromptConfig, userState *userFlowState, flow *Flow) error {
	fm.notifyUserIfNeeded(ctx, config.Message)

	// Try to render the step again without the problematic image
	if originalPrompt != nil {
		// Create a fallback prompt without image to avoid repeated errors
		fallbackPrompt := &PromptConfig{
			Message:  originalPrompt.Message,
			Keyboard: originalPrompt.Keyboard,
			// Image is intentionally omitted to avoid repeated render errors
		}

		// Try to render without image using PromptComposer - if this fails, we'll advance to next step
		if err := fm.bot.promptComposer.composeAndSend(ctx, fallbackPrompt); err != nil {
			// If even the fallback fails, advance to next step
			_, err := fm.advanceToNextStep(ctx, userState, flow)
			return err
		}
	}
	return nil
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
	if err := ctx.sendSimpleText(message); err != nil {
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
