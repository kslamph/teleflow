package handlers

import (
	"strconv"
	"strings"

	teleflow "github.com/kslamph/teleflow/core"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/models"
)

// RegisterFlows registers all flow definitions with the bot
func RegisterFlows(bot *teleflow.Bot, userService models.UserService) {
	// Register change name flow
	changeNameFlow := teleflow.NewFlow("change_name").
		Step("request_name").
		OnStart(func(ctx *teleflow.Context) error {
			// Get target user ID from context
			userIDInterface, exists := ctx.Get("target_user_id")
			if !exists {
				return ctx.Reply("❌ User ID not found")
			}

			userID := userIDInterface.(int64)

			user, err := userService.GetUserByID(userID)
			if err != nil {
				return ctx.Reply("❌ User not found")
			}

			return ctx.ReplyTemplate("current_name", map[string]interface{}{
				"User": user,
			})
		}).
		OnInput(func(ctx *teleflow.Context) error {
			// Get new name from message
			// We must check if Message is nil, as this handler can be triggered by a callback.
			if ctx.Update.Message == nil {
				// This means a callback (like confirmation) likely triggered this.
				// The actual logic for this case should be handled by the callback itself
				// or the flow should transition to a different step.
				// For now, if it's not a text message, we can't process it as a name.
				return nil // Or an error, or a specific handling
			}
			newName := ctx.Update.Message.Text

			// Validate name
			if valid, errorMsg := nameValidator(newName); !valid {
				return ctx.Reply(errorMsg)
			}

			// Store new name for confirmation
			ctx.Set("new_name", newName)

			// Get target user for confirmation
			userIDInterface, _ := ctx.Get("target_user_id")
			userID := userIDInterface.(int64)

			// Create confirmation keyboard
			// The callback data for "changename" should include userID for context.
			keyboard := createConfirmationKeyboard("changename", strconv.FormatInt(userID, 10))

			return ctx.ReplyTemplate("confirm_name_change", map[string]interface{}{
				"NewName": newName,
			}, keyboard)
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			// This is called when the flow completes successfully.
			// The actual name update should happen *before* this, triggered by the confirmation.
			// This handler just sends the final success message.

			newNameInterface, newNameExists := ctx.Get("new_name")
			if !newNameExists {
				// Should not happen if flow proceeded correctly
				return ctx.Reply("Error: New name not found in context.")
			}
			newName := newNameInterface.(string)

			keyboard := createBackToListKeyboard()
			return ctx.ReplyTemplate("name_change_success", map[string]interface{}{
				"NewName": newName,
			}, keyboard)
		}).
		OnCancel(func(ctx *teleflow.Context) error {
			keyboard := createBackToListKeyboard()
			return ctx.Reply("❌ Name change cancelled.", keyboard)
		}).
		Build()

	// Register transfer balance flow
	transferFlow := teleflow.NewFlow("transfer_balance").
		Step("request_amount").
		OnStart(func(ctx *teleflow.Context) error {
			// Get sender ID from context
			senderIDInterface, exists := ctx.Get("sender_id")
			if !exists {
				return ctx.Reply("❌ Sender ID not found")
			}

			senderID := senderIDInterface.(int64)

			sender, err := userService.GetUserByID(senderID)
			if err != nil {
				return ctx.Reply("❌ User not found")
			}

			return ctx.ReplyTemplate("request_transfer_amount", map[string]interface{}{
				"Sender": sender,
			})
		}).
		OnInput(func(ctx *teleflow.Context) error {
			if ctx.Update.Message == nil {
				// This step's OnInput expects a text message (the amount).
				// If triggered by a callback, it means the flow is being advanced by a callback handler.
				return nil // Let the callback handler (e.g., for recipient selection) drive the next action.
			}
			// Get amount from message
			amountText := ctx.Update.Message.Text

			// Validate amount
			if valid, errorMsg := amountValidator(amountText); !valid {
				return ctx.Reply(errorMsg)
			}

			// Parse amount
			amount, _ := strconv.ParseFloat(amountText, 64)

			// Check sender balance
			senderIDInterface, _ := ctx.Get("sender_id")
			senderID := senderIDInterface.(int64)

			sender, _ := userService.GetUserByID(senderID)
			if !sender.CanTransfer(amount) {
				return ctx.ReplyTemplate("error_insufficient_balance", map[string]interface{}{
					"User":   sender,
					"Amount": amount,
				})
			}

			// Store amount and show recipient selection
			ctx.Set("transfer_amount", amount)

			// Get all users for recipient selection
			users := userService.GetAllUsers()
			keyboard := createReceiverKeyboard(senderID, users)

			return ctx.ReplyTemplate("select_receiver", map[string]interface{}{
				"Sender": sender,
				"Amount": amount,
			}, keyboard)
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			// This would be triggered by confirmation callback
			keyboard := createBackToListKeyboard()
			return ctx.Reply("✅ Transfer completed!", keyboard)
		}).
		OnCancel(func(ctx *teleflow.Context) error {
			keyboard := createBackToListKeyboard()
			return ctx.Reply("❌ Transfer cancelled.", keyboard)
		}).
		Build()

	// Register flows with bot
	bot.RegisterFlow(changeNameFlow)
	bot.RegisterFlow(transferFlow)
}

