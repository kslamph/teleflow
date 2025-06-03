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

// HandlerFunc defines the function signature for all interaction handlers
type HandlerFunc func(ctx *Context) error

// MiddlewareFunc defines the function signature for middleware
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

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
	api              *tgbotapi.BotAPI
	handlers         map[string]HandlerFunc
	textHandlers     map[string]HandlerFunc
	callbackRegistry *CallbackRegistry
	stateManager     StateManager
	flowManager      *FlowManager
	templates        *template.Template
	middleware       []MiddlewareFunc

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
		api:              api,
		handlers:         make(map[string]HandlerFunc),
		textHandlers:     make(map[string]HandlerFunc),
		callbackRegistry: NewCallbackRegistry(),
		stateManager:     NewInMemoryStateManager(),
		flowManager:      NewFlowManager(NewInMemoryStateManager()),
		templates:        template.New("botMessages"),
		middleware:       []MiddlewareFunc{},
		flowConfig: FlowConfig{
			ExitCommands:        []string{"/cancel", "/exit"},
			ExitMessage:         "Operation cancelled.",
			AllowGlobalCommands: false,
			HelpCommands:        []string{"/help"},
		},
	}

	// Apply options
	for _, opt := range options {
		opt(b)
	}

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

// Use adds middleware to the bot
func (b *Bot) Use(middleware MiddlewareFunc) {
	b.middleware = append(b.middleware, middleware)
}

// HandleCommand registers a handler for a specific command
func (b *Bot) HandleCommand(command string, handler HandlerFunc) {
	b.handlers[command] = b.applyMiddleware(handler)
}

// HandleText registers a handler for specific text input
// If text is empty string, it registers a default handler for any unmatched text
func (b *Bot) HandleText(text string, handler HandlerFunc) {
	b.textHandlers[text] = b.applyMiddleware(handler)
}

// RegisterCallback registers a type-safe callback handler
func (b *Bot) RegisterCallback(handler CallbackHandler) {
	b.callbackRegistry.Register(handler)
}

// RegisterFlow registers a flow with the flow manager
func (b *Bot) RegisterFlow(flow *Flow) {
	b.flowManager.RegisterFlow(flow)
}

// applyMiddleware applies all registered middleware to a handler
func (b *Bot) applyMiddleware(handler HandlerFunc) HandlerFunc {
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}
	return handler
}

// processUpdate processes incoming updates with full middleware chain
func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := NewContext(b, update)
	var handler HandlerFunc

	// Check for global flow exit commands first
	if b.flowManager.IsUserInFlow(ctx.UserID()) {
		if ctx.Update.Message != nil && b.isGlobalExitCommand(ctx.Update.Message.Text) {
			handler = func(ctx *Context) error {
				b.flowManager.CancelFlow(ctx.UserID())
				return ctx.Reply(b.flowConfig.ExitMessage)
			}
		} else if b.flowConfig.AllowGlobalCommands && ctx.Update.Message != nil && ctx.Update.Message.IsCommand() {
			// Allow certain global commands during flows
			if globalHandler := b.resolveGlobalCommand(ctx.Update.Message.Command()); globalHandler != nil {
				handler = globalHandler
			}
		}
	}

	// Check if user is in a flow
	if handler == nil {
		if handled, err := b.flowManager.HandleUpdate(ctx); handled {
			if err != nil {
				handler = func(ctx *Context) error { return err }
			} else {
				return // Flow handled the update, no further processing needed
			}
		} else {
			// Regular handler resolution
			handler = b.resolveHandler(update)
		}
	}

	if handler != nil {
		if err := handler(ctx); err != nil {
			log.Printf("Handler error for UserID %d: %v", ctx.UserID(), err)
			if ctx.Reply("An error occurred. Please try again.") != nil {
				log.Printf("Failed to send error reply to UserID %d: %v", ctx.UserID(), err)
			}
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

// resolveGlobalCommand resolves global commands that should work during flows
func (b *Bot) resolveGlobalCommand(command string) HandlerFunc {
	// Only allow specific global commands during flows
	for _, helpCmd := range b.flowConfig.HelpCommands {
		if "/"+command == helpCmd || command == helpCmd {
			return b.handlers[command]
		}
	}

	return nil
}

// resolveHandler resolves the appropriate handler for the update
func (b *Bot) resolveHandler(update tgbotapi.Update) HandlerFunc {
	if update.Message != nil {
		if update.Message.IsCommand() {
			return b.handlers[update.Message.Command()]
		}
		// First try specific text handler, then fall back to default
		if handler, exists := b.textHandlers[update.Message.Text]; exists {
			return handler
		}
		return b.textHandlers[""] // Default text handler
	}

	if update.CallbackQuery != nil {
		return b.callbackRegistry.Handle(update.CallbackQuery.Data)
	}

	return nil
}

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
