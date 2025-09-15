package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context provides a convenient interface for handling a Telegram update.
type Context struct {
	update tgbotapi.Update
	api    *tgbotapi.BotAPI
	bot    *Bot
	data   map[string]interface{}
}

// NewContext creates a new Context for an update.
func NewContext(update tgbotapi.Update, api *tgbotapi.BotAPI, bot *Bot) *Context {
	return &Context{
		update: update,
		api:    api,
		bot:    bot,
		data:   make(map[string]interface{}),
	}
}

// Update returns the raw Telegram update.
func (c *Context) Update() tgbotapi.Update {
	return c.update
}

// API returns the underlying Bot API instance.
func (c *Context) API() *tgbotapi.BotAPI {
	return c.api
}

// UserID returns the ID of the user who initiated the update.
func (c *Context) UserID() int64 {
	if c.update.Message != nil {
		return c.update.Message.From.ID
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.From.ID
	}
	return 0
}

// ChatID returns the ID of the chat where the update originated.
func (c *Context) ChatID() int64 {
	if c.update.Message != nil {
		return c.update.Message.Chat.ID
	}
	if c.update.CallbackQuery != nil && c.update.CallbackQuery.Message != nil {
		return c.update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

// Text returns the text from a message or the data from a callback query.
func (c *Context) Text() string {
	if c.update.Message != nil {
		return c.update.Message.Text
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.Data
	}
	return ""
}

// Set stores a key-value pair in the context's data store for the lifetime of the update.
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get retrieves a value from the context's data store.
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

// RegisterCallback registers a one-time handler for a callback query and returns the unique ID.
func (c *Context) RegisterCallback(handler func(ctx *Context) error) string {
	return c.bot.registerCallback(handler)
}

// RegisterData registers arbitrary data for one-time retrieval and returns a unique ID.
func (c *Context) RegisterData(data interface{}) string {
	return c.bot.RegisterData(data)
}

// GetData retrieves the data associated with the current callback query's data field.
func (c *Context) GetData() (interface{}, bool) {
	if c.update.CallbackQuery == nil {
		return nil, false
	}
	return c.bot.GetData(c.update.CallbackQuery.Data)
}

// ReplyText sends a simple text message back to the user.
func (c *Context) ReplyText(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	_, err := c.api.Send(msg)
	return err
}

// ReplyInlineKeyboard sends a message with a keyboard.
func (c *Context) ReplyInlineKeyboard(text string, markup interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	msg.ReplyMarkup = markup
	_, err := c.api.Send(msg)
	return err
}

// DeleteMessage deletes a message by its ID in the current chat.
func (c *Context) DeleteMessage(messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(c.ChatID(), messageID)
	_, err := c.api.Request(deleteMsg)
	return err
}

// EditMessageText edits the text of a message.
func (c *Context) EditMessageText(messageID int, text string) error {
	editMsg := tgbotapi.NewEditMessageText(c.ChatID(), messageID, text)
	_, err := c.api.Request(editMsg)
	return err
}

// EditMessageReplyMarkup edits the reply markup of a message.
func (c *Context) EditMessageReplyMarkup(messageID int, markup *tgbotapi.InlineKeyboardMarkup) error {
	config := tgbotapi.NewEditMessageReplyMarkup(c.ChatID(), messageID, *markup)
	_, err := c.api.Request(config)
	return err
}

// AnswerCallback answers a callback query, required after a user presses an inline button.
func (c *Context) AnswerCallback() error {
	if c.update.CallbackQuery == nil {
		return fmt.Errorf("not a callback query")
	}
	callback := tgbotapi.NewCallback(c.update.CallbackQuery.ID, "")
	_, err := c.api.Request(callback)
	return err
}
