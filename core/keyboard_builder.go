package teleflow

import (
	"fmt"
	"sync"
)

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

	if err := builder.ValidateBuilder(); err != nil {
		return nil, fmt.Errorf("invalid inline keyboard: %w", err)
	}

	pkh.mu.Lock()
	defer pkh.mu.Unlock()

	userID := ctx.UserID()
	if pkh.userUUIDMappings[userID] == nil {
		pkh.userUUIDMappings[userID] = make(map[string]interface{})
	}

	for uuid, data := range builder.GetUUIDMapping() {
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
