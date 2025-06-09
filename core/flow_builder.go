package teleflow

import (
	"fmt"
	"time"
)

// NewFlow creates a new FlowBuilder for constructing a conversation flow.
// The name parameter identifies the flow and is used when starting it via ctx.StartFlow().
// Flows have a default timeout of 30 minutes.
//
// Example:
//
//	flow := teleflow.NewFlow("registration").
//		Step("ask_name").Prompt("What's your name?").Process(...).
//		Step("ask_email").Prompt("What's your email?").Process(...).
//		OnComplete(func(ctx *teleflow.Context) error {
//			return ctx.SendPromptText("Registration completed!")
//		}).
//		Build()
func NewFlow(name string) *FlowBuilder {
	return &FlowBuilder{
		name:    name,
		steps:   make(map[string]*StepBuilder),
		order:   make([]string, 0),
		timeout: 30 * time.Minute,
	}
}

// Step adds a new step to the flow with the specified name.
// Steps are executed in the order they are defined unless explicitly redirected.
// Step names must be unique within a flow - duplicate names will cause a panic.
// Returns a StepBuilder for configuring the step's prompt and processing logic.
//
// Example:
//
//	flow.Step("collect_info").
//		Prompt("Please provide your information:").
//		Process(func(ctx *teleflow.Context, input string, click *teleflow.ButtonClick) teleflow.ProcessResult {
//			// Process the user's input
//			return teleflow.NextStep()
//		})
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

// OnComplete sets a callback function that is executed when the flow completes successfully.
// This is called after the last step processes successfully or when CompleteFlow() is returned.
// The completion handler can access flow data and send final messages to the user.
//
// Example:
//
//	flow.OnComplete(func(ctx *teleflow.Context) error {
//		// Save data, send confirmation, etc.
//		name, _ := ctx.GetFlowData("name")
//		return ctx.SendPromptText("Thank you, " + name.(string) + "!")
//	})
func (fb *FlowBuilder) OnComplete(handler func(*Context) error) *FlowBuilder {
	fb.onComplete = handler
	return fb
}

// OnError configures how the flow handles errors during step processing.
// This applies to all steps in the flow unless overridden at the step level.
// The error configuration determines whether to cancel, retry, or ignore errors.
//
// Example:
//
//	flow.OnError(teleflow.OnErrorRetry("Something went wrong. Please try again."))
func (fb *FlowBuilder) OnError(config *ErrorConfig) *FlowBuilder {
	fb.onError = config
	return fb
}

// WithTimeout sets a timeout duration for the entire flow.
// If the flow is not completed within this time, it will be automatically cancelled.
// The default timeout is 30 minutes.
//
// Example:
//
//	flow.WithTimeout(30 * time.Minute)
func (fb *FlowBuilder) WithTimeout(duration time.Duration) *FlowBuilder {
	fb.timeout = duration
	return fb
}

// OnButtonClick configures the default action to take when inline keyboard buttons are clicked.
// This can be overridden at the step level if needed. Options include keeping the message,
// deleting the entire message, or just removing the keyboard buttons.
//
// Example:
//
//	flow.OnButtonClick(teleflow.DeleteMessage) // Delete messages after button clicks
func (fb *FlowBuilder) OnButtonClick(action ButtonClickAction) *FlowBuilder {
	fb.onProcessAction = ProcessMessageAction(action)
	return fb
}

// Build constructs and validates the final Flow from the FlowBuilder configuration.
// Returns an error if the flow is invalid (e.g., no steps, steps missing prompts or processing).
// Once built, the Flow can be registered with a bot using bot.RegisterFlow().
//
// Example:
//
//	flow, err := teleflow.NewFlow("example").
//		Step("step1").Prompt("Hello").Process(...).
//		Build()
//	if err != nil {
//		log.Fatal(err)
//	}
//	bot.RegisterFlow(flow)
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

