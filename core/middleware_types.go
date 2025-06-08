// Package teleflow provides middleware type definitions for the unified middleware system.
//
// The middleware system in Teleflow uses a unified approach where a single MiddlewareFunc
// type can intercept and process any type of update (commands, text messages, callbacks, etc.).
// This eliminates the complexity of type-specific middleware and provides a consistent
// interface for cross-cutting concerns like authentication, logging, and rate limiting.
//
// Middleware Execution Order:
//
// Middleware functions are applied in reverse order of registration (LIFO - Last In, First Out).
// If you register middleware A, B, C in that order, they execute as:
// C -> B -> A -> handler -> A -> B -> C
//
// This allows the last registered middleware to be the outermost wrapper, which is
// typically desired for security middleware like authentication and rate limiting.
//
// Basic Custom Middleware:
//
//	func CustomLoggingMiddleware() teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				log.Printf("Processing update from user %d", ctx.UserID())
//				err := next(ctx)
//				if err != nil {
//					log.Printf("Handler error: %v", err)
//				}
//				return err
//			}
//		}
//	}
//
//	// Apply to bot - will intercept all handler executions
//	bot.UseMiddleware(CustomLoggingMiddleware())
//
// Context-Aware Middleware:
//
//	func AuthMiddleware(accessManager AccessManager) teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				// Access user info, update type, etc. from context
//				if !isAuthorized(ctx.UserID(), ctx.ChatID()) {
//					return ctx.Reply("🚫 Access denied")
//				}
//				return next(ctx)
//			}
//		}
//	}
//
// Built-in Middleware Usage:
//
//	// Apply multiple middleware in registration order
//	bot.UseMiddleware(teleflow.RecoveryMiddleware())     // Outermost
//	bot.UseMiddleware(teleflow.LoggingMiddleware())      // Middle
//	bot.UseMiddleware(teleflow.RateLimitMiddleware(10))  // Innermost
//
// The unified approach ensures that all handler types (commands, text, callbacks, flows)
// receive the same middleware processing, providing consistent behavior across your bot.
package teleflow

// MiddlewareFunc defines the unified middleware function signature for the Teleflow framework.
//
// This function type enables intercepting and processing any type of update before it reaches
// specific handlers. The middleware receives the next handler in the chain and returns a
// new handler that wraps the original with additional functionality.
//
// The middleware pattern allows for:
//   - Pre-processing: Execute code before the handler (logging, auth checks, validation)
//   - Post-processing: Execute code after the handler (cleanup, response modification)
//   - Error handling: Catch and handle errors from downstream handlers
//   - Request modification: Modify the context before passing to the next handler
//
// Parameters:
//   - next: The next HandlerFunc in the middleware chain or the final handler
//
// Returns:
//   - HandlerFunc: A new handler that wraps the next handler with middleware logic
//
// Example Implementation:
//
//	func TimingMiddleware() MiddlewareFunc {
//		return func(next HandlerFunc) HandlerFunc {
//			return func(ctx *Context) error {
//				start := time.Now()
//				err := next(ctx)
//				duration := time.Since(start)
//				log.Printf("Handler took %v", duration)
//				return err
//			}
//		}
//	}
type MiddlewareFunc func(next HandlerFunc) HandlerFunc
