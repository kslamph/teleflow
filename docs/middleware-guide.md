# Teleflow Middleware Guide

Middleware in Teleflow provides a powerful mechanism to process incoming updates before they reach your main [Handlers](handlers-guide.md) and to perform actions after handlers have executed. Teleflow supports both **general middleware** that applies to all handler types and **type-specific middleware** for more targeted control. This dual approach allows you to implement cross-cutting concerns like logging, authentication, rate limiting, and error recovery with maximum flexibility.

## Table of Contents

- [General Middleware System](#general-middleware-system)
  - [`bot.Use()` - Universal Middleware](#botuse---universal-middleware)
  - [General Middleware Benefits](#general-middleware-benefits)
- [Type-Specific Middleware System](#type-specific-middleware-system)
  - [What is Type-Specific Middleware?](#what-is-type-specific-middleware)
  - [Benefits of Type-Specific Middleware](#benefits-of-type-specific-middleware)
- [Defining Middleware](#defining-middleware)
  - [General Middleware: `MiddlewareFunc`](#general-middleware-middlewarefunc)
  - [Type-Specific Signatures](#type-specific-signatures)
    - [`CommandMiddlewareFunc`](#commandmiddlewarefunc)
    - [`TextMiddlewareFunc`](#textmiddlewarefunc)
    - [`DefaultTextMiddlewareFunc`](#defaulttextmiddlewarefunc)
    - [`CallbackMiddlewareFunc`](#callbackmiddlewarefunc)
    - [`FlowStepInputMiddlewareFunc`](#flowstepinputmiddlewarefunc)
  - [The `next` Parameter](#the-next-parameter)
  - [Conceptual Example: Logging Middleware](#conceptual-example-logging-middleware)
- [Registering Middleware](#registering-middleware)
  - [`bot.Use()` - General Middleware](#botuse---general-middleware)
  - [`bot.UseCommandMiddleware()`](#botusecommandmiddleware)
  - [`bot.UseTextMiddleware()`](#botusetextmiddleware)
  - [`bot.UseDefaultTextMiddleware()`](#botusedefaulttextmiddleware)
  - [`bot.UseCallbackMiddleware()`](#botusecallbackmiddleware)
  - [`bot.UseFlowStepInputMiddleware()`](#botuseflowstepinputmiddleware)
  - [Registration Example](#registration-example)
- [Middleware Execution Order](#middleware-execution-order)
- [Common Middleware Examples](#common-middleware-examples)
  - [General Middleware Examples](#general-middleware-examples)
  - [Type-Specific Middleware Examples](#type-specific-middleware-examples)
  - [Recovery Middleware](#recovery-middleware)
  - [Authentication Middleware](#authentication-middleware)
- [Middleware for Flow Handlers](#middleware-for-flow-handlers)
- [Best Practices for Middleware](#best-practices-for-middleware)
- [Next Steps](#next-steps)

## General Middleware System

### `bot.Use()` - Universal Middleware

The general middleware system allows you to apply middleware to **all handler types** with a single method call. This is perfect for cross-cutting concerns that should apply universally across your bot.

```go
// General middleware that applies to ALL handler types
func loggingMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
    return func(ctx *teleflow.Context) error {
        start := time.Now()
        log.Printf("Handler started: %s", ctx.Update.Message.Text)
        
        err := next(ctx)
        
        log.Printf("Handler completed in %v", time.Since(start))
        return err
    }
}

// Apply to all handler types at once
bot.Use(loggingMiddleware)
```

When you use `bot.Use()`, Teleflow automatically converts your general middleware to the appropriate type-specific middleware and applies it to:
- Command handlers
- Text handlers
- Default text handlers
- Callback handlers
- Flow step input handlers

### General Middleware Benefits

- **Simplicity**: One method call applies middleware everywhere
- **Consistency**: Ensures the same behavior across all handler types
- **Less Code**: No need to adapt middleware for each handler type manually
- **Perfect for Cross-Cutting Concerns**: Logging, authentication, recovery, rate limiting

## Type-Specific Middleware System

### What is Type-Specific Middleware?

Previously, Teleflow used a generic middleware signature. The new system introduces distinct middleware types for different handler categories: commands, specific text messages, default text messages, and callback queries. This means a middleware function is now designed explicitly for the kind of handler it will wrap.

Middleware functions are "interceptors" that sit between the raw Telegram update and your specific bot logic. They form a chain where each middleware can:
- Inspect or modify the `teleflow.Context`.
- Perform actions before calling the next middleware or the actual handler in the chain.
- Perform actions after the next middleware or handler has completed.
- Choose to halt the processing chain and respond directly (e.g., for an authentication failure).

## Benefits of Type-Specific Middleware

- **Type Safety**: Middleware signatures are now aligned with the specific handler signatures they wrap. This means you get compile-time checks for parameter types (e.g., a command middleware knows it will receive `command` and `args` strings).
- **Clarity and Intent**: It's clearer what kind of processing a middleware is intended for.
- **Reduced Boilerplate**: Less type assertion or casting might be needed within middleware, as the types are more specific.

## Defining Middleware

Middleware in Teleflow can be defined in two ways: as general middleware that applies to all handler types, or as type-specific middleware for particular handler categories.

### General Middleware: `MiddlewareFunc`

General middleware uses the universal `MiddlewareFunc` signature and works with all handler types:

```go
// General middleware signature
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// Where HandlerFunc is the universal handler signature
type HandlerFunc func(ctx *Context) error
```

**Example of general middleware:**
```go
func loggingMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
    return func(ctx *teleflow.Context) error {
        log.Printf("Processing update from user %d", ctx.Update.Message.From.ID)
        err := next(ctx)
        if err != nil {
            log.Printf("Handler error: %v", err)
        }
        return err
    }
}

// Apply to all handler types
bot.Use(loggingMiddleware)
```

### Type-Specific Signatures

For more granular control, you can define middleware specific to particular handler types. Each type-specific middleware is aligned with the signature of the handlers it wraps.

### `CommandMiddlewareFunc`
For middleware that processes commands.
```go
// From core/middleware_types.go
type CommandMiddlewareFunc func(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc

// The CommandHandlerFunc it wraps (from core/bot.go)
// type CommandHandlerFunc func(ctx *teleflow.Context, command string, args string) error
```

### `TextMiddlewareFunc`
For middleware that processes specific registered text messages.
```go
// From core/middleware_types.go
type TextMiddlewareFunc func(next teleflow.TextHandlerFunc) teleflow.TextHandlerFunc

// The TextHandlerFunc it wraps (from core/bot.go)
// type TextHandlerFunc func(ctx *teleflow.Context, text string) error
```

### `DefaultTextMiddlewareFunc`
For middleware that processes messages handled by the default text handler.
```go
// From core/middleware_types.go
type DefaultTextMiddlewareFunc func(next teleflow.DefaultTextHandlerFunc) teleflow.DefaultTextHandlerFunc

// The DefaultTextHandlerFunc it wraps (from core/bot.go)
// type DefaultTextHandlerFunc func(ctx *teleflow.Context, fullMessageText string) error
```

### `CallbackMiddlewareFunc`
For middleware that processes callback queries.
```go
// From core/middleware_types.go
type CallbackMiddlewareFunc func(next teleflow.CallbackHandlerFunc) teleflow.CallbackHandlerFunc

// The CallbackHandlerFunc it wraps (from core/middleware_types.go)
// type CallbackHandlerFunc func(ctx *teleflow.Context, fullCallbackData string, extractedData string) error
```

### The `next` Parameter
In each middleware signature, the `next` parameter represents the subsequent middleware in the chain or, if it's the last middleware, the actual handler function (e.g., `CommandHandlerFunc`, `TextHandlerFunc`).

Your middleware function **must** call `next(...)` to continue processing the update down the chain. If `next(...)` is not called, the chain is short-circuited, and subsequent middleware or the main handler will not execute.

### Conceptual Example: Logging Middleware
Here's how a logging middleware might look for commands:
```go
import (
	"log"
	teleflow "github.com/kslamph/teleflow/core" // Adjust import path as needed
)

// LoggingCommandMiddleware logs incoming commands and their arguments.
func LoggingCommandMiddleware() teleflow.CommandMiddlewareFunc {
	return func(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc {
		// This inner function is the actual middleware logic for commands
		return func(ctx *teleflow.Context, command string, args string) error {
			// 1. Code here runs BEFORE the next middleware/handler
			log.Printf("CommandMiddleware: UserID: %d, Command: /%s, Args: %s", ctx.UserID(), command, args)

			// 2. Call the next middleware/handler in the chain
			err := next(ctx, command, args) // Pass along command-specific parameters

			// 3. Code here runs AFTER the next middleware/handler completes
			if err != nil {
				log.Printf("CommandMiddleware: Handler for command /%s (UserID: %d) returned error: %v", command, ctx.UserID(), err)
			} else {
				log.Printf("CommandMiddleware: Handler for command /%s (UserID: %d) completed.", command, ctx.UserID())
			}
			return err // Propagate or handle the error
		}
	}
}
```
A similar logging middleware could be created for `TextMiddlewareFunc`, `DefaultTextMiddlewareFunc`, or `CallbackMiddlewareFunc`, adjusting the logged parameters accordingly.

## Registering Middleware

Teleflow provides both universal and type-specific middleware registration methods, allowing you to choose the approach that best fits your needs.

### `bot.Use()` - General Middleware

The `bot.Use()` method applies middleware to **all handler types** automatically. This is the recommended approach for cross-cutting concerns.

```go
// Apply to all handler types at once
bot.Use(loggingMiddleware)
bot.Use(authMiddleware)
bot.Use(recoveryMiddleware)
```

When you call `bot.Use()`, Teleflow automatically:
1. Stores the general middleware in the bot's middleware chain
2. Converts it to type-specific middleware using adapter functions
3. Applies it to all handler types (commands, text, callbacks, flow steps)

### Type-Specific Registration Methods

For more granular control, you can register middleware for specific handler categories:

### `bot.UseCommandMiddleware()`
Registers middleware specifically for command handlers.
```go
bot.UseCommandMiddleware(m teleflow.CommandMiddlewareFunc)
```

### `bot.UseTextMiddleware()`
Registers middleware specifically for text handlers (those registered with `bot.HandleText()`).
```go
bot.UseTextMiddleware(m teleflow.TextMiddlewareFunc)
```

### `bot.UseDefaultTextMiddleware()`
Registers middleware specifically for the default text handler (registered with `bot.SetDefaultTextHandler()`).
```go
bot.UseDefaultTextMiddleware(m teleflow.DefaultTextMiddlewareFunc)
```

### `bot.UseCallbackMiddleware()`
Registers middleware specifically for callback query handlers.
```go
bot.UseCallbackMiddleware(m teleflow.CallbackMiddlewareFunc)
```

### `bot.UseFlowStepInputMiddleware()`
Registers middleware specifically for flow step input handlers.
```go
bot.UseFlowStepInputMiddleware(m teleflow.FlowStepInputMiddlewareFunc)
```

### Registration Examples

**General Middleware Approach (Recommended for Cross-Cutting Concerns):**
```go
// Define general middleware once
func loggingMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
    return func(ctx *teleflow.Context) error {
        start := time.Now()
        log.Printf("Handler started")
        err := next(ctx)
        log.Printf("Handler completed in %v", time.Since(start))
        return err
    }
}

func authMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
    return func(ctx *teleflow.Context) error {
        if !isAuthorized(ctx.Update.Message.From.ID) {
            return ctx.Reply("Access denied")
        }
        return next(ctx)
    }
}

// Apply to all handler types at once
bot.Use(loggingMiddleware)
bot.Use(authMiddleware)
```

**Type-Specific Middleware Approach:**
```go
// Define type-specific middleware for granular control
func commandLoggingMiddleware(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc {
    return func(ctx *teleflow.Context, command string, args string) error {
        log.Printf("Command executed: %s %s", command, args)
        return next(ctx, command, args)
    }
}

// Apply only to command handlers
bot.UseCommandMiddleware(commandLoggingMiddleware)
```

**Mixed Approach:**
```go
// Apply general middleware for common concerns
bot.Use(recoveryMiddleware)      // Apply to all handlers
bot.Use(loggingMiddleware)       // Apply to all handlers

// Apply specific middleware for targeted functionality  
bot.UseCommandMiddleware(commandRateLimitMiddleware)  // Only for commands
bot.UseCallbackMiddleware(callbackValidationMiddleware) // Only for callbacks
```

## Middleware Execution Order

Middleware for a specific category is executed in the order it was added for that category. More precisely, they form an "onion-like" chain:
1. The last middleware added is the first one to execute its "pre-processing" logic (before calling `next`).
2. Control passes down the chain via `next` calls.
3. The actual handler for the category is executed.
4. Control passes back up the chain, with each middleware executing its "post-processing" logic (after the `next` call returns).

If you register them like this for commands:
```go
bot.UseCommandMiddleware(MiddlewareA) // CommandMiddlewareFunc
bot.UseCommandMiddleware(MiddlewareB) // CommandMiddlewareFunc
bot.UseCommandMiddleware(MiddlewareC) // CommandMiddlewareFunc
```
The execution order for the "before `next`" part will be C -> B -> A -> Actual Command Handler.
The "after `next`" part will execute in the order Actual Command Handler -> A -> B -> C.

## Common Middleware Examples

Here are practical examples of both general and type-specific middleware:

### General Middleware Examples

**Recovery Middleware (General):**
```go
func RecoveryMiddleware() teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) (err error) {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("Panic recovered: %v\n%s", r, debug.Stack())
                    err = ctx.Reply("Sorry, an internal error occurred.")
                }
            }()
            return next(ctx)
        }
    }
}

// Apply to all handler types
bot.Use(RecoveryMiddleware())
```

**Logging Middleware (General):**
```go
func LoggingMiddleware() teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            start := time.Now()
            userID := ctx.Update.Message.From.ID
            
            log.Printf("Handler started for user %d", userID)
            err := next(ctx)
            
            duration := time.Since(start)
            if err != nil {
                log.Printf("Handler failed for user %d after %v: %v", userID, duration, err)
            } else {
                log.Printf("Handler completed for user %d in %v", userID, duration)
            }
            return err
        }
    }
}

// Apply to all handler types
bot.Use(LoggingMiddleware())
```

**Authentication Middleware (General):**
```go
func AuthMiddleware(allowedUserIDs []int64) teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            userID := ctx.Update.Message.From.ID
            
            for _, allowedID := range allowedUserIDs {
                if userID == allowedID {
                    return next(ctx) // User is authorized
                }
            }
            
            return ctx.Reply("Access denied. You are not authorized to use this bot.")
        }
    }
}

// Apply to all handler types
bot.Use(AuthMiddleware([]int64{123456789, 987654321}))
```

### Type-Specific Middleware Examples

For cases where you need different behavior for different handler types:

### Recovery Middleware
This example shows a `RecoveryMiddleware` for `CommandMiddlewareFunc`. It catches panics, logs them, and returns an error to the user, preventing the bot from crashing.
```go
import (
	"fmt"
	"log"
	"runtime/debug"
	teleflow "github.com/kslamph/teleflow/core" // Adjust import path
)

func RecoveryCommandMiddleware() teleflow.CommandMiddlewareFunc {
	return func(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc {
		return func(ctx *teleflow.Context, command string, args string) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic recovered in RecoveryCommandMiddleware for command /%s: %v\n%s", command, r, debug.Stack())
					// It's often good practice to try and inform the user,
					// but be careful not to panic again while doing so.
					replyErr := ctx.Reply("Sorry, an internal error occurred while processing your command.")
					if replyErr != nil {
						log.Printf("Failed to send error reply after panic: %v", replyErr)
					}
					// Set a generic error to be returned
					err = fmt.Errorf("internal server error (panic recovery)")
				}
			}()
			return next(ctx, command, args)
		}
	}
}

// Usage:
// bot.UseCommandMiddleware(RecoveryCommandMiddleware())
```
You would create similar recovery middleware for `TextMiddlewareFunc`, `DefaultTextMiddlewareFunc`, and `CallbackMiddlewareFunc` by adapting the signature and logged information.

### Authentication Middleware
An `AuthMiddleware` can check permissions and short-circuit if unauthorized. Here's a conceptual example for `CommandMiddlewareFunc`:
```go
import (
	"errors"
	"log"
	teleflow "github.com/kslamph/teleflow/core" // Adjust import path
)

// isAdmin is a placeholder for your actual admin checking logic
func isAdmin(userID int64) bool {
	// Replace with your actual logic, e.g., check against a database
	return userID == 123456789 // Example admin UserID
}

func AuthCommandMiddleware() teleflow.CommandMiddlewareFunc {
	return func(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc {
		return func(ctx *teleflow.Context, command string, args string) error {
			// Example: Restrict "/admin" command
			if command == "admin" && !isAdmin(ctx.UserID()) {
				log.Printf("AuthCommandMiddleware: Unauthorized attempt to use /admin by UserID: %d", ctx.UserID())
				// Short-circuit and reply to the user
				return ctx.Reply("You are not authorized to use this command.")
			}

			// If authorized, or not a protected command, proceed to the next handler
			return next(ctx, command, args)
		}
	}
}

// Usage:
// bot.UseCommandMiddleware(AuthCommandMiddleware())
```
This `AuthCommandMiddleware` specifically checks command names. For text or callback middleware, the authentication logic might be based on `ctx.UserID()` or data within `ctx` set by a previous, more general authentication middleware (if you choose to layer them).

## Middleware for Flow Handlers

The current refactoring primarily focuses on type-specific middleware for global command, text, default text, and callback handlers, registered via `bot.UseCommandMiddleware()`, `bot.UseTextMiddleware()`, etc.

Teleflow also defines a `FlowStepInputMiddlewareFunc`:
```go
// From core/middleware_types.go
// type FlowStepInputHandlerFunc func(ctx *Context, input string) error
type FlowStepInputMiddlewareFunc func(next teleflow.FlowStepInputHandlerFunc) teleflow.FlowStepInputHandlerFunc
```
And a corresponding registration method on the `Bot` object:
```go
// From core/bot.go
bot.UseFlowStepInputMiddleware(m teleflow.FlowStepInputMiddlewareFunc)
```
This allows you to add middleware that specifically targets the input handling phase of flow steps.

Applying middleware directly to other individual flow lifecycle handlers (like `FlowStepStartHandlerFunc`, `FlowCompletionHandlerFunc`, `FlowCancellationHandlerFunc`) is a potential area for future enhancements and is not covered by the current `Use*Middleware` methods on the `Bot` object beyond `UseFlowStepInputMiddleware`.

## Best Practices for Middleware

- **Specificity**: Write middleware that is specific to the handler category it targets (command, text, callback, flow step input). This leverages the type safety benefits.
- **Call `next(...)`**: Always ensure `next(...)` is called within your middleware unless you are intentionally short-circuiting the request (e.g., an authentication failure). Forgetting to call `next` will silently stop processing for that update.
- **Error Handling**: Decide if your middleware should handle errors from `next(ctx, ...)` or propagate them. Recovery middleware typically handles errors, while others might just log and propagate.
- **Context Modification**: If middleware modifies the context (e.g., `ctx.Set()`), ensure keys are unique or well-namespaced to avoid collisions.
- **Order Matters**: Be mindful of the order in which you register middleware for a given category. For example, recovery middleware should usually be among the first, and authentication middleware before sensitive operations.
- **Performance**: Be cautious of computationally expensive operations in middleware, as they run for every applicable request in their category.

## Next Steps

- [Handlers Guide](handlers-guide.md): Learn about the specific handler functions that type-specific middleware ultimately wrap.
- [Flow Guide](flow-guide.md): Understand how conversational flows work and how `FlowStepInputMiddlewareFunc` can be used.
- [API Reference](api-reference.md): For detailed information on middleware types, handler function signatures, and other related types.
- Explore `core/middleware_types.go` and `core/bot.go` for the definitions of these types and functions.