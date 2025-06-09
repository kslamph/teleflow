package teleflow

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerFunc represents a generic handler function for processing Telegram updates.
// It receives a Context containing update information and bot capabilities,
// and returns an error if processing fails.
type HandlerFunc func(ctx *Context) error

// CommandHandlerFunc represents a handler function specifically for Telegram commands.
// It receives the command name and any arguments passed with the command.
// Commands are messages that start with "/" (e.g., "/start", "/help").
type CommandHandlerFunc func(ctx *Context, command string, args string) error

// TextHandlerFunc represents a handler function for specific text messages.
// It's triggered when the bot receives a message that exactly matches the configured text.
type TextHandlerFunc func(ctx *Context, text string) error

// DefaultHandlerFunc represents a fallback handler for unmatched messages.
// It's called when no specific command or text handler matches the incoming message.
type DefaultHandlerFunc func(ctx *Context, text string) error

// BotOption represents a configuration option for customizing bot behavior.
// Options are applied during bot creation to configure features like flow management
// and access control.
type BotOption func(*Bot)

// PermissionContext contains information needed for access control decisions.
// It provides context about the user, chat, and operation being performed,
// allowing AccessManager implementations to make informed authorization decisions.
type PermissionContext struct {
	UserID    int64            // Telegram user ID making the request
	ChatID    int64            // Chat ID where the request originated
	Command   string           // Command being executed (if applicable)
	Arguments []string         // Command arguments (if applicable)
	IsGroup   bool             // True if the request is from a group chat
	IsChannel bool             // True if the request is from a channel
	MessageID int              // Message ID of the request
	Update    *tgbotapi.Update // Full Telegram update object for advanced processing
}

// AccessManager defines the interface for controlling user access to bot features.
// Implementations can restrict access based on user permissions, chat types,
// or any other criteria relevant to the bot's security requirements.
type AccessManager interface {
	// CheckPermission validates whether the given context has permission to proceed.
	// Returns nil if access is granted, or an error describing why access was denied.
	CheckPermission(ctx *PermissionContext) error

	// GetReplyKeyboard returns a custom reply keyboard for the given context.
	// This allows dynamic keyboard generation based on user permissions.
	// Returns nil if no custom keyboard should be used.
	GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard
}

// Bot represents the main bot instance that handles Telegram updates and manages conversations.
// It provides methods for registering handlers, managing flows, and configuring bot behavior.
// The Bot is the central component that coordinates all other framework features.
type Bot struct {
	api  TelegramClient // Interface for communicating with Telegram API
	self tgbotapi.User  // Bot's own user information from Telegram

	handlers           map[string]HandlerFunc // Registered command handlers
	textHandlers       map[string]HandlerFunc // Registered text message handlers
	defaultTextHandler HandlerFunc            // Fallback handler for unmatched messages

	flowManager           *flowManager          // Manages multi-step conversation flows
	promptKeyboardHandler PromptKeyboardActions // Handles inline keyboard interactions
	promptComposer        *PromptComposer       // Composes and sends rich messages
	templateManager       TemplateManager       // Manages message templates

	middleware []MiddlewareFunc // Chain of middleware functions

	accessManager AccessManager // Controls user access to bot features
	flowConfig    FlowConfig    // Configuration for flow behavior
}

// newBotInternal creates a new Bot instance with the provided client and configuration.
// This internal function is used by NewBot and for testing with mock clients.
// It initializes all bot components and applies the provided options.
func newBotInternal(client TelegramClient, botUser tgbotapi.User, options ...BotOption) (*Bot, error) {
	b := &Bot{
		api:                   client,
		self:                  botUser,
		handlers:              make(map[string]HandlerFunc),
		textHandlers:          make(map[string]HandlerFunc),
		promptKeyboardHandler: newPromptKeyboardHandler(),
		templateManager:       GetDefaultTemplateManager(),
		middleware:            make([]MiddlewareFunc, 0),
		flowConfig: FlowConfig{
			ExitCommands:        []string{"/cancel"},
			ExitMessage:         "🚫 Operation cancelled.",
			AllowGlobalCommands: false,
			HelpCommands:        []string{"/help"},
			OnProcessAction:     ProcessKeepMessage,
		},
	}

	msgHandler := newMessageHandler(b.templateManager)
	imageHandler := newImageHandler()
	b.promptComposer = newPromptComposer(b.api, msgHandler, imageHandler, b.promptKeyboardHandler.(*PromptKeyboardHandler))

	for _, opt := range options {
		opt(b)
	}

	// Initialize flowManager with its new dependencies
	b.flowManager = newFlowManager(&b.flowConfig, b.promptComposer, b.promptKeyboardHandler, b)
	return b, nil
}

