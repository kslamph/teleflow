// Package teleflow provides a comprehensive framework for building Telegram bots
// with advanced flow management, keyboard handling, and context-aware operations.
package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context provides comprehensive information and helper methods for handling
// Telegram bot interactions within the teleflow framework. It serves as the
// central point of access for bot operations, maintaining both request-scoped
// data and flow-persistent data throughout the interaction lifecycle.
//
// The Context encapsulates:
//   - Current Telegram update (messages, callback queries, etc.)
//   - Bot instance for API operations
//   - User and chat identification
//   - Request-scoped data storage (temporary, per-update)
//   - Flow-persistent data storage (maintained across flow steps)
//   - Helper methods for sending messages, templates, and flow control
//
// Context is passed to all handlers and provides two distinct data storage mechanisms:
//   - Request-scoped data: Use Set()/Get() for temporary data during update processing
//   - Flow-persistent data: Use SetFlowData()/GetFlowData() for data that persists across flow steps
//
// Example usage in command handlers:
//
//	bot.HandleCommand("/start", func(ctx *teleflow.Context, command string, args string) error {
//		// Store temporary data for this request
//		ctx.Set("start_time", time.Now())
//
//		// Start a conversational flow
//		return ctx.StartFlow("registration")
//	})
//
// Example usage in flow definitions:
//
//	flow := teleflow.NewFlow("registration").
//		Step("name").
//		Prompt("What's your name?", nil, nil).
//		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//			if buttonClick != nil {
//				return teleflow.NextStep() // Handle button clicks
//			}
//			if len(input) < 2 {
//				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{
//					Message: "Name must be at least 2 characters"
//				})
//			}
//			// Store data that persists across flow steps
//			ctx.SetFlowData("user_name", input)
//			return teleflow.NextStep()
//		})
//
// Example message sending usage:
//
//	// Send informational message with template rendering
//	ctx.SendPromptWithTemplate("welcome", map[string]interface{}{
//		"UserName": "John",
//	})
//
//	// Send simple text message
//	ctx.SendPromptText("Choose an option!")
//
//	// Send complex message with image and template
//	ctx.SendPrompt(&teleflow.PromptConfig{
//		Message: "template:user_profile",
//		Image: "path/to/image.jpg",
//		TemplateData: map[string]interface{}{"Name": "John"},
//	})

// Context provides information and helpers for the current Telegram interaction.
// It maintains both temporary request-scoped data and flow-persistent data,
// along with user/chat identification and bot API access.
type Context struct {
	bot    *Bot                   // Bot instance for API operations
	update tgbotapi.Update        // Current Telegram update being processed
	data   map[string]interface{} // Request-scoped data storage (temporary, per-update lifecycle)

	// User and chat identification extracted from the update
	userID    int64 // ID of the user who initiated the update
	chatID    int64 // ID of the chat where the update originated
	isGroup   bool  // True if the chat is a group or supergroup
	isChannel bool  // True if the chat is a channel
}

