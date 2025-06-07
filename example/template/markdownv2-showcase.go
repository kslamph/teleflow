package main

import (
	"log"
	"os"

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	// Get token from environment
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Create bot
	bot, err := teleflow.NewBot(token)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Register MarkdownV2 templates showcasing various features
	registerMarkdownV2Templates(bot)

	// Create showcase flow
	flow, err := createTemplateShowcaseFlow()
	if err != nil {
		log.Fatal("Failed to create showcase flow:", err)
	}

	// Register the flow
	bot.RegisterFlow(flow)

	// Handle commands
	setupCommands(bot)

	// Start bot
	log.Println("MarkdownV2 Template Showcase Bot started...")
	bot.Start()
}

func registerMarkdownV2Templates(bot *teleflow.Bot) {
	// 1. Basic formatting template
	bot.AddTemplate("basic_formatting",
		"*Bold text* and _italic text_ and __underline__ and ~strikethrough~\n\n"+
			"You can also combine *_bold italic_* and *__bold underline__*",
		teleflow.ParseModeMarkdownV2)

	// 2. Code and preformatted text
	bot.AddTemplate("code_example",
		"Here's some `inline code` and a code block:\n\n"+
			"```go\n"+
			"func main() {\n"+
			"    fmt.Println(\"Hello {{.Name}}!\")\n"+
			"}\n"+
			"```",
		teleflow.ParseModeMarkdownV2)

	// 3. Links and mentions
	bot.AddTemplate("links_example",
		"Visit our [website](https://example.com) or check out [TeleFlow](https://github.com/kslamph/teleflow)\\.\n\n"+
			"You can also mention users like @username or use [inline links](tg://user?id={{.UserID}})\\.",
		teleflow.ParseModeMarkdownV2)

	// 4. Lists and organization
	bot.AddTemplate("lists_example",
		"*Available Features:*\n\n"+
			"{{range $i, $feature := .Features}}"+
			"{{.Index}}\\. {{.Name}}\n"+
			"{{end}}\n"+
			"*Pros:*\n"+
			"• Easy to use\n"+
			"• Powerful templates\n"+
			"• MarkdownV2 support",
		teleflow.ParseModeMarkdownV2)

	// 5. Spoilers and special formatting
	bot.AddTemplate("spoilers_example",
		"Here's a secret: ||{{.Secret}}||\n\n"+
			"And here's some `monospace` text with a ||hidden spoiler||\\.",
		teleflow.ParseModeMarkdownV2)

	// 6. Complex user profile template
	bot.AddTemplate("user_profile",
		"*👤 User Profile*\n\n"+
			"*Name:* {{.Name}}\n"+
			"*Role:* `{{.Role}}`\n"+
			"*Join Date:* {{.JoinDate}}\n"+
			"*Status:* {{if .IsActive}}✅ _Active_{{else}}❌ _Inactive_{{end}}\n\n"+
			"*Recent Activity:*\n"+
			"{{range $i, $activity := .Activities}}"+
			"{{.Index}}\\. {{.Activity}}\n"+
			"{{end}}\n"+
			"||Last Login: {{.LastLogin}}||",
		teleflow.ParseModeMarkdownV2)

	// 7. Notification template with emojis
	bot.AddTemplate("notification",
		"🔔 *{{.Type}} Notification*\n\n"+
			"{{if eq .Priority \"high\"}}🚨 *HIGH PRIORITY*{{else if eq .Priority \"medium\"}}⚠️ *Medium Priority*{{else}}ℹ️ *Low Priority*{{end}}\n\n"+
			"*Message:* {{.Message}}\n"+
			"*Time:* `{{.Time}}`\n\n"+
			"{{if .ActionRequired}}*Action Required:* ||{{.Action}}||{{end}}",
		teleflow.ParseModeMarkdownV2)

	// 8. Product showcase template
	bot.AddTemplate("product_showcase",
		"🛍️ *{{.ProductName}}*\n\n"+
			"{{.Description}}\n\n"+
			"*💰 Price:* ~${{.OldPrice}}~ *${{.Price}}*\n"+
			"*⭐ Rating:* {{.Rating}}/5\n"+
			"*📦 Stock:* {{if gt .Stock 0}}`{{.Stock}} available`{{else}}_Out of stock_{{end}}\n\n"+
			"*Features:*\n"+
			"{{range .Features}}"+
			"✅ {{.}}\n"+
			"{{end}}\n"+
			"[🛒 Buy Now]({{.PurchaseLink}}) \\| [📖 Learn More]({{.InfoLink}})",
		teleflow.ParseModeMarkdownV2)

	// 9. Main menu template
	bot.AddTemplate("main_menu",
		"🎨 *MarkdownV2 Template Showcase*\n\n"+
			"Choose a template to see in action:\n\n"+
			"📝 Basic Formatting\n"+
			"💻 Code Examples\n"+
			"🔗 Links & Mentions\n"+
			"📋 Lists\n"+
			"🤫 Spoilers\n"+
			"👤 User Profile\n"+
			"🔔 Notifications\n"+
			"🛍️ Product Showcase",
		teleflow.ParseModeMarkdownV2)
}

