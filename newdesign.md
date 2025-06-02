# Teleflow - Enhanced Telegram Bot Framework Design

**Package Name:** `teleflow` (Telegram Flow)

## Core Philosophy

Teleflow is designed with three key principles:
1. **Flat Learning Curve**: Start simple, grow complex naturally
2. **Type Safety**: Compile-time safety for callbacks and flows
3. **Extensible Architecture**: Plugin-based middleware and modular components

## Package Structure

```
teleflow/
├── bot.go           // Main Bot struct, setup, routing
├── context.go       // Context struct and helper methods
├── handlers.go      // Handler types, registration, middleware
├── keyboards.go     // Type-safe keyboard abstractions
├── state.go         // State management with validation
├── flow.go          // Structured multi-step flows with DSL
├── middleware.go    // Middleware framework (auth, logging, etc.)
├── callbacks.go     // Type-safe callback system
├── templates.go     // Message templating
└── examples/        // Example usage patterns
```

## 1. Enhanced Bot Core (`bot.go`)

```go
package teleflow

import (
	"context"
	"log"
	"text/template"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerFunc defines the function signature for all interaction handlers
type HandlerFunc func(ctx *Context) error

// MiddlewareFunc defines the function signature for middleware
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

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
	mainMenu         *ReplyKeyboard
	menuButton       *MenuButtonConfig
	userPermissions  UserPermissionChecker
}

// UserPermissionChecker interface for authorization
type UserPermissionChecker interface {
	CanExecute(userID int64, action string) bool
	GetMainMenuForUser(userID int64) *ReplyKeyboard
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
		flowManager:      NewFlowManager(),
		templates:        template.New("botMessages"),
		middleware:       []MiddlewareFunc{},
	}

	// Apply options
	for _, opt := range options {
		opt(b)
	}

	return b, nil
}

// BotOption defines functional options for Bot configuration
type BotOption func(*Bot)

// WithMainMenu sets the default main ReplyKeyboard
func WithMainMenu(kb *ReplyKeyboard) BotOption {
	return func(b *Bot) {
		b.mainMenu = kb
	}
}

// WithUserPermissions sets the user permission checker
func WithUserPermissions(checker UserPermissionChecker) BotOption {
	return func(b *Bot) {
		b.userPermissions = checker
	}
}

// Use adds middleware to the bot
func (b *Bot) Use(middleware ...MiddlewareFunc) {
	b.middleware = append(b.middleware, middleware...)
}

// HandleCommand registers a handler for a specific command with permission check
func (b *Bot) HandleCommand(command string, handler HandlerFunc, permissions ...string) {
	wrappedHandler := b.wrapWithPermissions(handler, permissions...)
	b.handlers[command] = b.applyMiddleware(wrappedHandler)
}

// HandleText registers a handler for specific text input
func (b *Bot) HandleText(text string, handler HandlerFunc, permissions ...string) {
	wrappedHandler := b.wrapWithPermissions(handler, permissions...)
	b.textHandlers[text] = b.applyMiddleware(wrappedHandler)
}

// RegisterCallback registers a type-safe callback handler
func (b *Bot) RegisterCallback(callback CallbackHandler) {
	b.callbackRegistry.Register(callback)
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

// wrapWithPermissions wraps a handler with permission checking
func (b *Bot) wrapWithPermissions(handler HandlerFunc, permissions ...string) HandlerFunc {
	if len(permissions) == 0 || b.userPermissions == nil {
		return handler
	}

	return func(ctx *Context) error {
		for _, permission := range permissions {
			if !b.userPermissions.CanExecute(ctx.UserID(), permission) {
				return ctx.Reply("❌ You don't have permission to perform this action.")
			}
		}
		return handler(ctx)
	}
// processUpdate processes incoming updates with full middleware chain
func (b *Bot) processUpdate(ctx *Context) {
	var handler HandlerFunc

	// Check for global flow exit commands first
	if b.flowManager.IsUserInFlow(ctx.UserID()) {
		if b.isGlobalExitCommand(ctx) {
			handler = func(ctx *Context) error {
				b.flowManager.CancelFlow(ctx.UserID())
				return ctx.Reply(b.flowConfig.ExitMessage)
			}
		} else if b.flowConfig.AllowGlobalCommands && ctx.Update.Message != nil && ctx.Update.Message.IsCommand() {
			// Allow certain global commands during flows
			if globalHandler := b.resolveGlobalCommand(ctx); globalHandler != nil {
				handler = globalHandler
			}
		}
	}

	// Check if user is in a flow
	if handler == nil {
		if flowHandler := b.flowManager.HandleUpdate(ctx); flowHandler != nil {
			handler = b.applyMiddleware(flowHandler)
		} else {
			// Regular handler resolution
			handler = b.resolveHandler(ctx)
		}
	}

	if handler != nil {
		handler = b.applyMiddleware(handler)
		if err := handler(ctx); err != nil {
			log.Printf("Handler error for UserID %d: %v", ctx.UserID(), err)
			ctx.Reply("An error occurred. Please try again.")
		}
	}
}

// isGlobalExitCommand checks if the update contains a global exit command
func (b *Bot) isGlobalExitCommand(ctx *Context) bool {
	if ctx.Update.Message == nil {
		return false
	}
	
	text := ctx.Update.Message.Text
	for _, cmd := range b.flowConfig.ExitCommands {
		if text == cmd {
			return true
		}
	}
	return false
}

// resolveGlobalCommand resolves global commands that should work during flows
func (b *Bot) resolveGlobalCommand(ctx *Context) HandlerFunc {
	command := ctx.Update.Message.Command()
	
	// Only allow specific global commands during flows
	for _, helpCmd := range b.flowConfig.HelpCommands {
		if "/"+command == helpCmd || command == helpCmd {
			return b.handlers[command]
		}
	}
	
	return nil
}
	}
}

// resolveHandler resolves the appropriate handler for the update
func (b *Bot) resolveHandler(ctx *Context) HandlerFunc {
	if ctx.Update.Message != nil {
		if ctx.Update.Message.IsCommand() {
			return b.handlers[ctx.Update.Message.Command()]
		}
		return b.textHandlers[ctx.Update.Message.Text]
	}
	
	if ctx.Update.CallbackQuery != nil {
		return b.callbackRegistry.Handle(ctx.Update.CallbackQuery.Data)
	}
	
	return nil
}

// Start begins listening for updates
func (b *Bot) Start() {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		ctx := NewContext(b, update)
		go b.processUpdate(ctx)
	}
}
```

