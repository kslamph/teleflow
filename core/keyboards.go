package teleflow

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Package teleflow/core provides keyboard systems for creating Telegram reply keyboards and inline keyboards.
//
// This file contains core keyboard structures and utility functions that support both traditional
// keyboard builders and modern map-based keyboard approaches used in flow steps.
//
// # Keyboard Types
//
// Two main keyboard types are supported:
//   - Reply Keyboards: Appear below the message input field, send text when pressed
//   - Inline Keyboards: Appear below messages as clickable buttons with callback data
//
// # Modern Map-based Approach (Recommended)
//
// For flow steps, use the simple map-based syntax in Prompt() calls:
//
//	.Prompt(
//		"Please choose an option:",
//		nil, // optional image
//		func(ctx *teleflow.Context) map[string]interface{} {
//			return map[string]interface{}{
//				"✅ Approve": "approve_123",
//				"❌ Reject":  "reject_123",
//				"ℹ️ More Info": "info_123",
//			}
//		},
//	)
//
//	// Handle button clicks in Process function
//	.Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//		if buttonClick != nil {
//			switch buttonClick.Data {
//			case "approve_123":
//				return teleflow.CompleteFlow()
//			case "reject_123":
//				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{Message: "Please provide reason:"})
//			}
//		}
//		return teleflow.NextStep()
//	})
//
// # Traditional Builder Approach (Advanced Use Cases)
//
// For complex keyboards requiring special button types:
//
//	// Request contact information
//	keyboard := teleflow.NewReplyKeyboard().
//		AddRow(teleflow.NewReplyButton("📱 Share Contact").SetRequestContact()).
//		AddRow(teleflow.NewReplyButton("📍 Share Location").SetRequestLocation())
//

// ReplyKeyboardButton represents a single button in a reply keyboard.
//
// Reply keyboard buttons appear below the message input field and send their
// text content when pressed. They can optionally request contact or location
// information from the user.
type ReplyKeyboardButton struct {
	Text            string `json:"text"`                       // Button text displayed to user
	RequestContact  bool   `json:"request_contact,omitempty"`  // Request user's contact when pressed
	RequestLocation bool   `json:"request_location,omitempty"` // Request user's location when pressed
}

// ReplyKeyboard represents a custom reply keyboard with multiple buttons arranged in rows.
//
// Reply keyboards appear below the message input area and remain visible until
// hidden or replaced. They're ideal for main menus and persistent navigation.
//
// Use fluent methods for configuration:
//
//	keyboard := BuildReplyKeyboard([]string{"Option A", "Option B"}, 2).
//		Resize().
//		OneTime().
//		Placeholder("Choose an option...")
type ReplyKeyboard struct {
	Keyboard              [][]ReplyKeyboardButton `json:"keyboard"`                          // 2D array of keyboard buttons
	ResizeKeyboard        bool                    `json:"resize_keyboard,omitempty"`         // Resize keyboard to fit buttons
	OneTimeKeyboard       bool                    `json:"one_time_keyboard,omitempty"`       // Hide keyboard after use
	InputFieldPlaceholder string                  `json:"input_field_placeholder,omitempty"` // Placeholder text in input field
	Selective             bool                    `json:"selective,omitempty"`               // Show keyboard only to mentioned users
}

// InlineKeyboardButton represents a single button in an inline keyboard.
//
// Inline keyboard buttons appear directly below messages and support various
// interaction types including callbacks, URLs, and inline query switching.
type InlineKeyboardButton struct {
	Text                         string `json:"text"`                                       // Button text displayed to user
	URL                          string `json:"url,omitempty"`                              // URL to open when pressed
	CallbackData                 string `json:"callback_data,omitempty"`                    // Data sent in callback query
	SwitchInlineQuery            string `json:"switch_inline_query,omitempty"`              // Switch to inline mode in any chat
	SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat,omitempty"` // Switch to inline mode in current chat
}

// InlineKeyboard represents an inline keyboard with buttons arranged in rows.
//
// Inline keyboards appear directly below messages and provide interactive
// elements without cluttering the input area. They're ideal for quick actions,
// confirmations, and navigation within conversations.
//
// Use InlineKeyboardBuilder for construction:
//
//	keyboard := NewInlineKeyboard().
//		ButtonCallback("Approve", "approve_data").
//		ButtonCallback("Reject", "reject_data").
//		Row().
//		ButtonUrl("Learn More", "https://example.com")
type InlineKeyboard struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"` // 2D array of inline keyboard buttons
}

