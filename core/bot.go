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
//	bot := teleflow.NewBot("YOUR_BOT_TOKEN")
//
//	bot.HandleCommand("/start", func(ctx *teleflow.Context) error {
//		return ctx.Reply("Hello! Welcome to Teleflow!")
//	})
//
//	log.Fatal(bot.Start())
//
// Flow-based Conversations:
//
//	flow := teleflow.NewFlow("registration").
//		AddStep("name", teleflow.StepTypeText, "What's your name?").
//		AddStep("age", teleflow.StepTypeText, "How old are you?")
//
//	bot.RegisterFlow(flow, func(ctx *teleflow.Context, result map[string]string) error {
//		return ctx.Reply(fmt.Sprintf("Welcome %s, age %s!", result["name"], result["age"]))
//	})
//
// Middleware Integration:
//
//	bot.Use(teleflow.LoggingMiddleware())
//	bot.Use(teleflow.AuthMiddleware(authChecker))
//	bot.Use(teleflow.RateLimitMiddleware(10, time.Minute))
//
// For comprehensive examples and documentation, see the examples/ directory
// and the documentation at docs/.
package teleflow

import (
	"log"
	"text/template"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Types implemented in keyboards.go will be imported

// StateManager interface is implemented in state.go

// FlowManager, Flow, and related types are now implemented in flow.go

// HandlerFunc defines the general function signature for interaction handlers.
// It will be replaced by more specific handler types.
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

// DefaultTextHandlerFunc defines the function signature for the default text message handler.
// It is called for any text message that is not a command and does not match any specific TextHandlerFunc.
//
// Parameters:
//   - ctx: The context for the current update.
//   - fullMessageText: The full text of the received message.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
type DefaultTextHandlerFunc func(ctx *Context, fullMessageText string) error

// MiddlewareFunc defines the function signature for middleware.
// TODO: Refactor middleware system to work with specific handler types.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc // This HandlerFunc is generic and needs refactoring

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
	GetReplyKeyboard(ctx *MenuContext) *ReplyKeyboard

	// GetMenuButton returns the menu button configuration for the user based on context
	// This menu button will be automatically set for the chat
	GetMenuButton(ctx *MenuContext) *MenuButtonConfig
}

// Bot is the main application structure
type Bot struct {
	api                     *tgbotapi.BotAPI
	handlers                map[string]CommandHandlerFunc
	textHandlers            map[string]TextHandlerFunc
	defaultTextHandler      DefaultTextHandlerFunc // Field for the default text handler
	callbackRegistry        *CallbackRegistry
	stateManager            StateManager
	flowManager             *FlowManager
	templates               *template.Template
	commandMiddleware       []CommandMiddlewareFunc
	textMiddleware          []TextMiddlewareFunc
	defaultTextMiddleware   []DefaultTextMiddlewareFunc
	callbackMiddleware      []CallbackMiddlewareFunc
	flowStepInputMiddleware []FlowStepInputMiddlewareFunc
	// middleware         []MiddlewareFunc // Old generic middleware

	// Configuration
	replyKeyboard *ReplyKeyboard
	menuButton    *MenuButtonConfig
	accessManager AccessManager // AccessManager for user permission checking and replyKeyboard management
	flowConfig    FlowConfig
}

