package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// InlineKeyboardBuilder provides a fluent API for building inline keyboards
type InlineKeyboardBuilder struct {
	rows        [][]tgbotapi.InlineKeyboardButton
	currentRow  []tgbotapi.InlineKeyboardButton
	uuidMapping map[string]interface{} // Maps UUIDs to original callback data
}

// NewInlineKeyboard creates a new inline keyboard builder
func NewInlineKeyboard() *InlineKeyboardBuilder {
	return &InlineKeyboardBuilder{
		rows:        make([][]tgbotapi.InlineKeyboardButton, 0),
		currentRow:  make([]tgbotapi.InlineKeyboardButton, 0),
		uuidMapping: make(map[string]interface{}),
	}
}

// ButtonCallback adds a callback button with interface{} data that gets UUID-mapped
func (kb *InlineKeyboardBuilder) ButtonCallback(text string, data interface{}) *InlineKeyboardBuilder {
	// Generate UUID for the callback data
	callbackUUID := uuid.New().String()

	// Store the mapping
	kb.uuidMapping[callbackUUID] = data

	// Create the button with UUID as callback data
	button := tgbotapi.NewInlineKeyboardButtonData(text, callbackUUID)
	kb.currentRow = append(kb.currentRow, button)

	return kb
}

// ButtonUrl adds a URL button
func (kb *InlineKeyboardBuilder) ButtonUrl(text string, url string) *InlineKeyboardBuilder {
	button := tgbotapi.NewInlineKeyboardButtonURL(text, url)
	kb.currentRow = append(kb.currentRow, button)

	return kb
}

// Row finishes the current row and starts a new one
func (kb *InlineKeyboardBuilder) Row() *InlineKeyboardBuilder {
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
		kb.currentRow = make([]tgbotapi.InlineKeyboardButton, 0)
	}
	return kb
}

// Build finalizes the keyboard and returns the Telegram markup
func (kb *InlineKeyboardBuilder) Build() tgbotapi.InlineKeyboardMarkup {
	// Add any remaining buttons in current row
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(kb.rows...)
}

// GetUUIDMapping returns the UUID to data mapping for callback handling
func (kb *InlineKeyboardBuilder) GetUUIDMapping() map[string]interface{} {
	return kb.uuidMapping
}

// ValidateBuilder ensures the builder has at least one button
func (kb *InlineKeyboardBuilder) ValidateBuilder() error {
	totalButtons := len(kb.currentRow)
	for _, row := range kb.rows {
		totalButtons += len(row)
	}

	if totalButtons == 0 {
		return fmt.Errorf("inline keyboard must have at least one button")
	}

	return nil
}
