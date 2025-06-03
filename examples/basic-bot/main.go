package main

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	// Get bot token from environment variable
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable is required")
	}

	// Create new bot instance
	bot, err := teleflow.NewBot(token)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Add logging middleware to track all interactions
	bot.Use(teleflow.LoggingMiddleware())

	// Register command handlers
	registerCommands(bot)

	// Register text handlers for keyboard buttons
	registerTextHandlers(bot)

	// Start the bot
	log.Println("Starting basic bot...")
	if err := bot.Start(); err != nil {
		log.Fatal("Failed to start bot:", err)
	}
}

// registerCommands sets up all command handlers
func registerCommands(bot *teleflow.Bot) {
	// /start command - Welcome message with reply keyboard
	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		welcomeText := "🤖 Welcome to the Basic Teleflow Bot!\n\n" +
			"This bot demonstrates basic API usage including:\n" +
			"• Command handling (/start, /help, /ping)\n" +
			"• Reply keyboards with buttons\n" +
			"• Text message handling\n" +
			"• Logging middleware\n\n" +
			"Use the buttons below or type commands directly!"

		// Create reply keyboard with main menu buttons
		keyboard := createMainKeyboard()
		messageReceived := ctx.Update.Message

		chatidText := strconv.FormatInt(messageReceived.Chat.ID, 10)

		welcomeText += "\n\n" +
			"Message received at: " + time.Unix(int64(messageReceived.Date), 0).Format("2006-01-02 15:04:05") + "\n" +
			"From: " + messageReceived.From.String() + "\n" +
			"Chat ID: " + chatidText + "\n" +
			"Message: " + messageReceived.Text + "\n"

		queryParams := toQueryParams(messageReceived.Text)
		if orderID, ok := queryParams["orderid"]; ok {
			welcomeText += " (Order ID: " + orderID + ")"
		}
		welcomeText += "\n\n"
		return ctx.Reply(welcomeText, keyboard)
	})

	// /help command - Help information
	bot.HandleCommand("help", func(ctx *teleflow.Context) error {
		helpText := "🆘 **Teleflow Basic Bot Help**\n\n" +
			"**Available Commands:**\n" +
			"• `/start` - Show welcome message and main menu\n" +
			"• `/help` - Show this help information\n" +
			"• `/ping` - Test bot responsiveness\n\n" +
			"**Keyboard Buttons:**\n" +
			"• 🏠 Home - Return to main menu\n" +
			"• ℹ️ Info - Show bot information\n" +
			"• ❓ Help - Show help text\n\n" +
			"**Features Demonstrated:**\n" +
			"• Command handling with HandleCommand\n" +
			"• Text handling with HandleText\n" +
			"• Reply keyboards with custom buttons\n" +
			"• Logging middleware for request tracking\n" +
			"• Environment variable configuration\n" +
			"• Proper error handling patterns"

		keyboard := createMainKeyboard()
		return ctx.Reply(helpText, keyboard)
	})

	// /ping command - Simple ping response
	bot.HandleCommand("ping", func(ctx *teleflow.Context) error {
		return ctx.Reply("🏓 Pong! Bot is working correctly.", createMainKeyboard())
	})
}

// registerTextHandlers sets up handlers for keyboard button presses
func registerTextHandlers(bot *teleflow.Bot) {
	// Handle specific keyboard button presses
	bot.HandleText("🏠 Home", handleHomeButton)
	bot.HandleText("ℹ️ Info", handleInfoButton)
	bot.HandleText("❓ Help", handleHelpButton)

	// Handle all other text messages (default handler)
	bot.HandleText("", func(ctx *teleflow.Context) error {
		// Get the text from the message
		if ctx.Update.Message == nil {
			return ctx.Reply("❌ No message received")
		}
		text := ctx.Update.Message.Text
		return handleUnknownText(ctx, text)
	})
}

// handleHomeButton processes the Home keyboard button
func handleHomeButton(ctx *teleflow.Context) error {
	homeText := "🏠 **Main Menu**\n\n" +
		"Welcome back to the main menu! This basic bot demonstrates:\n\n" +
		"✅ **Basic API Usage:**\n" +
		"• Bot creation and configuration\n" +
		"• Command registration with HandleCommand\n" +
		"• Text handling with HandleText\n" +
		"• Middleware integration\n\n" +
		"✅ **Interactive Features:**\n" +
		"• Reply keyboards with custom buttons\n" +
		"• Command and text message routing\n" +
		"• Environment-based configuration\n\n" +
		"Use the buttons below or type commands like `/ping`, `/help`!"

	keyboard := createMainKeyboard()
	return ctx.Reply(homeText, keyboard)
}

