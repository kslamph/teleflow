package teleflow

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ReplyKeyboardButton represents a single button in a reply keyboard.
// Reply keyboards appear below the message input field and can request
// special information like contact details or location.
type ReplyKeyboardButton struct {
	Text            string `json:"text"`                       // Button display text
	RequestContact  bool   `json:"request_contact,omitempty"`  // Request user's contact info
	RequestLocation bool   `json:"request_location,omitempty"` // Request user's location
}

// ReplyKeyboard represents a custom reply keyboard that replaces the user's keyboard.
// Unlike inline keyboards, reply keyboards persist until replaced or removed,
// and their buttons send regular text messages when pressed.
type ReplyKeyboard struct {
	Keyboard              [][]ReplyKeyboardButton `json:"keyboard"`                          // 2D array of keyboard buttons
	ResizeKeyboard        bool                    `json:"resize_keyboard,omitempty"`         // Resize keyboard to fit buttons
	OneTimeKeyboard       bool                    `json:"one_time_keyboard,omitempty"`       // Hide keyboard after one use
	InputFieldPlaceholder string                  `json:"input_field_placeholder,omitempty"` // Placeholder text in input field
	Selective             bool                    `json:"selective,omitempty"`               // Show only to mentioned users
}

// BuildReplyKeyboard creates a reply keyboard from a list of button texts.
// Buttons are arranged in rows based on the buttonsPerRow parameter.
// Returns nil if no buttons are provided.
//
// Example:
//
//	keyboard := teleflow.BuildReplyKeyboard([]string{"Yes", "No", "Maybe"}, 2)
//	// Creates a keyboard with "Yes" and "No" on the first row, "Maybe" on the second
func BuildReplyKeyboard(buttons []string, buttonsPerRow int) *ReplyKeyboard {
	if len(buttons) == 0 {
		return nil
	}

	if buttonsPerRow <= 0 {
		buttonsPerRow = 1
	}

	kb := &ReplyKeyboard{
		Keyboard: make([][]ReplyKeyboardButton, 0),
	}

	for i := 0; i < len(buttons); i += buttonsPerRow {
		var row []ReplyKeyboardButton

		for j := 0; j < buttonsPerRow && i+j < len(buttons); j++ {
			row = append(row, ReplyKeyboardButton{Text: buttons[i+j]})
		}

		kb.Keyboard = append(kb.Keyboard, row)
	}

	return kb
}

// Resize configures the keyboard to automatically resize to fit the buttons.
// This makes the keyboard smaller if it has fewer buttons than usual.
//
// Example:
//
//	keyboard := teleflow.BuildReplyKeyboard([]string{"Yes", "No"}, 2).Resize()
func (kb *ReplyKeyboard) Resize() *ReplyKeyboard {
	kb.ResizeKeyboard = true
	return kb
}

// OneTime configures the keyboard to hide automatically after the user presses a button.
// This is useful for one-time choices where the keyboard shouldn't persist.
//
// Example:
//
//	keyboard := teleflow.BuildReplyKeyboard([]string{"Confirm", "Cancel"}, 2).OneTime()
func (kb *ReplyKeyboard) OneTime() *ReplyKeyboard {
	kb.OneTimeKeyboard = true
	return kb
}

// Placeholder sets placeholder text that appears in the message input field.
// This text provides a hint about what the user should type.
//
// Example:
//
//	keyboard := teleflow.BuildReplyKeyboard([]string{"Option 1", "Option 2"}, 2).
//		Placeholder("Choose an option or type your own...")
func (kb *ReplyKeyboard) Placeholder(text string) *ReplyKeyboard {
	kb.InputFieldPlaceholder = text
	return kb
}