func createTemplateShowcaseFlow() (*teleflow.Flow, error) {
	return teleflow.NewFlow("template_showcase").
		OnError(teleflow.OnErrorCancel()).
		Step("start").
		Prompt(
			"template:main_menu",
			nil,
			func(ctx *teleflow.Context) map[string]interface{} {
				return map[string]interface{}{
					"inline_keyboard": []map[string]interface{}{

						{"text": "📝 Basic Formatting", "callback_data": "basic"},
						{"text": "💻 Code Examples", "callback_data": "code"},

						{"text": "🔗 Links & Mentions", "callback_data": "links"},
						{"text": "📋 Lists", "callback_data": "lists"},

						{"text": "🤫 Spoilers", "callback_data": "spoilers"},
						{"text": "👤 User Profile", "callback_data": "profile"},

						{"text": "🔔 Notifications", "callback_data": "notification"},
						{"text": "🛍️ Product Showcase", "callback_data": "product"},
					},
				}
			},
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick == nil {
				return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
					Message: "Please click one of the buttons above to see a template example.",
				})
			}

			switch buttonClick.Data {
			case "basic":
				return teleflow.GoToStep("show_basic")
			case "code":
				return teleflow.GoToStep("show_code")
			case "links":
				return teleflow.GoToStep("show_links")
			case "lists":
				return teleflow.GoToStep("show_lists")
			case "spoilers":
				return teleflow.GoToStep("show_spoilers")
			case "profile":
				return teleflow.GoToStep("show_profile")
			case "notification":
				return teleflow.GoToStep("show_notification")
			case "product":
				return teleflow.GoToStep("show_product")
			default:
				return teleflow.Retry()
			}
		}).

		// Basic formatting example
		Step("show_basic").
		Prompt(
			"template:basic_formatting",
			nil,
			getBackButton(),
		).
		Process(handleBackButton).

		// Code example
		Step("show_code").
		Prompt(
			"template:code_example",
			nil,
			getBackButton(),
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Set template data for this step
			ctx.Set("Name", "World")
			return handleBackButton(ctx, input, buttonClick)
		}).

		// Links example
		Step("show_links").
		Prompt(
			"template:links_example",
			nil,
			getBackButton(),
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Set template data for this step
			ctx.Set("UserID", "123456789")
			return handleBackButton(ctx, input, buttonClick)
		}).

		// Lists example
		Step("show_lists").
		Prompt(
			"template:lists_example",
			nil,
			getBackButton(),
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Set template data for this step
			ctx.Set("Features", []map[string]interface{}{
				{"Index": "1", "Name": "Template system with MarkdownV2 support"},
				{"Index": "2", "Name": "Data precedence \\(TemplateData over Context\\)"},
				{"Index": "3", "Name": "Parse mode auto\\-detection"},
				{"Index": "4", "Name": "Backwards compatibility"},
			})
			return handleBackButton(ctx, input, buttonClick)
		}).

		// Spoilers example
		Step("show_spoilers").
		Prompt(
			"template:spoilers_example",
			nil,
			getBackButton(),
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Set template data for this step
			ctx.Set("Secret", "Templates are awesome\\!")
			return handleBackButton(ctx, input, buttonClick)
		}).

		// User profile example
		Step("show_profile").
		Prompt(
			"template:user_profile",
			nil,
			getBackButton(),
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Set template data for this step
			ctx.Set("Name", "John Doe")
			ctx.Set("Role", "Developer")
			ctx.Set("JoinDate", "2024\\-01\\-15")
			ctx.Set("IsActive", true)
			ctx.Set("LastLogin", "2024\\-06\\-07 10:30 AM")
			ctx.Set("Activities", []map[string]interface{}{
				{"Index": "1", "Activity": "Implemented template system"},
				{"Index": "2", "Activity": "Fixed markdown parsing"},
				{"Index": "3", "Activity": "Added new features"},
			})
			return handleBackButton(ctx, input, buttonClick)
		}).

		// Notification example
		Step("show_notification").
		Prompt(
			"template:notification",
			nil,
			getBackButton(),
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Set template data for this step
			ctx.Set("Type", "System")
			ctx.Set("Priority", "high")
			ctx.Set("Message", "Template system has been successfully implemented\\!")
			ctx.Set("Time", "2024\\-06\\-07 10:45 AM")
			ctx.Set("ActionRequired", true)
			ctx.Set("Action", "Review the implementation and test all features")
			return handleBackButton(ctx, input, buttonClick)
		}).

		// Product showcase example
		Step("show_product").
		Prompt(
			"template:product_showcase",
			nil,
			getBackButton(),
		).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Set template data for this step
			ctx.Set("ProductName", "TeleFlow Pro")
			ctx.Set("Description", "_The ultimate Telegram bot framework with advanced template support\\._")
			ctx.Set("OldPrice", "99\\.99")
			ctx.Set("Price", "79\\.99")
			ctx.Set("Rating", "4\\.9")
			ctx.Set("Stock", 15)
			ctx.Set("Features", []string{
				"Advanced template system",
				"MarkdownV2 support",
				"Flow\\-based architecture",
				"Type\\-safe APIs",
			})
			ctx.Set("PurchaseLink", "https://example.com/buy")
			ctx.Set("InfoLink", "https://example.com/info")
			return handleBackButton(ctx, input, buttonClick)
		}).
		Build()
}

