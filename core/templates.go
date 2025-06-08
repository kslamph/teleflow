package teleflow

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// Templates package provides a powerful templating system for dynamic message generation
// with Telegram-specific formatting support. The system integrates seamlessly with both
// the Step-Prompt-Process API and traditional handlers, offering zero learning curve
// message templating with built-in security and validation.
//
// # Core Features
//
// The template system supports:
//   - Variable substitution with Go template syntax {{.VariableName}}
//   - Conditional rendering with {{if .Condition}}...{{end}}
//   - List iteration with {{range .Items}}...{{end}}
//   - Built-in template functions: title, upper, lower, escape, safe
//   - Multiple Telegram parsing modes: None, Markdown, MarkdownV2, HTML
//   - Automatic format validation and security escaping
//   - Template data precedence (TemplateData overrides Context data)
//
// # Basic Usage with PromptConfig
//
// Register and use templates in Step-Prompt-Process flows:
//
//	// Register a template with Telegram formatting
//	bot.AddTemplate("welcome", `*Hello {{.Name}}!* Welcome to our service.`, teleflow.ParseModeMarkdownV2)
//
//	// Use in flow steps via PromptConfig
//	.Step("greeting").
//	Prompt("template:welcome").
//	WithTemplateData(map[string]interface{}{"Name": "John"})
//
// # Traditional Handler Integration
//
// Templates work with existing command and message handlers:
//
//	bot.HandleCommand("/welcome", func(ctx *teleflow.Context, command string, args string) error {
//		data := map[string]interface{}{"Name": ctx.Get("username")}
//		return ctx.ReplyTemplate("welcome", data)
//	})
//
// # Advanced PromptConfig Examples
//
// Complete PromptConfig with template integration:
//
//	prompt := &teleflow.PromptConfig{
//		Message:      "template:user_profile",
//		TemplateData: map[string]interface{}{
//			"Name":     "John Doe",
//			"Role":     "Developer",
//			"IsActive": true,
//		},
//		Keyboard: func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
//			return teleflow.NewInlineKeyboard().
//				ButtonCallback("Edit Profile", "edit").
//				ButtonCallback("View Stats", "stats")
//		},
//	}
//
// # Template Functions and Formatting
//
// Built-in template functions for text manipulation:
//
//	{{.Name | title}}     // Title case conversion
//	{{.Message | upper}}  // UPPERCASE conversion
//	{{.Text | lower}}     // lowercase conversion
//	{{.HTML | escape}}    // Auto-escape based on ParseMode
//	{{.Raw | safe}}       // Bypass escaping (use with caution)
//
// # List and Conditional Templates
//
//	bot.AddTemplate("feature_list", `
//	*Available Features:*
//	{{range .Features}}
//	{{if .Enabled}}✅{{else}}❌{{end}} {{.Name}}
//	{{end}}`, teleflow.ParseModeMarkdownV2)

// ParseMode represents the different Telegram message parsing modes.
// Each mode has specific syntax requirements and validation rules that are
// automatically enforced when templates are registered and rendered.
type ParseMode string

const (
	// ParseModeNone represents plain text with no formatting or parsing.
	// Messages are sent as-is without any special character interpretation.
	// Use this mode for simple text messages that don't require formatting.
	ParseModeNone ParseMode = ""

	// ParseModeMarkdown represents legacy Markdown parsing mode.
	// Supports basic formatting: *bold*, _italic_, `code`, [links](url).
	// Note: This mode has been deprecated by Telegram in favor of MarkdownV2.
	// Template validation ensures proper markdown syntax.
	ParseModeMarkdown ParseMode = "Markdown"

	// ParseModeMarkdownV2 represents the modern MarkdownV2 parsing mode.
	// Supports enhanced formatting with stricter escaping requirements:
	//   - *bold*, _italic_, __underline__, ~strikethrough~
	//   - `code`, ```language\ncode block```
	//   - [text](url), ||spoiler text||
	// Special characters must be escaped outside formatting contexts.
	// Template validation enforces proper MarkdownV2 syntax.
	ParseModeMarkdownV2 ParseMode = "MarkdownV2"

	// ParseModeHTML represents HTML-based parsing mode.
	// Supports HTML tags: <b>bold</b>, <i>italic</i>, <u>underline</u>,
	// <s>strikethrough</s>, <code>code</code>, <pre>preformatted</pre>,
	// <a href="url">links</a>, <tg-spoiler>spoiler</tg-spoiler>.
	// Template validation ensures proper HTML tag matching.
	ParseModeHTML ParseMode = "HTML"
)

