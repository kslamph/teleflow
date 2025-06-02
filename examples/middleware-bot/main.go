package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	teleflow "github.com/kslamph/teleflow/core"
)

// AdminAccessManager implements AccessManager with admin/user roles
type AdminAccessManager struct {
	adminUsers map[int64]bool
}

// NewAdminAccessManager creates a new AccessManager with admin users from environment
func NewAdminAccessManager() *AdminAccessManager {
	checker := &AdminAccessManager{
		adminUsers: make(map[int64]bool),
	}

	// Parse admin users from environment variable ADMIN_USERS (comma-separated user IDs)
	adminUsersEnv := os.Getenv("ADMIN_USERS")
	if adminUsersEnv != "" {
		userIDs := strings.Split(adminUsersEnv, ",")
		for _, userIDStr := range userIDs {
			userIDStr = strings.TrimSpace(userIDStr)
			if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
				checker.adminUsers[userID] = true
				log.Printf("Registered admin user: %d", userID)
			} else {
				log.Printf("Invalid admin user ID in ADMIN_USERS: %s", userIDStr)
			}
		}
	}

	return checker
}

// CheckPermission checks if user can execute based on permission context
func (c *AdminAccessManager) CheckPermission(ctx *teleflow.PermissionContext) error {
	// Check if user is admin
	isAdmin := c.adminUsers[ctx.UserID]

	// Allow admin users to do anything
	if isAdmin {
		return nil
	}

	// Check command-specific permissions for non-admin users
	if ctx.Command != "" {
		switch ctx.Command {
		case "admin", "status", "panic":
			return fmt.Errorf("Admin access required for command '%s'", ctx.Command)
		default:
			// Allow basic commands for all users
			return nil
		}
	}

	// Allow basic access for non-command interactions
	return nil
}

// IsAdmin is a helper method to check if a user is admin (for legacy compatibility)
func (c *AdminAccessManager) IsAdmin(userID int64) bool {
	return c.adminUsers[userID]
}

// GetReplyKeyboard returns different keyboards based on user role
func (c *AdminAccessManager) GetReplyKeyboard(ctx *teleflow.MenuContext) *teleflow.ReplyKeyboard {
	keyboard := teleflow.NewReplyKeyboard()

	// Basic buttons for all users
	keyboard.AddButton("🏠 Home").AddButton("ℹ️ Help").AddRow()
	keyboard.AddButton("⚡ Spam Test").AddButton("💥 Panic Test").AddRow()

	// Admin-only button
	if c.adminUsers[ctx.UserID] {
		keyboard.AddButton("👑 Admin Panel").AddRow()
	}

	keyboard.Resize()
	return keyboard
}

// GetMenuButton returns different menu buttons based on user role
func (c *AdminAccessManager) GetMenuButton(ctx *teleflow.MenuContext) *teleflow.MenuButtonConfig {
	// Check if user is admin
	if c.adminUsers[ctx.UserID] {
		// Admin menu button with advanced options
		return &teleflow.MenuButtonConfig{
			Type: teleflow.MenuButtonTypeCommands,
			Items: []teleflow.MenuButtonItem{
				{Text: "🏠 Home", Command: "/start"},
				{Text: "👑 Admin", Command: "/admin"},
				{Text: "📊 Status", Command: "/status"},
				{Text: "⚡ Spam Test", Command: "/spam"},
				{Text: "💥 Panic Test", Command: "/panic"},
				{Text: "ℹ️ Help", Command: "/help"},
			},
		}
	}

	// Regular user menu button with basic options
	return &teleflow.MenuButtonConfig{
		Type: teleflow.MenuButtonTypeCommands,
		Items: []teleflow.MenuButtonItem{
			{Text: "🏠 Home", Command: "/start"},
			{Text: "⚡ Spam Test", Command: "/spam"},
			{Text: "💥 Panic Test", Command: "/panic"},
			{Text: "ℹ️ Help", Command: "/help"},
		},
	}
}