// BuildMenuButton is DEPRECATED and has been removed.
//
// MIGRATION: Use Bot.SetBotCommands() for command management or
// NewDefaultMenuButton() from menu_button.go for default menu buttons.
//
//	// Old way (removed):
//	// menuButton := teleflow.BuildMenuButton()
//
//	// New way (recommended):
//	err := bot.SetBotCommands(map[string]string{
//		"help":  "📖 Show help information",
//		"start": "🚀 Start the bot",
//	})
//
//	// Or for default menu button:
//	menuButton := teleflow.NewDefaultMenuButton()

// newReplyKeyboard creates a new reply keyboard from button rows.
//
// This is an internal constructor function used by other keyboard creation
// utilities. For public API, use BuildReplyKeyboard instead.
//
// Parameters:
//   - rows: Variable number of button rows to initialize the keyboard with
//
// Returns a new ReplyKeyboard instance with the provided button rows.
func newReplyKeyboard(rows ...[]ReplyKeyboardButton) *ReplyKeyboard {
	kb := &ReplyKeyboard{
		Keyboard: make([][]ReplyKeyboardButton, 0),
	}
	kb.Keyboard = append(kb.Keyboard, rows...)
	return kb
}

// BuildReplyKeyboard creates a reply keyboard with buttons distributed across rows.
//
// This function provides a convenient way to create reply keyboards by automatically
// distributing buttons across rows based on the specified buttons-per-row count.
// The resulting keyboard can be further customized using fluent methods.
//
// Parameters:
//   - buttons: Slice of button text labels to create buttons for
//   - buttonsPerRow: Number of buttons to place in each row (must be > 0, defaults to 1 if invalid)
//
// Returns a ReplyKeyboard instance ready for further configuration or use.
//
// Example usage:
//
//	// Create a keyboard with 3 buttons per row
//	keyboard := teleflow.BuildReplyKeyboard(
//		[]string{"Option A", "Option B", "Option C", "Option D", "Option E"}, 3)
//	// Results in: [Option A | Option B | Option C] [Option D | Option E]
//
//	// Configure the keyboard with fluent methods
//	keyboard.Resize().OneTime().Placeholder("Choose an option")
//
//	// Single column layout
//	singleColumn := teleflow.BuildReplyKeyboard(
//		[]string{"Main Menu", "Settings", "Help", "Exit"}, 1)
func BuildReplyKeyboard(buttons []string, buttonsPerRow int) *ReplyKeyboard {
	if len(buttons) == 0 {
		return newReplyKeyboard()
	}

	if buttonsPerRow <= 0 {
		buttonsPerRow = 1 // Default to 1 button per row if invalid value
	}

	kb := &ReplyKeyboard{
		Keyboard: make([][]ReplyKeyboardButton, 0),
	}

	// Split buttons into rows based on buttonsPerRow
	for i := 0; i < len(buttons); i += buttonsPerRow {
		var row []ReplyKeyboardButton

		// Add up to buttonsPerRow buttons to current row
		for j := 0; j < buttonsPerRow && i+j < len(buttons); j++ {
			row = append(row, ReplyKeyboardButton{Text: buttons[i+j]})
		}

		kb.Keyboard = append(kb.Keyboard, row)
	}

	return kb
}

// Resize enables automatic keyboard resizing to fit the button layout.
//
// When enabled, Telegram clients will resize the keyboard vertically to
// accommodate the optimal button size. This typically results in a more
// compact and visually appealing keyboard layout.
//
// Returns the same ReplyKeyboard instance for method chaining.
//
// Example:
//
//	keyboard := BuildReplyKeyboard(buttons, 2).Resize()
func (kb *ReplyKeyboard) Resize() *ReplyKeyboard {
	kb.ResizeKeyboard = true
	return kb
}

