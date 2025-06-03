package handlers

import (
	"fmt"

	teleflow "github.com/kslamph/teleflow/core"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/models"
)

// createUserListKeyboard creates an inline keyboard for user selection
func createUserListKeyboard(users []models.User) *teleflow.InlineKeyboard {
	var keyboard [][]teleflow.InlineKeyboardButton

	// Create buttons for users in rows of 3
	var currentRow []teleflow.InlineKeyboardButton

	for _, user := range users {
		// Create button text with status and balance
		statusIcon := "✅"
		if !user.Enabled {
			statusIcon = "❌"
		}

		buttonText := fmt.Sprintf("👤 %s ($%.0f) %s", user.Name, user.Balance, statusIcon)
		if len(buttonText) > 40 {
			// Truncate long names
			shortName := user.Name
			if len(shortName) > 15 {
				shortName = shortName[:12] + "..."
			}
			buttonText = fmt.Sprintf("👤 %s ($%.0f) %s", shortName, user.Balance, statusIcon)
		}

		button := teleflow.InlineKeyboardButton{
			Text:         buttonText,
			CallbackData: fmt.Sprintf("user_select_%d", user.ID),
		}
		currentRow = append(currentRow, button)

		// Add row when we have 3 buttons or it's the last user
		if len(currentRow) == 3 {
			keyboard = append(keyboard, currentRow)
			currentRow = []teleflow.InlineKeyboardButton{}
		}
	}

	// Add remaining buttons if any
	if len(currentRow) > 0 {
		keyboard = append(keyboard, currentRow)
	}

	// Add close menu button
	keyboard = append(keyboard, []teleflow.InlineKeyboardButton{
		{Text: "❌ Close Menu", CallbackData: "close_menu"},
	})

	return &teleflow.InlineKeyboard{
		InlineKeyboard: keyboard,
	}
}

// createUserActionKeyboard creates action buttons for a specific user
func createUserActionKeyboard(userID int64) *teleflow.InlineKeyboard {
	return &teleflow.InlineKeyboard{
		InlineKeyboard: [][]teleflow.InlineKeyboardButton{
			{
				{Text: "✏️ Change Name", CallbackData: fmt.Sprintf("action_changename_%d", userID)},
				{Text: "🔄 Enable/Disable", CallbackData: fmt.Sprintf("action_toggle_%d", userID)},
				{Text: "💰 Transfer", CallbackData: fmt.Sprintf("action_transfer_%d", userID)},
			},
			{
				{Text: "⬅️ Back to List", CallbackData: "back_to_list"},
			},
		},
	}
}

// createReceiverKeyboard creates a keyboard for selecting transfer recipients
func createReceiverKeyboard(senderID int64, users []models.User) *teleflow.InlineKeyboard {
	var keyboard [][]teleflow.InlineKeyboardButton
	var currentRow []teleflow.InlineKeyboardButton

	for _, user := range users {
		// Skip the sender and disabled users
		if user.ID == senderID || !user.Enabled {
			continue
		}

		buttonText := fmt.Sprintf("👤 %s ($%.0f)", user.Name, user.Balance)
		if len(buttonText) > 35 {
			// Truncate long names
			shortName := user.Name
			if len(shortName) > 15 {
				shortName = shortName[:12] + "..."
			}
			buttonText = fmt.Sprintf("👤 %s ($%.0f)", shortName, user.Balance)
		}

		button := teleflow.InlineKeyboardButton{
			Text:         buttonText,
			CallbackData: fmt.Sprintf("receiver_%d_%d", senderID, user.ID),
		}
		currentRow = append(currentRow, button)

		// Add row when we have 2 buttons
		if len(currentRow) == 2 {
			keyboard = append(keyboard, currentRow)
			currentRow = []teleflow.InlineKeyboardButton{}
		}
	}

	// Add remaining buttons if any
	if len(currentRow) > 0 {
		keyboard = append(keyboard, currentRow)
	}

	// Add cancel button
	keyboard = append(keyboard, []teleflow.InlineKeyboardButton{
		{Text: "❌ Cancel Transfer", CallbackData: "cancel_transfer"},
	})

	return &teleflow.InlineKeyboard{
		InlineKeyboard: keyboard,
	}
}

// createConfirmationKeyboard creates Yes/No confirmation buttons
func createConfirmationKeyboard(action string, data string) *teleflow.InlineKeyboard {
	return &teleflow.InlineKeyboard{
		InlineKeyboard: [][]teleflow.InlineKeyboardButton{
			{
				{Text: "✅ Yes", CallbackData: fmt.Sprintf("confirm_%s_%s", action, data)},
				{Text: "❌ No", CallbackData: fmt.Sprintf("cancel_%s", action)},
			},
		},
	}
}

// createTransferConfirmationKeyboard creates confirmation buttons for transfers
func createTransferConfirmationKeyboard(senderID, receiverID int64, amount float64) *teleflow.InlineKeyboard {
	return &teleflow.InlineKeyboard{
		InlineKeyboard: [][]teleflow.InlineKeyboardButton{
			{
				{Text: "✅ Confirm Transfer", CallbackData: fmt.Sprintf("confirm_transfer_%d_%d_%.2f", senderID, receiverID, amount)},
				{Text: "❌ Cancel", CallbackData: "cancel_transfer"},
			},
		},
	}
}

// createBackToListKeyboard creates a simple back to list button
func createBackToListKeyboard() *teleflow.InlineKeyboard {
	return &teleflow.InlineKeyboard{
		InlineKeyboard: [][]teleflow.InlineKeyboardButton{
			{
				{Text: "⬅️ Back to List", CallbackData: "back_to_list"},
			},
		},
	}
}
