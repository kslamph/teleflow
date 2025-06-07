package teleflow

import (
	"fmt"
)

// inlineKeyboardBuilder handles building keyboards from KeyboardFunc
type inlineKeyboardBuilder struct {
	uuidMappings map[string]map[string]interface{} // Maps message IDs to UUID mappings
}

// newInlineKeyboardBuilder creates a new keyboard builder
func newInlineKeyboardBuilder() *inlineKeyboardBuilder {
	return &inlineKeyboardBuilder{
		uuidMappings: make(map[string]map[string]interface{}),
	}
}

// buildInlineKeyboard processes a KeyboardFunc and returns the appropriate keyboard structure
func (kb *inlineKeyboardBuilder) buildInlineKeyboard(keyboardFunc KeyboardFunc, ctx *Context) (interface{}, error) {
	if keyboardFunc == nil {
		return nil, nil // No keyboard specified
	}

	// Execute the keyboard function to get the builder
	builder := keyboardFunc(ctx)
	if builder == nil {
		return nil, nil // Function returned nil, no keyboard
	}

	// Validate the builder
	if err := builder.ValidateBuilder(); err != nil {
		return nil, fmt.Errorf("invalid keyboard: %v", err)
	}

	// Store UUID mappings for this message (we'll need message ID from context)
	messageKey := fmt.Sprintf("user_%d", ctx.UserID()) // Simple key for now
	kb.uuidMappings[messageKey] = builder.GetUUIDMapping()

	// Build and return the keyboard markup
	return builder.Build(), nil
}

// getCallbackData retrieves the original callback data from UUID
func (kb *inlineKeyboardBuilder) getCallbackData(userID int64, uuid string) (interface{}, bool) {
	messageKey := fmt.Sprintf("user_%d", userID)
	if mapping, exists := kb.uuidMappings[messageKey]; exists {
		if data, found := mapping[uuid]; found {
			return data, true
		}
	}
	return nil, false
}

// cleanupMappings removes UUID mappings for a user (called on flow end or callback handling)
func (kb *inlineKeyboardBuilder) cleanupMappings(userID int64) {
	messageKey := fmt.Sprintf("user_%d", userID)
	delete(kb.uuidMappings, messageKey)
}
