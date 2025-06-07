package main

import (
	"fmt"
	"log"
	"os"

	teleflow "github.com/kslamph/teleflow/core"
)

// MyAccessManager implements AccessManager interface for custom keyboard management
type MyAccessManager struct {
	mainMenu *teleflow.ReplyKeyboard
}

// CheckPermission checks if user has permission for an action
func (m *MyAccessManager) CheckPermission(ctx *teleflow.PermissionContext) error {
	return nil // Allow all actions for this example
}

// GetReplyKeyboard returns appropriate keyboard based on context
func (m *MyAccessManager) GetReplyKeyboard(ctx *teleflow.PermissionContext) *teleflow.ReplyKeyboard {
	// Return main menu keyboard for all users
	return m.mainMenu
}

// GetMenuButton returns menu button configuration with bot commands
func (m *MyAccessManager) GetMenuButton(ctx *teleflow.PermissionContext) *teleflow.MenuButtonConfig {
	return teleflow.BuildMenuButton(map[string]string{
		"start":  "🚀 Start Registration",
		"demo":   "📸 Photo Demo",
		"help":   "❓ Help",
		"cancel": "❌ Cancel",
	})
}

func main() {
	// Initialize bot
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Create keyboards using proper API
	// Reply keyboard only contains non-command actions
	// mainMenuKeyboard := teleflow.NewReplyKeyboard(
	// 	[]teleflow.ReplyKeyboardButton{
	// 		{Text: "📝 Register"}, // Main action for registration
	// 	},
	// ).Resize()

	mainMenuKeyboard := teleflow.BuildReplyKeyboard([]string{"📝 Register", "🏠 Home", "⚙️ Settings", "❓ Help"}, 3).Resize()

	// Initialize AccessManager
	accessManager := &MyAccessManager{
		mainMenu: mainMenuKeyboard,
	}

	// Create bot with access manager
	// WithAccessManager automatically applies optimized middleware stack:
	// 1. RateLimitMiddleware (60 req/min) - blocks spam before expensive operations
	// 2. AuthMiddleware - permission checking after rate limiting
	//
	// UI Management:
	// - Menu Button: Contains commands (/start, /demo, /help, /cancel)
	// - Reply Keyboard: Contains action buttons (📝 Register)
	bot, err := teleflow.NewBot(token,
		teleflow.WithFlowConfig(teleflow.FlowConfig{
			ExitCommands:        []string{"/cancel", "/exit"},
			ExitMessage:         "🚫 操作已取消。", // Chinese custom exit message
			AllowGlobalCommands: false,
		}),
		teleflow.WithAccessManager(accessManager), // Auto-applies rate limiting + auth middleware
	)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Note: No need to manually call bot.UseMiddleware(teleflow.AuthMiddleware(accessManager))
	// The framework automatically applies the optimal middleware stack for security and performance
	// Create a user registration flow using the new Step-Prompt-Process API with photo capabilities
	registrationFlow, err := teleflow.NewFlow("user_registration").
		OnError(teleflow.OnErrorCancel()).
		Step("welcome").
		Prompt("👋 Welcome! Let's get you registered. What's your name?").
		WithImage(
			"data:image/png;base64,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", // Welcome image
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if input == "" {
				return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
					Message: "Please enter your name:",
					Image:   "https://via.placeholder.com/300x150/FF9800/white?text=Name+Required",
				})
			}

			// Store the name
			ctx.Set("user_name", input)
			return teleflow.NextStep().WithPrompt(&teleflow.PromptConfig{
				Message: "✅ Name saved! Moving to the next step...",
				// Image:   "https://via.placeholder.com/300x150/2196F3/white?text=Name+Saved",
			})
		}).
		Step("age").
		Prompt(
			func(ctx *teleflow.Context) string {
				name, _ := ctx.Get("user_name")
				return fmt.Sprintf("Nice to meet you, %s! How old are you?", name)
			}).
		WithImage(
			func(ctx *teleflow.Context) string {
				// Dynamic image based on context
				name, _ := ctx.Get("user_name")
				return fmt.Sprintf("https://via.placeholder.com/400x200/9C27B0/white?text=Hello+%s", name)
			}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Simple age validation
			if len(input) == 0 || len(input) > 3 {
				return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
					Message: "Please enter a valid age (1-3 digits):",
					Image:   "https://via.placeholder.com/300x150/F44336/white?text=Invalid+Age",
				})
			}

			// Store the age
			ctx.Set("user_age", input)
			return teleflow.NextStep().WithPrompt(&teleflow.PromptConfig{
				Message: "✅ Age recorded! Let's confirm your details...",
				Image:   "https://via.placeholder.com/300x150/4CAF50/white?text=Age+Saved",
			})
		}).
		Step("confirmation").
		Prompt(
			func(ctx *teleflow.Context) string {
				name, _ := ctx.Get("user_name")
				age, _ := ctx.Get("user_age")
				return fmt.Sprintf("Great! So your name is %s and you're %s years old. Is this correct?", name, age)
			}).
		WithInlineKeyboard(
			func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
				return teleflow.NewInlineKeyboard().
					ButtonCallback("✅ Yes, that's correct", "confirm").
					ButtonCallback("❌ No, let me try again", "restart")
			},
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {

			switch input {
			case "confirm":
				return teleflow.CompleteFlow().WithPrompt(&teleflow.PromptConfig{
					Message: "🎉 Perfect! Processing your registration...",
				})
			case "restart":
				return teleflow.GoToStep("age").WithPrompt(&teleflow.PromptConfig{
					Message: "🔄 No problem! Let's start over...",
				})
			default:
				return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
					Message: "Please click one of the buttons above.",
				})
			}
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			name, _ := ctx.Get("user_name")
			age, _ := ctx.Get("user_age")

			// Use SendPrompt with a celebration image
			return ctx.SendPrompt(&teleflow.PromptConfig{
				Message: fmt.Sprintf("🎉 Registration complete!\nName: %s\nAge: %s\n\nWelcome to our service!", name, age),
				Image:   "https://via.placeholder.com/500x300/4CAF50/white?text=🎉+Welcome+Aboard!",
			})
		}).
		Build()

	if err != nil {
		log.Fatal("Failed to build flow:", err)
	}

	// Register the flow
	bot.RegisterFlow(registrationFlow)

	// Handle the start command to begin registration
	bot.HandleCommand("start", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.StartFlow("user_registration")
	})

	// Handle the register button from reply keyboard
	bot.HandleText("📝 Register", func(ctx *teleflow.Context, text string) error {
		return ctx.StartFlow("user_registration")
	})

	// Demo command to showcase photo capabilities
	bot.HandleCommand("demo", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.SendPrompt(&teleflow.PromptConfig{
			Message: "📸 TeleFlow Photo Demo\n\nThis showcases different image capabilities:\n• Static URLs\n• Dynamic images based on context\n• Placeholder images for testing\n• Images with text and keyboards",
			Image:   "https://via.placeholder.com/600x400/673AB7/white?text=TeleFlow+Photo+Demo",
		})
	})

	// Help command to show available commands
	bot.HandleCommand("help", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.SendPrompt(&teleflow.PromptConfig{
			Message: "🤖 TeleFlow Bot Commands:\n\n/start - Begin user registration flow\n/register - Alternative registration command\n/demo - Photo capabilities demo\n/cancel - Cancel current flow\n/help - Show this help",
			Image:   "https://via.placeholder.com/500x250/2196F3/white?text=Help+Menu",
		})
	})
	bot.AddTemplate("not_understood", `❓ I didn't understand '{{.Input}}'`, teleflow.ParseModeMarkdownV2)
	bot.DefaultHandler(func(ctx *teleflow.Context, text string) error {
		return ctx.SendPrompt(&teleflow.PromptConfig{
			Message: "template:not_understood",
			TemplateData: map[string]interface{}{
				"Input": text,
			},
		})
	})

	// Start the bot
	log.Println("🤖 Bot starting with new Step-Prompt-Process API...")
	log.Fatal(bot.Start())
}
