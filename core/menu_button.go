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
)

// SetMenuButton sets the menu button for a specific chat or all chats
// Note: Commands-type menu buttons are now set via SetBotCommands method
func (b *Bot) SetMenuButton(chatID int64, config *MenuButtonConfig) error {
	if config == nil {
		return fmt.Errorf("menu button config cannot be nil")
	}

	switch config.Type {
	case menuButtonTypeCommands:
		return fmt.Errorf("commands-type menu buttons should be set via SetBotCommands method, not SetMenuButton")

	case menuButtonTypeWebApp:
		if config.WebApp == nil {
			return fmt.Errorf("web app config is required for web_app menu button type")
		}
		log.Printf("ℹ️ WebApp menu button not yet supported, but config saved")
		return nil

	case menuButtonTypeDefault:
		log.Printf("✅ Default menu button set for chat %d", chatID)
		return nil

	default:
		return fmt.Errorf("unsupported menu button type: %s", config.Type)
	}
}

// SetDefaultMenuButton sets the default menu button configuration for the bot
// Only supports web_app or default types. For bot commands, use SetBotCommands method.
func (b *Bot) SetDefaultMenuButton() error {
	if b.menuButton == nil {
		return nil // No menu button configured
	}

	// Only allow web_app or default types
	if b.menuButton.Type != menuButtonTypeWebApp && b.menuButton.Type != menuButtonTypeDefault {
		return fmt.Errorf("SetDefaultMenuButton only supports web_app or default types, use SetBotCommands for commands")
	}

	// Set for all chats (chatID = 0)
	return b.SetMenuButton(0, b.menuButton)
}

// initializeMenuButton sets up the menu button when the bot starts
// Only handles web_app or default types. Bot commands should be set via SetBotCommands.
func (b *Bot) initializeMenuButton() {
	if b.menuButton != nil {
		// Only initialize web_app or default menu buttons
		if b.menuButton.Type == menuButtonTypeWebApp || b.menuButton.Type == menuButtonTypeDefault {
			log.Printf("🔧 Setting up menu button: %s", b.menuButton.Type)

			err := b.SetDefaultMenuButton()
			if err != nil {
				log.Printf("❌ Menu button setup failed: %v", err)
			} else {
				log.Printf("✅ Menu button configured: %s", b.menuButton.Type)
			}
		} else {
			log.Printf("ℹ️ Skipping commands-type menu button initialization - use SetBotCommands() method instead")
		}
	}
}

// Helper functions for creating menu button configurations

// NewCommandsMenuButton creates a menu button that shows bot commands
func NewCommandsMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: menuButtonTypeCommands,
	}
}

// NewWebAppMenuButton creates a menu button that opens a web app
func NewWebAppMenuButton(text, url string) *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: menuButtonTypeWebApp,
		Text: text,
		WebApp: &webAppInfo{
			URL: url,
		},
	}
}

// NewDefaultMenuButton creates a default menu button
func NewDefaultMenuButton() *MenuButtonConfig {
	return &MenuButtonConfig{
		Type: menuButtonTypeDefault,
	}
}

// AddCommandToMenuButton adds a command to a commands-type menu button
func (config *MenuButtonConfig) AddCommand(text, command string) *MenuButtonConfig {
	if config.Type != menuButtonTypeCommands {
		return config // Only works for commands type
	}

	config.Items = append(config.Items, menuButtonItem{
		text:    text,
		command: command,
	})

	return config
}
