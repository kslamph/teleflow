package main

import (
	"fmt"
	// "sync" // Removed sync import

	"github.com/google/uuid"
)

// UserAccount represents a user's account
type UserAccount struct {
	AccountID string  `json:"account_id"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
}

// UserData represents user data with their accounts
type UserData struct {
	TelegramUserID int64         `json:"telegram_user_id"`
	Accounts       []UserAccount `json:"accounts"`
}

// BusinessService manages user accounts and orders
type BusinessService struct {
	users map[int64]*UserData
	// mutex sync.RWMutex // Mutex removed
}

// NewBusinessService creates a new business service
func NewBusinessService() *BusinessService {
	return &BusinessService{
		users: make(map[int64]*UserData),
	}
}

// GetAccounts returns all accounts for a user
func (bs *BusinessService) GetAccounts(userID int64) []UserAccount {
	// bs.mutex.RLock() // Mutex removed
	userData, exists := bs.users[userID]
	// bs.mutex.RUnlock() // Mutex removed

	if !exists {
		// Initialize with default accounts
		// bs.mutex.Lock() // Mutex removed
		// Double-check after acquiring write lock
		userData, exists = bs.users[userID]
		if !exists {
			userData = &UserData{
				TelegramUserID: userID,
				Accounts: []UserAccount{
					{
						AccountID: uuid.New().String(),
						Name:      "Main Account",
						Balance:   1000.00,
					},
					{
						AccountID: uuid.New().String(),
						Name:      "Savings Account",
						Balance:   5000.00,
					},
				},
			}
			bs.users[userID] = userData
		}
		// bs.mutex.Unlock() // Mutex removed
	}

	return userData.Accounts
}

// AddAccount adds a new account for a user
func (bs *BusinessService) AddAccount(userID int64, accountName string, initialBalance float64) error {
	// bs.mutex.Lock() // Mutex removed
	// defer bs.mutex.Unlock() // Mutex removed

	userData, exists := bs.users[userID]
	if !exists {
		userData = &UserData{
			TelegramUserID: userID,
			Accounts:       make([]UserAccount, 0),
		}
		bs.users[userID] = userData
	}

	newAccount := UserAccount{
		AccountID: uuid.New().String(),
		Name:      accountName,
		Balance:   initialBalance,
	}

	userData.Accounts = append(userData.Accounts, newAccount)
	return nil
}

// GetAccountBalance returns the balance of a specific account
func (bs *BusinessService) GetAccountBalance(userID int64, accountID string) (float64, error) {
	// bs.mutex.RLock() // Mutex removed
	// defer bs.mutex.RUnlock() // Mutex removed

	userData, exists := bs.users[userID]
	if !exists {
		return 0, fmt.Errorf("user not found")
	}

	for _, account := range userData.Accounts {
		if account.AccountID == accountID {
			return account.Balance, nil
		}
	}

	return 0, fmt.Errorf("account not found")
}

// TransferFunds transfers money between accounts
func (bs *BusinessService) TransferFunds(userID int64, fromAccountID string, toAccountID string, amount float64) error {
	// bs.mutex.Lock() // Mutex removed
	// defer bs.mutex.Unlock() // Mutex removed

	userData, exists := bs.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	var fromAccount, toAccount *UserAccount

	// Find accounts
	for i := range userData.Accounts {
		if userData.Accounts[i].AccountID == fromAccountID {
			fromAccount = &userData.Accounts[i]
		}
		if userData.Accounts[i].AccountID == toAccountID {
			toAccount = &userData.Accounts[i]
		}
	}

	if fromAccount == nil {
		return fmt.Errorf("source account not found")
	}
	if toAccount == nil {
		return fmt.Errorf("destination account not found")
	}

	if fromAccount.Balance < amount {
		return fmt.Errorf("insufficient funds")
	}

	// Perform transfer
	fromAccount.Balance -= amount
	toAccount.Balance += amount

	return nil
}

// ProcessPayment processes a payment from an account
func (bs *BusinessService) ProcessPayment(userID int64, accountID string, amount float64) error {
	// bs.mutex.Lock() // Mutex removed
	// defer bs.mutex.Unlock() // Mutex removed

	userData, exists := bs.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	for i := range userData.Accounts {
		if userData.Accounts[i].AccountID == accountID {
			if userData.Accounts[i].Balance < amount {
				return fmt.Errorf("insufficient funds")
			}
			userData.Accounts[i].Balance -= amount
			return nil
		}
	}

	return fmt.Errorf("account not found")
}

// GetAccountByID returns a specific account by ID
func (bs *BusinessService) GetAccountByID(userID int64, accountID string) (*UserAccount, error) {
	// bs.mutex.RLock() // Mutex removed
	// defer bs.mutex.RUnlock() // Mutex removed

	userData, exists := bs.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	for _, account := range userData.Accounts {
		if account.AccountID == accountID {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account not found")
}