func getBackButton() func(ctx *teleflow.Context) map[string]interface{} {
	return func(ctx *teleflow.Context) map[string]interface{} {
		return map[string]interface{}{
			"inline_keyboard": [][]map[string]interface{}{
				{
					{"text": "🔙 Back to Menu", "callback_data": "back"},
				},
			},
		}
	}
}

func handleBackButton(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
	if buttonClick != nil && buttonClick.Data == "back" {
		return teleflow.GoToStep("start")
	}
	return teleflow.Retry()
}

func setupCommands(bot *teleflow.Bot) {
	// Handle the start command to begin showcase
	bot.HandleCommand("start", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.StartFlow("template_showcase")
	})

	// Handle showcase command
	bot.HandleCommand("showcase", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.StartFlow("template_showcase")
	})

	// Demo individual templates using convenience methods
	bot.HandleCommand("demo_basic", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.ReplyTemplate("basic_formatting", nil)
	})

	bot.HandleCommand("demo_code", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.ReplyTemplate("code_example", map[string]interface{}{
			"Name": "TeleFlow",
		})
	})

	bot.HandleCommand("demo_profile", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.SendPromptWithTemplate("user_profile", map[string]interface{}{
			"Name":      "Jane Smith",
			"Role":      "Product Manager",
			"JoinDate":  "2023\\-08\\-10",
			"IsActive":  true,
			"LastLogin": "2024\\-06\\-07 09:15 AM",
			"Activities": []map[string]interface{}{
				{"Index": "1", "Activity": "Reviewed template designs"},
				{"Index": "2", "Activity": "Planned feature roadmap"},
				{"Index": "3", "Activity": "Coordinated with development team"},
			},
		})
	})

	// Help command
	bot.HandleCommand("help", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.SendPrompt(&teleflow.PromptConfig{
			Message: "*🤖 MarkdownV2 Template Showcase Bot*\n\n" +
				"*Commands:*\n" +
				"/start \\- Start the interactive showcase\n" +
				"/showcase \\- Same as /start\n" +
				"/demo\\_basic \\- Show basic formatting\n" +
				"/demo\\_code \\- Show code examples\n" +
				"/demo\\_profile \\- Show user profile template\n" +
				"/help \\- Show this help\n\n" +
				"*Features Demonstrated:*\n" +
				"• MarkdownV2 formatting \\(*bold*, _italic_, __underline__, ~strikethrough~\\)\n" +
				"• Code blocks and `inline code`\n" +
				"• [Links](https://example.com) and mentions\n" +
				"• Lists and organization\n" +
				"• ||Spoilers|| and special formatting\n" +
				"• Template data merging and precedence",
		})
	})

	// Default handler using template
	bot.AddTemplate("not_understood", "❓ I didn't understand `{{.Input}}`\\. Type /help for available commands\\.", teleflow.ParseModeMarkdownV2)
	bot.DefaultHandler(func(ctx *teleflow.Context, text string) error {
		return ctx.ReplyTemplate("not_understood", map[string]interface{}{
			"Input": text,
		})
	})
}
