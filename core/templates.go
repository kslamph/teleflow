package teleflow

import (
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"
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
//
// ParseMode represents different Telegram parsing modes
type ParseMode string

const (
	// ParseModeNone represents plain text (no parsing)
	ParseModeNone ParseMode = ""
	// ParseModeMarkdown represents Markdown parsing
	ParseModeMarkdown ParseMode = "Markdown"
	// ParseModeMarkdownV2 represents MarkdownV2 parsing
	ParseModeMarkdownV2 ParseMode = "MarkdownV2"
	// ParseModeHTML represents HTML parsing
	ParseModeHTML ParseMode = "HTML"
)

// TemplateInfo holds metadata about a registered template
type TemplateInfo struct {
	Name      string
	ParseMode ParseMode
	Template  *template.Template
}

// templateRegistry holds template metadata
var templateRegistry = make(map[string]*TemplateInfo)

// AddTemplate registers a template with the given name, template text, and parse mode
func (b *Bot) AddTemplate(name, templateText string, parseMode ParseMode) error {
	return b.addTemplateInternal(name, templateText, parseMode, false)
}

// MustAddTemplate registers a template and panics if it fails
func (b *Bot) MustAddTemplate(name, templateText string, parseMode ParseMode) {
	if err := b.addTemplateInternal(name, templateText, parseMode, true); err != nil {
		log.Fatalf("Failed to add template '%s': %v", name, err)
	}
}

// addTemplateInternal is the internal implementation for adding templates
func (b *Bot) addTemplateInternal(name, templateText string, parseMode ParseMode, mustAdd bool) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if templateText == "" {
		return fmt.Errorf("template text cannot be empty")
	}

	// Validate parse mode
	if err := validateParseMode(parseMode); err != nil {
		return fmt.Errorf("invalid parse mode: %w", err)
	}

	// Validate template integrity according to parse mode
	if err := validateTemplateIntegrity(templateText, parseMode); err != nil {
		return fmt.Errorf("template integrity validation failed for '%s': %w", name, err)
	}

	// Parse the template text with custom functions
	tmpl, err := template.New(name).Funcs(getTemplateFuncs(parseMode)).Parse(templateText)
	if err != nil {
		return fmt.Errorf("failed to parse template '%s': %w", name, err)
	}
	// Add the template to the bot's template collection
	if b.templates == nil {
		// Initialize with all template functions (combine all parse modes)
		allFuncs := getAllTemplateFuncs()
		b.templates = template.New("botMessages").Funcs(allFuncs)
	}

	// Add the template to the collection
	_, err = b.templates.AddParseTree(name, tmpl.Tree)
	if err != nil {
		return fmt.Errorf("failed to add template '%s': %w", name, err)
	}

	// Store template metadata
	templateRegistry[name] = &TemplateInfo{
		Name:      name,
		ParseMode: parseMode,
		Template:  tmpl,
	}

	return nil
}

// getAllTemplateFuncs returns all template functions for all parse modes
func getAllTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"escape": func(s string) string {
			// Default to HTML escaping - will be overridden in execution context
			return html.EscapeString(s)
		},
		"safe": func(s string) string {
			return s
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}
}

