package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type PromptKeyboardBuilder struct {
	rows        [][]tgbotapi.InlineKeyboardButton
	currentRow  []tgbotapi.InlineKeyboardButton
	uuidMapping map[string]interface{}
}

func NewPromptKeyboard() *PromptKeyboardBuilder {
	return &PromptKeyboardBuilder{
		rows:        make([][]tgbotapi.InlineKeyboardButton, 0),
		currentRow:  make([]tgbotapi.InlineKeyboardButton, 0),
		uuidMapping: make(map[string]interface{}),
	}
}

func (kb *PromptKeyboardBuilder) ButtonCallback(text string, data interface{}) *PromptKeyboardBuilder {

	callbackUUID := uuid.New().String()

	kb.uuidMapping[callbackUUID] = data

	button := tgbotapi.NewInlineKeyboardButtonData(text, callbackUUID)
	kb.currentRow = append(kb.currentRow, button)

	return kb
}

func (kb *PromptKeyboardBuilder) ButtonUrl(text string, url string) *PromptKeyboardBuilder {
	button := tgbotapi.NewInlineKeyboardButtonURL(text, url)
	kb.currentRow = append(kb.currentRow, button)

	return kb
}

func (kb *PromptKeyboardBuilder) Row() *PromptKeyboardBuilder {
	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
		kb.currentRow = make([]tgbotapi.InlineKeyboardButton, 0)
	}
	return kb
}

func (kb *PromptKeyboardBuilder) Build() tgbotapi.InlineKeyboardMarkup {

	if len(kb.currentRow) > 0 {
		kb.rows = append(kb.rows, kb.currentRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(kb.rows...)
}

func (kb *PromptKeyboardBuilder) validateBuilder() error {
	totalButtons := len(kb.currentRow)
	for _, row := range kb.rows {
		totalButtons += len(row)
	}

	if totalButtons == 0 {
		return fmt.Errorf("inline keyboard must have at least one button")
	}

	return nil
}