## 2. Enhanced Context (`context.go`)

```go
package teleflow

import (
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context provides information and helpers for the current interaction
type Context struct {
	Bot    *Bot
	Update tgbotapi.Update
	data   map[string]interface{}
	
	// User context
	userID int64
	chatID int64
}

// NewContext creates a new Context
func NewContext(bot *Bot, update tgbotapi.Update) *Context {
	ctx := &Context{
		Bot:    bot,
		Update: update,
		data:   make(map[string]interface{}),
	}
	
	ctx.userID = ctx.extractUserID()
	ctx.chatID = ctx.extractChatID()
	
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
func (c *Context) Reply(text string, keyboardMarkup ...interface{}) error {
	return c.send(text, "", keyboardMarkup...)
}

// ReplyTemplate sends a text message using a template
func (c *Context) ReplyTemplate(templateName string, data interface{}, keyboardMarkup ...interface{}) error {
	var buf bytes.Buffer
	if err := c.Bot.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return fmt.Errorf("executing template %s: %w", templateName, err)
	}
	return c.send(buf.String(), "", keyboardMarkup...)
}

// EditOrReply attempts to edit current message or sends new one
func (c *Context) EditOrReply(text string, keyboardMarkup ...interface{}) error {
	if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
		msg := tgbotapi.NewEditMessageText(
			c.ChatID(),
			c.Update.CallbackQuery.Message.MessageID,
			text,
		)
		
		if len(keyboardMarkup) > 0 {
			if ik, ok := keyboardMarkup[0].(*InlineKeyboard); ok {
				msg.ReplyMarkup = ik.toTgbotapi()
			}
		}
		
		if _, err := c.Bot.api.Send(msg); err == nil {
			cb := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, "")
			c.Bot.api.AnswerCallbackQuery(cb)
			return nil
		}
	}
	return c.Reply(text, keyboardMarkup...)
}

// StartFlow initiates a new flow for the user
func (c *Context) StartFlow(flowName string, initialData map[string]interface{}) error {
	return c.Bot.flowManager.StartFlow(c.UserID(), flowName, initialData)
}

// send is an internal helper for sending messages
func (c *Context) send(text, parseMode string, keyboardMarkup ...interface{}) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	// Apply keyboard markup
	if len(keyboardMarkup) > 0 && keyboardMarkup[0] != nil {
		switch kb := keyboardMarkup[0].(type) {
		case *ReplyKeyboard:
			msg.ReplyMarkup = kb.toTgbotapi()
		case *InlineKeyboard:
			msg.ReplyMarkup = kb.toTgbotapi()
		case tgbotapi.ReplyKeyboardRemove:
			msg.ReplyMarkup = kb
		}
	} else {
		// Apply user-specific main menu
		if c.Bot.userPermissions != nil {
			if userMenu := c.Bot.userPermissions.GetMainMenuForUser(c.UserID()); userMenu != nil {
				msg.ReplyMarkup = userMenu.toTgbotapi()
			}
		} else if c.Bot.mainMenu != nil {
			msg.ReplyMarkup = c.Bot.mainMenu.toTgbotapi()
		}
	}

	_, err := c.Bot.api.Send(msg)
	return err
}

// Helper methods to extract IDs
func (c *Context) extractUserID() int64 {
	if c.Update.Message != nil {
		return c.Update.Message.From.ID
	}
	if c.Update.CallbackQuery != nil {
		return c.Update.CallbackQuery.From.ID
	}
	return 0
}

func (c *Context) extractChatID() int64 {
	if c.Update.Message != nil {
		return c.Update.Message.Chat.ID
	}
	if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
		return c.Update.CallbackQuery.Message.Chat.ID
	}
	return 0
}
```

## 3. Type-Safe Callbacks (`callbacks.go`)

