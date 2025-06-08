// Package teleflow menu button system provides data structures and utilities for
// configuring Telegram's native menu button that appears next to the text input field.
//
// This system works in close integration with Bot.SetBotCommands() method to provide
// a unified approach to command management. The menu button can display either:
//   - Bot commands (configured via Bot.SetBotCommands())
//   - Telegram's default menu button behavior
//
// Relationship with Bot.SetBotCommands():
// The MenuButtonConfig structures defined here are primarily used as data containers
// and builders for command configurations. The actual API calls to Telegram
// (SetMyCommands, SetChatMenuButton) are handled by Bot.SetBotCommands() method,
// which processes the command data and makes appropriate Telegram API calls.
//
// Basic Usage:
//
//	// Method 1: Direct command mapping (recommended)
//	err := bot.SetBotCommands(map[string]string{
//		"start":    "Start the bot",
//		"help":     "Show help information",
//		"settings": "Open settings menu",
//	})
//
//	// Method 2: Using MenuButtonConfig builders (alternative approach)
//	commandsConfig := teleflow.NewCommandsMenuButton().
//		AddCommand("Help", "/help").
//		AddCommand("Settings", "/settings")
//	// Process this config through Bot.SetBotCommands() or similar method
//
//	// Method 3: Default menu button
//	defaultConfig := teleflow.NewDefaultMenuButton()
//	// Configure bot to use Telegram's default menu button
//
// Architecture:
// This file provides the data structures and builder patterns, while the actual
// Telegram API integration is handled by the Bot struct methods. This separation
// allows for flexible command configuration while maintaining clean API boundaries.
package teleflow

// menuButtonType represents the type of Telegram's native menu button.
// This determines how the menu button behaves when users interact with it.
type menuButtonType string

const (
	// MenuButtonTypeCommands indicates the menu button should display bot commands.
	// When this type is used, the menu button will show a list of commands that
	// can be configured through Bot.SetBotCommands() method. The commands are
	// defined in the Items field of MenuButtonConfig and processed by the Bot
	// to make the appropriate SetMyCommands API call to Telegram.
	//
	// This type is used internally by NewCommandsMenuButton() constructor.
	menuButtonTypeCommands menuButtonType = "commands"

	// MenuButtonTypeDefault indicates Telegram's default menu button behavior should be used.
	// When this type is used, Telegram will display its standard menu button,
	// which typically shows options like "Menu" or other default behaviors
	// determined by Telegram's client implementation.
	//
	// This type is used internally by NewDefaultMenuButton() constructor.
	menuButtonTypeDefault menuButtonType = "default"
)

// menuButtonItem represents a single command entry for menu buttons.
// This internal structure stores the display text and command string for
// individual bot commands that will be processed by Bot.SetBotCommands().
//
// Fields:
//   - text: The human-readable description that appears in the command menu
//   - command: The actual command string (typically starting with "/")
//
// This struct is used internally when MenuButtonConfig.Type is menuButtonTypeCommands
// and is populated through the AddCommand() method on MenuButtonConfig.
type menuButtonItem struct {
	text    string // Display text for the command (e.g., "Show help information")
	command string // Command string (e.g., "/help")
}

// MenuButtonConfig represents the configuration for Telegram's native menu button.
// This struct serves as a data container and builder for menu button configurations
// that work in conjunction with Bot.SetBotCommands() method.
//
// The struct supports two primary use cases:
//  1. Commands menu button: Contains a list of bot commands with descriptions
//  2. Default menu button: Instructs Telegram to use its default behavior
//
// Integration with Bot.SetBotCommands():
// When Type is menuButtonTypeCommands, the Items field contains command data
// that can be processed by Bot.SetBotCommands() to make the appropriate
// SetMyCommands API call to Telegram. The bot method handles the conversion
// from this configuration format to Telegram's API format.
//
// Usage patterns:
//   - Use NewCommandsMenuButton() to create a commands-type configuration
//   - Use NewDefaultMenuButton() to create a default-type configuration
//   - Chain AddCommand() calls to build command lists fluently
type MenuButtonConfig struct {
	// Type specifies the menu button behavior type.
	// Valid values: menuButtonTypeCommands, menuButtonTypeDefault
	Type menuButtonType `json:"type"`

	// Items contains the list of commands for commands-type menu buttons.
	// This field is only relevant when Type is menuButtonTypeCommands.
	// Each item contains a display text and command string that will be
	// processed by Bot.SetBotCommands() method to configure Telegram's command menu.
	// For default-type menu buttons, this field should be empty.
	Items []menuButtonItem `json:"items,omitempty"`
}

