package teleflow

import (
	"fmt"
	"sync"
)

// Package teleflow/core provides PromptKeyboardHandler for managing keyboard creation and UUID mapping.
//
// This file contains the PromptKeyboardHandler which serves as the central component for processing
// KeyboardFunc functions from PromptConfig, building inline keyboards, and managing the UUID-to-data
// mappings required for secure callback handling in conversational flows.
//
// # Core Responsibilities
//
// The PromptKeyboardHandler performs several critical functions:
//   - Processes KeyboardFunc from flow steps and prompts
//   - Manages per-user UUID mappings for callback data security
//   - Validates keyboards before sending to prevent API errors
//   - Provides thread-safe access to callback data mappings
//   - Handles cleanup of user mappings when flows complete
//
// # UUID Management
//
// To prevent callback data conflicts and ensure security, the handler:
//   - Generates unique UUIDs for each callback button
//   - Maps UUIDs to original callback data per user
//   - Provides secure retrieval of original data during callback processing
//   - Isolates user data to prevent cross-user data access
//
// # Usage in Flows
//
// The handler is automatically used by the prompt system when KeyboardFunc
// is provided in PromptConfig. It seamlessly integrates with the flow system
// to provide interactive keyboards without manual UUID management.

// PromptKeyboardHandler manages inline keyboard creation and UUID-to-callback-data mapping for prompts.
//
// This handler serves as the bridge between user-defined KeyboardFunc functions and the underlying
// Telegram bot API. It processes keyboard functions, builds the resulting inline keyboards,
// and maintains secure UUID mappings for callback data on a per-user basis.
//
// Key features:
//   - Thread-safe per-user UUID mapping management
//   - Automatic keyboard validation before building
//   - Integration with InlineKeyboardBuilder for UUID generation
//   - Secure isolation of user callback data
//   - Cleanup methods for memory management
//
// The handler is designed to work seamlessly with the Step-Prompt-Process API, where
// KeyboardFunc functions are defined in flow steps and automatically processed during
// prompt rendering.
//
// Example integration:
//
//	handler := NewPromptKeyboardHandler()
//
//	// Used internally by prompt system
//	keyboard, err := handler.BuildKeyboard(ctx, func(ctx *Context) *InlineKeyboardBuilder {
//		return NewInlineKeyboard().
//			ButtonCallback("Approve", "approve_123").
//			ButtonCallback("Reject", "reject_123")
//	})
//
//	// Later, during callback processing
//	originalData, found := handler.GetCallbackData(ctx.UserID(), callbackUUID)
type PromptKeyboardHandler struct {
	// userUUIDMappings stores UUID-to-data mappings per user ID
	// Key: UserID (int64), Value: map[uuid_string]original_data
	// This ensures each user's callback data is isolated and secure
	userUUIDMappings map[int64]map[string]interface{}

	// mu provides thread-safe access to userUUIDMappings
	// Uses RWMutex for efficient concurrent read access during callback processing
	mu sync.RWMutex
}

// NewPromptKeyboardHandler creates a new PromptKeyboardHandler instance.
//
// This constructor initializes an empty handler with no user mappings.
// The handler is ready to process KeyboardFunc functions and manage
// UUID mappings for callback data.
//
// Returns a new PromptKeyboardHandler ready for use with the prompt system.
//
// Example:
//
//	handler := NewPromptKeyboardHandler()
//	// Handler is now ready to process keyboards in prompts
func NewPromptKeyboardHandler() *PromptKeyboardHandler {
	return &PromptKeyboardHandler{
		userUUIDMappings: make(map[int64]map[string]interface{}),
	}
}