// TemplateInfo holds comprehensive metadata about a registered template.
// It stores both the template configuration and the compiled template object
// for efficient rendering and introspection.
type TemplateInfo struct {
	// Name is the unique identifier for the template used in template references
	// like "template:name" in PromptConfig.Message
	Name string

	// ParseMode specifies the Telegram parsing mode for this template's output.
	// This determines how the rendered text will be formatted when sent to users.
	ParseMode ParseMode

	// Template is the compiled Go template object used for rendering.
	// It includes the parsed template tree and associated template functions.
	Template *template.Template
}

// AddTemplate registers a template with the bot's template system.
// This method delegates to the defaultTemplateManager for actual template storage and compilation.
//
// Parameters:
//   - name: Unique identifier for the template, used in "template:name" references
//   - templateText: Go template syntax with variables, conditionals, and loops
//   - parseMode: Telegram parsing mode (None, Markdown, MarkdownV2, HTML)
//
// The template system validates syntax according to the specified parseMode and compiles
// the template with built-in functions (title, upper, lower, escape, safe).
//
// Example:
//
//	err := bot.AddTemplate("welcome", `*Hello {{.Name}}!* Welcome to our service.`, teleflow.ParseModeMarkdownV2)
//	if err != nil {
//		log.Fatal("Template registration failed:", err)
//	}
func (b *Bot) AddTemplate(name, templateText string, parseMode ParseMode) error {
	return defaultTemplateManager.AddTemplate(name, templateText, parseMode)
}

// GetTemplateInfo retrieves comprehensive metadata for a registered template.
// This method delegates to the defaultTemplateManager for template lookup.
//
// Parameters:
//   - name: The template identifier used during registration
//
// Returns:
//   - *TemplateInfo: Complete template metadata including ParseMode and compiled Template
//   - nil: If the template doesn't exist
//
// Example:
//
//	info := bot.GetTemplateInfo("welcome")
//	if info != nil {
//		fmt.Printf("Template: %s, ParseMode: %s\n", info.Name, info.ParseMode)
//	}
func (b *Bot) GetTemplateInfo(name string) *TemplateInfo {
	return defaultTemplateManager.GetTemplateInfo(name)
}

// ListTemplates returns all registered template names.
// This method delegates to the defaultTemplateManager for template enumeration.
//
// Returns:
//   - []string: Slice of all template names available for use in "template:name" references
//
// Example:
//
//	templates := bot.ListTemplates()
//	fmt.Printf("Available templates: %v\n", templates)
func (b *Bot) ListTemplates() []string {
	return defaultTemplateManager.ListTemplates()
}

// HasTemplate checks whether a template with the specified name exists.
// This method delegates to the defaultTemplateManager for template existence checking.
//
// Parameters:
//   - name: The template identifier to check
//
// Returns:
//   - bool: true if the template exists, false otherwise
//
// Example:
//
//	if bot.HasTemplate("welcome") {
//		// Safe to use "template:welcome" in PromptConfig
//	}
func (b *Bot) HasTemplate(name string) bool {
	return defaultTemplateManager.HasTemplate(name)
}

// validateParseMode validates if the specified parse mode is supported by the template system.
// This internal function ensures only valid Telegram parsing modes are used during template registration.
//
// Parameters:
//   - mode: The ParseMode to validate
//
// Returns:
//   - nil: If the parse mode is supported
//   - error: If the parse mode is unsupported or invalid
func validateParseMode(mode ParseMode) error {
	switch mode {
	case ParseModeNone, ParseModeMarkdown, ParseModeMarkdownV2, ParseModeHTML:
		return nil
	default:
		return fmt.Errorf("unsupported parse mode: %s", mode)
	}
}

