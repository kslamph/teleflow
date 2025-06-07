package teleflow

import (
	"fmt"
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
	bot    *Bot
	update tgbotapi.Update
	data   map[string]interface{}

	// User context
	userID    int64
	chatID    int64
	isGroup   bool
	isChannel bool
}

// NewContext creates a new Context
func NewContext(bot *Bot, update tgbotapi.Update) *Context {
	ctx := &Context{
		bot:    bot,
		update: update,
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

// ReplyWithParseMode sends a text message with specific parse mode and keyboard
func (c *Context) ReplyWithParseMode(text string, parseMode ParseMode, keyboard ...interface{}) error {
	return c.sendWithParseMode(text, parseMode, keyboard...)
}

// StartFlow initiates a new flow for the user
func (c *Context) StartFlow(flowName string) error {
	// StartFlow now takes a Context parameter instead of initialData
	return c.bot.flowManager.startFlow(c.UserID(), flowName, c)
}

// IsUserInFlow checks if the user is currently in a flow
func (c *Context) IsUserInFlow() bool {
	return c.bot.flowManager.isUserInFlow(c.UserID())
}

// CancelFlow cancels the current flow for the user
// This is an idempotent operation
func (c *Context) CancelFlow() {
	c.bot.flowManager.cancelFlow(c.UserID())
}

// SendPrompt renders and sends a PromptConfig message without keyboard interaction.
// This is useful for informational messages that use the same rendering system
// as flow prompts but don't require user interaction.
func (c *Context) SendPrompt(prompt *PromptConfig) error {
	if c.bot.promptComposer == nil {
		return fmt.Errorf("PromptComposer not initialized - this should not happen as initialization is automatic")
	}

	// Create a copy of the prompt without keyboard for informational use
	infoPrompt := &PromptConfig{
		Message:      prompt.Message,
		Image:        prompt.Image,
		TemplateData: prompt.TemplateData,
		// Keyboard is intentionally omitted for informational messages
	}

	return c.bot.promptComposer.ComposeAndSend(c, infoPrompt)
}

// ReplyTemplate renders and sends a template with data and optional keyboard.
// This is a convenience method that uses the template system with "template:" + templateName format.
func (c *Context) ReplyTemplate(templateName string, data map[string]interface{}, keyboard ...interface{}) error {
	prompt := &PromptConfig{
		Message:      "template:" + templateName,
		TemplateData: data,
	}

	// If keyboard is provided, we need to handle it differently since PromptConfig.Keyboard
	// expects a KeyboardFunc, but traditional Reply methods accept various keyboard types
	if len(keyboard) > 0 && keyboard[0] != nil {
		// For templates with keyboards, we need to render manually and use Reply
		// since PromptComposer only handles KeyboardFunc in PromptConfig
		if c.bot.promptComposer == nil {
			return fmt.Errorf("PromptComposer not initialized")
		}

		// Render the template message using PromptComposer's messageRenderer
		message, parseMode, err := c.bot.promptComposer.messageRenderer.renderMessage(prompt, c)
		if err != nil {
			return fmt.Errorf("failed to render template '%s': %w", templateName, err)
		}

		// Send with the keyboard using the appropriate method
		if parseMode != ParseModeNone {
			return c.ReplyWithParseMode(message, parseMode, keyboard[0])
		}
		return c.Reply(message, keyboard[0])
	}

	// No keyboard, use PromptComposer
	if c.bot.promptComposer == nil {
		return fmt.Errorf("PromptComposer not initialized")
	}
	return c.bot.promptComposer.ComposeAndSend(c, prompt)
}

// SendPromptWithTemplate is a convenience method for template-based prompts.
// It renders the specified template with the given data and sends it as a prompt.
func (c *Context) SendPromptWithTemplate(templateName string, data map[string]interface{}) error {
	return c.SendPrompt(&PromptConfig{
		Message:      "template:" + templateName,
		TemplateData: data,
	})
}

func (c *Context) IsGroup() bool {
	return c.isGroup
}

func (c *Context) IsChannel() bool {
	return c.isChannel
}

func (c *Context) getPermissionContext() *PermissionContext {
	if c.bot.accessManager != nil {
		return &PermissionContext{
			UserID:    c.UserID(),
			ChatID:    c.ChatID(),
			IsGroup:   c.isGroup,
			IsChannel: c.isChannel,
		}

	}
	return nil // No access manager, no menu button
}

// applyAutomaticMenuButton automatically sets the menu button for the chat based on user context
// Note: Bot commands are now set via SetBotCommands method, not through AccessManager
func (c *Context) applyAutomaticMenuButton() {
	// AccessManager no longer provides menu buttons for bot commands
	// This method is kept for potential future use with web_app menu buttons
	// Bot commands should be set directly via Bot.SetBotCommands()
}

// send is an internal helper for sending messages with automatic UI management
func (c *Context) send(text string, keyboard ...interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)

	// Check if a parse mode was set during rendering
	if parseMode, exists := c.Get("__render_parse_mode"); exists {
		if pm, ok := parseMode.(ParseMode); ok && pm != ParseModeNone {
			msg.ParseMode = string(pm)
		}
		// Clean up the temporary parse mode
		delete(c.data, "__render_parse_mode")
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
			// Check if the inline keyboard is empty (has no buttons)
			if len(kb.InlineKeyboard) > 0 {
				msg.ReplyMarkup = kb
			}
			// If empty, don't set ReplyMarkup (leave it nil)
		}
	} else {
		// Apply user-specific reply keyboard automatically only if no explicit keyboard provided
		if c.bot.accessManager != nil {
			permissionContext := c.getPermissionContext()
			if userMenu := c.bot.accessManager.GetReplyKeyboard(permissionContext); userMenu != nil {
				msg.ReplyMarkup = userMenu.ToTgbotapi()
			}
		} else if c.bot.replyKeyboard != nil {
			msg.ReplyMarkup = c.bot.replyKeyboard.ToTgbotapi()
		}
	}

	_, err := c.bot.api.Send(msg)
	return err
}

