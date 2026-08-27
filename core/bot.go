package teleflow

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// registryTTL is how long a registered callback or data entry lives before
// being eligible for eviction. Entries are cleaned up lazily on access and
// periodically by a background sweeper.
const registryTTL = 24 * time.Hour

// callbackEntry wraps a callback handler with an expiry timestamp.
type callbackEntry struct {
	handler   func(ctx *Context) error
	expiresAt time.Time
}

// dataEntry wraps a data value with an expiry timestamp.
type dataEntry struct {
	value     interface{}
	expiresAt time.Time
}

// Handler defines the interface for handling Telegram updates.
type Handler interface {
	HandleUpdate(ctx *Context) error
}

// Bot represents the main bot instance. It listens for updates from Telegram,
// manages callback and data registries, and dispatches updates to the appropriate
// handler.
type Bot struct {
	api              *tgbotapi.BotAPI
	self             tgbotapi.User
	handler          Handler
	callbackRegistry map[string]callbackEntry
	dataRegistry     map[string]dataEntry
	mu               sync.Mutex
}

// NewBot creates and initializes a new Bot instance. It requires a valid Telegram
// bot token and a root Handler (typically a Router) to which all non-callback
// updates will be dispatched. It returns an error if the token is invalid or
// the bot's details cannot be fetched.
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
		callbackRegistry: make(map[string]callbackEntry),
		dataRegistry:     make(map[string]dataEntry),
	}, nil
}

// processUpdate checks the callback registry first, then passes the update to the root handler.
func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := NewContext(update, b.api, b)

	// Check if this is a callback query with a registered handler
	if update.CallbackQuery != nil {
		b.mu.Lock()
		callbackID := update.CallbackQuery.Data
		entry, ok := b.callbackRegistry[callbackID]
		if ok && time.Now().Before(entry.expiresAt) {
			// Handler found, remove it for one-time use and execute
			delete(b.callbackRegistry, callbackID)
			b.mu.Unlock()
			if err := entry.handler(ctx); err != nil {
				log.Printf("Callback handler error: %v", err)
			}
			return // Stop further processing
		} else if ok {
			// Entry expired, remove it
			delete(b.callbackRegistry, callbackID)
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
	b.callbackRegistry[id] = callbackEntry{handler: handler, expiresAt: time.Now().Add(registryTTL)}
	return id
}

// RegisterData stores an arbitrary data object for one-time retrieval via a
// unique ID. This is the core mechanism for passing complex data through
// callback buttons, overcoming Telegram's 64-byte callback data limit. The
// returned ID should be used as the callback data for a button. The data is
// automatically deleted after it is retrieved once.
func (b *Bot) RegisterData(data interface{}) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := uuid.New().String()
	b.dataRegistry[id] = dataEntry{value: data, expiresAt: time.Now().Add(registryTTL)}
	return id
}

// GetData retrieves and immediately deletes a data object from the registry using
// the unique ID provided by a callback query. It returns the data and a boolean
// indicating if the data was found.
func (b *Bot) GetData(id string) (interface{}, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry, ok := b.dataRegistry[id]
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			delete(b.dataRegistry, id)
		}
		return nil, false
	}
	delete(b.dataRegistry, id)
	return entry.value, true
}

// Start begins the bot's main event loop. It connects to Telegram and starts
// polling for updates. Each update is processed in a separate goroutine.
func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.self.UserName)

	// Start background sweeper for registry eviction.
	go b.sweepRegistries()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.processUpdate(update)
	}

	return nil
}

// sweepRegistries periodically removes expired entries from both registries.
func (b *Bot) sweepRegistries() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		b.mu.Lock()
		now := time.Now()
		for id, entry := range b.callbackRegistry {
			if now.After(entry.expiresAt) {
				delete(b.callbackRegistry, id)
			}
		}
		for id, entry := range b.dataRegistry {
			if now.After(entry.expiresAt) {
				delete(b.dataRegistry, id)
			}
		}
		b.mu.Unlock()
	}
}

// SetBotCommands sets the list of commands that will be displayed to the user
// in the Telegram interface. This should be called once after the bot is created.
func (b *Bot) SetBotCommands(commands map[string]string) error {
	var tgCommands []tgbotapi.BotCommand
	keys := make([]string, 0, len(commands))
	for cmd := range commands {
		keys = append(keys, cmd)
	}
	sort.Strings(keys)
	for _, cmd := range keys {
		tgCommands = append(tgCommands, tgbotapi.BotCommand{Command: cmd, Description: commands[cmd]})
	}

	config := tgbotapi.NewSetMyCommands(tgCommands...)
	_, err := b.api.Request(config)
	if err != nil {
		return fmt.Errorf("failed to set bot commands: %w", err)
	}
	return nil
}
