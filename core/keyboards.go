package teleflow

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ReplyKeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact,omitempty"`
	RequestLocation bool   `json:"request_location,omitempty"`
}

type ReplyKeyboard struct {
	Keyboard              [][]ReplyKeyboardButton `json:"keyboard"`
	ResizeKeyboard        bool                    `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard       bool                    `json:"one_time_keyboard,omitempty"`
	InputFieldPlaceholder string                  `json:"input_field_placeholder,omitempty"`
	Selective             bool                    `json:"selective,omitempty"`
}

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

func (kb *ReplyKeyboard) Resize() *ReplyKeyboard {
	kb.ResizeKeyboard = true
	return kb
}

func (kb *ReplyKeyboard) OneTime() *ReplyKeyboard {
	kb.OneTimeKeyboard = true
	return kb
}

func (kb *ReplyKeyboard) Placeholder(text string) *ReplyKeyboard {
	kb.InputFieldPlaceholder = text
	return kb
}

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

// ReplyKeyboardBuilder provides a fluent interface for building ReplyKeyboard
type ReplyKeyboardBuilder struct {
	rows                  [][]ReplyKeyboardButton
	currentRow            []ReplyKeyboardButton
	resizeKeyboard        bool
	oneTimeKeyboard       bool
	inputFieldPlaceholder string
	selective             bool
}

// NewReplyKeyboard creates a new ReplyKeyboardBuilder with default values
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

// AddButton adds a standard button to the current row
func (kb *ReplyKeyboardBuilder) AddButton(text string) *ReplyKeyboardBuilder {
	button := ReplyKeyboardButton{Text: text}
	kb.currentRow = append(kb.currentRow, button)
	return kb
}

// AddContactButton adds a contact request button to the current row
func (kb *ReplyKeyboardBuilder) AddContactButton(text string) *ReplyKeyboardBuilder {
	button := ReplyKeyboardButton{
		Text:           text,
		RequestContact: true,
	}
	kb.currentRow = append(kb.currentRow, button)
	return kb
}

// AddLocationButton adds a location request button to the current row
func (kb *ReplyKeyboardBuilder) AddLocationButton(text string) *ReplyKeyboardBuilder {
	button := ReplyKeyboardButton{
		Text:            text,
		RequestLocation: true,
	}
	kb.currentRow = append(kb.currentRow, button)
	return kb
}

// Row finishes the current row and starts a new one
func (kb *ReplyKeyboardBuilder) Row() *ReplyKeyboardBuilder {
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
		kb.currentRow = make([]ReplyKeyboardButton, 0)
	}
	return kb
}

// Resize sets the keyboard to be resized
func (kb *ReplyKeyboardBuilder) Resize() *ReplyKeyboardBuilder {
	kb.resizeKeyboard = true
	return kb
}

// OneTime sets the keyboard to be one-time only
func (kb *ReplyKeyboardBuilder) OneTime() *ReplyKeyboardBuilder {
	kb.oneTimeKeyboard = true
	return kb
}

// Placeholder sets the input field placeholder text
func (kb *ReplyKeyboardBuilder) Placeholder(text string) *ReplyKeyboardBuilder {
	kb.inputFieldPlaceholder = text
	return kb
}

// Selective sets the keyboard to be selective
func (kb *ReplyKeyboardBuilder) Selective() *ReplyKeyboardBuilder {
	kb.selective = true
	return kb
}

// Build constructs and returns the final ReplyKeyboard
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
