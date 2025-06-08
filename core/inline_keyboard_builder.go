package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// Package teleflow/core provides InlineKeyboardBuilder for creating interactive inline keyboards.
//
// This file contains the InlineKeyboardBuilder which provides a fluent API for constructing
// inline keyboards with buttons, callback data management, and UUID mapping for secure
// callback handling in Telegram bot applications.
//
// # Usage Patterns
//
// The InlineKeyboardBuilder is primarily used in two contexts:
//   - Directly via KeyboardFunc in flow steps
//   - Through PromptKeyboardHandler for UUID management
//
// # Basic Usage
//
//	keyboard := NewInlineKeyboard().
//		ButtonCallback("Approve", "approve_data").
//		ButtonCallback("Reject", "reject_data").
//		Row().
//		ButtonUrl("Help", "https://example.com/help")
//
// # UUID Management
//
// The builder automatically generates UUIDs for callback data to prevent
// callback data conflicts and provide secure callback handling. Original
// callback data is preserved in the UUID mapping for retrieval.

// InlineKeyboardBuilder provides a fluent API for building inline keyboards with automatic UUID management.
//
// This builder creates inline keyboards that appear directly below messages in Telegram.
// It automatically generates UUID-based callback data to ensure uniqueness and security,
// while maintaining a mapping to the original callback data for processing.
//
// Key features:
//   - Fluent method chaining for easy keyboard construction
//   - Automatic UUID generation for callback data security
//   - Support for multiple button types (callback, URL, inline query)
//   - Row-based layout control
//   - Built-in validation
//
// Example usage:
//
//	builder := NewInlineKeyboard().
//		ButtonCallback("Yes", "confirm_action").
//		ButtonCallback("No", "cancel_action").
//		Row().
//		ButtonUrl("Learn More", "https://example.com")
//
//	keyboard := builder.Build()
//	uuidMap := builder.GetUUIDMapping()
type InlineKeyboardBuilder struct {
	rows        [][]tgbotapi.InlineKeyboardButton // Completed rows of buttons
	currentRow  []tgbotapi.InlineKeyboardButton   // Current row being built
	uuidMapping map[string]interface{}            // Maps generated UUIDs to original callback data
}

// NewInlineKeyboard creates a new inline keyboard builder instance.
//
// This constructor initializes an empty keyboard builder with no buttons or rows.
// The builder uses fluent method chaining to construct the keyboard layout.
//
// Returns a new InlineKeyboardBuilder ready for button addition and configuration.
//
// Example:
//
//	builder := NewInlineKeyboard().
//		ButtonCallback("Option 1", "opt1").
//		ButtonCallback("Option 2", "opt2")
func NewInlineKeyboard() *InlineKeyboardBuilder {
	return &InlineKeyboardBuilder{
		rows:        make([][]tgbotapi.InlineKeyboardButton, 0),
		currentRow:  make([]tgbotapi.InlineKeyboardButton, 0),
		uuidMapping: make(map[string]interface{}),
	}
}

// ButtonCallback adds a callback button with automatic UUID generation for secure callback handling.
//
// This method creates a button that sends callback data when pressed. The original callback
// data is automatically replaced with a generated UUID to ensure uniqueness and security.
// The mapping between UUIDs and original data is maintained internally for retrieval.
//
// Parameters:
//   - text: The text displayed on the button
//   - data: The original callback data (can be any type: string, int, struct, etc.)
//
// Returns the same InlineKeyboardBuilder instance for method chaining.
//
// Example:
//
//	builder.ButtonCallback("Approve", "approve_request_123").
//		ButtonCallback("Reject", map[string]string{"action": "reject", "id": "123"})
func (kb *InlineKeyboardBuilder) ButtonCallback(text string, data interface{}) *InlineKeyboardBuilder {
	// Generate UUID for the callback data to ensure uniqueness
	callbackUUID := uuid.New().String()

	// Store the mapping from UUID to original data
	kb.uuidMapping[callbackUUID] = data

	// Create the button with UUID as callback data
	button := tgbotapi.NewInlineKeyboardButtonData(text, callbackUUID)
	kb.currentRow = append(kb.currentRow, button)

	return kb
}

