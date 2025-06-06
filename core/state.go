package teleflow

import (
	"sync"
)

// State management system provides persistent storage of user data across
// bot interactions and conversation flows. The system includes an in-memory
// implementation for development and can be extended with custom storage
// backends for production use.
//
// State features:
//   - User-specific data persistence across conversations
//   - Automatic flow state tracking
//   - Simple key-value storage interface
//   - Custom storage backend support
//
// Basic Usage (handled automatically in flows):
//
//	// In Process functions, use context methods
//	ctx.Set("username", "john_doe")  // Store in current context
//
//	// Access previous step data automatically
//	name, exists := ctx.Get("name")  // Get from current flow context
//
// Flow State Management:
//
//	// Flow progress and data are tracked automatically
//	.Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//		ctx.Set("user_choice", input)  // Stored for this flow
//		return teleflow.NextStep()
//	})
//
// Custom Storage Backend:
//
//	// Implement StateManager interface for production storage
//	type RedisStateManager struct {
//		client *redis.Client
//	}
//
//	func (r *RedisStateManager) SetState(userID int64, key string, value interface{}) error {
//		// Implementation for Redis storage
//	}
//
//	bot.SetStateManager(&RedisStateManager{client: redisClient})

// StateManager defines the interface for managing user state data
type StateManager interface {
	SetState(userID int64, key string, value interface{}) error
	GetState(userID int64, key string) (interface{}, bool)
	ClearState(userID int64) error
}

// InMemoryStateManager is a thread-safe in-memory implementation of StateManager
// that persists data across handler calls within the same bot session
type InMemoryStateManager struct {
	mu   sync.RWMutex
	data map[int64]map[string]interface{} // userID -> key -> value mapping
}

// NewInMemoryStateManager creates and returns a new StateManager implementation
func NewInMemoryStateManager() StateManager {
	return &InMemoryStateManager{
		data: make(map[int64]map[string]interface{}),
	}
}

// SetState stores a key-value pair for a specific user
// This method is thread-safe and handles concurrent access from multiple users
func (m *InMemoryStateManager) SetState(userID int64, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize user's state map if it doesn't exist
	if m.data[userID] == nil {
		m.data[userID] = make(map[string]interface{})
	}

	// Set the key-value pair for the user
	m.data[userID][key] = value

	return nil
}

// GetState retrieves a value for a specific user and key
// Returns the value and a boolean indicating if the key was found
func (m *InMemoryStateManager) GetState(userID int64, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if user has any state data
	userData, userExists := m.data[userID]
	if !userExists {
		return nil, false
	}

	// Check if the specific key exists for the user
	value, keyExists := userData[key]
	return value, keyExists
}

// ClearState removes all state data for a specific user
// This method is thread-safe and handles concurrent access
func (m *InMemoryStateManager) ClearState(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove all state data for the user
	delete(m.data, userID)

	return nil
}
