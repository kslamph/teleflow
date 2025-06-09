package teleflow

import "time"

type FlowBuilder struct {
	name            string
	steps           map[string]*StepBuilder
	order           []string
	onComplete      func(*Context) error
	onError         *ErrorConfig
	onProcessAction ProcessMessageAction
	currentStep     *StepBuilder
	timeout         time.Duration
}

type StepBuilder struct {
	name         string
	promptConfig *PromptConfig
	processFunc  ProcessFunc
	flowBuilder  *FlowBuilder
}

type PromptConfig struct {
	Message      MessageSpec
	Image        ImageSpec
	Keyboard     KeyboardFunc
	TemplateData map[string]interface{}
}

type MessageSpec interface{}

type ImageSpec interface{}

type KeyboardFunc func(ctx *Context) *PromptKeyboardBuilder

type ProcessFunc func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult

type ButtonClick struct {
	Data     interface{}
	Text     string
	UserID   int64
	ChatID   int64
	Metadata map[string]interface{}
}

type ProcessResult struct {
	Action     processAction
	TargetStep string
	Prompt     *PromptConfig
}

func (pr ProcessResult) WithPrompt(prompt MessageSpec) ProcessResult {

	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Message = prompt

	return pr
}

func (pr ProcessResult) WithImage(image ImageSpec) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Image = image
	return pr
}

func (pr ProcessResult) WithTemplateData(data map[string]interface{}) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.TemplateData = data
	return pr
}

type ButtonClickAction int

const (
	KeepMessage   ButtonClickAction = iota // Do nothing (default)
	DeleteMessage                          // Delete entire message with buttons
	DeleteButtons                          // Delete only the inline buttons
)

// Keep the old type for internal use to maintain compatibility
type ProcessMessageAction ButtonClickAction

const (
	ProcessKeepMessage    ProcessMessageAction = ProcessMessageAction(KeepMessage)
	ProcessDeleteMessage  ProcessMessageAction = ProcessMessageAction(DeleteMessage)
	ProcessDeleteKeyboard ProcessMessageAction = ProcessMessageAction(DeleteButtons)
)

type processAction int

const (
	actionNextStep processAction = iota
	actionGoToStep
	actionRetryStep
	actionCompleteFlow
	actionCancelFlow
)

func NextStep() ProcessResult {
	return ProcessResult{Action: actionNextStep}
}

func GoToStep(stepName string) ProcessResult {
	return ProcessResult{Action: actionGoToStep, TargetStep: stepName}
}

func Retry() ProcessResult {
	return ProcessResult{Action: actionRetryStep}
}

func CompleteFlow() ProcessResult {
	return ProcessResult{Action: actionCompleteFlow}
}

func CancelFlow() ProcessResult {
	return ProcessResult{Action: actionCancelFlow}
}

func isTemplateMessage(message string) (bool, string) {
	const templatePrefix = "template:"
	if len(message) > len(templatePrefix) && message[:len(templatePrefix)] == templatePrefix {
		templateName := message[len(templatePrefix):]
		return true, templateName
	}
	return false, ""
}
