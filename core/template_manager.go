package teleflow

import (
	"fmt"
	"html"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TemplateManager interface {

	//

	//

	AddTemplate(name, templateText string, parseMode ParseMode) error

	//

	//

	HasTemplate(name string) bool

	//

	//

	GetTemplateInfo(name string) *TemplateInfo

	//

	ListTemplates() []string

	//

	//

	RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error)
}

type templateManager struct {
	templates *template.Template

	registry map[string]*TemplateInfo
}

func newTemplateManager() *templateManager {
	return &templateManager{
		templates: template.New("templateManager").Funcs(getAllTemplateFuncs()),
		registry:  make(map[string]*TemplateInfo),
	}
}

func (tm *templateManager) AddTemplate(name, templateText string, parseMode ParseMode) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if templateText == "" {
		return fmt.Errorf("template text cannot be empty")
	}

	if err := validateParseMode(parseMode); err != nil {
		return fmt.Errorf("invalid parse mode: %w", err)
	}

	if err := validateTemplateIntegrity(templateText, parseMode); err != nil {
		return fmt.Errorf("template integrity validation failed for '%s': %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(getTemplateFuncs(parseMode)).Parse(templateText)
	if err != nil {
		return fmt.Errorf("failed to parse template '%s': %w", name, err)
	}

	_, err = tm.templates.AddParseTree(name, tmpl.Tree)
	if err != nil {
		return fmt.Errorf("failed to add template '%s': %w", name, err)
	}

	tm.registry[name] = &TemplateInfo{
		Name:      name,
		ParseMode: parseMode,
		Template:  tmpl,
	}

	return nil
}

func (tm *templateManager) HasTemplate(name string) bool {
	return tm.templates.Lookup(name) != nil
}

func (tm *templateManager) GetTemplateInfo(name string) *TemplateInfo {
	return tm.registry[name]
}

func (tm *templateManager) ListTemplates() []string {
	var names []string
	for _, tmpl := range tm.templates.Templates() {
		if tmpl.Name() != "templateManager" {
			names = append(names, tmpl.Name())
		}
	}
	return names
}

func (tm *templateManager) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {

	tmpl := tm.templates.Lookup(name)
	if tmpl == nil {
		return "", ParseModeNone, fmt.Errorf("template '%s' not found", name)
	}

	info := tm.registry[name]
	if info == nil {
		return "", ParseModeNone, fmt.Errorf("template info not found for '%s'", name)
	}

	mergedData := tm.mergeTemplateData(data, nil)

	var buf strings.Builder
	err := tmpl.Execute(&buf, mergedData)
	if err != nil {
		return "", ParseModeNone, fmt.Errorf("failed to render template '%s': %w", name, err)
	}

	return buf.String(), info.ParseMode, nil
}

func (tm *templateManager) mergeTemplateData(templateData map[string]interface{}, contextData map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	for k, v := range contextData {
		merged[k] = v
	}

	for k, v := range templateData {
		merged[k] = v
	}

	return merged
}

func getAllTemplateFuncs() template.FuncMap {
	titleCaser := cases.Title(language.Und)
	return template.FuncMap{
		"escape": func(s string) string {

			return html.EscapeString(s)
		},
		"safe": func(s string) string {
			return s
		},
		"title": func(s string) string {
			return titleCaser.String(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}
}

func getTemplateFuncs(parseMode ParseMode) template.FuncMap {
	titleCaser := cases.Title(language.Und)
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

			return s
		},
		"title": func(s string) string {
			return titleCaser.String(s)
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

var defaultTemplateManager = newTemplateManager()

func GetDefaultTemplateManager() TemplateManager {
	return defaultTemplateManager
}
