package teleflow

// MiddlewareFunc represents a middleware function that wraps a handler.
// Middleware functions can perform operations before and after calling the next handler,
// enabling cross-cutting concerns like logging, authentication, and error handling.
//
// Middleware follows the standard pattern where each function takes a handler
// and returns a new handler that wraps the original functionality.
//
// Example custom middleware:
//
//	func CustomTimingMiddleware() teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				start := time.Now()
//				err := next(ctx)
//				duration := time.Since(start)
//				log.Printf("Handler for user %d took %v", ctx.UserID(), duration)
//				return err
//			}
//		}
//	}
type MiddlewareFunc func(HandlerFunc) HandlerFunc