// ToTgbotapi converts the ReplyKeyboard to a tgbotapi.ReplyKeyboardMarkup
// for use with the Telegram Bot API library.
func (kb *ReplyKeyboard) ToTgbotapi() tgbotapi.ReplyKeyboardMarkup {

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

// ReplyKeyboardBuilder provides a fluent interface for building ReplyKeyboard instances.
// It allows adding buttons row by row with various button types and keyboard options.
// Use NewReplyKeyboard() to create a new builder instance.
type ReplyKeyboardBuilder struct {
	rows                  [][]ReplyKeyboardButton // Completed rows of buttons
	currentRow            []ReplyKeyboardButton   // Current row being built
	resizeKeyboard        bool                    // Whether to resize keyboard
	oneTimeKeyboard       bool                    // Whether to hide after one use
	inputFieldPlaceholder string                  // Placeholder text for input field
	selective             bool                    // Whether to show selectively
}

// NewReplyKeyboard creates a new ReplyKeyboardBuilder with default settings.
// Use this to start building a custom reply keyboard with specific button arrangements.
//
// Example:
//
//	keyboard := teleflow.NewReplyKeyboard().
//		AddButton("Home").
//		AddButton("Settings").
//		Row().
//		AddContactButton("Share Contact").
//		Build()
func NewReplyKeyboard() *ReplyKeyboardBuilder {
	return &ReplyKeyboardBuilder{
		rows:                  make([][]ReplyKeyboardButton, 0),
		currentRow:            make([]ReplyKeyboardButton, 0),
		resizeKeyboard:        false,
		oneTimeKeyboard:       false,
		inputFieldPlaceholder: "",
		selective:             false,
	}
}

// AddButton adds a standard text button to the current row.
// When pressed, this button will send its text as a regular message.
//
// Example:
//
//	keyboard.AddButton("Home").AddButton("Settings")
func (kb *ReplyKeyboardBuilder) AddButton(text string) *ReplyKeyboardBuilder {
	button := ReplyKeyboardButton{Text: text}
	kb.currentRow = append(kb.currentRow, button)
	return kb
}

// AddContactButton adds a button that requests the user's contact information.
// When pressed, this button will prompt the user to share their phone number and name.
//
// Example:
//
//	keyboard.AddContactButton("Share My Contact")
func (kb *ReplyKeyboardBuilder) AddContactButton(text string) *ReplyKeyboardBuilder {
	button := ReplyKeyboardButton{
		Text:           text,
		RequestContact: true,
	}
	kb.currentRow = append(kb.currentRow, button)
	return kb
}

// AddLocationButton adds a button that requests the user's location.
// When pressed, this button will prompt the user to share their current location.
//
// Example:
//
//	keyboard.AddLocationButton("Share My Location")
func (kb *ReplyKeyboardBuilder) AddLocationButton(text string) *ReplyKeyboardBuilder {
	button := ReplyKeyboardButton{
		Text:            text,
		RequestLocation: true,
	}
	kb.currentRow = append(kb.currentRow, button)
	return kb
}

// Row finishes the current row and starts a new one.
// This allows organizing buttons into multiple rows for better layout.
//
// Example:
//
//	keyboard.AddButton("Button 1").AddButton("Button 2").
//		Row().
//		AddButton("Button 3")  // This will be on a new row
func (kb *ReplyKeyboardBuilder) Row() *ReplyKeyboardBuilder {
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
		kb.currentRow = make([]ReplyKeyboardButton, 0)
	}
	return kb
}

// Resize configures the keyboard to automatically resize to fit the buttons.
// This makes the keyboard take up less space when it has fewer buttons.
//
// Example:
//
//	keyboard.AddButton("Yes").AddButton("No").Resize()
func (kb *ReplyKeyboardBuilder) Resize() *ReplyKeyboardBuilder {
	kb.resizeKeyboard = true
	return kb
}

// OneTime configures the keyboard to hide after the user presses any button.
// This is useful for one-time confirmations or selections.
//
// Example:
//
//	keyboard.AddButton("Confirm").AddButton("Cancel").OneTime()
func (kb *ReplyKeyboardBuilder) OneTime() *ReplyKeyboardBuilder {
	kb.oneTimeKeyboard = true
	return kb
}

// Placeholder sets placeholder text that appears in the message input field.
// This provides a hint about what the user should type or do.
//
// Example:
//
//	keyboard.AddButton("Option 1").Placeholder("Choose an option or type custom text")
func (kb *ReplyKeyboardBuilder) Placeholder(text string) *ReplyKeyboardBuilder {
	kb.inputFieldPlaceholder = text
	return kb
}

// Selective configures the keyboard to be shown only to specific users.
// This is typically used in group chats to show the keyboard only to mentioned users.
//
// Example:
//
//	keyboard.AddButton("Admin Only").Selective()
func (kb *ReplyKeyboardBuilder) Selective() *ReplyKeyboardBuilder {
	kb.selective = true
	return kb
}

// Build constructs and returns the final ReplyKeyboard instance.
// This finalizes the keyboard configuration and makes it ready for use.
// Any buttons in the current row are automatically added to the final keyboard.
//
// Example:
//
//	keyboard := teleflow.NewReplyKeyboard().
//		AddButton("Home").AddButton("Settings").
//		Row().
//		AddContactButton("Share Contact").
//		Build()
func (kb *ReplyKeyboardBuilder) Build() *ReplyKeyboard {
	// If there's a current row with buttons, add it to rows
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
	}

	return &ReplyKeyboard{
		Keyboard:              kb.rows,
		ResizeKeyboard:        kb.resizeKeyboard,
		OneTimeKeyboard:       kb.oneTimeKeyboard,
		InputFieldPlaceholder: kb.inputFieldPlaceholder,
		Selective:             kb.selective,
	}
}
