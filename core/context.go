package teleflow

import (
	"bytes"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context provides comprehensive information and helper methods for handling
// Telegram bot interactions. It encapsulates the current update, provides
// access to the bot instance, and offers convenient methods for common
// operations like sending replies, managing keyboards, and handling user data.
//
// The Context is passed to all handler functions and contains:
//   - The current Telegram update (message, callback query, etc.)
//   - Bot instance for advanced operations
//   - User and chat identification
//   - Key-value data storage for request-scoped information
//   - Helper methods for replying, keyboard management, and state operations
//
// Example usage in handlers:
//
//	bot.HandleCommand("/start", func(ctx *teleflow.Context) error {
//		// Get user information
//		userID := ctx.UserID()
//
//		// Store temporary data
//		ctx.Set("step", "greeting")
//
//		// Reply with keyboard
//		keyboard := teleflow.NewReplyKeyboard().AddRow("Help", "Settings")
//		return ctx.ReplyWithKeyboard("Welcome!", keyboard)
//	})
//
//	bot.HandleCallback("button_*", func(ctx *teleflow.Context) error {
//		// Extract callback data
//		data := ctx.CallbackData()
//
//		// Update message
//		return ctx.EditMessage("Button clicked: " + data)
//	})

// Context provides information and helpers for the current interaction
type Context struct {
	Bot    *Bot
	Update tgbotapi.Update
	data   map[string]interface{}

	// User context
	userID int64
	chatID int64
}

// NewContext creates a new Context
func NewContext(bot *Bot, update tgbotapi.Update) *Context {
	ctx := &Context{
		Bot:    bot,
		Update: update,
		data:   make(map[string]interface{}),
	}

	ctx.userID = ctx.extractUserID(update)
	ctx.chatID = ctx.extractChatID(update)

	return ctx
}

// UserID returns the ID of the user who initiated the update
func (c *Context) UserID() int64 {
	return c.userID
}

// ChatID returns the ID of the chat where the update originated
func (c *Context) ChatID() int64 {
	return c.chatID
}

// Set stores a value in the context's data map
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get retrieves a value from the context's data map
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

// Reply sends a text message with appropriate keyboard for user
func (c *Context) Reply(text string, keyboard ...interface{}) error {
	return c.send(text, keyboard...)
}

// ReplyTemplate sends a text message using a template
func (c *Context) ReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error {
	var buf bytes.Buffer
	if err := c.Bot.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return fmt.Errorf("executing template %s: %w", templateName, err)
	}
	return c.send(buf.String(), keyboard...)
}

// EditOrReply attempts to edit current message or sends new one
func (c *Context) EditOrReply(text string, keyboard ...interface{}) error {
	if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
		msg := tgbotapi.NewEditMessageText(
			c.ChatID(),
			c.Update.CallbackQuery.Message.MessageID,
			text,
		)

		if len(keyboard) > 0 && keyboard[0] != nil {
			switch kb := keyboard[0].(type) {
			case *InlineKeyboard:
				markup := kb.ToTgbotapi()
				msg.ReplyMarkup = &markup
			case tgbotapi.InlineKeyboardMarkup:
				msg.ReplyMarkup = &kb
			}
		}

		if _, err := c.Bot.api.Send(msg); err == nil {
			cb := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, "")
			c.Bot.api.Request(cb)
			return nil
		}
	}
	return c.Reply(text, keyboard...)
}

// StartFlow initiates a new flow for the user
func (c *Context) StartFlow(flowName string) error {
	// StartFlow now takes a Context parameter instead of initialData
	return c.Bot.flowManager.StartFlow(c.UserID(), flowName, c)
}

// IsUserInFlow checks if the user is currently in a flow
func (c *Context) IsUserInFlow() bool {
	return c.Bot.flowManager.IsUserInFlow(c.UserID())
}

// CancelFlow cancels the current flow for the user
func (c *Context) CancelFlow() error {
	return c.Bot.flowManager.CancelFlow(c.UserID())
}

// send is an internal helper for sending messages
func (c *Context) send(text string, keyboard ...interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)

	// Apply keyboard markup
	if len(keyboard) > 0 && keyboard[0] != nil {
		switch kb := keyboard[0].(type) {
		case *ReplyKeyboard:
			msg.ReplyMarkup = kb.ToTgbotapi()
		case *InlineKeyboard:
			msg.ReplyMarkup = kb.ToTgbotapi()
		case tgbotapi.ReplyKeyboardRemove:
			msg.ReplyMarkup = kb
		case tgbotapi.ReplyKeyboardMarkup:
			msg.ReplyMarkup = kb
		case tgbotapi.InlineKeyboardMarkup:
			msg.ReplyMarkup = kb
		}
	} else {
		// Apply user-specific main menu
		if c.Bot.userPermissions != nil {
			if userMenu := c.Bot.userPermissions.GetMainMenuForUser(c.UserID()); userMenu != nil {
				msg.ReplyMarkup = userMenu.ToTgbotapi()
			}
		} else if c.Bot.mainMenu != nil {
			msg.ReplyMarkup = c.Bot.mainMenu.ToTgbotapi()
		}
	}

	_, err := c.Bot.api.Send(msg)
	return err
}

// extractUserID extracts user ID from update
func (c *Context) extractUserID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

// extractChatID extracts chat ID from update
func (c *Context) extractChatID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}