```go
package teleflow

import (
	"fmt"
	"reflect"
	"strings"
)

// CallbackHandler interface for type-safe callback handling
type CallbackHandler interface {
	Pattern() string
	Handle(ctx *Context, data string) error
}

// CallbackRegistry manages type-safe callback handlers
type CallbackRegistry struct {
	handlers map[string]CallbackHandler
	patterns []string
}

// NewCallbackRegistry creates a new callback registry
func NewCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry{
		handlers: make(map[string]CallbackHandler),
		patterns: []string{},
	}
}

// Register registers a callback handler
func (r *CallbackRegistry) Register(handler CallbackHandler) {
	pattern := handler.Pattern()
	r.handlers[pattern] = handler
	r.patterns = append(r.patterns, pattern)
}

// Handle finds and executes the appropriate callback handler
func (r *CallbackRegistry) Handle(callbackData string) HandlerFunc {
	for _, pattern := range r.patterns {
		if handler := r.handlers[pattern]; handler != nil {
			if data := r.matchPattern(pattern, callbackData); data != "" {
				return func(ctx *Context) error {
					return handler.Handle(ctx, data)
				}
			}
		}
	}
	return nil
}

// matchPattern checks if callback data matches pattern and extracts data
func (r *CallbackRegistry) matchPattern(pattern, callbackData string) string {
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		if strings.HasPrefix(callbackData, prefix) {
			return callbackData[len(prefix):]
		}
	} else if pattern == callbackData {
		return ""
	}
	return ""
}

// Helper function to create simple callback handlers
func SimpleCallback(pattern string, handler func(ctx *Context, data string) error) CallbackHandler {
	return &simpleCallbackHandler{
		pattern: pattern,
		handler: handler,
	}
}

type simpleCallbackHandler struct {
	pattern string
	handler func(ctx *Context, data string) error
}

func (h *simpleCallbackHandler) Pattern() string {
	return h.pattern
}

func (h *simpleCallbackHandler) Handle(ctx *Context, data string) error {
	return h.handler(ctx, data)
}

// Typed callback helpers for common patterns
type ActionCallback struct {
	Action  string
	Handler func(ctx *Context, actionData string) error
}

func (ac *ActionCallback) Pattern() string {
	return fmt.Sprintf("action_%s_*", ac.Action)
}

func (ac *ActionCallback) Handle(ctx *Context, data string) error {
	return ac.Handler(ctx, data)
}
```

## 4. Flow Management (`flow.go`)

