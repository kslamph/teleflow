package handlers

import (
	"strconv"
	"strings"

	teleflow "github.com/kslamph/teleflow/core"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/models"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/services"
)

// RegisterCallbacks registers all callback handlers with pattern matching
func RegisterCallbacks(bot *teleflow.Bot, accessManager *services.AccessManager, userService models.UserService) {
	// Register user selection callbacks with pattern matching
	bot.RegisterCallback(teleflow.SimpleCallback("user_select_*", func(ctx *teleflow.Context, data string) error {
		return handleUserSelect(ctx, data, accessManager, userService)
	}))

	// Register action callbacks with pattern matching
	bot.RegisterCallback(teleflow.SimpleCallback("action_changename_*", func(ctx *teleflow.Context, data string) error {
		return handleChangeNameAction(ctx, data, accessManager)
	}))

	bot.RegisterCallback(teleflow.SimpleCallback("action_toggle_*", func(ctx *teleflow.Context, data string) error {
		return handleToggleAction(ctx, data, accessManager, userService)
	}))

	bot.RegisterCallback(teleflow.SimpleCallback("action_transfer_*", func(ctx *teleflow.Context, data string) error {
		return handleTransferAction(ctx, data, accessManager, userService)
	}))

	// Register simple callbacks
	bot.RegisterCallback(teleflow.SimpleCallback("close_menu", func(ctx *teleflow.Context, data string) error {
		return ctx.EditOrReply("✅ Menu closed. Use the buttons below to access features.")
	}))

	bot.RegisterCallback(teleflow.SimpleCallback("back_to_list", func(ctx *teleflow.Context, data string) error {
		return handleBackToList(ctx, accessManager, userService)
	}))

	// Register confirmation callbacks
	bot.RegisterCallback(teleflow.SimpleCallback("confirm_changename_*", func(ctx *teleflow.Context, data string) error {
		// 'data' here is the userID string matched by the wildcard "*"
		return handleConfirmNameChange(ctx, data, accessManager, userService)
	}))

	bot.RegisterCallback(teleflow.SimpleCallback("cancel_changename", func(ctx *teleflow.Context, data string) error {
		return handleCancelNameChange(ctx, accessManager)
	}))

	// Register receiver selection callbacks
	bot.RegisterCallback(teleflow.SimpleCallback("receiver_*", func(ctx *teleflow.Context, data string) error {
		return handleReceiverSelection(ctx, data, accessManager, userService)
	}))

	bot.RegisterCallback(teleflow.SimpleCallback("cancel_transfer", func(ctx *teleflow.Context, data string) error {
		return handleCancelTransfer(ctx, accessManager)
	}))

	// Register transfer confirmation callbacks
	bot.RegisterCallback(teleflow.SimpleCallback("confirm_transfer_*", func(ctx *teleflow.Context, data string) error {
		return handleConfirmTransfer(ctx, data, accessManager, userService)
	}))
}

// handleUserSelect handles user selection from the user list
func handleUserSelect(ctx *teleflow.Context, data string, accessManager *services.AccessManager, userService models.UserService) error {
	userID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	user, err := userService.GetUserByID(userID)
	if err != nil {
		return ctx.Reply("❌ User not found")
	}

	// Create action keyboard
	keyboard := createUserActionKeyboard(userID)

	return ctx.EditOrReplyTemplate("user_details", map[string]interface{}{
		"User": user,
	}, keyboard)
}

func handleChangeNameAction(ctx *teleflow.Context, data string, accessManager *services.AccessManager) error {
	userID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	// Check permissions
	if !accessManager.CanEditUserNames(ctx) {
		return ctx.Reply("❌ You don't have permission to edit user names.")
	}

	// Log access
	accessManager.LogAccess(ctx, "change_name_start")

	// Store target user ID and start flow
	ctx.Set("target_user_id", userID)
	return ctx.StartFlow("change_name")
}

func handleToggleAction(ctx *teleflow.Context, data string, accessManager *services.AccessManager, userService models.UserService) error {
	userID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	// Check permissions
	if !accessManager.CanToggleUserStatus(ctx) {
		return ctx.Reply("❌ You don't have permission to toggle user status.")
	}

	// Log access
	accessManager.LogAccess(ctx, "toggle_status")

	// Toggle status
	err = userService.ToggleUserStatus(userID)
	if err != nil {
		return ctx.Reply("❌ Failed to toggle user status: " + err.Error())
	}

	// Get updated user
	updatedUser, _ := userService.GetUserByID(userID)

	// Create back to list keyboard
	keyboard := createBackToListKeyboard()

	return ctx.EditOrReplyTemplate("status_toggle_success", map[string]interface{}{
		"User": updatedUser,
	}, keyboard)
}

func handleTransferAction(ctx *teleflow.Context, data string, accessManager *services.AccessManager, userService models.UserService) error {
	senderID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	// Check permissions
	if !accessManager.CanTransferBalance(ctx) {
		return ctx.Reply("❌ You don't have permission to transfer balance.")
	}

	sender, err := userService.GetUserByID(senderID)
	if err != nil {
		return ctx.Reply("❌ User not found")
	}

	if sender.Balance <= 0 {
		return ctx.EditOrReplyTemplate("error_insufficient_balance", map[string]interface{}{
			"User":   sender,
			"Amount": 0.01,
		})
	}

	// Log access
	accessManager.LogAccess(ctx, "transfer_start")

	// Store sender ID and start flow
	ctx.Set("sender_id", senderID)
	return ctx.StartFlow("transfer_balance")
}

