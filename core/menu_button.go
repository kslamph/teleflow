package teleflow

import (
	"encoding/json"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SetMenuButton sets the menu button for a specific chat or all chats
func (b *Bot) SetMenuButton(chatID int64, config *MenuButtonConfig) error {
	if config == nil {
		return fmt.Errorf("menu button config cannot be nil")
	}

	log.Printf("Setting menu button for chat %d: %+v", chatID, config)

	var menuButton interface{}

	switch config.Type {
	case MenuButtonTypeCommands:
		// For commands type, first register the commands with Telegram
		if len(config.Items) > 0 {
			if err := b.setMyCommands(config.Items); err != nil {
				log.Printf("Warning: Failed to set bot commands: %v", err)
				// Continue anyway - the menu button might still work
			}
		}

		// Then set the menu button to show commands
		menuButton = map[string]string{
			"type": "commands",
		}
	case MenuButtonTypeWebApp:
		if config.WebApp == nil {
			return fmt.Errorf("web app config is required for web_app menu button type")
		}
		menuButton = map[string]interface{}{
			"type":    "web_app",
			"text":    config.Text,
			"web_app": config.WebApp,
		}
	case MenuButtonTypeDefault:
		menuButton = map[string]string{
			"type": "default",
		}
	default:
		return fmt.Errorf("unsupported menu button type: %s", config.Type)
	}

	// Create the request parameters
	params := map[string]interface{}{
		"menu_button": menuButton,
	}

	// Add chat_id if specified (if 0, applies to all chats)
	if chatID != 0 {
		params["chat_id"] = chatID
	}

	// Make the API call
	err := b.makeMenuButtonAPICall("setChatMenuButton", params)
	if err != nil {
		log.Printf("Failed to set menu button: %v", err)
		return err
	}

	log.Printf("✅ Menu button set successfully for chat %d", chatID)
	return nil
}

// GetMenuButton gets the current menu button for a specific chat
func (b *Bot) GetMenuButton(chatID int64) (*MenuButtonConfig, error) {
	params := map[string]interface{}{}

	// Add chat_id if specified (if 0, gets default)
	if chatID != 0 {
		params["chat_id"] = chatID
	}

	response, err := b.makeMenuButtonAPICallWithResponse("getChatMenuButton", params)
	if err != nil {
		return nil, err
	}

	var menuButton MenuButtonConfig
	if err := json.Unmarshal(response, &menuButton); err != nil {
		return nil, fmt.Errorf("failed to parse menu button response: %w", err)
	}

	return &menuButton, nil
}

// SetDefaultMenuButton sets the default menu button configuration for the bot
func (b *Bot) SetDefaultMenuButton() error {
	if b.menuButton == nil {
		return nil // No menu button configured
	}

	// Set for all chats (chatID = 0)
	return b.SetMenuButton(0, b.menuButton)
}

// makeMenuButtonAPICall makes a custom API call for menu button operations
func (b *Bot) makeMenuButtonAPICall(method string, params map[string]interface{}) error {
	// Convert params to the format expected by telegram-bot-api
	tgParams := make(tgbotapi.Params)
	for key, value := range params {
		switch v := value.(type) {
		case string:
			tgParams[key] = v
		case int64:
			tgParams[key] = fmt.Sprintf("%d", v)
		default:
			// For complex objects, marshal to JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal %s: %w", key, err)
			}
			tgParams[key] = string(jsonBytes)
		}
	}

	_, err := b.api.MakeRequest(method, tgParams)
	if err != nil {
		return fmt.Errorf("menu button API call failed: %w", err)
	}

	return nil
}

// makeMenuButtonAPICallWithResponse makes a custom API call and returns the response
func (b *Bot) makeMenuButtonAPICallWithResponse(method string, params map[string]interface{}) ([]byte, error) {
	// Convert params to the format expected by telegram-bot-api
	tgParams := make(tgbotapi.Params)
	for key, value := range params {
		switch v := value.(type) {
		case string:
			tgParams[key] = v
		case int64:
			tgParams[key] = fmt.Sprintf("%d", v)
		default:
			// For complex objects, marshal to JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal %s: %w", key, err)
			}
			tgParams[key] = string(jsonBytes)
		}
	}

	response, err := b.api.MakeRequest(method, tgParams)
	if err != nil {
		return nil, fmt.Errorf("menu button API call failed: %w", err)
	}

	return response.Result, nil
}

// setMyCommands registers bot commands with Telegram
func (b *Bot) setMyCommands(items []MenuButtonItem) error {
	if len(items) == 0 {
		return nil
	}

	// Convert MenuButtonItems to Telegram BotCommand format
	var commands []map[string]string
	for _, item := range items {
		// Remove leading slash if present
		command := item.Command
		if len(command) > 0 && command[0] == '/' {
			command = command[1:]
		}

		commands = append(commands, map[string]string{
			"command":     command,
			"description": item.Text,
		})
	}

	// Create the request parameters
	params := map[string]interface{}{
		"commands": commands,
	}

	// Make the API call
	err := b.makeMenuButtonAPICall("setMyCommands", params)
	if err != nil {
		return fmt.Errorf("failed to set bot commands: %w", err)
	}

	log.Printf("✅ Registered %d bot commands with Telegram", len(commands))
	return nil
}

// InitializeMenuButton sets up the menu button when the bot starts
func (b *Bot) InitializeMenuButton() {
	if b.menuButton != nil {
		err := b.SetDefaultMenuButton()
		if err != nil {
			log.Printf("Failed to set default menu button: %v", err)
		} else {
			log.Printf("✅ Menu button configured: %s", b.menuButton.Type)

			// Log menu items for commands type
			if b.menuButton.Type == MenuButtonTypeCommands && len(b.menuButton.Items) > 0 {
				log.Printf("📋 Menu commands available:")
				for _, item := range b.menuButton.Items {
					log.Printf("   %s - %s", item.Text, item.Command)
				}
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
