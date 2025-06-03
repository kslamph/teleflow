package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kslamph/teleflow/examples/advanced-user-manager/models"
)

// userService implements the UserService interface using mock data
type userService struct {
	users []models.User
}

// NewUserService creates a new user service with mock data
func NewUserService() models.UserService {
	// Copy mock data to avoid modifying the original
	users := make([]models.User, len(models.MockUsers))
	copy(users, models.MockUsers)

	return &userService{
		users: users,
	}
}

// GetAllUsers returns all users
func (s *userService) GetAllUsers() []models.User {
	return s.users
}

// GetUserByID returns a user by ID
func (s *userService) GetUserByID(id int64) (*models.User, error) {
	for i, user := range s.users {
		if user.ID == id {
			return &s.users[i], nil
		}
	}
	return nil, fmt.Errorf("user with ID %d not found", id)
}

// UpdateUser updates a user's information
func (s *userService) UpdateUser(user *models.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user data: %w", err)
	}

	for i, u := range s.users {
		if u.ID == user.ID {
			s.users[i] = *user
			return nil
		}
	}
	return fmt.Errorf("user with ID %d not found", user.ID)
}

// TransferBalance transfers balance between users
func (s *userService) TransferBalance(fromID, toID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("transfer amount must be positive")
	}

	if fromID == toID {
		return errors.New("cannot transfer to the same user")
	}

	// Find sender and receiver
	var sender, receiver *models.User
	for i := range s.users {
		if s.users[i].ID == fromID {
			sender = &s.users[i]
		}
		if s.users[i].ID == toID {
			receiver = &s.users[i]
		}
	}

	if sender == nil {
		return fmt.Errorf("sender with ID %d not found", fromID)
	}
	if receiver == nil {
		return fmt.Errorf("receiver with ID %d not found", toID)
	}

	// Validate transfer conditions
	if !sender.Enabled {
		return errors.New("sender account is disabled")
	}
	if !receiver.Enabled {
		return errors.New("receiver account is disabled")
	}
	if !sender.CanTransfer(amount) {
		return fmt.Errorf("insufficient balance: sender has $%.2f, transfer amount is $%.2f",
			sender.Balance, amount)
	}

	// Perform transfer
	sender.Balance -= amount
	receiver.Balance += amount

	return nil
}

// GetActiveUsers returns only enabled users
func (s *userService) GetActiveUsers() []models.User {
	var activeUsers []models.User
	for _, user := range s.users {
		if user.Enabled {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers
}

// GetUserCount returns the total number of users
func (s *userService) GetUserCount() int {
	return len(s.users)
}

// ToggleUserStatus toggles the enabled status of a user
func (s *userService) ToggleUserStatus(id int64) error {
	for i, user := range s.users {
		if user.ID == id {
			s.users[i].Enabled = !s.users[i].Enabled
			return nil
		}
	}
	return fmt.Errorf("user with ID %d not found", id)
}

// UpdateUserName updates a user's name with validation
func (s *userService) UpdateUserName(id int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return errors.New("name cannot be empty")
	}
	if len(newName) < 2 {
		return errors.New("name must be at least 2 characters long")
	}
	if len(newName) > 50 {
		return errors.New("name must be less than 50 characters")
	}

	for i, user := range s.users {
		if user.ID == id {
			s.users[i].Name = newName
			return nil
		}
	}
	return fmt.Errorf("user with ID %d not found", id)
}
