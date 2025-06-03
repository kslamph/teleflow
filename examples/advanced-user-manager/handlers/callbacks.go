package handlers

import (
	"strconv"

	teleflow "github.com/kslamph/teleflow/core"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/models"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/services"
)

// RegisterCallbacks registers all callback handlers with pattern matching
func RegisterCallbacks(bot *teleflow.Bot) {
	// Note: Current teleflow API uses RegisterCallback method
	// Pattern matching callbacks would be implemented differently
	// This function is kept for structure but simplified for the current API

	// User selection callbacks would be registered as:
	// bot.RegisterCallback(handler_for_user_select)

	// For now, we'll keep this as a placeholder for future callback registration
}

// handleUserSelect handles user selection from the user list
func handleUserSelect(ctx *teleflow.Context, data string) error {
	userID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	// Get user service
	userServiceVal, _ := ctx.Get("userService")
	userService := userServiceVal.(models.UserService)

	user, err := userService.GetUserByID(userID)
	if err != nil {
		return ctx.Reply("❌ User not found")
	}

	// Create action keyboard
	keyboard := createUserActionKeyboard(userID)

	return ctx.ReplyTemplate("user_details", map[string]interface{}{
		"User": user,
	}, keyboard)
}

// Note: The following functions are kept for reference but would need
// to be adapted to the actual teleflow callback API implementation

func handleChangeNameAction(ctx *teleflow.Context, data string) error {
	userID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	// Check permissions
	accessManagerVal, _ := ctx.Get("accessManager")
	accessManager := accessManagerVal.(*services.AccessManager)
	if !accessManager.CanEditUserNames(ctx) {
		return ctx.Reply("❌ You don't have permission to edit user names.")
	}

	// Log access
	accessManager.LogAccess(ctx, "change_name_start")

	// Store target user ID and start flow
	ctx.Set("target_user_id", userID)
	return ctx.StartFlow("change_name")
}

func handleToggleAction(ctx *teleflow.Context, data string) error {
	userID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	// Check permissions
	accessManagerVal, _ := ctx.Get("accessManager")
	accessManager := accessManagerVal.(*services.AccessManager)
	if !accessManager.CanToggleUserStatus(ctx) {
		return ctx.Reply("❌ You don't have permission to toggle user status.")
	}

	// Log access
	accessManager.LogAccess(ctx, "toggle_status")

	// Get user service
	userServiceVal, _ := ctx.Get("userService")
	userService := userServiceVal.(models.UserService)

	// Toggle status
	err = userService.ToggleUserStatus(userID)
	if err != nil {
		return ctx.Reply("❌ Failed to toggle user status: " + err.Error())
	}

	// Get updated user
	updatedUser, _ := userService.GetUserByID(userID)

	// Create back to list keyboard
	keyboard := createBackToListKeyboard()

	return ctx.ReplyTemplate("status_toggle_success", map[string]interface{}{
		"User": updatedUser,
	}, keyboard)
}

func handleTransferAction(ctx *teleflow.Context, data string) error {
	senderID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return ctx.Reply("❌ Invalid user selection")
	}

	// Check permissions
	accessManagerVal, _ := ctx.Get("accessManager")
	accessManager := accessManagerVal.(*services.AccessManager)
	if !accessManager.CanTransferBalance(ctx) {
		return ctx.Reply("❌ You don't have permission to transfer balance.")
	}

	// Get user service
	userServiceVal, _ := ctx.Get("userService")
	userService := userServiceVal.(models.UserService)

	sender, err := userService.GetUserByID(senderID)
	if err != nil {
		return ctx.Reply("❌ User not found")
	}

	if sender.Balance <= 0 {
		return ctx.ReplyTemplate("error_insufficient_balance", map[string]interface{}{
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

// Additional callback handlers would be implemented here
// but simplified for the current teleflow API
