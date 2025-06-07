// Menu Button System provides structures for defining menu button configurations.
// The menu button appears next to the text input field in Telegram chats.
//
// Menu Button Types:
//   - Commands: Used in conjunction with `Bot.SetBotCommands` to define
//               the commands that appear in Telegram's command menu.
//   - Default:  Instructs Telegram to use its default menu button behavior.
//
// Basic Usage:
//
//	// To define commands for the menu button (used with Bot.SetBotCommands):
//	commandsConfig := teleflow.NewCommandsMenuButton().
//		AddCommand("Help", "/help").
//		AddCommand("Settings", "/settings")
//	// This config would then be processed by Bot.SetBotCommands.
//
//	// To set the menu button to Telegram's default:
//	defaultConfig := teleflow.NewDefaultMenuButton()
//	// This config would be set on the Bot, and the Bot would make the
//	// appropriate API call to Telegram.
//
// The actual API calls to Telegram to set the menu button (e.g., SetMyCommands
// or SetChatMenuButton) are handled by methods on the Bot struct, not directly here.
// This file primarily provides the data structures and helpers for these configurations.
package teleflow

// menuButtonType represents the type of native menu button
type menuButtonType string

const (
	// menuButtonTypeCommands indicates the menu button should display bot commands.
	// The actual commands are defined in the Items field of MenuButtonConfig.
	menuButtonTypeCommands menuButtonType = "commands"

	// menuButtonTypeDefault indicates Telegram's default menu button behavior should be used.
	menuButtonTypeDefault menuButtonType = "default"
)

// menuButtonItem represents a command item for menu buttons.
// This is used internally when MenuButtonConfig.Type is menuButtonTypeCommands.
type menuButtonItem struct {
	text    string
	command string
}

// MenuButtonConfig represents the configuration for Telegram's native menu button.
// It's used to define either a list of commands or to specify the default menu button.
type MenuButtonConfig struct {
	// Type specifies the type of menu button.
	Type menuButtonType `json:"type"`
	// Items is a list of commands, relevant only if Type is menuButtonTypeCommands.
	// This list is typically processed by `Bot.SetBotCommands`.
	Items []menuButtonItem `json:"items,omitempty"`
}

// NewCommandsMenuButton creates a menu button configuration intended for defining bot commands.
// The actual commands are added using the AddCommand method.
// This configuration is typically used with `Bot.SetBotCommands`.
func NewCommandsMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type:  menuButtonTypeCommands,
		Items: make([]menuButtonItem, 0), // Initialize Items
	}
}

// NewDefaultMenuButton creates a menu button configuration that signifies
// Telegram's default menu button should be used.
// When this configuration is applied to a Bot, the Bot should instruct Telegram
// to display its standard menu button.
func NewDefaultMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: menuButtonTypeDefault,
	}
}

// AddCommand adds a command to a commands-type menu button configuration.
// It is a no-op if the MenuButtonConfig is not of type menuButtonTypeCommands.
// This method allows for fluent construction of command lists.
//
// Example:
//
//	config := NewCommandsMenuButton().
//		AddCommand("Start", "/start").
//		AddCommand("Help", "/help")
func (config *MenuButtonConfig) AddCommand(text, command string) *MenuButtonConfig {
	if config.Type != menuButtonTypeCommands {
		// Only add commands if the type is explicitly 'commands'.
		// This prevents accidental modification of a 'default' type config.
		return config
	}

	config.Items = append(config.Items, menuButtonItem{
		text:    text,
		command: command,
	})

	return config
}