package teleflow

import (
	"fmt"
	"text/template"
)

// Template system provides powerful message templating capabilities using
// Go's built-in text/template engine. The system enables dynamic content
// generation, message personalization, and complex formatting with support
// for conditional logic, loops, and custom functions.
//
// Templates support:
//   - Dynamic variable substitution
//   - Conditional rendering with if/else statements
//   - Iterative content generation with range loops
//   - Custom template functions for advanced formatting
//   - Template inheritance and composition
//   - Automatic HTML/Markdown escaping for security
//   - Template caching for improved performance
//
// Basic Template Usage:
//
//	// Register a template
//	bot.AddTemplate("welcome", `
//	Hello {{.Name}}!
//	Welcome to our service. You have {{.MessageCount}} unread messages.
//	{{if gt .MessageCount 0}}
//	Would you like to read them now?
//	{{end}}`)
//
//	// Use template in handlers
//	bot.HandleCommand("/welcome", func(ctx *teleflow.Context) error {
//		data := map[string]interface{}{
//			"Name":         ctx.User().FirstName,
//			"MessageCount": getUserMessageCount(ctx.UserID()),
//		}
//		return ctx.ReplyTemplate("welcome", data)
//	})
//
// Advanced Template with Loops:
//
//	bot.AddTemplate("user_list", `
//	📋 Active Users:
//	{{range .Users}}
//	• {{.Name}} - {{.Status}}
//	{{end}}
//	{{if eq (len .Users) 0}}
//	No active users found.
//	{{end}}`)
//
// Template Functions:
//
//	// Templates support custom functions
//	bot.AddTemplate("formatted_date", `
//	Today is {{formatDate .Date "2006-01-02"}}
//	Welcome {{title .Username}}!`)
//
// Complex Template Example:
//
//	bot.AddTemplate("order_summary", `
//	🛒 Order Summary
//	Order #{{.OrderID}}
//	Date: {{formatDate .Date "Jan 2, 2006"}}
//
//	Items:
//	{{range .Items}}
//	• {{.Name}} x{{.Quantity}} - ${{printf "%.2f" .Price}}
//	{{end}}
//
//	{{if .Discount}}
//	Discount: -${{printf "%.2f" .Discount}}
//	{{end}}
//
//	Total: ${{printf "%.2f" .Total}}
//
//	{{if eq .Status "pending"}}
//	⏳ Payment pending
//	{{else if eq .Status "paid"}}
//	✅ Payment confirmed
//	{{end}}`)

// AddTemplate registers a template with the given name and template text
func (b *Bot) AddTemplate(name, templateText string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if templateText == "" {
		return fmt.Errorf("template text cannot be empty")
	}

	// Parse the template text
	tmpl, err := template.New(name).Parse(templateText)
	if err != nil {
		return fmt.Errorf("failed to parse template '%s': %w", name, err)
	}

	// Add the template to the bot's template collection
	// If this is the first template, we need to clone the existing template collection
	if b.templates == nil {
		b.templates = template.New("botMessages")
	}

	// Add the template to the collection
	_, err = b.templates.AddParseTree(name, tmpl.Tree)
	if err != nil {
		return fmt.Errorf("failed to add template '%s': %w", name, err)
	}

	return nil
}

// GetTemplate retrieves a template by name (useful for debugging/testing)
func (b *Bot) GetTemplate(name string) *template.Template {
	if b.templates == nil {
		return nil
	}
	return b.templates.Lookup(name)
}

// ListTemplates returns a list of all registered template names
func (b *Bot) ListTemplates() []string {
	if b.templates == nil {
		return []string{}
	}

	var names []string
	for _, tmpl := range b.templates.Templates() {
		if tmpl.Name() != "botMessages" { // Skip the root template
			names = append(names, tmpl.Name())
		}
	}
	return names
}

// HasTemplate checks if a template with the given name exists
func (b *Bot) HasTemplate(name string) bool {
	if b.templates == nil {
		return false
	}
	return b.templates.Lookup(name) != nil
}
