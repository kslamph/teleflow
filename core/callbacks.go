package teleflow

import (
	"strings"
	"sync"
)

type callbackHandler interface {
	pattern() string

	handle(ctx *Context, fullCallbackData string, extractedData string) error
}

type callbackRegistry struct {
	mu sync.RWMutex

	handlers map[string]callbackHandler

	patterns []string
}

func newCallbackRegistry() *callbackRegistry {
	return &callbackRegistry{
		handlers: make(map[string]callbackHandler),
		patterns: []string{},
	}
}

func (r *callbackRegistry) handle(callbackData string) HandlerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, pattern := range r.patterns {
		if specificHandler := r.handlers[pattern]; specificHandler != nil {
			extractedData := r.matchPattern(pattern, callbackData)

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
	return nil
}

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