```go
package teleflow

import (
	"fmt"
	"time"
)
// FlowManager manages all flows and user flow states
type FlowManager struct {
	flows     map[string]*Flow
	userFlows map[int64]*UserFlowState
	botConfig *FlowConfig
}

// NewFlowManager creates a new flow manager
func NewFlowManager() *FlowManager {
	return &FlowManager{
		flows:     make(map[string]*Flow),
		userFlows: make(map[int64]*UserFlowState),
	}
}

// SetBotConfig sets the bot configuration for the flow manager
func (fm *FlowManager) SetBotConfig(config *FlowConfig) {
	fm.botConfig = config
}

// IsUserInFlow checks if a user is currently in a flow
func (fm *FlowManager) IsUserInFlow(userID int64) bool {
	_, exists := fm.userFlows[userID]
	return exists
}

// CancelFlow cancels the current flow for a user
func (fm *FlowManager) CancelFlow(userID int64) {
	delete(fm.userFlows, userID)
}

// Flow represents a structured multi-step conversation
type Flow struct {
	Name         string
	Steps        []*FlowStep
	stepMap      map[string]*FlowStep
	transitions  map[string][]string
	OnComplete   HandlerFunc
	OnCancel     HandlerFunc
	Timeout      time.Duration
}
// FlowStep represents a single step in a flow
type FlowStep struct {
	Name               string
	Handler            HandlerFunc
	Validator          func(string) error // Simple validation function
	NextStep           string
	Transitions        map[string]string // input -> next step
	Timeout            time.Duration
	StayOnInvalidInput bool              // Stay in step (true) or cancel flow (false) on invalid input
}

// UserFlowState tracks a user's current position in a flow
type UserFlowState struct {
	FlowName    string
	CurrentStep string
	Data        map[string]interface{}
	StartedAt   time.Time
	LastActive  time.Time
}


// FlowBuilder provides a fluent interface for building flows
type FlowBuilder struct {
	flow *Flow
}

// NewFlow creates a new flow with the given name
func NewFlow(name string) *FlowBuilder {
	return &FlowBuilder{
		flow: &Flow{
			Name:        name,
			Steps:       []*FlowStep{},
			stepMap:     make(map[string]*FlowStep),
			transitions: make(map[string][]string),
			Timeout:     30 * time.Minute,
		},
	}
}

// Step adds a step to the flow
func (fb *FlowBuilder) Step(name string, handler HandlerFunc) *FlowStepBuilder {
	step := &FlowStep{
		Name:               name,
		Handler:            handler,
		Transitions:        make(map[string]string),
		Timeout:            5 * time.Minute,
		StayOnInvalidInput: true, // Default: stay on invalid input for better UX
	}
	
	fb.flow.Steps = append(fb.flow.Steps, step)
	fb.flow.stepMap[name] = step
	
	// Auto-link to previous step if this is not the first
	if len(fb.flow.Steps) > 1 {
		prevStep := fb.flow.Steps[len(fb.flow.Steps)-2]
		if prevStep.NextStep == "" {
			prevStep.NextStep = name
		}
	}
	
	return &FlowStepBuilder{
		flowBuilder: fb,
		step:        step,
	}
}

// OnComplete sets the completion handler
func (fb *FlowBuilder) OnComplete(handler HandlerFunc) *FlowBuilder {
	fb.flow.OnComplete = handler
	return fb
}

// OnCancel sets the cancellation handler
func (fb *FlowBuilder) OnCancel(handler HandlerFunc) *FlowBuilder {
	fb.flow.OnCancel = handler
	return fb
}

// Build creates the final flow
func (fb *FlowBuilder) Build() *Flow {
	return fb.flow
}

// FlowStepBuilder provides a fluent interface for building flow steps
type FlowStepBuilder struct {
	flowBuilder *FlowBuilder
	step        *FlowStep
}

// WithValidator adds input validation to the step
func (fsb *FlowStepBuilder) WithValidator(validator func(string) error) *FlowStepBuilder {
	fsb.step.Validator = validator
	return fsb
}

// NextStep sets the default next step
func (fsb *FlowStepBuilder) NextStep(stepName string) *FlowStepBuilder {
	fsb.step.NextStep = stepName
	return fsb
}

// OnInput adds a conditional transition based on input
func (fsb *FlowStepBuilder) OnInput(input, nextStep string) *FlowStepBuilder {
	fsb.step.Transitions[input] = nextStep
	return fsb
}

// WithTimeout sets a timeout for this step
func (fsb *FlowStepBuilder) WithTimeout(timeout time.Duration) *FlowStepBuilder {
	fsb.step.Timeout = timeout
	return fsb
}

// StayOnInvalidInput configures whether to stay in step on invalid input
func (fsb *FlowStepBuilder) StayOnInvalidInput(stay bool) *FlowStepBuilder {
	fsb.step.StayOnInvalidInput = stay
	return fsb
}


// Step continues building the flow with a new step
func (fsb *FlowStepBuilder) Step(name string, handler HandlerFunc) *FlowStepBuilder {
	return fsb.flowBuilder.Step(name, handler)
}

// OnComplete sets the completion handler
func (fsb *FlowStepBuilder) OnComplete(handler HandlerFunc) *FlowBuilder {
	return fsb.flowBuilder.OnComplete(handler)
}

// Build creates the final flow
func (fsb *FlowStepBuilder) Build() *Flow {
	return fsb.flowBuilder.Build()
}

// Flow management methods

// RegisterFlow registers a flow with the manager
func (fm *FlowManager) RegisterFlow(flow *Flow) {
	fm.flows[flow.Name] = flow
}

// StartFlow starts a flow for a user
func (fm *FlowManager) StartFlow(userID int64, flowName string, initialData map[string]interface{}) error {
	flow, exists := fm.flows[flowName]
	if !exists {
		return fmt.Errorf("flow %s not found", flowName)
	}
	
	if len(flow.Steps) == 0 {
		return fmt.Errorf("flow %s has no steps", flowName)
	}
	
	userState := &UserFlowState{
		FlowName:    flowName,
		CurrentStep: flow.Steps[0].Name,
		Data:        initialData,
		StartedAt:   time.Now(),
		LastActive:  time.Now(),
	}
	
	if userState.Data == nil {
		userState.Data = make(map[string]interface{})
	}
	
	fm.userFlows[userID] = userState
	return nil
}

// HandleUpdate processes an update for a user in a flow
func (fm *FlowManager) HandleUpdate(ctx *Context) HandlerFunc {
	userState, exists := fm.userFlows[ctx.UserID()]
	if !exists {
		return nil
	}
	
	flow := fm.flows[userState.FlowName]
	if flow == nil {
		delete(fm.userFlows, ctx.UserID())
		return nil
	}
	
	currentStep := flow.stepMap[userState.CurrentStep]
	if currentStep == nil {
		delete(fm.userFlows, ctx.UserID())
		return nil
	}
	
	return func(ctx *Context) error {
		// Update last active time
		userState.LastActive = time.Now()
		
		// Store user state data in context
		for key, value := range userState.Data {
			ctx.Set(key, value)
		}
		
		// Validate input if validator exists
		if currentStep.Validator != nil && ctx.Update.Message != nil {
			if err := currentStep.Validator.Validate(ctx.Update.Message.Text); err != nil {
				helpText := currentStep.Validator.GetHelpText()
				exitHint := ""
				if fm.botConfig != nil && len(fm.botConfig.ExitCommands) > 0 {
					exitHint = fmt.Sprintf("\n\nType '%s' to cancel.", fm.botConfig.ExitCommands[0])
				}
				return ctx.Reply(fmt.Sprintf("❌ %s\n\n%s%s", err.Error(), helpText, exitHint))
			}
		}
		
		// Execute step handler
		if err := currentStep.Handler(ctx); err != nil {
			return err
		}

		// Determine next step
		nextStep := fm.determineNextStep(ctx, currentStep, userState)
		
		if nextStep == "" {
			// Flow completed
			delete(fm.userFlows, ctx.UserID())
			if flow.OnComplete != nil {
				return flow.OnComplete(ctx)
			}
		} else if nextStep == "_cancel_" {
			// Flow cancelled
			delete(fm.userFlows, ctx.UserID())
			if flow.OnCancel != nil {
				return flow.OnCancel(ctx)
			}
		} else if nextStep == currentStep.Name {
			// Staying in current step due to invalid input
			if currentStep.InvalidInputMessage != "" {
				exitHint := ""
				if fm.botConfig != nil && len(fm.botConfig.ExitCommands) > 0 {
					exitHint = fmt.Sprintf(" Type '%s' to cancel.", fm.botConfig.ExitCommands[0])
				}
				ctx.Reply(currentStep.InvalidInputMessage + exitHint)
			}
			// Save context data back to user state but don't advance
			for key, value := range ctx.data {
				userState.Data[key] = value
			}
		} else {
			// Move to next step
			userState.CurrentStep = nextStep
			
			// Save context data back to user state
			for key, value := range ctx.data {
				userState.Data[key] = value
			}
		}
		
		return nil
	}
}
// determineNextStep determines the next step based on current step and user input
func (fm *FlowManager) determineNextStep(ctx *Context, currentStep *FlowStep, userState *UserFlowState) string {
	// 1. Check for global exit commands (handled at bot level, shouldn't reach here)
	
	// 2. Check for input-based transitions
	if ctx.Update.Message != nil {
		text := ctx.Update.Message.Text
		if nextStep, exists := currentStep.Transitions[text]; exists {
			return nextStep
		}
	}
	
	if ctx.Update.CallbackQuery != nil {
		data := ctx.Update.CallbackQuery.Data
		if nextStep, exists := currentStep.Transitions[data]; exists {
			return nextStep
		}
	}
	
	// 3. For unexpected input, check StayOnInvalidInput behavior
	if currentStep.StayOnInvalidInput {
		return currentStep.Name // Stay in current step
	} else {
		return "_cancel_" // Cancel flow on invalid input
	}
}

// Common validator helper functions
func NumberValidator(min, max float64) func(string) error {
	return func(input string) error {
		// Parse the number (implementation would go here)
		// if value < min || value > max {
		//     return fmt.Errorf("please enter a number between %.2f and %.2f", min, max)
		// }
		// For now, just return nil
		return nil
	}
}

func ChoiceValidator(validChoices ...string) func(string) error {
	return func(input string) error {
		for _, choice := range validChoices {
			if input == choice {
				return nil
			}
		}
		return fmt.Errorf("please choose one of: %v", validChoices)
	}
}
```

