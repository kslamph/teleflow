package teleflow

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// TelegramClient defines an interface for interacting with the Telegram Bot API.
// It abstracts the underlying Telegram client, allowing for mock implementations
// for testing and decoupling components from a concrete API client.
type TelegramClient interface {
	// Send sends a Chattable object (e.g., MessageConfig, PhotoConfig) to Telegram.
	// It returns the sent Message object on success, or an error.
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)

	// Request makes a direct request to the Telegram Bot API.
	// It is used for less common API calls that might not have dedicated methods.
	// It returns an APIResponse object or an error.
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)

	// GetUpdatesChan returns a channel for receiving updates from Telegram.
	// The config parameter specifies how updates should be fetched.
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel

	// GetMe fetches the bot's own user information.
	// It returns a User object representing the bot, or an error.
	GetMe() (tgbotapi.User, error)
}