// Flow handlers and validation functions are kept for reference
// but would need to be adapted to the actual teleflow flow implementation

// nameValidator validates name input
func nameValidator(input string) (bool, string) {
	name := strings.TrimSpace(input)

	if len(name) < 2 {
		return false, "❌ Name must be at least 2 characters long"
	}
	if len(name) > 50 {
		return false, "❌ Name must be less than 50 characters"
	}
	if name == "" {
		return false, "❌ Name cannot be empty"
	}

	// Check for invalid characters (basic validation)
	for _, char := range name {
		if char < 32 || char > 126 {
			return false, "❌ Name contains invalid characters"
		}
	}

	return true, ""
}

// amountValidator validates transfer amount input
func amountValidator(input string) (bool, string) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		return false, "❌ Please enter a valid number"
	}
	if amount <= 0 {
		return false, "❌ Amount must be greater than 0"
	}
	if amount > 10000 {
		return false, "❌ Maximum transfer amount is $10,000"
	}

	// Check for reasonable decimal places
	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)
	if strings.Contains(amountStr, ".") {
		parts := strings.Split(amountStr, ".")
		if len(parts[1]) > 2 {
			return false, "❌ Amount can have at most 2 decimal places"
		}
	}

	return true, ""
}

// Example flow step handlers (would be adapted to actual teleflow flow API)

func handleNameChangeStep(ctx *teleflow.Context) error {
	// Get target user ID from context
	userIDInterface, exists := ctx.Get("target_user_id")
	if !exists {
		return ctx.Reply("❌ User ID not found")
	}

	userID := userIDInterface.(int64)
	userServiceVal, _ := ctx.Get("userService")
	userService := userServiceVal.(models.UserService)

	user, err := userService.GetUserByID(userID)
	if err != nil {
		return ctx.Reply("❌ User not found")
	}

	return ctx.ReplyTemplate("current_name", map[string]interface{}{
		"User": user,
	})
}

func handleTransferAmountStep(ctx *teleflow.Context) error {
	// Get sender ID from context
	senderIDInterface, exists := ctx.Get("sender_id")
	if !exists {
		return ctx.Reply("❌ Sender ID not found")
	}

	senderID := senderIDInterface.(int64)
	userServiceVal, _ := ctx.Get("userService")
	userService := userServiceVal.(models.UserService)

	sender, err := userService.GetUserByID(senderID)
	if err != nil {
		return ctx.Reply("❌ User not found")
	}

	return ctx.ReplyTemplate("request_transfer_amount", map[string]interface{}{
		"Sender": sender,
	})
}

// Additional flow step handlers would be implemented here
// based on the actual teleflow flow API
