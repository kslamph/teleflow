package teleflow

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
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

	info := tm.registry[name]
	if info == nil {
		return "", ParseModeNone, fmt.Errorf("template info not found for '%s'", name)
	}
	// Ensure the template instance from the registry is used, as it was parsed with the correct FuncMap
	if info.Template == nil {
		return "", ParseModeNone, fmt.Errorf("parsed template not found in registry for '%s'", name)
	}
	tmplToExecute := info.Template // Use the template from the registry, not from tm.templates.Lookup(name)

	mergedData := tm.mergeTemplateData(data, nil)

	var buf strings.Builder

	jsonData, jsonErr := json.Marshal(mergedData)
	if jsonErr != nil {
		log.Printf("DEBUG: Error marshaling data for template %s: %v", name, jsonErr)
	}
	log.Printf("DEBUG: Rendering template '%s' with ParseMode '%s' and data: %s", name, info.ParseMode, string(jsonData))

	err := tmplToExecute.Execute(&buf, mergedData) // Execute the specific template instance from the registry
	if err != nil {
		log.Printf("ERROR: Failed to execute template '%s'. Data: %s. Error: %v", name, string(jsonData), err)
		return "", ParseModeNone, fmt.Errorf("failed to render template '%s': %w", name, err)
	}

	renderedString := buf.String()
	log.Printf("DEBUG: Successfully rendered template '%s'. Output: %s", name, renderedString)

	return renderedString, info.ParseMode, nil
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
			originalS := s
			var escapedS string
			switch parseMode {
			case ParseModeHTML:
				escapedS = html.EscapeString(s)
			case ParseModeMarkdown:
				escapedS = escapeMarkdown(s)
			case ParseModeMarkdownV2:
				escapedS = escapeMarkdownV2(s)
				// Log specifically for MarkdownV2
				log.Printf("DEBUG: escapeMarkdownV2 called with ParseMode '%s'. Input: '%s', Output: '%s'", parseMode, originalS, escapedS)
			default:
				escapedS = s
			}
			return escapedS
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