// NewBot creates a new Bot instance
func NewBot(token string, options ...BotOption) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		api:          api,
		handlers:     make(map[string]CommandHandlerFunc),
		textHandlers: make(map[string]TextHandlerFunc),
		// defaultTextHandler will be initialized as nil by default
		callbackRegistry: NewCallbackRegistry(make([]CallbackMiddlewareFunc, 0)...), // Pass empty slice initially
		stateManager:     NewInMemoryStateManager(),
		flowManager:      NewFlowManager(NewInMemoryStateManager()),
		templates:        template.New("botMessages"),
		// middleware:       []MiddlewareFunc{}, // Old generic middleware
		commandMiddleware:       make([]CommandMiddlewareFunc, 0),
		textMiddleware:          make([]TextMiddlewareFunc, 0),
		defaultTextMiddleware:   make([]DefaultTextMiddlewareFunc, 0),
		callbackMiddleware:      make([]CallbackMiddlewareFunc, 0),
		flowStepInputMiddleware: make([]FlowStepInputMiddlewareFunc, 0),
		flowConfig: FlowConfig{
			ExitCommands:        []string{"/cancel", "/exit"},
			ExitMessage:         "Operation cancelled.",
			AllowGlobalCommands: false,
			HelpCommands:        []string{"/help"},
		},
	}

	// Apply options
	// Apply options AFTER core Bot fields (like middleware slices) are initialized
	for _, opt := range options {
		opt(b)
	}

	// Now that b.callbackMiddleware is potentially populated by options (e.g., WithCallbackMiddleware option if added),
	// re-initialize CallbackRegistry with the Bot's actual middleware slice.
	// This is a bit indirect; a cleaner way might involve BotOptions that can access/modify the registry post-init,
	// or deferring registry creation until after options. For now, this ensures it gets the Bot's slice.
	b.callbackRegistry = NewCallbackRegistry(b.callbackMiddleware...)

	return b, nil
}

// WithMenuButton sets the menu button configuration using functional options
func WithMenuButton(config *MenuButtonConfig) BotOption {
	return func(b *Bot) {
		b.menuButton = config
	}
}

// WithMenuButton sets the menu button configuration (method style)
func (b *Bot) WithMenuButton(config *MenuButtonConfig) {
	b.menuButton = config
}

// WithFlowConfig sets the flow configuration
func WithFlowConfig(config FlowConfig) BotOption {
	return func(b *Bot) {
		b.flowConfig = config
	}
}

// WithExitCommands sets the global exit commands
func WithExitCommands(commands []string) BotOption {
	return func(b *Bot) {
		b.flowConfig.ExitCommands = commands
	}
}

// WithAccessManager sets the user permission checker
func WithAccessManager(accessManager AccessManager) BotOption {
	return func(b *Bot) {
		b.accessManager = accessManager
	}
}

// UseCommandMiddleware adds a middleware that will be applied to all command handlers.
// Middlewares are applied in the reverse order they are added.
//
// Parameters:
//   - m: The CommandMiddlewareFunc to add.
func (b *Bot) UseCommandMiddleware(m CommandMiddlewareFunc) {
	b.commandMiddleware = append(b.commandMiddleware, m)
}

// UseTextMiddleware adds a middleware that will be applied to all specific text handlers
// (registered via HandleText).
// Middlewares are applied in the reverse order they are added.
//
// Parameters:
//   - m: The TextMiddlewareFunc to add.
func (b *Bot) UseTextMiddleware(m TextMiddlewareFunc) {
	b.textMiddleware = append(b.textMiddleware, m)
}

// UseDefaultTextMiddleware adds a middleware that will be applied to the default text handler
// (registered via SetDefaultTextHandler).
// Middlewares are applied in the reverse order they are added.
//
// Parameters:
//   - m: The DefaultTextMiddlewareFunc to add.
func (b *Bot) UseDefaultTextMiddleware(m DefaultTextMiddlewareFunc) {
	b.defaultTextMiddleware = append(b.defaultTextMiddleware, m)
}

// UseCallbackMiddleware adds a middleware that will be applied to all callback handlers
// registered via the CallbackRegistry.
// Middlewares are applied in the reverse order they are added.
// The CallbackRegistry is responsible for applying these middlewares.
//
// Parameters:
//   - m: The CallbackMiddlewareFunc to add.
func (b *Bot) UseCallbackMiddleware(m CallbackMiddlewareFunc) {
	b.callbackMiddleware = append(b.callbackMiddleware, m)
}

