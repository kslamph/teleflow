package teleflow

import "time"

// Assuming Context type is defined in core/context.go

// FlowBuilder is used to construct a new flow.
type FlowBuilder struct {
	name            string
	steps           map[string]*StepBuilder
	order           []string
	onComplete      func(*Context) error
	onError         *ErrorConfig         // Flow-level error handling
	onProcessAction ProcessMessageAction // How to handle previous messages on button clicks
	currentStep     *StepBuilder         // Helper for fluent API
	timeout         time.Duration        // Flow timeout duration
}

// StepBuilder is used to construct a step within a flow.
type StepBuilder struct {
	name         string
	promptConfig *PromptConfig
	processFunc  ProcessFunc
	onComplete   func(*Context) error // Step-specific completion handler
	flowBuilder  *FlowBuilder         // Parent builder
}

// PromptConfig represents the declarative specification for a prompt.
type PromptConfig struct {
	Message      MessageSpec            // Can be a string, func(*Context) string, or a template.
	Image        ImageSpec              // Can be a string (URL, file path, base64), func(*Context) string, or nil.
	Keyboard     KeyboardFunc           // Can be a func(*Context) map[string]interface{} or nil.
	TemplateData map[string]interface{} // Template variables that take precedence over context data
}

// MessageSpec defines the type for message content.
// It can be a string or a function that takes a context and returns a string.
type MessageSpec interface{}

// ImageSpec defines the type for image content.
// It can be a string (URL, file path, base64), a function that returns such a string, or nil.
type ImageSpec interface{}

// KeyboardFunc defines the function signature for generating an inline keyboard.
// It takes a Context and returns an InlineKeyboardBuilder for fluent keyboard construction.
type KeyboardFunc func(ctx *Context) *InlineKeyboardBuilder

// ProcessFunc defines the signature for the function that processes user input for a step.
// It takes the current context, user input string, and optional button click information.
// It returns a ProcessResult indicating the next action for the flow.
type ProcessFunc func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult

// ButtonClick holds information about a button click event passed to ProcessFunc.
// With the new InlineKeyboardBuilder system, Data contains the original interface{}
// that was passed to ButtonCallback(), not just the UUID string.
type ButtonClick struct {
	Data     interface{}            // Original callback data passed to ButtonCallback() (UUID-mapped).
	Text     string                 // Text of the button clicked.
	UserID   int64                  // ID of the user who clicked the button.
	ChatID   int64                  // ID of the chat where the button was clicked.
	Metadata map[string]interface{} // Optional metadata.
}

// ProcessResult specifies the outcome of a ProcessFunc execution.
// It dictates the next action in the flow and can optionally provide a new prompt.
type ProcessResult struct {
	Action     processAction
	TargetStep string        // Name of the step to go to, if Action is ActionGoToStep.
	Prompt     *PromptConfig // Optional prompt to display before taking the action or on retry.
}

// WithPrompt adds a custom prompt to any ProcessResult.
// It accepts either a *PromptConfig or a MessageSpec (same as StepBuilder.Prompt).
// This allows for fluent chaining like:
// NextStep().WithPrompt(&PromptConfig{Message: "Moving to next step!"})
// or NextStep().WithPrompt("Moving to next step!")
// or NextStep().WithPrompt(func(ctx *Context) string { return "Dynamic message" })
func (pr ProcessResult) WithPrompt(prompt MessageSpec) ProcessResult {

	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Message = prompt

	return pr
}

// WithImage adds an image to the ProcessResult's prompt using ImageSpec (same as PromptBuilder.WithImage).
// This allows for fluent chaining like:
// NextStep().WithPrompt("Message").WithImage("https://example.com/image.jpg")
// or NextStep().WithPrompt("Message").WithImage(func(ctx *Context) string { return "dynamic-url" })
func (pr ProcessResult) WithImage(image ImageSpec) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Image = image
	return pr
}

// WithTemplateData adds template data to the ProcessResult's prompt (same as PromptBuilder.WithTemplateData).
// This allows for fluent chaining like:
// NextStep().WithPrompt("template:my_template").WithTemplateData(map[string]interface{}{"name": "John"})
func (pr ProcessResult) WithTemplateData(data map[string]interface{}) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.TemplateData = data
	return pr
}

// ProcessMessageAction defines how to handle previous messages when processing button clicks
type ProcessMessageAction int

const (
	ProcessKeepMessage    ProcessMessageAction = iota // Keep previous messages untouched (default)
	ProcessDeleteMessage                              // Delete entire previous message
	ProcessDeleteKeyboard                             // Remove only keyboard from previous message
)

// processAction defines the type of action to be taken after a step's ProcessFunc completes.
type processAction int

// Defines the possible actions that can be taken by a ProcessFunc.
const (
	actionNextStep     processAction = iota // Proceed to the next step in sequence.
	actionGoToStep                          // Jump to a specific step by name.
	actionRetryStep                         // Retry the current step, optionally with a new prompt.
	actionCompleteFlow                      // Mark the entire flow as completed.
	actionCancelFlow                        // Cancel the current flow.
)

// Helper functions for creating ProcessResult instances.

// NextStep creates a ProcessResult that directs the flow to the next sequential step.
func NextStep() ProcessResult {
	return ProcessResult{Action: actionNextStep}
}

// GoToStep creates a ProcessResult that directs the flow to a specific step by name.
func GoToStep(stepName string) ProcessResult {
	return ProcessResult{Action: actionGoToStep, TargetStep: stepName}
}

// Retry creates a ProcessResult that retries the current step (re-renders original prompt).
func Retry() ProcessResult {
	return ProcessResult{Action: actionRetryStep}
}

// CompleteFlow creates a ProcessResult that marks the current flow as completed.
func CompleteFlow() ProcessResult {
	return ProcessResult{Action: actionCompleteFlow}
}

// CancelFlow creates a ProcessResult that cancels the current flow.
func CancelFlow() ProcessResult {
	return ProcessResult{Action: actionCancelFlow}
}

// isTemplateMessage checks if a message string is a template reference.
// It returns true and the template name if the message has the format "template:templateName",
// otherwise returns false and an empty string.
func isTemplateMessage(message string) (bool, string) {
	const templatePrefix = "template:"
	if len(message) > len(templatePrefix) && message[:len(templatePrefix)] == templatePrefix {
		templateName := message[len(templatePrefix):]
		return true, templateName
	}
	return false, ""
}
