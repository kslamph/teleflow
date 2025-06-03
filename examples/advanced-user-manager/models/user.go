package models

import (
	"errors"
	"fmt"
	"strings"
)

// User represents a user in the management system
type User struct {
	ID      int64   `json:"id"`      // Unique identifier
	Name    string  `json:"name"`    // Display name (editable)
	Enabled bool    `json:"enabled"` // Account status (toggleable)
	Balance float64 `json:"balance"` // Balance for transfers (transferable)
}

// String returns a string representation of the user
func (u User) String() string {
	status := "✅"
	if !u.Enabled {
		status = "❌"
	}
	return fmt.Sprintf("%s %s ($%.2f) %s", status, u.Name, u.Balance, status)
}

// Validate checks if the user data is valid
func (u User) Validate() error {
	if u.ID <= 0 {
		return errors.New("user ID must be positive")
	}
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("user name cannot be empty")
	}
	if len(u.Name) > 50 {
		return errors.New("user name must be less than 50 characters")
	}
	if u.Balance < 0 {
		return errors.New("user balance cannot be negative")
	}
	return nil
}

// CanTransfer checks if the user can transfer the specified amount
func (u User) CanTransfer(amount float64) bool {
	return u.Enabled && u.Balance >= amount && amount > 0
}

// UserService interface defines the operations for user management
type UserService interface {
	GetAllUsers() []User
	GetUserByID(id int64) (*User, error)
	UpdateUser(user *User) error
	TransferBalance(fromID, toID int64, amount float64) error
	GetActiveUsers() []User
	GetUserCount() int
	ToggleUserStatus(id int64) error
	UpdateUserName(id int64, newName string) error
}
