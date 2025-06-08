package main

import (
	"log"
	"os"

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	// Initialize bot
	bot, err := teleflow.NewBot(token)
	if err != nil {
		log.Fatal(err)
	}

	// Example 1: Flow with OnProcessDeleteMessage - completely removes previous messages
	deleteMessageFlow, err := teleflow.NewFlow("delete_message_demo").
		OnProcessDeleteKeyboard(). // All button clicks will delete previous messages
		Step("menu").
		Prompt("Choose an option:").
		WithInlineKeyboard(func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
			return teleflow.NewInlineKeyboard().
				ButtonCallback("🔄 Refresh", "refresh").
				Row().
				ButtonCallback("📊 Stats", "stats").
				ButtonCallback("⚙️ Settings", "settings")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				switch buttonClick.Data {
				case "refresh":
					return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
						Message: "🔄 Refreshed! Choose again:",
					})
				case "stats":
					return teleflow.NextStep()
				case "settings":
					return teleflow.GoToStep("settings")
				}
			}
			return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
				Message: "Please use the buttons:",
			})
		}).
		Step("stats").
		Prompt("📊 Here are your stats. What would you like to do?").
		WithInlineKeyboard(func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
			return teleflow.NewInlineKeyboard().
				ButtonCallback("🔙 Back to Menu", "back").
				ButtonCallback("✅ Done", "done")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				switch buttonClick.Data {
				case "back":
					return teleflow.GoToStep("menu")
				case "done":
					return teleflow.CompleteFlow()
				}
			}
			return teleflow.Retry()
		}).
		Step("settings").
		Prompt("⚙️ Settings panel. Configure your preferences:").
		WithInlineKeyboard(func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
			return teleflow.NewInlineKeyboard().
				ButtonCallback("🔙 Back to Menu", "back").
				ButtonCallback("✅ Save & Exit", "save")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				switch buttonClick.Data {
				case "back":
					return teleflow.GoToStep("menu")
				case "save":
					return teleflow.CompleteFlow()
				}
			}
			return teleflow.Retry()
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			return ctx.SendPromptText("✅ Flow completed! All previous messages were deleted when you clicked buttons.")
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Example 2: Flow with OnProcessDeleteKeyboard - keeps messages but removes keyboards
	deleteKeyboardFlow, err := teleflow.NewFlow("delete_keyboard_demo").
		OnProcessDeleteKeyboard(). // Button clicks will remove keyboards from previous messages
		Step("welcome").
		Prompt("Welcome! This flow will disable previous keyboards when you click buttons.").
		WithInlineKeyboard(func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
			return teleflow.NewInlineKeyboard().
				ButtonCallback("➡️ Continue", "continue").
				ButtonCallback("❌ Cancel", "cancel")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				switch buttonClick.Data {
				case "continue":
					return teleflow.NextStep()
				case "cancel":
					return teleflow.CancelFlow()
				}
			}
			return teleflow.Retry()
		}).
		Step("confirm").
		Prompt("Are you sure you want to proceed?").
		WithInlineKeyboard(func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
			return teleflow.NewInlineKeyboard().
				ButtonCallback("✅ Yes, proceed", "yes").
				ButtonCallback("🔙 Go back", "back")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				switch buttonClick.Data {
				case "yes":
					return teleflow.CompleteFlow()
				case "back":
					return teleflow.GoToStep("welcome")
				}
			}
			return teleflow.Retry()
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			return ctx.SendPromptText("✅ Completed! Notice how previous message text remained but keyboards were disabled.")
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Example 3: Flow with default behavior - keeps everything untouched
	keepMessagesFlow, err := teleflow.NewFlow("keep_messages_demo").
		// No OnProcessDelete* methods called - default behavior keeps messages untouched
		Step("demo").
		Prompt("This flow keeps all messages and keyboards intact. Try scrolling back and clicking old buttons!").
		WithInlineKeyboard(func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
			return teleflow.NewInlineKeyboard().
				ButtonCallback("🔄 Refresh (keeps old keyboards)", "refresh").
				ButtonCallback("✅ Finish", "finish")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				switch buttonClick.Data {
				case "refresh":
					return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
						Message: "🔄 Refreshed! Old keyboards remain functional - you can scroll back and click them!",
					})
				case "finish":
					return teleflow.CompleteFlow()
				}
			}
			return teleflow.Retry()
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			return ctx.SendPromptText("✅ Demo completed! All previous keyboards remained functional throughout the flow.")
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Register flows
	bot.RegisterFlow(deleteMessageFlow)
	bot.RegisterFlow(deleteKeyboardFlow)
	bot.RegisterFlow(keepMessagesFlow)

	// Command handlers to start different demos
	bot.HandleCommand("delete_messages", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.StartFlow("delete_message_demo")
	})

	bot.HandleCommand("delete_keyboards", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.StartFlow("delete_keyboard_demo")
	})

	bot.HandleCommand("keep_messages", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.StartFlow("keep_messages_demo")
	})

	bot.HandleCommand("help", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.SendPromptText(`🎛️ **Message Handling Demo**

Choose a demo to see different keyboard behaviors:

/delete_messages - Delete entire previous messages on button clicks
/delete_keyboards - Remove only keyboards, keep message text
/keep_messages - Keep everything untouched (default behavior)

**The Problem Solved:**
Before this feature, when users scrolled back and clicked old keyboard buttons, they would trigger unexpected behavior because UUID mappings were cleaned up too early.

**The Solutions:**
• **OnProcessDeleteMessage()** - Clean UX, no scroll-back confusion
• **OnProcessDeleteKeyboard()** - Keep message history but disable old interactions
• **Default behavior** - Traditional behavior where all keyboards remain functional`)
	})

	// Start the bot
	log.Println("Bot starting... Use /help to see demo commands")
	log.Fatal(bot.Start())
}