// handleBackToList handles the back to list button
func handleBackToList(ctx *teleflow.Context, accessManager *services.AccessManager, userService models.UserService) error {
	// Get all users and active count
	users := userService.GetAllUsers()
	activeUsers := userService.GetActiveUsers()

	// Create user list keyboard
	keyboard := createUserListKeyboard(users)

	// Send user list with template
	return ctx.EditOrReplyTemplate("user_list", map[string]interface{}{
		"Users":       users,
		"ActiveCount": len(activeUsers),
	}, keyboard)
}

// handleConfirmNameChange handles name change confirmation.
// userIDStr is the string representation of the user ID, extracted from the callback data by SimpleCallback.
func handleConfirmNameChange(ctx *teleflow.Context, userIDStr string, accessManager *services.AccessManager, userService models.UserService) error {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user ID in confirmation.")
	}

	// Retrieve the new name from the flow's context
	newNameInterface, newNameExists := ctx.Get("new_name")
	if !newNameExists {
		// This could happen if the flow state was lost or this callback was triggered out of sequence
		ctx.CancelFlow() // Cancel the flow as a precaution
		return ctx.Reply("❌ Error: Could not find the new name. Please try again.")
	}
	newName := newNameInterface.(string)

	// Perform the name update
	err = userService.UpdateUserName(userID, newName)
	if err != nil {
		ctx.CancelFlow() // Cancel the flow on error
		return ctx.Reply("❌ Failed to update name: " + err.Error())
	}

	// Log access
	accessManager.LogAccess(ctx, "change_name_confirmed")

	// Manually send success message and cancel the flow to ensure UI updates and state clears.
	// This is because relying on the flow's OnComplete being triggered by a callback's nil return
	// might not be consistently updating the UI as expected.
	keyboard := createBackToListKeyboard()
	replyErr := ctx.ReplyTemplate("name_change_success", map[string]interface{}{
		"NewName": newName, // newName is from earlier in this function
	}, keyboard)
	if replyErr != nil {
		// Log this error but don't let it stop the flow cancellation.
		// Consider logging to a more persistent store in a real app.
		// For now, printing to console if available or just ignoring.
		// log.Printf("Error sending success message for name change: %v", replyErr)
	}

	ctx.CancelFlow() // Explicitly cancel the flow to clear state.
	return nil       // Return nil as per callback handler convention.
}

// handleCancelNameChange handles name change cancellation
func handleCancelNameChange(ctx *teleflow.Context, accessManager *services.AccessManager) error {
	// Cancel the flow and return to list
	ctx.CancelFlow()
	keyboard := createBackToListKeyboard()
	return ctx.Reply("❌ Name change cancelled.", keyboard)
}

// handleReceiverSelection handles transfer recipient selection
func handleReceiverSelection(ctx *teleflow.Context, data string, accessManager *services.AccessManager, userService models.UserService) error {
	// Parse callback data: "senderID_receiverID"
	parts := strings.Split(data, "_")
	if len(parts) != 2 {
		return ctx.Reply("❌ Invalid selection")
	}

	senderID, err1 := strconv.ParseInt(parts[0], 10, 64)
	receiverID, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		return ctx.Reply("❌ Invalid user IDs")
	}

	// Get transfer amount from context
	amountInterface, exists := ctx.Get("transfer_amount")
	if !exists {
		return ctx.Reply("❌ Transfer amount not found")
	}
	amount := amountInterface.(float64)

	// Get users
	sender, err := userService.GetUserByID(senderID)
	if err != nil {
		return ctx.Reply("❌ Sender not found")
	}
	receiver, err := userService.GetUserByID(receiverID)
	if err != nil {
		return ctx.Reply("❌ Receiver not found")
	}

	// Store receiver ID
	ctx.Set("receiver_id", receiverID)

	// Create confirmation keyboard
	keyboard := createTransferConfirmationKeyboard(senderID, receiverID, amount)

	// Show confirmation
	return ctx.EditOrReplyTemplate("confirm_transfer", map[string]interface{}{
		"Sender":               sender,
		"Receiver":             receiver,
		"Amount":               amount,
		"SenderBalanceAfter":   sender.Balance - amount,
		"ReceiverBalanceAfter": receiver.Balance + amount,
	}, keyboard)
}

// handleCancelTransfer handles transfer cancellation
func handleCancelTransfer(ctx *teleflow.Context, accessManager *services.AccessManager) error {
	// Cancel the flow and return to list
	ctx.CancelFlow()
	keyboard := createBackToListKeyboard()
	return ctx.Reply("❌ Transfer cancelled.", keyboard)
}

// handleConfirmTransfer handles transfer confirmation and execution
func handleConfirmTransfer(ctx *teleflow.Context, data string, accessManager *services.AccessManager, userService models.UserService) error {
	// Parse callback data: "senderID_receiverID_amount"
	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return ctx.Reply("❌ Invalid transfer data")
	}

	senderID, err1 := strconv.ParseInt(parts[0], 10, 64)
	receiverID, err2 := strconv.ParseInt(parts[1], 10, 64)
	amount, err3 := strconv.ParseFloat(parts[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return ctx.Reply("❌ Invalid transfer parameters")
	}

	// Perform the transfer
	err := userService.TransferBalance(senderID, receiverID, amount)
	if err != nil {
		return ctx.Reply("❌ Transfer failed: " + err.Error())
	}

	// Get updated users
	sender, _ := userService.GetUserByID(senderID)
	receiver, _ := userService.GetUserByID(receiverID)

	// Log the transfer
	accessManager.LogAccess(ctx, "transfer_completed")

	// Create back to list keyboard
	keyboard := createBackToListKeyboard()

	// Show success message
	return ctx.EditOrReplyTemplate("transfer_success", map[string]interface{}{
		"Sender":   sender,
		"Receiver": receiver,
		"Amount":   amount,
	}, keyboard)
}