// UseFlowStepInputMiddleware adds a middleware that will be applied to all flow step input handlers.
// These are handlers associated with steps in a conversation flow that expect user input.
// Middlewares are applied in the reverse order they are added.
// The FlowManager is responsible for applying these middlewares.
//
// Parameters:
//   - m: The FlowStepInputMiddlewareFunc to add.
func (b *Bot) UseFlowStepInputMiddleware(m FlowStepInputMiddlewareFunc) {
	b.flowStepInputMiddleware = append(b.flowStepInputMiddleware, m)
}

// HandleCommand registers a CommandHandlerFunc for a specific command.
// The commandName should be the command without the leading slash (e.g., "start" for "/start").
// Any registered command middleware will be applied to the handler.
//
// Parameters:
//   - commandName: The name of the command to handle (e.g., "help").
//   - handler: The CommandHandlerFunc to execute when the command is received.
func (b *Bot) HandleCommand(commandName string, handler CommandHandlerFunc) {
	b.handlers[commandName] = b.applyCommandMiddleware(handler)
}

// HandleText registers a TextHandlerFunc for a message that exactly matches the given text.
// Any registered text middleware will be applied to the handler.
//
// Parameters:
//   - textToMatch: The exact text string to match.
//   - handler: The TextHandlerFunc to execute when the text is matched.
func (b *Bot) HandleText(textToMatch string, handler TextHandlerFunc) {
	b.textHandlers[textToMatch] = b.applyTextMiddleware(handler)
}

// SetDefaultTextHandler registers a DefaultTextHandlerFunc to be called for any text message
// that is not a command and does not match any handler registered with HandleText.
// Only one default text handler can be set; subsequent calls will overwrite the previous one.
// Any registered default text middleware will be applied to the handler.
//
// Parameters:
//   - handler: The DefaultTextHandlerFunc to execute.
func (b *Bot) SetDefaultTextHandler(handler DefaultTextHandlerFunc) {
	b.defaultTextHandler = b.applyDefaultTextMiddleware(handler)
}

// RegisterCallback registers a CallbackHandler with the bot's CallbackRegistry.
// The CallbackRegistry manages matching callback queries to their handlers and applying
// any relevant callback middleware.
//
// Parameters:
//   - handler: The CallbackHandler to register. This handler must implement the
//     CallbackHandler interface, defining its pattern and handling logic.
func (b *Bot) RegisterCallback(handler CallbackHandler) {
	b.callbackRegistry.Register(handler)
}

// RegisterFlow registers a flow with the flow manager
func (b *Bot) RegisterFlow(flow *Flow) {
	b.flowManager.RegisterFlow(flow)
}

// applyCommandMiddleware applies all registered command middleware to a handler.
// This is an internal method called during handler registration.
func (b *Bot) applyCommandMiddleware(handler CommandHandlerFunc) CommandHandlerFunc {
	for i := len(b.commandMiddleware) - 1; i >= 0; i-- {
		handler = b.commandMiddleware[i](handler)
	}
	return handler
}

// applyTextMiddleware applies all registered text middleware to a handler.
// This is an internal method called during handler registration.
func (b *Bot) applyTextMiddleware(handler TextHandlerFunc) TextHandlerFunc {
	for i := len(b.textMiddleware) - 1; i >= 0; i-- {
		handler = b.textMiddleware[i](handler)
	}
	return handler
}

// applyDefaultTextMiddleware applies all registered default text middleware to a handler.
// This is an internal method called during handler registration.
func (b *Bot) applyDefaultTextMiddleware(handler DefaultTextHandlerFunc) DefaultTextHandlerFunc {
	for i := len(b.defaultTextMiddleware) - 1; i >= 0; i-- {
		handler = b.defaultTextMiddleware[i](handler)
	}
	return handler
}

// applyCallbackMiddleware applies all registered callback middleware to a handler.
// This is an internal method, typically used by the CallbackRegistry.
func (b *Bot) applyCallbackMiddleware(handler CallbackHandlerFunc) CallbackHandlerFunc {
	for i := len(b.callbackMiddleware) - 1; i >= 0; i-- {
		handler = b.callbackMiddleware[i](handler)
	}
	return handler
}

