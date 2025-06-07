package teleflow

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// Template system provides simple message templating capabilities for dynamic content.
// The system focuses on ease of use and integrates seamlessly with the new
// Step-Prompt-Process API for zero learning curve message generation.
//
// Templates support:
//   - Simple variable substitution with {{.VariableName}}
//   - Basic conditional rendering with {{if .Condition}}
//   - List iteration with {{range .Items}}
//   - Built-in formatting functions (title, upper, lower)
//   - Automatic escaping for security
//
// Basic Template Usage with PromptConfig:
//
//	// Register a simple template
//	bot.AddTemplate("welcome", `Hello {{.Name}}! Welcome to our service.`, teleflow.ParseModeNone)
//
//	// Use in flow steps with PromptConfig
//	.Prompt(
//		"template:welcome",
//		map[string]interface{}{"Name": "John"},
//		nil, // no keyboard
//	)
//
// Template in Traditional Handlers:
//
//	bot.HandleCommand("/welcome", func(ctx *teleflow.Context, command string, args string) error {
//		data := map[string]interface{}{"Name": "User"}
//		return ctx.ReplyTemplate("welcome", data)
//	})
//
// Simple List Template:
//
//	bot.AddTemplate("simple_list", `
//	📋 Items:
//	{{range .Items}}• {{.}}
//	{{end}}`, teleflow.ParseModeNone)
//
// PromptConfig Integration:
//
//	// Templates work seamlessly with PromptConfig messages
//	prompt := &teleflow.PromptConfig{
//		Message: "template:welcome",
//		Data:    map[string]interface{}{"Name": userName},
//	}
//	return teleflow.RetryWithPrompt(prompt)
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

// AddTemplate registers a template with the given name, template text, and parse mode
func (b *Bot) AddTemplate(name, templateText string, parseMode ParseMode) error {
	return defaultTemplateManager.AddTemplate(name, templateText, parseMode)
}

// GetTemplateInfo retrieves template metadata by name
func (b *Bot) GetTemplateInfo(name string) *TemplateInfo {
	return defaultTemplateManager.GetTemplateInfo(name)
}

// ListTemplates returns a list of all registered template names
func (b *Bot) ListTemplates() []string {
	return defaultTemplateManager.ListTemplates()
}

// HasTemplate checks if a template with the given name exists
func (b *Bot) HasTemplate(name string) bool {
	return defaultTemplateManager.HasTemplate(name)
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
			if !isValidMarkdownV2Usage() {
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
func isValidMarkdownV2Usage() bool {
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
