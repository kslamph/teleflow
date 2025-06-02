package main

import (
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {

	var menuButton = &teleflow.MenuButtonConfig{
		{
			Text:    "📖 Help",
			Command: "/help",
		},
		{
			Text:    "📝 Markdown",
			Command: "/markdown",
		},
		{
			Text:    "📝 MarkdownV2",
			Command: "/markdownv2",
		},
		{
			Text:    "🌐 HTML",
			Command: "/html",
		},
		{
			Text:    "🧑 Profile",
			Command: "/profile",
		},
		{
			Text:    "⚙️ Status",
			Command: "/status",
		},
	}

	// Get bot token from environment variable
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable is required")
	}
	bot, err := teleflow.NewBot(token)
	if err != nil {
		log.Fatal(err)
	}

	bot.WithMenuButton(menuButton)

	// Setup templates for different parsing modes
	setupTemplates(bot)

	// Setup command handlers
	setupHandlers(bot)

	log.Printf("🚀 Enhanced Template Demo Bot started...")
	log.Printf("📝 Features: Markdown, MarkdownV2, HTML parsing modes")
	log.Printf("🛡️ Security: Automatic content escaping")
	log.Printf("⚡ Enhanced: EditOrReplyTemplate support")

	log.Fatal(bot.Start())
}

func setupTemplates(bot *teleflow.Bot) {
	// Plain text template for welcome
	bot.MustAddTemplate("welcome", `
🎉 Welcome to the Enhanced Template Demo!

This bot showcases Teleflow's powerful template system with:
• Multiple parsing modes (Plain, Markdown, MarkdownV2, HTML)
• Automatic content escaping for security
• Template validation and integrity checking
• Advanced formatting capabilities

Send any of these commands to see demos:
/markdown - Markdown formatting demo
/markdownv2 - MarkdownV2 advanced demo
/html - HTML formatting demo
/profile - User profile with HTML
/status - System status with refresh
/help - Show help information
`, teleflow.ParseModeNone)

	// Markdown template
	bot.MustAddTemplate("markdown_demo", `
**Markdown Demo** 🎨

*Hello {{.Name | escape}}!*
_This text is italicized_
**This text is bold**

Here's what you can do:
• Send _formatted_ messages
• Use **bold** and *italic* text
• Create [links]({{.URL}})
• Add inline `+"`code`"+` snippets

*Current time:* {{.Time | escape}}
*Message count:* {{.Count}}

[🔗 Learn more about Markdown](https://core.telegram.org/bots/api#markdown-style)
`, teleflow.ParseModeMarkdown)

	// MarkdownV2 template with advanced formatting
	bot.MustAddTemplate("markdownv2_demo", `
*MarkdownV2 Demo* 🚀

__Hello {{.Name | escape}}\!__
_This is italic text_
*This is bold text*
~This is strikethrough~
||This is spoiler text||
__This is underlined text__

*Code examples:*
`+"`"+`inline code`+"`"+`

`+"```python"+`
def hello_world():
    print("Hello from MarkdownV2!")
`+"```"+`

*Links and formatting:*
[Click here]({{.URL}})
[User profile](tg://user?id={{.UserID}})

*Escaped special characters:*
Characters like \. \! \- \= are properly escaped

*Current data:*
• Time: {{.Time | escape}}
• User ID: {{.UserID}}
• Message count: {{.Count}}

_Note: MarkdownV2 requires precise escaping\!_
`, teleflow.ParseModeMarkdownV2)

	// HTML template with rich formatting
	// HTML template with rich formatting
	bot.MustAddTemplate("html_demo", `
<b>HTML Demo</b> 🌐

<u>Hello <i>{{.Name | escape}}</i>!</u>

<b>Formatting Options:</b>
• <b>Bold text</b>
• <i>Italic text</i>
• <u>Underlined text</u>
• <s>Strikethrough text</s>
• <code>Inline code</code>
• <tg-spoiler>Spoiler text</tg-spoiler>

<b>Code Block:</b>
<pre><code class="language-javascript">
function greetUser(name) {
	   return "Hello, " + name + "!";
}
</code></pre>

<b>Links:</b>
<a href="{{.URL}}">External Link</a>
<a href="tg://user?id={{.UserID}}">User Profile</a>

<b>Current Status:</b>
<blockquote>
Time: <code>{{.Time | escape}}</code>
User ID: <code>{{.UserID}}</code>
Messages: <code>{{.Count}}</code>
</blockquote>

<i>💡 HTML allows the richest formatting options!</i>
`, teleflow.ParseModeHTML)
	// Interactive template for user data
	bot.MustAddTemplate("user_profile", `
<b>👤 User Profile</b>

<u>Account Information:</u>
• <b>Name:</b> {{.FirstName | escape}} {{.LastName | escape}}
• <b>Username:</b> {{if .Username}}@{{.Username | escape}}{{else}}<i>Not set</i>{{end}}
• <b>User ID:</b> <code>{{.UserID}}</code>
• <b>Language:</b> {{.LanguageCode | escape}}

<u>Activity Stats:</u>
• <b>Join Date:</b> {{.JoinDate | escape}}
• <b>Messages Sent:</b> {{.MessageCount}}
• <b>Commands Used:</b> {{.CommandCount}}
• <b>Last Active:</b> {{.LastActive | escape}}

{{if .IsPremium}}<b>⭐ Premium User</b>{{else}}<i>Standard User</i>{{end}}

<blockquote>
<i>This profile uses HTML templates with automatic content escaping for security.</i>
</blockquote>
`, teleflow.ParseModeHTML)

	// Dynamic content template
	bot.MustAddTemplate("system_status", `
<b>🔧 System Status</b>

<u>Server Information:</u>
• <b>Uptime:</b> <code>{{.Uptime}}</code>
• <b>Memory Usage:</b> {{.MemoryUsage}}%
• <b>CPU Usage:</b> {{.CPUUsage}}%
• <b>Active Users:</b> {{.ActiveUsers}}

<u>Bot Statistics:</u>
• <b>Total Messages:</b> {{.TotalMessages}}
• <b>Commands Processed:</b> {{.CommandsProcessed}}
• <b>Templates Rendered:</b> {{.TemplatesRendered}}

<u>Status Indicators:</u>
{{if lt .MemoryUsage 80}}✅{{else}}⚠️{{end}} Memory: {{.MemoryUsage}}%
{{if lt .CPUUsage 70}}✅{{else}}⚠️{{end}} CPU: {{.CPUUsage}}%
{{if gt .ActiveUsers 0}}✅{{else}}❌{{end}} Users: {{.ActiveUsers}} online

<i>📊 Click refresh to update data</i>
`, teleflow.ParseModeHTML)

	// Echo response template for user input
	bot.MustAddTemplate("echo_response", `
<b>🤖 Echo Response</b>

<u>You sent:</u>
<blockquote>{{.UserMessage | escape}}</blockquote>

<u>Message Analysis:</u>
• <b>Length:</b> {{.Length}} characters
• <b>Word Count:</b> {{.WordCount}} words
• <b>Contains Emoji:</b> {{if .HasEmoji}}Yes ✅{{else}}No ❌{{end}}
• <b>Is Command:</b> {{if .IsCommand}}Yes{{else}}No{{end}}

<i>💡 This demonstrates automatic content escaping - your input is safely displayed!</i>

Try the demo commands:
/markdown /markdownv2 /html /profile /status
`, teleflow.ParseModeHTML)
}