// OneTime configures the keyboard to hide automatically after the user presses any button.
//
// This is useful for keyboards that represent one-time choices or confirmations.
// The user will see the keyboard disappear after making their selection, keeping
// the chat interface clean.
//
// Returns the same ReplyKeyboard instance for method chaining.
//
// Example:
//
//	confirmKeyboard := BuildReplyKeyboard([]string{"Yes", "No"}, 2).OneTime()
func (kb *ReplyKeyboard) OneTime() *ReplyKeyboard {
	kb.OneTimeKeyboard = true
	return kb
}

// Placeholder sets the placeholder text shown in the input field when the keyboard is active.
//
// This text appears in the message input field to guide users on what action
// is expected. It helps provide context for what the keyboard buttons represent.
//
// Parameters:
//   - text: The placeholder text to display in the input field
//
// Returns the same ReplyKeyboard instance for method chaining.
//
// Example:
//
//	menuKeyboard := BuildReplyKeyboard(menuOptions, 2).
//		Placeholder("Choose a menu option")
func (kb *ReplyKeyboard) Placeholder(text string) *ReplyKeyboard {
	kb.InputFieldPlaceholder = text
	return kb
}

// ToTgbotapi converts the ReplyKeyboard to the telegram-bot-api library format.
//
// This conversion function transforms the internal ReplyKeyboard representation
// into the format expected by the telegram-bot-api library for sending to Telegram.
// All keyboard properties and button configurations are preserved during conversion.
//
// Returns a tgbotapi.ReplyKeyboardMarkup ready for use with telegram-bot-api methods.
//
// Example:
//
//	keyboard := BuildReplyKeyboard([]string{"Option 1", "Option 2"}, 2).Resize()
//	tgKeyboard := keyboard.ToTgbotapi()
//	// Use tgKeyboard with telegram-bot-api functions
func (kb *ReplyKeyboard) ToTgbotapi() tgbotapi.ReplyKeyboardMarkup {
	// Convert internal button format to telegram-bot-api format
	var keyboard [][]tgbotapi.KeyboardButton
	for _, row := range kb.Keyboard {
		var tgRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			tgBtn := tgbotapi.KeyboardButton{
				Text:            btn.Text,
				RequestContact:  btn.RequestContact,
				RequestLocation: btn.RequestLocation,
			}
			tgRow = append(tgRow, tgBtn)
		}
		keyboard = append(keyboard, tgRow)
	}

	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard:              keyboard,
		ResizeKeyboard:        kb.ResizeKeyboard,
		OneTimeKeyboard:       kb.OneTimeKeyboard,
		InputFieldPlaceholder: kb.InputFieldPlaceholder,
		Selective:             kb.Selective,
	}
}

// ToTgbotapi converts the InlineKeyboard to the telegram-bot-api library format.
//
// This conversion function transforms the internal InlineKeyboard representation
// into the format expected by the telegram-bot-api library for sending to Telegram.
// All button properties including callback data, URLs, and inline query settings
// are preserved during conversion.
//
// Returns a tgbotapi.InlineKeyboardMarkup ready for use with telegram-bot-api methods.
//
// Example:
//
//	keyboard := &InlineKeyboard{
//		InlineKeyboard: [][]InlineKeyboardButton{
//			{{Text: "Button 1", CallbackData: "data1"}},
//		},
//	}
//	tgKeyboard := keyboard.ToTgbotapi()
//	// Use tgKeyboard with telegram-bot-api functions
func (kb *InlineKeyboard) ToTgbotapi() tgbotapi.InlineKeyboardMarkup {
	// Convert internal button format to telegram-bot-api format
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, row := range kb.InlineKeyboard {
		var tgRow []tgbotapi.InlineKeyboardButton
		for _, btn := range row {
			tgBtn := tgbotapi.InlineKeyboardButton{
				Text: btn.Text,
			}

			// Set optional fields as pointers (telegram-bot-api requirement)
			if btn.URL != "" {
				tgBtn.URL = &btn.URL
			}
			if btn.CallbackData != "" {
				tgBtn.CallbackData = &btn.CallbackData
			}
			if btn.SwitchInlineQuery != "" {
				tgBtn.SwitchInlineQuery = &btn.SwitchInlineQuery
			}
			if btn.SwitchInlineQueryCurrentChat != "" {
				tgBtn.SwitchInlineQueryCurrentChat = &btn.SwitchInlineQueryCurrentChat
			}
			tgRow = append(tgRow, tgBtn)
		}
		keyboard = append(keyboard, tgRow)
	}

	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}
