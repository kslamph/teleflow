package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context provides a rich interface for handling Telegram updates within handlers and flows.
// It encapsulates the update information, bot capabilities, and user session data,
// offering convenient methods for sending messages, managing flows, and working with templates.
// Each Context instance is specific to a single update and should not be shared between handlers.
type Context struct {
	telegramClient  TelegramClient        // Interface for sending messages to Telegram
	templateManager TemplateManager       // Manager for message templates
	flowOps         ContextFlowOperations // Interface for flow operations
	promptSender    PromptSender          // Component for sending rich prompts
	accessManager   AccessManager         // Access control manager

	update tgbotapi.Update        // The original Telegram update
	data   map[string]interface{} // Context-specific data storage

	userID    int64 // User ID extracted from the update
	chatID    int64 // Chat ID extracted from the update
	isGroup   bool  // True if the update is from a group chat
	isChannel bool  // True if the update is from a channel

	pendingReplyKeyboard *ReplyKeyboard // Reply keyboard to be attached to next message
}

// newContext creates a new Context instance for handling a Telegram update.
// This internal function initializes all context components and extracts
// user and chat information from the update.
func newContext(
	update tgbotapi.Update,
	client TelegramClient,
	tm TemplateManager,
	fo ContextFlowOperations,
	ps PromptSender,
	am AccessManager,
) *Context {
	ctx := &Context{
		telegramClient:  client,
		templateManager: tm,
		flowOps:         fo,
		promptSender:    ps,
		accessManager:   am,
		update:          update,
		data:            make(map[string]interface{}),
	}

	ctx.userID = ctx.extractUserID(update)
	ctx.chatID = ctx.extractChatID(update)
	ctx.isGroup = update.Message != nil && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup())
	ctx.isChannel = update.Message != nil && update.Message.Chat.IsChannel()

	return ctx
}

// UserID returns the Telegram user ID associated with this update.
// This ID uniquely identifies the user across all chats and is consistent
// across all interactions with the bot.
func (c *Context) UserID() int64 {
	return c.userID
}

// ChatID returns the chat ID where this update originated.
// For private chats, this is the same as the user ID.
// For groups and channels, this identifies the specific chat.
func (c *Context) ChatID() int64 {
	return c.chatID
}

// Set stores a key-value pair in the context's data storage.
// This data is specific to the current update/handler execution and
// is not persisted beyond the current request.
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get retrieves a value from the context's data storage.
// Returns the value and a boolean indicating whether the key was found.
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

// SetFlowData stores data specific to the user's current flow.
// This data persists across flow steps and can be accessed throughout
// the flow's execution. Returns an error if the user is not in a flow.
func (c *Context) SetFlowData(key string, value interface{}) error {
	if !c.isUserInFlow() {
		return fmt.Errorf("user not in a flow, cannot set flow data")
	}

	return c.flowOps.setUserFlowData(c.UserID(), key, value)
}

// GetFlowData retrieves data specific to the user's current flow.
// Returns the value and a boolean indicating whether the key was found.
// Returns false if the user is not currently in a flow.
func (c *Context) GetFlowData(key string) (interface{}, bool) {
	if !c.isUserInFlow() {
		return nil, false
	}

	return c.flowOps.getUserFlowData(c.UserID(), key)
}

// StartFlow initiates a named flow for the current user.
// The flow must be previously registered with the bot using RegisterFlow.
// Returns an error if the flow doesn't exist or cannot be started.
func (c *Context) StartFlow(flowName string) error {

	return c.flowOps.startFlow(c.UserID(), flowName, c)
}

// isUserInFlow checks if the current user is in any active flow.
// This is used internally to determine flow state.
func (c *Context) isUserInFlow() bool {
	return c.flowOps.isUserInFlow(c.UserID())
}

// CancelFlow cancels the current user's active flow.
// If the user is not in a flow, this operation has no effect.
func (c *Context) CancelFlow() {
	c.flowOps.cancelFlow(c.UserID())
}

// SendPrompt sends a rich prompt message with optional images, keyboards, and templates.
// This is the primary method for sending complex messages in flows and handlers.
//
// Example:
//
//	prompt := &teleflow.PromptConfig{
//		Message: "Choose an option:",
//		Keyboard: func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
//			return teleflow.NewPromptKeyboard().
//				ButtonCallback("Option 1", "opt1").
//				ButtonCallback("Option 2", "opt2")
//		},
//	}
//	err := ctx.SendPrompt(prompt)
func (c *Context) SendPrompt(prompt *PromptConfig) error {
	if c.promptSender == nil {
		return fmt.Errorf("PromptSender not initialized - this should not happen as initialization is automatic")
	}

	return c.promptSender.ComposeAndSend(c, &PromptConfig{
		Message:      prompt.Message,
		Image:        prompt.Image,
		TemplateData: prompt.TemplateData,
	})
}

