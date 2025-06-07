package teleflow

import (
	"fmt"
)

// messageRenderer handles rendering of MessageSpec into string messages
type messageRenderer struct {
	templateManager TemplateManager
}

// newMessageRenderer creates a new message renderer
func newMessageRenderer() *messageRenderer {
	return &messageRenderer{
		templateManager: GetDefaultTemplateManager(),
	}
}

// renderMessage processes a MessageSpec and returns the final message string with parse mode
func (mr *messageRenderer) renderMessage(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
	if config.Message == nil {
		return "", ParseModeNone, nil // No message specified
	}

	switch msg := config.Message.(type) {
	case string:
		// Static string message - check if it's a template
		return mr.handleStringMessage(msg, config, ctx)

	case func(*Context) string:
		// Dynamic message function - evaluate then check if result is template
		result := msg(ctx)
		return mr.handleStringMessage(result, config, ctx)

	default:
		return "", ParseModeNone, fmt.Errorf("unsupported message type: %T (expected string or func(*Context) string)", msg)
	}
}

// handleStringMessage processes a string message, checking if it's a template
func (mr *messageRenderer) handleStringMessage(message string, config *PromptConfig, ctx *Context) (string, ParseMode, error) {
	// Check if the message is a template reference
	isTemplate, templateName := isTemplateMessage(message)
	if isTemplate {
		// Render as template
		return mr.renderTemplateMessage(templateName, config, ctx)
	}

	// Return as plain text
	return message, ParseModeNone, nil
}

// renderTemplateMessage renders a template with merged data sources
func (mr *messageRenderer) renderTemplateMessage(templateName string, config *PromptConfig, ctx *Context) (string, ParseMode, error) {
	// Check if template exists
	if !mr.templateManager.HasTemplate(templateName) {
		return "", ParseModeNone, fmt.Errorf("template '%s' not found", templateName)
	}

	// Merge data sources: context data + template data (template data takes precedence)
	mergedData := mr.mergeDataSources(config.TemplateData, ctx.data)

	// Render the template
	renderedText, parseMode, err := mr.templateManager.RenderTemplate(templateName, mergedData)
	if err != nil {
		return "", ParseModeNone, fmt.Errorf("failed to render template '%s': %w", templateName, err)
	}

	return renderedText, parseMode, nil
}

// mergeDataSources merges TemplateData and Context data, giving precedence to TemplateData
func (mr *messageRenderer) mergeDataSources(templateData map[string]interface{}, contextData map[string]interface{}) map[string]interface{} {
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
