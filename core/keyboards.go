package teleflow

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Keyboard system provides intuitive abstractions for creating and managing
// Telegram reply keyboards and inline keyboards. The system supports both
// simple text-based keyboards and complex interactive keyboards with callbacks,
// web apps, and special request buttons.
//
// Reply Keyboards appear below the message input field and send their text
// as regular messages when pressed. They're ideal for main menus, options
// selection, and persistent navigation elements.
//
// Inline Keyboards appear directly below messages as clickable buttons.
// They support callback data, URLs, web apps, and other interactive elements
// without sending text messages.
//
// Example - Reply Keyboard:
//
//	keyboard := teleflow.NewReplyKeyboard().
//		AddRow("🏠 Home", "📊 Stats").
//		AddRow("⚙️ Settings").
//		SetResizable(true).
//		SetOneTime(false)
//
//	ctx.ReplyWithKeyboard("Choose an option:", keyboard)
//
// Example - Inline Keyboard:
//
//	keyboard := teleflow.NewInlineKeyboard().
//		AddCallbackRow("✅ Approve", "approve_123", "❌ Reject", "reject_123").
//		AddURLRow("📖 Documentation", "https://docs.example.com").
//		AddRow(teleflow.NewInlineButton("🔧 Settings", "settings"))
//
//	ctx.ReplyWithInlineKeyboard("Please review:", keyboard)
//
// Special Button Types:
//
//	// Request contact information
//	keyboard.AddRow(teleflow.NewReplyButton("📱 Share Contact").RequestContact())
//
//	// Request location
//	keyboard.AddRow(teleflow.NewReplyButton("📍 Share Location").RequestLocation())
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

// MenuButtonConfig represents a menu button configuration
type MenuButtonConfig struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

// WebAppInfo represents a web app
type WebAppInfo struct {
	URL string `json:"url"`
}

// NewReplyKeyboard creates a new reply keyboard
func NewReplyKeyboard(rows ...[]ReplyKeyboardButton) *ReplyKeyboard {
	kb := &ReplyKeyboard{
		Keyboard:   make([][]ReplyKeyboardButton, 0),
		currentRow: make([]ReplyKeyboardButton, 0),
	}

	for _, row := range rows {
		kb.Keyboard = append(kb.Keyboard, row)
	}

	return kb
}

// NewInlineKeyboard creates a new inline keyboard
func NewInlineKeyboard(rows ...[]InlineKeyboardButton) *InlineKeyboard {
	kb := &InlineKeyboard{
		InlineKeyboard: make([][]InlineKeyboardButton, 0),
		currentRow:     make([]InlineKeyboardButton, 0),
	}

	for _, row := range rows {
		kb.InlineKeyboard = append(kb.InlineKeyboard, row)
	}

	return kb
}

// AddRow adds a new row to the reply keyboard
func (kb *ReplyKeyboard) AddRow() *ReplyKeyboard {
	if len(kb.currentRow) > 0 {
		kb.Keyboard = append(kb.Keyboard, kb.currentRow)
		kb.currentRow = make([]ReplyKeyboardButton, 0)
	}
	return kb
}

// AddButton adds a button to the current row of the reply keyboard
func (kb *ReplyKeyboard) AddButton(text string) *ReplyKeyboard {
	kb.currentRow = append(kb.currentRow, ReplyKeyboardButton{
		Text: text,
	})
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

// AddRow adds a new row to the inline keyboard
func (kb *InlineKeyboard) AddRow() *InlineKeyboard {
	if len(kb.currentRow) > 0 {
		kb.InlineKeyboard = append(kb.InlineKeyboard, kb.currentRow)
		kb.currentRow = make([]InlineKeyboardButton, 0)
	}
	return kb
}

// AddButton adds a callback button to the current row of the inline keyboard
func (kb *InlineKeyboard) AddButton(text, data string) *InlineKeyboard {
	kb.currentRow = append(kb.currentRow, InlineKeyboardButton{
		Text:         text,
		CallbackData: data,
	})
	return kb
}

// AddURL adds a URL button to the current row of the inline keyboard
func (kb *InlineKeyboard) AddURL(text, url string) *InlineKeyboard {
	kb.currentRow = append(kb.currentRow, InlineKeyboardButton{
		Text: text,
		URL:  url,
	})
	return kb
}

// AddWebApp adds a web app button to the current row of the inline keyboard
func (kb *InlineKeyboard) AddWebApp(text string, webApp WebAppInfo) *InlineKeyboard {
	kb.currentRow = append(kb.currentRow, InlineKeyboardButton{
		Text:   text,
		WebApp: &webApp,
	})
	return kb
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
