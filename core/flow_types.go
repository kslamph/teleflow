package teleflow

import "time"

// Package core/flow_types.go defines all type definitions and structures for the Step-Prompt-Process API.
//
// This file contains the complete type system for building and executing conversational flows,
// including builder types, configuration structures, function signatures, and result types.
// All types work together to provide a comprehensive, type-safe API for flow development.
//
// Key Type Categories:
//   - Builder Types: FlowBuilder, StepBuilder for flow construction
//   - Configuration Types: PromptConfig, ErrorConfig for declarative setup
//   - Function Types: ProcessFunc, KeyboardFunc for custom logic
//   - Result Types: ProcessResult, ButtonClick for flow control
//   - Action Types: processAction, ProcessMessageAction for behavior control
//   - Helper Functions: NextStep(), GoToStep(), etc. for flow navigation
//
// Type Relationships:
//   FlowBuilder -> StepBuilder -> PromptConfig -> ProcessFunc -> ProcessResult
//   The types form a pipeline from flow definition to execution and result handling.

// FlowBuilder is used to construct a new flow with fluent API support.
// It accumulates flow configuration including steps, error handling, timeouts,
// and lifecycle callbacks before building the final Flow object.
type FlowBuilder struct {
	name            string                  // Unique flow identifier
	steps           map[string]*StepBuilder // Map of step builders by name
	order           []string                // Ordered list of step names for execution sequence
	onComplete      func(*Context) error    // Flow completion handler
	onError         *ErrorConfig            // Flow-level error handling configuration
	onProcessAction ProcessMessageAction    // Default message handling behavior for button clicks
	currentStep     *StepBuilder            // Currently active step builder for fluent chaining
	timeout         time.Duration           // Maximum flow execution duration
}

// StepBuilder is used to construct an individual step within a flow.
// It accumulates step configuration including prompt specification, processing logic,
// and completion handlers before being converted to a flowStep.
type StepBuilder struct {
	name         string        // Unique step identifier within the flow
	promptConfig *PromptConfig // Declarative prompt configuration (message, image, keyboard)
	processFunc  ProcessFunc   // Function to process user input and determine next action
	flowBuilder  *FlowBuilder  // Reference to parent flow builder for fluent chaining
}

// PromptConfig represents the complete declarative specification for a step's prompt.
// It defines all visual and interactive elements that will be presented to the user,
// including the message content, optional image, interactive keyboard, and template data.
// The configuration is processed by the PromptComposer during step rendering.
type PromptConfig struct {
	Message      MessageSpec            // Message content: string, func(*Context) string, or "template:name"
	Image        ImageSpec              // Optional image: URL, file path, base64, func(*Context) string, or nil
	Keyboard     KeyboardFunc           // Optional interactive keyboard generator function
	TemplateData map[string]interface{} // Template variables that override context data during rendering
}

// MessageSpec defines the flexible type for message content in prompts.
// Supports static strings, dynamic functions, and template references for maximum flexibility.
//
// Supported Types:
//   - string: Static message text or template reference ("template:name")
//   - func(*Context) string: Dynamic message based on current context
//   - Other types are converted to string using fmt.Sprintf("%v", value)
//
// Examples:
//
//	MessageSpec("Hello World")                    // Static text
//	MessageSpec("template:welcome")               // Template reference
//	MessageSpec(func(ctx *Context) string {       // Dynamic content
//	  return fmt.Sprintf("Hello %s", ctx.Get("name"))
//	})
type MessageSpec interface{}

// ImageSpec defines the flexible type for image content in prompts.
// Supports static URLs, dynamic image generation, and various image formats.
//
// Supported Types:
//   - string: Static URL, file path, or base64-encoded image data
//   - func(*Context) string: Dynamic image URL/path based on current context
//   - nil: No image for this prompt
//
// Examples:
//
//	ImageSpec("https://example.com/image.jpg")    // Static URL
//	ImageSpec("./assets/welcome.png")             // File path
//	ImageSpec("data:image/png;base64,...")        // Base64 data
//	ImageSpec(func(ctx *Context) string {         // Dynamic image
//	  return fmt.Sprintf("https://api.avatar/%s", ctx.Get("username"))
//	})
type ImageSpec interface{}

// KeyboardFunc defines the function signature for generating interactive inline keyboards.
// The function receives the current context and returns an InlineKeyboardBuilder for
// fluent keyboard construction with buttons, callbacks, and other interactive elements.
//
// Parameters:
//   - ctx: Current context containing user data and flow state
//
// Returns:
//   - *InlineKeyboardBuilder: Builder for creating interactive keyboard layouts
//
// Example:
//
//	KeyboardFunc(func(ctx *Context) *InlineKeyboardBuilder {
//	  return NewInlineKeyboard().
//	    ButtonCallback("Confirm", "confirm").
//	    ButtonCallback("Cancel", "cancel")
//	})
type KeyboardFunc func(ctx *Context) *PromptKeyboardBuilder

