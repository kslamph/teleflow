// Package teleflow provides a powerful, type-safe Go framework for building
// sophisticated Telegram bots with intuitive APIs and advanced flow management.
//
// The framework offers comprehensive features including:
//   - Type-safe APIs for commands, callbacks, and user interactions
//   - Advanced conversation flow management with validation and branching
//   - Comprehensive middleware system for authentication, logging, and rate limiting
//   - Intuitive keyboard abstractions for reply and inline keyboards
//   - Persistent state management across conversation flows
//   - Powerful template system for dynamic message content
//   - Rich context objects with helper methods for common operations
//
// Basic Usage:
//
//	bot, err := teleflow.NewBot("YOUR_BOT_TOKEN")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Initialize the flow system before using flows
//	bot.InitializeFlowSystem()
//
//	bot.HandleCommand("start", func(ctx *teleflow.Context, command string, args string) error {
//		return ctx.Reply("Hello! Welcome to Teleflow!")
//	})
//
//	log.Fatal(bot.Start())
//
// Flow-based Conversations (New Step-Prompt-Process API):
//
//	flow := teleflow.NewFlow("registration").
//		Step("name").
//		Prompt("What's your name?", nil, nil).
//		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//			if input == "" {
//				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{Message: "Please enter your name:"})
//			}
//			ctx.Set("name", input)
//			return teleflow.NextStep()
//		}).
//		Step("age").
//		Prompt("How old are you?", nil, nil).
//		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//			if input == "" {
//				return teleflow.RetryWithPrompt(&teleflow.PromptConfig{Message: "Please enter your age:"})
//			}
//			ctx.Set("age", input)
//			return teleflow.NextStep()
//		}).
//		OnComplete(func(ctx *teleflow.Context, flowData map[string]interface{}) error {
//			name, _ := flowData["name"].(string)
//			age, _ := flowData["age"].(string)
//			return ctx.Reply(fmt.Sprintf("Welcome %s, age %s!", name, age))
//		}).
//		Build()
//
//	bot.RegisterFlow(flow)
//
// Middleware Integration:
//
//	// Manual middleware setup
//	bot.UseMiddleware(teleflow.LoggingMiddleware())
//	bot.UseMiddleware(teleflow.RecoveryMiddleware())
//
//	// OR automatic with AccessManager (includes rate limiting + auth)
//	bot, _ := teleflow.NewBot(token, teleflow.WithAccessManager(accessManager))
//
// For comprehensive examples and documentation, see the examples/ directory
// and the documentation at docs/.
package teleflow

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Types implemented in keyboards.go will be imported

// StateManager interface is implemented in state.go

// FlowManager, Flow, and related types are now implemented in flow.go

// HandlerFunc defines the general function signature for interaction handlers.
type HandlerFunc func(ctx *Context) error

// CommandHandlerFunc defines the function signature for command handlers.
// It is called when a registered command is received.
//
// Parameters:
//   - ctx: The context for the current update, providing access to bot API, state, etc.
//   - command: The command name (e.g., "start" for "/start").
//   - args: The arguments passed with the command as a single string.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type CommandHandlerFunc func(ctx *Context, command string, args string) error

// TextHandlerFunc defines the function signature for specific text message handlers.
// It is called when a message exactly matches a registered text string.
//
// Parameters:
//   - ctx: The context for the current update.
//   - text: The full text of the message that matched.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type TextHandlerFunc func(ctx *Context, text string) error

// DefaultHandlerFunc defines the function signature for the default text message handler.
// It is called for any text message that is not a command and does not match any specific TextHandlerFunc.
//
// Parameters:
//   - ctx: The context for the current update.
//   - text: The full text of the received message.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type DefaultHandlerFunc func(ctx *Context, text string) error

// BotOption defines functional options for Bot configuration
type BotOption func(*Bot)

// PermissionContext provides rich context for permission checking
type PermissionContext struct {
	UserID    int64
	ChatID    int64
	Command   string
	Arguments []string
	IsGroup   bool
	IsChannel bool
	MessageID int
	Update    *tgbotapi.Update
}

// AccessManager interface for authorization and automatic UI management
// It provides context-aware keyboards and menu buttons based on user permissions
type AccessManager interface {
	// CheckPermission checks if the user has permission to perform an action
	// Returns an error if permission is denied, nil if allowed
	// The error message is used to inform the user
	CheckPermission(ctx *PermissionContext) error

	// GetReplyKeyboard returns the reply keyboard for the user based on context
	// This keyboard will be automatically applied to reply messages
	GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard
}

