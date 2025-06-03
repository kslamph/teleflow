package main

import (
	"log"
	"os"

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
			// Inject services into context
			ctx.Set("userService", userService)
			return next(ctx)
		}
	})

	// Register command handlers
	handlers.RegisterCommands(bot)

	// Register text message handlers
	handlers.RegisterTextHandlers(bot)

	// Register callback handlers
	handlers.RegisterCallbacks(bot)

	// Register flow handlers
	handlers.RegisterFlows(bot)

	// Start the bot
	log.Println("Starting Advanced User Manager Bot...")

	if err := bot.Start(); err != nil {
		log.Fatal("Failed to start bot:", err)
	}
}
