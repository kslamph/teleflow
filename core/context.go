package teleflow

import (
	"bytes"
	"fmt"
	"log"

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
//	bot.HandleCommand("/start", func(ctx *teleflow.Context, command string, args string) error {
//		// Start a flow using the new Step-Prompt-Process API
//		return ctx.StartFlow("registration")
//	})
//
//	// In flow steps, use unified ProcessFunc instead of separate callback handlers
//	.Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//		// Handle both text input and button clicks in one function
//		if buttonClick != nil {
//			// Button was clicked
//			return teleflow.NextStep()
//		}
//		// Text was typed
//		return teleflow.RetryWithPrompt(&teleflow.PromptConfig{Message: "Please use the buttons"})
//	})

// Context provides information and helpers for the current interaction
type Context struct {
	Bot    *Bot
	Update tgbotapi.Update
	data   map[string]interface{}

	// User context
	userID    int64
	chatID    int64
	isGroup   bool
	isChannel bool
}

type MenuContext struct {
	UserID    int64
	ChatID    int64
	IsGroup   bool
	IsChannel bool
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
	ctx.isGroup = update.Message != nil && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup())
	ctx.isChannel = update.Message != nil && update.Message.Chat.IsChannel()

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
	text, parseMode, err := c.executeTemplate(templateName, data)
	if err != nil {
		return err
	}
	return c.sendWithParseMode(text, parseMode, keyboard...)
}

// EditOrReplyTemplate attempts to edit current message using template or sends new one
func (c *Context) EditOrReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error {
	text, parseMode, err := c.executeTemplate(templateName, data)
	if err != nil {
		return err
	}
	return c.editOrReplyWithParseMode(text, parseMode, keyboard...)
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
// This is an idempotent operation
func (c *Context) CancelFlow() {
	c.Bot.flowManager.CancelFlow(c.UserID())
}

// SendPrompt renders and sends a PromptConfig message without keyboard interaction.
// This is useful for informational messages that use the same rendering system
// as flow prompts but don't require user interaction.
func (c *Context) SendPrompt(prompt *PromptConfig) error {
	if c.Bot.flowManager.promptRenderer == nil {
		return fmt.Errorf("PromptRenderer not initialized - call InitializeFlowSystem() first")
	}

	// Create a copy of the prompt without keyboard for informational use
	infoPrompt := &PromptConfig{
		Message: prompt.Message,
		Image:   prompt.Image,
		// Keyboard is intentionally omitted for informational messages
	}

	renderCtx := &RenderContext{
		ctx:          c,
		promptConfig: infoPrompt,
		stepName:     "info",
		flowName:     "system",
	}

	return c.Bot.flowManager.promptRenderer.Render(renderCtx)
}

func (c *Context) IsGroup() bool {
	return c.isGroup
}

func (c *Context) IsChannel() bool {
	return c.isChannel
}

func (c *Context) GetMenuContext() *MenuContext {
	return &MenuContext{
		UserID:    c.UserID(),
		ChatID:    c.ChatID(),
		IsGroup:   c.isGroup,
		IsChannel: c.isChannel,
	}
}

// applyAutomaticMenuButton automatically sets the menu button for the chat based on user context
func (c *Context) applyAutomaticMenuButton() {
	if c.Bot.accessManager != nil {
		menuContext := &MenuContext{
			UserID:    c.UserID(),
			ChatID:    c.ChatID(),
			IsGroup:   c.isGroup,
			IsChannel: c.isChannel,
		}

		if menuButton := c.Bot.accessManager.GetMenuButton(menuContext); menuButton != nil {
			// Set menu button for this specific chat
			log.Printf("Setting menu button for chat %d: %+v", c.ChatID(), menuButton)
			if err := c.Bot.SetMenuButton(c.ChatID(), menuButton); err != nil {
				// Log error but don't fail the message send
				// In production, you might want to log this more appropriately
				log.Printf("Failed to set menu button for chat %d: %v", c.ChatID(), err)
				_ = err
			}
		}
	}
}

// send is an internal helper for sending messages with automatic UI management
func (c *Context) send(text string, keyboard ...interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)

	// Automatic menu button management
	c.applyAutomaticMenuButton()

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
		// Apply user-specific reply keyboard automatically
		if c.Bot.accessManager != nil {
			menuContext := &MenuContext{
				UserID:    c.UserID(),
				ChatID:    c.ChatID(),
				IsGroup:   c.Update.Message != nil && (c.Update.Message.Chat.IsGroup() || c.Update.Message.Chat.IsSuperGroup()),
				IsChannel: c.Update.Message != nil && c.Update.Message.Chat.IsChannel(),
			}
			if userMenu := c.Bot.accessManager.GetReplyKeyboard(menuContext); userMenu != nil {
				msg.ReplyMarkup = userMenu.ToTgbotapi()
			}
		} else if c.Bot.replyKeyboard != nil {
			msg.ReplyMarkup = c.Bot.replyKeyboard.ToTgbotapi()
		}
	}

	_, err := c.Bot.api.Send(msg)
	return err
}

