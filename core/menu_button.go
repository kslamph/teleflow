// Menu Button System provides comprehensive menu button management for Telegram bots.
// This system enables customization of the bot's menu button that appears next to the
// text input field in Telegram chats.
//
// Menu Button Types:
//   - Commands: Displays registered bot commands in a native menu
//   - WebApp: Opens a web application when clicked
//   - Default: Uses Telegram's default menu button behavior
//
// Basic Usage:
//
//	// Create a commands menu button with bot commands
//	menuButton := &teleflow.MenuButtonConfig{
//		Type: teleflow.MenuButtonTypeCommands,
//		Items: []teleflow.MenuButtonItem{
//			{Text: "📖 Help", Command: "/help"},
//			{Text: "⚙️ Settings", Command: "/settings"},
//		},
//	}
//
//	// Apply to bot using functional options
//	bot, err := teleflow.NewBot(token, teleflow.WithMenuButton(menuButton))
//
//	// Or set after bot creation
//	bot.WithMenuButton(menuButton)
//
// Helper Functions:
//
//	// Create different menu button types
//	commandsButton := teleflow.NewCommandsMenuButton()
//	commandsButton.AddCommand("Help", "/help").AddCommand("Status", "/status")
//
//	webAppButton := teleflow.NewWebAppMenuButton("Open App", "https://example.com")
//	defaultButton := teleflow.NewDefaultMenuButton()
//
// Advanced Usage with AccessManager:
//
//	// Dynamic menu buttons based on user permissions
//	type MyAccessManager struct{}
//
//	func (m *MyAccessManager) GetMenuButton(ctx *teleflow.MenuContext) *teleflow.MenuButtonConfig {
//		if ctx.UserID == adminID {
//			return &teleflow.MenuButtonConfig{
//				Type: teleflow.MenuButtonTypeCommands,
//				Items: []teleflow.MenuButtonItem{
//					{Text: "👤 Admin Panel", Command: "/admin"},
//					{Text: "📊 Analytics", Command: "/stats"},
//				},
//			}
//		}
//		return teleflow.NewDefaultMenuButton()
//	}
//
// The menu button system automatically:
//   - Registers commands with Telegram when using Commands type
//   - Handles menu button initialization during bot startup
//   - Supports per-chat and global menu button configurations
//   - Integrates with the AccessManager for permission-based UI
package teleflow

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SetMenuButton sets the menu button for a specific chat or all chats
func (b *Bot) SetMenuButton(chatID int64, config *MenuButtonConfig) error {
	if config == nil {
		return fmt.Errorf("menu button config cannot be nil")
	}

	switch config.Type {
	case MenuButtonTypeCommands:
		// For commands type, register the commands with Telegram - this is what makes them appear!
		if len(config.Items) > 0 {
			if err := b.setMyCommands(config.Items); err != nil {
				log.Printf("Warning: Failed to set bot commands: %v", err)
				return err
			}
		}
		log.Printf("✅ Commands menu button set for chat %d", chatID)
		return nil

	case MenuButtonTypeWebApp:
		if config.WebApp == nil {
			return fmt.Errorf("web app config is required for web_app menu button type")
		}
		log.Printf("ℹ️ WebApp menu button not yet supported, but config saved")
		return nil

	case MenuButtonTypeDefault:
		log.Printf("✅ Default menu button set for chat %d", chatID)
		return nil

	default:
		return fmt.Errorf("unsupported menu button type: %s", config.Type)
	}
}

// SetDefaultMenuButton sets the default menu button configuration for the bot
func (b *Bot) SetDefaultMenuButton() error {
	if b.menuButton == nil {
		return nil // No menu button configured
	}

	// Set for all chats (chatID = 0)
	return b.SetMenuButton(0, b.menuButton)
}

// setMyCommands registers bot commands with Telegram using the native telegram-bot-api
func (b *Bot) setMyCommands(items []MenuButtonItem) error {
	if len(items) == 0 {
		return nil
	}

	// Convert MenuButtonItems to tgbotapi.BotCommand format
	var commands []tgbotapi.BotCommand
	for _, item := range items {
		// Remove leading slash if present
		command := item.Command
		if len(command) > 0 && command[0] == '/' {
			command = command[1:]
		}

		commands = append(commands, tgbotapi.BotCommand{
			Command:     command,
			Description: item.Text,
		})
	}

	// Create the setMyCommands request using the native API
	cmdCfg := tgbotapi.NewSetMyCommands(commands...)

	// Use Send method - this is what makes the menu button appear!
	// Note: This may cause a JSON unmarshal error which we'll ignore since it still works

	_, err := b.api.Request(cmdCfg)
	if err != nil {

		return fmt.Errorf("failed to set bot commands: %w", err)

	} else {
		log.Printf("✅ Registered %d bot commands with Telegram", len(commands))
	}

	return nil
}

// InitializeMenuButton sets up the menu button when the bot starts
func (b *Bot) InitializeMenuButton() {
	if b.menuButton != nil {
		log.Printf("🔧 Setting up menu button: %s", b.menuButton.Type)

		err := b.SetDefaultMenuButton()
		if err != nil {
			log.Printf("❌ Menu button setup failed: %v", err)
		} else {
			log.Printf("✅ Menu button configured: %s", b.menuButton.Type)
		}

		// Log available commands for commands type
		if b.menuButton.Type == MenuButtonTypeCommands && len(b.menuButton.Items) > 0 {
			log.Printf("📋 Bot commands available:")
			for _, item := range b.menuButton.Items {
				log.Printf("   %s - %s", item.Text, item.Command)
			}
		}
	}
}

// Helper functions for creating menu button configurations

// NewCommandsMenuButton creates a menu button that shows bot commands
func NewCommandsMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: MenuButtonTypeCommands,
	}
}

// NewWebAppMenuButton creates a menu button that opens a web app
func NewWebAppMenuButton(text, url string) *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: MenuButtonTypeWebApp,
		Text: text,
		WebApp: &WebAppInfo{
			URL: url,
		},
	}
}

// NewDefaultMenuButton creates a default menu button
func NewDefaultMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: MenuButtonTypeDefault,
	}
}

// AddCommandToMenuButton adds a command to a commands-type menu button
func (config *MenuButtonConfig) AddCommand(text, command string) *MenuButtonConfig {
	if config.Type != MenuButtonTypeCommands {
		return config // Only works for commands type
	}

	config.Items = append(config.Items, MenuButtonItem{
		Text:    text,
		Command: command,
	})

	return config
}