func setupHandlers(bot *teleflow.Bot) {
	// Start command
	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		data := map[string]interface{}{
			"Name": "Developer",
		}
		return ctx.ReplyTemplate("welcome", data, nil)
	})

	// Markdown demo
	bot.HandleCommand("markdown", func(ctx *teleflow.Context) error {
		data := map[string]interface{}{
			"Name":  "Developer",
			"URL":   "https://telegram.org",
			"Time":  time.Now().Format("15:04:05"),
			"Count": 42,
		}
		return ctx.ReplyTemplate("markdown_demo", data)
	})

	// MarkdownV2 demo
	bot.HandleCommand("markdownv2", func(ctx *teleflow.Context) error {
		data := map[string]interface{}{
			"Name":   "Developer",
			"URL":    "https://telegram.org",
			"Time":   time.Now().Format("2006-01-02 15:04:05"),
			"UserID": ctx.UserID(),
			"Count":  123,
		}
		return ctx.ReplyTemplate("markdownv2_demo", data)
	})

	// HTML demo
	bot.HandleCommand("html", func(ctx *teleflow.Context) error {
		data := map[string]interface{}{
			"Name":   "Developer",
			"URL":    "https://telegram.org",
			"Time":   time.Now().Format("2006-01-02 15:04:05"),
			"UserID": ctx.UserID(),
			"Count":  456,
		}
		return ctx.ReplyTemplate("html_demo", data)
	})

	// User profile demo
	bot.HandleCommand("profile", func(ctx *teleflow.Context) error {
		if ctx.Update.Message == nil || ctx.Update.Message.From == nil {
			return ctx.Reply("Unable to get user information")
		}

		user := ctx.Update.Message.From
		data := map[string]interface{}{
			"FirstName":    user.FirstName,
			"LastName":     getStringOrEmpty(user.LastName),
			"Username":     getStringOrEmpty(user.UserName),
			"UserID":       user.ID,
			"LanguageCode": getStringOrDefault(user.LanguageCode, "en"),
			"JoinDate":     "2024-01-15",
			"MessageCount": 789,
			"CommandCount": 45,
			"LastActive":   time.Now().Format("2006-01-02 15:04"),
			"IsPremium":    false, // IsPremium not available in basic API
		}
		return ctx.ReplyTemplate("user_profile", data)
	})

	// System status demo
	bot.HandleCommand("status", func(ctx *teleflow.Context) error {
		data := map[string]interface{}{
			"Uptime":            "2d 14h 32m",
			"MemoryUsage":       65,
			"CPUUsage":          23,
			"ActiveUsers":       127,
			"TotalMessages":     15420,
			"CommandsProcessed": 3421,
			"TemplatesRendered": 8765,
		}

		// Create inline keyboard for refresh
		keyboard := teleflow.NewInlineKeyboard().
			AddButton("🔄 Refresh", "refresh_status").AddRow()

		return ctx.ReplyTemplate("system_status", data, keyboard)
	})

	// Help command - add as template first
	bot.MustAddTemplate("help_message", `
<b>🎯 Template Demo Help</b>

<u>Available Commands:</u>

<b>/markdown</b> - Basic Markdown formatting demo
<b>/markdownv2</b> - Advanced MarkdownV2 formatting
<b>/html</b> - Rich HTML formatting demo
<b>/profile</b> - Your user profile with HTML
<b>/status</b> - System status with refresh button

<u>Security Features:</u>
• All user input is automatically escaped
• Templates validated for syntax correctness
• Parse mode specific content handling

<u>Template Functions:</u>
• <code>escape</code> - Safe content escaping
• <code>safe</code> - Unescaped content (careful!)
• <code>title/upper/lower</code> - Text transformation

<i>💡 Try different commands to see how Teleflow handles various parsing modes securely!</i>
`, teleflow.ParseModeHTML)

	bot.HandleCommand("/help", func(ctx *teleflow.Context) error {
		return ctx.ReplyTemplate("help_message", map[string]interface{}{})
	})

	// Text handler for general messages
	bot.HandleText(func(ctx *teleflow.Context) error {
		if ctx.Update.Message == nil {
			return nil
		}

		text := ctx.Update.Message.Text

		// Skip commands as they're handled separately
		if strings.HasPrefix(text, "/") {
			return nil
		}

		data := map[string]interface{}{
			"UserMessage": text,
			"Length":      len(text),
			"WordCount":   len(strings.Fields(text)),
			"HasEmoji":    containsEmoji(text),
			"IsCommand":   strings.HasPrefix(text, "/"),
		}

		return ctx.ReplyTemplate("echo_response", data)
	})

	// Setup callback handler for refresh button
	setupCallbackHandlers(bot)
}

