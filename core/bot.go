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
//		return ctx.SendPromptText("Hello! Welcome to Teleflow!")
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
//			return ctx.SendPromptText(fmt.Sprintf("Welcome %s, age %s!", name, age))
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
// This is the most basic handler type used internally by the middleware system
// and as a common interface for all handler types after middleware application.
//
// Parameters:
//   - ctx: The context for the current update, providing access to bot API, state, user info, etc.
//
// Returns:
//   - error: An error if processing failed, nil otherwise.
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

// BotOption defines functional options for Bot configuration.
// BotOptions are used with NewBot to customize the bot's behavior during initialization.
// Multiple options can be chained together for flexible configuration.
//
// Example:
//
//	bot, err := NewBot(token,
//	  WithFlowConfig(customFlowConfig),
//	  WithAccessManager(myAccessManager))
type BotOption func(*Bot)

// PermissionContext provides rich context for permission checking within AccessManager.
// This struct contains comprehensive information about the current user interaction,
// enabling fine-grained permission control and context-aware decision making.
type PermissionContext struct {
	UserID    int64            // Telegram user ID of the person making the request
	ChatID    int64            // Telegram chat ID where the interaction occurred
	Command   string           // The command being executed (without leading slash)
	Arguments []string         // Command arguments split into individual strings
	IsGroup   bool             // True if the interaction occurred in a group chat
	IsChannel bool             // True if the interaction occurred in a channel
	MessageID int              // Telegram message ID of the triggering message
	Update    *tgbotapi.Update // Full Telegram update object for advanced context analysis
}

// AccessManager interface provides authorization and automatic UI management capabilities.
// It enables context-aware permission checking and automatic keyboard generation based on user roles.
// When used with WithAccessManager, it automatically applies optimal middleware stack including
// rate limiting and authentication.
//
// Implementation example:
//
//	type MyAccessManager struct {
//	  // your fields
//	}
//
//	func (m *MyAccessManager) CheckPermission(ctx *PermissionContext) error {
//	  // your permission logic
//	}
//
//	func (m *MyAccessManager) GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard {
//	  // return appropriate keyboard based on user permissions
//	}
type AccessManager interface {
	// CheckPermission checks if the user has permission to perform the requested action.
	// The method receives full context about the user, chat, command, and interaction type.
	//
	// Parameters:
	//   - ctx: Rich context containing user ID, chat info, command details, and full update
	//
	// Returns:
	//   - error: An error if permission is denied (message will be shown to user), nil if allowed
	CheckPermission(ctx *PermissionContext) error

	// GetReplyKeyboard returns the appropriate reply keyboard for the user based on their context.
	// This keyboard will be automatically applied to bot reply messages, providing
	// context-aware UI that adapts to user permissions and current state.
	//
	// Parameters:
	//   - ctx: Rich context for determining appropriate keyboard layout
	//
	// Returns:
	//   - *ReplyKeyboard: The keyboard to display, or nil for no keyboard
	GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard
}

// Bot is the main application structure that orchestrates all Telegram bot functionality.
// It manages handlers, middleware, flows, state, and provides the core update processing logic.
// The Bot struct is initialized via NewBot and configured through BotOption functions.
type Bot struct {
	// Core Telegram API integration
	api *tgbotapi.BotAPI // Direct interface to Telegram Bot API

	// Handler management
	handlers           map[string]HandlerFunc // Registered command handlers (key: command name without slash)
	textHandlers       map[string]HandlerFunc // Registered exact text match handlers
	defaultTextHandler HandlerFunc            // Fallback handler for unmatched text messages

	// Callback and interaction management
	callbackRegistry *callbackRegistry // Manages inline keyboard callback handlers with UUID mapping

	// State and flow management
	stateManager          StateManager           // Handles persistent user state across conversations
	flowManager           *flowManager           // Manages conversational flows and step transitions
	promptKeyboardHandler *PromptKeyboardHandler // Handles UUID mappings for prompt keyboards
	promptComposer        *PromptComposer        // Handles dynamic prompt composition and sending

	// Middleware system
	middleware []MiddlewareFunc // Unified middleware stack applied to all handler types

	// Configuration and access control
	accessManager AccessManager // Permission checking and automatic UI management
	flowConfig    FlowConfig    // Flow behavior configuration (exit commands, global commands, etc.)
}

