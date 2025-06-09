package teleflow

import (
	"fmt"
)

type messageHandler struct {
	templateManager TemplateManager
}

func newMessageHandler(tm TemplateManager) *messageHandler {
	return &messageHandler{
		templateManager: tm,
	}
}

func (mr *messageHandler) renderMessage(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
	if config.Message == nil {
		return "", ParseModeNone, nil
	}

	switch msg := config.Message.(type) {
	case string:

		return mr.handleStringMessage(msg, config, ctx)

	case func(*Context) string:

		result := msg(ctx)
		return mr.handleStringMessage(result, config, ctx)

	default:
		return "", ParseModeNone, fmt.Errorf("unsupported message type: %T (expected string or func(*Context) string)", msg)
	}
}

func (mr *messageHandler) handleStringMessage(message string, config *PromptConfig, ctx *Context) (string, ParseMode, error) {

	isTemplate, templateName := isTemplateMessage(message)
	if isTemplate {

		return mr.renderTemplateMessage(templateName, config, ctx)
	}

	return message, ParseModeNone, nil
}

func (mr *messageHandler) renderTemplateMessage(templateName string, config *PromptConfig, ctx *Context) (string, ParseMode, error) {

	if !mr.templateManager.HasTemplate(templateName) {
		return "", ParseModeNone, fmt.Errorf("template '%s' not found", templateName)
	}

	// Use only explicit TemplateData - no context data merging
	templateData := config.TemplateData
	if templateData == nil {
		templateData = make(map[string]interface{})
	}

	renderedText, parseMode, err := mr.templateManager.RenderTemplate(templateName, templateData)
	if err != nil {
		return "", ParseModeNone, fmt.Errorf("failed to render template '%s': %w", templateName, err)
	}

	return renderedText, parseMode, nil
}