// ProcessFunc defines the signature for functions that process user input within a flow step.
// This is the core processing function that receives user input (text or button clicks),
// applies business logic, updates context data, and determines the next flow action.
//
// Input Processing:
//   - input: Contains text message content or button callback data (as string)
//   - buttonClick: Provides detailed button click information if user clicked a button
//   - ctx: Current context with user data, flow state, and bot capabilities
//
// Return Value:
//   - ProcessResult: Specifies next action (NextStep, GoToStep, Retry, Complete, Cancel)
//   - Result can include optional prompt for user feedback
//
// Parameters:
//   - ctx: Current context containing user data and bot capabilities
//   - input: User input string (message text or button callback data)
//   - buttonClick: Detailed button information if input came from button click, nil for text
//
// Returns:
//   - ProcessResult: Flow navigation decision with optional user prompt
//
// Example:
//
//	ProcessFunc(func(ctx *Context, input string, click *ButtonClick) ProcessResult {
//	  if click != nil {
//	    // Handle button click
//	    return handleButtonClick(ctx, click.Data)
//	  }
//	  // Handle text input
//	  return handleTextInput(ctx, input)
//	})
type ProcessFunc func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult

// ButtonClick holds comprehensive information about a button click event passed to ProcessFunc.
// With the UUID mapping system, the Data field contains the original interface{} value
// that was passed to ButtonCallback(), providing type-safe access to callback data.
type ButtonClick struct {
	Data     interface{}            // Original callback data passed to ButtonCallback() (UUID-mapped, preserves original type)
	Text     string                 // Display text of the button that was clicked
	UserID   int64                  // Telegram user ID of the user who clicked the button
	ChatID   int64                  // Telegram chat ID where the button click occurred
	Metadata map[string]interface{} // Optional metadata for additional button context
}

// ProcessResult specifies the outcome of a ProcessFunc execution and determines flow navigation.
// It combines the next action to take with an optional prompt for user feedback.
// Results enable declarative flow control with clear separation of processing logic and navigation.
type ProcessResult struct {
	Action     processAction // Flow navigation action to execute
	TargetStep string        // Target step name for GoToStep action
	Prompt     *PromptConfig // Optional informational prompt to display before executing action
}

// WithPrompt adds a custom informational prompt to any ProcessResult for user feedback.
// The prompt is displayed before executing the navigation action, providing context about
// what's happening in the flow. Accepts flexible MessageSpec for static or dynamic content.
//
// Parameters:
//   - prompt: MessageSpec defining the prompt content (string, func(*Context) string, or template)
//
// Returns:
//   - ProcessResult: The same result with prompt attached for method chaining
//
// Examples:
//
//	NextStep().WithPrompt("Moving to next step!")
//	NextStep().WithPrompt(func(ctx *Context) string { return "Dynamic message" })
//	NextStep().WithPrompt("template:transition_message")
func (pr ProcessResult) WithPrompt(prompt MessageSpec) ProcessResult {

	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Message = prompt

	return pr
}

// WithImage adds an image to the ProcessResult's informational prompt using ImageSpec.
// This enables rich visual feedback when transitioning between flow steps or providing
// user guidance. The image supports the same flexibility as PromptBuilder.WithImage().
//
// Parameters:
//   - image: ImageSpec defining the image source (URL, file path, base64, or function)
//
// Returns:
//   - ProcessResult: The same result with image attached for method chaining
//
// Examples:
//
//	NextStep().WithPrompt("Success!").WithImage("https://example.com/success.jpg")
//	Retry().WithPrompt("Try again").WithImage(func(ctx *Context) string { return getDynamicImage(ctx) })
func (pr ProcessResult) WithImage(image ImageSpec) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.Image = image
	return pr
}

// WithTemplateData adds template variables to the ProcessResult's informational prompt.
// This enables dynamic template rendering with context-specific data for rich user feedback.
// Template data takes precedence over context data during prompt rendering.
//
// Parameters:
//   - data: Map of template variables to make available during prompt rendering
//
// Returns:
//   - ProcessResult: The same result with template data attached for method chaining
//
// Examples:
//
//	NextStep().WithPrompt("template:success").WithTemplateData(map[string]interface{}{
//	  "step_name": "registration",
//	  "progress": "50%",
//	})
func (pr ProcessResult) WithTemplateData(data map[string]interface{}) ProcessResult {
	if pr.Prompt == nil {
		pr.Prompt = &PromptConfig{}
	}
	pr.Prompt.TemplateData = data
	return pr
}

// ProcessMessageAction defines how previous messages are handled when processing button clicks.
// This controls the user experience by managing message history and preventing interaction
// with outdated buttons or content during flow progression.
type ProcessMessageAction int

const (
	ProcessKeepMessage    ProcessMessageAction = iota // Keep previous messages untouched, allowing scroll-back (default behavior)
	ProcessDeleteMessage                              // Delete entire previous message for clean conversation history
	ProcessDeleteKeyboard                             // Remove only keyboard from previous message, keeping content visible
)