// NewBot creates a new Bot instance using the provided Telegram bot token.
// It establishes a connection to the Telegram Bot API and retrieves bot information.
// Additional configuration can be applied using BotOption functions.
//
// Example:
//
//	bot, err := teleflow.NewBot("YOUR_BOT_TOKEN",
//		teleflow.WithFlowConfig(config),
//		teleflow.WithAccessManager(accessManager),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
func NewBot(token string, options ...BotOption) (*Bot, error) {
	realAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	botUser, err := realAPI.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	return newBotInternal(realAPI, botUser, options...)
}

// WithFlowConfig returns a BotOption that configures flow management behavior.
// This option allows customization of exit commands, help commands, and flow processing options.
//
// Example:
//
//	config := teleflow.FlowConfig{
//		ExitCommands:        []string{"/cancel", "/exit"},
//		ExitMessage:         "Operation cancelled.",
//		AllowGlobalCommands: true,
//	}
//	bot, err := teleflow.NewBot(token, teleflow.WithFlowConfig(config))
func WithFlowConfig(config FlowConfig) BotOption {
	return func(b *Bot) {
		b.flowConfig = config
	}
}

// WithAccessManager returns a BotOption that configures access control for the bot.
// It automatically adds the AuthMiddleware to enforce permission checks.
// The AccessManager will be consulted for all incoming requests to determine access rights.
//
// Example:
//
//	accessManager := &MyAccessManager{}
//	bot, err := teleflow.NewBot(token, teleflow.WithAccessManager(accessManager))
func WithAccessManager(accessManager AccessManager) BotOption {
	return func(b *Bot) {
		b.accessManager = accessManager
		b.UseMiddleware(AuthMiddleware(accessManager))
	}
}

// UseMiddleware adds a middleware function to the bot's middleware chain.
// Middleware functions are executed in the order they are added, wrapping the final handler.
// This allows for cross-cutting concerns like logging, authentication, and error handling.
//
// Example:
//
//	bot.UseMiddleware(teleflow.LoggingMiddleware())
//	bot.UseMiddleware(teleflow.RecoveryMiddleware())
//	bot.UseMiddleware(teleflow.RateLimitMiddleware(10))
func (b *Bot) UseMiddleware(m MiddlewareFunc) {
	b.middleware = append(b.middleware, m)
}

// HandleCommand registers a handler for a specific Telegram command.
// Commands are messages that start with "/" (e.g., "/start", "/help").
// The handler receives the command name and any arguments that follow it.
//
// Example:
//
//	bot.HandleCommand("start", func(ctx *teleflow.Context, command, args string) error {
//		return ctx.SendPromptText("Welcome! Arguments: " + args)
//	})
func (b *Bot) HandleCommand(commandName string, handler CommandHandlerFunc) {

	wrappedHandler := func(ctx *Context) error {

		command := commandName
		args := ""
		if ctx.update.Message != nil && len(ctx.update.Message.Text) > len(command)+1 {
			args = ctx.update.Message.Text[len(command)+1:]
		}
		return handler(ctx, command, args)
	}
	b.handlers[commandName] = b.applyMiddleware(wrappedHandler)
}

// HandleText registers a handler for exact text message matches.
// The handler is triggered only when the received message text exactly matches the specified text.
// This is useful for handling specific keywords or phrases.
//
// Example:
//
//	bot.HandleText("Hello", func(ctx *teleflow.Context, text string) error {
//		return ctx.SendPromptText("Hello there!")
//	})
func (b *Bot) HandleText(textToMatch string, handler TextHandlerFunc) {

	wrappedHandler := func(ctx *Context) error {
		return handler(ctx, textToMatch)
	}
	b.textHandlers[textToMatch] = b.applyMiddleware(wrappedHandler)
}