// sendWithParseMode is an internal helper for sending messages with parse mode
func (c *Context) sendWithParseMode(text string, parseMode ParseMode, keyboard ...interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)

	// Set parse mode if specified
	if parseMode != ParseModeNone && parseMode != "" {
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
			// Check if the inline keyboard is empty (has no buttons)
			if len(kb.InlineKeyboard) > 0 {
				msg.ReplyMarkup = kb
			}
			// If empty, don't set ReplyMarkup (leave it nil)
		}
	} else {
		// Apply user-specific reply keyboard automatically only if no explicit keyboard provided
		if c.bot.accessManager != nil {
			permissionContext := c.getPermissionContext()
			if userMenu := c.bot.accessManager.GetReplyKeyboard(permissionContext); userMenu != nil {
				msg.ReplyMarkup = userMenu.ToTgbotapi()
			}
		} else if c.bot.replyKeyboard != nil {
			msg.ReplyMarkup = c.bot.replyKeyboard.ToTgbotapi()
		}
	}

	_, err := c.bot.api.Send(msg)
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
func (c *Context) SendPhoto(image *processedImage, caption string, keyboard ...interface{}) error {
	var photoConfig tgbotapi.PhotoConfig

	// Handle different image types
	if image.data != nil {
		// Send image data directly (base64 or file data)
		photoConfig = tgbotapi.NewPhoto(c.ChatID(), tgbotapi.FileBytes{
			Name:  "image",
			Bytes: image.data,
		})
	} else if image.filePath != "" {
		// Handle URL or file path
		if c.isURL(image.filePath) {
			// Send URL directly
			photoConfig = tgbotapi.NewPhoto(c.ChatID(), tgbotapi.FileURL(image.filePath))
		} else {
			// Send file path
			photoConfig = tgbotapi.NewPhoto(c.ChatID(), tgbotapi.FilePath(image.filePath))
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
			// Check if the inline keyboard is empty (has no buttons)
			if len(kb.InlineKeyboard) > 0 {
				photoConfig.ReplyMarkup = kb
			}
			// If empty, don't set ReplyMarkup (leave it nil)
		}
	}

	// Apply automatic menu button management
	c.applyAutomaticMenuButton()

	_, err := c.bot.api.Send(photoConfig)
	return err
}

// isURL checks if a string is likely a URL (method on Context)
func (c *Context) isURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// answerCallbackQuery answers a callback query to dismiss the loading indicator (internal use only)
func (c *Context) answerCallbackQuery(text string) error {
	if c.update.CallbackQuery == nil {
		return nil // Not a callback query, nothing to answer
	}

	cb := tgbotapi.NewCallback(c.update.CallbackQuery.ID, text)
	_, err := c.bot.api.Request(cb)
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

// ReplyKeyboardOption defines functional options for SendReplyKeyboard.
type ReplyKeyboardOption func(*tgbotapi.ReplyKeyboardMarkup)

// WithResize configures the reply keyboard to resize.
func WithResize() ReplyKeyboardOption {
	return func(kb *tgbotapi.ReplyKeyboardMarkup) {
		kb.ResizeKeyboard = true
	}
}

// WithOneTime configures the reply keyboard to be one-time.
func WithOneTime() ReplyKeyboardOption {
	return func(kb *tgbotapi.ReplyKeyboardMarkup) {
		kb.OneTimeKeyboard = true
	}
}

// WithPlaceholder sets the input field placeholder for the reply keyboard.
func WithPlaceholder(text string) ReplyKeyboardOption {
	return func(kb *tgbotapi.ReplyKeyboardMarkup) {
		kb.InputFieldPlaceholder = text
	}
}

// SendReplyKeyboard sends a new reply keyboard, replacing any existing one for the user.
// It uses the BuildReplyKeyboard logic internally.
// Example: ctx.SendReplyKeyboard([]string{"Button 1", "Button 2"}, 2, WithResize(), WithOneTime())
func (c *Context) SendReplyKeyboard(buttons []string, buttonsPerRow int, options ...ReplyKeyboardOption) error {
	if c.bot == nil || c.bot.api == nil {
		return fmt.Errorf("bot or bot API not available in context for SendReplyKeyboard")
	}

	tempReplyKeyboard := BuildReplyKeyboard(buttons, buttonsPerRow)
	tgAPIReplyKeyboard := tempReplyKeyboard.ToTgbotapi()

	for _, opt := range options {
		opt(&tgAPIReplyKeyboard)
	}

	msg := tgbotapi.NewMessage(c.ChatID(), "\u200B") // Use an invisible char
	msg.ReplyMarkup = tgAPIReplyKeyboard
	_, err := c.bot.api.Send(msg)
	return err
}
