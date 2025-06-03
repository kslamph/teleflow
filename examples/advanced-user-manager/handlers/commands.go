package handlers

import (
	teleflow "github.com/kslamph/teleflow/core"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/models"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/services"
)

// RegisterCommands registers all command handlers with the bot
func RegisterCommands(bot *teleflow.Bot, accessManager *services.AccessManager) {
	// Start command - shows welcome message
	bot.HandleCommand("/start", func(ctx *teleflow.Context) error {
		return handleStart(ctx, accessManager)
	})

	// Help command - shows help information
	bot.HandleCommand("/help", func(ctx *teleflow.Context) error {
		return handleHelp(ctx, accessManager)
	})

	// Cancel command - cancels current operation
	bot.HandleCommand("/cancel", func(ctx *teleflow.Context) error {
		return handleCancel(ctx, accessManager)
	})
}

// handleStart handles the /start command
func handleStart(ctx *teleflow.Context, accessManager *services.AccessManager) error {
	// Log access
	accessManager.LogAccess(ctx, "start_command")

	// Create main keyboard
	keyboard := createMainKeyboard()

	return ctx.ReplyTemplate("welcome", nil, keyboard)
}

// handleHelp handles the /help command
func handleHelp(ctx *teleflow.Context, accessManager *services.AccessManager) error {
	// Log access
	accessManager.LogAccess(ctx, "help_command")

	// Create main keyboard
	keyboard := createMainKeyboard()

	return ctx.ReplyTemplate("help", nil, keyboard)
}

// handleCancel handles the /cancel command
func handleCancel(ctx *teleflow.Context, accessManager *services.AccessManager) error {
	// Log access
	accessManager.LogAccess(ctx, "cancel_command")

	// Cancel any active flow
	if ctx.IsUserInFlow() {
		ctx.CancelFlow()
	}

	// Create main keyboard
	keyboard := createMainKeyboard()

	return ctx.ReplyTemplate("operation_cancelled", nil, keyboard)
}

// RegisterTextHandlers registers text message handlers
// RegisterTextHandlers registers text message handlers
func RegisterTextHandlers(bot *teleflow.Bot, accessManager *services.AccessManager) {
	// Register handlers for reply keyboard button presses
	bot.HandleText("👥 User Manager", func(ctx *teleflow.Context) error {
		return handleUserManagerButton(ctx, accessManager)
	})
	bot.HandleText("❓ Help", func(ctx *teleflow.Context) error {
		return handleHelpButton(ctx, accessManager)
	})
}

// handleUserManagerButton handles the User Manager button press
func handleUserManagerButton(ctx *teleflow.Context, accessManager *services.AccessManager) error {
	// Check permissions
	if !accessManager.CanManageUsers(ctx) {
		return ctx.Reply("❌ You don't have permission to manage users.")
	}

	// Log access
	accessManager.LogAccess(ctx, "user_manager_access")

	// Get user service
	userServiceVal, _ := ctx.Get("userService")
	userService := userServiceVal.(models.UserService)

	// Get all users and active count
	users := userService.GetAllUsers()
	activeUsers := userService.GetActiveUsers()

	// Create user list keyboard
	keyboard := createUserListKeyboard(users)

	// Send user list with template
	return ctx.ReplyTemplate("user_list", map[string]interface{}{
		"Users":       users,
		"ActiveCount": len(activeUsers),
	}, keyboard)
}

// handleHelpButton handles the Help button press
func handleHelpButton(ctx *teleflow.Context, accessManager *services.AccessManager) error {
	return handleHelp(ctx, accessManager)
}

// createMainKeyboard creates the main reply keyboard
func createMainKeyboard() *teleflow.ReplyKeyboard {
	return &teleflow.ReplyKeyboard{
		Keyboard: [][]teleflow.ReplyKeyboardButton{
			{
				{Text: "👥 User Manager"},
				{Text: "❓ Help"},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}