// DefaultHandler registers a fallback handler for unmatched messages.
// This handler is called when no specific command or text handler matches the incoming message.
// Only one default handler can be registered; subsequent calls will replace the previous handler.
//
// Example:
//
//	bot.DefaultHandler(func(ctx *teleflow.Context, text string) error {
//		return ctx.SendPromptText("I don't understand: " + text)
//	})
func (b *Bot) DefaultHandler(handler DefaultHandlerFunc) {

	wrappedHandler := func(ctx *Context) error {
		var text string
		if ctx.update.Message != nil {
			text = ctx.update.Message.Text
		}
		return handler(ctx, text)
	}
	b.defaultTextHandler = b.applyMiddleware(wrappedHandler)
}

// RegisterFlow adds a conversation flow to the bot.
// Flows enable multi-step conversations with state management and complex user interactions.
// Once registered, flows can be started by calling ctx.StartFlow(flowName).
//
// Example:
//
//	flow := teleflow.NewFlow("registration").
//		Step("ask_name").Prompt("What's your name?").Process(...).
//		Step("ask_age").Prompt("How old are you?").Process(...).
//		Build()
//	bot.RegisterFlow(flow)
func (b *Bot) RegisterFlow(flow *Flow) {
	b.flowManager.registerFlow(flow)
}

// GetPromptKeyboardHandler returns the bot's keyboard handler for advanced keyboard management.
// This is typically used internally or for advanced use cases where direct keyboard manipulation is needed.
func (b *Bot) GetPromptKeyboardHandler() PromptKeyboardActions {
	return b.promptKeyboardHandler
}

// applyMiddleware applies the middleware chain to a handler function.
// Middleware is applied in reverse order (LIFO), so the last added middleware
// runs first, allowing for proper request/response wrapping.
func (b *Bot) applyMiddleware(handler HandlerFunc) HandlerFunc {
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}
	return handler
}

// processUpdate handles an incoming Telegram update by routing it to the appropriate handler.
// It manages flow state, applies global exit commands, and provides fallback error handling.
// This method is called concurrently for each update, ensuring responsive bot behavior.
func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := newContext(update, b.api, b.templateManager, b.flowManager, b.promptComposer, b.accessManager)
	var err error

	if b.flowManager.isUserInFlow(ctx.UserID()) {
		if ctx.update.Message != nil && b.isGlobalExitCommand(ctx.update.Message.Text) {
			b.flowManager.cancelFlow(ctx.UserID())
			err = ctx.sendSimpleText(b.flowConfig.ExitMessage)
			if err != nil {
				log.Printf("Error sending flow exit message: %v", err)
			}
			return
		} else if b.flowConfig.AllowGlobalCommands && ctx.update.Message != nil && ctx.update.Message.IsCommand() {
			commandName := ctx.update.Message.Command()
			if cmdHandler := b.resolveGlobalCommandHandler(commandName); cmdHandler != nil {

				err = cmdHandler(ctx)
				if err != nil {
					log.Printf("Global command handler error for UserID %d, command '%s': %v", ctx.UserID(), commandName, err)

				}
				return
			}
		}
	}

	if handledByFlow, flowErr := b.flowManager.HandleUpdate(ctx); handledByFlow {
		if flowErr != nil {
			log.Printf("Flow handler error for UserID %d: %v", ctx.UserID(), flowErr)

		}
		return
	}

	if update.Message != nil {
		if update.Message.IsCommand() {
			commandName := update.Message.Command()
			if cmdHandler, ok := b.handlers[commandName]; ok {
				err = cmdHandler(ctx)
			} else {

				if b.defaultTextHandler != nil {
					err = b.defaultTextHandler(ctx)
				}
			}
		} else {
			text := update.Message.Text
			if textHandler, ok := b.textHandlers[text]; ok {
				err = textHandler(ctx)
			} else if b.defaultTextHandler != nil {
				err = b.defaultTextHandler(ctx)
			}

		}
	} else if update.CallbackQuery != nil {

		if answerErr := ctx.answerCallbackQuery(""); answerErr != nil {

			log.Printf("Failed to answer callback query for UserID %d: %v", ctx.UserID(), answerErr)
		}
	}

	if err != nil {
		log.Printf("Handler error for UserID %d: %v", ctx.UserID(), err)
		if replyErr := ctx.sendSimpleText("An error occurred. Please try again."); replyErr != nil {
			log.Printf("Failed to send error reply to UserID %d: %v", ctx.UserID(), replyErr)
		}
	}
}

// isGlobalExitCommand checks if the given text matches any configured exit command.
// Exit commands allow users to cancel flows regardless of the current flow state.
func (b *Bot) isGlobalExitCommand(text string) bool {
	for _, cmd := range b.flowConfig.ExitCommands {
		if text == cmd {
			return true
		}
	}
	return false
}

