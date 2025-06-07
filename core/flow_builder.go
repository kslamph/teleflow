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

// Prompt sets the prompt configuration for the current step.
// This defines what message, image, and keyboard to show to the user.
func (sb *StepBuilder) Prompt(message MessageSpec, image ImageSpec, keyboard KeyboardFunc) *StepBuilder {
	sb.promptConfig = &PromptConfig{
		Message:  message,
		Image:    image,
		Keyboard: keyboard,
	}
	return sb
}

// Process sets the processing function for the current step.
// This function handles user input and determines the next action.
func (sb *StepBuilder) Process(processFunc ProcessFunc) *StepBuilder {
	sb.processFunc = processFunc
	return sb
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