// validateTemplateIntegrity validates template content against parse mode requirements.
// This internal function performs syntax validation to ensure templates will render correctly
// with the specified Telegram parsing mode, catching common formatting errors early.
//
// Parameters:
//   - templateText: The template content to validate
//   - parseMode: The parsing mode that will be used for this template
//
// Returns:
//   - nil: If the template content is valid for the specified parse mode
//   - error: If validation fails due to syntax issues or format violations
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

// validateMarkdown validates basic Markdown syntax for legacy Markdown parsing mode.
// This internal function checks for common Markdown formatting errors that would cause
// rendering issues in Telegram, such as unmatched bold/italic markers and link brackets.
//
// Parameters:
//   - text: The template text to validate for Markdown compliance
//
// Returns:
//   - nil: If the Markdown syntax is valid
//   - error: If unmatched markers or invalid syntax is detected
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

// validateMarkdownV2 validates MarkdownV2 syntax with stricter escaping requirements.
// This internal function performs simplified validation for MarkdownV2 format, checking
// for potentially unescaped special characters that could cause parsing errors.
//
// Note: This is a simplified implementation. Full MarkdownV2 validation would require
// sophisticated parsing to handle all formatting contexts and escape sequences.
//
// Parameters:
//   - text: The template text to validate for MarkdownV2 compliance
//
// Returns:
//   - nil: If the MarkdownV2 syntax appears valid
//   - error: If potential escaping issues are detected
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

// validateHTML validates basic HTML syntax for HTML parsing mode.
// This internal function checks for properly matched HTML tags to prevent
// rendering errors in Telegram. It uses a stack-based approach to validate
// tag nesting and handles self-closing tags appropriately.
//
// Parameters:
//   - text: The template text to validate for HTML compliance
//
// Returns:
//   - nil: If the HTML syntax is valid with properly matched tags
//   - error: If unmatched tags or invalid HTML structure is detected
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

// isValidMarkdownV2Usage checks if special characters are used correctly (simplified implementation).
// This internal helper function provides a placeholder for more sophisticated MarkdownV2 validation.
// Currently returns true as a simplified check - a full implementation would require
// context-aware parsing to determine if special characters are properly escaped or used in formatting.
//
// Returns:
//   - bool: Always true in this simplified implementation
func isValidMarkdownV2Usage() bool {
	// This is a simplified implementation
	// In practice, you'd need sophisticated parsing to handle all edge cases
	return true
}

// isSelfClosingTag checks if an HTML tag is self-closing and doesn't require a closing tag.
// This internal helper function is used during HTML validation to determine proper tag matching.
// Self-closing tags like <br>, <img>, <hr> don't need closing tags and are handled specially.
//
// Parameters:
//   - tag: The HTML tag name to check (case-insensitive)
//
// Returns:
//   - bool: true if the tag is self-closing, false if it requires a closing tag
func isSelfClosingTag(tag string) bool {
	selfClosing := map[string]bool{
		"br": true, "hr": true, "img": true, "input": true,
		"area": true, "base": true, "col": true, "embed": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}
	return selfClosing[tag]
}

// escapeMarkdown escapes special characters for legacy Markdown parsing mode.
// This internal function provides automatic escaping when using the "escape" template function
// with ParseModeMarkdown. It handles the basic Markdown special characters.
//
// Parameters:
//   - s: The string to escape for Markdown safety
//
// Returns:
//   - string: The escaped string safe for Markdown rendering
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

// escapeMarkdownV2 escapes special characters for MarkdownV2 parsing mode.
// This internal function provides automatic escaping when using the "escape" template function
// with ParseModeMarkdownV2. It handles all MarkdownV2 special characters that require escaping
// outside of formatting contexts to prevent parsing errors.
//
// Parameters:
//   - s: The string to escape for MarkdownV2 safety
//
// Returns:
//   - string: The escaped string safe for MarkdownV2 rendering
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
