package teleflow

import (
	"fmt"
	"log"
	"maps"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type errorStrategy int

const (
	errorStrategyCancel errorStrategy = iota
	errorStrategyRetry
	errorStrategyIgnore
)

type ErrorConfig struct {
	Action  errorStrategy
	Message string
}

const ON_ERROR_SILENT = "__SILENT__"

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

type FlowConfig struct {
	ExitCommands        []string
	ExitMessage         string
	AllowGlobalCommands bool
	HelpCommands        []string
	OnProcessAction     ProcessMessageAction
}

type flowManager struct {
	flows        map[string]*Flow
	userFlows    map[int64]*userFlowState
	stateManager StateManager
	botConfig    *FlowConfig
	bot          *Bot
}

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

func (fm *flowManager) isUserInFlow(userID int64) bool {
	_, exists := fm.userFlows[userID]
	return exists
}

func (fm *flowManager) cancelFlow(userID int64) {
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

	fm.userFlows[userID] = userState

	if ctx != nil {
		return fm.renderStepPrompt(ctx, flow, flow.Order[0], userState)
	}

	return nil
}

func (fm *flowManager) renderStepPrompt(ctx *Context, flow *Flow, stepName string, userState *userFlowState) error {
	step := flow.Steps[stepName]
	if step == nil {
		return fmt.Errorf("step %s not found", stepName)
	}

	if step.PromptConfig == nil {
		return fmt.Errorf("step %s has no prompt configuration", stepName)
	}

	for key, value := range userState.Data {
		ctx.Set(key, value)
	}

	if fm.bot == nil || fm.bot.promptComposer == nil {
		return fmt.Errorf("PromptComposer not initialized - Bot not properly set")
	}

	err := fm.bot.promptComposer.composeAndSend(ctx, step.PromptConfig)
	if err != nil {
		return fm.handleRenderError(ctx, err, flow, stepName, userState)
	}

	return nil
}

func (fm *flowManager) HandleUpdate(ctx *Context) (bool, error) {
	userID := ctx.UserID()
	userState, exists := fm.userFlows[userID]
	if !exists {
		return false, nil
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

	userState.LastActive = time.Now()

	input, buttonClick := fm.extractInputData(ctx)

	for key, value := range userState.Data {
		ctx.Set(key, value)
	}

	if currentStep.ProcessFunc == nil {
		return true, fmt.Errorf("step %s has no process function", userState.CurrentStep)
	}

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

	maps.Copy(userState.Data, ctx.data)

	return fm.handleProcessResult(ctx, result, userState, flow)
}

func (fm *flowManager) extractInputData(ctx *Context) (string, *ButtonClick) {
	var input string
	var buttonClick *ButtonClick

	if ctx.update.Message != nil {
		input = ctx.update.Message.Text
	} else if ctx.update.CallbackQuery != nil {
		input = ctx.update.CallbackQuery.Data
		var originalData interface{} = input

		if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
			if mappedData, found := pkh.GetCallbackData(ctx.UserID(), input); found {
				originalData = mappedData
			}
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

func (fm *flowManager) handleProcessResult(ctx *Context, result ProcessResult, userState *userFlowState, flow *Flow) (bool, error) {

	if result.Prompt != nil {
		if err := fm.renderInformationalPrompt(ctx, result.Prompt); err != nil {

			return true, fm.handleRenderError(ctx, err, flow, userState.CurrentStep, userState)
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

func (fm *flowManager) renderInformationalPrompt(ctx *Context, config *PromptConfig) error {
	if fm.bot == nil || fm.bot.promptComposer == nil {
		return fmt.Errorf("PromptComposer not initialized - Bot not properly set")
	}

	infoPrompt := &PromptConfig{
		Message: config.Message,
		Image:   config.Image,
	}

	return fm.bot.promptComposer.composeAndSend(ctx, infoPrompt)
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

	return true, fm.renderStepPrompt(ctx, flow, nextStepName, userState)
}

func (fm *flowManager) goToSpecificStep(ctx *Context, userState *userFlowState, flow *Flow, targetStep string) (bool, error) {
	if _, exists := flow.Steps[targetStep]; !exists {
		return true, fmt.Errorf("target step %s not found in flow", targetStep)
	}

	userState.CurrentStep = targetStep
	return true, fm.renderStepPrompt(ctx, flow, targetStep, userState)
}

func (fm *flowManager) completeFlow(ctx *Context, flow *Flow) (bool, error) {

	if flow.OnComplete != nil {
		if err := flow.OnComplete(ctx); err != nil {
			delete(fm.userFlows, ctx.UserID())

			if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
				pkh.CleanupUserMappings(ctx.UserID())
			}
			return true, err
		}
	} else {
		log.Printf("[FLOW_COMPLETE] Flow %s called for user %d without completion handler", flow.Name, ctx.UserID())
		return true, fmt.Errorf("no completion handler defined for flow %s", flow.Name)
	}

	if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
		pkh.CleanupUserMappings(ctx.UserID())
	}

	delete(fm.userFlows, ctx.UserID())
	return true, nil
}

func (fm *flowManager) handleRenderError(ctx *Context, renderErr error, flow *Flow, stepName string, userState *userFlowState) error {

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

		fm.handleErrorStrategyCancel(ctx, &ErrorConfig{
			Action:  errorStrategyCancel,
			Message: "❗ A technical error occurred. Flow has been cancelled.",
		})
		return nil
	}
}

func (fm *flowManager) handleErrorStrategyCancel(ctx *Context, config *ErrorConfig) {

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

		if err := fm.bot.promptComposer.composeAndSend(ctx, fallbackPrompt); err != nil {

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

func (fm *flowManager) cancelFlowAction(ctx *Context) (bool, error) {

	if pkh := ctx.bot.GetPromptKeyboardHandler(); pkh != nil {
		pkh.CleanupUserMappings(ctx.UserID())
	}

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
	deleteConfig := tgbotapi.DeleteMessageConfig{
		ChatID:    ctx.ChatID(),
		MessageID: messageID,
	}

	_, err := ctx.bot.api.Request(deleteConfig)
	return err
}

func (fm *flowManager) deletePreviousKeyboard(ctx *Context, messageID int) error {
	editConfig := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    ctx.ChatID(),
			MessageID: messageID,
		},
	}

	_, err := ctx.bot.api.Request(editConfig)
	return err
}

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