## 5. Middleware System (`middleware.go`)

```go
package teleflow

import (
	"log"
	"time"
)

// Built-in middleware functions

// LoggingMiddleware logs all incoming updates and handler execution time
func LoggingMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()
			
			// Log incoming update
			updateType := "unknown"
			if ctx.Update.Message != nil {
				if ctx.Update.Message.IsCommand() {
					updateType = "command: " + ctx.Update.Message.Command()
				} else {
					updateType = "text: " + ctx.Update.Message.Text
				}
			} else if ctx.Update.CallbackQuery != nil {
				updateType = "callback: " + ctx.Update.CallbackQuery.Data
			}
			
			log.Printf("[%d] Processing %s", ctx.UserID(), updateType)
			
			// Execute handler
			err := next(ctx)
			
			// Log execution time
			duration := time.Since(start)
			if err != nil {
				log.Printf("[%d] Handler failed in %v: %v", ctx.UserID(), duration, err)
			} else {
				log.Printf("[%d] Handler completed in %v", ctx.UserID(), duration)
			}
			
			return err
		}
	}
}

// AuthMiddleware checks if user is authorized
func AuthMiddleware(checker UserPermissionChecker) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			if !checker.CanExecute(ctx.UserID(), "basic_access") {
				return ctx.Reply("🚫 You are not authorized to use this bot.")
			}
			return next(ctx)
		}
	}
}

// RateLimitMiddleware implements simple rate limiting
func RateLimitMiddleware(requestsPerMinute int) MiddlewareFunc {
	userLastRequest := make(map[int64]time.Time)
	minInterval := time.Minute / time.Duration(requestsPerMinute)
	
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			userID := ctx.UserID()
			now := time.Now()
			
			if lastRequest, exists := userLastRequest[userID]; exists {
				if now.Sub(lastRequest) < minInterval {
					return ctx.Reply("⏳ Please wait before sending another message.")
				}
			}
			
			userLastRequest[userID] = now
			return next(ctx)
		}
	}
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in handler for user %d: %v", ctx.UserID(), r)
					err = ctx.Reply("An unexpected error occurred. Please try again.")
				}
			}()
			return next(ctx)
		}
	}
}
```

## 6. Example Usage

