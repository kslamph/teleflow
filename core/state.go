package teleflow

import (
	"sync"
)

type StateManager interface {

	//

	//

	SetState(userID int64, key string, value interface{}) error

	//

	//

	GetState(userID int64, key string) (interface{}, bool)

	//

	//

	ClearState(userID int64) error
}

type inMemoryStateManager struct {
	mu sync.RWMutex

	data map[int64]map[string]interface{}
}

func NewInMemoryStateManager() StateManager {
	return &inMemoryStateManager{
		data: make(map[int64]map[string]interface{}),
	}
}

func (m *inMemoryStateManager) SetState(userID int64, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data[userID] == nil {
		m.data[userID] = make(map[string]interface{})
	}

	m.data[userID][key] = value

	return nil
}

func (m *inMemoryStateManager) GetState(userID int64, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userData, userExists := m.data[userID]
	if !userExists {
		return nil, false
	}

	value, keyExists := userData[key]
	return value, keyExists
}

func (m *inMemoryStateManager) ClearState(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, userID)

	return nil
}
