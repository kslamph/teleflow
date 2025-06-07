package teleflow

import (
	"fmt"
	"sync"
)

// PromptKeyboardHandler handles building inline keyboards from KeyboardFunc for prompts
// and manages the UUID to original callback data mappings.
type PromptKeyboardHandler struct {
	// Stores UUID mappings per user (or per message if keyboards are long-lived and tied to messages)
	// Key: UserID (int64), Value: map[string]interface{} (uuid -> originalData)
	userUUIDMappings map[int64]map[string]interface{}
	mu               sync.RWMutex // Protects userUUIDMappings
}

// NewPromptKeyboardHandler creates a new PromptKeyboardHandler.
func NewPromptKeyboardHandler() *PromptKeyboardHandler {
	return &PromptKeyboardHandler{
		userUUIDMappings: make(map[int64]map[string]interface{}),
	}
}

// BuildKeyboard processes a KeyboardFunc, generates the tgbotapi.InlineKeyboardMarkup,
// and stores the UUID mappings for callback data.
func (pkh *PromptKeyboardHandler) BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
	if keyboardFunc == nil {
		return nil, nil // No keyboard specified - return nil to indicate no keyboard
	}

	builder := keyboardFunc(ctx) // User's InlineKeyboardBuilder
	if builder == nil {
		return nil, nil // Function returned nil - return nil to indicate no keyboard
	}

	if err := builder.ValidateBuilder(); err != nil {
		return nil, fmt.Errorf("invalid inline keyboard: %w", err)
	}

	// Store UUID mappings from the user's builder instance
	pkh.mu.Lock()
	defer pkh.mu.Unlock()

	userID := ctx.UserID()
	if pkh.userUUIDMappings[userID] == nil {
		pkh.userUUIDMappings[userID] = make(map[string]interface{})
	}
	for uuid, data := range builder.GetUUIDMapping() {
		pkh.userUUIDMappings[userID][uuid] = data
	}

	// Return the tgbotapi.InlineKeyboardMarkup built by the builder
	// The builder.Build() method itself returns tgbotapi.InlineKeyboardMarkup
	builtKeyboard := builder.Build()
	if numButtons(builtKeyboard) == 0 { // Ensure we don't return an empty keyboard markup
		return nil, nil
	}

	return builtKeyboard, nil
}

// GetCallbackData retrieves the original callback data for a given user and UUID.
func (pkh *PromptKeyboardHandler) GetCallbackData(userID int64, uuid string) (interface{}, bool) {
	pkh.mu.RLock()
	defer pkh.mu.RUnlock()

	if userMappings, exists := pkh.userUUIDMappings[userID]; exists {
		data, found := userMappings[uuid]
		return data, found
	}
	return nil, false
}

// CleanupUserMappings removes all UUID mappings for a specific user.
// This should be called when a flow ends or when messages with keyboards are no longer relevant.
func (pkh *PromptKeyboardHandler) CleanupUserMappings(userID int64) {
	pkh.mu.Lock()
	defer pkh.mu.Unlock()
	delete(pkh.userUUIDMappings, userID)
}