```go
package main

import (
	"log"
	"os"
	"teleflow"
)

// Simple permission checker implementation
type SimplePermissionChecker struct {
	adminUsers map[int64]bool
}

func (spc *SimplePermissionChecker) CanExecute(userID int64, action string) bool {
	if action == "admin" {
		return spc.adminUsers[userID]
	}
	return true // Allow basic access for all users
}

func (spc *SimplePermissionChecker) GetMainMenuForUser(userID int64) *teleflow.ReplyKeyboard {
	if spc.adminUsers[userID] {
		return teleflow.NewReplyKeyboard(
			[]teleflow.ReplyKeyboardButton{{Text: "💰 Balance"}, {Text: "💸 Transfer"}},
			[]teleflow.ReplyKeyboardButton{{Text: "👥 Admin Panel"}, {Text: "❓ Help"}},
		)
	}
	return teleflow.NewReplyKeyboard(
		[]teleflow.ReplyKeyboardButton{{Text: "💰 Balance"}, {Text: "💸 Transfer"}},
		[]teleflow.ReplyKeyboardButton{{Text: "❓ Help"}},
	)
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	// Create permission checker
	permissionChecker := &SimplePermissionChecker{
		adminUsers: map[int64]bool{
			123456789: true, // Replace with actual admin user IDs
		},
	}

	// Create bot
	bot, err := teleflow.NewBot(token,
		teleflow.WithUserPermissions(permissionChecker),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add middleware
	bot.Use(
		teleflow.RecoveryMiddleware(),
		teleflow.LoggingMiddleware(),
		teleflow.AuthMiddleware(permissionChecker),
		teleflow.RateLimitMiddleware(10), // 10 requests per minute
	)

	// Register templates
	bot.AddTemplate("welcome", "Hello {{.UserName}}! Welcome to our bot.")
	bot.AddTemplate("balance", "Your current balance is: ${{.Amount}}")

	// Register handlers
	bot.HandleCommand("start", startHandler)
	bot.HandleCommand("admin", adminHandler, "admin") // Requires admin permission

	// Register type-safe callbacks
	bot.RegisterCallback(&teleflow.ActionCallback{
		Action: "balance",
		Handler: func(ctx *teleflow.Context, data string) error {
			return ctx.EditOrReply("Balance: $100.50")
		},
	})

	// Register transfer flow
	transferFlow := teleflow.NewFlow("transfer").
		Step("amount", func(ctx *teleflow.Context) error {
			return ctx.Reply("Please enter the amount to transfer:")
		}).WithValidator(teleflow.NumberValidator(0.01, 10000)).
		Step("recipient", func(ctx *teleflow.Context) error {
			amount := ctx.Update.Message.Text
			ctx.Set("amount", amount)
			return ctx.Reply("Please enter the recipient's username:")
		}).
		Step("confirm", func(ctx *teleflow.Context) error {
			recipient := ctx.Update.Message.Text
			amount, _ := ctx.Get("amount")
			
			keyboard := teleflow.NewInlineKeyboard(
				[]teleflow.InlineKeyboardButton{
					{Text: "✅ Confirm", CallbackData: "transfer_confirm"},
					{Text: "❌ Cancel", CallbackData: "transfer_cancel"},
				},
			)
			
			return ctx.Reply(
				fmt.Sprintf("Transfer $%s to %s?", amount, recipient),
				keyboard,
			)
		}).OnInput("transfer_confirm", "").OnInput("transfer_cancel", "_cancel_").
		OnComplete(func(ctx *teleflow.Context) error {
			return ctx.EditOrReply("✅ Transfer completed successfully!")
		}).
		OnCancel(func(ctx *teleflow.Context) error {
			return ctx.EditOrReply("❌ Transfer cancelled.")
		}).
		Build()

	bot.RegisterFlow(transferFlow)

	// Text handlers for main menu
	bot.HandleText("💸 Transfer", func(ctx *teleflow.Context) error {
		return ctx.StartFlow("transfer", nil)
	})

	log.Println("Bot starting...")
	bot.Start()
}

func startHandler(ctx *teleflow.Context) error {
	return ctx.ReplyTemplate("welcome", map[string]string{
		"UserName": ctx.Update.Message.From.FirstName,
	})
}

func adminHandler(ctx *teleflow.Context) error {
	return ctx.Reply("🛠️ Admin Panel - You have administrative privileges.")
}
```

## Key Design Features

### 🎯 **Flat Learning Curve**
- Start with simple `HandleCommand` and `Reply`
- Gradually add flows, middleware, and advanced features
- Sensible defaults for everything

### 🔒 **Type Safety**
- Compile-time safe callback handlers
- Structured flow definitions
- Interface-based extensions

### 🚀 **Extensibility**
- Middleware system for cross-cutting concerns
- Pluggable state managers and permission checkers
- Template system for message management
- Flow DSL for complex conversations

### 🛡️ **Production Ready**
- Built-in authentication and authorization
- Comprehensive error handling and recovery
- Logging and monitoring capabilities
- Rate limiting and abuse prevention

This design maintains simplicity for beginners while providing enterprise-grade features for production use. The flow system makes complex conversational bots intuitive to build, and the middleware architecture ensures clean separation of concerns.

## UI Component Extensibility

The framework's architecture naturally supports custom UI components without requiring framework modifications. Users can build sophisticated interactive elements using the existing callback system and keyboard builders.

### UI Component Pattern

```go
package teleflow

// UIComponent interface for reusable UI elements
type UIComponent interface {
	Render(ctx *Context) (*InlineKeyboard, error)
	HandleCallback(ctx *Context, data string) error
	GetCallbackPattern() string
}

// BaseUIComponent provides common functionality
type BaseUIComponent struct {
	CallbackPrefix string
}

func (b *BaseUIComponent) GetCallbackPattern() string {
	return b.CallbackPrefix + "*"
}
```

### Example: Paginated Data Table

