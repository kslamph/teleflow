// Middleware Type System provides type-specific middleware function signatures
// for Teleflow's handler system. This enables granular middleware application
// to specific handler types rather than applying middleware globally.
//
// The type-specific middleware system offers:
//   - Command-specific middleware for command handlers only
//   - Text-specific middleware for text message handlers
//   - Callback-specific middleware for callback query handlers
//   - Flow step input middleware for flow input handlers
//   - Type-safe middleware chaining with proper signatures
//
// Basic Usage:
//
//	// Define command-specific middleware
//	func commandLoggingMiddleware(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc {
//		return func(ctx *teleflow.Context, command string, args string) error {
//			log.Printf("Command: %s, Args: %s", command, args)
//			return next(ctx, command, args)
//		}
//	}
//
//	// Apply to bot
//	bot.UseCommandMiddleware(commandLoggingMiddleware)
//
// Text Handler Middleware:
//
//	func textValidationMiddleware(next teleflow.TextHandlerFunc) teleflow.TextHandlerFunc {
//		return func(ctx *teleflow.Context, text string) error {
//			if len(text) > 1000 {
//				return ctx.Reply("Message too long")
//			}
//			return next(ctx, text)
//		}
//	}
//
//	bot.UseTextMiddleware(textValidationMiddleware)
//
// Callback Query Middleware:
//
//	func callbackAuthMiddleware(next teleflow.CallbackHandlerFunc) teleflow.CallbackHandlerFunc {
//		return func(ctx *teleflow.Context, fullCallbackData string, extractedData string) error {
//			if !isAuthorized(ctx.UserID()) {
//				return ctx.Reply("Access denied")
//			}
//			return next(ctx, fullCallbackData, extractedData)
//		}
//	}
//
//	bot.UseCallbackMiddleware(callbackAuthMiddleware)
//
// Flow Step Input Middleware:
//
//	func flowInputValidationMiddleware(next teleflow.FlowStepInputHandlerFunc) teleflow.FlowStepInputHandlerFunc {
//		return func(ctx *teleflow.Context, input string) error {
//			if input == "" {
//				return ctx.Reply("Input cannot be empty")
//			}
//			return next(ctx, input)
//		}
//	}
//
//	bot.UseFlowStepInputMiddleware(flowInputValidationMiddleware)
//
// This type-specific approach provides better type safety, clearer intent,
// and more efficient middleware execution compared to generic middleware.
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
