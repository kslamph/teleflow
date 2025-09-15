package teleflow

import (
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// Handler defines the interface for handling Telegram updates.
type Handler interface {
	HandleUpdate(ctx *Context) error
}

// Bot represents the main bot instance. It listens for updates from Telegram
// and dispatches them to a registered handler or a callback from the registry.
type Bot struct {
	api              *tgbotapi.BotAPI
	self             tgbotapi.User
	handler          Handler
	callbackRegistry map[string]func(ctx *Context) error
	mu               sync.Mutex
}

// NewBot creates a new Bot instance. It requires a Telegram bot token and a root Handler.
func NewBot(token string, handler Handler) (*Bot, error) {
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	botUser, err := api.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	return &Bot{
		api:              api,
		self:             botUser,
		handler:          handler,
		callbackRegistry: make(map[string]func(ctx *Context) error),
	}, nil
}

// processUpdate checks the callback registry first, then passes the update to the root handler.
func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := NewContext(update, b.api, b)

	// Check if this is a callback query with a registered handler
	if update.CallbackQuery != nil {
		b.mu.Lock()
		callbackID := update.CallbackQuery.Data
		handler, ok := b.callbackRegistry[callbackID]
		if ok {
			// Handler found, remove it for one-time use and execute
			delete(b.callbackRegistry, callbackID)
			b.mu.Unlock()
			if err := handler(ctx); err != nil {
				log.Printf("Callback handler error: %v", err)
			}
			return // Stop further processing
		}
		b.mu.Unlock()
	}

	// No specific callback found, proceed with the main handler
	if err := b.handler.HandleUpdate(ctx); err != nil {
		log.Printf("Handler error: %v", err)
	}
}

// registerCallback adds a handler to the registry and returns its unique ID.
func (b *Bot) registerCallback(handler func(ctx *Context) error) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := uuid.New().String()
	b.callbackRegistry[id] = handler
	return id
}

// Start begins the bot's main event loop for polling updates from Telegram.
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

// SetBotCommands sets the list of commands for the bot.
func (b *Bot) SetBotCommands(commands map[string]string) error {
	var tgCommands []tgbotapi.BotCommand
	for cmd, desc := range commands {
		tgCommands = append(tgCommands, tgbotapi.BotCommand{Command: cmd, Description: desc})
	}

	config := tgbotapi.NewSetMyCommands(tgCommands...)
	_, err := b.api.Request(config)
	if err != nil {
		return fmt.Errorf("failed to set bot commands: %w", err)
	}
	return nil
}