// newContext creates a new Context instance for the given bot and Telegram update.
// It automatically extracts user ID, chat ID, and chat type information from the update.
// The context is initialized with an empty request-scoped data map.
//
// Parameters:
//   - bot: The Bot instance that will handle operations
//   - update: The Telegram update to process
//
// Returns a fully initialized Context ready for use in handlers and flows.
func newContext(bot *Bot, update tgbotapi.Update) *Context {
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

// UserID returns the ID of the user who initiated the current update.
// This ID is extracted from the update's message or callback query and
// is used for flow management and user-specific operations.
func (c *Context) UserID() int64 {
	return c.userID
}

// ChatID returns the ID of the chat where the current update originated.
// This can be a private chat, group, supergroup, or channel ID.
func (c *Context) ChatID() int64 {
	return c.chatID
}

// Set stores a value in the context's request-scoped data map.
// This data is temporary and only exists for the duration of the current
// update processing lifecycle, including middleware execution.
// For data that needs to persist across flow steps, use SetFlowData instead.
//
// Parameters:
//   - key: The key to store the value under
//   - value: The value to store (can be any type)
//
// Example:
//
//	ctx.Set("processing_start", time.Now())
//	ctx.Set("user_input_valid", true)
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get retrieves a value from the context's request-scoped data map.
// This only accesses temporary data stored during the current update processing.
// For persistent flow data, use GetFlowData instead.
//
// Parameters:
//   - key: The key to retrieve the value for
//
// Returns:
//   - interface{}: The stored value, or nil if not found
//   - bool: True if the key was found, false otherwise
//
// Example:
//
//	if startTime, found := ctx.Get("processing_start"); found {
//	    // Use startTime
//	}
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

// SetFlowData sets persistent data for the user's active flow.
// This data persists across flow steps and can be accessed throughout
// the entire flow lifecycle. The user must be in an active flow.
//
// Parameters:
//   - key: The key to store the value under
//   - value: The value to store (can be any type)
//
// Returns an error if the user is not currently in a flow.
//
// Example:
//
//	// Store user's name to use in later flow steps
//	err := ctx.SetFlowData("user_name", "John Doe")
//	if err != nil {
//	    // Handle error - user not in flow
//	}
func (c *Context) SetFlowData(key string, value interface{}) error {
	if !c.IsUserInFlow() {
		return fmt.Errorf("user not in a flow, cannot set flow data")
	}
	// Delegate to flowManager to update UserFlowState.Data
	return c.bot.flowManager.setUserFlowData(c.UserID(), key, value)
}

// GetFlowData retrieves persistent data from the user's active flow state.
// This accesses data that was stored using SetFlowData and persists across
// flow steps. Returns nil and false if the user is not in a flow.
//
// Parameters:
//   - key: The key to retrieve the value for
//
// Returns:
//   - interface{}: The stored value, or nil if not found or user not in flow
//   - bool: True if the key was found and user is in flow, false otherwise
//
// Example:
//
//	if name, found := ctx.GetFlowData("user_name"); found {
//	    // Use the stored name from previous flow steps
//	    welcomeMsg := fmt.Sprintf("Hello, %s!", name)
//	}
func (c *Context) GetFlowData(key string) (interface{}, bool) {
	if !c.IsUserInFlow() {
		return nil, false
	}
	// Delegate to flowManager to get from UserFlowState.Data
	return c.bot.flowManager.getUserFlowData(c.UserID(), key)
}

// StartFlow initiates a new conversational flow for the current user.
// If the user is already in a flow, it will be cancelled and replaced
// with the new flow. The flow must be registered with the bot beforehand.
//
// Parameters:
//   - flowName: The name of the flow to start (must be registered)
//
// Returns an error if the flow is not found or cannot be started.
//
// Example:
//
//	// Start a registration flow
//	err := ctx.StartFlow("user_registration")
//	if err != nil {
//	    ctx.SendPromptText("Sorry, registration is not available right now.")
//	}
func (c *Context) StartFlow(flowName string) error {
	// StartFlow now takes a Context parameter instead of initialData
	return c.bot.flowManager.startFlow(c.UserID(), flowName, c)
}

// IsUserInFlow checks if the current user is actively participating in a flow.
// This is useful for determining whether flow-specific operations can be performed.
//
// Returns true if the user is in an active flow, false otherwise.
//
// Example:
//
//	if ctx.IsUserInFlow() {
//	    // User is in a flow, can access flow data
//	    if name, found := ctx.GetFlowData("name"); found {
//	        // Process flow data
//	    }
//	}
func (c *Context) IsUserInFlow() bool {
	return c.bot.flowManager.isUserInFlow(c.UserID())
}

// CancelFlow cancels the current flow for the user, if any.
// This operation is idempotent - calling it multiple times or when
// the user is not in a flow will not cause errors.
//
// After cancellation, the user's flow state and associated data are cleared.
//
// Example:
//
//	// Cancel flow on user command
//	ctx.CancelFlow()
//	ctx.SendPromptText("Flow cancelled. You can start over anytime.")
func (c *Context) CancelFlow() {
	c.bot.flowManager.cancelFlow(c.UserID())
}

// SendPrompt renders and sends a PromptConfig message for informational purposes.
// This is the primary method for sending messages outside of flows, using the same
// template rendering and image handling system as flow prompts but omits any
// keyboard interaction, making it ideal for status messages, notifications, or
// informational content in command handlers and OnComplete callbacks.
//
// Parameters:
//   - prompt: The PromptConfig containing message, optional image, and template data
//
// Returns an error if the PromptComposer is not initialized or sending fails.
//
// Example:
//
//	// Send welcome message with template
//	ctx.SendPrompt(&teleflow.PromptConfig{
//	    Message: "template:welcome",
//	    TemplateData: map[string]interface{}{
//	        "UserName": ctx.GetFlowData("name"),
//	    },
//	})
//
//	// Send message with image
//	ctx.SendPrompt(&teleflow.PromptConfig{
//	    Message: "Check out this image!",
//	    Image: "path/to/image.jpg",
//	})
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

	return c.bot.promptComposer.composeAndSend(c, infoPrompt)
}

