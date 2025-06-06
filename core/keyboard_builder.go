package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// inlineKeyboardBuilder handles building keyboards from KeyboardFunc
type inlineKeyboardBuilder struct{}

// newInlineKeyboardBuilder creates a new keyboard builder
func newInlineKeyboardBuilder() *inlineKeyboardBuilder {
	return &inlineKeyboardBuilder{}
}

// buildInlineKeyboard processes a KeyboardFunc and returns the appropriate keyboard structure
func (kb *inlineKeyboardBuilder) buildInlineKeyboard(keyboardFunc KeyboardFunc, ctx *Context) (interface{}, error) {
	if keyboardFunc == nil {
		return nil, nil // No keyboard specified
	}

	// Execute the keyboard function
	keyboardMap := keyboardFunc(ctx)
	if keyboardMap == nil {
		return nil, nil // Function returned nil, no keyboard
	}

	// Convert map to inline keyboard structure
	return kb.convertMapToInlineKeyboard(keyboardMap)
}

// convertMapToInlineKeyboard converts a map[string]interface{} to an inline keyboard
func (kb *inlineKeyboardBuilder) convertMapToInlineKeyboard(keyboardMap map[string]interface{}) (interface{}, error) {
	if len(keyboardMap) == 0 {
		return nil, nil
	}

	// Create a proper tgbotapi.InlineKeyboardMarkup
	var rows [][]tgbotapi.InlineKeyboardButton

	for text, callbackData := range keyboardMap {
		// Validate callback data
		callbackStr, ok := callbackData.(string)
		if !ok {
			return nil, fmt.Errorf("keyboard callback data must be string, got %T for button '%s'", callbackData, text)
		}

		// Create button
		button := tgbotapi.NewInlineKeyboardButtonData(text, callbackStr)

		// Add as single-button row
		rows = append(rows, []tgbotapi.InlineKeyboardButton{button})
	}

	// Return proper inline keyboard markup
	return tgbotapi.NewInlineKeyboardMarkup(rows...), nil
}