func main() {
	// Get bot token from environment variable
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable is required")
	}

	// Create access manager with admin users from environment
	accessManager := NewAdminAccessManager()

	// Create new bot instance
	bot, err := teleflow.NewBot(token)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// ========================================
	// MIDDLEWARE STACK SETUP (Order matters!)
	// ========================================

	// 1. Recovery middleware (outermost) - catches panics and prevents crashes
	bot.Use(teleflow.RecoveryMiddleware())

	// 2. Logging middleware - logs all requests and execution times
	bot.Use(teleflow.LoggingMiddleware())

	// 3. Rate limiting middleware - limits requests to 2 per minute for demonstration
	bot.Use(teleflow.RateLimitMiddleware(2))

	// 4. Auth middleware (innermost) - checks basic user authorization
	bot.Use(teleflow.AuthMiddleware(accessManager))

	// Register command handlers
	registerCommands(bot, accessManager)

	// Register text handlers for keyboard buttons
	registerTextHandlers(bot, accessManager)

	// Start the bot
	log.Println("Starting middleware demonstration bot...")
	log.Println("Middleware stack active:")
	log.Println("  1. RecoveryMiddleware() - Panic recovery")
	log.Println("  2. LoggingMiddleware() - Request logging")
	log.Println("  3. RateLimitMiddleware(2) - Rate limiting (2 req/min)")
	log.Println("  4. AuthMiddleware() - Authorization")
	log.Println("")
	log.Printf("Admin users configured: %s", os.Getenv("ADMIN_USERS"))

	if err := bot.Start(); err != nil {
		log.Fatal("Failed to start bot:", err)
	}
}

// registerCommands sets up all command handlers with middleware demonstrations
func registerCommands(bot *teleflow.Bot, permissionChecker *AdminAccessManager) {
	// /start command - Welcome message explaining middleware demo
	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		welcomeText := "🛡️ **Middleware Demonstration Bot**\n\n" +
			"This bot showcases all Teleflow middleware types in action:\n\n" +
			"🔍 **Active Middleware:**\n" +
			"• `RecoveryMiddleware()` - Handles panics gracefully\n" +
			"• `LoggingMiddleware()` - Logs all requests and timing\n" +
			"• `RateLimitMiddleware(2)` - Limits to 2 requests/minute\n" +
			"• `AuthMiddleware()` - Checks user permissions\n\n" +
			"🎯 **Commands to Try:**\n" +
			"• `/admin` - Admin-only command (requires admin role)\n" +
			"• `/spam` - Test rate limiting (try multiple times quickly)\n" +
			"• `/panic` - Trigger panic to test recovery\n" +
			"• `/help` - Detailed middleware information\n\n" +
			"Use the buttons below to interact with different middleware features!"

		keyboard := permissionChecker.GetReplyKeyboard(ctx.GetMenuContext())
		return ctx.Reply(welcomeText, keyboard)
	})

	// /admin command - Admin-only command demonstrating authorization middleware
	bot.HandleCommand("admin", func(ctx *teleflow.Context) error {
		// This command has additional admin-specific authorization check
		if !permissionChecker.IsAdmin(ctx.UserID()) {
			return ctx.Reply("🚫 **Access Denied**\n\n" +
				"This command requires admin privileges.\n\n" +
				"🔒 **Authorization Middleware in Action:**\n" +
				"• Your user ID: `" + strconv.FormatInt(ctx.UserID(), 10) + "`\n" +
				"• Required permission: `admin_access`\n" +
				"• Result: **DENIED**\n\n" +
				"💡 **How to become admin:**\n" +
				"Set the `ADMIN_USERS` environment variable with your user ID.\n" +
				"Example: `ADMIN_USERS=123456789,987654321`")
		}

		adminText := "👑 **Admin Panel Access Granted**\n\n" +
			"🎉 **Authorization Middleware Success:**\n" +
			"• Your user ID: `" + strconv.FormatInt(ctx.UserID(), 10) + "`\n" +
			"• Required permission: `admin_access`\n" +
			"• Result: **GRANTED**\n\n" +
			"🔧 **Admin Features:**\n" +
			"• Full system access\n" +
			"• User management capabilities\n" +
			"• Advanced configuration options\n" +
			"• Monitoring and analytics\n\n" +
			"⚠️ **Note:** This demonstrates how AuthMiddleware works with custom permission checkers to implement role-based access control."

		keyboard := permissionChecker.GetReplyKeyboard(ctx.GetMenuContext())
		return ctx.Reply(adminText, keyboard)
	})

	// /spam command - Test rate limiting middleware
	bot.HandleCommand("spam", func(ctx *teleflow.Context) error {
		spamText := "⚡ **Rate Limiting Test**\n\n" +
			"🚦 **RateLimitMiddleware(2) in Action:**\n" +
			"• Limit: 2 requests per minute\n" +
			"• Your user ID: `" + strconv.FormatInt(ctx.UserID(), 10) + "`\n" +
			"• This request: **ALLOWED**\n\n" +
			"🧪 **How to Test:**\n" +
			"1. Send this command multiple times quickly\n" +
			"2. After 2 requests, you'll get rate limited\n" +
			"3. Wait 30 seconds between requests to avoid limits\n\n" +
			"💡 **Implementation Details:**\n" +
			"• Uses in-memory tracking per user ID\n" +
			"• Calculates minimum interval based on rate\n" +
			"• Blocks requests that exceed the limit\n" +
			"• Provides clear feedback to users"

		keyboard := permissionChecker.GetReplyKeyboard(ctx.GetMenuContext())
		return ctx.Reply(spamText, keyboard)
	})

	// /panic command - Test recovery middleware
	bot.HandleCommand("panic", func(ctx *teleflow.Context) error {
		panicText := "💥 **Panic Recovery Test**\n\n" +
			"🛡️ **RecoveryMiddleware() Demonstration:**\n" +
			"This command will intentionally trigger a panic to show how the recovery middleware handles it gracefully.\n\n" +
			"⚠️ **What happens next:**\n" +
			"1. A panic will be triggered in the handler\n" +
			"2. RecoveryMiddleware catches the panic\n" +
			"3. Error is logged to server console\n" +
			"4. User receives friendly error message\n" +
			"5. Bot continues running normally\n\n" +
			"🎯 **Ready to trigger panic?**"

		keyboard := teleflow.NewReplyKeyboard()
		keyboard.AddButton("💥 YES - Trigger Panic").AddButton("🚫 Cancel").AddRow()
		keyboard.Resize()

		return ctx.Reply(panicText, keyboard)
	})

	// /help command - Detailed middleware information
	bot.HandleCommand("help", func(ctx *teleflow.Context) error {
		helpText := "📚 **Middleware System Help**\n\n" +
			"🔄 **Middleware Execution Order:**\n" +
			"1. `RecoveryMiddleware()` - Panic recovery (outermost)\n" +
			"2. `LoggingMiddleware()` - Request logging\n" +
			"3. `RateLimitMiddleware(2)` - Rate limiting\n" +
			"4. `AuthMiddleware()` - Authorization (innermost)\n\n" +
			"🛡️ **RecoveryMiddleware:**\n" +
			"• Catches panics in handlers\n" +
			"• Logs panic details to console\n" +
			"• Returns user-friendly error message\n" +
			"• Prevents bot crashes\n\n" +
			"📝 **LoggingMiddleware:**\n" +
			"• Logs all incoming updates\n" +
			"• Tracks handler execution time\n" +
			"• Shows success/failure status\n" +
			"• Includes user ID and update type\n\n" +
			"⚡ **RateLimitMiddleware(2):**\n" +
			"• Limits to 2 requests per minute per user\n" +
			"• Uses in-memory tracking\n" +
			"• Calculates 30-second intervals\n" +
			"• Provides rate limit feedback\n\n" +
			"🔐 **AuthMiddleware:**\n" +
			"• Checks basic user authorization\n" +
			"• Uses custom permission checker\n" +
			"• Supports role-based access control\n" +
			"• Blocks unauthorized access\n\n" +
			"🎯 **Test Commands:**\n" +
			"• `/admin` - Test authorization\n" +
			"• `/spam` - Test rate limiting\n" +
			"• `/panic` - Test panic recovery"

		keyboard := permissionChecker.GetReplyKeyboard(ctx.GetMenuContext())
		return ctx.Reply(helpText, keyboard)
	})
}

