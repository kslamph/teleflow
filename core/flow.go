package teleflow

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// errorStrategy defines the internal enumeration of error handling strategies.
type errorStrategy int

const (
	errorStrategyCancel errorStrategy = iota
	errorStrategyRetry
	errorStrategyIgnore
)

// ErrorConfig defines how flows should handle errors during step processing.
// It specifies both the action to take and an optional user-facing message.
type ErrorConfig struct {
	Action  errorStrategy // The strategy to use when handling errors
	Message string        // Message to display to the user (optional)
}

// ON_ERROR_SILENT is a special constant that can be used as a message
// to indicate that no error message should be shown to the user.
const ON_ERROR_SILENT = "__SILENT__"

// OnErrorCancel creates an ErrorConfig that cancels the flow when an error occurs.
// This is the most conservative approach, stopping the flow immediately.
//
// Example:
//
//	flow := teleflow.NewFlow("example").
//		OnError(teleflow.OnErrorCancel("Something went wrong. Please try again later.")).
//		// ... define steps
//		Build()
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

// OnErrorRetry creates an ErrorConfig that retries the current step when an error occurs.
// This allows flows to recover from temporary issues automatically.
//
// Example:
//
//	flow := teleflow.NewFlow("example").
//		OnError(teleflow.OnErrorRetry("Please try again.")).
//		// ... define steps
//		Build()
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

// OnErrorIgnore creates an ErrorConfig that ignores errors and continues the flow.
// This should be used carefully as it may lead to unexpected behavior.
//
// Example:
//
//	flow := teleflow.NewFlow("example").
//		OnError(teleflow.OnErrorIgnore("Continuing despite error...")).
//		// ... define steps
//		Build()
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

// FlowConfig configures global flow behavior and command handling.
// It defines exit commands, help commands, and default message processing actions.
type FlowConfig struct {
	ExitCommands        []string             // Commands that exit any active flow
	ExitMessage         string               // Message shown when flow is exited
	AllowGlobalCommands bool                 // Whether global commands work during flows
	HelpCommands        []string             // Commands considered "help" commands
	OnProcessAction     ProcessMessageAction // Default action for processing messages
}

// flowManager manages all active conversation flows and their state.
// It handles flow registration, user state tracking, and flow execution.
// This is an internal component not exposed to bot users directly.
type flowManager struct {
	flows       map[string]*Flow         // Registered flows by name
	userFlows   map[int64]*userFlowState // Active user flow states
	muUserFlows sync.RWMutex             // Mutex for thread-safe flow operations
	flowConfig  *FlowConfig              // Global flow configuration

	promptSender   PromptSender          // Component for sending prompts
	keyboardAccess PromptKeyboardActions // Handler for keyboard interactions
	messageCleaner MessageCleaner        // Component for message management
}

func newFlowManager(config *FlowConfig, pSender PromptSender, kAccess PromptKeyboardActions, mCleaner MessageCleaner) *flowManager {
	return &flowManager{
		flows:          make(map[string]*Flow),
		userFlows:      make(map[int64]*userFlowState),
		flowConfig:     config,
		promptSender:   pSender,
		keyboardAccess: kAccess,
		messageCleaner: mCleaner,
	}
}

func (fm *flowManager) isUserInFlow(userID int64) bool {
	fm.muUserFlows.RLock()
	defer fm.muUserFlows.RUnlock()
	_, exists := fm.userFlows[userID]
	return exists
}

func (fm *flowManager) cancelFlow(userID int64) {
	fm.muUserFlows.Lock()
	defer fm.muUserFlows.Unlock()
	delete(fm.userFlows, userID)
}

type Flow struct {
	Name            string
	Steps           map[string]*flowStep
	Order           []string
	OnComplete      func(*Context) error
	OnError         *ErrorConfig
	OnProcessAction ProcessMessageAction
	Timeout         time.Duration
}

type flowStep struct {
	Name         string
	PromptConfig *PromptConfig
	ProcessFunc  ProcessFunc
}

type userFlowState struct {
	FlowName      string
	CurrentStep   string
	Data          map[string]interface{}
	StartedAt     time.Time
	LastActive    time.Time
	LastMessageID int
}

func (fm *flowManager) registerFlow(flow *Flow) {
	fm.flows[flow.Name] = flow
}

