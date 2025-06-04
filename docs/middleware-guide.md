# Teleflow Middleware Guide

Middleware in Teleflow provides a powerful mechanism to process incoming updates before they reach your main [Handlers](handlers-guide.md) and to perform actions after handlers have executed. This allows you to implement cross-cutting concerns like logging, authentication, rate limiting, error recovery, and more, in a clean and reusable way.

## Table of Contents

- [What is Middleware?](#what-is-middleware)
- [How Middleware Works](#how-middleware-works)
  - [The `MiddlewareFunc` Signature](#the-middlewarefunc-signature)
  - [Execution Flow](#execution-flow)
- [Registering Middleware](#registering-middleware)
  - [Global Middleware](#global-middleware)
- [Built-in Middleware](#built-in-middleware)
  - [1. `LoggingMiddleware()`](#1-loggingmiddleware)
  - [2. `AuthMiddleware(accessManager AccessManager)`](#2-authmiddlewareaccessmanager-accessmanager)
  - [3. `RateLimitMiddleware(requestsPerMinute int)`](#3-ratelimitmiddlewarerequestsperminute-int)
  - [4. `RecoveryMiddleware()`](#4-recoverymiddleware)
- [Creating Custom Middleware](#creating-custom-middleware)
  - [Basic Structure](#basic-structure)
  - [Example: Request Timer Middleware](#example-request-timer-middleware)
  - [Example: User Language Setter Middleware](#example-user-language-setter-middleware)
- [Middleware Chaining and Order](#middleware-chaining-and-order)
- [Best Practices for Middleware](#best-practices-for-middleware)
- [Next Steps](#next-steps)

## What is Middleware?

Middleware functions are essentially "interceptors" that sit between the raw Telegram update and your specific bot logic (handlers). They form a chain where each middleware can:
- Inspect or modify the `teleflow.Context`.
- Perform actions before calling the next middleware or handler in the chain.
- Perform actions after the next middleware or handler has completed.
- Choose to halt the processing chain and respond directly (e.g., for an authentication failure).

## How Middleware Works

### The `MiddlewareFunc` Signature
A middleware in Teleflow is a function that matches the `teleflow.MiddlewareFunc` signature:
```go
type MiddlewareFunc func(next teleflow.HandlerFunc) teleflow.HandlerFunc
```
It takes one `teleflow.HandlerFunc` (representing the next step in the chain) and returns another `teleflow.HandlerFunc`. The returned function is what actually gets executed for an incoming request.

### Execution Flow
Consider a middleware:
```go
func MyMiddleware() teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            // 1. Code here runs BEFORE the next middleware/handler
            log.Println("MyMiddleware: Before next handler")

            // 2. Call the next middleware/handler in the chain
            err := next(ctx)

            // 3. Code here runs AFTER the next middleware/handler completes
            log.Println("MyMiddleware: After next handler")
            if err != nil {
                log.Printf("MyMiddleware: Next handler returned error: %v", err)
            }
            return err // Propagate or handle the error
        }
    }
}
```

## Registering Middleware

### Global Middleware
You can apply middleware globally to all handlers (command, text, and callback) using `bot.Use()`:
```go
bot.Use(teleflow.LoggingMiddleware())
bot.Use(MyCustomMiddleware())
```
Middleware registered with `bot.Use()` will be applied to every handler.

## Built-in Middleware

Teleflow comes with several pre-built middleware functions found in `core/middleware.go`:

### 1. `LoggingMiddleware()`
Logs incoming updates and the time taken for handlers to execute.
```go
bot.Use(teleflow.LoggingMiddleware())
```
Output might look like:
```
[INFO][12345678] Processing command: start
[DEBUG][12345678] Handler completed in 2.5ms
```

### 2. `AuthMiddleware(accessManager AccessManager)`
Checks user permissions using a provided `teleflow.AccessManager` implementation. If `accessManager.CheckPermission()` returns an error, the middleware replies with the error message and stops further processing.
```go
type MyAccessManager struct { /* ... */ }
func (am *MyAccessManager) CheckPermission(ctx *teleflow.PermissionContext) error {
    if ctx.Command == "admin" && !isAdmin(ctx.UserID) {
        return errors.New("You are not authorized to use this command.")
    }
    return nil
}
// ... implement other AccessManager methods ...

accessMgr := &MyAccessManager{}
bot.Use(teleflow.AuthMiddleware(accessMgr))
```
The `PermissionContext` passed to `CheckPermission` contains details like `UserID`, `ChatID`, `Command`, `Arguments`, and the raw `Update`.

### 3. `RateLimitMiddleware(requestsPerMinute int)`
Provides simple rate limiting based on user ID. If a user exceeds the specified number of requests per minute, they receive a "Please wait" message.
```go
// Allow 10 requests per minute per user
bot.Use(teleflow.RateLimitMiddleware(10))
```

### 4. `RecoveryMiddleware()`
Recovers from panics that might occur in subsequent middleware or handlers. It logs the panic and sends a generic error message to the user, preventing the bot from crashing.
```go
bot.Use(teleflow.RecoveryMiddleware())
```
It's generally a good idea to add this early in your middleware chain.

## Creating Custom Middleware

### Basic Structure
Follow the `MiddlewareFunc` signature:
```go
package main

import (
	"log"
	teleflow "github.com/kslamph/teleflow/core"
)

func MyCustomMiddleware() teleflow.MiddlewareFunc {
	return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
		// This inner function is the actual middleware logic
		return func(ctx *teleflow.Context) error {
			// --- Pre-processing ---
			log.Printf("CustomMiddleware: User %d triggered an update.", ctx.UserID())
			
			// You can modify the context here if needed, e.g., ctx.Set("my_key", "my_value")

			// Call the next middleware or the main handler
			err := next(ctx)

			// --- Post-processing ---
			if err != nil {
				log.Printf("CustomMiddleware: Handler for user %d returned error: %v", ctx.UserID(), err)
				// You could potentially transform the error or handle it here
			} else {
				log.Printf("CustomMiddleware: Handler for user %d completed successfully.", ctx.UserID())
			}
			
			return err // Return the error (or nil) from the next handler
		}
	}
}

// Usage:
// bot.Use(MyCustomMiddleware())
```

### Example: Request Timer Middleware
```go
import (
	"log"
	"time"
	teleflow "github.com/kslamph/teleflow/core"
)

func RequestTimerMiddleware() teleflow.MiddlewareFunc {
	return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
		return func(ctx *teleflow.Context) error {
			start := time.Now()
			
			err := next(ctx) // Execute the rest of the chain
			
			duration := time.Since(start)
			log.Printf("Request for user %d took %s", ctx.UserID(), duration)
			
			return err
		}
	}
}
```

### Example: User Language Setter Middleware
This middleware could try to determine the user's language and set it in the context for other handlers or templates to use.
```go
func UserLanguageMiddleware(defaultLang string) teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            userLang := defaultLang
            if ctx.Update.Message != nil && ctx.Update.Message.From != nil {
                // Basic language detection (can be more sophisticated)
                if ctx.Update.Message.From.LanguageCode != "" {
                    userLang = ctx.Update.Message.From.LanguageCode
                }
            }
            ctx.Set("user_language", userLang) // Store in context
            log.Printf("User %d language set to: %s", ctx.UserID(), userLang)
            return next(ctx)
        }
    }
}

// Usage:
// bot.Use(UserLanguageMiddleware("en"))
```

## Middleware Chaining and Order

Middleware functions are executed in the reverse order they are added using `bot.Use()`.
If you register them like this:
```go
bot.Use(MiddlewareA)
bot.Use(MiddlewareB)
bot.Use(MiddlewareC)
```
The execution order for the "before next" part will be C -> B -> A -> Handler.
The "after next" part will execute in the order Handler -> A -> B -> C.

This "onion-like" layering is standard in many web frameworks.

## Best Practices for Middleware

- **Keep Middleware Focused**: Each middleware should ideally do one thing well (e.g., logging, auth, rate limiting).
- **Order Matters**: Be mindful of the order in which you register middleware, especially for dependencies (e.g., recovery middleware early, auth before sensitive operations).
- **Error Handling**: Decide if your middleware should handle errors from `next(ctx)` or propagate them.
- **Context Modification**: If middleware modifies the context (e.g., `ctx.Set()`), ensure keys are unique or well-namespaced to avoid collisions.
- **Performance**: Be cautious of computationally expensive operations in middleware, as they run for every applicable request.

## Next Steps

- [Handlers Guide](handlers-guide.md): Learn about the functions that middleware ultimately wrap.
- [Flow Guide](flow-guide.md): See how middleware interacts with conversational flows.
- [API Reference](api-reference.md): For detailed information on `MiddlewareFunc`, `AccessManager`, and other related types.
- Explore the `core/middleware.go` file to see the implementation of built-in middleware.