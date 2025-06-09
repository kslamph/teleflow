//

//

//

//

//
//
//	bot.UseMiddleware(teleflow.LoggingMiddleware())
//	bot.UseMiddleware(teleflow.RecoveryMiddleware())
//	bot.UseMiddleware(teleflow.RateLimitMiddleware(10))
//	bot.UseMiddleware(teleflow.AuthMiddleware(accessManager))
//
//
//	bot, err := teleflow.NewBot(token, teleflow.WithAccessManager(accessManager))
//
//

//
//	func CustomTimingMiddleware() teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//				start := time.Now()
//
//
//				err := next(ctx)
//
//
//				duration := time.Since(start)
//				log.Printf("Handler for user %d took %v", ctx.UserID(), duration)
//				return err
//			}
//		}
//	}
//

//

//

//

package teleflow

import (
	"log"
	"sync"
	"time"
)

func LoggingMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()

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

			updateType := "unknown"
			if ctx.update.Message != nil {
				if ctx.update.Message.IsCommand() {
					updateType = "command: " + ctx.update.Message.Command()
				} else {
					updateType = "text: " + ctx.update.Message.Text
					if len(updateType) > 100 {
						updateType = updateType[:100] + "..."
					}
				}
			} else if ctx.update.CallbackQuery != nil {
				updateType = "callback: " + ctx.update.CallbackQuery.Data
			}

			if debug || logLevel == "debug" {
				log.Printf("[DEBUG][%d] Processing %s", ctx.UserID(), updateType)
			} else if logLevel == "info" {
				log.Printf("[INFO][%d] Processing %s", ctx.UserID(), updateType)
			}

			err := next(ctx)

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

func AuthMiddleware(accessManager AccessManager) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {

			permCtx := ctx.getPermissionContext()

			if ctx.update.Message != nil && ctx.update.Message.IsCommand() {
				permCtx.Command = ctx.update.Message.Command()
				if args := ctx.update.Message.CommandArguments(); args != "" {
					permCtx.Arguments = []string{args}
				}
			}

			if ctx.update.Message != nil {
				permCtx.IsGroup = ctx.update.Message.Chat.IsGroup() || ctx.update.Message.Chat.IsSuperGroup()
				permCtx.MessageID = ctx.update.Message.MessageID
			} else if ctx.update.CallbackQuery != nil && ctx.update.CallbackQuery.Message != nil {
				permCtx.IsGroup = ctx.update.CallbackQuery.Message.Chat.IsGroup() || ctx.update.CallbackQuery.Message.Chat.IsSuperGroup()
				permCtx.MessageID = ctx.update.CallbackQuery.Message.MessageID
			}

			if err := accessManager.CheckPermission(permCtx); err != nil {
				return ctx.sendSimpleText("🚫 " + err.Error())
			}

			// Get reply keyboard from access manager and set it as pending
			if keyboard := accessManager.GetReplyKeyboard(permCtx); keyboard != nil {
				ctx.SetPendingReplyKeyboard(keyboard)
			}

			return next(ctx)
		}
	}
}

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
					return ctx.sendSimpleText("⏳ Please wait before sending another message.")
				}
			}

			userLastRequest[userID] = now
			return next(ctx)
		}
	}
}

func RecoveryMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in handler for user %d: %v", ctx.UserID(), r)
					err = ctx.sendSimpleText("❗An unexpected error occurred. Please try again.")
				}
			}()
			return next(ctx)
		}
	}
}