// handleInfoButton processes the Info keyboard button
func handleInfoButton(ctx *teleflow.Context) error {
	infoText := "ℹ️ **Bot Information**\n\n" +
		"**Teleflow Basic Bot v1.0**\n\n" +
		"🔧 **Technical Details:**\n" +
		"• Built with Teleflow framework\n" +
		"• Uses telegram-bot-api for Telegram integration\n" +
		"• Implements logging middleware for request tracking\n" +
		"• Demonstrates reply keyboard functionality\n" +
		"• Shows proper error handling patterns\n\n" +
		"📋 **Implemented Features:**\n" +
		"• Command handlers (/start, /help, /ping)\n" +
		"• Text message routing for keyboard buttons\n" +
		"• Reply keyboards with emoji buttons\n" +
		"• Middleware usage for logging\n" +
		"• Environment variable configuration\n\n" +
		"🎯 **Purpose:**\n" +
		"This bot serves as a beginner-friendly example showing how to use the basic Teleflow API for creating interactive Telegram bots."

	keyboard := createMainKeyboard()
	return ctx.Reply(infoText, keyboard)
}

// handleHelpButton processes the Help keyboard button
func handleHelpButton(ctx *teleflow.Context) error {
	// Reuse the same help logic as the /help command
	helpText := "❓ **Quick Help**\n\n" +
		"**How to Use This Bot:**\n\n" +
		"1️⃣ **Commands** - Type commands directly:\n" +
		"   • `/start` - Reset to main menu\n" +
		"   • `/help` - Detailed help information\n" +
		"   • `/ping` - Test bot responsiveness\n\n" +
		"2️⃣ **Buttons** - Tap the keyboard buttons:\n" +
		"   • 🏠 Home - Return to main menu\n" +
		"   • ℹ️ Info - Bot technical information\n" +
		"   • ❓ Help - This quick help\n\n" +
		"3️⃣ **Features Shown:**\n" +
		"   • Command handling with middleware\n" +
		"   • Reply keyboards for better UX\n" +
		"   • Text message processing\n" +
		"   • Proper bot lifecycle management\n\n" +
		"💡 **Tip:** This bot demonstrates core Teleflow concepts that you can build upon for more complex applications!"

	keyboard := createMainKeyboard()
	return ctx.Reply(helpText, keyboard)
}

// handleUnknownText processes any text that doesn't match known buttons
func handleUnknownText(ctx *teleflow.Context, text string) error {
	responseText := "🤔 I received your message: \"" + text + "\"\n\n" +
		"This basic bot recognizes:\n" +
		"• Commands: `/start`, `/help`, `/ping`\n" +
		"• Keyboard buttons: 🏠 Home, ℹ️ Info, ❓ Help\n\n" +
		"Try using the buttons below or type `/help` for more information!"

	keyboard := createMainKeyboard()
	return ctx.Reply(responseText, keyboard)
}

// createMainKeyboard creates the main reply keyboard with all buttons
func createMainKeyboard() *teleflow.ReplyKeyboard {
	// Create a new reply keyboard
	keyboard := teleflow.NewReplyKeyboard()

	// Add first row with Home and Info buttons
	keyboard.AddButton("🏠 Home").AddButton("ℹ️ Info").AddRow()

	// Add second row with Help button
	keyboard.AddButton("❓ Help").AddRow()

	// Configure keyboard properties
	keyboard.Resize() // Make keyboard smaller

	return keyboard
}

func toQueryParams(fullText string) map[string]string {
	// Parse the full text into query parameters
	params := make(map[string]string)
	if fullText == "" {
		return params
	}
	parts := strings.SplitN(fullText, "?", 2)
	var queryString string
	if len(parts) > 1 {
		queryString = parts[1] // "orderid=319"
	}

	// Split by spaces to get key-value pairs
	pairs, err := url.ParseQuery(queryString)
	if err != nil {
		log.Println("Error parsing query parameters:", err)
		return params
	}
	for key, values := range pairs {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}
