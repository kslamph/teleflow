package teleflow

// Assuming Context type is defined in core/context.go

// FlowBuilder is used to construct a new flow.
type FlowBuilder struct {
	name        string
	steps       map[string]*StepBuilder
	order       []string
	onComplete  func(*Context) error
	onError     *ErrorConfig // Flow-level error handling
	currentStep *StepBuilder // Helper for fluent API
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
	Message  MessageSpec  // Can be a string, func(*Context) string, or a template.
	Image    ImageSpec    // Can be a string (URL, file path, base64), func(*Context) string, or nil.
	Keyboard KeyboardFunc // Can be a func(*Context) map[string]interface{} or nil.
}

// MessageSpec defines the type for message content.
// It can be a string or a function that takes a context and returns a string.
type MessageSpec interface{}

// ImageSpec defines the type for image content.
// It can be a string (URL, file path, base64), a function that returns such a string, or nil.
type ImageSpec interface{}

// KeyboardFunc defines the function signature for generating a keyboard.
// It takes a Context and returns a map representing the keyboard structure.
// Example: map[string]interface{}{"Button Text": "callback_data"}
type KeyboardFunc func(ctx *Context) map[string]interface{}

// ProcessFunc defines the signature for the function that processes user input for a step.
// It takes the current context, user input string, and optional button click information.
// It returns a ProcessResult indicating the next action for the flow.
type ProcessFunc func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult

// ButtonClick holds information about a button click event passed to ProcessFunc.
type ButtonClick struct {
	Data     string                 // Callback data associated with the button.
	Text     string                 // Text of the button clicked.
	UserID   int64                  // ID of the user who clicked the button.
	ChatID   int64                  // ID of the chat where the button was clicked.
	Metadata map[string]interface{} // Optional metadata.
}

// ProcessResult specifies the outcome of a ProcessFunc execution.
// It dictates the next action in the flow and can optionally provide a new prompt.
type ProcessResult struct {
	Action     ProcessAction
	TargetStep string        // Name of the step to go to, if Action is ActionGoToStep.
	Prompt     *PromptConfig // Optional prompt to display before taking the action or on retry.
}

// WithPrompt adds a custom prompt to any ProcessResult.
// This allows for fluent chaining like: NextStep().WithPrompt(&PromptConfig{Message: "Moving to next step!"})
func (pr ProcessResult) WithPrompt(prompt *PromptConfig) ProcessResult {
	pr.Prompt = prompt
	return pr
}

// ProcessAction defines the type of action to be taken after a step's ProcessFunc completes.
type ProcessAction int

// Defines the possible actions that can be taken by a ProcessFunc.
const (
	ActionNextStep     ProcessAction = iota // Proceed to the next step in sequence.
	ActionGoToStep                          // Jump to a specific step by name.
	ActionRetry                             // Retry the current step, optionally with a new prompt.
	ActionCompleteFlow                      // Mark the entire flow as completed.
	ActionCancelFlow                        // Cancel the current flow.
)

// Helper functions for creating ProcessResult instances.

// NextStep creates a ProcessResult that directs the flow to the next sequential step.
func NextStep() ProcessResult {
	return ProcessResult{Action: ActionNextStep}
}

// GoToStep creates a ProcessResult that directs the flow to a specific step by name.
func GoToStep(stepName string) ProcessResult {
	return ProcessResult{Action: ActionGoToStep, TargetStep: stepName}
}

// RetryWithPrompt creates a ProcessResult that retries the current step,
// optionally displaying a new prompt. If prompt is nil, the original step prompt may be reshown.
func RetryWithPrompt(prompt *PromptConfig) ProcessResult {
	return ProcessResult{Action: ActionRetry, Prompt: prompt}
}

// Retry creates a ProcessResult that retries the current step (re-renders original prompt).
func Retry() ProcessResult {
	return ProcessResult{Action: ActionRetry}
}

// CompleteFlow creates a ProcessResult that marks the current flow as completed.
func CompleteFlow() ProcessResult {
	return ProcessResult{Action: ActionCompleteFlow}
}

// CancelFlow creates a ProcessResult that cancels the current flow.
func CancelFlow() ProcessResult {
	return ProcessResult{Action: ActionCancelFlow}
}