// NewCommandsMenuButton creates a new MenuButtonConfig for bot commands.
// This constructor initializes a configuration that can hold multiple command
// definitions, which are then processed by Bot.SetBotCommands() to configure
// Telegram's command menu that appears when users type "/".
//
// The returned configuration starts empty and commands should be added using
// the AddCommand() method, which supports fluent chaining for easy construction.
//
// Relationship with Bot.SetBotCommands():
// The commands added to this configuration are intended to be processed by
// Bot.SetBotCommands(), which extracts the command data and makes the
// appropriate SetMyCommands API call to Telegram.
//
// Returns:
//   - *MenuButtonConfig: A new configuration with Type set to menuButtonTypeCommands
//     and an empty Items slice ready for command additions.
//
// Example:
//
//	config := teleflow.NewCommandsMenuButton().
//		AddCommand("Start bot", "/start").
//		AddCommand("Get help", "/help").
//		AddCommand("Settings", "/settings")
//
//	// Process with Bot.SetBotCommands() or similar method
//	commands := make(map[string]string)
//	for _, item := range config.Items {
//		commands[strings.TrimPrefix(item.command, "/")] = item.text
//	}
//	err := bot.SetBotCommands(commands)
func NewCommandsMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type:  menuButtonTypeCommands,
		Items: make([]menuButtonItem, 0),
	}
}

// NewDefaultMenuButton creates a new MenuButtonConfig for Telegram's default menu button.
// This constructor creates a configuration that instructs Telegram to use its
// standard menu button behavior instead of displaying custom bot commands.
//
// When this configuration is processed by the bot system, it typically results
// in a SetChatMenuButton API call to Telegram with a "default" type, which
// tells Telegram to display its standard menu button interface.
//
// Use cases:
//   - Reverting from custom commands back to default behavior
//   - Simplifying the user interface for bots that don't need custom commands
//   - Providing a consistent experience across different bot implementations
//
// Integration with Bot methods:
// While this configuration can be used with various bot methods, the actual
// API call to set the default menu button is handled by Bot methods, not
// directly by this configuration structure.
//
// Returns:
//   - *MenuButtonConfig: A new configuration with Type set to menuButtonTypeDefault
//     and no Items (as default menu buttons don't require command definitions).
//
// Example:
//
//	defaultConfig := teleflow.NewDefaultMenuButton()
//	// Process this configuration through appropriate Bot method
//	// to set Telegram's default menu button behavior
func NewDefaultMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: menuButtonTypeDefault,
	}
}

// AddCommand adds a command entry to a commands-type menu button configuration.
// This method enables fluent construction of command lists that will be processed
// by Bot.SetBotCommands() to configure Telegram's command menu.
//
// The method only operates on MenuButtonConfig instances with Type set to
// menuButtonTypeCommands. If called on a default-type configuration, it performs
// no operation and returns the config unchanged, preventing accidental modification.
//
// Parameters:
//   - text: Human-readable description of the command that appears in the menu
//   - command: The actual command string, typically starting with "/" (e.g., "/start")
//
// Returns:
//   - *MenuButtonConfig: The same configuration instance to enable method chaining
//
// Integration with Bot.SetBotCommands():
// Commands added through this method are stored in the Items field and can be
// processed by Bot.SetBotCommands(), which converts them to the appropriate
// format for Telegram's SetMyCommands API call.
//
// Examples:
//
//	// Basic usage with method chaining
//	config := teleflow.NewCommandsMenuButton().
//		AddCommand("Start the bot", "/start").
//		AddCommand("Show help information", "/help").
//		AddCommand("Open settings menu", "/settings")
//
//	// Safe to call on default-type config (no-op)
//	defaultConfig := teleflow.NewDefaultMenuButton().
//		AddCommand("This won't be added", "/ignored") // No effect
//
//	// Processing commands for Bot.SetBotCommands()
//	commands := make(map[string]string)
//	for _, item := range config.Items {
//		cmdName := strings.TrimPrefix(item.command, "/")
//		commands[cmdName] = item.text
//	}
//	err := bot.SetBotCommands(commands)
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