func setupCallbackHandlers(bot *teleflow.Bot) {
	// Create a callback handler using the SimpleCallback helper
	refreshHandler := teleflow.SimpleCallback("refresh_status", func(ctx *teleflow.Context, data string) error {
		// Simulate updated data
		templateData := map[string]interface{}{
			"Uptime":            "2d 14h 35m",
			"MemoryUsage":       68,
			"CPUUsage":          rand.Intn(100), // Random CPU usage for demo
			"ActiveUsers":       131,
			"TotalMessages":     15445,
			"CommandsProcessed": 3428,
			"TemplatesRendered": 8789,
		}

		keyboard := teleflow.NewInlineKeyboard().
			AddButton("🔄 Refresh", "refresh_status").AddRow()

		// This will edit the message if possible, otherwise send new
		return ctx.EditOrReplyTemplate("system_status", templateData, keyboard)
	})

	// Register the callback handler
	bot.RegisterCallback(refreshHandler)
}

// Helper functions
func getStringOrEmpty(s string) string {
	return s
}

func getStringOrDefault(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

func containsEmoji(s string) bool {
	// Simple emoji detection (basic implementation)
	for _, r := range s {
		if r >= 0x1F600 && r <= 0x1F64F || // emoticons
			r >= 0x1F300 && r <= 0x1F5FF || // misc symbols
			r >= 0x1F680 && r <= 0x1F6FF || // transport
			r >= 0x2600 && r <= 0x26FF || // misc symbols
			r >= 0x2700 && r <= 0x27BF { // dingbats
			return true
		}
	}
	return false
}

func mainMenu()
