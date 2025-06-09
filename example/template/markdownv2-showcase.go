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
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}

func registerMarkdownV2Templates(bot *teleflow.Bot) {
	// 1. Basic formatting template
	if err := teleflow.AddTemplate("basic_formatting",
		"*Bold text* and _italic text_ and __underline__ and ~strikethrough~\n\n"+
			"You can also combine *_bold italic_* and *__bold underline__*",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add basic_formatting template: %v", err)
	}

	// 2. Code and preformatted text
	if err := teleflow.AddTemplate("code_example",
		"Here's some `inline code` and a code block:\n\n"+
			"```go\n"+
			"func main() {\n"+
			"    fmt.Println(\"Hello {{.Name}}!\")\n"+
			"}\n"+
			"```",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add code_example template: %v", err)
	}

	// 3. Links and mentions
	if err := teleflow.AddTemplate("links_example",
		"Visit our [website](https://example.com) or check out [TeleFlow](https://github.com/kslamph/teleflow)\\.\n\n"+
			"You can also mention users like @username or use [inline links](tg://user?id={{.UserID}})\\.",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add links_example template: %v", err)
	}

	// 4. Lists and organization
	if err := teleflow.AddTemplate("lists_example",
		"*Available Features:*\n\n"+
			"{{range $i, $feature := .Features}}"+
			"{{.Index}}\\. {{.Name}}\n"+
			"{{end}}\n"+
			"*Pros:*\n"+
			"• Easy to use\n"+
			"• Powerful templates\n"+
			"• MarkdownV2 support",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add lists_example template: %v", err)
	}

	// 5. Spoilers and special formatting
	if err := teleflow.AddTemplate("spoilers_example",
		"Here's a secret: ||{{.Secret}}||\n\n"+
			"And here's some `monospace` text with a ||hidden spoiler||\\.",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add spoilers_example template: %v", err)
	}

	// 6. Complex user profile template
	if err := teleflow.AddTemplate("user_profile",
		"*👤 User Profile*\n\n"+
			"*Name:* {{.Name}}\n"+
			"*Role:* `{{.Role}}`\n"+
			"*Join Date:* {{.JoinDate}}\n"+
			"*Status:* {{if .IsActive}}✅ _Active_{{else}}❌ _Inactive_{{end}}\n\n"+
			"*Recent Activity:*\n"+
			"{{range .Activities}}"+
			"{{.Index}}\\. {{.Activity}}\n"+
			"{{end}}\n"+
			"||Last Login: {{.LastLogin}}||",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add user_profile template: %v", err)
	}

	// 7. Notification template with emojis
	if err := teleflow.AddTemplate("notification",
		"🔔 *{{.Type}} Notification*\n\n"+
			"{{if eq .Priority \"high\"}}🚨 *HIGH PRIORITY*{{else if eq .Priority \"medium\"}}⚠️ *Medium Priority*{{else}}ℹ️ *Low Priority*{{end}}\n\n"+
			"*Message:* {{.Message}}\n"+
			"*Time:* `{{.Time}}`\n\n"+
			"{{if .ActionRequired}}*Action Required:* ||{{.Action}}||{{end}}",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add notification template: %v", err)
	}

	// 8. Product showcase template
	if err := teleflow.AddTemplate("product_showcase",
		"🛍️ *{{.ProductName}}*\n\n"+
			"{{.Description}}\n\n"+
			"*💰 Price:* ~${{.OldPrice}}~ *${{.Price}}*\n"+
			"*⭐ Rating:* {{.Rating}}/5\n"+
			"*📦 Stock:* {{if .Stock}}{{if gt .Stock 0}}`{{.Stock}} available`{{else}}_Out of stock_{{end}}{{else}}_Stock info unavailable_{{end}}\n\n"+
			"*Features:*\n"+
			"{{range .Features}}"+
			"✅ {{.}}\n"+
			"{{end}}\n"+
			"[🛒 Buy Now]({{.PurchaseLink}}) \\| [📖 Learn More]({{.InfoLink}})",
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add product_showcase template: %v", err)
	}

	// 9. Main menu template
	if err := teleflow.AddTemplate("main_menu",
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
		teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add main_menu template: %v", err)
	}
}

