package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context provides a convenient, user-facing API for handling a single Telegram
// update. It acts as a toolkit, offering helper methods to access update data
// and perform common bot actions (e.g., sending replies) without needing to
// interact directly with the low-level Telegram API. A new Context is created
// for each incoming update.
type Context struct {
	update tgbotapi.Update
	api    *tgbotapi.BotAPI
	bot    *Bot
	data   map[string]interface{}
}

// NewContext creates a new Context for a given update, Bot API instance, and
// Bot instance. It is used internally by the Bot to construct the context for
// each update.
func NewContext(update tgbotapi.Update, api *tgbotapi.BotAPI, bot *Bot) *Context {
	return &Context{
		update: update,
		api:    api,
		bot:    bot,
		data:   make(map[string]interface{}),
	}
}

// Update returns the raw, underlying `tgbotapi.Update` object.
func (c *Context) Update() tgbotapi.Update {
	return c.update
}

// API returns the underlying `tgbotapi.BotAPI` instance, allowing for direct
// access to the Telegram API for advanced use cases.
func (c *Context) API() *tgbotapi.BotAPI {
	return c.api
}

// UserID returns the unique identifier for the user who initiated the update.
// It gracefully handles both messages and callback queries. Returns 0 if the
// user cannot be identified.
func (c *Context) UserID() int64 {
	if c.update.Message != nil {
		return c.update.Message.From.ID
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.From.ID
	}
	return 0
}

// ChatID returns the unique identifier for the chat where the update originated.
// It handles both messages and callback queries. Returns 0 if the chat cannot
// be identified.
func (c *Context) ChatID() int64 {
	if c.update.Message != nil {
		return c.update.Message.Chat.ID
	}
	if c.update.CallbackQuery != nil && c.update.CallbackQuery.Message != nil {
		return c.update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

// Text returns the text content from a message or the data from a callback query.
// This provides a unified way to get the primary string content of an update.
func (c *Context) Text() string {
	if c.update.Message != nil {
		return c.update.Message.Text
	}
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery.Data
	}
	return ""
}

// Set stores a key-value pair in the context's ephemeral data store. This data
// persists only for the lifetime of the current update's processing. It is
// useful for passing data between middleware and handlers.
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get retrieves a value from the context's ephemeral data store.
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

// RegisterCallback registers a one-time handler for a callback query. It returns
// a unique ID to be used as the callback data in a `tgbotapi.InlineKeyboardButton`.
// The handler will be executed once when the button is pressed, then automatically
// removed.
func (c *Context) RegisterCallback(handler func(ctx *Context) error) string {
	return c.bot.registerCallback(handler)
}

// RegisterData registers an arbitrary data object for one-time retrieval via a
// callback query. It returns a unique ID to be used as the callback data. This
// allows passing complex data through buttons, bypassing Telegram's size limit.
// The data is automatically deleted after being retrieved by GetData.
func (c *Context) RegisterData(data interface{}) string {
	return c.bot.RegisterData(data)
}

// GetData retrieves and deletes a data object that was previously stored via
// RegisterData. It uses the data from the current callback query as the lookup key.
func (c *Context) GetData() (interface{}, bool) {
	if c.update.CallbackQuery == nil {
		return nil, false
	}
	return c.bot.GetData(c.update.CallbackQuery.Data)
}

// ReplyText sends a simple text message to the originating chat.
func (c *Context) ReplyText(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	_, err := c.api.Send(msg)
	return err
}

// ReplyWithKeyboard sends a message with a keyboard markup. The markup can be
// either a `tgbotapi.InlineKeyboardMarkup`, `tgbotapi.ReplyKeyboardMarkup`,
// or `tgbotapi.ReplyKeyboardRemove`.
func (c *Context) ReplyWithKeyboard(text string, markup interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	msg.ReplyMarkup = markup
	_, err := c.api.Send(msg)
	return err
}

// DeleteMessage deletes a message by its ID from the originating chat.
func (c *Context) DeleteMessage(messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(c.ChatID(), messageID)
	_, err := c.api.Request(deleteMsg)
	return err
}

// EditMessageText edits the text of a previously sent message.
func (c *Context) EditMessageText(messageID int, text string) error {
	editMsg := tgbotapi.NewEditMessageText(c.ChatID(), messageID, text)
	_, err := c.api.Request(editMsg)
	return err
}

// EditMessageReplyMarkup edits the inline keyboard of a previously sent message.
func (c *Context) EditMessageReplyMarkup(messageID int, markup *tgbotapi.InlineKeyboardMarkup) error {
	config := tgbotapi.NewEditMessageReplyMarkup(c.ChatID(), messageID, *markup)
	_, err := c.api.Request(config)
	return err
}

// AnswerCallback acknowledges a callback query. This must be called after a user
// presses an inline keyboard button to stop the loading animation on their client.
func (c *Context) AnswerCallback() error {
	if c.update.CallbackQuery == nil {
		return fmt.Errorf("not a callback query")
	}
	callback := tgbotapi.NewCallback(c.update.CallbackQuery.ID, "")
	_, err := c.api.Request(callback)
	return err
}
