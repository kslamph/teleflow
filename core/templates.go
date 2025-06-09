package teleflow

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

type ParseMode string

const (
	ParseModeNone ParseMode = ""

	ParseModeMarkdown ParseMode = "Markdown"

	ParseModeMarkdownV2 ParseMode = "MarkdownV2"

	ParseModeHTML ParseMode = "HTML"
)

type TemplateInfo struct {
	Name string

	ParseMode ParseMode

	Template *template.Template
}

func AddTemplate(name, templateText string, parseMode ParseMode) error {
	return defaultTemplateManager.AddTemplate(name, templateText, parseMode)
}

func GetTemplateInfo(name string) *TemplateInfo {
	return defaultTemplateManager.GetTemplateInfo(name)
}

func ListTemplates() []string {
	return defaultTemplateManager.ListTemplates()
}

func HasTemplate(name string) bool {
	return defaultTemplateManager.HasTemplate(name)
}

func validateParseMode(mode ParseMode) error {
	switch mode {
	case ParseModeNone, ParseModeMarkdown, ParseModeMarkdownV2, ParseModeHTML:
		return nil
	default:
		return fmt.Errorf("unsupported parse mode: %s", mode)
	}
}

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
