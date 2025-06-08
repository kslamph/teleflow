package teleflow

import (
	"sync"
)

// Package state provides persistent storage management for user data across
// bot interactions and conversation flows. It includes an in-memory implementation
// suitable for development and small-scale deployments, with support for custom
// storage backends for production environments.
//
// The state management system handles two types of data:
//   - User state: persistent data that survives across different conversations
//   - Flow state: temporary data that exists only during active flow execution
//
// State features:
//   - User-specific data persistence across bot restarts (with custom backends)
//   - Thread-safe concurrent access from multiple users
//   - Simple key-value storage interface
//   - Extensible design for custom storage backends
//
// Basic Usage:
//
//	// Create a state manager (usually handled by bot initialization)
//	stateManager := teleflow.NewInMemoryStateManager()
//
//	// Store user data (typically done through Context in flows)
//	err := stateManager.SetState(userID, "username", "john_doe")
//
//	// Retrieve user data
//	value, exists := stateManager.GetState(userID, "username")
//	if exists {
//		username := value.(string)
//	}
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
//		return nil
//	}
//
//	func (r *RedisStateManager) GetState(userID int64, key string) (interface{}, bool) {
//		// Implementation for Redis retrieval
//		return nil, false
//	}
//
//	func (r *RedisStateManager) ClearState(userID int64) error {
//		// Implementation for Redis cleanup
//		return nil
//	}
//
//	// Use custom backend
//	bot.SetStateManager(&RedisStateManager{client: redisClient})

// StateManager defines the interface for managing persistent user state data.
// It provides methods for storing, retrieving, and clearing user-specific data
// that persists across bot interactions and conversation flows.
//
// The interface is designed to be storage-agnostic, allowing implementations
// for various backends including in-memory (for development), Redis, databases,
// or cloud storage solutions.
//
// All methods must be thread-safe to handle concurrent access from multiple users.
type StateManager interface {
	// SetState stores a key-value pair for a specific user.
	// The value can be any type that the storage backend supports.
	//
	// Parameters:
	//   - userID: Telegram user ID to associate the data with
	//   - key: string identifier for the stored value
	//   - value: data to store (any type supported by the backend)
	//
	// Returns:
	//   - error: nil on success, or an error if storage fails
	SetState(userID int64, key string, value interface{}) error

	// GetState retrieves a stored value for a specific user and key.
	//
	// Parameters:
	//   - userID: Telegram user ID to retrieve data for
	//   - key: string identifier for the value to retrieve
	//
	// Returns:
	//   - interface{}: the stored value (nil if not found)
	//   - bool: true if the key exists, false otherwise
	GetState(userID int64, key string) (interface{}, bool)

	// ClearState removes all stored data for a specific user.
	// This is useful for cleanup when a user leaves or resets their data.
	//
	// Parameters:
	//   - userID: Telegram user ID to clear data for
	//
	// Returns:
	//   - error: nil on success, or an error if cleanup fails
	ClearState(userID int64) error
}

// inMemoryStateManager is a thread-safe in-memory implementation of the StateManager interface.
// It stores user state data in memory using nested maps, making it suitable for development,
// testing, and small-scale bot deployments.
//
// Key characteristics:
//   - Thread-safe: uses RWMutex for concurrent access protection
//   - Memory-based: data is lost when the bot process restarts
//   - Fast access: O(1) average case for get/set operations
//   - No persistence: not suitable for production environments requiring data durability
//   - Scalability: memory usage grows with number of users and stored data
//
// Best suited for:
//   - Development and testing environments
//   - Small bots with limited users
//   - Temporary data that doesn't need persistence
//   - Prototyping before implementing a persistent backend
type inMemoryStateManager struct {
	// mu provides thread-safe access to the data map, protecting against
	// concurrent read/write operations from multiple goroutines handling
	// different user requests simultaneously
	mu sync.RWMutex

	// data stores user state in a nested map structure:
	// - First level key: Telegram user ID (int64)
	// - Second level key: state key name (string)
	// - Value: stored data (interface{} to support any type)
	data map[int64]map[string]interface{}
}

// NewInMemoryStateManager creates and returns a new in-memory StateManager implementation.
// This constructor initializes the internal data structures needed for state storage.
//
// The returned StateManager is immediately ready for use and is thread-safe for
// concurrent access from multiple goroutines.
//
// Returns:
//   - StateManager: A new in-memory state manager instance
//
// Example:
//
//	stateManager := teleflow.NewInMemoryStateManager()
//	bot.SetStateManager(stateManager)
func NewInMemoryStateManager() StateManager {
	return &inMemoryStateManager{
		data: make(map[int64]map[string]interface{}),
	}
}

// SetState stores a key-value pair for a specific user in the in-memory storage.
// This method implements the StateManager interface and is fully thread-safe,
// handling concurrent access from multiple goroutines processing different users.
//
// The method automatically initializes the user's state map if this is the first
// time data is being stored for the user. Any existing value for the same key
// will be overwritten.
//
// Parameters:
//   - userID: Telegram user ID to associate the data with
//   - key: string identifier for the stored value
//   - value: data to store (any type that can be stored in interface{})
//
// Returns:
//   - error: always returns nil for the in-memory implementation
//
// Thread Safety:
// Uses a write lock to ensure atomic updates and prevent data races.
func (m *inMemoryStateManager) SetState(userID int64, key string, value interface{}) error {
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

// GetState retrieves a stored value for a specific user and key from in-memory storage.
// This method implements the StateManager interface and is fully thread-safe,
// using a read lock to allow concurrent read operations while preventing conflicts
// with write operations.
//
// Parameters:
//   - userID: Telegram user ID to retrieve data for
//   - key: string identifier for the value to retrieve
//
// Returns:
//   - interface{}: the stored value, or nil if not found
//   - bool: true if the key exists for the user, false otherwise
//
// Behavior:
//   - Returns (nil, false) if the user has never stored any data
//   - Returns (nil, false) if the user exists but the specific key doesn't
//   - Returns (value, true) if the key exists and has a value (even if nil)
//
// Thread Safety:
// Uses a read lock to allow multiple concurrent read operations.
func (m *inMemoryStateManager) GetState(userID int64, key string) (interface{}, bool) {
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

// ClearState removes all stored data for a specific user from in-memory storage.
// This method implements the StateManager interface and is fully thread-safe,
// handling concurrent access from multiple goroutines.
//
// After calling this method, the user will have no stored state, as if they
// had never interacted with the bot before. This is useful for:
//   - User-requested data deletion
//   - Cleanup when users leave or block the bot
//   - Resetting user state for testing
//   - Memory management for inactive users
//
// Parameters:
//   - userID: Telegram user ID to clear all data for
//
// Returns:
//   - error: always returns nil for the in-memory implementation
//
// Behavior:
//   - If the user has no stored data, the operation is a no-op
//   - All keys and values for the user are permanently removed
//   - Subsequent GetState calls for this user will return (nil, false)
//
// Thread Safety:
// Uses a write lock to ensure atomic deletion and prevent data races.
func (m *inMemoryStateManager) ClearState(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove all state data for the user
	delete(m.data, userID)

	return nil
}
