package teleflow

import (
	"fmt"
	"time"
)

func NewFlow(name string) *FlowBuilder {
	return &FlowBuilder{
		name:    name,
		steps:   make(map[string]*StepBuilder),
		order:   make([]string, 0),
		timeout: 30 * time.Minute,
	}
}

func (fb *FlowBuilder) Step(name string) *StepBuilder {
	if _, exists := fb.steps[name]; exists {
		panic(fmt.Sprintf("Step '%s' already exists in flow '%s'", name, fb.name))
	}

	stepBuilder := &StepBuilder{
		name:        name,
		flowBuilder: fb,
	}

	fb.steps[name] = stepBuilder
	fb.order = append(fb.order, name)
	fb.currentStep = stepBuilder

	return stepBuilder
}

func (fb *FlowBuilder) OnComplete(handler func(*Context) error) *FlowBuilder {
	fb.onComplete = handler
	return fb
}

func (fb *FlowBuilder) OnError(config *ErrorConfig) *FlowBuilder {
	fb.onError = config
	return fb
}

func (fb *FlowBuilder) WithTimeout(duration time.Duration) *FlowBuilder {
	fb.timeout = duration
	return fb
}

func (fb *FlowBuilder) OnButtonClick(action ButtonClickAction) *FlowBuilder {
	fb.onProcessAction = ProcessMessageAction(action)
	return fb
}

func (fb *FlowBuilder) Build() (*Flow, error) {
	if len(fb.steps) == 0 {
		return nil, fmt.Errorf("flow '%s' must have at least one step", fb.name)
	}

	flow := &Flow{
		Name:            fb.name,
		Steps:           make(map[string]*flowStep),
		Order:           fb.order,
		OnError:         fb.onError,
		OnProcessAction: fb.onProcessAction,
		Timeout:         fb.timeout,
	}

	for _, stepName := range fb.order {
		stepBuilder := fb.steps[stepName]

		if stepBuilder.promptConfig == nil {
			return nil, fmt.Errorf("step '%s' must have a prompt configuration", stepName)
		}

		if stepBuilder.processFunc == nil {
			return nil, fmt.Errorf("step '%s' must have a process function", stepName)
		}

		flowStep := &flowStep{
			Name:         stepBuilder.name,
			PromptConfig: stepBuilder.promptConfig,
			ProcessFunc:  stepBuilder.processFunc,
		}

		flow.Steps[stepName] = flowStep
	}

	if fb.onComplete != nil {
		flow.OnComplete = fb.onComplete
	}

	return flow, nil
}

type PromptBuilder struct {
	stepBuilder  *StepBuilder
	promptConfig *PromptConfig
}

func (sb *StepBuilder) Prompt(message MessageSpec) *PromptBuilder {
	promptConfig := &PromptConfig{
		Message: message,
	}

	return &PromptBuilder{
		stepBuilder:  sb,
		promptConfig: promptConfig,
	}
}

func (pb *PromptBuilder) WithTemplateData(data map[string]interface{}) *PromptBuilder {
	pb.promptConfig.TemplateData = data
	return pb
}

func (pb *PromptBuilder) WithImage(image ImageSpec) *PromptBuilder {
	pb.promptConfig.Image = image
	return pb
}

func (pb *PromptBuilder) WithPromptKeyboard(keyboard KeyboardFunc) *PromptBuilder {
	pb.promptConfig.Keyboard = keyboard
	return pb
}

func (pb *PromptBuilder) Process(processFunc ProcessFunc) *StepBuilder {
	pb.stepBuilder.promptConfig = pb.promptConfig
	pb.stepBuilder.processFunc = processFunc
	return pb.stepBuilder
}

func (sb *StepBuilder) Step(name string) *StepBuilder {
	return sb.flowBuilder.Step(name)
}

func (sb *StepBuilder) OnComplete(handler func(*Context) error) *FlowBuilder {
	return sb.flowBuilder.OnComplete(handler)
}

func (sb *StepBuilder) Build() (*Flow, error) {
	return sb.flowBuilder.Build()
}