// NewBot creates a new Bot instance with the specified Telegram bot token and optional configuration.
// This is the primary constructor for teleflow bots, setting up all necessary components with sensible defaults.
//
// The function performs several initialization steps:
//  1. Creates and validates the Telegram Bot API connection
//  2. Initializes core components (handlers, state management, flow system)
//  3. Sets up default flow configuration with exit commands and security settings
//  4. Applies any provided BotOption configurations
//  5. Initializes the flow system for immediate use
//
// Default Configuration:
//   - In-memory state management
//   - Flow exit commands: ["/cancel"]
//   - Exit message: "🚫 Operation cancelled."
//   - Global commands during flows: disabled (security)
//   - Help commands: ["/help"]
//   - Message processing: keep messages untouched
//
// Parameters:
//   - token: Telegram bot token obtained from @BotFather
//   - options: Optional BotOption functions for customizing bot behavior
//
// Returns:
//   - *Bot: Configured bot instance ready for handler registration and startup
//   - error: Error if bot token is invalid or initialization fails
//
// Example:
//
//	bot, err := NewBot("123456:ABC-DEF...",
//	  WithFlowConfig(customConfig),
//	  WithAccessManager(myAccessManager))
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
		promptKeyboardHandler: newPromptKeyboardHandler(),
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
	b.promptComposer = newPromptComposer(api, messageRenderer, imageHandler, b.promptKeyboardHandler)

	// Apply options
	for _, opt := range options {
		opt(b)
	}

	// initialize the flow system
	b.flowManager.initialize(b)
	return b, nil
}

// WithFlowConfig sets the flow configuration that controls conversational flow behavior.
// This BotOption allows customization of flow exit commands, global command handling,
// help commands, and message processing behavior during flows.
//
// Parameters:
//   - config: FlowConfig with exit commands, global command settings, and processing options
//
// Returns:
//   - BotOption: Configuration function to be used with NewBot
//
// Example:
//
//	flowConfig := FlowConfig{
//	  ExitCommands: []string{"/cancel", "/quit"},
//	  ExitMessage: "Operation cancelled!",
//	  AllowGlobalCommands: true,
//	  HelpCommands: []string{"/help", "/info"},
//	}
//	bot, err := NewBot(token, WithFlowConfig(flowConfig))
func WithFlowConfig(config FlowConfig) BotOption {
	return func(b *Bot) {
		b.flowConfig = config
	}
}

// WithAccessManager sets the user permission checker and automatically applies optimal middleware stack.
// This BotOption not only configures permission checking but also automatically applies
// a production-ready middleware stack in the correct order for security and performance.
//
// Automatic Middleware Stack Applied:
//  1. RateLimitMiddleware(60) - Prevents spam/abuse (60 requests per minute default)
//  2. AuthMiddleware - Permission checking with your AccessManager implementation
//
// The middleware order is optimized so rate limiting occurs before expensive permission
// checks, protecting against both spam attacks and unauthorized access attempts.
//
// Parameters:
//   - accessManager: Your AccessManager implementation for permission checking and UI management
//
// Returns:
//   - BotOption: Configuration function to be used with NewBot
//
// Example:
//
//	type MyAccessManager struct { /* your implementation */ }
//
//	accessManager := &MyAccessManager{}
//	bot, err := NewBot(token, WithAccessManager(accessManager))
//	// RateLimit + Auth middleware automatically applied
func WithAccessManager(accessManager AccessManager) BotOption {
	return func(b *Bot) {
		b.accessManager = accessManager

		// Apply middleware in optimal order for security and performance
		// Rate limiting comes before auth to prevent expensive operations on spam/attacks
		b.UseMiddleware(RateLimitMiddleware(60))       // 60 requests per minute default
		b.UseMiddleware(AuthMiddleware(accessManager)) // Permission checking after rate limiting
	}
}

// UseMiddleware adds general middleware that intercepts all message types and handler executions.
// This middleware will be applied to commands, text messages, callbacks, and flow processing.
//
// Middleware Execution Order:
// Middlewares are applied in reverse order of registration (LIFO - Last In, First Out).
// If you register middlewares A, B, C in that order, they execute as: C -> B -> A -> handler -> A -> B -> C
//
// This method provides a unified middleware system that works across all interaction types,
// eliminating the need for separate command, text, and callback middleware registration.
//
// Parameters:
//   - m: The MiddlewareFunc to add to the middleware stack
//
// Example:
//
//	bot.UseMiddleware(LoggingMiddleware())
//	bot.UseMiddleware(RecoveryMiddleware())
//	// Execution order: Recovery -> Logging -> handler -> Logging -> Recovery
func (b *Bot) UseMiddleware(m MiddlewareFunc) {
	b.middleware = append(b.middleware, m)
}

