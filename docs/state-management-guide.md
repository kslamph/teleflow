# Teleflow State Management Guide

State management in Teleflow allows your bot to remember information about users across different interactions, messages, and even conversational flows. This is crucial for creating personalized experiences and managing multi-step processes.

## Table of Contents

- [What is State Management?](#what-is-state-management)
- [The `StateManager` Interface](#the-statemanager-interface)
- [Built-in State Manager: `InMemoryStateManager`](#built-in-state-manager-inmemorystatemanager)
  - [Characteristics](#characteristics)
  - [Usage](#usage)
- [Using State in `teleflow.Context`](#using-state-in-teleflowcontext)
  - [`ctx.SetState(key string, value interface{}) error`](#ctxsetstatekey-string-value-interface-error)
  - [`ctx.GetState(key string) (interface{}, bool)`](#ctxgetstatekey-string-interface-bool)
  - [`ctx.ClearState() error`](#ctxclearstate-error)
- [State Management in Flows](#state-management-in-flows)
- [Custom State Managers](#custom-state-managers)
  - [Implementing the Interface](#implementing-the-interface)
  - [Example: Conceptual Redis State Manager](#example-conceptual-redis-state-manager)
  - [Setting a Custom State Manager](#setting-a-custom-state-manager)
- [Request-Scoped Data vs. Persistent State](#request-scoped-data-vs-persistent-state)
  - [`ctx.Set()` / `ctx.Get()`](#ctxset--ctxget)
  - [`ctx.SetState()` / `ctx.GetState()`](#ctxsetstate--ctxgetstate)
- [Best Practices for State Management](#best-practices-for-state-management)
- [Next Steps](#next-steps)

## What is State Management?
State refers to any data that your bot needs to remember about a user or a chat session. This could include:
- User preferences (e.g., language, notification settings).
- Progress in a multi-step task (e.g., current step in a registration flow).
- Temporary data needed for a short interaction (e.g., an item ID selected from a list).
- User authentication status.

Teleflow provides a flexible system for managing this state.

## The `StateManager` Interface
At the core of Teleflow's state management is the `StateManager` interface (defined in `core/state.go`):
```go
type StateManager interface {
    SetState(userID int64, key string, value interface{}) error
    GetState(userID int64, key string) (interface{}, bool)
    ClearState(userID int64) error
}
```
This interface defines the contract for how state is stored, retrieved, and cleared on a per-user basis.

## Built-in State Manager: `InMemoryStateManager`
Teleflow comes with a default, thread-safe, in-memory implementation of `StateManager` called `InMemoryStateManager`.

### Characteristics
- **In-Memory**: Data is stored in the bot's RAM.
- **Session-Based**: State persists as long as the bot process is running. If the bot restarts, all in-memory state is lost.
- **User-Scoped**: State is stored per `userID`.
- **Suitable For**:
    - Development and testing.
    - Bots where persistent state across restarts is not critical.
    - Managing temporary state within active user sessions or flows.

### Usage
By default, `teleflow.NewBot()` initializes an `InMemoryStateManager`. You don't need to do anything special to use it for basic state operations via the `Context`.

## Using State in `teleflow.Context`
The `teleflow.Context` object, available in all handlers, provides convenient methods to interact with the bot's configured `StateManager`. These methods operate on the state of the user associated with the current update.

### `ctx.SetState(key string, value interface{}) error`
Stores a key-value pair for the current user.
```go
bot.HandleCommand("setlang", func(ctx *teleflow.Context) error {
    lang := ctx.Update.Message.CommandArguments()
    if lang == "" {
        return ctx.Reply("Please specify a language (e.g., /setlang en)")
    }
    err := ctx.SetState("user_language", lang)
    if err != nil {
        log.Printf("Error setting state for user %d: %v", ctx.UserID(), err)
        return ctx.Reply("Sorry, couldn't save your language preference.")
    }
    return ctx.Reply("Language preference saved: " + lang)
})
```

### `ctx.GetState(key string) (interface{}, bool)`
Retrieves a value for the current user by its key. It returns the value and a boolean indicating if the key was found.
```go
bot.HandleCommand("greet", func(ctx *teleflow.Context) error {
    lang, exists := ctx.GetState("user_language")
    greeting := "Hello!" // Default greeting

    if exists {
        userLang := lang.(string) // Type assertion might be needed
        if userLang == "es" {
            greeting = "¡Hola!"
        } else if userLang == "fr" {
            greeting = "Bonjour!"
        }
    }
    return ctx.Reply(greeting + " " + ctx.Update.Message.From.FirstName)
})
```

### `ctx.ClearState() error`
Removes all state data for the current user.
```go
bot.HandleCommand("resetprefs", func(ctx *teleflow.Context) error {
    err := ctx.ClearState()
    if err != nil {
        log.Printf("Error clearing state for user %d: %v", ctx.UserID(), err)
        return ctx.Reply("Sorry, couldn't reset your preferences.")
    }
    return ctx.Reply("Your preferences have been reset.")
})
```

## State Management in Flows
The [Conversational Flow System](flow-guide.md) heavily relies on state management.
- The `FlowManager` uses its own instance of `StateManager` (by default, an `InMemoryStateManager`) to store `UserFlowState`.
- `UserFlowState` includes a `Data map[string]interface{}` field. When you use `ctx.Set("key", value)` *within a flow step handler*, this data is stored in the `UserFlowState.Data` for the current flow instance.
- This flow-specific data is automatically managed (loaded and saved) by the `FlowManager` as the user progresses through the flow.
- `ctx.Get("key")` within a flow will retrieve data from this `UserFlowState.Data`.

**Key Distinction**:
- `ctx.SetState()` / `ctx.GetState()`: Interact with the bot's main `StateManager`, for user-level state that might persist beyond a single flow.
- `ctx.Set()` / `ctx.Get()` *inside a flow*: Interact with the current flow's temporary data, managed by the `FlowManager`'s state manager.

The `FlowManager` itself is initialized with a `StateManager`. By default, `NewBot` creates a `FlowManager` that also uses a new `InMemoryStateManager`. If you set a custom `StateManager` on the bot, you might also want to ensure the `FlowManager` uses a compatible or the same instance if you need flow state to be managed by your custom backend.

## Custom State Managers
For production bots requiring persistent state across restarts (e.g., storing data in Redis, PostgreSQL, etc.), you can implement your own `StateManager`.

### Implementing the Interface
Create a struct and implement the three methods of the `StateManager` interface:
```go
package mycustomstorage

import (
	// Your database/storage client library, e.g., "github.com/go-redis/redis/v8"
	"sync" // For thread-safety if needed by your backend logic
)

type MyCustomStateManager struct {
	// Example: client *redis.Client
	// Example: db *sql.DB
	// For a simple map-based example (not persistent):
	mu   sync.RWMutex
	data map[int64]map[string]interface{}
}

func NewMyCustomStateManager(/* dbConfig, redisAddr, etc. */) *MyCustomStateManager {
	return &MyCustomStateManager{
		data: make(map[int64]map[string]interface{}),
	}
}

func (m *MyCustomStateManager) SetState(userID int64, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Logic to store in your backend (e.g., Redis SET, SQL INSERT/UPDATE)
	// For the map example:
	if m.data[userID] == nil {
		m.data[userID] = make(map[string]interface{})
	}
	m.data[userID][key] = value
	return nil // Return error if backend operation fails
}

func (m *MyCustomStateManager) GetState(userID int64, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Logic to retrieve from your backend (e.g., Redis GET, SQL SELECT)
	// For the map example:
	userData, userExists := m.data[userID]
	if !userExists {
		return nil, false
	}
	value, keyExists := userData[key]
	return value, keyExists // Return error if backend operation fails (how to signal this?)
                            // The interface doesn't return error for GetState.
                            // Log errors internally or handle them gracefully.
}

func (m *MyCustomStateManager) ClearState(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Logic to delete from your backend (e.g., Redis DEL, SQL DELETE)
	// For the map example:
	delete(m.data, userID)
	return nil // Return error if backend operation fails
}
```

### Example: Conceptual Redis State Manager
```go
/*
// This is a conceptual example, not fully implemented.
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	teleflow "github.com/kslamph/teleflow/core" // Assuming this path
)

type RedisStateManager struct {
	client *redis.Client
	prefix string // e.g., "teleflow_state:"
}

func NewRedisStateManager(addr string, password string, db int, prefix string) (*RedisStateManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return &RedisStateManager{client: rdb, prefix: prefix}, nil
}

func (rsm *RedisStateManager) userKey(userID int64) string {
	return fmt.Sprintf("%suser:%d", rsm.prefix, userID)
}

func (rsm *RedisStateManager) SetState(userID int64, key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal state value for key %s: %w", key, err)
	}
	return rsm.client.HSet(context.Background(), rsm.userKey(userID), key, jsonData).Err()
}

func (rsm *RedisStateManager) GetState(userID int64, key string) (interface{}, bool) {
	jsonData, err := rsm.client.HGet(context.Background(), rsm.userKey(userID), key).Result()
	if err == redis.Nil {
		return nil, false // Key does not exist
	}
	if err != nil {
		log.Printf("Redis GetState error for user %d, key %s: %v", userID, key, err)
		return nil, false // Error occurred
	}

	var value interface{}
	if err := json.Unmarshal([]byte(jsonData), &value); err != nil {
		log.Printf("Redis Unmarshal error for user %d, key %s: %v", userID, key, err)
		return nil, false
	}
	return value, true
}

func (rsm *RedisStateManager) ClearState(userID int64) error {
	return rsm.client.Del(context.Background(), rsm.userKey(userID)).Err()
}
*/
```

### Setting a Custom State Manager
You can provide your custom `StateManager` when creating the bot using a `BotOption`.
**Recommendation**: Introduce a `WithStateManager(StateManager)` `BotOption` (as suggested in `improvement-recommendations.md`).

Currently, you would set it on the bot instance after creation, and potentially also on the `FlowManager` if you want flows to use the same persistent store:
```go
// Assuming WithStateManager BotOption exists:
// customManager := mycustomstorage.NewMyCustomStateManager()
// bot, err := teleflow.NewBot(token, teleflow.WithStateManager(customManager))

// Current workaround (if WithStateManager option is not yet available):
bot, err := teleflow.NewBot(token)
if err != nil { /* ... */ }
customManager := mycustomstorage.NewMyCustomStateManager()
bot.SetStateManager(customManager) // Assuming a setter method exists or direct field access
bot.FlowManager().SetStateManager(customManager) // Assuming FlowManager has a similar setter
```
*Self-correction*: `bot.stateManager` and `bot.flowManager.stateManager` are distinct by default. If a `WithStateManager` option is added, it should ideally set both, or provide separate options. For now, direct assignment after `NewBot` is the way if you need to override the default `InMemoryStateManager`.

## Request-Scoped Data vs. Persistent State

Teleflow's `Context` offers two ways to store data, serving different purposes:

### `ctx.Set()` / `ctx.Get()`
- Stores data directly in the `ctx.data map[string]interface{}`.
- **Request-scoped**: This data lives only for the duration of the current update processing. It's useful for passing information between middleware and handlers for a single request.
- **Not persistent**: Data is lost once the handler finishes.
- **Flows Exception**: When used *inside a flow*, `ctx.Set/Get` interacts with `UserFlowState.Data`, which *is* persisted by the `FlowManager`'s `StateManager` for the duration of that flow instance.

### `ctx.SetState()` / `ctx.GetState()`
- Interacts with the bot's main `StateManager` (e.g., `InMemoryStateManager` or your custom one).
- **User-scoped and potentially persistent**: Data is associated with a `userID` and its lifespan depends on the `StateManager` implementation (session-long for in-memory, longer for custom backends).
- Use this for data that needs to be remembered across different messages or bot restarts (if using a persistent manager).

## Best Practices for State Management
- **Choose the Right Scope**: Use `ctx.Set/Get` for temporary, request-specific data. Use `ctx.SetState/GetState` for data that needs to persist for a user.
- **Data Serialization**: If using a custom state manager that stores data externally (like Redis or a DB), ensure your values are serializable (e.g., to JSON, gob).
- **Minimize State**: Only store what's necessary. Large state objects can impact performance.
- **Clear Obsolete State**: Use `ctx.ClearState()` or selectively remove keys when data is no longer needed to free up resources and avoid stale data.
- **Error Handling**: Always check for errors when calling `ctx.SetState()` and `ctx.ClearState()`.
- **Concurrency**: If your custom `StateManager` interacts with a backend, ensure your implementation is thread-safe. The built-in `InMemoryStateManager` is thread-safe.

## Next Steps
- [Flow Guide](flow-guide.md): See how state is integral to managing multi-step conversations.
- [Context Deep Dive (coming soon)](): A more detailed look at the `teleflow.Context` object.
- [API Reference](api-reference.md): For details on `StateManager`, `Context` methods, and related types.
- Consider implementing a custom `StateManager` if your bot requires persistent data storage.