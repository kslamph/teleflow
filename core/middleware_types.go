// Simplified Middleware System provides a unified middleware approach
// for intercepting any type of message in TeleFlow. This replaces the
// complex type-specific middleware system with a single, general-purpose
// middleware that can handle all message types.
//
// The unified middleware intercepts all incoming updates before they
// reach specific handlers, allowing for consistent logging, auth,
// rate limiting, and other cross-cutting concerns.
//
// Basic Usage:
//
//	// Define general middleware that handles all message types
//	func loggingMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//		return func(ctx *teleflow.Context) error {
//			log.Printf("Update from user %d: %+v", ctx.UserID(), ctx.Update)
//			return next(ctx)
//		}
//	}
//
//	// Apply to bot - will intercept all messages
//	bot.UseMiddleware(loggingMiddleware)
//
// Authentication Example:
//
//	func authMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//		return func(ctx *teleflow.Context) error {
//			if !isAuthorized(ctx.UserID()) {
//				return ctx.Reply("Access denied")
//			}
//			return next(ctx)
//		}
//	}
package teleflow

// MiddlewareFunc defines the unified middleware signature that can intercept
// any type of message or update. This replaces all type-specific middleware.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// Internal middleware types (not exposed to users)
// These are used internally for callback and flow system operations