// HandleCommand registers a CommandHandlerFunc for a specific command.
// The commandName should be the command without the leading slash (e.g., "start" for "/start").
// All registered middleware will be automatically applied to the handler.
//
// The handler receives the command name and arguments as separate parameters for easy processing.
// Command arguments are extracted as a single string containing everything after the command.
//
// Parameters:
//   - commandName: The command name without leading slash (e.g., "help" for "/help")
//   - handler: The CommandHandlerFunc to execute when the command is received
//
// Example:
//
//	bot.HandleCommand("start", func(ctx *teleflow.Context, command string, args string) error {
//	  return ctx.Reply(fmt.Sprintf("Hello! Command: %s, Args: %s", command, args))
//	})
//
//	// Handles: "/start" and "/start some arguments"
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

// HandleText registers a TextHandlerFunc for messages that exactly match the given text.
// This enables handling of specific phrases, keywords, or menu button responses.
// All registered middleware will be automatically applied to the handler.
//
// The text matching is case-sensitive and requires exact string equality.
// For pattern matching or partial matches, use DefaultHandler with custom logic.
//
// Parameters:
//   - textToMatch: The exact text string to match (case-sensitive)
//   - handler: The TextHandlerFunc to execute when the text is matched
//
// Example:
//
//	bot.HandleText("📋 Show Menu", func(ctx *teleflow.Context, text string) error {
//	  return ctx.Reply("Here's your menu!")
//	})
//
//	bot.HandleText("help", func(ctx *teleflow.Context, text string) error {
//	  return ctx.Reply("Help information...")
//	})
func (b *Bot) HandleText(textToMatch string, handler TextHandlerFunc) {
	// Convert TextHandlerFunc to HandlerFunc and apply unified middleware
	wrappedHandler := func(ctx *Context) error {
		return handler(ctx, textToMatch)
	}
	b.textHandlers[textToMatch] = b.applyMiddleware(wrappedHandler)
}

// DefaultHandler registers a DefaultHandlerFunc to handle any text message that doesn't match
// other registered handlers. This serves as a fallback for unmatched text messages.
//
// Only one default handler can be registered; subsequent calls will overwrite the previous handler.
// All registered middleware will be automatically applied to the handler.
//
// The default handler is called for:
//   - Text messages that are not commands
//   - Text messages that don't match any HandleText registrations
//   - When no specific handler is found for the message
//
// Parameters:
//   - handler: The DefaultHandlerFunc to execute for unmatched text messages
//
// Example:
//
//	bot.DefaultHandler(func(ctx *teleflow.Context, text string) error {
//	  return ctx.Reply("I didn't understand that. Type /help for assistance.")
//	})
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

// RegisterFlow registers a conversational flow with the bot's flow management system.
// Flows enable complex, multi-step conversations with validation, branching, and state management.
//
// The flow must be built using the FlowBuilder before registration. Once registered,
// flows can be triggered by their trigger commands or programmatically via Context methods.
//
// Parameters:
//   - flow: A complete flow built using NewFlow().Step()...Build()
//
// Example:
//
//	flow := teleflow.NewFlow("registration").
//	  Step("name").
//	  Prompt("What's your name?", nil, nil).
//	  Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//	    ctx.Set("name", input)
//	    return teleflow.NextStep()
//	  }).
//	  Build()
//
//	bot.RegisterFlow(flow)
func (b *Bot) RegisterFlow(flow *Flow) {
	b.flowManager.registerFlow(flow)
}

// GetPromptKeyboardHandler returns the PromptKeyboardHandler instance for advanced flow integration.
// This is primarily used internally by the flow system but may be needed for custom implementations
// that require direct access to prompt keyboard UUID mapping and callback handling.
//
// Returns:
//   - *PromptKeyboardHandler: The bot's prompt keyboard handler instance
func (b *Bot) GetPromptKeyboardHandler() *PromptKeyboardHandler {
	return b.promptKeyboardHandler
}

// applyMiddleware applies all registered middleware to a handler in reverse order (LIFO).
// This is an internal method called during handler registration to wrap handlers with middleware.
// The middleware stack is applied once during registration for optimal performance.
//
// Parameters:
//   - handler: The base handler to wrap with middleware
//
// Returns:
//   - HandlerFunc: The handler wrapped with all registered middleware
func (b *Bot) applyMiddleware(handler HandlerFunc) HandlerFunc {
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}
	return handler
}