// Bot is the main application structure
type Bot struct {
	api                   *tgbotapi.BotAPI
	handlers              map[string]HandlerFunc
	textHandlers          map[string]HandlerFunc
	defaultTextHandler    HandlerFunc // Field for the default text handler
	callbackRegistry      *callbackRegistry
	stateManager          StateManager
	flowManager           *flowManager
	promptKeyboardHandler *PromptKeyboardHandler // Handles UUID mappings for prompt keyboards
	promptComposer        *PromptComposer        // Handles prompt sending
	// Unified middleware system - intercepts all message types
	middleware []MiddlewareFunc

	// Configuration
	replyKeyboard *ReplyKeyboard
	menuButton    *MenuButtonConfig // Only for web_app or default types
	accessManager AccessManager     // AccessManager for user permission checking and replyKeyboard management
	flowConfig    FlowConfig
}

// NewBot creates a new Bot instance
func NewBot(token string, options ...BotOption) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	b := &Bot{
		api:                   api,
		handlers:              make(map[string]HandlerFunc),
		textHandlers:          make(map[string]HandlerFunc),
		callbackRegistry:      newCallbackRegistry(),
		stateManager:          NewInMemoryStateManager(),
		flowManager:           newFlowManager(NewInMemoryStateManager()),
		promptKeyboardHandler: NewPromptKeyboardHandler(),
		middleware:            make([]MiddlewareFunc, 0),
		flowConfig: FlowConfig{
			ExitCommands:        []string{"/cancel"},
			ExitMessage:         "🚫 Operation cancelled.",
			AllowGlobalCommands: false,
			HelpCommands:        []string{"/help"},
			OnProcessAction:     ProcessKeepMessage, // Default: keep messages untouched
		},
	}

	// Initialize PromptComposer with required dependencies
	messageRenderer := newMessageRenderer()
	imageHandler := newImageHandler()
	b.promptComposer = NewPromptComposer(api, messageRenderer, imageHandler, b.promptKeyboardHandler)

	// Apply options
	for _, opt := range options {
		opt(b)
	}

	// initialize the flow system
	b.flowManager.initialize(b)
	return b, nil
}

// WithMenuButton sets the default menu button configuration for web_app or default types only.
// For bot commands, use SetBotCommands() method instead.
func WithMenuButton(config *MenuButtonConfig) BotOption {
	return func(b *Bot) {
		// Only allow web_app or default types for WithMenuButton
		if config != nil && (config.Type == menuButtonTypeWebApp || config.Type == menuButtonTypeDefault) {
			b.menuButton = config
		}
	}
}

// WithFlowConfig sets the flow configuration
func WithFlowConfig(config FlowConfig) BotOption {
	return func(b *Bot) {
		b.flowConfig = config
	}
}

// WithAccessManager sets the user permission checker and automatically applies auth middleware
// This automatically sets up the optimal middleware order:
// 1. RecoveryMiddleware (catch panics)
// 2. LoggingMiddleware (monitor requests)
// 3. RateLimitMiddleware (block spam before expensive operations)
// 4. AuthMiddleware (permission checking with database/business logic)
func WithAccessManager(accessManager AccessManager) BotOption {
	return func(b *Bot) {
		b.accessManager = accessManager

		// Apply middleware in optimal order for security and performance
		// Rate limiting comes before auth to prevent expensive operations on spam/attacks
		b.UseMiddleware(RateLimitMiddleware(60))       // 60 requests per minute default
		b.UseMiddleware(AuthMiddleware(accessManager)) // Permission checking after rate limiting
	}
}

// Adapter functions to convert general middleware to type-specific middleware

// Use adds a general middleware that will be applied to ALL handler types (commands, text, callbacks, flow steps).
// This middleware is automatically converted to the appropriate type-specific middleware and applied to all handlers.
// Middlewares are applied in the reverse order they are added.
//
// Parameters:
//   - m: The MiddlewareFunc to add to all handler types.
//
// UseMiddleware adds general middleware that intercepts all message types.
// This middleware will be applied to commands, text messages, callbacks, and flows.
// Middlewares are applied in the reverse order they are added.
//
// Parameters:
//   - m: The MiddlewareFunc to add.
func (b *Bot) UseMiddleware(m MiddlewareFunc) {
	b.middleware = append(b.middleware, m)
}

