package teleflow

import (
	"fmt"
	"log"
	"strings"

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
// Example usage in handlers:
//
//	bot.HandleCommand("/start", func(ctx *teleflow.Context, command string, args string) error {
//		// Start a flow using the new Step-Prompt-Process API
//		return ctx.StartFlow("registration")
//	})
//
//	// Modern flow definition with Step-Prompt-Process API:
//	flow := teleflow.NewFlow("registration").
//		Step("name").
//		Prompt("What's your name?", nil, nil).
//		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//			// Handle both text input and button clicks in one unified function
//			if buttonClick != nil {
//				// Button was clicked
//				return teleflow.NextStep()
//			}
//			// Text input validation
//			if len(input) < 2 {
//				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{Message: "Name must be at least 2 characters"})
//			}
//			ctx.Set("name", input)
//			return teleflow.NextStep()
//		})
//
//	// Use SendPrompt for informational messages with PromptConfig rendering:
//	ctx.SendPrompt(&teleflow.PromptConfig{
//		Message: "Welcome to our service!",
//		Image:   nil, // Optional image
//	})
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

// StartFlow initiates a new flow for the user
func (c *Context) StartFlow(flowName string) error {
	// StartFlow now takes a Context parameter instead of initialData
	return c.Bot.flowManager.startFlow(c.UserID(), flowName, c)
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
		return fmt.Errorf("PromptRenderer not initialized - this should not happen as initialization is automatic")
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

// SendPhoto sends a photo message with optional caption and keyboard
func (c *Context) SendPhoto(image *ProcessedImage, caption string, keyboard ...interface{}) error {
	var photoConfig tgbotapi.PhotoConfig

	// Handle different image types
	if image.Data != nil {
		// Send image data directly (base64 or file data)
		photoConfig = tgbotapi.NewPhoto(c.ChatID(), tgbotapi.FileBytes{
			Name:  "image",
			Bytes: image.Data,
		})
	} else if image.FilePath != "" {
		// Handle URL or file path
		if c.isURL(image.FilePath) {
			// Send URL directly
			photoConfig = tgbotapi.NewPhoto(c.ChatID(), tgbotapi.FileURL(image.FilePath))
		} else {
			// Send file path
			photoConfig = tgbotapi.NewPhoto(c.ChatID(), tgbotapi.FilePath(image.FilePath))
		}
	} else {
		return fmt.Errorf("no valid image data or path provided")
	}

	// Set caption
	if caption != "" {
		photoConfig.Caption = caption
	}

	// Apply keyboard markup
	if len(keyboard) > 0 && keyboard[0] != nil {
		switch kb := keyboard[0].(type) {
		case *ReplyKeyboard:
			photoConfig.ReplyMarkup = kb.ToTgbotapi()
		case *InlineKeyboard:
			photoConfig.ReplyMarkup = kb.ToTgbotapi()
		case tgbotapi.ReplyKeyboardRemove:
			photoConfig.ReplyMarkup = kb
		case tgbotapi.ReplyKeyboardMarkup:
			photoConfig.ReplyMarkup = kb
		case tgbotapi.InlineKeyboardMarkup:
			photoConfig.ReplyMarkup = kb
		}
	}

	// Apply automatic menu button management
	c.applyAutomaticMenuButton()

	_, err := c.Bot.api.Send(photoConfig)
	return err
}

// isURL checks if a string is likely a URL (method on Context)
func (c *Context) isURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
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
