package handlers

import (
	"strconv"
	"strings"

	teleflow "github.com/kslamph/teleflow/core"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/models"
)

// RegisterFlows registers all flow definitions with the bot
func RegisterFlows(bot *teleflow.Bot) {
	// Note: Flow registration would depend on the actual teleflow flow API
	// This function is kept as a placeholder for flow registration
	// Individual flow handlers would be implemented based on the actual API
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
