package teleflow

import (
	"fmt"
)

// messageRenderer handles rendering of MessageSpec into string messages
type messageRenderer struct{}

// newMessageRenderer creates a new message renderer
func newMessageRenderer() *messageRenderer {
	return &messageRenderer{}
}

// renderMessage processes a MessageSpec and returns the final message string
func (mr *messageRenderer) renderMessage(config *PromptConfig, ctx *Context) (string, error) {
	if config.Message == nil {
		return "", nil // No message specified
	}

	switch msg := config.Message.(type) {
	case string:
		// Static string message
		return msg, nil

	case func(*Context) string:
		// Dynamic message function
		result := msg(ctx)
		return result, nil

	default:
		return "", fmt.Errorf("unsupported message type: %T (expected string or func(*Context) string)", msg)
	}
}
