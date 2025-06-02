**Package Name Idea:** `teleflow` (Telegram Flow) or `teleguide`

**Core Concepts & Design Principles:**

1.  **Declarative Approach:** Developers define *what* should happen rather than *how* the Telegram API details work.
2.  **Handler-Based:** Similar to web frameworks, specific functions (handlers) will manage different bot interactions.
3.  **Context Object:** A central `Context` object passed to every handler, providing access to the update, bot API, state, and helper methods.
4.  **State Management:** Built-in, simple state management for multi-step conversations.
5.  **Flows/Wizards:** Define sequences of interactions for complex business logic.
6.  **Templating:** Easy-to-use message templating.
7.  **Keyboard Abstractions:** Simple ways to define `ReplyKeyboardMarkup` and `InlineKeyboardMarkup`.
8.  **Clear Separation:** Business logic (developer's code) vs. framework logic.

**Proposed Structure and Components:**

```
teleflow/
├── bot.go           // Main Bot struct, setup, routing
├── context.go       // Context struct and its helper methods
├── handlers.go      // Handler types, registration
├── keyboards.go     // Helpers for creating keyboards
├── state.go         // State management logic
├── flow.go          // (Optional, advanced) For defining complex, multi-step flows
├── templates.go     // Message templating
└── (examples/)      // Example usage
```

**1. `bot.go` - The Core Engine**

```go
package teleflow

import (
	"log"
	"text/template"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerFunc defines the function signature for all interaction handlers.
type HandlerFunc func(ctx *Context) error

// Bot is the main application structure.
type Bot struct {
	api           *tgbotapi.BotAPI
	handlers      map[string]HandlerFunc         // For commands like /start, /help
	textHandlers  map[string]HandlerFunc         // For specific text inputs (e.g., from ReplyKeyboard)
	callbackHandlers map[string]HandlerFunc    // For inline button callbacks (can use prefix matching)
	stateHandlers map[string]map[string]HandlerFunc // stateName -> inputText/callback -> HandlerFunc
	defaultHandler HandlerFunc                  // Fallback for unhandled messages
	menuButton    *MenuButtonConfig              // Configuration for the bot menu button
	mainMenu      *ReplyKeyboard               // Default main reply keyboard
	stateManager  StateManager
	templates     *template.Template
}

// NewBot creates a new Bot instance.
func NewBot(token string, options ...BotOption) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	// api.Debug = true // Optional: enable debugging

	b := &Bot{
		api:           api,
		handlers:      make(map[string]HandlerFunc),
		textHandlers:  make(map[string]HandlerFunc),
		callbackHandlers: make(map[string]HandlerFunc),
		stateHandlers: make(map[string]map[string]HandlerFunc),
		stateManager:  NewInMemoryStateManager(), // Default state manager
		templates:     template.New("botMessages"),
	}

	for _, opt := range options {
		opt(b)
	}
	return b, nil
}

// BotOption defines functional options for Bot configuration.
type BotOption func(*Bot)

// WithMainMenu sets the default main ReplyKeyboard.
func WithMainMenu(kb *ReplyKeyboard) BotOption {
	return func(b *Bot) {
		b.mainMenu = kb
	}
}

// WithMenuButton configures the bot's main menu button.
func WithMenuButton(config *MenuButtonConfig) BotOption {
	return func(b *Bot) {
		b.menuButton = config
		// TODO: Call setChatMenuButton API here or on Start()
	}
}

// AddTemplate parses and adds a named message template.
func (b *Bot) AddTemplate(name, tmpl string) error {
	_, err := b.templates.New(name).Parse(tmpl)
	return err
}

// HandleCommand registers a handler for a specific command (e.g., "/start").
func (b *Bot) HandleCommand(command string, handler HandlerFunc) {
	b.handlers[command] = handler
}

// HandleText registers a handler for specific text input (often from ReplyKeyboard).
func (b *Bot) HandleText(text string, handler HandlerFunc) {
	b.textHandlers[text] = handler
}

// HandleCallback registers a handler for an inline keyboard callback query.
// Supports prefix matching if callbackData ends with "*", e.g., "action_*".
func (b *Bot) HandleCallback(callbackData string, handler HandlerFunc) {
	b.callbackHandlers[callbackData] = handler
}

// OnState registers a handler for messages received when the user is in a specific state.
// 'trigger' can be text from a reply keyboard or a callback_data prefix.
func (b *Bot) OnState(stateName string, trigger string, handler HandlerFunc) {
	if _, ok := b.stateHandlers[stateName]; !ok {
		b.stateHandlers[stateName] = make(map[string]HandlerFunc)
	}
	b.stateHandlers[stateName][trigger] = handler
}

// OnStateDefault registers a handler for any message received when in a specific state
// if no specific trigger matches.
func (b *Bot) OnStateDefault(stateName string, handler HandlerFunc) {
	b.OnState(stateName, "_default_", handler)
}


// SetDefaultHandler sets a handler for messages that don't match any other handlers.
func (b *Bot) SetDefaultHandler(handler HandlerFunc) {
	b.defaultHandler = handler
}

// Start begins listening for updates and processing them.
func (b *Bot) Start() {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

    // Set the bot menu button if configured
	if b.menuButton != nil {
		cfg := tgbotapi.MenuButton{Type: b.menuButton.Type}
		if b.menuButton.Type == "commands" {
			// Default behavior, nothing to set explicitly unless you want to override default
		} else if b.menuButton.Type == "web_app" {
			cfg.Text = b.menuButton.Text
			cfg.WebApp = &tgbotapi.WebAppInfo{URL: b.menuButton.WebAppURL}
		}
        // You can set it for all private chats, or allow per-chat configuration
		// For simplicity, we'll assume a global setting for now.
		// _, err := b.api.Request(tgbotapi.SetChatMenuButtonConfig{MenuButton: cfg})
        // if err != nil {
        //  log.Printf("Error setting menu button: %v", err)
        // }
        // More robust: call SetChatMenuButton API endpoint
        log.Println("Menu button configuration would be applied here.")
	}


	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		ctx := NewContext(b, update) // Create context for each update
		go b.processUpdate(ctx)      // Process updates concurrently (optional, consider implications)
	}
}

func (b *Bot) processUpdate(ctx *Context) {
	var handler HandlerFunc
	var err error

	currentState, _ := b.stateManager.GetState(ctx.UserID())

	// 1. Check state-specific handlers first
	if currentState != "" {
		if stateSpecificHandlers, ok := b.stateHandlers[currentState]; ok {
			trigger := ""
			if ctx.Update.Message != nil {
				trigger = ctx.Update.Message.Text
			} else if ctx.Update.CallbackQuery != nil {
				trigger = ctx.Update.CallbackQuery.Data
			}

			if h, ok := stateSpecificHandlers[trigger]; ok {
				handler = h
			} else if h, ok := stateSpecificHandlers["_default_"]; ok { // Default for the state
                handler = h
            }
		}
	}

	// 2. If no state handler, check general handlers
	if handler == nil {
		if ctx.Update.Message != nil {
			if ctx.Update.Message.IsCommand() {
				handler = b.handlers[ctx.Update.Message.Command()]
			} else if h, ok := b.textHandlers[ctx.Update.Message.Text]; ok {
				handler = h
			}
		} else if ctx.Update.CallbackQuery != nil {
			// Simple prefix matching for callbacks
			for pattern, h := range b.callbackHandlers {
				if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
					prefix := pattern[:len(pattern)-1]
					if strings.HasPrefix(ctx.Update.CallbackQuery.Data, prefix) {
						handler = h
						break
					}
				} else if ctx.Update.CallbackQuery.Data == pattern {
					handler = h
					break
				}
			}
		}
	}
	
	// 3. Fallback to default handler
	if handler == nil {
		handler = b.defaultHandler
	}

	if handler != nil {
		if err = handler(ctx); err != nil {
			log.Printf("Handler error for UserID %d: %v", ctx.UserID(), err)
			// Optionally send an error message to the user
			// ctx.Reply("An unexpected error occurred. Please try again.")
		}
	} else {
		log.Printf("No handler found for update from UserID %d", ctx.UserID())
	}
}
```

**2. `context.go` - The Interaction Context**

```go
package teleflow

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context provides information and helpers for the current interaction.
type Context struct {
	Bot    *Bot // Reference to the main Bot instance
	Update tgbotapi.Update
	data   map[string]interface{} // For passing data between handlers in a flow
}

// NewContext creates a new Context.
func NewContext(bot *Bot, update tgbotapi.Update) *Context {
	return &Context{
		Bot:    bot,
		Update: update,
		data:   make(map[string]interface{}),
	}
}

// UserID returns the ID of the user who initiated the update.
func (c *Context) UserID() int64 {
	if c.Update.Message != nil {
		return c.Update.Message.From.ID
	}
	if c.Update.CallbackQuery != nil {
		return c.Update.CallbackQuery.From.ID
	}
	// Add other update types (inline query, etc.) if needed
	return 0
}

// ChatID returns the ID of the chat where the update originated.
func (c *Context) ChatID() int64 {
	if c.Update.Message != nil {
		return c.Update.Message.Chat.ID
	}
	if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
		return c.Update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

// Set stores a value in the context's data map (for this request only).
func (c *Context) Set(key string, value interface{}) {
    c.data[key] = value
}

// Get retrieves a value from the context's data map.
func (c *Context) Get(key string) (interface{}, bool) {
    val, ok := c.data[key]
    return val, ok
}

// --- State Management Wrappers ---

// SetState sets the user's current state.
// 'data' can be used to store information relevant to this state.
func (c *Context) SetState(stateName string, data map[string]interface{}) error {
	return c.Bot.stateManager.SetState(c.UserID(), stateName, data)
}

// GetState retrieves the user's current state and associated data.
func (c *Context) GetState() (string, map[string]interface{}, error) {
	return c.Bot.stateManager.GetState(c.UserID())
}

// ClearState clears the user's current state.
func (c *Context) ClearState() error {
	return c.Bot.stateManager.ClearState(c.UserID())
}

// --- Reply Helpers ---

// Reply sends a text message back to the user.
func (c *Context) Reply(text string, keyboardMarkup ...interface{}) error {
	return c.send(text, "", keyboardMarkup...)
}

// ReplyTemplate sends a text message using a pre-loaded template.
func (c *Context) ReplyTemplate(templateName string, data interface{}, keyboardMarkup ...interface{}) error {
	var buf bytes.Buffer
	if err := c.Bot.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return fmt.Errorf("executing template %s: %w", templateName, err)
	}
	return c.send(buf.String(), "", keyboardMarkup...)
}

// EditOrReply attempts to edit the current message if it's a callback, otherwise replies.
func (c *Context) EditOrReply(text string, keyboardMarkup ...interface{}) error {
    if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
        // Try to edit
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
        _, err := c.Bot.api.Send(msg)
        if err == nil {
             // Answer callback query to remove the "loading" state on the button
            cb := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, "")
            _, _ = c.Bot.api.AnswerCallbackQuery(cb)
            return nil
        }
        log.Printf("Failed to edit message, will reply instead: %v", err)
    }
    return c.Reply(text, keyboardMarkup...)
}


// send is an internal helper for sending messages.
func (c *Context) send(text, parseMode string, keyboardMarkup ...interface{}) error {
	if c.ChatID() == 0 {
		return fmt.Errorf("cannot determine chat ID to send message")
	}
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	if len(keyboardMarkup) > 0 && keyboardMarkup[0] != nil {
		switch kb := keyboardMarkup[0].(type) {
		case *ReplyKeyboard:
			msg.ReplyMarkup = kb.toTgbotapi()
		case *InlineKeyboard:
			msg.ReplyMarkup = kb.toTgbotapi()
		case tgbotapi.ReplyKeyboardRemove:
			msg.ReplyMarkup = kb
		default:
			return fmt.Errorf("unsupported keyboard type: %T", kb)
		}
	} else if c.Bot.mainMenu != nil { // Default to main menu if no specific keyboard given
        // Only add default main menu if not in a specific state (or make this configurable)
        currentState, _, _ := c.GetState()
        if currentState == "" {
		    msg.ReplyMarkup = c.Bot.mainMenu.toTgbotapi()
        }
	}


	_, err := c.Bot.api.Send(msg)
	return err
}

// AnswerCallback sends an answer to a callback query (e.g., for inline buttons).
// Useful for showing notifications or alerts.
func (c *Context) AnswerCallback(text string, showAlert bool) error {
    if c.Update.CallbackQuery == nil {
        return fmt.Errorf("not a callback query")
    }
    callback := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, text)
    callback.ShowAlert = showAlert
    _, err := c.Bot.api.AnswerCallbackQuery(callback)
    return err
}

// RawAPI returns the underlying tgbotapi.BotAPI instance for advanced use.
func (c *Context) RawAPI() *tgbotapi.BotAPI {
	return c.Bot.api
}
```

**3. `keyboards.go` - Keyboard Helpers**

```go
package teleflow

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// --- ReplyKeyboard ---
type ReplyKeyboardButton struct {
	Text            string
	RequestContact  bool
	RequestLocation bool
}

type ReplyKeyboardRow []ReplyKeyboardButton

type ReplyKeyboard struct {
	Buttons         [][]ReplyKeyboardButton // Rows of buttons
	ResizeKeyboard  bool
	OneTimeKeyboard bool
	Selective       bool
}

func NewReplyKeyboard(rows ...[]ReplyKeyboardButton) *ReplyKeyboard {
	return &ReplyKeyboard{Buttons: rows, ResizeKeyboard: true}
}

func (rk *ReplyKeyboard) toTgbotapi() tgbotapi.ReplyKeyboardMarkup {
	var keyboard [][]tgbotapi.KeyboardButton
	for _, row := range rk.Buttons {
		var tgRow []tgbotapi.KeyboardButton
		for _, btn := range row {
			tgBtn := tgbotapi.NewKeyboardButton(btn.Text)
			if btn.RequestContact {
				tgBtn.RequestContact = true
			}
			if btn.RequestLocation {
				tgBtn.RequestLocation = true
			}
			tgRow = append(tgRow, tgBtn)
		}
		keyboard = append(keyboard, tgRow)
	}
	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		ResizeKeyboard:  rk.ResizeKeyboard,
		OneTimeKeyboard: rk.OneTimeKeyboard,
		Selective:       rk.Selective,
	}
}

// --- InlineKeyboard ---
type InlineKeyboardButton struct {
	Text         string
	URL          string
	CallbackData string
	WebApp       *WebAppInfo // For launching Web Apps
	// Add other types: LoginURL, SwitchInlineQuery, etc.
}

type WebAppInfo struct {
	URL string
}

type InlineKeyboardRow []InlineKeyboardButton

type InlineKeyboard struct {
	Buttons [][]InlineKeyboardButton
}

func NewInlineKeyboard(rows ...[]InlineKeyboardButton) *InlineKeyboard {
	return &InlineKeyboard{Buttons: rows}
}

func (ik *InlineKeyboard) toTgbotapi() tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, row := range ik.Buttons {
		var tgRow []tgbotapi.InlineKeyboardButton
		for _, btn := range row {
			tgBtn := tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.CallbackData) // Default
			if btn.URL != "" {
				tgBtn = tgbotapi.NewInlineKeyboardButtonURL(btn.Text, btn.URL)
			} else if btn.WebApp != nil {
                tgBtn.WebApp = &tgbotapi.WebAppInfo{URL: btn.WebApp.URL}
                // Text for WebApp button is usually handled by the button itself,
                // but CallbackData might still be useful if WebApp sends data back.
                // For simplicity, we assume Text is for display, WebApp URL is key.
                // tgBtn.Text = btn.Text // Already set by NewInlineKeyboardButtonData if used as base
            }
			// Add other button types
			tgRow = append(tgRow, tgBtn)
		}
		keyboard = append(keyboard, tgRow)
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

// --- MenuButton --- (for the chat menu button)
type MenuButtonConfig struct {
	Type      string // "commands", "web_app", "default"
	Text      string // Required if type is "web_app"
	WebAppURL string // Required if type is "web_app"
}
```

**4. `state.go` - State Management**

```go
package teleflow

import "sync"

// StateManager defines the interface for managing user states.
type StateManager interface {
	SetState(userID int64, stateName string, data map[string]interface{}) error
	GetState(userID int64) (string, map[string]interface{}, error)
	ClearState(userID int64) error
}

// InMemoryStateManager is a simple thread-safe in-memory state manager.
type InMemoryStateManager struct {
	mu     sync.RWMutex
	states map[int64]string
	data   map[int64]map[string]interface{}
}

func NewInMemoryStateManager() *InMemoryStateManager {
	return &InMemoryStateManager{
		states: make(map[int64]string),
		data:   make(map[int64]map[string]interface{}),
	}
}

func (m *InMemoryStateManager) SetState(userID int64, stateName string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[userID] = stateName
	if data != nil {
		m.data[userID] = data
	} else {
		delete(m.data, userID) // Clear data if nil is passed
	}
	return nil
}

func (m *InMemoryStateManager) GetState(userID int64) (string, map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.states[userID]
	if !ok {
		return "", nil, nil // Or an error indicating no state
	}
	userData, _ := m.data[userID]
	return state, userData, nil
}

func (m *InMemoryStateManager) ClearState(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, userID)
	delete(m.data, userID)
	return nil
}
```

**5. `templates.go` (already integrated into `bot.go` and `context.go`)**

The `text/template` package from Go's standard library is used. Templates are added via `bot.AddTemplate()` and rendered via `ctx.ReplyTemplate()`.

**Example Usage (`examples/main.go`):**

```go
package main

import (
	"fmt"
	"log"
	"os"
	"teleflow" // Assuming this is your package path
)

// ---- Handlers ----
func startHandler(ctx *teleflow.Context) error {
	// Using a pre-loaded template
	return ctx.ReplyTemplate("welcome", map[string]string{"UserName": ctx.Update.Message.From.FirstName})
	// The main menu (if set via WithMainMenu) will be shown by default with Reply/ReplyTemplate
}

func helpHandler(ctx *teleflow.Context) error {
	return ctx.Reply("This is the help message. Available commands: /start, /help, /balance, /transfer")
}

func balanceHandler(ctx *teleflow.Context) error {
	// Business logic to get balance
	balance := 100.50
	inlineKb := teleflow.NewInlineKeyboard(
		[]teleflow.InlineKeyboardButton{
			{Text: "View Transactions", CallbackData: "view_tx_history"},
			{Text: "Make Deposit (Web)", WebApp: &teleflow.WebAppInfo{URL: "https://yourdomain.com/deposit"}},
		},
	)
	return ctx.ReplyTemplate("balance", map[string]interface{}{"Amount": balance}, inlineKb)
}

func viewTransactionsHandler(ctx *teleflow.Context) error {
    // Business logic for transactions
    // For simplicity, just answering the callback
    ctx.AnswerCallback("Loading transactions...", false)
	return ctx.EditOrReply("Your last 3 transactions: ... (details here)", nil) // Edit previous message
}

// --- State-based flow for Transfer ---
const (
	StateTransferEnterAmount   = "transfer_enter_amount"
	StateTransferEnterRecepient = "transfer_enter_recipient"
	StateTransferConfirm       = "transfer_confirm"
)

func transferStartHandler(ctx *teleflow.Context) error {
	ctx.SetState(StateTransferEnterAmount, nil)
	return ctx.Reply("Please enter the amount to transfer:", teleflow.ReplyKeyboardRemove{}) // Remove main menu temporarily
}

func transferAmountHandler(ctx *teleflow.Context) error {
	amount := ctx.Update.Message.Text // TODO: Validate this is a number
	stateData := map[string]interface{}{"amount": amount}
	ctx.SetState(StateTransferEnterRecepient, stateData)
	return ctx.Reply(fmt.Sprintf("Amount: %s. Now, please enter the recipient's username or ID.", amount))
}

func transferRecipientHandler(ctx *teleflow.Context) error {
	recipient := ctx.Update.Message.Text
	_, prevStateData, _ := ctx.GetState()
	prevStateData["recipient"] = recipient

	confirmKb := teleflow.NewInlineKeyboard(
		[]teleflow.InlineKeyboardButton{
			{Text: "Confirm Transfer", CallbackData: "transfer_do_confirm"},
			{Text: "Cancel", CallbackData: "transfer_cancel"},
		},
	)
	ctx.SetState(StateTransferConfirm, prevStateData)
	return ctx.Reply(
		fmt.Sprintf("Transfer %s to %s?", prevStateData["amount"], recipient),
		confirmKb,
	)
}

func transferConfirmActionHandler(ctx *teleflow.Context) error {
    // Actual transfer logic here
    _, stateData, _ := ctx.GetState()
    log.Printf("CONFIRMED: Transfer %s to %s for user %d", stateData["amount"], stateData["recipient"], ctx.UserID())

    ctx.AnswerCallback("Transfer successful!", true)
    ctx.ClearState() // Clear state after flow completion
	return ctx.EditOrReply("Transfer successful!", nil) // Edit the confirmation message
}

func transferCancelActionHandler(ctx *teleflow.Context) error {
    ctx.AnswerCallback("Transfer cancelled.", false)
    ctx.ClearState()
	return ctx.EditOrReply("Transfer cancelled.", nil)
}


func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	// Define Main Menu (ReplyKeyboard)
	mainMenu := teleflow.NewReplyKeyboard(
		[]teleflow.ReplyKeyboardButton{{Text: "💰 Balance"}, {Text: "💸 Transfer"}},
		[]teleflow.ReplyKeyboardButton{{Text: "❓ Help"}},
	)
    mainMenu.ResizeKeyboard = true


    // Define Bot Menu Button (optional)
    botMenuCfg := &teleflow.MenuButtonConfig{
        Type: "web_app", // or "commands" or "default"
        Text: "Open App",
        WebAppURL: "https://yourwebapp.com/main",
    }

	bot, err := teleflow.NewBot(token,
		teleflow.WithMainMenu(mainMenu),
        // teleflow.WithMenuButton(botMenuCfg), // Uncomment to use bot menu button
    )
	if err != nil {
		log.Fatal(err)
	}

	// Add Message Templates
	bot.AddTemplate("welcome", "Hello {{.UserName}}! Welcome to our bot.")
	bot.AddTemplate("balance", "Your current balance is: {{.Amount}}.")

	// Register Handlers
	bot.HandleCommand("start", startHandler)
	bot.HandleCommand("help", helpHandler)

	bot.HandleText("💰 Balance", balanceHandler) // From ReplyKeyboard
	bot.HandleText("💸 Transfer", transferStartHandler)
	bot.HandleText("❓ Help", helpHandler)

	bot.HandleCallback("view_tx_history", viewTransactionsHandler)
    bot.HandleCallback("transfer_do_confirm", transferConfirmActionHandler)
    bot.HandleCallback("transfer_cancel", transferCancelActionHandler)


	// Register State Handlers for the transfer flow
    // These are triggered by general text input when in a specific state
	bot.OnStateDefault(StateTransferEnterAmount, transferAmountHandler)
	bot.OnStateDefault(StateTransferEnterRecepient, transferRecipientHandler)
    // For StateTransferConfirm, we expect callbacks, not text, so those are handled by HandleCallback

	bot.SetDefaultHandler(func(ctx *teleflow.Context) error {
		return ctx.Reply("Sorry, I didn't understand that. Use /help for commands.")
	})

	log.Println("Bot is starting...")
	bot.Start()
}
```

**Key Features of this Design:**

*   **Intuitive API:** `bot.HandleCommand()`, `bot.HandleText()`, `ctx.Reply()`, `ctx.SetState()`.
*   **Focus on Business Logic:** Developers write simple handler functions.
*   **State Management:** `ctx.SetState()`, `ctx.GetState()`, `bot.OnState()` make multi-step interactions manageable.
*   **Main Menus:**
    *   `teleflow.WithMainMenu()` for a persistent `ReplyKeyboardMarkup`.
    *   `teleflow.WithMenuButton()` for the dedicated bot menu button (can launch a command or Web App).
*   **Conditional Routing (Inline Keyboards):** `teleflow.NewInlineKeyboard()` and `bot.HandleCallback()` allow for branching within a function's response.
*   **Message Templating:** `bot.AddTemplate()` and `ctx.ReplyTemplate()` keep message strings out of business logic code.
*   **Clean Code:** The structure promotes separation of concerns.
*   **Flat Learning Curve:** Starts simple, but allows for more complex stateful interactions.

**Further Enhancements (Beyond Initial Design):**

*   **Flow/Wizard DSL:** A more structured way to define multi-step flows than just state handlers (e.g., `bot.NewFlow("transfer").Step(askAmount).Step(askRecipient).OnConfirm(doTransfer)`).
*   **Middleware:** For logging, authentication, rate limiting before handlers.
*   **More Sophisticated Callback Matching:** Beyond simple prefix.
*   **Internationalization (i18n):** Built-in support for multi-language bots.
*   **Persistence for State:** Allow plugging in different state backends (Redis, database) instead of just in-memory.
*   **Error Handling Strategies:** More granular control over how errors are reported to users.
*   **File Handling:** Helpers for receiving and sending files.

This design provides a solid foundation. The key is to start with these core elements and then iteratively build upon them based on feedback and common use cases.