func (fm *flowManager) startFlow(userID int64, flowName string, ctx *Context) error {
	flow, exists := fm.flows[flowName]
	if !exists {
		return fmt.Errorf("flow %s not found", flowName)
	}

	if len(flow.Order) == 0 {
		return fmt.Errorf("flow %s has no steps", flowName)
	}

	initialData := make(map[string]interface{})
	if ctx != nil {
		for key, value := range ctx.data {
			initialData[key] = value
		}
	}

	userState := &userFlowState{
		FlowName:    flowName,
		CurrentStep: flow.Order[0],
		Data:        initialData,
		StartedAt:   time.Now(),
		LastActive:  time.Now(),
	}

	fm.muUserFlows.Lock()
	fm.userFlows[userID] = userState
	fm.muUserFlows.Unlock()

	if ctx != nil {
		return fm.renderStepPrompt(ctx, flow, flow.Order[0], userState)
	}

	return nil
}

func (fm *flowManager) renderStepPrompt(ctx *Context, flow *Flow, stepName string, userState *userFlowState) error {
	return fm.renderStepPrompt_nolock(ctx, flow, stepName, userState)
}

func (fm *flowManager) renderStepPrompt_withLockRelease(ctx *Context, flow *Flow, stepName string, userState *userFlowState) error {
	step := flow.Steps[stepName]
	if step == nil {
		return fmt.Errorf("step %s not found", stepName)
	}

	if step.PromptConfig == nil {
		return fmt.Errorf("step %s has no prompt configuration", stepName)
	}

	// Data copy removed - flow data should be accessed via GetFlowData() only

	// Release the mutex before prompt rendering to avoid deadlock
	// Prompt functions may call GetFlowData/SetFlowData which need the same mutex
	fm.muUserFlows.Unlock()

	err := fm.promptSender.ComposeAndSend(ctx, step.PromptConfig)

	// Re-acquire the mutex after prompt rendering
	fm.muUserFlows.Lock()

	if err != nil {
		return fm.handleRenderError_nolock(ctx, err, flow, stepName, userState)
	}

	return nil
}

func (fm *flowManager) renderStepPrompt_nolock(ctx *Context, flow *Flow, stepName string, userState *userFlowState) error {
	step := flow.Steps[stepName]
	if step == nil {
		return fmt.Errorf("step %s not found", stepName)
	}

	if step.PromptConfig == nil {
		return fmt.Errorf("step %s has no prompt configuration", stepName)
	}

	// Data copy removed - flow data should be accessed via GetFlowData() only

	err := fm.promptSender.ComposeAndSend(ctx, step.PromptConfig)

	if err != nil {
		return fm.handleRenderError_nolock(ctx, err, flow, stepName, userState)
	}

	return nil
}
func (fm *flowManager) HandleUpdate(ctx *Context) (bool, error) {
	// First, acquire lock to get flow state info
	fm.muUserFlows.Lock()

	userID := ctx.UserID()
	userState, exists := fm.userFlows[userID]
	if !exists {
		fm.muUserFlows.Unlock()
		return false, nil
	}

	flow := fm.flows[userState.FlowName]
	if flow == nil {
		delete(fm.userFlows, userID)
		fm.muUserFlows.Unlock()
		return false, fmt.Errorf("flow %s not found", userState.FlowName)
	}

	currentStep := flow.Steps[userState.CurrentStep]
	if currentStep == nil {
		delete(fm.userFlows, userID)
		fm.muUserFlows.Unlock()
		return false, fmt.Errorf("step %s not found", userState.CurrentStep)
	}

	userState.LastActive = time.Now()

	input, buttonClick := fm.extractInputData(ctx)

	// Data copy removed - flow data should be accessed via GetFlowData() only

	if currentStep.ProcessFunc == nil {
		fm.muUserFlows.Unlock()
		return true, fmt.Errorf("step %s has no process function", userState.CurrentStep)
	}

	// Release the lock before calling ProcessFunc to avoid deadlock
	// ProcessFunc might call SetFlowData which needs flowDataMutex
	fm.muUserFlows.Unlock()

	// Call ProcessFunc without holding any locks
	result := currentStep.ProcessFunc(ctx, input, buttonClick)

	if buttonClick != nil {
		if err := ctx.answerCallbackQuery(""); err != nil {

			_ = err
		}

		var messageIDToDelete int
		if ctx.update.CallbackQuery != nil && ctx.update.CallbackQuery.Message != nil {
			messageIDToDelete = ctx.update.CallbackQuery.Message.MessageID
		}

		if messageIDToDelete > 0 {
			if err := fm.handleMessageAction(ctx, flow, messageIDToDelete); err != nil {
				log.Printf("Error handling message action for UserID %d: %v", ctx.UserID(), err)

			}
		}
	}

	// Re-acquire lock for state modifications
	fm.muUserFlows.Lock()
	defer fm.muUserFlows.Unlock()

	// Re-check that user is still in flow (in case it was cancelled during ProcessFunc)
	userState, exists = fm.userFlows[userID]
	if !exists {
		return true, nil // Flow was cancelled, but we handled the update
	}

	// Bidirectional sync removed - use SetFlowData() to modify flow data

	return fm.handleProcessResult_nolock(ctx, result, userState, flow)
}