// processUpdate processes incoming Telegram updates through a comprehensive decision tree.
// This is the core method that determines how each update should be handled based on:
// user flow state, message type, registered handlers, and bot configuration.
//
// Processing Decision Tree:
//  1. Check if user is in a flow
//     a. Handle flow exit commands (global exit commands like /cancel)
//     b. Handle global commands if enabled (help commands during flows)
//     c. Delegate to flow manager for step processing
//  2. If not in flow, handle regular messages:
//     a. Commands -> specific command handlers
//     b. Text messages -> exact text handlers or default handler
//     c. Callback queries -> registered callback handlers
//  3. Error handling and user feedback for all cases
//
// The method runs in its own goroutine (called from Start) to handle concurrent updates.
//
// Parameters:
//   - update: The Telegram update object containing the message/callback/etc.
func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := newContext(b, update)
	var err error

	// Check for global flow exit commands first
	if b.flowManager.isUserInFlow(ctx.UserID()) {
		if ctx.update.Message != nil && b.isGlobalExitCommand(ctx.update.Message.Text) {
			b.flowManager.cancelFlow(ctx.UserID()) // Consider if OnCancel should be triggered
			err = ctx.sendSimpleText(b.flowConfig.ExitMessage)
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
		if replyErr := ctx.sendSimpleText("An error occurred. Please try again."); replyErr != nil {
			log.Printf("Failed to send error reply to UserID %d: %v", ctx.UserID(), replyErr)
		}
	}
}

// isGlobalExitCommand checks if the given text matches any configured flow exit command.
// Flow exit commands allow users to cancel ongoing flows and return to normal bot operation.
// These commands work globally regardless of the current flow step.
//
// Parameters:
//   - text: The message text to check against configured exit commands
//
// Returns:
//   - bool: True if the text matches any exit command, false otherwise
func (b *Bot) isGlobalExitCommand(text string) bool {
	for _, cmd := range b.flowConfig.ExitCommands {
		if text == cmd {
			return true
		}
	}
	return false
}

// resolveGlobalCommandHandler resolves commands that are allowed to execute during flows.
// When AllowGlobalCommands is enabled in FlowConfig, certain commands (like help commands)
// can be executed without interrupting the current flow state.
//
// The method checks if the command is in the configured HelpCommands list and returns
// the appropriate handler if found. Command matching is flexible, supporting both
// "/command" and "command" formats.
//
// Parameters:
//   - commandName: The command name to resolve (with or without leading slash)
//
// Returns:
//   - HandlerFunc: The handler for the global command, or nil if not allowed/found
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

// resolveCallbackHandler resolves callback handlers for inline keyboard interactions.
// It returns a HandlerFunc that wraps the specific CallbackHandler.Handle method,
// enabling callback handling through the unified middleware system.
//
// The callback registry uses UUID-based routing to match callback data with registered handlers,
// providing type-safe callback handling with automatic data extraction.
//
// Parameters:
//   - callbackData: The callback_data from the inline keyboard button press
//
// Returns:
//   - HandlerFunc: A wrapped handler that processes the callback, or nil if not found
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

// SetBotCommands configures the bot's command menu that appears when users type "/".
// These commands are set globally for the bot on Telegram and appear in the command suggestions.
// This is different from WithMenuButton, which sets persistent menu buttons.
//
// The method handles both setting new commands and clearing existing commands (when map is empty).
// Commands are automatically prefixed with "/" in the Telegram UI.
//
// Parameters:
//   - commands: Map of command names (without slash) to their descriptions
//
// Returns:
//   - error: Error if the API request fails, nil on success
//
// Example:
//
//	err := bot.SetBotCommands(map[string]string{
//	  "start": "Start the bot",
//	  "help":  "Show help information",
//	  "about": "About this bot",
//	})
//
//	// To clear commands:
//	err := bot.SetBotCommands(map[string]string{})
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
	// log.Printf("Successfully set %d bot commands.", len(tgCommands))
	return nil
}

// Start begins listening for Telegram updates and processes them concurrently.
// This method blocks and runs the bot's main event loop, processing incoming messages,
// commands, callbacks, and other updates through the registered handlers and flows.
//
// Each update is processed in its own goroutine to handle concurrent users efficiently.
// The method sets up long polling with a 60-second timeout for optimal performance.
//
// The bot will log its authorization status and continue running until an error occurs
// or the process is terminated.
//
// Returns:
//   - error: Error if bot startup fails (invalid token, network issues, etc.)
//
// Example:
//
//	bot, err := NewBot(token)
//	if err != nil {
//	  log.Fatal(err)
//	}
//
//	// Register handlers...
//	bot.HandleCommand("start", startHandler)
//
//	// Start the bot (blocks)
//	log.Fatal(bot.Start())
func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	// Initialize menu button if configured (only for web_app or default types)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.processUpdate(update)
	}

	return nil
}