// PromptBuilder provides a fluent interface for configuring step prompts.
// It allows setting message content, images, keyboards, and template data
// before defining the processing logic for user responses.
type PromptBuilder struct {
	stepBuilder  *StepBuilder  // Reference to the step being built
	promptConfig *PromptConfig // Configuration for the prompt
}

// Prompt sets the message content for a flow step.
// The message can be a string, template reference, or function that returns content.
// Returns a PromptBuilder for further configuration of the prompt.
//
// Example:
//
//	step.Prompt("What's your name?")
//	step.Prompt("template:greeting")  // Use a template
//	step.Prompt(func(ctx *teleflow.Context) string {
//		return "Hello " + getUserName(ctx)
//	})
func (sb *StepBuilder) Prompt(message MessageSpec) *PromptBuilder {
	promptConfig := &PromptConfig{
		Message: message,
	}

	return &PromptBuilder{
		stepBuilder:  sb,
		promptConfig: promptConfig,
	}
}

// WithTemplateData adds data for template rendering to the prompt.
// This data is available to templates when rendering the prompt message.
//
// Example:
//
//	step.Prompt("template:greeting").
//		WithTemplateData(map[string]interface{}{
//			"name": "John",
//			"time": time.Now(),
//		})
func (pb *PromptBuilder) WithTemplateData(data map[string]interface{}) *PromptBuilder {
	pb.promptConfig.TemplateData = data
	return pb
}

// WithImage adds an image to the prompt.
// The image can be a URL, file path, byte slice, or function that returns any of these.
//
// Example:
//
//	step.Prompt("Choose an option:").
//		WithImage("https://example.com/menu.jpg")
func (pb *PromptBuilder) WithImage(image ImageSpec) *PromptBuilder {
	pb.promptConfig.Image = image
	return pb
}

// WithPromptKeyboard adds an inline keyboard to the prompt.
// The keyboard function receives the context and returns a keyboard builder.
//
// Example:
//
//	step.Prompt("Choose an option:").
//		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
//			return teleflow.NewPromptKeyboard().
//				ButtonCallback("Option 1", "opt1").
//				ButtonCallback("Option 2", "opt2")
//		})
func (pb *PromptBuilder) WithPromptKeyboard(keyboard KeyboardFunc) *PromptBuilder {
	pb.promptConfig.Keyboard = keyboard
	return pb
}

// Process sets the processing function for handling user responses to the prompt.
// This function receives user input and button clicks, returning a ProcessResult
// that determines the next action in the flow.
//
// Example:
//
//	step.Prompt("Enter your age:").
//		Process(func(ctx *teleflow.Context, input string, click *teleflow.ButtonClick) teleflow.ProcessResult {
//			age, err := strconv.Atoi(input)
//			if err != nil || age < 0 {
//				return teleflow.Retry().WithPrompt("Please enter a valid age")
//			}
//			ctx.SetFlowData("age", age)
//			return teleflow.NextStep()
//		})
func (pb *PromptBuilder) Process(processFunc ProcessFunc) *StepBuilder {
	pb.stepBuilder.promptConfig = pb.promptConfig
	pb.stepBuilder.processFunc = processFunc
	return pb.stepBuilder
}

// Step allows adding another step to the flow from within a StepBuilder.
// This provides a convenient way to chain step definitions.
func (sb *StepBuilder) Step(name string) *StepBuilder {
	return sb.flowBuilder.Step(name)
}

// OnComplete allows setting the completion handler from within a StepBuilder.
// This provides a convenient way to set the completion handler after defining steps.
func (sb *StepBuilder) OnComplete(handler func(*Context) error) *FlowBuilder {
	return sb.flowBuilder.OnComplete(handler)
}

// Build constructs the final Flow from within a StepBuilder.
// This provides a convenient way to build the flow after defining the last step.
func (sb *StepBuilder) Build() (*Flow, error) {
	return sb.flowBuilder.Build()
}