// applyFlowStepInputMiddleware applies all registered flow step input middleware to a handler.
// This is an internal method, typically used by the FlowManager.
func (b *Bot) applyFlowStepInputMiddleware(handler FlowStepInputHandlerFunc) FlowStepInputHandlerFunc {
	for i := len(b.flowStepInputMiddleware) - 1; i >= 0; i-- {
		handler = b.flowStepInputMiddleware[i](handler)
	}
	return handler
}

// processUpdate processes incoming updates
func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := NewContext(b, update)
	var err error

	// Check for global flow exit commands first
	if b.flowManager.IsUserInFlow(ctx.UserID()) {
		if ctx.Update.Message != nil && b.isGlobalExitCommand(ctx.Update.Message.Text) {
			b.flowManager.CancelFlow(ctx.UserID()) // Consider if OnCancel should be triggered
			err = ctx.Reply(b.flowConfig.ExitMessage)
			if err != nil {
				log.Printf("Error sending flow exit message: %v", err)
			}
			return // Exit command processed
		} else if b.flowConfig.AllowGlobalCommands && ctx.Update.Message != nil && ctx.Update.Message.IsCommand() {
			commandName := ctx.Update.Message.Command()
			if cmdHandler := b.resolveGlobalCommandHandler(commandName); cmdHandler != nil {
				args := ctx.Update.Message.CommandArguments()
				// TODO: Apply middleware if/when adapted for CommandHandlerFunc
				err = cmdHandler(ctx, commandName, args)
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
				args := ctx.Update.Message.CommandArguments()
				err = cmdHandler(ctx, commandName, args) // Middleware already applied at registration
			} else {
				// No specific command handler, check for default text handler if configured
				if b.defaultTextHandler != nil {
					err = b.defaultTextHandler(ctx, update.Message.Text)
				}
			}
		} else { // Regular text message
			text := update.Message.Text
			if textHandler, ok := b.textHandlers[text]; ok {
				err = textHandler(ctx, text) // Middleware already applied at registration
			} else if b.defaultTextHandler != nil { // Fallback to default text handler
				err = b.defaultTextHandler(ctx, text)
			}
			// If no specific text handler and no default, message is ignored unless flow handles it
		}
	} else if update.CallbackQuery != nil {
		// CallbackQuery handling is now more direct in CallbackRegistry.Handle
		// which returns a HandlerFunc. That HandlerFunc (wrapper) calls the specific CallbackHandler.Handle
		// The CallbackRegistry.Handle itself is called from resolveHandler (which needs update)
		// or directly if we refactor resolveHandler out for callbacks.
		// For now, let's assume resolveHandler will give us the wrapped HandlerFunc.
		// This part needs to align with how resolveHandler is refactored.
		// The old resolveHandler returned a generic HandlerFunc.
		// The new approach should involve CallbackRegistry.Handle returning the specific handler or a wrapper.
		// Let's adjust resolveHandler first.
		// For now, this logic path will be hit if resolveHandler is called.
		// The CallbackRegistry.Handle method returns a HandlerFunc that wraps the specific CallbackHandler.Handle call.
		genericHandler := b.resolveCallbackHandler(update.CallbackQuery.Data)
		if genericHandler != nil {
			err = genericHandler(ctx) // This genericHandler internally calls the specific CallbackHandler.Handle
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
func (b *Bot) resolveGlobalCommandHandler(commandName string) CommandHandlerFunc {
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
	return b.callbackRegistry.Handle(callbackData)
}

// Note: resolveHandler is largely integrated into processUpdate now.
// Individual resolution logic for commands and text is directly in processUpdate.
// Callbacks are handled via resolveCallbackHandler.
// If a more complex, unified resolveHandler is needed later, it would need to return an interface{}
// and then processUpdate would do type assertions.
// For now, processUpdate directly accesses b.handlers, b.textHandlers, and calls resolveCallbackHandler.

// Start begins listening for updates
func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	// Initialize menu button if configured
	b.InitializeMenuButton()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.processUpdate(update)
	}

	return nil
}