// registerTextHandlers sets up handlers for keyboard button presses
func registerTextHandlers(bot *teleflow.Bot, permissionChecker *AdminAccessManager) {
	// Handle all text messages (keyboard button presses and regular text)
	bot.HandleText(func(ctx *teleflow.Context) error {
		if ctx.Update.Message == nil {
			return ctx.Reply("❌ No message received")
		}

		text := ctx.Update.Message.Text

		// Handle specific keyboard button presses
		switch text {
		case "🏠 Home":
			return handleHomeButton(ctx, permissionChecker)
		case "ℹ️ Help":
			return handleHelpButton(ctx, permissionChecker)
		case "⚡ Spam Test":
			return handleSpamTestButton(ctx, permissionChecker)
		case "💥 Panic Test":
			return handlePanicTestButton(ctx, permissionChecker)
		case "👑 Admin Panel":
			return handleAdminPanelButton(ctx, permissionChecker)
		case "💥 YES - Trigger Panic":
			return handleActualPanic(ctx)
		case "🚫 Cancel":
			return handleCancelPanic(ctx, permissionChecker)
		default:
			return handleUnknownText(ctx, text, permissionChecker)
		}
	})
}

// handleHomeButton returns to main menu with middleware status
func handleHomeButton(ctx *teleflow.Context, checker *AdminAccessManager) error {
	userType := "Regular User"
	if checker.IsAdmin(ctx.UserID()) {
		userType = "Admin User"
	}

	homeText := "🏠 **Middleware Demo - Main Menu**\n\n" +
		"🎯 **Current Status:**\n" +
		"• User ID: `" + strconv.FormatInt(ctx.UserID(), 10) + "`\n" +
		"• Role: " + userType + "\n" +
		"• All middleware layers active\n\n" +
		"🛡️ **Active Protection:**\n" +
		"✅ Panic recovery enabled\n" +
		"✅ Request logging active\n" +
		"✅ Rate limiting (2/min) enforced\n" +
		"✅ Authorization checks enabled\n\n" +
		"Use the buttons below to test different middleware features!"

	keyboard := checker.GetReplyKeyboard(ctx.GetMenuContext())
	return ctx.Reply(homeText, keyboard)
}