// SendPromptText sends a simple text message without any additional formatting or features.
// This is a convenience method for sending plain text responses.
//
// Example:
//
//	err := ctx.SendPromptText("Hello, World!")
func (c *Context) SendPromptText(text string) error {
	return c.sendSimpleText(text)
}

// SendPromptWithTemplate sends a message using a named template with the provided data.
// The template must be previously registered using AddTemplate.
//
// Example:
//
//	ctx.AddTemplate("greeting", "Hello {{.name}}!", teleflow.ParseModeMarkdown)
//	err := ctx.SendPromptWithTemplate("greeting", map[string]interface{}{
//		"name": "John",
//	})
func (c *Context) SendPromptWithTemplate(templateName string, data map[string]interface{}) error {
	return c.SendPrompt(&PromptConfig{
		Message:      "template:" + templateName,
		TemplateData: data,
	})
}

// AddTemplate registers a new message template with the specified parse mode.
// Templates support Go template syntax and can include custom functions.
//
// Example:
//
//	err := ctx.AddTemplate("welcome", "Welcome {{.name}}! Your balance: {{.balance | currency}}", teleflow.ParseModeMarkdown)
func (c *Context) AddTemplate(name, templateText string, parseMode ParseMode) error {
	return c.templateManager.AddTemplate(name, templateText, parseMode)
}

// GetTemplateInfo returns information about a registered template.
// Returns nil if the template doesn't exist.
func (c *Context) GetTemplateInfo(name string) *TemplateInfo {
	return c.templateManager.GetTemplateInfo(name)
}

// ListTemplates returns a list of all registered template names.
func (c *Context) ListTemplates() []string {
	return c.templateManager.ListTemplates()
}

// HasTemplate checks if a template with the given name is registered.
func (c *Context) HasTemplate(name string) bool {
	return c.templateManager.HasTemplate(name)
}

// RenderTemplate renders a template with the provided data and returns the result.
// This is useful for testing templates or using them in complex scenarios.
func (c *Context) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	return c.templateManager.RenderTemplate(name, data)
}

// TemplateManager returns the underlying template manager for advanced operations.
func (c *Context) TemplateManager() TemplateManager {
	return c.templateManager
}

// IsGroup returns true if the current update originated from a group chat.
func (c *Context) IsGroup() bool {
	return c.isGroup
}

// IsChannel returns true if the current update originated from a channel.
func (c *Context) IsChannel() bool {
	return c.isChannel
}

// getPermissionContext creates a PermissionContext for access control decisions.
// Returns nil if no access manager is configured.
func (c *Context) getPermissionContext() *PermissionContext {
	if c.accessManager != nil {
		return &PermissionContext{
			UserID:    c.UserID(),
			ChatID:    c.ChatID(),
			IsGroup:   c.isGroup,
			IsChannel: c.isChannel,
		}

	}
	return nil
}

// extractUserID extracts the user ID from different types of Telegram updates.
// Supports both message updates and callback query updates.
func (c *Context) extractUserID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

// answerCallbackQuery responds to a callback query with optional text.
// This is required by Telegram's API when handling inline keyboard button presses.
func (c *Context) answerCallbackQuery(text string) error {
	if c.update.CallbackQuery == nil {
		return nil
	}

	cb := tgbotapi.NewCallback(c.update.CallbackQuery.ID, text)
	_, err := c.telegramClient.Request(cb)
	return err
}

// extractChatID extracts the chat ID from different types of Telegram updates.
// Supports both message updates and callback query updates.
func (c *Context) extractChatID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

// sendSimpleText sends a plain text message to the current chat.
// Automatically attaches any pending reply keyboard and disables web page previews.
func (c *Context) sendSimpleText(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	msg.DisableWebPagePreview = true

	// Attach pending reply keyboard if available
	if c.pendingReplyKeyboard != nil {
		msg.ReplyMarkup = c.pendingReplyKeyboard.ToTgbotapi()
		c.pendingReplyKeyboard = nil // Clear after use
	}

	_, err := c.telegramClient.Send(msg)
	return err
}

// SetPendingReplyKeyboard sets a reply keyboard to be attached to the next outgoing message.
// The keyboard will be automatically attached and cleared when the next message is sent.
//
// Example:
//
//	keyboard := teleflow.BuildReplyKeyboard([]string{"Yes", "No"}, 2)
//	ctx.SetPendingReplyKeyboard(keyboard)
//	ctx.SendPromptText("Do you agree?") // Keyboard will be attached to this message
func (c *Context) SetPendingReplyKeyboard(keyboard *ReplyKeyboard) {
	c.pendingReplyKeyboard = keyboard
}
