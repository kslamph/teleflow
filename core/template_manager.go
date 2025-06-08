package teleflow

import (
	"fmt"
	"html"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateManager interface defines the contract for template management and rendering.
// It provides a complete abstraction for template operations, allowing different
// implementations while maintaining consistent behavior across the templating system.
//
// The interface supports the full template lifecycle:
//   - Registration: Adding templates with validation and compilation
//   - Introspection: Checking existence and retrieving metadata
//   - Rendering: Processing templates with data to generate formatted output
//
// Implementations must ensure thread safety and proper error handling.
// The defaultTemplateManager provides the standard implementation used by Bot methods.
type TemplateManager interface {
	// AddTemplate registers a new template with validation and compilation.
	// The template text is parsed according to Go template syntax and validated
	// against the specified ParseMode requirements.
	//
	// Parameters:
	//   - name: Unique identifier for the template
	//   - templateText: Go template syntax with variables and control structures
	//   - parseMode: Telegram parsing mode for output formatting
	//
	// Returns:
	//   - error: If registration fails due to invalid syntax or duplicate names
	AddTemplate(name, templateText string, parseMode ParseMode) error

	// HasTemplate checks if a template with the specified name exists.
	// This method provides efficient existence checking without retrieving full metadata.
	//
	// Parameters:
	//   - name: The template identifier to check
	//
	// Returns:
	//   - bool: true if the template exists, false otherwise
	HasTemplate(name string) bool

	// GetTemplateInfo retrieves comprehensive metadata for a registered template.
	// Returns detailed information including ParseMode and compiled Template object.
	//
	// Parameters:
	//   - name: The template identifier
	//
	// Returns:
	//   - *TemplateInfo: Complete template metadata, or nil if template doesn't exist
	GetTemplateInfo(name string) *TemplateInfo

	// ListTemplates returns all registered template names.
	// Useful for introspection and debugging of available templates.
	//
	// Returns:
	//   - []string: Slice of all template names currently registered
	ListTemplates() []string

	// RenderTemplate processes a template with data and returns formatted output.
	// This is the core rendering method that combines template execution with
	// parse mode information for proper message formatting.
	//
	// Parameters:
	//   - name: The template identifier to render
	//   - data: Template variables and values for substitution
	//
	// Returns:
	//   - string: The rendered template output
	//   - ParseMode: The parsing mode associated with this template
	//   - error: If rendering fails due to missing template or execution errors
	RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error)
}

// templateManager is the default implementation of the TemplateManager interface.
// It provides a complete template management system with built-in validation,
// compilation, and rendering capabilities. The implementation uses Go's text/template
// package internally with custom template functions for Telegram-specific formatting.
//
// Architecture:
//   - templates: Root template collection containing all compiled templates
//   - registry: Metadata storage mapping template names to TemplateInfo objects
//
// The templateManager ensures thread safety and maintains consistency between
// the compiled templates and their associated metadata.
type templateManager struct {
	// templates holds the root template collection with all compiled templates.
	// Each template is stored as a named template within this collection.
	templates *template.Template

	// registry maps template names to their TemplateInfo metadata.
	// This provides quick access to template ParseMode and other properties.
	registry map[string]*TemplateInfo
}

// newTemplateManager creates a new template manager instance with initialized state.
// This constructor function sets up the template collection with built-in template
// functions and prepares the metadata registry for template storage.
//
// The created manager includes all standard template functions:
//   - escape: Automatic escaping based on ParseMode
//   - safe: Bypass escaping (use with caution)
//   - title: Title case conversion
//   - upper: UPPERCASE conversion
//   - lower: lowercase conversion
//
// Returns:
//   - *templateManager: A fully initialized template manager ready for use
func newTemplateManager() *templateManager {
	return &templateManager{
		templates: template.New("templateManager").Funcs(getAllTemplateFuncs()),
		registry:  make(map[string]*TemplateInfo),
	}
}

// AddTemplate registers and compiles a new template with comprehensive validation.
// This method implements the TemplateManager interface, providing the core template
// registration functionality with validation, compilation, and metadata storage.
//
// The registration process includes:
//  1. Input validation (non-empty name and template text)
//  2. ParseMode validation against supported modes
//  3. Template integrity validation for the specified ParseMode
//  4. Template compilation with ParseMode-specific functions
//  5. Metadata storage for efficient lookup and rendering
//
// Parameters:
//   - name: Unique identifier for the template (must be non-empty)
//   - templateText: Go template syntax content (must be non-empty)
//   - parseMode: Telegram parsing mode for output formatting
//
// Returns:
//   - error: If validation fails, compilation errors occur, or name conflicts exist
//
// Example:
//
//	err := manager.AddTemplate("welcome", `*Hello {{.Name}}!*`, teleflow.ParseModeMarkdownV2)
//	if err != nil {
//		log.Printf("Template registration failed: %v", err)
//	}
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

