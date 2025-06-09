package teleflow

import "time"

// FlowBuilder provides a fluent interface for constructing conversation flows.
// It allows defining multi-step conversations with branching logic, error handling,
// and completion callbacks. Use NewFlow() to create a new FlowBuilder instance.
type FlowBuilder struct {
	name            string                  // Flow name for identification
	steps           map[string]*StepBuilder // Map of step name to step builder
	order           []string                // Order of steps as they were added
	onComplete      func(*Context) error    // Callback when flow completes successfully
	onError         *ErrorConfig            // Error handling configuration
	onProcessAction ProcessMessageAction    // Default action for processing messages
	currentStep     *StepBuilder            // Currently being built step
	timeout         time.Duration           // Flow timeout duration
}

// StepBuilder represents a single step in a conversation flow.
// Each step consists of a prompt to show the user and a processing function
// to handle the user's response.
type StepBuilder struct {
	name         string        // Step name for identification and navigation
	promptConfig *PromptConfig // Configuration for the prompt to display
	processFunc  ProcessFunc   // Function to process user input
	flowBuilder  *FlowBuilder  // Reference to parent flow builder
}

// PromptConfig defines the configuration for a prompt message in a flow step.
// It can include text messages, images, keyboards, and template data for dynamic content.
type PromptConfig struct {
	Message      MessageSpec            // Message content (string, function, or template)
	Image        ImageSpec              // Optional image (URL, file path, or bytes)
	Keyboard     KeyboardFunc           // Optional keyboard generator function
	TemplateData map[string]interface{} // Data for template rendering
}

// MessageSpec represents various ways to specify message content.
// Can be a string, a function that returns a string, or template reference.
type MessageSpec interface{}

// ImageSpec represents various ways to specify image content.
// Can be a URL string, file path, byte slice, or function returning any of these.
type ImageSpec interface{}

// KeyboardFunc is a function that generates an inline keyboard for a prompt.
// It receives the current context and returns a keyboard builder.
type KeyboardFunc func(ctx *Context) *PromptKeyboardBuilder

// ProcessFunc processes user input for a flow step and determines the next action.
// It receives the context, user text input, and any button click data,
// returning a ProcessResult that controls flow progression.
type ProcessFunc func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult

// ButtonClick contains information about an inline keyboard button that was pressed.
// It provides access to the button's data, text, and metadata for processing decisions.
type ButtonClick struct {
	Data     interface{}            // Custom data associated with the button
	Text     string                 // Display text of the button
	UserID   int64                  // ID of the user who clicked the button
	ChatID   int64                  // ID of the chat where the button was clicked
	Metadata map[string]interface{} // Additional metadata for the button
}

// ProcessResult defines the outcome of processing user input in a flow step.
// It specifies what action to take next (continue, retry, jump to step, etc.)
// and can include an optional prompt to display.
type ProcessResult struct {
	Action     processAction // What action to take (next step, retry, etc.)
	TargetStep string        // Target step name for jump actions
	Prompt     *PromptConfig // Optional prompt to display before action
}

// WithPrompt adds a prompt message to a ProcessResult.
// This allows displaying a message before executing the result action.
//
// Example:
//
//	return teleflow.NextStep().WithPrompt("Moving to next step...")
func (pr ProcessResult) WithPrompt(prompt MessageSpec) ProcessResult {

	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Message = prompt

	return pr
}

// WithImage adds an image to a ProcessResult's prompt.
// The image will be displayed along with any message before executing the action.
//
// Example:
//
//	return teleflow.CompleteFlow().WithImage("https://example.com/success.jpg")
func (pr ProcessResult) WithImage(image ImageSpec) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Image = image
	return pr
}

// WithTemplateData adds template data to a ProcessResult's prompt.
// This data can be used for rendering template messages.
//
// Example:
//
//	return teleflow.Retry().WithTemplateData(map[string]interface{}{
//		"error": "Invalid input",
//		"attempts": 2,
//	})
func (pr ProcessResult) WithTemplateData(data map[string]interface{}) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.TemplateData = data
	return pr
}

