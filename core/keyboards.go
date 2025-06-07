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
	WebApp          *webAppInfo `json:"web_app,omitempty"`
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
	WebApp                       *webAppInfo `json:"web_app,omitempty"`
	SwitchInlineQuery            string      `json:"switch_inline_query,omitempty"`
	SwitchInlineQueryCurrentChat string      `json:"switch_inline_query_current_chat,omitempty"`
}

// InlineKeyboard represents an inline keyboard
type InlineKeyboard struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
	currentRow     []InlineKeyboardButton
}

// menuButtonType represents the type of native menu button
type menuButtonType string

const (
	menuButtonTypeCommands menuButtonType = "commands"
	menuButtonTypeWebApp   menuButtonType = "web_app"
	menuButtonTypeDefault  menuButtonType = "default"
)

// menuButtonItem represents a command item for menu buttons
type menuButtonItem struct {
	text    string
	command string
}

// MenuButtonConfig represents the configuration for Telegram's native menu button
type MenuButtonConfig struct {
	Type   menuButtonType   `json:"type"`
	Text   string           `json:"text,omitempty"`    // For web_app type
	WebApp *webAppInfo      `json:"web_app,omitempty"` // For web_app type
	Items  []menuButtonItem `json:"items,omitempty"`   // For commands type (not sent to API, used internally)
}

// webAppInfo represents a web app
type webAppInfo struct {
	URL string `json:"url"`
}

// BuildMenuButton creates a menu button configuration for web_app or default types only.
// For bot commands, use Bot.SetBotCommands() method instead.
//
// This function is deprecated for command-type menu buttons. Use the following pattern instead:
//
//	// Old way (deprecated):
//	// menuButton := teleflow.BuildMenuButton(map[string]string{"help": "📖 Help"})
//
//	// New way:
//	// err := bot.SetBotCommands(map[string]string{"help": "📖 Help"})
//
// BuildMenuButton now only supports creating default menu button configurations.
func BuildMenuButton(commandMap map[string]string) *MenuButtonConfig {
	// BuildMenuButton is now deprecated for commands - return default type
	// Users should use Bot.SetBotCommands() for setting bot commands
	return &MenuButtonConfig{Type: menuButtonTypeDefault}
}

// newReplyKeyboard creates a new reply keyboard (internal use)
func newReplyKeyboard(rows ...[]ReplyKeyboardButton) *ReplyKeyboard {
	kb := &ReplyKeyboard{
		Keyboard:   make([][]ReplyKeyboardButton, 0),
		currentRow: make([]ReplyKeyboardButton, 0),
	}

	kb.Keyboard = append(kb.Keyboard, rows...)

	return kb
}

// BuildReplyKeyboard creates a reply keyboard with custom buttons per row
//
// This function allows you to specify how many buttons should appear in each row.
// This provides more control over the keyboard layout compared to BuildReplyKeyboard.
//
// Parameters:
//   - buttons: slice of button texts
//   - buttonsPerRow: number of buttons to place in each row (must be > 0)
//
// Example usage:
//
//	// Create a keyboard with 3 buttons per row
//	keyboard := teleflow.BuildReplyKeyboard(
//		[]string{"A", "B", "C", "D", "E", "F", "G"}, 3)
//	// Results in: [A B C] [D E F] [G]
//
//	keyboard.Resize().OneTime()
func BuildReplyKeyboard(buttons []string, buttonsPerRow int) *ReplyKeyboard {
	if len(buttons) == 0 {
		return newReplyKeyboard()
	}

	if buttonsPerRow <= 0 {
		buttonsPerRow = 1 // Default to 1 button per row if invalid value
	}

	kb := &ReplyKeyboard{
		Keyboard:   make([][]ReplyKeyboardButton, 0),
		currentRow: make([]ReplyKeyboardButton, 0),
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
