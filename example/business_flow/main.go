package main

import (
	"log"
	"os"

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	// Initialize bot
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Initialize business service
	businessService := NewBusinessService()

	// Create main menu keyboard
	mainMenuKeyboard := teleflow.NewReplyKeyboard().
		AddButton("💼 Account Info").
		AddButton("💸 Transfer Funds").
		Row().
		AddButton("🛒 Place Order").
		Resize().
		Build()

	// Initialize AccessManager
	accessManager := &BusinessAccessManager{
		mainMenu: mainMenuKeyboard,
	}

	// Create bot with flow config and access manager
	bot, err := teleflow.NewBot(token,
		teleflow.WithFlowConfig(teleflow.FlowConfig{
			ExitCommands:        []string{"/cancel"},
			ExitMessage:         "🚫 Operation cancelled.",
			AllowGlobalCommands: false,
		}),
		teleflow.WithAccessManager(accessManager),
	)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Initialize template manager and register templates
	registerTemplates()

	// Set bot commands
	err = bot.SetBotCommands(map[string]string{
		"start":  "Welcome & Intro",
		"help":   "Show Help",
		"cancel": "Cancel Current Action",
	})
	if err != nil {
		log.Printf("Failed to set bot commands: %v", err)
	}

	// Register command handlers
	bot.HandleCommand("start", handleStartCommand)
	bot.HandleCommand("help", handleHelpCommand)

	// Register flows
	registerFlows(bot, businessService)

	// Register text handlers for main menu
	bot.HandleText("💼 Account Info", func(ctx *teleflow.Context, text string) error {
		return ctx.StartFlow("account_info")
	})

	bot.HandleText("💸 Transfer Funds", func(ctx *teleflow.Context, text string) error {
		return ctx.StartFlow("transfer_funds")
	})

	bot.HandleText("🛒 Place Order", func(ctx *teleflow.Context, text string) error {
		return ctx.StartFlow("place_order")
	})

	log.Println("Bot started successfully...")
	bot.Start()
}

// BusinessAccessManager implements AccessManager interface
type BusinessAccessManager struct {
	mainMenu *teleflow.ReplyKeyboard
}

func (m *BusinessAccessManager) CheckPermission(ctx *teleflow.PermissionContext) error {
	return nil // Allow all actions for this example
}

func (m *BusinessAccessManager) GetReplyKeyboard(ctx *teleflow.PermissionContext) *teleflow.ReplyKeyboard {
	return m.mainMenu
}

func handleStartCommand(ctx *teleflow.Context, command string, args string) error {
	// Generate welcome image
	imageBytes, err := GenerateImage("Welcome to Teleflow Business Bot!", 600, 300)
	if err != nil {
		log.Printf("Failed to generate welcome image: %v", err)
	}

	templateData := map[string]interface{}{
		"UserID": ctx.UserID(),
	}

	return ctx.SendPrompt(&teleflow.PromptConfig{
		Message:      "template:start_message_template",
		Image:        imageBytes,
		TemplateData: templateData,
	})
}

func handleHelpCommand(ctx *teleflow.Context, command string, args string) error {
	return ctx.SendPrompt(&teleflow.PromptConfig{
		Message: "template:help_message_template",
	})
}