// executeTemplate executes a template and returns the result with parse mode
func (c *Context) executeTemplate(templateName string, data interface{}) (string, ParseMode, error) {
	// Get template info to determine parse mode first
	templateInfo := templateRegistry[templateName]
	parseMode := ParseModeNone
	if templateInfo != nil {
		parseMode = templateInfo.ParseMode
	}

	// Create a template with the correct functions for this parse mode
	tmpl := c.Bot.templates.Lookup(templateName)
	if tmpl == nil {
		return "", parseMode, fmt.Errorf("template %s not found", templateName)
	}

	// Clone the template and add the correct functions
	clonedTmpl, err := tmpl.Clone()
	if err != nil {
		return "", parseMode, fmt.Errorf("failed to clone template %s: %w", templateName, err)
	}

	// Add parse mode specific functions
	clonedTmpl = clonedTmpl.Funcs(getTemplateFuncs(parseMode))

	var buf bytes.Buffer
	if err := clonedTmpl.Execute(&buf, data); err != nil {
		return "", parseMode, fmt.Errorf("executing template %s: %w", templateName, err)
	}

	return buf.String(), parseMode, nil
}

// sendWithParseMode sends a message with the specified parse mode and automatic UI management
func (c *Context) sendWithParseMode(text string, parseMode ParseMode, keyboard ...interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)

	// Set parse mode if specified
	if parseMode != ParseModeNone {
		msg.ParseMode = string(parseMode)
	}

	// Automatic menu button management
	c.applyAutomaticMenuButton()

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
		// Apply user-specific reply keyboard automatically
		if c.Bot.accessManager != nil {
			menuContext := &MenuContext{
				UserID:    c.UserID(),
				ChatID:    c.ChatID(),
				IsGroup:   c.Update.Message != nil && (c.Update.Message.Chat.IsGroup() || c.Update.Message.Chat.IsSuperGroup()),
				IsChannel: c.Update.Message != nil && c.Update.Message.Chat.IsChannel(),
			}
			if userMenu := c.Bot.accessManager.GetReplyKeyboard(menuContext); userMenu != nil {
				msg.ReplyMarkup = userMenu.ToTgbotapi()
			}
		} else if c.Bot.replyKeyboard != nil {
			msg.ReplyMarkup = c.Bot.replyKeyboard.ToTgbotapi()
		}
	}

	_, err := c.Bot.api.Send(msg)
	return err
}

// editOrReplyWithParseMode attempts to edit current message with parse mode or sends new one
func (c *Context) editOrReplyWithParseMode(text string, parseMode ParseMode, keyboard ...interface{}) error {
	if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
		msg := tgbotapi.NewEditMessageText(
			c.ChatID(),
			c.Update.CallbackQuery.Message.MessageID,
			text,
		)

		// Set parse mode if specified
		if parseMode != ParseModeNone {
			msg.ParseMode = string(parseMode)
		}

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
		} else {
			// Log why the edit failed (optional - could be removed for production)
			// Common reasons: identical content, message too old, rate limiting
			_ = err // Acknowledge the error but continue to fallback
		}
	}
	return c.sendWithParseMode(text, parseMode, keyboard...)
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

// answerCallbackQuery answers a callback query to dismiss the loading indicator (internal use only)
func (c *Context) answerCallbackQuery(text string) error {
	if c.Update.CallbackQuery == nil {
		return nil // Not a callback query, nothing to answer
	}

	cb := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, text)
	_, err := c.Bot.api.Request(cb)
	return err
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
