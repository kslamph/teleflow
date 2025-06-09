package teleflow

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// ParseMode defines the parse mode for Telegram message formatting.
// Different parse modes support different formatting features and syntax.
type ParseMode string

const (
	// ParseModeNone disables all formatting - messages are sent as plain text.
	ParseModeNone ParseMode = ""

	// ParseModeMarkdown enables basic Markdown formatting (deprecated by Telegram).
	// Supports *bold*, _italic_, and `code` formatting.
	ParseModeMarkdown ParseMode = "Markdown"

	// ParseModeMarkdownV2 enables enhanced Markdown formatting with more features.
	// Supports __underline__, **bold**, _italic_, `code`, ```pre```, and more.
	// Requires escaping of special characters.
	ParseModeMarkdownV2 ParseMode = "MarkdownV2"

	// ParseModeHTML enables HTML formatting with standard HTML tags.
	// Supports <b>bold</b>, <i>italic</i>, <u>underline</u>, <code>code</code>, etc.
	ParseModeHTML ParseMode = "HTML"
)

// TemplateInfo contains metadata about a registered message template.
// It includes the template content, formatting mode, and compiled template.
type TemplateInfo struct {
	Name string // Unique template name

	ParseMode ParseMode // Telegram formatting mode for the template

	Template *template.Template // Compiled Go template
}

// AddTemplate registers a new message template with the default template manager.
// Templates use Go template syntax and support the specified Telegram parse mode.
// This is a convenience function for the global template manager.
//
// Example:
//
//	err := teleflow.AddTemplate("greeting", "Hello {{.name}}!", teleflow.ParseModeMarkdown)
func AddTemplate(name, templateText string, parseMode ParseMode) error {
	return defaultTemplateManager.AddTemplate(name, templateText, parseMode)
}

// GetTemplateInfo retrieves information about a registered template.
// Returns nil if the template doesn't exist. This is a convenience function
// for the global template manager.
func GetTemplateInfo(name string) *TemplateInfo {
	return defaultTemplateManager.GetTemplateInfo(name)
}

// ListTemplates returns a list of all registered template names.
// This is a convenience function for the global template manager.
func ListTemplates() []string {
	return defaultTemplateManager.ListTemplates()
}

// HasTemplate checks if a template with the given name is registered.
// This is a convenience function for the global template manager.
func HasTemplate(name string) bool {
	return defaultTemplateManager.HasTemplate(name)
}

// validateParseMode checks if the provided parse mode is supported.
// Returns an error if the parse mode is not recognized.
func validateParseMode(mode ParseMode) error {
	switch mode {
	case ParseModeNone, ParseModeMarkdown, ParseModeMarkdownV2, ParseModeHTML:
		return nil
	default:
		return fmt.Errorf("unsupported parse mode: %s", mode)
	}
}

// validateTemplateIntegrity validates that a template is compatible with its parse mode.
// This helps catch formatting issues early before templates are used.
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

func validateMarkdown(text string) error {

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

func validateMarkdownV2(text string) error {

	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	for _, char := range specialChars {

		if strings.Contains(text, char) {

			if !isValidMarkdownV2Usage() {
				return fmt.Errorf("potentially unescaped special character '%s' in MarkdownV2", char)
			}
		}
	}
	return nil
}

func validateHTML(text string) error {

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

func isValidMarkdownV2Usage() bool {

	return true
}

func isSelfClosingTag(tag string) bool {
	selfClosing := map[string]bool{
		"br": true, "hr": true, "img": true, "input": true,
		"area": true, "base": true, "col": true, "embed": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}
	return selfClosing[tag]
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"-", "\\-",
		"`", "\\`",
		"[", "\\[",
		"]", "\\]",
	)
	return replacer.Replace(s)
}

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
