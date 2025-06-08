//

//

//

//

//

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
//
//	bot.UseMiddleware(CustomLoggingMiddleware())
//

//
//	func AuthMiddleware(accessManager AccessManager) teleflow.MiddlewareFunc {
//		return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
//			return func(ctx *teleflow.Context) error {
//
//				if !isAuthorized(ctx.UserID(), ctx.ChatID()) {
//					return ctx.Reply("🚫 Access denied")
//				}
//				return next(ctx)
//			}
//		}
//	}
//

//
//
//	bot.UseMiddleware(teleflow.RecoveryMiddleware())
//	bot.UseMiddleware(teleflow.LoggingMiddleware())
//	bot.UseMiddleware(teleflow.RateLimitMiddleware(10))
//

package teleflow

type MiddlewareFunc func(next HandlerFunc) HandlerFunc