// handleHelpButton shows middleware help information
func handleHelpButton(ctx *teleflow.Context, checker *AdminAccessManager) error {
	helpText := "ℹ️ **Quick Middleware Guide**\n\n" +
		"🚀 **How to Test Each Middleware:**\n\n" +
		"1️⃣ **Rate Limiting:**\n" +
		"   • Tap 'Spam Test' rapidly\n" +
		"   • Try sending multiple messages quickly\n" +
		"   • You'll be blocked after 2 requests/minute\n\n" +
		"2️⃣ **Authorization:**\n" +
		"   • Try 'Admin Panel' button\n" +
		"   • Non-admins will be denied access\n" +
		"   • Set ADMIN_USERS env var to become admin\n\n" +
		"3️⃣ **Panic Recovery:**\n" +
		"   • Tap 'Panic Test' to trigger controlled panic\n" +
		"   • Bot will recover gracefully\n" +
		"   • Check server logs for panic details\n\n" +
		"4️⃣ **Logging:**\n" +
		"   • All actions are automatically logged\n" +
		"   • Check server console for detailed logs\n" +
		"   • Includes timing and status information\n\n" +
		"💡 **Tip:** Run with verbose logging to see all middleware actions!"

	keyboard := checker.GetReplyKeyboard(ctx.GetMenuContext())
	return ctx.Reply(helpText, keyboard)
}

// handleSpamTestButton initiates rate limiting test
func handleSpamTestButton(ctx *teleflow.Context, checker *AdminAccessManager) error {
	spamText := "⚡ **Rate Limiting Active Test**\n\n" +
		"🎯 **Current Request Status:**\n" +
		"• This message: **DELIVERED** ✅\n" +
		"• Rate limit: 2 requests per minute\n" +
		"• Your next request available in: ~30 seconds\n\n" +
		"🧪 **Quick Test:**\n" +
		"1. Tap this button again immediately\n" +
		"2. Try sending any other command\n" +
		"3. You should get rate limited\n\n" +
		"📊 **RateLimitMiddleware Details:**\n" +
		"• Tracks per-user request timestamps\n" +
		"• Calculates minimum intervals\n" +
		"• Blocks excess requests with clear feedback\n" +
		"• Thread-safe with mutex protection"

	keyboard := checker.GetReplyKeyboard(ctx.GetMenuContext())
	return ctx.Reply(spamText, keyboard)
}

// handlePanicTestButton shows panic test information
func handlePanicTestButton(ctx *teleflow.Context, checker *AdminAccessManager) error {
	panicText := "💥 **Panic Recovery Demonstration**\n\n" +
		"⚠️ **Safety Notice:**\n" +
		"This is a controlled panic test to demonstrate RecoveryMiddleware.\n\n" +
		"🔬 **What Will Happen:**\n" +
		"1. Handler will call `panic()` with test message\n" +
		"2. RecoveryMiddleware catches the panic\n" +
		"3. Panic details logged to server console\n" +
		"4. You receive friendly error message\n" +
		"5. Bot continues running normally\n\n" +
		"🛡️ **RecoveryMiddleware Benefits:**\n" +
		"• Prevents bot crashes from handler panics\n" +
		"• Provides user-friendly error responses\n" +
		"• Logs detailed panic information\n" +
		"• Maintains bot availability\n\n" +
		"Ready to trigger the test panic?"

	keyboard := teleflow.NewReplyKeyboard()
	keyboard.AddButton("💥 YES - Trigger Panic").AddButton("🚫 Cancel").AddRow()
	keyboard.Resize()

	return ctx.Reply(panicText, keyboard)
}

