package main

import (
	"fmt"

	teleflow "github.com/kslamph/teleflow/core"
)

// MainMenuKeyboard creates the main menu keyboard
func MainMenuKeyboard() *teleflow.ReplyKeyboard {
	return teleflow.NewReplyKeyboard().
		AddButton("💼 Account Info").
		AddButton("💸 Transfer Funds").
		Row().
		AddButton("🛒 Place Order").
		Resize().
		Build()
}

// AccountActionsKeyboard creates inline keyboard for account actions
func AccountActionsKeyboard() *teleflow.PromptKeyboardBuilder {
	return teleflow.NewPromptKeyboard().
		ButtonCallback("👀 View My Accounts", "view_accounts").
		ButtonCallback("➕ Add New Account", "add_account")
}

// CategorySelectionKeyboard creates keyboard for product categories
func CategorySelectionKeyboard(categories []string) *teleflow.PromptKeyboardBuilder {
	kb := teleflow.NewPromptKeyboard()

	for _, category := range categories {
		kb.ButtonCallback(category, category)
		kb.Row()
	}

	return kb
}

// Merchandise represents a product item
type Merchandise struct {
	ID    string
	Name  string
	Price float64
}

// MerchandiseSelectionKeyboard creates keyboard for merchandise selection
func MerchandiseSelectionKeyboard(items []Merchandise) *teleflow.PromptKeyboardBuilder {
	kb := teleflow.NewPromptKeyboard()

	for _, item := range items {
		buttonText := item.Name + " - $" + formatPrice(item.Price)
		kb.ButtonCallback(buttonText, item.ID)
		kb.Row()
	}

	return kb
}

// ShippingOption represents a shipping method
type ShippingOption struct {
	ID     string
	Name   string
	Cost   float64
	Method string
}

// ShippingSelectionKeyboard creates keyboard for shipping options
func ShippingSelectionKeyboard(methods []ShippingOption) *teleflow.PromptKeyboardBuilder {
	kb := teleflow.NewPromptKeyboard()

	for _, method := range methods {
		buttonText := method.Method + " - $" + formatPrice(method.Cost)
		kb.ButtonCallback(buttonText, method.ID)
		kb.Row()
	}

	return kb
}

// PaymentAccountSelectionKeyboard creates keyboard for payment account selection
func PaymentAccountSelectionKeyboard(accounts []UserAccount) *teleflow.PromptKeyboardBuilder {
	kb := teleflow.NewPromptKeyboard()

	for _, account := range accounts {
		buttonText := account.Name + " - $" + formatPrice(account.Balance)
		kb.ButtonCallback(buttonText, account.AccountID)
		kb.Row()
	}

	return kb
}

// AccountSelectionKeyboard creates a general account selection keyboard
func AccountSelectionKeyboard(accounts []UserAccount, action string) *teleflow.PromptKeyboardBuilder {
	kb := teleflow.NewPromptKeyboard()

	for _, account := range accounts {
		buttonText := account.Name + " - $" + formatPrice(account.Balance)
		kb.ButtonCallback(buttonText, map[string]interface{}{
			"action":       action,
			"account_id":   account.AccountID,
			"account_name": account.Name,
		})
		kb.Row()
	}

	return kb
}

// ConfirmationKeyboard creates a yes/no confirmation keyboard
func ConfirmationKeyboard(confirmAction, cancelAction string) *teleflow.PromptKeyboardBuilder {
	return teleflow.NewPromptKeyboard().
		ButtonCallback("✅ Yes, Confirm", confirmAction).
		ButtonCallback("❌ No, Cancel", cancelAction)
}

// LocationConfirmationKeyboard creates keyboard for location confirmation
func LocationConfirmationKeyboard() *teleflow.PromptKeyboardBuilder {
	return teleflow.NewPromptKeyboard().
		ButtonCallback("✅ Yes, Confirm", "confirm_location").
		ButtonCallback("📝 No, Re-enter Address", "reenter_address")
}

// formatPrice formats a price to 2 decimal places
func formatPrice(price float64) string {
	return fmt.Sprintf("%.2f", price)
}
