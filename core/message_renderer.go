// Package teleflow provides a Telegram bot framework for building conversational flows.
// This file contains the messageRenderer component which handles the conversion
// of MessageSpec specifications into final rendered text with appropriate parse modes.
package teleflow

import (
	"fmt"
)

// messageRenderer handles the rendering of MessageSpec into final string messages
// with appropriate parse modes. It supports two main types of MessageSpec:
//   - Static strings: Direct text or template references
//   - Dynamic functions: Functions that generate text based on context
//
// The renderer integrates with the template system to process template references
// (strings starting with "template:") and merges data from both PromptConfig
// template data and Context data, with PromptConfig data taking precedence.
type messageRenderer struct {
	// templateManager provides template rendering capabilities including
	// template lookup, data merging, and parse mode determination
	templateManager TemplateManager
}

// newMessageRenderer creates and initializes a new message renderer with the default
// template manager. The renderer is ready to process MessageSpec instances into
// final rendered text with appropriate parse modes.
//
// Returns a configured messageRenderer instance.
func newMessageRenderer() *messageRenderer {
	return &messageRenderer{
		templateManager: GetDefaultTemplateManager(),
	}
}

// renderMessage processes a MessageSpec from a PromptConfig and returns the final
// rendered message string along with the appropriate parse mode for Telegram.
//
// The method supports two types of MessageSpec:
//   - string: Static text or template reference (format: "template:templateName")
//   - func(*Context) string: Dynamic function that generates text based on context
//
// Template Processing:
//   - Template references are identified by the "template:" prefix
//   - Template data from PromptConfig.TemplateData is merged with Context.data
//   - PromptConfig template data takes precedence over Context data
//   - Templates can specify their own parse mode (HTML, Markdown, or None)
//
// Parameters:
//   - config: PromptConfig containing the MessageSpec and optional template data
//   - ctx: Conversation context providing user data and context variables
//
// Returns:
//   - string: The final rendered message text ready for Telegram
//   - ParseMode: The parse mode to use (HTML, Markdown, or None)
//   - error: Any error that occurred during rendering or template processing
func (mr *messageRenderer) renderMessage(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
	if config.Message == nil {
		return "", ParseModeNone, nil // No message specified
	}

	switch msg := config.Message.(type) {
	case string:
		// Static string message - check if it's a template reference
		return mr.handleStringMessage(msg, config, ctx)

	case func(*Context) string:
		// Dynamic message function - evaluate result then check if it's a template
		result := msg(ctx)
		return mr.handleStringMessage(result, config, ctx)

	default:
		return "", ParseModeNone, fmt.Errorf("unsupported message type: %T (expected string or func(*Context) string)", msg)
	}
}

// handleStringMessage processes a string message, determining whether it's a plain text
// message or a template reference that needs to be rendered through the template system.
//
// Template Detection:
//   - Template references are identified by the "template:" prefix
//   - Plain strings without this prefix are returned as-is with ParseModeNone
//   - Template references are processed through renderTemplateMessage
//
// Parameters:
//   - message: The string message to process (either plain text or template reference)
//   - config: PromptConfig containing template data for template rendering
//   - ctx: Context providing additional data for template rendering
//
// Returns the processed message text, parse mode, and any processing error.
func (mr *messageRenderer) handleStringMessage(message string, config *PromptConfig, ctx *Context) (string, ParseMode, error) {
	// Check if the message is a template reference
	isTemplate, templateName := isTemplateMessage(message)
	if isTemplate {
		// Render as template with data merging
		return mr.renderTemplateMessage(templateName, config, ctx)
	}

	// Return as plain text with no parse mode
	return message, ParseModeNone, nil
}

// renderTemplateMessage renders a named template with merged data from multiple sources.
// This method handles the complete template rendering process including data merging,
// template lookup, rendering, and parse mode determination.
//
// Data Merging Strategy:
//   - Context data provides the base data layer from conversation state
//   - PromptConfig.TemplateData provides template-specific overrides
//   - PromptConfig data takes precedence over Context data for conflicts
//
// Template Processing:
//   - Validates template exists before rendering
//   - Delegates actual rendering to TemplateManager
//   - Preserves parse mode determined by template configuration
//
// Parameters:
//   - templateName: Name of the template to render (without "template:" prefix)
//   - config: PromptConfig containing template-specific data overrides
//   - ctx: Context containing base conversation data
//
// Returns the rendered text, appropriate parse mode, and any rendering error.
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

// mergeDataSources combines data from Context and PromptConfig template data,
// giving precedence to PromptConfig data for any conflicting keys.
//
// This merging strategy allows:
//   - Context data to provide base conversation state and user information
//   - PromptConfig template data to override specific values for this message
//   - Template-specific customization without modifying global context
//
// Merge Process:
//  1. Start with Context data as the base layer
//  2. Overlay PromptConfig template data, overwriting any conflicts
//  3. Return the merged data map for template rendering
//
// Parameters:
//   - templateData: Template-specific data from PromptConfig (takes precedence)
//   - contextData: Base conversation data from Context
//
// Returns a merged map containing data from both sources with proper precedence.
func (mr *messageRenderer) mergeDataSources(templateData map[string]interface{}, contextData map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// First add context data as the base layer
	for k, v := range contextData {
		merged[k] = v
	}

	// Then add template data (overwrites context data for conflicts)
	for k, v := range templateData {
		merged[k] = v
	}

	return merged
}
