package teleflow

import (
	"fmt"
	"log"
	"strings"
)

// promptRenderer handles the rendering of PromptConfig objects into Telegram messages
type promptRenderer struct {
	bot             *Bot
	messageRenderer *messageRenderer
	imageHandler    *imageHandler
	keyboardBuilder *inlineKeyboardBuilder
}

// renderContext provides context for rendering a prompt
type renderContext struct {
	ctx          *Context
	promptConfig *PromptConfig
	stepName     string
	flowName     string
}

// newPromptRenderer creates a new prompt renderer
func newPromptRenderer(bot *Bot) *promptRenderer {
	return &promptRenderer{
		bot:             bot,
		messageRenderer: newMessageRenderer(),
		imageHandler:    newImageHandler(),
		keyboardBuilder: newInlineKeyboardBuilder(),
	}
}

// render processes a PromptConfig and sends the appropriate Telegram message
func (pr *promptRenderer) render(renderCtx *renderContext) error {
	// Validate PromptConfig
	if err := pr.validatePromptConfig(renderCtx.promptConfig); err != nil {
		return fmt.Errorf("invalid PromptConfig in step %s: %w", renderCtx.stepName, err)
	}

	// Render components
	message, err := pr.messageRenderer.renderMessage(renderCtx.promptConfig, renderCtx.ctx)
	if err != nil {
		return pr.logFriendlyError("message rendering", renderCtx, err)
	}

	image, err := pr.imageHandler.processImage(renderCtx.promptConfig.Image, renderCtx.ctx)
	if err != nil {
		return pr.logFriendlyError("image processing", renderCtx, err)
	}

	keyboard, err := pr.keyboardBuilder.buildInlineKeyboard(renderCtx.promptConfig.Keyboard, renderCtx.ctx)
	if err != nil {
		return pr.logFriendlyError("keyboard generation", renderCtx, err)
	}

	// Send message according to Telegram rules
	return pr.sendMessage(renderCtx.ctx, message, image, keyboard)
}

// validatePromptConfig ensures the PromptConfig has at least one non-nil field
func (pr *promptRenderer) validatePromptConfig(config *PromptConfig) error {
	if config.Message == nil && config.Image == nil && config.Keyboard == nil {
		return fmt.Errorf("PromptConfig cannot have all fields nil - at least one of Message, Image, or Keyboard must be specified")
	}
	return nil
}

// sendMessage determines the appropriate message type and sends it
func (pr *promptRenderer) sendMessage(ctx *Context, message string, image *processedImage, keyboard interface{}) error {
	// Determine message type and content
	if image != nil {
		// Photo message with caption
		return pr.sendPhotoMessage(ctx, message, image, keyboard)
	}

	if message != "" {
		// Text message
		return pr.sendTextMessage(ctx, message, keyboard)
	}

	if keyboard != nil {
		// Keyboard only - send invisible message
		return pr.sendInvisibleMessage(ctx, keyboard)
	}

	return fmt.Errorf("no content to send - this should not happen after validation")
}

// sendPhotoMessage sends a photo with optional caption and keyboard
func (pr *promptRenderer) sendPhotoMessage(ctx *Context, caption string, image *processedImage, keyboard interface{}) error {
	// Use the new SendPhoto method from Context
	return ctx.SendPhoto(image, caption, keyboard)
}

// sendTextMessage sends a text message with optional keyboard
func (pr *promptRenderer) sendTextMessage(ctx *Context, message string, keyboard interface{}) error {
	if keyboard != nil {
		return ctx.Reply(message, keyboard)
	}
	return ctx.Reply(message)
}

// sendInvisibleMessage sends a keyboard-only message using zero-width space
func (pr *promptRenderer) sendInvisibleMessage(ctx *Context, keyboard interface{}) error {
	// Zero-width space character (invisible)
	invisibleText := "\u200B"
	return ctx.Reply(invisibleText, keyboard)
}

// logFriendlyError logs developer-friendly error messages with suggestions
func (pr *promptRenderer) logFriendlyError(component string, renderCtx *renderContext, err error) error {
	suggestions := pr.getSuggestions(component)

	log.Printf("🚨 TeleFlow Rendering Error in Flow '%s', Step '%s' (User %d):\n"+
		"   Component: %s\n"+
		"   Error: %s\n"+
		"   Suggestions:\n%s\n",
		renderCtx.flowName, renderCtx.stepName, renderCtx.ctx.UserID(),
		component, err.Error(),
		formatSuggestions(suggestions))

	return err
}

// getSuggestions returns component-specific error suggestions
func (pr *promptRenderer) getSuggestions(component string) []string {
	switch component {
	case "message rendering":
		return []string{
			"Check if message function returns valid string",
			"Verify template syntax if using template strings",
			"Ensure message is not nil in PromptConfig",
		}
	case "image processing":
		return []string{
			"Verify image file exists at specified path",
			"Check base64 encoding format",
			"Ensure image size is within limits",
			"Verify image file extension is supported",
		}
	case "keyboard generation":
		return []string{
			"Check keyboard function returns valid map[string]interface{}",
			"Verify callback_data values are strings",
			"Ensure keyboard function doesn't panic",
		}
	default:
		return []string{"Check the component implementation"}
	}
}

// formatSuggestions formats a list of suggestions for display
func formatSuggestions(suggestions []string) string {
	var result strings.Builder
	for i, suggestion := range suggestions {
		result.WriteString(fmt.Sprintf("   %d. %s\n", i+1, suggestion))
	}
	return result.String()
}
