package teleflow

import (
	"fmt"
	"time"
)

// NewFlow creates a new flow builder with the given name.
// This is the main entry point for creating flows using the new Step-Prompt-Process API.
func NewFlow(name string) *FlowBuilder {
	return &FlowBuilder{
		name:  name,
		steps: make(map[string]*StepBuilder),
		order: make([]string, 0),
	}
}

// Step creates and adds a new step to the flow.
// Steps are executed in the order they are added.
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

// OnComplete sets the completion handler for the entire flow.
// This handler is called when the flow successfully completes all steps.
func (fb *FlowBuilder) OnComplete(handler func(*Context) error) *FlowBuilder {
	fb.onComplete = handler
	return fb
}

// OnError sets the error handling configuration for the flow.
// This defines how the flow should behave when rendering errors occur.
func (fb *FlowBuilder) OnError(config *ErrorConfig) *FlowBuilder {
	fb.onError = config
	return fb
}

// Build creates the final Flow object from the builder.
// This converts the new API structure to the internal Flow representation.
func (fb *FlowBuilder) Build() (*Flow, error) {
	if len(fb.steps) == 0 {
		return nil, fmt.Errorf("flow '%s' must have at least one step", fb.name)
	}

	// Convert to internal Flow structure
	flow := &Flow{
		Name:    fb.name,
		Steps:   make(map[string]*flowStep),
		Order:   fb.order,
		OnError: fb.onError,       // Set flow-level error handling
		Timeout: 30 * time.Minute, // Default timeout
	}

	// Convert each step
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
			OnComplete:   stepBuilder.onComplete,
			Timeout:      5 * time.Minute, // Default step timeout
		}

		flow.Steps[stepName] = flowStep
	}

	// Set flow completion handler
	if fb.onComplete != nil {
		flow.OnComplete = fb.onComplete
	}

	return flow, nil
}

// PromptBuilder provides a fluent interface for building prompts
type PromptBuilder struct {
	stepBuilder  *StepBuilder
	promptConfig *PromptConfig
}

// Prompt starts building a prompt with a message.
// The message can be a string or a template reference like "template:my_template".
func (sb *StepBuilder) Prompt(message MessageSpec) *PromptBuilder {
	promptConfig := &PromptConfig{
		Message: message,
	}

	return &PromptBuilder{
		stepBuilder:  sb,
		promptConfig: promptConfig,
	}
}

// WithTemplateData sets the template data for the prompt.
// This data is available to templates and takes precedence over context data.
func (pb *PromptBuilder) WithTemplateData(data map[string]interface{}) *PromptBuilder {
	pb.promptConfig.TemplateData = data
	return pb
}

// WithImage sets the image for the prompt.
// The image can be a URL, file path, base64 string, or a function that returns such a string.
func (pb *PromptBuilder) WithImage(image ImageSpec) *PromptBuilder {
	pb.promptConfig.Image = image
	return pb
}

// WithInlineKeyboard sets the inline keyboard for the prompt.
func (pb *PromptBuilder) WithInlineKeyboard(keyboard KeyboardFunc) *PromptBuilder {
	pb.promptConfig.Keyboard = keyboard
	return pb
}

// Process completes the prompt configuration and sets the processing function.
func (pb *PromptBuilder) Process(processFunc ProcessFunc) *StepBuilder {
	pb.stepBuilder.promptConfig = pb.promptConfig
	pb.stepBuilder.processFunc = processFunc
	return pb.stepBuilder
}

// OnComplete sets a completion handler for this specific step.
// This is called after the step's ProcessFunc completes successfully.
func (sb *StepBuilder) OnComplete(handler func(*Context) error) *StepBuilder {
	sb.onComplete = handler
	return sb
}

// Step continues building the flow by adding a new step.
// This allows for fluent chaining: Step("a").Prompt(...).Process(...).Step("b")...
func (sb *StepBuilder) Step(name string) *StepBuilder {
	return sb.flowBuilder.Step(name)
}

// OnFlowComplete sets the completion handler for the entire flow.
// This is a convenience method that delegates to the FlowBuilder.
func (sb *StepBuilder) OnFlowComplete(handler func(*Context) error) *FlowBuilder {
	return sb.flowBuilder.OnComplete(handler)
}

// Build creates the final Flow object.
// This is a convenience method that delegates to the FlowBuilder.
func (sb *StepBuilder) Build() (*Flow, error) {
	return sb.flowBuilder.Build()
}
