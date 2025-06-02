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

// CallbackHandler interface for type-safe callback handling
type CallbackHandler interface {
	Pattern() string
	Handle(ctx *Context, data string) error
}

// CallbackRegistry manages type-safe callback handlers
type CallbackRegistry struct {
	mu       sync.RWMutex
	handlers map[string]CallbackHandler
	patterns []string
}

// NewCallbackRegistry creates a new callback registry
func NewCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry{
		handlers: make(map[string]CallbackHandler),
		patterns: []string{},
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

// Handle finds and executes the appropriate callback handler
func (r *CallbackRegistry) Handle(callbackData string) HandlerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, pattern := range r.patterns {
		if handler := r.handlers[pattern]; handler != nil {
			if data := r.matchPattern(pattern, callbackData); data != "" || (pattern == callbackData) {
				return func(ctx *Context) error {
					return handler.Handle(ctx, data)
				}
			}
		}
	}
	return nil
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

// Helper function to create simple callback handlers
func SimpleCallback(pattern string, handler func(ctx *Context, data string) error) CallbackHandler {
	return &simpleCallbackHandler{
		pattern: pattern,
		handler: handler,
	}
}

type simpleCallbackHandler struct {
	pattern string
	handler func(ctx *Context, data string) error
}

func (h *simpleCallbackHandler) Pattern() string {
	return h.pattern
}

func (h *simpleCallbackHandler) Handle(ctx *Context, data string) error {
	return h.handler(ctx, data)
}

// Typed callback helpers for common patterns
type ActionCallback struct {
	Action  string
	Handler func(ctx *Context, actionData string) error
}

func (ac *ActionCallback) Pattern() string {
	return fmt.Sprintf("action_%s_*", ac.Action)
}

func (ac *ActionCallback) Handle(ctx *Context, data string) error {
	return ac.Handler(ctx, data)
}