// HasTemplate checks if a template with the specified name exists in the manager.
// This method implements efficient template existence checking by looking up the
// compiled template directly without retrieving metadata.
//
// Parameters:
//   - name: The template identifier to check
//
// Returns:
//   - bool: true if the template exists and is available for rendering, false otherwise
func (tm *templateManager) HasTemplate(name string) bool {
	return tm.templates.Lookup(name) != nil
}

// GetTemplateInfo retrieves comprehensive metadata for a registered template.
// This method provides access to all template properties including ParseMode
// and the compiled Template object for introspection purposes.
//
// Parameters:
//   - name: The template identifier to look up
//
// Returns:
//   - *TemplateInfo: Complete template metadata if found, nil if template doesn't exist
func (tm *templateManager) GetTemplateInfo(name string) *TemplateInfo {
	return tm.registry[name]
}

// ListTemplates returns all registered template names for enumeration and debugging.
// This method provides a complete list of available templates, excluding internal
// template manager structures. Useful for template introspection and validation.
//
// Returns:
//   - []string: Slice containing all template names currently registered
func (tm *templateManager) ListTemplates() []string {
	var names []string
	for _, tmpl := range tm.templates.Templates() {
		if tmpl.Name() != "templateManager" { // Skip the root template
			names = append(names, tmpl.Name())
		}
	}
	return names
}

// RenderTemplate processes a template with provided data and returns formatted output.
// This method implements the core template rendering functionality, combining template
// execution with data merging and ParseMode information for proper message formatting.
//
// The rendering process includes:
//  1. Template existence validation
//  2. Metadata retrieval for ParseMode information
//  3. Data merging with context precedence handling
//  4. Template execution with merged data
//  5. Result packaging with ParseMode for message sending
//
// Parameters:
//   - name: The template identifier to render
//   - data: Template variables and values for substitution
//
// Returns:
//   - string: The rendered template output ready for sending
//   - ParseMode: The parsing mode associated with this template
//   - error: If rendering fails due to missing template or execution errors
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

// mergeTemplateData merges template data with context data using template data precedence.
// This internal method implements the data precedence rule where TemplateData from PromptConfig
// takes priority over Context data when both contain the same keys. This ensures predictable
// template rendering behavior and allows explicit data overrides.
//
// Parameters:
//   - templateData: Data from PromptConfig.TemplateData (higher precedence)
//   - contextData: Data from Context state (lower precedence)
//
// Returns:
//   - map[string]interface{}: Merged data map with template data overriding context data
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

// getAllTemplateFuncs returns the complete set of template functions for all parsing modes.
// This internal function provides the base template functions used during template manager
// initialization. The functions include text manipulation and escaping utilities.
//
// Returns:
//   - template.FuncMap: Map of function names to their implementations
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

// getTemplateFuncs returns ParseMode-specific template functions for template compilation.
// This internal function provides template functions customized for the specified parsing mode,
// particularly important for the "escape" function which applies appropriate escaping rules.
//
// Template Functions Provided:
//   - escape: Automatic escaping based on ParseMode (HTML, Markdown, MarkdownV2, or none)
//   - safe: Bypass escaping - returns unmodified string (use with caution)
//   - title: Title case conversion using Unicode-aware rules
//   - upper: UPPERCASE conversion
//   - lower: lowercase conversion
//
// Parameters:
//   - parseMode: The parsing mode that determines escaping behavior
//
// Returns:
//   - template.FuncMap: Map of function names to ParseMode-aware implementations
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

// defaultTemplateManager is the global template manager instance used by Bot methods.
// This singleton provides backwards compatibility and convenient access to template
// functionality without requiring explicit template manager instantiation.
//
// All Bot template methods (AddTemplate, HasTemplate, GetTemplateInfo, ListTemplates)
// delegate to this global instance, ensuring consistent template storage and behavior
// across the application.
var defaultTemplateManager = newTemplateManager()

// GetDefaultTemplateManager returns the global default template manager instance.
// This function provides access to the same template manager used by Bot methods,
// allowing direct interaction with the template system when needed for advanced use cases.
//
// The returned manager implements the full TemplateManager interface and shares
// the same template registry as Bot template methods.
//
// Returns:
//   - TemplateManager: The global template manager instance used by the framework
//
// Example:
//
//	manager := teleflow.GetDefaultTemplateManager()
//	templates := manager.ListTemplates()
//	fmt.Printf("Available templates: %v\n", templates)
func GetDefaultTemplateManager() TemplateManager {
	return defaultTemplateManager
}
