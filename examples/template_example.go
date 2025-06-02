package main

import (
	"fmt"
	"log"

	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	// This is a demonstration of the template system
	// Note: This example shows the template functionality but cannot run
	// without a valid Telegram bot token

	fmt.Println("Template System Example")
	fmt.Println("=======================")

	// Create a bot with proper initialization
	// Note: This would need a real token in practice
	bot, err := teleflow.NewBot("fake-token-for-demo")
	if err != nil {
		log.Fatal("This is just a demo - would work with real token")
	}

	// Add various templates
	templates := map[string]string{
		"welcome":      "🎉 Welcome {{.Name}}! You're user #{{.UserID}}",
		"notification": "📢 You have {{.Count}} new messages",
		"profile": `👤 Profile Information
Name: {{.Name}}
ID: {{.ID}}
{{if .IsAdmin}}🔑 Admin privileges enabled{{end}}
{{if .Messages}}Recent messages:
{{range .Messages}}- {{.}}
{{end}}{{end}}`,
		"menu": `🏠 Main Menu
Available options:
{{range .Options}}- {{.Name}}: {{.Description}}
{{end}}`,
	}

	// Register all templates
	for name, templateText := range templates {
		if err := bot.AddTemplate(name, templateText); err != nil {
			log.Printf("Failed to add template %s: %v", name, err)
		} else {
			fmt.Printf("✅ Added template: %s\n", name)
		}
	}

	// Show registered templates
	fmt.Printf("\n📋 Registered templates: %v\n", bot.ListTemplates())

	// Example of how templates would be used in handlers:
	fmt.Println("\n💡 Usage in handlers:")
	fmt.Println("ctx.ReplyTemplate(\"welcome\", map[string]interface{}{")
	fmt.Println("    \"Name\": \"John\",")
	fmt.Println("    \"UserID\": 12345,")
	fmt.Println("})")

	fmt.Println("\nThis would send: 🎉 Welcome John! You're user #12345")
}
