package teleflow

// CallbackHandlerFunc defines the signature for callback query handlers
// specifically for use with the new middleware system. This signature
// is based on the requirements for SimpleCallback or similar mechanisms.
type CallbackHandlerFunc func(ctx *Context, fullCallbackData string, extractedData string) error

// CommandMiddlewareFunc defines the signature for command-specific middleware.
// It uses the existing CommandHandlerFunc type, which is expected to be defined
// in core/bot.go (e.g., func(ctx *Context, command string, args string) error).
type CommandMiddlewareFunc func(next CommandHandlerFunc) CommandHandlerFunc

// TextMiddlewareFunc defines the signature for text-specific middleware.
// It uses the existing TextHandlerFunc type, which is expected to be defined
// in core/bot.go (e.g., func(ctx *Context, text string) error).
type TextMiddlewareFunc func(next TextHandlerFunc) TextHandlerFunc

// DefaultTextMiddlewareFunc defines the signature for default text-specific middleware.
// It uses the existing DefaultTextHandlerFunc type, which is expected to be defined
// in core/bot.go (e.g., func(ctx *Context, fullMessageText string) error).
type DefaultTextMiddlewareFunc func(next DefaultTextHandlerFunc) DefaultTextHandlerFunc

// CallbackMiddlewareFunc defines the signature for callback-specific middleware.
// It uses the CallbackHandlerFunc type defined in this file.
type CallbackMiddlewareFunc func(next CallbackHandlerFunc) CallbackHandlerFunc

// FlowStepInputMiddlewareFunc defines the signature for flow step input-specific middleware.
// It uses the existing FlowStepInputHandlerFunc type, which is expected to be defined
// in core/flow.go (e.g., func(ctx *Context, input string) error).
type FlowStepInputMiddlewareFunc func(next FlowStepInputHandlerFunc) FlowStepInputHandlerFunc

// Note: Middleware for other flow handlers (FlowStepStart, FlowCompletion, FlowCancellation)
// can be defined similarly if needed, following the pattern:
// type FlowStepStartMiddlewareFunc func(next FlowStepStartHandlerFunc) FlowStepStartHandlerFunc
// type FlowCompletionMiddlewareFunc func(next FlowCompletionHandlerFunc) FlowCompletionHandlerFunc
// type FlowCancellationMiddlewareFunc func(next FlowCancellationHandlerFunc) FlowCancellationHandlerFunc