// ButtonClickAction defines what happens to a message when its inline keyboard button is clicked.
type ButtonClickAction int

const (
	KeepMessage   ButtonClickAction = iota // Do nothing (default)
	DeleteMessage                          // Delete entire message with buttons
	DeleteButtons                          // Delete only the inline buttons
)

// ProcessMessageAction is an internal type for backward compatibility.
// Use ButtonClickAction for new code.
type ProcessMessageAction ButtonClickAction

const (
	ProcessKeepMessage    ProcessMessageAction = ProcessMessageAction(KeepMessage)
	ProcessDeleteMessage  ProcessMessageAction = ProcessMessageAction(DeleteMessage)
	ProcessDeleteKeyboard ProcessMessageAction = ProcessMessageAction(DeleteButtons)
)

// processAction defines internal actions that can be taken after processing user input.
type processAction int

const (
	actionNextStep processAction = iota
	actionGoToStep
	actionRetryStep
	actionCompleteFlow
	actionCancelFlow
)

// NextStep creates a ProcessResult that advances to the next step in the flow.
// This is the most common action for continuing a flow progression.
//
// Example:
//
//	func processName(ctx *teleflow.Context, input string, click *teleflow.ButtonClick) teleflow.ProcessResult {
//		ctx.SetFlowData("name", input)
//		return teleflow.NextStep()
//	}
func NextStep() ProcessResult {
	return ProcessResult{Action: actionNextStep}
}

// GoToStep creates a ProcessResult that jumps to a specific step in the flow.
// This enables conditional flow branching based on user input or business logic.
//
// Example:
//
//	func processChoice(ctx *teleflow.Context, input string, click *teleflow.ButtonClick) teleflow.ProcessResult {
//		if input == "advanced" {
//			return teleflow.GoToStep("advanced_options")
//		}
//		return teleflow.NextStep()
//	}
func GoToStep(stepName string) ProcessResult {
	return ProcessResult{Action: actionGoToStep, TargetStep: stepName}
}

// Retry creates a ProcessResult that repeats the current step.
// This is useful for handling invalid input or validation failures.
//
// Example:
//
//	func processAge(ctx *teleflow.Context, input string, click *teleflow.ButtonClick) teleflow.ProcessResult {
//		age, err := strconv.Atoi(input)
//		if err != nil || age < 0 {
//			return teleflow.Retry().WithPrompt("Please enter a valid age:")
//		}
//		return teleflow.NextStep()
//	}
func Retry() ProcessResult {
	return ProcessResult{Action: actionRetryStep}
}

// CompleteFlow creates a ProcessResult that successfully completes the flow.
// This triggers any completion handlers registered with the flow.
//
// Example:
//
//	func processConfirmation(ctx *teleflow.Context, input string, click *teleflow.ButtonClick) teleflow.ProcessResult {
//		if input == "confirm" {
//			return teleflow.CompleteFlow().WithPrompt("Registration completed!")
//		}
//		return teleflow.CancelFlow()
//	}
func CompleteFlow() ProcessResult {
	return ProcessResult{Action: actionCompleteFlow}
}

// CancelFlow creates a ProcessResult that cancels the flow.
// This immediately terminates the flow without calling completion handlers.
//
// Example:
//
//	func processCancel(ctx *teleflow.Context, input string, click *teleflow.ButtonClick) teleflow.ProcessResult {
//		return teleflow.CancelFlow().WithPrompt("Flow cancelled.")
//	}
func CancelFlow() ProcessResult {
	return ProcessResult{Action: actionCancelFlow}
}

// isTemplateMessage checks if a message string is a template reference.
// Template references are prefixed with "template:" followed by the template name.
// Returns true and the template name if it's a template, false otherwise.
func isTemplateMessage(message string) (bool, string) {
	const templatePrefix = "template:"
	if len(message) > len(templatePrefix) && message[:len(templatePrefix)] == templatePrefix {
		templateName := message[len(templatePrefix):]
		return true, templateName
	}
	return false, ""
}
