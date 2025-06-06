package main

import (
	"fmt"
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

	bot, err := teleflow.NewBot(token)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Initialize the flow system
	bot.InitializeFlowSystem()

	// Create a user registration flow using the new Step-Prompt-Process API
	registrationFlow, err := teleflow.NewFlow("user_registration").
		Step("welcome").
		Prompt(
			"👋 Welcome! Let's get you registered. What's your name?",
			nil, // No image
			nil, // No keyboard
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if input == "" {
				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{
					Message: "Please enter your name:",
				})
			}

			// Store the name
			ctx.Set("user_name", input)
			return teleflow.NextStep().WithPrompt(&teleflow.PromptConfig{
				Message: "✅ Name saved! Moving to the next step...",
			})
		}).
		Step("age").
		Prompt(
			func(ctx *teleflow.Context) string {
				name, _ := ctx.Get("user_name")
				return fmt.Sprintf("Nice to meet you, %s! How old are you?", name)
			},
			nil, // No image
			nil, // No keyboard
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Simple age validation
			if len(input) == 0 || len(input) > 3 {
				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{
					Message: "Please enter a valid age (1-3 digits):",
				})
			}

			// Store the age
			ctx.Set("user_age", input)
			return teleflow.NextStep().WithPrompt(&teleflow.PromptConfig{
				Message: "✅ Age recorded! Let's confirm your details...",
			})
		}).
		Step("confirmation").
		Prompt(
			func(ctx *teleflow.Context) string {
				name, _ := ctx.Get("user_name")
				age, _ := ctx.Get("user_age")
				return fmt.Sprintf("Great! So your name is %s and you're %s years old. Is this correct?", name, age)
			},
			nil, // No image
			func(ctx *teleflow.Context) map[string]interface{} {
				return map[string]interface{}{
					"✅ Yes, that's correct":  "confirm",
					"❌ No, let me try again": "restart",
				}
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
				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{
					Message: "Please click one of the buttons above.",
				})
			}
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			name, _ := ctx.Get("user_name")
			age, _ := ctx.Get("user_age")

			// Use SendPrompt instead of Reply for consistent rendering
			return ctx.SendPrompt(&teleflow.PromptConfig{
				Message: fmt.Sprintf("🎉 Registration complete!\nName: %s\nAge: %s\n\nWelcome to our service!", name, age),
				Image:   nil, // Could add a welcome image here
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

	// Handle the register command as an alternative
	bot.HandleCommand("register", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.StartFlow("user_registration")
	})

	// Handle cancellation
	bot.HandleCommand("cancel", func(ctx *teleflow.Context, command string, args string) error {
		if ctx.IsUserInFlow() {
			ctx.CancelFlow()
			return ctx.Reply("❌ Registration cancelled.")
		}
		return ctx.Reply("You're not currently in any process.")
	})

	// Start the bot
	log.Println("🤖 Bot starting with new Step-Prompt-Process API...")
	bot.Start()
}
