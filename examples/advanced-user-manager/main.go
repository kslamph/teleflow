package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	teleflow "github.com/kslamph/teleflow/core"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/handlers"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/services"
	"github.com/kslamph/teleflow/examples/advanced-user-manager/templates"
)

func main() {
	// Get bot token from environment
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	// Configure debug and logging
	debug := false
	if debugStr := os.Getenv("DEBUG"); debugStr != "" {
		if d, err := strconv.ParseBool(debugStr); err == nil {
			debug = d
		}
	}

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info"
	}

	if debug {
		log.Printf("Debug mode enabled, Log level: %s", logLevel)
	}

	// Initialize services
	userService := services.NewUserService()
	accessManager := services.NewAccessManager()

	// Create bot with access manager
	bot, err := teleflow.NewBot(token,
		teleflow.WithAccessManager(accessManager), // to provide UI main buttons and reply keyboard
	)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Register template system
	templates.RegisterTemplates(bot)
	templates.RegisterTemplateFunctions(bot)

	// Add middleware in the correct order
	bot.Use(teleflow.LoggingMiddleware())
	bot.Use(teleflow.AuthMiddleware(accessManager)) // to handle user authentication
	bot.Use(teleflow.RecoveryMiddleware())

	// Setup middleware to inject services into context
	bot.Use(func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
		return func(ctx *teleflow.Context) error {
			// Inject services into context (NOT accessManager - it's used by AuthMiddleware)
			ctx.Set("userService", userService)
			ctx.Set("debug", debug)
			ctx.Set("logLevel", logLevel)
			return next(ctx)
		}
	})

	// Register command handlers
	handlers.RegisterCommands(bot, accessManager)

	// Register text message handlers
	handlers.RegisterTextHandlers(bot, accessManager)

	// Register callback handlers
	handlers.RegisterCallbacks(bot, accessManager, userService)

	// Register flow handlers
	handlers.RegisterFlows(bot, userService)

	// Add a default text handler for unrecognized messages
	bot.HandleText("", func(ctx *teleflow.Context) error {
		if debug {
			log.Printf("Unrecognized message from user %d: %s", ctx.UserID(), ctx.Update.Message.Text)
		}

		// Create main keyboard
		keyboard := &teleflow.ReplyKeyboard{
			Keyboard: [][]teleflow.ReplyKeyboardButton{
				{
					{Text: "👥 User Manager"},
					{Text: "❓ Help"},
				},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: false,
		}

		return ctx.Reply("I didn't understand that. Please use the buttons below or type /help for assistance.", keyboard)
	})

	// Start the bot
	log.Println("Starting Advanced User Manager Bot...")

	if err := bot.Start(); err != nil {
		log.Fatal("Failed to start bot:", err)
	}
}
