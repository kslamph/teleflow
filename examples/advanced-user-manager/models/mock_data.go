package models

// MockUsers provides sample data for demonstrations
// This data represents a realistic user management scenario with:
// - Users with various balance levels (from $0 to $500)
// - Mix of enabled and disabled accounts
// - Realistic names for testing purposes
var MockUsers = []User{
	{
		ID:      1,
		Name:    "Alice Smith",
		Enabled: true,
		Balance: 150.50,
	},
	{
		ID:      2,
		Name:    "Bob Johnson",
		Enabled: true,
		Balance: 75.25,
	},
	{
		ID:      3,
		Name:    "Carol Williams",
		Enabled: false,
		Balance: 200.00,
	},
	{
		ID:      4,
		Name:    "Dave Brown",
		Enabled: true,
		Balance: 0.00,
	},
	{
		ID:      5,
		Name:    "Eve Davis",
		Enabled: true,
		Balance: 325.75,
	},
	{
		ID:      6,
		Name:    "Frank Wilson",
		Enabled: false,
		Balance: 50.00,
	},
	{
		ID:      7,
		Name:    "Grace Miller",
		Enabled: true,
		Balance: 500.00,
	},
	{
		ID:      8,
		Name:    "Henry Taylor",
		Enabled: true,
		Balance: 12.50,
	},
}

// GetMockUserByID returns a mock user by ID
func GetMockUserByID(id int64) (*User, bool) {
	for i, user := range MockUsers {
		if user.ID == id {
			return &MockUsers[i], true
		}
	}
	return nil, false
}

// GetActiveUsers returns only enabled users from mock data
func GetActiveUsers() []User {
	var activeUsers []User
	for _, user := range MockUsers {
		if user.Enabled {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers
}

// GetUserCount returns the total number of mock users
func GetUserCount() int {
	return len(MockUsers)
}

// GetActiveUserCount returns the number of enabled users
func GetActiveUserCount() int {
	count := 0
	for _, user := range MockUsers {
		if user.Enabled {
			count++
		}
	}
	return count
}