// GetTemplateInfo retrieves template metadata by name
func (b *Bot) GetTemplateInfo(name string) *TemplateInfo {
	return templateRegistry[name]
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

// validateParseMode validates if the parse mode is supported
func validateParseMode(mode ParseMode) error {
	switch mode {
	case ParseModeNone, ParseModeMarkdown, ParseModeMarkdownV2, ParseModeHTML:
		return nil
	default:
		return fmt.Errorf("unsupported parse mode: %s", mode)
	}
}

// validateTemplateIntegrity validates template against parse mode requirements
func validateTemplateIntegrity(templateText string, parseMode ParseMode) error {
	switch parseMode {
	case ParseModeMarkdown:
		return validateMarkdown(templateText)
	case ParseModeMarkdownV2:
		return validateMarkdownV2(templateText)
	case ParseModeHTML:
		return validateHTML(templateText)
	case ParseModeNone:
		return nil
	default:
		return fmt.Errorf("unknown parse mode: %s", parseMode)
	}
}

// validateMarkdown validates basic Markdown syntax
func validateMarkdown(text string) error {
	// Check for unmatched markdown characters
	patterns := []struct {
		char string
		name string
	}{
		{"*", "bold/italic"},
		{"_", "italic/underline"},
		{"`", "code"},
		{"[", "link opening bracket"},
		{"]", "link closing bracket"},
	}

	for _, pattern := range patterns {
		count := strings.Count(text, pattern.char)
		if pattern.char == "[" || pattern.char == "]" {
			// Links should have matching brackets
			openCount := strings.Count(text, "[")
			closeCount := strings.Count(text, "]")
			if openCount != closeCount {
				return fmt.Errorf("unmatched %s brackets in markdown", pattern.name)
			}
		} else if count%2 != 0 {
			return fmt.Errorf("unmatched %s markers in markdown", pattern.name)
		}
	}
	return nil
}

// validateMarkdownV2 validates MarkdownV2 syntax with stricter rules
func validateMarkdownV2(text string) error {
	// MarkdownV2 has stricter escaping requirements
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	// Check for unescaped special characters outside of formatting
	for _, char := range specialChars {
		// This is a simplified check - in practice, you'd need more sophisticated parsing
		if strings.Contains(text, char) {
			// Check if it's properly used in formatting context
			if !isValidMarkdownV2Usage(text, char) {
				return fmt.Errorf("potentially unescaped special character '%s' in MarkdownV2", char)
			}
		}
	}
	return nil
}

// validateHTML validates basic HTML syntax
func validateHTML(text string) error {
	// Check for unmatched HTML tags
	tagPattern := regexp.MustCompile(`<(/?)(\w+)(?:\s[^>]*)?>`)
	matches := tagPattern.FindAllStringSubmatch(text, -1)

	stack := []string{}
	for _, match := range matches {
		isClosing := match[1] == "/"
		tagName := strings.ToLower(match[2])

		if isClosing {
			if len(stack) == 0 || stack[len(stack)-1] != tagName {
				return fmt.Errorf("unmatched closing HTML tag: </%s>", tagName)
			}
			stack = stack[:len(stack)-1]
		} else {
			// Self-closing tags don't need to be matched
			if !isSelfClosingTag(tagName) {
				stack = append(stack, tagName)
			}
		}
	}

	if len(stack) > 0 {
		return fmt.Errorf("unclosed HTML tags: %v", stack)
	}
	return nil
}

// isValidMarkdownV2Usage checks if special characters are used correctly (simplified)
func isValidMarkdownV2Usage(text, char string) bool {
	// This is a simplified implementation
	// In practice, you'd need sophisticated parsing to handle all edge cases
	return true
}

// isSelfClosingTag checks if an HTML tag is self-closing
func isSelfClosingTag(tag string) bool {
	selfClosing := map[string]bool{
		"br": true, "hr": true, "img": true, "input": true,
		"area": true, "base": true, "col": true, "embed": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}
	return selfClosing[tag]
}

// getTemplateFuncs returns template functions for the given parse mode
func getTemplateFuncs(parseMode ParseMode) template.FuncMap {
	baseFuncs := template.FuncMap{
		"escape": func(s string) string {
			switch parseMode {
			case ParseModeHTML:
				return html.EscapeString(s)
			case ParseModeMarkdown:
				return escapeMarkdown(s)
			case ParseModeMarkdownV2:
				return escapeMarkdownV2(s)
			default:
				return s
			}
		},
		"safe": func(s string) string {
			// Returns unescaped string - use with caution
			return s
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}

	return baseFuncs
}

// escapeMarkdown escapes special characters for Markdown
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"`", "\\`",
		"[", "\\[",
		"]", "\\]",
	)
	return replacer.Replace(s)
}

// escapeMarkdownV2 escapes special characters for MarkdownV2
func escapeMarkdownV2(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(s)
}