// resolveGlobalCommandHandler finds a handler for commands that should be available globally,
// even when a user is in a flow. Currently supports help commands as defined in FlowConfig.
func (b *Bot) resolveGlobalCommandHandler(commandName string) HandlerFunc {

	for _, helpCmd := range b.flowConfig.HelpCommands {

		normalizedCmd := "/" + commandName
		if normalizedCmd == helpCmd || commandName == helpCmd {
			if handler, ok := b.handlers[commandName]; ok {
				return handler
			}
		}
	}
	return nil
}

// SetBotCommands configures the bot's command menu that appears in Telegram clients.
// This creates the command list that users see when typing "/" in a chat with the bot.
// Pass an empty map to clear all commands.
//
// Example:
//
//	commands := map[string]string{
//		"start": "Start the bot",
//		"help":  "Show help information",
//		"settings": "Configure bot settings",
//	}
//	err := bot.SetBotCommands(commands)
func (b *Bot) SetBotCommands(commands map[string]string) error {
	if b.api == nil {
		return fmt.Errorf("bot API not initialized")
	}
	if len(commands) == 0 {

		clearCmdCfg := tgbotapi.NewSetMyCommands()
		_, err := b.api.Request(clearCmdCfg)
		if err != nil {
			log.Printf("Warning: Failed to clear bot commands: %v", err)
			return fmt.Errorf("failed to clear bot commands: %w", err)
		}
		log.Printf("Bot commands cleared.")
		return nil
	}

	var tgCommands []tgbotapi.BotCommand
	for cmd, desc := range commands {
		tgCommands = append(tgCommands, tgbotapi.BotCommand{Command: cmd, Description: desc})
	}
	cmdCfg := tgbotapi.NewSetMyCommands(tgCommands...)
	_, err := b.api.Request(cmdCfg)
	if err != nil {
		log.Printf("Warning: Failed to set bot commands: %v", err)
		return fmt.Errorf("failed to set bot commands: %w", err)
	}

	return nil
}

// DeleteMessage deletes a specific message using the context and message ID.
// This is useful for cleaning up bot messages or implementing temporary notifications.
//
// Example:
//
//	err := bot.DeleteMessage(ctx, messageID)
func (b *Bot) DeleteMessage(ctx *Context, messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(ctx.ChatID(), messageID)
	_, err := b.api.Request(deleteMsg)
	return err
}

// EditMessageReplyMarkup edits the reply markup (inline keyboard) of a specific message.
// This allows dynamic updating of message keyboards without resending the entire message.
// To remove a keyboard completely, pass nil for replyMarkup.
//
// Example:
//
//	// Update keyboard
//	newKeyboard := teleflow.NewPromptKeyboard().
//		ButtonCallback("Updated", "data").Build()
//	err := bot.EditMessageReplyMarkup(ctx, messageID, newKeyboard)
//
//	// Remove keyboard
//	err := bot.EditMessageReplyMarkup(ctx, messageID, nil)
func (b *Bot) EditMessageReplyMarkup(ctx *Context, messageID int, replyMarkup interface{}) error {
	var editMsg tgbotapi.EditMessageReplyMarkupConfig

	if replyMarkup == nil {
		emptyKeyboard := tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		}
		// Remove keyboard by setting empty markup
		editMsg = tgbotapi.NewEditMessageReplyMarkup(ctx.ChatID(),
			messageID,
			emptyKeyboard)
	} else {
		// Assert to expected type
		keyboard, ok := replyMarkup.(tgbotapi.InlineKeyboardMarkup)
		if !ok {
			return fmt.Errorf("replyMarkup must be of type tgbotapi.InlineKeyboardMarkup")
		}
		editMsg = tgbotapi.NewEditMessageReplyMarkup(ctx.ChatID(), messageID, keyboard)
	}

	_, err := b.api.Request(editMsg)
	return err
}

// Start begins the bot's main event loop, listening for updates from Telegram.
// This method blocks indefinitely, processing updates concurrently as they arrive.
// It should typically be the last call in your main function.
//
// Example:
//
//	bot, err := teleflow.NewBot(token)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register handlers...
//
//	log.Println("Bot starting...")
//	if err := bot.Start(); err != nil {
//		log.Fatal(err)
//	}
func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.processUpdate(update)
	}

	return nil
}
