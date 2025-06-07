package teleflow

import (
	"fmt"
	"html"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateManager interface defines the contract for template management
type TemplateManager interface {
	// Template registration
	AddTemplate(name, templateText string, parseMode ParseMode) error
	HasTemplate(name string) bool
	GetTemplateInfo(name string) *TemplateInfo
	ListTemplates() []string

	// Template rendering
	RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error)
}

// templateManager implements the TemplateManager interface
type templateManager struct {
	templates *template.Template
	registry  map[string]*TemplateInfo
}

// newTemplateManager creates a new template manager instance
func newTemplateManager() *templateManager {
	return &templateManager{
		templates: template.New("templateManager").Funcs(getAllTemplateFuncs()),
		registry:  make(map[string]*TemplateInfo),
	}
}

// AddTemplate registers a template with the given name, template text, and parse mode
func (tm *templateManager) AddTemplate(name, templateText string, parseMode ParseMode) error {
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

	// Add the template to the collection
	_, err = tm.templates.AddParseTree(name, tmpl.Tree)
	if err != nil {
		return fmt.Errorf("failed to add template '%s': %w", name, err)
	}

	// Store template metadata
	tm.registry[name] = &TemplateInfo{
		Name:      name,
		ParseMode: parseMode,
		Template:  tmpl,
	}

	return nil
}

// HasTemplate checks if a template with the given name exists
func (tm *templateManager) HasTemplate(name string) bool {
	return tm.templates.Lookup(name) != nil
}

// GetTemplateInfo retrieves template metadata by name
func (tm *templateManager) GetTemplateInfo(name string) *TemplateInfo {
	return tm.registry[name]
}

// ListTemplates returns a list of all registered template names
func (tm *templateManager) ListTemplates() []string {
	var names []string
	for _, tmpl := range tm.templates.Templates() {
		if tmpl.Name() != "templateManager" { // Skip the root template
			names = append(names, tmpl.Name())
		}
	}
	return names
}

// RenderTemplate renders a template with the given data and returns the result with parse mode
func (tm *templateManager) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	// Check if template exists
	tmpl := tm.templates.Lookup(name)
	if tmpl == nil {
		return "", ParseModeNone, fmt.Errorf("template '%s' not found", name)
	}

	// Get template info for parse mode
	info := tm.registry[name]
	if info == nil {
		return "", ParseModeNone, fmt.Errorf("template info not found for '%s'", name)
	}

	// Merge template data with context data if needed
	mergedData := tm.mergeTemplateData(data, nil)

	// Render the template
	var buf strings.Builder
	err := tmpl.Execute(&buf, mergedData)
	if err != nil {
		return "", ParseModeNone, fmt.Errorf("failed to render template '%s': %w", name, err)
	}

	return buf.String(), info.ParseMode, nil
}

// mergeTemplateData merges template data with context data, giving precedence to template data
func (tm *templateManager) mergeTemplateData(templateData map[string]interface{}, contextData map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// First add context data
	for k, v := range contextData {
		merged[k] = v
	}

	// Then add template data (overwrites context data)
	for k, v := range templateData {
		merged[k] = v
	}

	return merged
}

// getAllTemplateFuncs returns all template functions for all parse modes
func getAllTemplateFuncs() template.FuncMap {
	titleCaser := cases.Title(language.Und)
	return template.FuncMap{
		"escape": func(s string) string {
			// Default to HTML escaping - will be overridden in execution context
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

// getTemplateFuncs returns template functions for the given parse mode
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
			// Returns unescaped string - use with caution
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

// Global default template manager instance for backwards compatibility
var defaultTemplateManager = newTemplateManager()

// GetDefaultTemplateManager returns the global default template manager instance
func GetDefaultTemplateManager() TemplateManager {
	return defaultTemplateManager
}