// handleAdminPanelButton tests admin authorization
func handleAdminPanelButton(ctx *teleflow.Context, checker *AdminAccessManager) error {
	if !checker.IsAdmin(ctx.UserID()) {
		return ctx.Reply("🚫 **Authorization Failed**\n\n" +
			"❌ **Access Denied by AuthMiddleware:**\n" +
			"• Command: Admin Panel Access\n" +
			"• Your User ID: `" + strconv.FormatInt(ctx.UserID(), 10) + "`\n" +
			"• Required Permission: `admin_access`\n" +
			"• Your Permission Level: `basic_access` only\n\n" +
			"🔐 **Authorization Details:**\n" +
			"• AuthMiddleware performed permission check\n" +
			"• Custom permission checker consulted\n" +
			"• Admin role required but not found\n" +
			"• Request blocked before reaching handler\n\n" +
			"💡 **To Gain Admin Access:**\n" +
			"Set environment variable: `ADMIN_USERS=" + strconv.FormatInt(ctx.UserID(), 10) + "`")
	}

	adminText := "👑 **Admin Panel - Access Granted**\n\n" +
		"🎉 **Authorization Successful:**\n" +
		"• User ID: `" + strconv.FormatInt(ctx.UserID(), 10) + "`\n" +
		"• Permission Level: `admin_access` ✅\n" +
		"• AuthMiddleware check: **PASSED**\n\n" +
		"🔧 **Admin Capabilities Unlocked:**\n" +
		"• Full middleware monitoring\n" +
		"• User permission management\n" +
		"• System configuration access\n" +
		"• Advanced debugging tools\n\n" +
		"📊 **Current System Status:**\n" +
		"✅ All middleware layers operational\n" +
		"✅ Rate limiting enforced for all users\n" +
		"✅ Panic recovery system active\n" +
		"✅ Authorization checks functioning\n\n" +
		"⚡ **Performance Metrics:**\n" +
		"• Request logging: All messages tracked\n" +
		"• Rate limiting: 2 requests/minute enforced\n" +
		"• Recovery: 0 unhandled panics\n" +
		"• Auth: Role-based access working"

	keyboard := checker.GetReplyKeyboard(ctx.GetMenuContext())
	return ctx.Reply(adminText, keyboard)
}

// handleActualPanic deliberately triggers a panic to test RecoveryMiddleware
func handleActualPanic(ctx *teleflow.Context) error {
	// Log the intentional panic for demonstration
	log.Printf("User %d triggered intentional panic for RecoveryMiddleware demonstration", ctx.UserID())

	// This panic will be caught by RecoveryMiddleware
	panic("🧪 DEMONSTRATION PANIC: This is an intentional panic triggered by user " + strconv.FormatInt(ctx.UserID(), 10) + " to test RecoveryMiddleware functionality. The middleware should catch this panic, log it, and return a user-friendly error message.")
}

// handleCancelPanic cancels the panic test
func handleCancelPanic(ctx *teleflow.Context, checker *AdminAccessManager) error {
	cancelText := "🚫 **Panic Test Cancelled**\n\n" +
		"✅ **Smart Choice!**\n" +
		"You cancelled the panic test. No panic was triggered.\n\n" +
		"🛡️ **RecoveryMiddleware Information:**\n" +
		"• Would have caught any panic\n" +
		"• Logs panic details to console\n" +
		"• Returns friendly error to user\n" +
		"• Keeps bot running smoothly\n\n" +
		"💡 **Alternative Testing:**\n" +
		"You can still test other middleware features:\n" +
		"• Rate limiting with spam test\n" +
		"• Authorization with admin panel\n" +
		"• Logging (happens automatically)"

	keyboard := checker.GetReplyKeyboard(ctx.GetMenuContext())
	return ctx.Reply(cancelText, keyboard)
}

// handleUnknownText processes any text that doesn't match known buttons
func handleUnknownText(ctx *teleflow.Context, text string, checker *AdminAccessManager) error {
	responseText := "🤔 **Unknown Message Received**\n\n" +
		"📝 **Logging Middleware Captured:**\n" +
		"• Message: \"" + text + "\"\n" +
		"• User ID: `" + strconv.FormatInt(ctx.UserID(), 10) + "`\n" +
		"• Type: Text message\n" +
		"• Timestamp: Logged automatically\n\n" +
		"🎯 **Available Actions:**\n" +
		"• Use keyboard buttons below\n" +
		"• Try commands: `/start`, `/help`, `/admin`, `/spam`, `/panic`\n" +
		"• Test middleware features with buttons\n\n" +
		"💡 **Middleware Status:**\n" +
		"✅ This message was processed by all middleware\n" +
		"✅ Rate limiting checked (if not exceeded)\n" +
		"✅ Authorization verified for basic access\n" +
		"✅ Logging recorded automatically"

	keyboard := checker.GetReplyKeyboard(ctx.GetMenuContext())
	return ctx.Reply(responseText, keyboard)
}