// ButtonUrl adds a URL button that opens a web link when pressed.
//
// URL buttons don't send callback data but instead open the specified URL
// in the user's browser or in-app browser. They're useful for external links,
// documentation, or web-based actions.
//
// Parameters:
//   - text: The text displayed on the button
//   - url: The URL to open when the button is pressed (should include protocol: http:// or https://)
//
// Returns the same InlineKeyboardBuilder instance for method chaining.
//
// Example:
//
//	builder.ButtonUrl("Visit Website", "https://example.com").
//		ButtonUrl("Documentation", "https://docs.example.com")
func (kb *InlineKeyboardBuilder) ButtonUrl(text string, url string) *InlineKeyboardBuilder {
	button := tgbotapi.NewInlineKeyboardButtonURL(text, url)
	kb.currentRow = append(kb.currentRow, button)

	return kb
}

// Row finalizes the current row of buttons and starts a new row.
//
// This method moves any buttons added to the current row into the keyboard's
// row collection and prepares for adding buttons to a new row. Call this method
// whenever you want to create a new line of buttons in the keyboard layout.
//
// If the current row is empty, this method has no effect.
//
// Returns the same InlineKeyboardBuilder instance for method chaining.
//
// Example:
//
//	builder.ButtonCallback("Button 1", "data1").
//		ButtonCallback("Button 2", "data2").
//		Row().  // Start new row
//		ButtonCallback("Button 3", "data3")
//	// Results in: [Button 1 | Button 2]
//	//            [Button 3]
func (kb *InlineKeyboardBuilder) Row() *InlineKeyboardBuilder {
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
		kb.currentRow = make([]tgbotapi.InlineKeyboardButton, 0)
	}
	return kb
}

// Build finalizes the keyboard construction and returns the Telegram inline keyboard markup.
//
// This method completes the keyboard building process by adding any remaining buttons
// from the current row and creating the final tgbotapi.InlineKeyboardMarkup structure
// that can be used with Telegram bot API calls.
//
// Any buttons in the current row that haven't been moved to a completed row via Row()
// will be automatically included in the final keyboard.
//
// Returns a tgbotapi.InlineKeyboardMarkup ready for use with Telegram API.
//
// Example:
//
//	keyboard := NewInlineKeyboard().
//		ButtonCallback("Yes", "confirm").
//		ButtonCallback("No", "cancel").
//		Build()
//	// Use keyboard with Telegram API
func (kb *InlineKeyboardBuilder) Build() tgbotapi.InlineKeyboardMarkup {
	// Add any remaining buttons in current row to the final keyboard
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(kb.rows...)
}

// GetUUIDMapping returns the mapping of generated UUIDs to original callback data.
//
// This method provides access to the internal UUID-to-data mapping that allows
// retrieval of original callback data when processing callback queries. The mapping
// is essential for the PromptKeyboardHandler to resolve callback data back to
// the original values provided during button creation.
//
// Returns a map where keys are generated UUIDs and values are the original callback data.
//
// Example:
//
//	builder := NewInlineKeyboard().ButtonCallback("Test", "original_data")
//	mapping := builder.GetUUIDMapping()
//	// mapping contains: {"generated-uuid": "original_data"}
func (kb *InlineKeyboardBuilder) GetUUIDMapping() map[string]interface{} {
	return kb.uuidMapping
}

// ValidateBuilder ensures the keyboard has at least one button before building.
//
// This validation method checks that the keyboard builder contains at least one button
// across all rows (including the current row being built). Empty keyboards are not
// allowed by Telegram and will cause API errors if sent.
//
// Returns an error if the keyboard is empty, nil if validation passes.
//
// Example:
//
//	builder := NewInlineKeyboard().ButtonCallback("Test", "data")
//	if err := builder.ValidateBuilder(); err != nil {
//		// Handle validation error
//	}
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
