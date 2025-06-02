package teleflow

import (
	"log"
	"sync"
	"time"
)

// Middleware system provides a powerful and flexible way to intercept,
// modify, and extend bot request processing. Middleware functions execute
// in a chain pattern, allowing for cross-cutting concerns like authentication,
// logging, rate limiting, error handling, and custom business logic.
//
// The middleware system supports:
//   - Request/response interception and modification
//   - Authentication and authorization
//   - Rate limiting and throttling
//   - Request logging and monitoring
//   - Error handling and recovery
//   - Custom business logic injection
//   - Conditional middleware execution
//   - Middleware composition and chaining
//
// Basic Middleware Usage:
//
//	// Apply middleware globally to all handlers
//	bot.Use(teleflow.LoggingMiddleware())
//	bot.Use(teleflow.AuthMiddleware(authChecker))
//	bot.Use(teleflow.RateLimitMiddleware(10, time.Minute))
//
//	// Apply middleware to specific handlers
//	bot.HandleCommand("/admin", adminHandler, teleflow.RequireAdmin())
//
// Custom Middleware Creation:
//
//	func CustomMiddleware() teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				// Pre-processing logic
//				log.Printf("Request from user %d", ctx.UserID())
//
//				// Call next handler
//				err := next(ctx)
//
//				// Post-processing logic
//				if err != nil {
//					log.Printf("Handler error: %v", err)
//				}
//
//				return err
//			}
//		}
//	}
//
// Authentication Middleware:
//
//	func AuthMiddleware(checker UserPermissionChecker) teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				if !checker.CanExecute(ctx.UserID(), "access") {
//					return ctx.Reply("🚫 Access denied")
//				}
//				return next(ctx)
//			}
//		}
//	}
//
// Rate Limiting Middleware:
//
//	func RateLimitMiddleware(limit int, window time.Duration) teleflow.MiddlewareFunc {
//		limiter := NewRateLimiter(limit, window)
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				if !limiter.Allow(ctx.UserID()) {
//					return ctx.Reply("⏰ Rate limit exceeded. Please try again later.")
//				}
//				return next(ctx)
//			}
//		}
//	}
//
// Error Recovery Middleware:
//
//	func RecoveryMiddleware() teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) (err error) {
//				defer func() {
//					if r := recover(); r != nil {
//						log.Printf("Panic recovered: %v", r)
//						err = ctx.Reply("❌ An error occurred. Please try again.")
//					}
//				}()
//				return next(ctx)
//			}
//		}
//	}

// LoggingMiddleware logs all incoming updates and handler execution time
func LoggingMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()

			// Log incoming update
			updateType := "unknown"
			if ctx.Update.Message != nil {
				if ctx.Update.Message.IsCommand() {
					updateType = "command: " + ctx.Update.Message.Command()
				} else {
					updateType = "text: " + ctx.Update.Message.Text
				}
			} else if ctx.Update.CallbackQuery != nil {
				updateType = "callback: " + ctx.Update.CallbackQuery.Data
			}

			log.Printf("[%d] Processing %s", ctx.UserID(), updateType)

			// Execute handler
			err := next(ctx)

			// Log execution time
			duration := time.Since(start)
			if err != nil {
				log.Printf("[%d] Handler failed in %v: %v", ctx.UserID(), duration, err)
			} else {
				log.Printf("[%d] Handler completed in %v", ctx.UserID(), duration)
			}

			return err
		}
	}
}

// AuthMiddleware checks if user is authorized
func AuthMiddleware(checker UserPermissionChecker) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			if !checker.CanExecute(ctx.UserID(), "basic_access") {
				return ctx.Reply("🚫 You are not authorized to use this bot.")
			}
			return next(ctx)
		}
	}
}

// RateLimitMiddleware implements simple rate limiting
func RateLimitMiddleware(requestsPerMinute int) MiddlewareFunc {
	userLastRequest := make(map[int64]time.Time)
	var mutex sync.RWMutex
	minInterval := time.Minute / time.Duration(requestsPerMinute)

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			userID := ctx.UserID()
			now := time.Now()

			mutex.Lock()
			defer mutex.Unlock()

			if lastRequest, exists := userLastRequest[userID]; exists {
				if now.Sub(lastRequest) < minInterval {
					return ctx.Reply("⏳ Please wait before sending another message.")
				}
			}

			userLastRequest[userID] = now
			return next(ctx)
		}
	}
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in handler for user %d: %v", ctx.UserID(), r)
					err = ctx.Reply("An unexpected error occurred. Please try again.")
				}
			}()
			return next(ctx)
		}
	}
}