// processAction defines the internal action type for flow navigation after ProcessFunc execution.
// These actions determine how the flow engine responds to ProcessFunc results and control
// the overall flow execution path. Used internally by ProcessResult helper functions.
type processAction int

// Constants defining all possible flow navigation actions available to ProcessFunc implementations.
// These actions provide comprehensive flow control from simple progression to complex navigation patterns.
const (
	actionNextStep     processAction = iota // Proceed to the next step in the flow's defined sequence
	actionGoToStep                          // Jump directly to a specific step by name, enabling conditional branching
	actionRetryStep                         // Retry the current step, optionally with a new prompt for user feedback
	actionCompleteFlow                      // Mark the entire flow as successfully completed and execute completion handlers
	actionCancelFlow                        // Cancel the current flow immediately, cleaning up resources and user state
)

// Helper functions for creating ProcessResult instances with fluent chaining support.
// These functions provide a clean, readable API for flow navigation decisions within ProcessFunc implementations.

// NextStep creates a ProcessResult that directs the flow to the next sequential step.
// This is the most common flow navigation action, moving through steps in their defined order.
// The result can be enhanced with informational prompts using .WithPrompt() chaining.
//
// Returns:
//   - ProcessResult: Result configured for next step navigation
//
// Example:
//
//	return NextStep() // Simple progression
//	return NextStep().WithPrompt("Moving to step 2...")
func NextStep() ProcessResult {
	return ProcessResult{Action: actionNextStep}
}

// GoToStep creates a ProcessResult that directs the flow to a specific step by name.
// This enables conditional branching and non-linear navigation within flows.
// Useful for implementing loops, conditional paths, or error recovery scenarios.
//
// Parameters:
//   - stepName: Name of the target step to navigate to
//
// Returns:
//   - ProcessResult: Result configured for targeted step navigation
//
// Example:
//
//	return GoToStep("confirmation") // Jump to specific step
//	return GoToStep("error_handler").WithPrompt("Redirecting to error handling...")
func GoToStep(stepName string) ProcessResult {
	return ProcessResult{Action: actionGoToStep, TargetStep: stepName}
}

// Retry creates a ProcessResult that retries the current step with optional new prompt.
// The original step prompt is re-rendered unless a custom prompt is provided via .WithPrompt().
// This is commonly used for input validation errors or user guidance scenarios.
//
// Returns:
//   - ProcessResult: Result configured for step retry
//
// Example:
//
//	return Retry() // Re-render original prompt
//	return Retry().WithPrompt("Invalid input. Please try again.")
func Retry() ProcessResult {
	return ProcessResult{Action: actionRetryStep}
}

// CompleteFlow creates a ProcessResult that marks the current flow as successfully completed.
// This triggers the flow's OnComplete handler (if defined) and cleans up user session state.
// Use this when the flow has accomplished its intended purpose.
//
// Returns:
//   - ProcessResult: Result configured for flow completion
//
// Example:
//
//	return CompleteFlow() // Simple completion
//	return CompleteFlow().WithPrompt("Registration completed successfully!")
func CompleteFlow() ProcessResult {
	return ProcessResult{Action: actionCompleteFlow}
}

// CancelFlow creates a ProcessResult that cancels the current flow immediately.
// This cleans up user session state and UUID mappings without executing completion handlers.
// Use this for error conditions or when the user explicitly requests cancellation.
//
// Returns:
//   - ProcessResult: Result configured for flow cancellation
//
// Example:
//
//	return CancelFlow() // Simple cancellation
//	return CancelFlow().WithPrompt("Flow cancelled at your request.")
func CancelFlow() ProcessResult {
	return ProcessResult{Action: actionCancelFlow}
}

// isTemplateMessage checks if a message string is a template reference with the "template:" prefix.
// This utility function is used internally by the prompt rendering system to distinguish between
// static text content and template references that need to be processed by the template engine.
//
// Template Format: "template:templateName"
// - The prefix "template:" indicates this is a template reference
// - Everything after the colon is treated as the template name
// - Template names should not contain additional colons
//
// Parameters:
//   - message: String to check for template reference format
//
// Returns:
//   - bool: true if the message is a template reference, false for static text
//   - string: extracted template name if it's a template reference, empty string otherwise
//
// Examples:
//
//	isTemplateMessage("Hello World")           // false, ""
//	isTemplateMessage("template:welcome")      // true, "welcome"
//	isTemplateMessage("template:user_greeting") // true, "user_greeting"
//	isTemplateMessage("template:")             // false, "" (empty template name)
func isTemplateMessage(message string) (bool, string) {
	const templatePrefix = "template:"
	if len(message) > len(templatePrefix) && message[:len(templatePrefix)] == templatePrefix {
		templateName := message[len(templatePrefix):]
		return true, templateName
	}
	return false, ""
}