// BuildKeyboard processes a KeyboardFunc and builds a Telegram inline keyboard with UUID management.
//
// This method is the core of the keyboard handling system. It takes a KeyboardFunc from a
// PromptConfig, executes it to get an InlineKeyboardBuilder, validates the resulting keyboard,
// stores UUID mappings for callback data, and returns the final Telegram keyboard markup.
//
// The method handles several important scenarios:
//   - Nil KeyboardFunc (returns nil to indicate no keyboard)
//   - KeyboardFunc returning nil builder (returns nil to indicate no keyboard)
//   - Empty keyboards (returns nil after validation failure)
//   - Valid keyboards (processes UUID mappings and returns markup)
//
// Parameters:
//   - ctx: Current context containing user information and flow state
//   - keyboardFunc: Function that generates an InlineKeyboardBuilder (can be nil)
//
// Returns:
//   - interface{}: tgbotapi.InlineKeyboardMarkup for valid keyboards, nil for no keyboard
//   - error: Validation error if keyboard is invalid, nil otherwise
//
// Example usage within prompt system:
//
//	keyboardFunc := func(ctx *Context) *InlineKeyboardBuilder {
//		return NewInlineKeyboard().
//			ButtonCallback("Approve", ctx.GetData("requestId")).
//			ButtonCallback("Reject", "reject")
//	}
//
//	keyboard, err := handler.BuildKeyboard(ctx, keyboardFunc)
//	if err != nil {
//		// Handle validation error
//	}
//	if keyboard != nil {
//		// Use keyboard in message
//	}
func (pkh *PromptKeyboardHandler) BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
	if keyboardFunc == nil {
		return nil, nil // No keyboard specified - return nil to indicate no keyboard
	}

	builder := keyboardFunc(ctx) // Execute user's KeyboardFunc to get builder
	if builder == nil {
		return nil, nil // Function returned nil - return nil to indicate no keyboard
	}

	if err := builder.ValidateBuilder(); err != nil {
		return nil, fmt.Errorf("invalid inline keyboard: %w", err)
	}

	// Store UUID mappings from the builder instance in thread-safe manner
	pkh.mu.Lock()
	defer pkh.mu.Unlock()

	userID := ctx.UserID()
	if pkh.userUUIDMappings[userID] == nil {
		pkh.userUUIDMappings[userID] = make(map[string]interface{})
	}

	// Copy all UUID mappings from builder to handler's per-user storage
	for uuid, data := range builder.GetUUIDMapping() {
		pkh.userUUIDMappings[userID][uuid] = data
	}

	// Build the final Telegram keyboard markup
	builtKeyboard := builder.Build()
	if numButtons(builtKeyboard) == 0 { // Ensure we don't return an empty keyboard markup
		return nil, nil
	}

	return builtKeyboard, nil
}

// GetCallbackData retrieves the original callback data for a given user and UUID.
//
// This method provides secure access to the original callback data that was mapped
// to a UUID during keyboard creation. It ensures that users can only access their
// own callback data, preventing cross-user data leakage.
//
// The method is typically called during callback query processing to resolve
// the UUID-based callback data back to the original values provided by the user.
//
// Parameters:
//   - userID: The user ID whose callback data should be retrieved
//   - uuid: The UUID generated during button creation that maps to the original data
//
// Returns:
//   - interface{}: The original callback data if found
//   - bool: true if the data was found, false otherwise
//
// Example usage in callback processing:
//
//	// During callback query handling
//	callbackUUID := update.CallbackQuery.Data
//	originalData, found := handler.GetCallbackData(userID, callbackUUID)
//	if found {
//		// Process originalData based on its type
//		switch data := originalData.(type) {
//		case string:
//			// Handle string callback data
//		case map[string]interface{}:
//			// Handle structured callback data
//		}
//	}
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
//
// This cleanup method should be called when a user's flow ends, when their session
// expires, or when keyboards with stored mappings are no longer relevant. It prevents
// memory leaks by removing stale UUID mappings that are no longer needed.
//
// The method is thread-safe and can be called concurrently with other operations.
// It's recommended to call this method:
//   - When a conversational flow completes
//   - During user session cleanup
//   - When keyboards expire or become invalid
//   - As part of periodic memory management
//
// Parameters:
//   - userID: The user ID whose UUID mappings should be removed
//
// Example usage:
//
//	// At end of flow
//	defer handler.CleanupUserMappings(ctx.UserID())
//
//	// In session management
//	func (bot *Bot) cleanupExpiredSessions() {
//		for _, expiredUserID := range getExpiredUsers() {
//			bot.keyboardHandler.CleanupUserMappings(expiredUserID)
//		}
//	}
func (pkh *PromptKeyboardHandler) CleanupUserMappings(userID int64) {
	pkh.mu.Lock()
	defer pkh.mu.Unlock()
	delete(pkh.userUUIDMappings, userID)
}