func createTemplateShowcaseFlow() (*teleflow.Flow, error) {
	return teleflow.NewFlow("template_showcase").
		OnButtonClick(teleflow.DeleteButtons). // Delete previous messages on button clicks
		Step("start").
		Prompt("template:main_menu").
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			// Demonstrate complex callback data - can be any interface{}
			basicData := map[string]interface{}{
				"type":      "template_demo",
				"category":  "formatting",
				"template":  "basic_formatting",
				"user_id":   ctx.UserID(),
				"timestamp": "2024-06-07",
			}

			codeData := struct {
				Type     string `json:"type"`
				Template string `json:"template"`
				Language string `json:"language"`
			}{
				Type:     "code_example",
				Template: "code_example",
				Language: "go",
			}

			return teleflow.NewPromptKeyboard().
				ButtonCallback("📝 Basic Formatting", basicData).
				ButtonCallback("💻 Code Examples", codeData).
				Row().
				ButtonCallback("🔗 Links & Mentions", "links").
				ButtonCallback("📋 Lists", "lists").
				Row().
				ButtonCallback("🤫 Spoilers", "spoilers").
				Row().
				ButtonCallback("👤 User Profile", "profile").
				ButtonCallback("🔔 Notifications", "notification").
				ButtonCallback("🛍️ Product Showcase", "product")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick == nil {
				log.Println("Button click is nil, retrying...")
				return teleflow.Retry().WithPrompt(&teleflow.PromptConfig{
					Message: "Please click one of the buttons above to see a template example.",
				})
			}

			// Demonstrate handling different data types
			switch data := buttonClick.Data.(type) {
			case map[string]interface{}:
				// Handle complex map data
				if data["type"] == "template_demo" && data["category"] == "formatting" {
					ctx.Set("demo_data", data) // Store for use in template
					return teleflow.GoToStep("show_basic")
				}
			case struct {
				Type     string `json:"type"`
				Template string `json:"template"`
				Language string `json:"language"`
			}:
				// Handle struct data
				ctx.Set("code_data", data)
				return teleflow.GoToStep("show_code")
			case string:
				// Handle simple string data (backward compatibility)
				switch data {
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
				}
			}

			return teleflow.Retry()
		}).

		// Basic formatting example
		Step("show_basic").
		Prompt("template:basic_formatting").
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).

		// Code example
		Step("show_code").
		Prompt("template:code_example").
		WithTemplateData(map[string]interface{}{
			"Name": "World 世界",
		}).
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).

		// Links example
		Step("show_links").
		Prompt("template:links_example").
		WithTemplateData(map[string]interface{}{
			"UserID": "123456789",
		}).
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).

		// Lists example
		Step("show_lists").
		Prompt("template:lists_example").
		WithTemplateData(map[string]interface{}{
			"Features": []map[string]interface{}{
				{"Index": "1", "Name": "Template system with MarkdownV2 support"},
				{"Index": "2", "Name": "Data precedence \\(TemplateData over Context\\)"},
				{"Index": "3", "Name": "Parse mode auto\\-detection"},
				{"Index": "4", "Name": "Backwards compatibility"},
			},
		}).
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).
		// Spoilers example
		Step("show_spoilers").
		Prompt("template:spoilers_example").
		WithTemplateData(map[string]interface{}{
			"Secret": "Templates are awesome\\!",
		}).
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).

		// User profile example
		// User profile example
		Step("show_profile").
		Prompt("template:user_profile").
		WithTemplateData(map[string]interface{}{
			"Name":      "John Doe",
			"Role":      "Developer",
			"JoinDate":  "2024\\-01\\-15",
			"IsActive":  true,
			"LastLogin": "2024\\-06\\-07 10:30 AM",
			"Activities": []map[string]interface{}{
				{"Index": "1", "Activity": "Implemented template system"},
				{"Index": "2", "Activity": "Fixed markdown parsing"},
				{"Index": "3", "Activity": "Added new features"},
			},
		}).
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).
		// Notification example
		Step("show_notification").
		Prompt("template:notification").
		WithTemplateData(map[string]interface{}{
			"Type":           "System",
			"Priority":       "high",
			"Message":        "Template system has been successfully implemented\\!",
			"Time":           "2024\\-06\\-07 10:45 AM",
			"ActionRequired": true,
			"Action":         "Review the implementation and test all features",
		}).
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).

		// Product showcase example
		Step("show_product").
		Prompt("template:product_showcase").
		WithTemplateData(map[string]interface{}{
			"ProductName": "TeleFlow Pro",
			"Description": "_The ultimate Telegram bot framework with advanced template support\\._",
			"OldPrice":    "99\\.99",
			"Price":       "79\\.99",
			"Rating":      "4\\.9",
			"Stock":       15,
			"Features": []string{
				"Advanced template system",
				"MarkdownV2 support",
				"Flow\\-based architecture",
				"Type\\-safe APIs",
			},
			"PurchaseLink": "https://example.com/buy",
			"InfoLink":     "https://example.com/info",
		}).
		WithPromptKeyboard(getBackButton()).
		Process(handleBackButton).
		Build()
}

func getBackButton() func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
	return func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
		return teleflow.NewPromptKeyboard().
			ButtonCallback("🔙 Back to Menu", "back")
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
		return ctx.SendPromptWithTemplate("basic_formatting", nil)
	})

	bot.HandleCommand("demo_code", func(ctx *teleflow.Context, command string, args string) error {
		return ctx.SendPromptWithTemplate("code_example", map[string]interface{}{
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
	if err := teleflow.AddTemplate("not_understood", "❓ I didn\\'t understand \\`{{.Input}}\\`\\. Type /help for available commands\\.", teleflow.ParseModeMarkdownV2); err != nil {
		log.Fatalf("Failed to add not_understood template: %v", err)
	}
	bot.DefaultHandler(func(ctx *teleflow.Context, text string) error {
		return ctx.SendPromptWithTemplate("not_understood", map[string]interface{}{
			"Input": text,
		})
	})
}