// HandleCommand registers a CommandHandlerFunc for a specific command.
// The commandName should be the command without the leading slash (e.g., "start" for "/start").
// Any registered command middleware will be applied to the handler.
//
// Parameters:
//   - commandName: The name of the command to handle (e.g., "help").
//   - handler: The CommandHandlerFunc to execute when the command is received.
func (b *Bot) HandleCommand(commandName string, handler CommandHandlerFunc) {
	// Convert CommandHandlerFunc to HandlerFunc and apply unified middleware
	wrappedHandler := func(ctx *Context) error {
		// Extract command and args from the update
		command := commandName
		args := ""
		if ctx.update.Message != nil && len(ctx.update.Message.Text) > len(command)+1 {
			args = ctx.update.Message.Text[len(command)+1:]
		}
		return handler(ctx, command, args)
	}
	b.handlers[commandName] = b.applyMiddleware(wrappedHandler)
}

// HandleText registers a TextHandlerFunc for a message that exactly matches the given text.
// Any registered text middleware will be applied to the handler.
//
// Parameters:
//   - textToMatch: The exact text string to match.
//   - handler: The TextHandlerFunc to execute when the text is matched.
func (b *Bot) HandleText(textToMatch string, handler TextHandlerFunc) {
	// Convert TextHandlerFunc to HandlerFunc and apply unified middleware
	wrappedHandler := func(ctx *Context) error {
		return handler(ctx, textToMatch)
	}
	b.textHandlers[textToMatch] = b.applyMiddleware(wrappedHandler)
}

// DefaultHandler registers a DefaultTextHandlerFunc to be called for any text message
// that is not a command and does not match any handler registered with HandleText.
// Only one default text handler can be set; subsequent calls will overwrite the previous one.
// Any registered unified middleware will be applied to the handler.
//
// Parameters:
//   - handler: The DefaultTextHandlerFunc to execute.
func (b *Bot) DefaultHandler(handler DefaultHandlerFunc) {
	// Convert DefaultTextHandlerFunc to HandlerFunc and apply unified middleware
	wrappedHandler := func(ctx *Context) error {
		var text string
		if ctx.update.Message != nil {
			text = ctx.update.Message.Text
		}
		return handler(ctx, text)
	}
	b.defaultTextHandler = b.applyMiddleware(wrappedHandler)
}

// RegisterFlow registers a flow with the flow manager
func (b *Bot) RegisterFlow(flow *Flow) {
	b.flowManager.registerFlow(flow)
}

// GetPromptKeyboardHandler returns the PromptKeyboardHandler instance
func (b *Bot) GetPromptKeyboardHandler() *PromptKeyboardHandler {
	return b.promptKeyboardHandler
}

// applyMiddleware applies all registered general middleware to a handler.
// This is an internal method called during handler registration.
func (b *Bot) applyMiddleware(handler HandlerFunc) HandlerFunc {
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}
	return handler
}

