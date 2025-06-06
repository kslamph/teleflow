package teleflow

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Keyboard system provides intuitive abstractions for creating and managing
// Telegram reply keyboards and inline keyboards. The Step-Prompt-Process API
// greatly simplifies keyboard creation using map-based approaches for common use cases.
//
// Reply Keyboards appear below the message input field and send their text
// as regular messages when pressed. They're ideal for main menus, options
// selection, and persistent navigation elements.
//
// Inline Keyboards appear directly below messages as clickable buttons.
// They support callback data, URLs, web apps, and other interactive elements
// without sending text messages.
//
// Simple Map-based Keyboards :
//
//	// In flow steps, use simple map syntax for inline keyboards
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

// Special Button Types (Still supported for complex use cases):
//
//	// Request contact information
//	keyboard.AddRow(teleflow.NewReplyButton("📱 Share Contact").SetRequestContact())
//
//	// Request location
//	keyboard.AddRow(teleflow.NewReplyButton("📍 Share Location").SetRequestLocation())
//
//	// Web app integration
//	keyboard.AddRow(teleflow.NewReplyButton("🌐 Open App").WithWebApp("https://app.example.com"))

// ReplyKeyboardButton represents a button in a reply keyboard
type ReplyKeyboardButton struct {
	Text            string      `json:"text"`
	RequestContact  bool        `json:"request_contact,omitempty"`
	RequestLocation bool        `json:"request_location,omitempty"`
	WebApp          *WebAppInfo `json:"web_app,omitempty"`
}

