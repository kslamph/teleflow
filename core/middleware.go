package teleflow

import (
	"log"
	"sync"
	"time"
)

// Middleware system provides a simple way to add cross-cutting functionality
// like authentication, logging, and rate limiting to your bot. Middleware
// functions execute in a chain pattern before your handlers run.
//
// Supported middleware:
//   - Authentication and authorization
//   - Request logging and monitoring
//   - Rate limiting and throttling
//   - Error handling and recovery
//   - Custom business logic
//
// Basic Usage:
//
//	// Apply middleware globally to all handlers
//	bot.UseMiddleware(teleflow.LoggingMiddleware())
//	bot.UseMiddleware(teleflow.AuthMiddleware(authChecker))
//	bot.UseMiddleware(teleflow.RateLimitMiddleware(10))
//
// Custom Middleware:
//
//	func CustomMiddleware() teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				// Pre-processing
//				log.Printf("Request from user %d", ctx.UserID())
//
//				// Call next handler
//				err := next(ctx)
//
//				// Post-processing
//				if err != nil {
//					log.Printf("Handler error: %v", err)
//				}
//				return err
//			}
//		}
//	}
//
// Built-in Middleware Examples:
//
//	// Authentication
//	bot.UseMiddleware(teleflow.AuthMiddleware(accessManager))
//
//	// Rate limiting (10 requests per minute)
//	bot.UseMiddleware(teleflow.RateLimitMiddleware(10))
//
//	// Error recovery
//	bot.UseMiddleware(teleflow.RecoveryMiddleware())
//
//	// Request logging
//	bot.UseMiddleware(teleflow.LoggingMiddleware())

// LoggingMiddleware logs all incoming updates and handler execution time
func LoggingMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()

			// Get debug and log level from context
			debug := false
			logLevel := "info"
			if debugVal, exists := ctx.Get("debug"); exists {
				if d, ok := debugVal.(bool); ok {
					debug = d
				}
			}
			if logLevelVal, exists := ctx.Get("logLevel"); exists {
				if ll, ok := logLevelVal.(string); ok {
					logLevel = ll
				}
			}

			// Log incoming update based on log level
			updateType := "unknown"
			if ctx.Update.Message != nil {
				if ctx.Update.Message.IsCommand() {
					updateType = "command: " + ctx.Update.Message.Command()
				} else {
					updateType = "text: " + ctx.Update.Message.Text
					if len(updateType) > 100 {
						updateType = updateType[:100] + "..."
					}
				}
			} else if ctx.Update.CallbackQuery != nil {
				updateType = "callback: " + ctx.Update.CallbackQuery.Data
			}

			// Log based on level and debug settings
			if debug || logLevel == "debug" {
				log.Printf("[DEBUG][%d] Processing %s", ctx.UserID(), updateType)
			} else if logLevel == "info" {
				log.Printf("[INFO][%d] Processing %s", ctx.UserID(), updateType)
			}

			// Execute handler
			err := next(ctx)

			// Log execution time
			duration := time.Since(start)
			if err != nil {
				log.Printf("[ERROR][%d] Handler failed in %v: %v", ctx.UserID(), duration, err)
			} else if debug || logLevel == "debug" {
				log.Printf("[DEBUG][%d] Handler completed in %v", ctx.UserID(), duration)
			}

			return err
		}
	}
}

// AuthMiddleware checks if user is authorized
func AuthMiddleware(accessManager AccessManager) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			// Create permission context
			permCtx := &PermissionContext{
				UserID: ctx.UserID(),
				ChatID: ctx.ChatID(),
				Update: &ctx.Update,
			}

			// Extract command and arguments if available
			if ctx.Update.Message != nil && ctx.Update.Message.IsCommand() {
				permCtx.Command = ctx.Update.Message.Command()
				if args := ctx.Update.Message.CommandArguments(); args != "" {
					permCtx.Arguments = []string{args}
				}
			}

			// Check if it's a group chat
			if ctx.Update.Message != nil {
				permCtx.IsGroup = ctx.Update.Message.Chat.IsGroup() || ctx.Update.Message.Chat.IsSuperGroup()
				permCtx.MessageID = ctx.Update.Message.MessageID
			} else if ctx.Update.CallbackQuery != nil && ctx.Update.CallbackQuery.Message != nil {
				permCtx.IsGroup = ctx.Update.CallbackQuery.Message.Chat.IsGroup() || ctx.Update.CallbackQuery.Message.Chat.IsSuperGroup()
				permCtx.MessageID = ctx.Update.CallbackQuery.Message.MessageID
			}

			if err := accessManager.CheckPermission(permCtx); err != nil {
				return ctx.Reply("🚫 " + err.Error())
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