// SendPromptText is a convenience method for sending simple text messages.
// This is the recommended approach for basic text replies in command handlers
// and other scenarios where you need to send plain text without templates or images.
//
// Parameters:
//   - text: The text message to send
//
// Returns an error if sending fails.
//
// Example:
//
//	// Send simple reply in command handler
//	ctx.SendPromptText("Hello! Welcome to the bot.")
func (c *Context) SendPromptText(text string) error {
	return c.sendSimpleText(text)
}

// SendPromptWithTemplate is a convenience method for sending informational messages
// using the template system. It combines template rendering with the SendPrompt functionality,
// making it easy to send formatted informational messages without keyboards.
//
// Parameters:
//   - templateName: Name of the template to render (without "template:" prefix)
//   - data: Template data for variable substitution
//
// Returns an error if template rendering or message sending fails.
//
// Example:
//
//	// Send status update using template
//	ctx.SendPromptWithTemplate("process_status", map[string]interface{}{
//	    "Status": "completed",
//	    "Progress": 100,
//	    "Duration": "2 minutes",
//	})
func (c *Context) SendPromptWithTemplate(templateName string, data map[string]interface{}) error {
	return c.SendPrompt(&PromptConfig{
		Message:      "template:" + templateName,
		TemplateData: data,
	})
}

// IsGroup returns true if the current chat is a group or supergroup.
// This is useful for applying group-specific logic or permissions.
func (c *Context) IsGroup() bool {
	return c.isGroup
}

// IsChannel returns true if the current chat is a channel.
// This is useful for handling channel-specific operations.
func (c *Context) IsChannel() bool {
	return c.isChannel
}

// getPermissionContext creates a PermissionContext for the current user and chat.
// This is used internally by the AccessManager to determine appropriate permissions
// and reply keyboards. Returns nil if no AccessManager is configured.
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

// extractUserID extracts the user ID from a Telegram update.
// It handles both message updates and callback query updates.
// Returns 0 if no user ID can be extracted.
func (c *Context) extractUserID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

// answerCallbackQuery answers a callback query to dismiss the loading indicator.
// This is used internally and only applies when the current update is a callback query.
// If the update is not a callback query, this method does nothing.
//
// Parameters:
//   - text: Optional text to show to the user (can be empty)
//
// Returns an error if answering the callback query fails.
func (c *Context) answerCallbackQuery(text string) error {
	if c.update.CallbackQuery == nil {
		return nil // Not a callback query, nothing to answer
	}

	cb := tgbotapi.NewCallback(c.update.CallbackQuery.ID, text)
	_, err := c.bot.api.Request(cb)
	return err
}

// extractChatID extracts the chat ID from a Telegram update.
// It handles both message updates and callback query updates.
// Returns 0 if no chat ID can be extracted.
func (c *Context) extractChatID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

// ReplyKeyboardOption defines a functional option for configuring reply keyboards.
// These options modify the behavior and appearance of reply keyboards sent via SendReplyKeyboard.
type ReplyKeyboardOption func(*tgbotapi.ReplyKeyboardMarkup)

// WithResize returns an option that configures the reply keyboard to automatically resize.
// When enabled, the keyboard will be displayed in a more compact way.
func WithResize() ReplyKeyboardOption {
	return func(kb *tgbotapi.ReplyKeyboardMarkup) {
		kb.ResizeKeyboard = true
	}
}

// WithOneTime returns an option that configures the reply keyboard to be one-time use.
// The keyboard will be hidden after the user presses any button.
func WithOneTime() ReplyKeyboardOption {
	return func(kb *tgbotapi.ReplyKeyboardMarkup) {
		kb.OneTimeKeyboard = true
	}
}

// WithPlaceholder returns an option that sets the input field placeholder text.
// This text appears in the message input field when the keyboard is displayed.
//
// Parameters:
//   - text: The placeholder text to display
func WithPlaceholder(text string) ReplyKeyboardOption {
	return func(kb *tgbotapi.ReplyKeyboardMarkup) {
		kb.InputFieldPlaceholder = text
	}
}

func (c *Context) sendSimpleText(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	msg.DisableWebPagePreview = true
	_, err := c.bot.api.Send(msg)
	return err
}
