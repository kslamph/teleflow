package teleflow

import (
	"fmt"
	"strings"
	"sync"
)

// Callback system provides type-safe handling of inline keyboard button
// interactions with pattern matching, data extraction, and structured
// callback management. The system supports both simple callback handlers
// and complex pattern-based routing with wildcard matching.
//
// Callbacks are triggered when users interact with inline keyboard buttons
// that contain callback data. The system can route callbacks based on exact
// matches or patterns, extract data from callback strings, and provide
// type-safe access to callback information.
//
// Basic Callback Handling:
//
//	// Register callback handler
//	bot.HandleCallback("approve_user", func(ctx *teleflow.Context) error {
//		userID := ctx.CallbackData() // Gets "approve_user"
//		return ctx.EditMessage("User approved!")
//	})
//
// Pattern-Based Callbacks:
//
//	// Handle callbacks with wildcards and data extraction
//	bot.HandleCallback("user_action_*", func(ctx *teleflow.Context) error {
//		parts := strings.Split(ctx.CallbackData(), "_")
//		action := parts[2] // Extract the action part
//		return ctx.EditMessage("Action: " + action)
//	})
//
// Advanced Pattern Matching:
//
//	// Complex patterns with multiple wildcards
//	bot.HandleCallback("admin_*_user_*", func(ctx *teleflow.Context) error {
//		data := ctx.CallbackData()
//		// Parse and handle admin actions on specific users
//		return handleAdminAction(ctx, data)
//	})
//
// The callback system integrates seamlessly with inline keyboards and
// provides automatic answer handling for callback queries to prevent
// loading states in the Telegram client.

// CallbackHandler defines the interface for handling callback queries.
// Implementations of this interface can be registered with the CallbackRegistry
// to process specific callback patterns.
type CallbackHandler interface {
	// Pattern returns the string pattern that this handler should match.
	// The pattern can be an exact string or end with a "*" to act as a wildcard prefix.
	// For example, "action_confirm" or "user_*".
	Pattern() string
	// Handle processes the callback query.
	//
	// Parameters:
	//   - ctx: The context for the current update.
	//   - fullCallbackData: The complete data string from the callback query.
	//   - extractedData: If the pattern used a wildcard, this is the part of the
	//     callback data that matched the wildcard. If it was an exact match,
	//     this will typically be the fullCallbackData or an empty string depending on registry logic.
	//
	// Returns:
	//   - error: An error if processing failed, nil otherwise.
	Handle(ctx *Context, fullCallbackData string, extractedData string) error
}

// CallbackRegistry manages type-safe callback handlers
type CallbackRegistry struct {
	mu                 sync.RWMutex
	handlers           map[string]CallbackHandler
	patterns           []string
	callbackMiddleware []CallbackMiddlewareFunc // Added to store callback-specific middleware
}

// NewCallbackRegistry creates a new callback registry
func NewCallbackRegistry(middlewares ...CallbackMiddlewareFunc) *CallbackRegistry {
	return &CallbackRegistry{
		handlers:           make(map[string]CallbackHandler),
		patterns:           []string{},
		callbackMiddleware: middlewares, // Initialize with provided middlewares
	}
}

// Register registers a callback handler
func (r *CallbackRegistry) Register(handler CallbackHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pattern := handler.Pattern()
	r.handlers[pattern] = handler
	r.patterns = append(r.patterns, pattern)
}

// Handle finds and executes the appropriate callback handler.
// It returns a generic HandlerFunc which wraps the call to the specific CallbackHandler.Handle method.
func (r *CallbackRegistry) Handle(callbackData string) HandlerFunc { // Keep existing signature for now, apply middleware internally
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, pattern := range r.patterns {
		if specificHandler := r.handlers[pattern]; specificHandler != nil {
			extractedData := r.matchPattern(pattern, callbackData)
			// A match occurs if:
			// 1. The pattern is an exact match to callbackData (extractedData will be "" from matchPattern, but it's a match).
			// 2. The pattern has a wildcard and callbackData matches the prefix (extractedData will be the part after the prefix).
			isExactMatch := (pattern == callbackData)
			isPrefixMatchWithWildcard := (strings.HasSuffix(pattern, "*") && extractedData != "" && strings.HasPrefix(callbackData, pattern[:len(pattern)-1]))

			if isExactMatch || isPrefixMatchWithWildcard {
				// Create the core CallbackHandlerFunc
				coreHandlerFunc := func(ctx *Context, cbd string, ed string) error {
					return specificHandler.Handle(ctx, cbd, ed)
				}

				// Apply middleware to this coreHandlerFunc
				wrappedHandler := coreHandlerFunc
				for i := len(r.callbackMiddleware) - 1; i >= 0; i-- {
					wrappedHandler = r.callbackMiddleware[i](wrappedHandler)
				}

				return func(ctx *Context) error {
					dataForHandler := extractedData
					if isExactMatch && extractedData == "" {
						dataForHandler = callbackData
					}
					return wrappedHandler(ctx, callbackData, dataForHandler)
				}
			}
		}
	}
	return nil // No handler found
}

// matchPattern checks if callback data matches pattern and extracts data
func (r *CallbackRegistry) matchPattern(pattern, callbackData string) string {
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		if strings.HasPrefix(callbackData, prefix) {
			return callbackData[len(prefix):]
		}
	} else if pattern == callbackData {
		return ""
	}
	return ""
}

// SimpleCallback is a helper function to easily create a CallbackHandler
// from a pattern string and a handler function.
// This is useful for straightforward callback scenarios where a full struct implementation
// of CallbackHandler is not necessary.
//
// Parameters:
//   - pattern: The string pattern to match for this callback. See CallbackHandler.Pattern().
//   - handler: A function that implements the logic for this callback. Its signature
//     matches the CallbackHandler.Handle method.
//
// Returns:
//   - CallbackHandler: An implementation of CallbackHandler that wraps the provided pattern and handler function.
func SimpleCallback(pattern string, handler func(ctx *Context, fullCallbackData string, extractedData string) error) CallbackHandler {
	return &simpleCallbackHandler{
		pattern: pattern,
		handler: handler,
	}
}

type simpleCallbackHandler struct {
	pattern string
	// Updated handler signature
	handler func(ctx *Context, fullCallbackData string, extractedData string) error
}

func (h *simpleCallbackHandler) Pattern() string {
	return h.pattern
}

// Updated Handle method signature to match the interface
func (h *simpleCallbackHandler) Handle(ctx *Context, fullCallbackData string, extractedData string) error {
	return h.handler(ctx, fullCallbackData, extractedData)
}

// Typed callback helpers for common patterns
type ActionCallback struct {
	Action string
	// Handler for ActionCallback should also be updated if it's intended to be used
	// with the new CallbackHandler interface directly.
	// For now, its direct usage might be broken until it's adapted or
	// a different way of handling such specific typed callbacks is devised.
	// This example focuses on the core CallbackHandler interface change.
	// Assuming it needs to match the new structure for consistency if registered directly:
	Handler func(ctx *Context, fullCallbackData string, extractedData string) error
}

func (ac *ActionCallback) Pattern() string {
	return fmt.Sprintf("action_%s_*", ac.Action)
}

// Updated Handle method signature to match the interface
func (ac *ActionCallback) Handle(ctx *Context, fullCallbackData string, extractedData string) error {
	// The `data` here was the extracted part.
	// The `actionData` in the original Handler func(ctx *Context, actionData string) error
	// corresponded to this extracted part.
	return ac.Handler(ctx, fullCallbackData, extractedData)
}
