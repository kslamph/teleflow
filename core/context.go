package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ParseMode specifies the format of a text message. Use the typed constants
// (ParseModeHTML, ParseModeMarkdown, ParseModeMarkdownV2) instead of raw strings.
type ParseMode string

const (
	ParseModeHTML       ParseMode = "HTML"
	ParseModeMarkdown   ParseMode = "Markdown"
	ParseModeMarkdownV2 ParseMode = "MarkdownV2"
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
// The handler is executed once when the button is pressed, then automatically
// removed. Callbacks are NEVER auto-answered by the framework; you must call
// ctx.AnswerCallback() inside your handler to stop the client loading spinner.
// Unmatched or stale callback queries silently fall through to the router's
// default route.
func (c *Context) RegisterCallback(handler func(ctx *Context) error) string {
	return c.bot.registerCallback(handler)
}

// RegisterData registers an arbitrary data object for one-time retrieval via a
// callback query. It returns a unique ID to be used as the callback data. This
// allows passing complex data through buttons, bypassing Telegram's size limit.
// The data is automatically deleted on the first call to GetData with the same ID.
func (c *Context) RegisterData(data interface{}) string {
	return c.bot.RegisterData(data)
}

// GetData retrieves and deletes a data object that was previously stored via
// RegisterData. It uses the data from the current callback query as the lookup key.
// The lookup key (the callback data string) is consumed on the first call; a
// subsequent call for the same key returns ok=false.
func (c *Context) GetData() (interface{}, bool) {
	if c.update.CallbackQuery == nil {
		return nil, false
	}
	return c.bot.GetData(c.update.CallbackQuery.Data)
}

// ReplyTextWithParseMode sends a text message with the given parse mode to
// the originating chat. Argument order: text first, then parseMode.
func (c *Context) ReplyTextWithParseMode(text string, parseMode ParseMode) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	msg.ParseMode = string(parseMode)
	_, err := c.api.Send(msg)
	return err
}

// ReplyText sends a simple text message to the originating chat.
func (c *Context) ReplyText(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	_, err := c.api.Send(msg)
	return err
}

// ReplyWithKeyboardAndParseMode sends a message with a keyboard markup and parse mode.
// Argument order: text, markup, parseMode. The markup can be an
// InlineKeyboardMarkup, ReplyKeyboardMarkup, or ReplyKeyboardRemove.
func (c *Context) ReplyWithKeyboardAndParseMode(text string, markup interface{}, parseMode ParseMode) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	msg.ReplyMarkup = markup
	msg.ParseMode = string(parseMode)
	_, err := c.api.Send(msg)
	return err
}

// ReplyWithKeyboard sends a message with a keyboard markup.
// Argument order: text first, then markup. The markup can be an
// InlineKeyboardMarkup, ReplyKeyboardMarkup, or ReplyKeyboardRemove.
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
// Edit* methods take messageID first, then the new content.
func (c *Context) EditMessageText(messageID int, text string) error {
	editMsg := tgbotapi.NewEditMessageText(c.ChatID(), messageID, text)
	_, err := c.api.Request(editMsg)
	return err
}

// EditMessageTextWithParseMode edits the text of a previously sent message with parse mode support.
// Edit* methods take messageID first, then the new content.
func (c *Context) EditMessageTextWithParseMode(messageID int, text string, parseMode ParseMode) error {
	editMsg := tgbotapi.NewEditMessageText(c.ChatID(), messageID, text)
	editMsg.ParseMode = string(parseMode)
	_, err := c.api.Request(editMsg)
	return err
}

// EditMessageReplyMarkup edits the inline keyboard of a previously sent message.
// Edit* methods take messageID first, then the new markup.
func (c *Context) EditMessageReplyMarkup(messageID int, markup tgbotapi.InlineKeyboardMarkup) error {
	config := tgbotapi.NewEditMessageReplyMarkup(c.ChatID(), messageID, markup)
	_, err := c.api.Request(config)
	return err
}

// EditMessageMedia edits the media of a previously sent message.
// Edit* methods take messageID first, then the new media.
func (c *Context) EditMessageMedia(messageID int, media interface{}) error {
	config := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    c.ChatID(),
			MessageID: messageID,
		},
		Media: media,
	}
	_, err := c.api.Request(config)
	return err
}

// AnswerCallback acknowledges a callback query. This MUST be called after a user
// presses an inline keyboard button to stop the loading animation on their client.
// The framework never auto-answers callbacks.
func (c *Context) AnswerCallback() error {
	if c.update.CallbackQuery == nil {
		return fmt.Errorf("not a callback query")
	}
	callback := tgbotapi.NewCallback(c.update.CallbackQuery.ID, "")
	_, err := c.api.Request(callback)
	return err
}