func (fm *flowManager) extractInputData(ctx *Context) (string, *ButtonClick) {
	var input string
	var buttonClick *ButtonClick

	if ctx.update.Message != nil {
		input = ctx.update.Message.Text
	} else if ctx.update.CallbackQuery != nil {
		input = ctx.update.CallbackQuery.Data
		var originalData interface{} = input

		if mappedData, found := fm.keyboardAccess.GetCallbackData(ctx.UserID(), input); found {
			originalData = mappedData
		}

		buttonClick = &ButtonClick{
			Data:     originalData,
			Text:     ctx.update.CallbackQuery.Message.Text,
			UserID:   ctx.UserID(),
			ChatID:   ctx.update.CallbackQuery.Message.Chat.ID,
			Metadata: make(map[string]interface{}),
		}
	}

	return input, buttonClick
}

func (fm *flowManager) handleProcessResult_nolock(ctx *Context, result ProcessResult, userState *userFlowState, flow *Flow) (bool, error) {

	if result.Prompt != nil {
		if err := fm.renderInformationalPrompt(ctx, result.Prompt); err != nil {

			return true, fm.handleRenderError_nolock(ctx, err, flow, userState.CurrentStep, userState)
		}
	}

	switch result.Action {
	case actionNextStep:
		return fm.advanceToNextStep(ctx, userState, flow)

	case actionGoToStep:
		return fm.goToSpecificStep(ctx, userState, flow, result.TargetStep)

	case actionRetryStep:

		if result.Prompt == nil {
			currentStep := flow.Steps[userState.CurrentStep]
			if currentStep != nil && currentStep.PromptConfig != nil {
				return true, fm.renderStepPrompt_withLockRelease(ctx, flow, userState.CurrentStep, userState)
			}
		}
		return true, nil

	case actionCompleteFlow:
		return fm.completeFlow_nolock(ctx, flow)

	case actionCancelFlow:
		return fm.cancelFlowAction_nolock(ctx)

	default:
		return true, fmt.Errorf("unknown ProcessAction: %d", result.Action)
	}
}

func (fm *flowManager) renderInformationalPrompt(ctx *Context, config *PromptConfig) error {

	infoPrompt := &PromptConfig{
		Message: config.Message,
		Image:   config.Image,
	}

	return fm.promptSender.ComposeAndSend(ctx, infoPrompt)
}

func (fm *flowManager) advanceToNextStep(ctx *Context, userState *userFlowState, flow *Flow) (bool, error) {

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

	if currentIndex+1 >= len(flow.Order) {

		return fm.completeFlow(ctx, flow)
	}

	nextStepName := flow.Order[currentIndex+1]
	userState.CurrentStep = nextStepName

	return true, fm.renderStepPrompt_withLockRelease(ctx, flow, nextStepName, userState)
}

func (fm *flowManager) goToSpecificStep(ctx *Context, userState *userFlowState, flow *Flow, targetStep string) (bool, error) {
	if _, exists := flow.Steps[targetStep]; !exists {
		return true, fmt.Errorf("target step %s not found in flow", targetStep)
	}

	userState.CurrentStep = targetStep
	return true, fm.renderStepPrompt_withLockRelease(ctx, flow, targetStep, userState)
}
func (fm *flowManager) completeFlow(ctx *Context, flow *Flow) (bool, error) {
	fm.muUserFlows.Lock()
	defer fm.muUserFlows.Unlock()
	return fm.completeFlow_nolock(ctx, flow)
}

func (fm *flowManager) completeFlow_nolock(ctx *Context, flow *Flow) (bool, error) {
	userID := ctx.UserID()
	var onCompleteErr error

	if flow.OnComplete != nil {
		// Release the lock before calling OnComplete to avoid deadlock
		// OnComplete handler may call GetFlowData/SetFlowData which need the same mutex
		fm.muUserFlows.Unlock()

		onCompleteErr = flow.OnComplete(ctx)

		// Re-acquire the lock after OnComplete completes
		fm.muUserFlows.Lock()
	} else {
		log.Printf("[FLOW_COMPLETE] Flow %s called for user %d without completion handler", flow.Name, userID)
	}

	// Always cleanup user flow and keyboard mappings regardless of OnComplete result
	fm.keyboardAccess.CleanupUserMappings(userID)
	delete(fm.userFlows, userID)

	// Return the OnComplete error if there was one
	if onCompleteErr != nil {
		return true, onCompleteErr
	}

	// If no OnComplete handler was defined, that's not an error - just complete successfully
	return true, nil
}