```go
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"teleflow"
)

type DataTable struct {
	teleflow.BaseUIComponent
	Data        [][]string
	Headers     []string
	PageSize    int
	CurrentPage int
	Title       string
}

func NewDataTable(title string, headers []string, data [][]string, pageSize int) *DataTable {
	return &DataTable{
		BaseUIComponent: teleflow.BaseUIComponent{CallbackPrefix: "table_"},
		Title:           title,
		Headers:         headers,
		Data:            data,
		PageSize:        pageSize,
		CurrentPage:     1,
	}
}

// Implement CallbackHandler interface
func (dt *DataTable) Pattern() string {
	return dt.GetCallbackPattern()
}

func (dt *DataTable) Handle(ctx *teleflow.Context, data string) error {
	switch {
	case data == "next":
		if dt.CurrentPage < dt.totalPages() {
			dt.CurrentPage++
		}
		return dt.Render(ctx)
	case data == "prev":
		if dt.CurrentPage > 1 {
			dt.CurrentPage--
		}
		return dt.Render(ctx)
	case strings.HasPrefix(data, "page_"):
		pageStr := strings.TrimPrefix(data, "page_")
		if page, err := strconv.Atoi(pageStr); err == nil {
			if page >= 1 && page <= dt.totalPages() {
				dt.CurrentPage = page
			}
		}
		return dt.Render(ctx)
	case strings.HasPrefix(data, "select_"):
		rowStr := strings.TrimPrefix(data, "select_")
		if rowIndex, err := strconv.Atoi(rowStr); err == nil {
			return dt.handleRowSelection(ctx, rowIndex)
		}
	}
	return nil
}

func (dt *DataTable) Render(ctx *teleflow.Context) error {
	keyboard := dt.buildKeyboard()
	message := dt.buildMessage()
	return ctx.EditOrReply(message, keyboard)
}

func (dt *DataTable) buildKeyboard() *teleflow.InlineKeyboard {
	var rows [][]teleflow.InlineKeyboardButton
	
	// Data rows for current page
	start := (dt.CurrentPage - 1) * dt.PageSize
	end := start + dt.PageSize
	if end > len(dt.Data) {
		end = len(dt.Data)
	}
	
	for i := start; i < end; i++ {
		row := dt.Data[i]
		displayText := strings.Join(row, " | ")
		if len(displayText) > 30 {
			displayText = displayText[:30] + "..."
		}
		
		rows = append(rows, []teleflow.InlineKeyboardButton{
			{Text: displayText, CallbackData: fmt.Sprintf("table_select_%d", i)},
		})
	}
	
	// Navigation row
	var navRow []teleflow.InlineKeyboardButton
	
	if dt.CurrentPage > 1 {
		navRow = append(navRow, teleflow.InlineKeyboardButton{
			Text: "◀️ Prev", CallbackData: "table_prev",
		})
	}
	
	navRow = append(navRow, teleflow.InlineKeyboardButton{
		Text: fmt.Sprintf("📄 %d/%d", dt.CurrentPage, dt.totalPages()),
		CallbackData: "table_info",
	})
	
	if dt.CurrentPage < dt.totalPages() {
		navRow = append(navRow, teleflow.InlineKeyboardButton{
			Text: "Next ▶️", CallbackData: "table_next",
		})
	}
	
	rows = append(rows, navRow)
	return teleflow.NewInlineKeyboard(rows...)
}

func (dt *DataTable) buildMessage() string {
	return fmt.Sprintf("📊 **%s**\n\nShowing page %d of %d (%d total records)",
		dt.Title, dt.CurrentPage, dt.totalPages(), len(dt.Data))
}

func (dt *DataTable) totalPages() int {
	return (len(dt.Data) + dt.PageSize - 1) / dt.PageSize
}

func (dt *DataTable) handleRowSelection(ctx *teleflow.Context, rowIndex int) error {
	if rowIndex >= 0 && rowIndex < len(dt.Data) {
		selectedRow := dt.Data[rowIndex]
		ctx.Set("selected_table_row", selectedRow)
		ctx.Set("selected_table_index", rowIndex)
		
		// Let the flow handle the selection
		return ctx.Reply(fmt.Sprintf("✅ Selected: %s", strings.Join(selectedRow, " | ")))
	}
	return nil
}
```

### Example: Date Picker Component

