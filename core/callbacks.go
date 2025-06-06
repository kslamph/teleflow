package teleflow

import (
	"strings"
	"sync"
)

// Callback system provides internal handling of inline keyboard button
// interactions for the Step-Prompt-Process API. This system is used
// internally by the flow system and is not exposed to end users.
//
// With the API, callback complexity is completely abstracted away.
// Users define keyboards using simple maps in Prompt() functions and handle
// all input (both text and button clicks) through unified ProcessFunc functions.
//
// Internal responsibilities:
//   - Automatic callback registration for flow-generated keyboards
//   - Pattern matching and routing of callback data
//   - Integration with the unified input processing system
//   - Transparent handling of button interactions

// callbackHandler defines the interface for handling callback queries (internal use only).
// This interface is used internally by the flow system and should not be used directly by users.
type callbackHandler interface {
	// pattern returns the string pattern that this handler should match.
	// The pattern can be an exact string or end with a "*" to act as a wildcard prefix.
	pattern() string
	// handle processes the callback query.
	handle(ctx *Context, fullCallbackData string, extractedData string) error
}

// callbackRegistry manages type-safe callback handlers (internal use only)
type callbackRegistry struct {
	mu       sync.RWMutex
	handlers map[string]callbackHandler
	patterns []string
}

// newCallbackRegistry creates a new callback registry (internal use only)
func newCallbackRegistry() *callbackRegistry {
	return &callbackRegistry{
		handlers: make(map[string]callbackHandler),
		patterns: []string{},
	}
}

// handle finds and executes the appropriate callback handler.
// It returns a generic HandlerFunc which wraps the call to the specific CallbackHandler.handle method.
func (r *callbackRegistry) handle(callbackData string) HandlerFunc { // Keep existing signature for now, apply middleware internally
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
				return func(ctx *Context) error {
					dataForHandler := extractedData
					if isExactMatch && extractedData == "" {
						dataForHandler = callbackData
					}
					return specificHandler.handle(ctx, callbackData, dataForHandler)
				}
			}
		}
	}
	return nil // No handler found
}

// matchPattern checks if callback data matches pattern and extracts data
func (r *callbackRegistry) matchPattern(pattern, callbackData string) string {
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