// ReplyKeyboard represents a custom reply keyboard
type ReplyKeyboard struct {
	Keyboard              [][]ReplyKeyboardButton `json:"keyboard"`
	ResizeKeyboard        bool                    `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard       bool                    `json:"one_time_keyboard,omitempty"`
	InputFieldPlaceholder string                  `json:"input_field_placeholder,omitempty"`
	Selective             bool                    `json:"selective,omitempty"`
	currentRow            []ReplyKeyboardButton
}

// InlineKeyboardButton represents a button in an inline keyboard
type InlineKeyboardButton struct {
	Text                         string      `json:"text"`
	URL                          string      `json:"url,omitempty"`
	CallbackData                 string      `json:"callback_data,omitempty"`
	WebApp                       *WebAppInfo `json:"web_app,omitempty"`
	SwitchInlineQuery            string      `json:"switch_inline_query,omitempty"`
	SwitchInlineQueryCurrentChat string      `json:"switch_inline_query_current_chat,omitempty"`
}

// InlineKeyboard represents an inline keyboard
type InlineKeyboard struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
	currentRow     []InlineKeyboardButton
}

// MenuButtonType represents the type of native menu button
type MenuButtonType string

const (
	MenuButtonTypeCommands MenuButtonType = "commands"
	MenuButtonTypeWebApp   MenuButtonType = "web_app"
	MenuButtonTypeDefault  MenuButtonType = "default"
)

// MenuButtonItem represents a command item for menu buttons
type MenuButtonItem struct {
	Text    string `json:"text"`
	Command string `json:"command"`
}

// MenuButtonConfig represents the configuration for Telegram's native menu button
type MenuButtonConfig struct {
	Type   MenuButtonType   `json:"type"`
	Text   string           `json:"text,omitempty"`    // For web_app type
	WebApp *WebAppInfo      `json:"web_app,omitempty"` // For web_app type
	Items  []MenuButtonItem `json:"items,omitempty"`   // For commands type (not sent to API, used internally)
}

// WebAppInfo represents a web app
type WebAppInfo struct {
	URL string `json:"url"`
}

// NewReplyKeyboard creates a new reply keyboard
//
// Reply keyboards appear below the message input field and are used for persistent
// navigation menus, main menu options, and other UI elements that should remain
// available to the user. They work with the AccessManager.GetReplyKeyboard() method
// to provide context-aware keyboard layouts.
//
// This is different from inline keyboards used in flow steps - those use the new
// map-based approach in the Step-Prompt-Process API.
func NewReplyKeyboard(rows ...[]ReplyKeyboardButton) *ReplyKeyboard {
	kb := &ReplyKeyboard{
		Keyboard:   make([][]ReplyKeyboardButton, 0),
		currentRow: make([]ReplyKeyboardButton, 0),
	}

	kb.Keyboard = append(kb.Keyboard, rows...)

	return kb
}

// NewReplyButton creates a new reply keyboard button with the given text
func NewReplyButton(text string) *ReplyKeyboardButton {
	return &ReplyKeyboardButton{
		Text: text,
	}
}

// SetRequestContact sets the button to request contact information
func (btn *ReplyKeyboardButton) SetRequestContact() *ReplyKeyboardButton {
	btn.RequestContact = true
	return btn
}

// SetRequestLocation sets the button to request location information
func (btn *ReplyKeyboardButton) SetRequestLocation() *ReplyKeyboardButton {
	btn.RequestLocation = true
	return btn
}

// WithWebApp sets the button as a web app button
func (btn *ReplyKeyboardButton) WithWebApp(url string) *ReplyKeyboardButton {
	btn.WebApp = &WebAppInfo{URL: url}
	return btn
}

// AddRow adds a new row of buttons to the keyboard
func (kb *ReplyKeyboard) AddRow(buttons ...*ReplyKeyboardButton) *ReplyKeyboard {
	// Convert pointers to values for the row
	var row []ReplyKeyboardButton
	for _, btn := range buttons {
		if btn != nil {
			row = append(row, *btn)
		}
	}
	if len(row) > 0 {
		kb.Keyboard = append(kb.Keyboard, row)
	}
	return kb
}

// AddButton adds a button to the current row
func (kb *ReplyKeyboard) AddButton(text string) *ReplyKeyboard {
	kb.currentRow = append(kb.currentRow, ReplyKeyboardButton{Text: text})
	return kb
}

// AddRequestContact adds a contact request button to the current row
func (kb *ReplyKeyboard) AddRequestContact() *ReplyKeyboard {
	kb.currentRow = append(kb.currentRow, ReplyKeyboardButton{
		Text:           "Share Contact",
		RequestContact: true,
	})
	return kb
}

// AddRequestLocation adds a location request button to the current row
func (kb *ReplyKeyboard) AddRequestLocation() *ReplyKeyboard {
	kb.currentRow = append(kb.currentRow, ReplyKeyboardButton{
		Text:            "Share Location",
		RequestLocation: true,
	})
	return kb
}

// Resize sets the resize keyboard flag
func (kb *ReplyKeyboard) Resize() *ReplyKeyboard {
	kb.ResizeKeyboard = true
	return kb
}

// OneTime sets the one time keyboard flag
func (kb *ReplyKeyboard) OneTime() *ReplyKeyboard {
	kb.OneTimeKeyboard = true
	return kb
}

// Placeholder sets the input field placeholder text
func (kb *ReplyKeyboard) Placeholder(text string) *ReplyKeyboard {
	kb.InputFieldPlaceholder = text
	return kb
}

// ToTgbotapi converts the reply keyboard to telegram-bot-api format
func (kb *ReplyKeyboard) ToTgbotapi() tgbotapi.ReplyKeyboardMarkup {
	// Add any remaining buttons in currentRow
	if len(kb.currentRow) > 0 {
		kb.Keyboard = append(kb.Keyboard, kb.currentRow)
		kb.currentRow = make([]ReplyKeyboardButton, 0)
	}

	// Convert to tgbotapi format
	var keyboard [][]tgbotapi.KeyboardButton
	for _, row := range kb.Keyboard {
		var tgRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			tgBtn := tgbotapi.KeyboardButton{
				Text:            btn.Text,
				RequestContact:  btn.RequestContact,
				RequestLocation: btn.RequestLocation,
			}
			// Note: WebApp support may vary by telegram-bot-api version
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

// ToTgbotapi converts the inline keyboard to telegram-bot-api format
func (kb *InlineKeyboard) ToTgbotapi() tgbotapi.InlineKeyboardMarkup {
	// Add any remaining buttons in currentRow
	if len(kb.currentRow) > 0 {
		kb.InlineKeyboard = append(kb.InlineKeyboard, kb.currentRow)
		kb.currentRow = make([]InlineKeyboardButton, 0)
	}
	// Convert to tgbotapi format
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, row := range kb.InlineKeyboard {
		var tgRow []tgbotapi.InlineKeyboardButton
		for _, btn := range row {
			tgBtn := tgbotapi.InlineKeyboardButton{
				Text: btn.Text,
			}

			// Set optional fields as pointers
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
			// Note: WebApp support may vary by telegram-bot-api version

			tgRow = append(tgRow, tgBtn)
		}
		keyboard = append(keyboard, tgRow)
	}

	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}
