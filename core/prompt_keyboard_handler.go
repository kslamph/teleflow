package teleflow

import (
	"fmt"
	"sync"
)

// PromptKeyboardActions defines the actions that can be performed by a prompt keyboard handler.
// It's used to decouple components like Bot from the concrete PromptKeyboardHandler implementation.
type PromptKeyboardActions interface {
	// BuildKeyboard constructs a keyboard based on the provided KeyboardFunc.
	// It registers callback UUIDs and their associated data for the user.
	BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error)

	// GetCallbackData retrieves the data associated with a specific callback UUID for a user.
	// It returns the data and a boolean indicating if the UUID was found.
	GetCallbackData(userID int64, uuid string) (interface{}, bool)

	// CleanupUserMappings removes all callback UUID mappings for a given user.
	// This is typically called when a user's session or flow ends.
	CleanupUserMappings(userID int64)
}

type PromptKeyboardHandler struct {
	userUUIDMappings map[int64]map[string]interface{}

	mu sync.RWMutex
}

func newPromptKeyboardHandler() *PromptKeyboardHandler {
	return &PromptKeyboardHandler{
		userUUIDMappings: make(map[int64]map[string]interface{}),
	}
}

func (pkh *PromptKeyboardHandler) BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
	if keyboardFunc == nil {
		return nil, nil
	}

	builder := keyboardFunc(ctx)
	if builder == nil {
		return nil, nil
	}

	if err := builder.validateBuilder(); err != nil {
		return nil, fmt.Errorf("invalid inline keyboard: %w", err)
	}

	pkh.mu.Lock()
	defer pkh.mu.Unlock()

	userID := ctx.UserID()
	if pkh.userUUIDMappings[userID] == nil {
		pkh.userUUIDMappings[userID] = make(map[string]interface{})
	}

	for uuid, data := range builder.uuidMapping {
		pkh.userUUIDMappings[userID][uuid] = data
	}

	builtKeyboard := builder.Build()
	if numButtons(builtKeyboard) == 0 {
		return nil, nil
	}

	return builtKeyboard, nil
}

func (pkh *PromptKeyboardHandler) GetCallbackData(userID int64, uuid string) (interface{}, bool) {
	pkh.mu.RLock()
	defer pkh.mu.RUnlock()

	if userMappings, exists := pkh.userUUIDMappings[userID]; exists {
		data, found := userMappings[uuid]
		return data, found
	}
	return nil, false
}

func (pkh *PromptKeyboardHandler) CleanupUserMappings(userID int64) {
	pkh.mu.Lock()
	defer pkh.mu.Unlock()
	delete(pkh.userUUIDMappings, userID)
}
