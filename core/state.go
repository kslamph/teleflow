package teleflow

import (
	"sync"
)

// State management system provides persistent storage of user data across
// bot interactions, conversations, and sessions. The system supports both
// in-memory storage for development and testing, and can be extended with
// custom storage backends for production deployments.
//
// The state system enables:
//   - User-specific data persistence across conversations
//   - Conversation flow state tracking
//   - Temporary data storage for multi-step interactions
//   - Session management and user preferences
//   - Custom storage backend integration
//
// Basic State Operations:
//
//	// Store user data
//	ctx.SetState("username", "john_doe")
//	ctx.SetState("registration_step", 2)
//
//	// Retrieve user data
//	if username, exists := ctx.GetState("username"); exists {
//		ctx.Reply("Hello " + username.(string))
//	}
//
//	// Clear user state
//	ctx.ClearState()
//
// Flow Integration:
//
//	// State is automatically managed during flows
//	flow := teleflow.NewFlow("user_onboarding").
//		AddStep("name", teleflow.StepTypeText, "What's your name?").
//		AddStep("email", teleflow.StepTypeText, "What's your email?")
//
//	// Flow progress is tracked in state automatically
//	bot.RegisterFlow(flow, func(ctx *teleflow.Context, result map[string]string) error {
//		// Save final results to persistent storage
//		return saveUserProfile(ctx.UserID(), result)
//	})
//
// Custom Storage Backend:
//
//	// Implement StateManager interface for custom storage
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
