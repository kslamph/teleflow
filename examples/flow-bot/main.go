package main

import (
	"fmt"
	"log"
	"os"

	teleflow "github.com/kslamph/teleflow/core"
)

// SimplePermissionChecker provides a basic permission implementation for demonstration
type SimplePermissionChecker struct{}

// CanExecute always returns true for this simple example
func (spc *SimplePermissionChecker) CanExecute(userID int64, action string) bool {
	// In a real implementation, you would check user permissions here
	return true
}

// GetMainMenuForUser returns a basic main menu keyboard for the user
func (spc *SimplePermissionChecker) GetMainMenuForUser(userID int64) *teleflow.ReplyKeyboard {
	return teleflow.NewReplyKeyboard(
		[]teleflow.ReplyKeyboardButton{
			{Text: "💸 Transfer"},
			{Text: "📊 Balance"},
		},
		[]teleflow.ReplyKeyboardButton{
			{Text: "❓ Help"},
			{Text: "⚙️ Settings"},
		},
	)
}

func main() {
	// Get bot token from environment variable
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable is required")
	}

	// Create permission checker
	permissionChecker := &SimplePermissionChecker{}

	// Create bot with configuration
	bot, err := teleflow.NewBot(token,
		teleflow.WithUserPermissions(permissionChecker),
		teleflow.WithFlowConfig(teleflow.FlowConfig{
			ExitCommands:        []string{"/cancel", "/exit"},
			ExitMessage:         "❌ Operation cancelled.",
			AllowGlobalCommands: true,
			HelpCommands:        []string{"/help"},
		}),
	)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Add middleware in the correct order
	bot.Use(teleflow.LoggingMiddleware())
	bot.Use(teleflow.AuthMiddleware(permissionChecker))
	bot.Use(teleflow.RecoveryMiddleware())

	// Register transfer flow - adapted from newdesign.md lines 1050-1100
	transferFlow := teleflow.NewFlow("transfer").
		// Step 1: Amount input with number validation
		Step("amount").
		OnInput(func(ctx *teleflow.Context) error {
			return ctx.Reply("💰 Please enter the amount to transfer:")
		}).
		WithValidator(teleflow.NumberValidator()).

		// Step 2: Recipient input (text)
		Step("recipient").
		OnInput(func(ctx *teleflow.Context) error {
			amount := ctx.Update.Message.Text
			ctx.Set("amount", amount)
			return ctx.Reply("👤 Please enter the recipient's username:")
		}).

		// Step 3: Confirmation with choice validation (yes/no)
		Step("confirm").
		OnInput(func(ctx *teleflow.Context) error {
			recipient := ctx.Update.Message.Text
			ctx.Set("recipient", recipient)
			amount, _ := ctx.Get("amount")

			// Create inline keyboard for confirmation
			keyboard := teleflow.NewInlineKeyboard(
				[]teleflow.InlineKeyboardButton{
					{Text: "✅ Confirm", CallbackData: "yes"},
					{Text: "❌ Cancel", CallbackData: "no"},
				},
			)

			return ctx.Reply(
				fmt.Sprintf("💸 Transfer $%s to %s?\n\nPlease confirm your transfer:", amount, recipient),
				keyboard,
			)
		}).
		WithValidator(teleflow.ChoiceValidator([]string{"yes", "no"})).

		// Flow completion handler
		OnComplete(func(ctx *teleflow.Context) error {
			amount, _ := ctx.Get("amount")
			recipient, _ := ctx.Get("recipient")

			return ctx.EditOrReply(fmt.Sprintf(
				"✅ Transfer completed successfully!\n\n"+
					"💰 Amount: $%s\n"+
					"👤 Recipient: %s\n"+
					"⏰ Status: Processed",
				amount, recipient,
			))
		}).

		// Flow cancellation handler
		OnCancel(func(ctx *teleflow.Context) error {
			return ctx.EditOrReply("❌ Transfer cancelled.")
		}).
		Build()

	// Register the transfer flow with the bot
	bot.RegisterFlow(transferFlow)

	// Command handlers

	// /start command - Welcome message with transfer flow option
	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		return ctx.Reply(
			"🤖 Welcome to Teleflow Bot!\n\n" +
				"This bot demonstrates the multi-step conversation system with a transfer flow.\n\n" +
				"Available commands:\n" +
				"💸 /transfer - Start a money transfer\n" +
				"❌ /cancel - Cancel current operation\n" +
				"❓ /help - Show this help message\n\n" +
				"You can also use the buttons below:",
		)
	})

	// /transfer command - Start the transfer flow
	bot.HandleCommand("transfer", func(ctx *teleflow.Context) error {
		log.Printf("Starting transfer flow for user %d", ctx.UserID())
		if err := ctx.StartFlow("transfer"); err != nil {
			return ctx.Reply("❌ Failed to start transfer flow. Please try again.")
		}
		// The flow will start automatically and show the first step
		return nil
	})

	// /cancel command - Cancel current flow
	bot.HandleCommand("cancel", func(ctx *teleflow.Context) error {
		return ctx.Reply("❌ Operation cancelled. Use this command during active flows.")
	})

	// /help command - Show help information
	bot.HandleCommand("help", func(ctx *teleflow.Context) error {
		return ctx.Reply(
			"❓ **Teleflow Bot Help**\n\n" +
				"**Available Commands:**\n" +
				"💸 /transfer - Start a money transfer flow\n" +
				"❌ /cancel - Cancel current operation\n" +
				"❓ /help - Show this help message\n\n" +
				"**Transfer Flow:**\n" +
				"1️⃣ Enter amount (numbers only)\n" +
				"2️⃣ Enter recipient username\n" +
				"3️⃣ Confirm the transfer\n\n" +
				"**Features:**\n" +
				"✅ Input validation at each step\n" +
				"✅ Flow cancellation support\n" +
				"✅ Error handling and user feedback\n" +
				"✅ Complete flow system demonstration\n\n" +
				"You can cancel any operation at any time using /cancel.",
		)
	})

	// Text handlers for button interactions

	// Handle "💸 Transfer" button
	bot.HandleText(func(ctx *teleflow.Context) error {
		if ctx.Update.Message != nil && ctx.Update.Message.Text == "💸 Transfer" {
			log.Printf("Starting transfer flow via button for user %d", ctx.UserID())
			if err := ctx.StartFlow("transfer"); err != nil {
				return ctx.Reply("❌ Failed to start transfer flow. Please try again.")
			}
			return nil
		}

		// Handle "📊 Balance" button
		if ctx.Update.Message != nil && ctx.Update.Message.Text == "📊 Balance" {
			return ctx.Reply("💳 Your current balance: $1,234.56")
		}

		// Handle "❓ Help" button
		if ctx.Update.Message != nil && ctx.Update.Message.Text == "❓ Help" {
			return ctx.Reply(
				"❓ **Teleflow Bot Help**\n\n" +
					"**Available Commands:**\n" +
					"💸 /transfer - Start a money transfer flow\n" +
					"❌ /cancel - Cancel current operation\n" +
					"❓ /help - Show this help message\n\n" +
					"**Transfer Flow:**\n" +
					"1️⃣ Enter amount (numbers only)\n" +
					"2️⃣ Enter recipient username\n" +
					"3️⃣ Confirm the transfer\n\n" +
					"You can cancel any operation at any time using /cancel.",
			)
		}

		// Handle "⚙️ Settings" button
		if ctx.Update.Message != nil && ctx.Update.Message.Text == "⚙️ Settings" {
			return ctx.Reply("⚙️ Settings feature coming soon!")
		}

		// Default response for unrecognized text
		return ctx.Reply("ℹ️ I don't understand that command. Type /help for available commands.")
	})

	// Callback handlers for inline keyboard buttons in the flow
	bot.RegisterCallback(teleflow.SimpleCallback("yes", func(ctx *teleflow.Context, data string) error {
		// This will be handled by the flow system automatically
		// since we're using ChoiceValidator with "yes" and "no"
		return nil
	}))

	bot.RegisterCallback(teleflow.SimpleCallback("no", func(ctx *teleflow.Context, data string) error {
		// This will trigger flow cancellation
		return nil
	}))

	log.Println("🚀 Flow Bot starting...")
	log.Println("📱 Features:")
	log.Println("   ✅ Multi-step transfer flow (amount → recipient → confirm)")
	log.Println("   ✅ Number validation for amount input")
	log.Println("   ✅ Choice validation for confirmation step")
	log.Println("   ✅ Flow cancellation with /cancel command")
	log.Println("   ✅ Logging middleware for request tracking")
	log.Println("   ✅ Auth middleware with permission checking")
	log.Println("   ✅ Recovery middleware for error handling")
	log.Println("   ✅ Complete flow system demonstration")
	log.Println("")
	log.Println("💡 Try the following:")
	log.Println("   1. Send /start to see the welcome message")
	log.Println("   2. Send /transfer to start the transfer flow")
	log.Println("   3. Follow the prompts: amount → recipient → confirm")
	log.Println("   4. Use /cancel to cancel the flow at any step")
	log.Println("   5. Use /help to see available commands")
	log.Println("")

	// Start the bot
	if err := bot.Start(); err != nil {
		log.Fatal("Failed to start bot:", err)
	}
}