// processUpdate processes incoming updates
func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := NewContext(b, update)
	var err error

	// Check for global flow exit commands first
	if b.flowManager.isUserInFlow(ctx.UserID()) {
		if ctx.update.Message != nil && b.isGlobalExitCommand(ctx.update.Message.Text) {
			b.flowManager.cancelFlow(ctx.UserID()) // Consider if OnCancel should be triggered
			err = ctx.Reply(b.flowConfig.ExitMessage)
			if err != nil {
				log.Printf("Error sending flow exit message: %v", err)
			}
			return // Exit command processed
		} else if b.flowConfig.AllowGlobalCommands && ctx.update.Message != nil && ctx.update.Message.IsCommand() {
			commandName := ctx.update.Message.Command()
			if cmdHandler := b.resolveGlobalCommandHandler(commandName); cmdHandler != nil {
				// Middleware already applied at registration
				err = cmdHandler(ctx)
				if err != nil {
					log.Printf("Global command handler error for UserID %d, command '%s': %v", ctx.UserID(), commandName, err)
					// Potentially send an error reply
				}
				return // Global command processed
			}
		}
	}

	// Check if user is in a flow
	if handledByFlow, flowErr := b.flowManager.HandleUpdate(ctx); handledByFlow {
		if flowErr != nil {
			log.Printf("Flow handler error for UserID %d: %v", ctx.UserID(), flowErr)
			// Potentially send an error reply
		}
		return // Flow handled the update or errored, no further regular processing
	}

	// Regular handler resolution and execution
	if update.Message != nil {
		if update.Message.IsCommand() {
			commandName := update.Message.Command()
			if cmdHandler, ok := b.handlers[commandName]; ok {
				err = cmdHandler(ctx) // Middleware already applied at registration
			} else {
				// No specific command handler, check for default text handler if configured
				if b.defaultTextHandler != nil {
					err = b.defaultTextHandler(ctx)
				}
			}
		} else { // Regular text message
			text := update.Message.Text
			if textHandler, ok := b.textHandlers[text]; ok {
				err = textHandler(ctx) // Middleware already applied at registration
			} else if b.defaultTextHandler != nil { // Fallback to default text handler
				err = b.defaultTextHandler(ctx)
			}
			// If no specific text handler and no default, message is ignored unless flow handles it
		}
	} else if update.CallbackQuery != nil {
		// Handle callback queries from inline keyboards
		genericHandler := b.resolveCallbackHandler(update.CallbackQuery.Data)
		if genericHandler != nil {
			err = genericHandler(ctx) // This genericHandler internally calls the specific CallbackHandler.Handle
		}

		// Always answer callback query to dismiss loading indicator
		if answerErr := ctx.answerCallbackQuery(""); answerErr != nil {
			// Log error but don't fail the main processing
			log.Printf("Failed to answer callback query for UserID %d: %v", ctx.UserID(), answerErr)
		}
	}

	if err != nil {
		log.Printf("Handler error for UserID %d: %v", ctx.UserID(), err)
		if replyErr := ctx.Reply("An error occurred. Please try again."); replyErr != nil {
			log.Printf("Failed to send error reply to UserID %d: %v", ctx.UserID(), replyErr)
		}
	}
}

// isGlobalExitCommand checks if the update contains a global exit command
func (b *Bot) isGlobalExitCommand(text string) bool {
	for _, cmd := range b.flowConfig.ExitCommands {
		if text == cmd {
			return true
		}
	}
	return false
}

// resolveGlobalCommandHandler resolves global commands that should work during flows
func (b *Bot) resolveGlobalCommandHandler(commandName string) HandlerFunc {
	// Only allow specific global commands during flows
	for _, helpCmd := range b.flowConfig.HelpCommands {
		// Ensure command matching is consistent (e.g. with or without '/')
		normalizedCmd := "/" + commandName
		if normalizedCmd == helpCmd || commandName == helpCmd { // Check both /cmd and cmd
			if handler, ok := b.handlers[commandName]; ok {
				return handler
			}
		}
	}
	return nil
}

// resolveCallbackHandler resolves callback handlers.
// It returns a generic HandlerFunc that wraps the specific CallbackHandler.Handle call.
// This is because the CallbackRegistry.Handle method is designed this way.
func (b *Bot) resolveCallbackHandler(callbackData string) HandlerFunc {
	// The CallbackRegistry's Handle method already returns a HandlerFunc
	// which, when executed, calls the specific CallbackHandler.Handle method
	// with the correct arguments (ctx, fullCallbackData, extractedData).
	return b.callbackRegistry.handle(callbackData)
}

// Note: resolveHandler is largely integrated into processUpdate now.
// Individual resolution logic for commands and text is directly in processUpdate.
// Callbacks are handled via resolveCallbackHandler.
// If a more complex, unified resolveHandler is needed later, it would need to return an interface{}
// and then processUpdate would do type assertions.
// For now, processUpdate directly accesses b.handlers, b.textHandlers, and calls resolveCallbackHandler.

// SetBotCommands configures the bot's menu button with the given commands.
// These commands are set globally for the bot on Telegram.
// The map keys are the command strings (without leading slash), and values are descriptions.
func (b *Bot) SetBotCommands(commands map[string]string) error {
	if b.api == nil {
		return fmt.Errorf("bot API not initialized")
	}
	if len(commands) == 0 {
		// To clear commands, send an empty list
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
	log.Printf("Successfully set %d bot commands.", len(tgCommands))
	return nil
}

// Start begins listening for updates
func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	// Initialize menu button if configured (only for web_app or default types)
	b.initializeMenuButton()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.processUpdate(update)
	}

	return nil
}
