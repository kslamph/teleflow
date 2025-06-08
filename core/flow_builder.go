package teleflow

import (
	"fmt"
	"time"
)

// Package core/flow_builder.go implements the fluent API for defining conversational flows.
//
// This file provides the builder pattern interface for creating flows using the Step-Prompt-Process API.
// The fluent API enables developers to declaratively define multi-step conversations with a zero-learning-curve
// approach that emphasizes readability and maintainability.
//
// Key Builder Components:
//   - FlowBuilder: Main flow construction with lifecycle management
//   - StepBuilder: Individual step configuration with prompt and processing
//   - PromptBuilder: Declarative prompt specification (message, image, keyboard)
//   - Fluent chaining for seamless flow definition
//   - Built-in validation and error checking during build process
//
// Example Flow Definition:
//   flow := NewFlow("registration").
//     OnError(OnErrorRetry()).
//     Step("name").
//       Prompt("What's your name?").
//       WithImage("welcome.jpg").
//       Process(func(ctx *Context, input string, click *ButtonClick) ProcessResult {
//         ctx.Set("name", input)
//         return NextStep()
//       }).
//     Step("confirm").
//       Prompt(func(ctx *Context) string {
//         return fmt.Sprintf("Hello %s! Continue?", ctx.Get("name"))
//       }).
//       WithInlineKeyboard(func(ctx *Context) *InlineKeyboardBuilder {
//         return NewInlineKeyboard().ButtonCallback("Yes", "confirm")
//       }).
//       Process(handleConfirmation).
//     OnComplete(handleCompletion).
//     Build()

// NewFlow creates a new flow builder with the specified name and default configuration.
// This is the main entry point for creating flows using the Step-Prompt-Process API.
// The returned FlowBuilder provides a fluent interface for adding steps, error handling,
// and lifecycle management.
//
// Parameters:
//   - name: Unique identifier for the flow, used for registration and starting
//
// Returns:
//   - *FlowBuilder: New flow builder instance ready for step definition
//
// Example:
//
//	flow := NewFlow("user_onboarding").
//	  Step("welcome").Prompt("Welcome!").Process(handleWelcome).
//	  Build()
func NewFlow(name string) *FlowBuilder {
	return &FlowBuilder{
		name:    name,
		steps:   make(map[string]*StepBuilder),
		order:   make([]string, 0),
		timeout: 30 * time.Minute, // Default timeout of 30 minutes
	}
}

// Step creates and adds a new step to the flow with the specified name.
// Steps are executed in the order they are added to the flow. Each step must have
// a unique name within the flow. The returned StepBuilder allows fluent configuration
// of the step's prompt and processing logic.
//
// Parameters:
//   - name: Unique identifier for the step within this flow
//
// Returns:
//   - *StepBuilder: Builder for configuring the step's prompt and processing
//
// Panics:
//   - If a step with the same name already exists in the flow
//
// Example:
//
//	flow.Step("welcome").Prompt("Hello!").Process(handleInput)
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
// This handler is executed when the flow successfully completes all steps without errors.
// The handler receives the final context with all accumulated flow data and can perform
// cleanup, notifications, or final processing.
//
// Parameters:
//   - handler: Function to execute on flow completion, receives final Context
//
// Returns:
//   - *FlowBuilder: The same builder for method chaining
//
// Example:
//
//	flow.OnComplete(func(ctx *Context) error {
//	  name := ctx.Get("user_name")
//	  return ctx.Reply(fmt.Sprintf("Welcome %s!", name))
//	})
func (fb *FlowBuilder) OnComplete(handler func(*Context) error) *FlowBuilder {
	fb.onComplete = handler
	return fb
}

// OnError sets the error handling configuration for the flow.
// This defines how the flow should behave when prompt rendering errors occur during execution.
// The error configuration applies to all steps in the flow unless overridden at the step level.
//
// Parameters:
//   - config: ErrorConfig specifying the recovery strategy and user notification
//
// Returns:
//   - *FlowBuilder: The same builder for method chaining
//
// Example:
//
//	flow.OnError(OnErrorRetry("Please try again"))
func (fb *FlowBuilder) OnError(config *ErrorConfig) *FlowBuilder {
	fb.onError = config
	return fb
}

// WithTimeout sets the maximum duration for the entire flow execution.
// If the user does not complete the flow within this time, it will be automatically cancelled.
// Set to 0 to disable timeout (flows will run indefinitely until completion or manual cancellation).
//
// Parameters:
//   - duration: Maximum flow execution time, or 0 for no timeout
//
// Returns:
//   - *FlowBuilder: The same builder for method chaining
//
// Example:
//
//	flow.WithTimeout(10 * time.Minute) // 10 minute timeout
func (fb *FlowBuilder) WithTimeout(duration time.Duration) *FlowBuilder {
	fb.timeout = duration
	return fb
}