func (fm *flowManager) handleRenderError_nolock(ctx *Context, renderErr error, flow *Flow, stepName string, userState *userFlowState) error {

	fm.logRenderError(renderErr, stepName, flow.Name, ctx.UserID())

	action := errorStrategyCancel
	config := &ErrorConfig{
		Action:  errorStrategyCancel,
		Message: "❗ A technical error occurred. Flow has been cancelled.",
	}

	if flow.OnError != nil {
		action = flow.OnError.Action
		config = flow.OnError
	}

	log.Printf("[FLOW_ERROR_ACTION] Flow: %s, Step: %s, User: %d, Action: %s",
		flow.Name, stepName, ctx.UserID(), fm.getActionName(action))

	switch action {
	case errorStrategyCancel:
		fm.handleErrorStrategyCancel_nolock(ctx, config)
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

		fm.handleErrorStrategyCancel_nolock(ctx, &ErrorConfig{
			Action:  errorStrategyCancel,
			Message: "❗ A technical error occurred. Flow has been cancelled.",
		})
		return nil
	}
}

func (fm *flowManager) handleErrorStrategyCancel_nolock(ctx *Context, config *ErrorConfig) {

	fm.notifyUserIfNeeded(ctx, config.Message)
	delete(fm.userFlows, ctx.UserID())
}

func (fm *flowManager) handleErrorStrategyRetry(ctx *Context, config *ErrorConfig) {
	fm.notifyUserIfNeeded(ctx, config.Message)

}

func (fm *flowManager) handleErrorStrategyIgnore(ctx *Context, config *ErrorConfig, originalPrompt *PromptConfig, userState *userFlowState, flow *Flow) error {
	fm.notifyUserIfNeeded(ctx, config.Message)

	if originalPrompt != nil {

		fallbackPrompt := &PromptConfig{
			Message:  originalPrompt.Message,
			Keyboard: originalPrompt.Keyboard,
		}

		if err := fm.promptSender.ComposeAndSend(ctx, fallbackPrompt); err != nil {

			_, err := fm.advanceToNextStep(ctx, userState, flow)
			return err
		}
	}
	return nil
}

func (fm *flowManager) logRenderError(err error, stepName, flowName string, userID int64) {
	log.Printf("[FLOW_RENDER_ERROR] Flow: %s, Step: %s, User: %d, Error: %v",
		flowName, stepName, userID, err)
}

func (fm *flowManager) notifyUserIfNeeded(ctx *Context, message string) {
	if message == ON_ERROR_SILENT {
		return
	}

	ctx.Set("__render_parse_mode", ParseModeNone)

	if err := ctx.sendSimpleText(message); err != nil {
		log.Printf("[FLOW_ERROR_NOTIFY_FAILED] Failed to notify user %d: %v", ctx.UserID(), err)
	}
}

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

func (fm *flowManager) cancelFlowAction_nolock(ctx *Context) (bool, error) {

	fm.keyboardAccess.CleanupUserMappings(ctx.UserID())

	delete(fm.userFlows, ctx.UserID())
	return true, nil
}

func (fm *flowManager) handleMessageAction(ctx *Context, flow *Flow, messageID int) error {
	switch flow.OnProcessAction {
	case ProcessDeleteMessage:
		return fm.deletePreviousMessage(ctx, messageID)
	case ProcessDeleteKeyboard:
		return fm.deletePreviousKeyboard(ctx, messageID)
	case ProcessKeepMessage:

		return nil
	default:

		return nil
	}
}

func (fm *flowManager) deletePreviousMessage(ctx *Context, messageID int) error {
	return fm.messageCleaner.DeleteMessage(ctx, messageID)
}

func (fm *flowManager) deletePreviousKeyboard(ctx *Context, messageID int) error {
	return fm.messageCleaner.EditMessageReplyMarkup(ctx, messageID, nil)
}
func (fm *flowManager) setUserFlowData(userID int64, key string, value interface{}) error {
	fm.muUserFlows.Lock()
	defer fm.muUserFlows.Unlock()

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

func (fm *flowManager) getUserFlowData(userID int64, key string) (interface{}, bool) {
	fm.muUserFlows.RLock()
	defer fm.muUserFlows.RUnlock()

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