```go
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"teleflow"
)

type DatePicker struct {
	teleflow.BaseUIComponent
	MinDate     time.Time
	MaxDate     time.Time
	CurrentDate time.Time
	Title       string
}

func NewDatePicker(title string, minDate, maxDate time.Time) *DatePicker {
	return &DatePicker{
		BaseUIComponent: teleflow.BaseUIComponent{CallbackPrefix: "date_"},
		Title:           title,
		MinDate:         minDate,
		MaxDate:         maxDate,
		CurrentDate:     time.Now(),
	}
}

func (dp *DatePicker) Pattern() string {
	return dp.GetCallbackPattern()
}

func (dp *DatePicker) Handle(ctx *teleflow.Context, data string) error {
	switch {
	case data == "prev_month":
		dp.CurrentDate = dp.CurrentDate.AddDate(0, -1, 0)
		return dp.Render(ctx)
	case data == "next_month":
		dp.CurrentDate = dp.CurrentDate.AddDate(0, 1, 0)
		return dp.Render(ctx)
	case strings.HasPrefix(data, "select_"):
		dayStr := strings.TrimPrefix(data, "select_")
		if day, err := strconv.Atoi(dayStr); err == nil {
			selectedDate := time.Date(dp.CurrentDate.Year(), dp.CurrentDate.Month(), day, 0, 0, 0, 0, time.UTC)
			if !selectedDate.Before(dp.MinDate) && !selectedDate.After(dp.MaxDate) {
				ctx.Set("selected_date", selectedDate)
				return ctx.EditOrReply(fmt.Sprintf("✅ Selected date: %s", selectedDate.Format("2006-01-02")))
			}
		}
	}
	return nil
}

func (dp *DatePicker) Render(ctx *teleflow.Context) error {
	keyboard := dp.buildCalendar()
	message := fmt.Sprintf("📅 **%s**\n\nSelect a date:", dp.Title)
	return ctx.EditOrReply(message, keyboard)
}

func (dp *DatePicker) buildCalendar() *teleflow.InlineKeyboard {
	var rows [][]teleflow.InlineKeyboardButton
	
	// Month navigation
	monthRow := []teleflow.InlineKeyboardButton{
		{Text: "◀️", CallbackData: "date_prev_month"},
		{Text: dp.CurrentDate.Format("January 2006"), CallbackData: "date_month_info"},
		{Text: "▶️", CallbackData: "date_next_month"},
	}
	rows = append(rows, monthRow)
	
	// Day headers
	dayHeaders := []teleflow.InlineKeyboardButton{
		{Text: "S", CallbackData: "date_header_sun"},
		{Text: "M", CallbackData: "date_header_mon"},
		{Text: "T", CallbackData: "date_header_tue"},
		{Text: "W", CallbackData: "date_header_wed"},
		{Text: "T", CallbackData: "date_header_thu"},
		{Text: "F", CallbackData: "date_header_fri"},
		{Text: "S", CallbackData: "date_header_sat"},
	}
	rows = append(rows, dayHeaders)
	
	// Calendar days (simplified for brevity)
	year, month, _ := dp.CurrentDate.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	
	startWeekday := int(firstDay.Weekday())
	var currentRow []teleflow.InlineKeyboardButton
	
	// Empty cells for days before the first day of the month
	for i := 0; i < startWeekday; i++ {
		currentRow = append(currentRow, teleflow.InlineKeyboardButton{Text: " ", CallbackData: "date_empty"})
	}
	
	// Days of the month
	for day := 1; day <= lastDay.Day(); day++ {
		dayDate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		text := fmt.Sprintf("%d", day)
		callbackData := fmt.Sprintf("date_select_%d", day)
		
		// Mark unavailable dates
		if dayDate.Before(dp.MinDate) || dayDate.After(dp.MaxDate) {
			text = "·"
			callbackData = "date_unavailable"
		}
		
		currentRow = append(currentRow, teleflow.InlineKeyboardButton{
			Text: text, CallbackData: callbackData,
		})
		
		// New row every 7 days
		if len(currentRow) == 7 {
			rows = append(rows, currentRow)
			currentRow = []teleflow.InlineKeyboardButton{}
		}
	}
	
	// Fill remaining cells in the last row
	for len(currentRow) < 7 && len(currentRow) > 0 {
		currentRow = append(currentRow, teleflow.InlineKeyboardButton{Text: " ", CallbackData: "date_empty"})
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}
	
	return teleflow.NewInlineKeyboard(rows...)
}
```

### Usage in Flows

```go
// Using UI components in conversational flows
func setupOrderFlow(bot *teleflow.Bot) {
	// Register UI components
	productTable := ui.NewDataTable("Products", []string{"ID", "Name", "Price"}, productData, 5)
	bot.RegisterCallback(productTable)
	
	datePicker := ui.NewDatePicker("Delivery Date", time.Now(), time.Now().AddDate(0, 1, 0))
	bot.RegisterCallback(datePicker)
	
	// Create flow using UI components
	orderFlow := teleflow.NewFlow("order").
		Step("select_product", func(ctx *teleflow.Context) error {
			return productTable.Render(ctx)
		}).
		OnInput("table_select_*", "select_date"). // Any product selection advances
		Step("select_date", func(ctx *teleflow.Context) error {
			return datePicker.Render(ctx)
		}).
		OnInput("date_select_*", "confirm"). // Any date selection advances
		Step("confirm", func(ctx *teleflow.Context) error {
			selectedRow, _ := ctx.Get("selected_table_row")
			selectedDate, _ := ctx.Get("selected_date")
			
			return ctx.Reply(fmt.Sprintf("Order Summary:\n📦 Product: %v\n📅 Delivery: %v",
				selectedRow, selectedDate))
		}).
		Build()
	
	bot.RegisterFlow(orderFlow)
}
```

### Key Extensibility Features

#### **Zero Framework Changes Required**
- UI components use existing [`CallbackHandler`](newdesign.md:432) interface
- Leverage current [`InlineKeyboard`](newdesign.md:483) system
- Work with existing flow and state management

#### **Component Benefits**
- **Reusable**: Create once, use in multiple flows
- **Stateful**: Components maintain their own state
- **Type-Safe**: Compile-time callback safety
- **Composable**: Multiple components work together
- **Packageable**: Distribute as separate packages

#### **Advanced Patterns**
- **Component Libraries**: Third-party UI component packages
- **Composite Components**: Components that use other components
- **Theme Support**: Customizable appearance and behavior
- **Event System**: Components can emit custom events
- **Validation Integration**: UI components work with flow validators

This extensibility pattern allows the framework core to remain minimal while supporting unlimited UI complexity through external packages. Users can build sophisticated interfaces like admin panels, data browsers, form wizards, and interactive dashboards using the same consistent patterns.