// OnProcessDeleteMessage configures the flow to delete entire previous messages when processing button clicks.
// This provides a clean user experience where old messages are completely removed, preventing
// scroll-back confusion and maintaining a tidy conversation history.
//
// Returns:
//   - *FlowBuilder: The same builder for method chaining
//
// Example:
//
//	flow.OnProcessDeleteMessage() // Clean up old messages on button clicks
func (fb *FlowBuilder) OnProcessDeleteMessage() *FlowBuilder {
	fb.onProcessAction = ProcessDeleteMessage
	return fb
}

// OnProcessDeleteKeyboard configures the flow to remove only keyboards from previous messages when processing button clicks.
// This keeps message content visible for reference but disables old interactive elements,
// preventing users from clicking outdated buttons while preserving conversation context.
//
// Returns:
//   - *FlowBuilder: The same builder for method chaining
//
// Example:
//
//	flow.OnProcessDeleteKeyboard() // Keep messages but remove old keyboards
func (fb *FlowBuilder) OnProcessDeleteKeyboard() *FlowBuilder {
	fb.onProcessAction = ProcessDeleteKeyboard
	return fb
}

// Build creates the final Flow object from the builder configuration.
// This method validates the flow configuration, converts the builder structure to the internal
// Flow representation, and performs comprehensive validation of all steps and their requirements.
//
// Validation performed:
//   - At least one step must be defined
//   - Each step must have a prompt configuration
//   - Each step must have a process function
//   - All step names must be unique
//
// Returns:
//   - *Flow: Fully configured and validated Flow ready for registration
//   - error: validation error if the flow configuration is invalid
//
// Example:
//
//	flow, err := NewFlow("example").
//	  Step("input").Prompt("Enter name:").Process(handleInput).
//	  Build()
func (fb *FlowBuilder) Build() (*Flow, error) {
	if len(fb.steps) == 0 {
		return nil, fmt.Errorf("flow '%s' must have at least one step", fb.name)
	}

	// Convert to internal Flow structure
	flow := &Flow{
		Name:            fb.name,
		Steps:           make(map[string]*flowStep),
		Order:           fb.order,
		OnError:         fb.onError,         // Set flow-level error handling
		OnProcessAction: fb.onProcessAction, // Set message handling behavior
		Timeout:         fb.timeout,         // Use configured timeout
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

// PromptBuilder provides a fluent interface for building declarative prompt configurations.
// It allows chaining of prompt components (message, image, keyboard, template data) before
// connecting to a ProcessFunc. The builder validates and constructs the final PromptConfig.
type PromptBuilder struct {
	stepBuilder  *StepBuilder  // Parent step builder for fluent chaining
	promptConfig *PromptConfig // Accumulated prompt configuration
}

// Prompt starts building a prompt configuration for the current step with the specified message.
// The message can be a static string, a function that returns a string based on context,
// or a template reference in the format "template:template_name".
//
// Parameters:
//   - message: MessageSpec defining the prompt content (string, func(*Context) string, or template reference)
//
// Returns:
//   - *PromptBuilder: Builder for adding additional prompt components (image, keyboard, template data)
//
// Example:
//
//	step.Prompt("What's your name?")
//	step.Prompt(func(ctx *Context) string { return fmt.Sprintf("Hello %s", ctx.Get("name")) })
//	step.Prompt("template:welcome_message")
func (sb *StepBuilder) Prompt(message MessageSpec) *PromptBuilder {
	promptConfig := &PromptConfig{
		Message: message,
	}

	return &PromptBuilder{
		stepBuilder:  sb,
		promptConfig: promptConfig,
	}
}

// WithTemplateData sets template variables for the prompt that take precedence over context data.
// This data is available to template rendering and can be used to provide step-specific variables
// that override or supplement the user's flow data.
//
// Parameters:
//   - data: Map of template variables to make available during rendering
//
// Returns:
//   - *PromptBuilder: The same builder for method chaining
//
// Example:
//
//	step.Prompt("template:greeting").WithTemplateData(map[string]interface{}{
//	  "title": "Welcome",
//	  "version": "2.0",
//	})
func (pb *PromptBuilder) WithTemplateData(data map[string]interface{}) *PromptBuilder {
	pb.promptConfig.TemplateData = data
	return pb
}

// WithImage adds an image to the prompt using ImageSpec.
// The image can be a static URL, file path, base64 string, or a function that dynamically
// generates image content based on the current context.
//
// Parameters:
//   - image: ImageSpec defining the image source (string, func(*Context) string, or nil)
//
// Returns:
//   - *PromptBuilder: The same builder for method chaining
//
// Example:
//
//	step.Prompt("Welcome!").WithImage("https://example.com/welcome.jpg")
//	step.Prompt("Hello").WithImage(func(ctx *Context) string {
//	  return fmt.Sprintf("https://api.avatar.com/%s", ctx.Get("username"))
//	})
func (pb *PromptBuilder) WithImage(image ImageSpec) *PromptBuilder {
	pb.promptConfig.Image = image
	return pb
}

// WithInlineKeyboard adds an interactive inline keyboard to the prompt.
// The keyboard function receives the current context and returns an InlineKeyboardBuilder
// for creating buttons with callbacks, URLs, and other interactive elements.
//
// Parameters:
//   - keyboard: KeyboardFunc that generates the keyboard based on current context
//
// Returns:
//   - *PromptBuilder: The same builder for method chaining
//
// Example:
//
//	step.Prompt("Choose option:").WithInlineKeyboard(func(ctx *Context) *InlineKeyboardBuilder {
//	  return NewInlineKeyboard().
//	    ButtonCallback("Option A", "choice_a").
//	    ButtonCallback("Option B", "choice_b")
//	})
func (pb *PromptBuilder) WithInlineKeyboard(keyboard KeyboardFunc) *PromptBuilder {
	pb.promptConfig.Keyboard = keyboard
	return pb
}

// Process completes the prompt configuration and sets the processing function for user input.
// This method finalizes the step definition by connecting the declarative prompt with the
// input processing logic. The ProcessFunc handles all user input and determines flow navigation.
//
// Parameters:
//   - processFunc: ProcessFunc that processes user input and returns flow navigation result
//
// Returns:
//   - *StepBuilder: Parent step builder for continued flow configuration
//
// Example:
//
//	step.Prompt("Enter name:").Process(func(ctx *Context, input string, click *ButtonClick) ProcessResult {
//	  if input == "" {
//	    return Retry().WithPrompt("Name cannot be empty")
//	  }
//	  ctx.Set("name", input)
//	  return NextStep()
//	})
func (pb *PromptBuilder) Process(processFunc ProcessFunc) *StepBuilder {
	pb.stepBuilder.promptConfig = pb.promptConfig
	pb.stepBuilder.processFunc = processFunc
	return pb.stepBuilder
}

// OnComplete sets a completion handler for this specific step.
// This handler is executed after the step's ProcessFunc completes successfully and before
// proceeding to the next step. It can be used for step-specific cleanup, validation, or side effects.
//
// Parameters:
//   - handler: Function to execute after successful step completion
//
// Returns:
//   - *StepBuilder: The same builder for method chaining
//
// Example:
//
//	step.OnComplete(func(ctx *Context) error {
//	  log.Printf("User %d completed step: %s", ctx.UserID(), stepName)
//	  return nil
//	})
func (sb *StepBuilder) OnComplete(handler func(*Context) error) *StepBuilder {
	sb.onComplete = handler
	return sb
}

// Step continues building the flow by adding a new step with the specified name.
// This enables fluent chaining where you can define multiple steps in sequence.
// The method delegates to the parent FlowBuilder to maintain the step order.
//
// Parameters:
//   - name: Unique identifier for the new step
//
// Returns:
//   - *StepBuilder: New step builder for configuring the next step
//
// Example:
//
//	flow.Step("welcome").Prompt("Hello!").Process(handleWelcome).
//	     Step("details").Prompt("Enter details:").Process(handleDetails)
func (sb *StepBuilder) Step(name string) *StepBuilder {
	return sb.flowBuilder.Step(name)
}

// OnFlowComplete sets the completion handler for the entire flow.
// This is a convenience method that allows setting the flow completion handler
// from within a step builder chain, maintaining the fluent interface.
//
// Parameters:
//   - handler: Function to execute when the entire flow completes successfully
//
// Returns:
//   - *FlowBuilder: Parent flow builder for continued flow configuration
//
// Example:
//
//	flow.Step("final").Prompt("Done!").Process(handleFinal).
//	     OnFlowComplete(func(ctx *Context) error {
//	       return ctx.Reply("Flow completed successfully!")
//	     })
func (sb *StepBuilder) OnFlowComplete(handler func(*Context) error) *FlowBuilder {
	return sb.flowBuilder.OnComplete(handler)
}

// Build creates the final Flow object from the current builder configuration.
// This is a convenience method that delegates to the parent FlowBuilder, allowing
// flow finalization from within a step builder chain.
//
// Returns:
//   - *Flow: Fully configured and validated Flow ready for registration
//   - error: validation error if the flow configuration is invalid
//
// Example:
//
//	flow, err := NewFlow("example").
//	  Step("input").Prompt("Enter name:").Process(handleInput).
//	  Build()
func (sb *StepBuilder) Build() (*Flow, error) {
	return sb.flowBuilder.Build()
}
