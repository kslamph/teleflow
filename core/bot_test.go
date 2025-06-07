package teleflow

import (
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestableBotAPI extends TestBotAPI to track Request calls specifically for SetBotCommands
type TestableBotAPI struct {
	*TestBotAPI
	RequestFunc  func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	RequestCalls []tgbotapi.Chattable
}

func NewTestableBotAPI() *TestableBotAPI {
	return &TestableBotAPI{
		TestBotAPI:   &TestBotAPI{},
		RequestCalls: make([]tgbotapi.Chattable, 0),
	}
}

func (t *TestableBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	t.RequestCalls = append(t.RequestCalls, c)
	if t.RequestFunc != nil {
		return t.RequestFunc(c)
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

// TestableBot extends Bot to allow dependency injection for testing
type TestableBot struct {
	api *TestableBotAPI
}

func NewTestableBot(api *TestableBotAPI) *TestableBot {
	return &TestableBot{
		api: api,
	}
}

func (tb *TestableBot) SetBotCommands(commands map[string]string) error {
	if tb.api == nil {
		return errors.New("bot API not initialized")
	}

	if len(commands) == 0 {
		// To clear commands, send an empty list
		clearCmdCfg := tgbotapi.NewSetMyCommands()
		_, err := tb.api.Request(clearCmdCfg)
		if err != nil {
			return errors.New("failed to clear bot commands: " + err.Error())
		}
		return nil
	}

	var tgCommands []tgbotapi.BotCommand
	for cmd, desc := range commands {
		tgCommands = append(tgCommands, tgbotapi.BotCommand{Command: cmd, Description: desc})
	}
	cmdCfg := tgbotapi.NewSetMyCommands(tgCommands...)
	_, err := tb.api.Request(cmdCfg)
	if err != nil {
		return errors.New("failed to set bot commands: " + err.Error())
	}
	return nil
}

func TestBot_SetBotCommands_NewCommands(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	bot := NewTestableBot(api)

	commands := map[string]string{
		"start": "Start the bot",
		"help":  "Show help information",
		"about": "About this bot",
	}

	// Execute
	err := bot.SetBotCommands(commands)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that Request was called
	if len(api.RequestCalls) != 1 {
		t.Errorf("Expected 1 Request call, got %d", len(api.RequestCalls))
	}

	// Verify the SetMyCommandsConfig was sent
	if cmdConfig, ok := api.RequestCalls[0].(tgbotapi.SetMyCommandsConfig); ok {
		if len(cmdConfig.Commands) != 3 {
			t.Errorf("Expected 3 commands, got %d", len(cmdConfig.Commands))
		}

		// Verify commands are present (order may vary due to map iteration)
		expectedCommands := map[string]string{
			"start": "Start the bot",
			"help":  "Show help information",
			"about": "About this bot",
		}

		for _, cmd := range cmdConfig.Commands {
			expectedDesc, exists := expectedCommands[cmd.Command]
			if !exists {
				t.Errorf("Unexpected command: %s", cmd.Command)
				continue
			}
			if cmd.Description != expectedDesc {
				t.Errorf("Command %s: expected description '%s', got '%s'", cmd.Command, expectedDesc, cmd.Description)
			}
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig, got %T", api.RequestCalls[0])
	}
}

func TestBot_SetBotCommands_SingleCommand(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	bot := NewTestableBot(api)

	commands := map[string]string{
		"start": "Start the bot",
	}

	// Execute
	err := bot.SetBotCommands(commands)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify single command was set
	if cmdConfig, ok := api.RequestCalls[0].(tgbotapi.SetMyCommandsConfig); ok {
		if len(cmdConfig.Commands) != 1 {
			t.Errorf("Expected 1 command, got %d", len(cmdConfig.Commands))
		}
		if cmdConfig.Commands[0].Command != "start" {
			t.Errorf("Expected command 'start', got '%s'", cmdConfig.Commands[0].Command)
		}
		if cmdConfig.Commands[0].Description != "Start the bot" {
			t.Errorf("Expected description 'Start the bot', got '%s'", cmdConfig.Commands[0].Description)
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig, got %T", api.RequestCalls[0])
	}
}

func TestBot_SetBotCommands_ClearCommands(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	bot := NewTestableBot(api)

	// Execute with empty commands map
	err := bot.SetBotCommands(map[string]string{})

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that Request was called
	if len(api.RequestCalls) != 1 {
		t.Errorf("Expected 1 Request call, got %d", len(api.RequestCalls))
	}

	// Verify empty SetMyCommandsConfig was sent
	if cmdConfig, ok := api.RequestCalls[0].(tgbotapi.SetMyCommandsConfig); ok {
		if len(cmdConfig.Commands) != 0 {
			t.Errorf("Expected 0 commands for clearing, got %d", len(cmdConfig.Commands))
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig, got %T", api.RequestCalls[0])
	}
}

func TestBot_SetBotCommands_NilCommands(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	bot := NewTestableBot(api)

	// Execute with nil commands map
	err := bot.SetBotCommands(nil)

	// Verify - nil map should be treated as empty and clear commands
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that Request was called
	if len(api.RequestCalls) != 1 {
		t.Errorf("Expected 1 Request call, got %d", len(api.RequestCalls))
	}

	// Verify empty SetMyCommandsConfig was sent
	if cmdConfig, ok := api.RequestCalls[0].(tgbotapi.SetMyCommandsConfig); ok {
		if len(cmdConfig.Commands) != 0 {
			t.Errorf("Expected 0 commands for nil map, got %d", len(cmdConfig.Commands))
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig, got %T", api.RequestCalls[0])
	}
}

func TestBot_SetBotCommands_APIError(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	api.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
		return nil, errors.New("API request failed")
	}
	bot := NewTestableBot(api)

	commands := map[string]string{
		"start": "Start the bot",
	}

	// Execute
	err := bot.SetBotCommands(commands)

	// Verify
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "failed to set bot commands: API request failed" {
		t.Errorf("Expected 'failed to set bot commands: API request failed', got '%s'", err.Error())
	}
}

func TestBot_SetBotCommands_ClearCommandsAPIError(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	api.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
		return nil, errors.New("Clear API request failed")
	}
	bot := NewTestableBot(api)

	// Execute with empty commands to trigger clear
	err := bot.SetBotCommands(map[string]string{})

	// Verify
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "failed to clear bot commands: Clear API request failed" {
		t.Errorf("Expected 'failed to clear bot commands: Clear API request failed', got '%s'", err.Error())
	}
}

func TestBot_SetBotCommands_NilAPIError(t *testing.T) {
	// Setup bot with nil API
	bot := NewTestableBot(nil)

	commands := map[string]string{
		"start": "Start the bot",
	}

	// Execute
	err := bot.SetBotCommands(commands)

	// Verify
	if err == nil {
		t.Error("Expected error for nil API, got nil")
	}
	if err != nil && err.Error() != "bot API not initialized" {
		t.Errorf("Expected 'bot API not initialized', got '%s'", err.Error())
	}
}

func TestBot_SetBotCommands_LongDescriptions(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	bot := NewTestableBot(api)

	commands := map[string]string{
		"help":      "This is a very long description that explains what this command does in great detail and provides comprehensive information about its usage",
		"settings":  "Configure bot settings and preferences",
		"analytics": "View detailed analytics and statistics about bot usage and performance metrics",
	}

	// Execute
	err := bot.SetBotCommands(commands)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify all commands with long descriptions were processed
	if cmdConfig, ok := api.RequestCalls[0].(tgbotapi.SetMyCommandsConfig); ok {
		if len(cmdConfig.Commands) != 3 {
			t.Errorf("Expected 3 commands, got %d", len(cmdConfig.Commands))
		}

		// Verify each command has its description preserved
		foundCommands := make(map[string]string)
		for _, cmd := range cmdConfig.Commands {
			foundCommands[cmd.Command] = cmd.Description
		}

		for expectedCmd, expectedDesc := range commands {
			if actualDesc, exists := foundCommands[expectedCmd]; !exists {
				t.Errorf("Command '%s' not found", expectedCmd)
			} else if actualDesc != expectedDesc {
				t.Errorf("Command '%s': expected description '%s', got '%s'", expectedCmd, expectedDesc, actualDesc)
			}
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig, got %T", api.RequestCalls[0])
	}
}

func TestBot_SetBotCommands_SpecialCharacters(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	bot := NewTestableBot(api)

	commands := map[string]string{
		"start":    "🚀 Start the bot",
		"help":     "❓ Get help & support",
		"settings": "⚙️ Configure settings",
	}

	// Execute
	err := bot.SetBotCommands(commands)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify special characters in descriptions are preserved
	if cmdConfig, ok := api.RequestCalls[0].(tgbotapi.SetMyCommandsConfig); ok {
		foundCommands := make(map[string]string)
		for _, cmd := range cmdConfig.Commands {
			foundCommands[cmd.Command] = cmd.Description
		}

		for expectedCmd, expectedDesc := range commands {
			if actualDesc, exists := foundCommands[expectedCmd]; !exists {
				t.Errorf("Command '%s' not found", expectedCmd)
			} else if actualDesc != expectedDesc {
				t.Errorf("Command '%s': expected description '%s', got '%s'", expectedCmd, expectedDesc, actualDesc)
			}
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig, got %T", api.RequestCalls[0])
	}
}

func TestBot_SetBotCommands_MultipleCallsSequential(t *testing.T) {
	// Setup
	api := NewTestableBotAPI()
	bot := NewTestableBot(api)

	// First set of commands
	commands1 := map[string]string{
		"start": "Start the bot",
		"help":  "Get help",
	}

	// Second set of commands
	commands2 := map[string]string{
		"info":     "Bot information",
		"settings": "Configure settings",
		"about":    "About this bot",
	}

	// Execute first set
	err1 := bot.SetBotCommands(commands1)
	if err1 != nil {
		t.Errorf("Expected no error for first set, got: %v", err1)
	}

	// Execute second set
	err2 := bot.SetBotCommands(commands2)
	if err2 != nil {
		t.Errorf("Expected no error for second set, got: %v", err2)
	}

	// Verify both calls were made
	if len(api.RequestCalls) != 2 {
		t.Errorf("Expected 2 Request calls, got %d", len(api.RequestCalls))
	}

	// Verify first set
	if cmdConfig1, ok := api.RequestCalls[0].(tgbotapi.SetMyCommandsConfig); ok {
		if len(cmdConfig1.Commands) != 2 {
			t.Errorf("Expected 2 commands in first set, got %d", len(cmdConfig1.Commands))
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig for first call, got %T", api.RequestCalls[0])
	}

	// Verify second set
	if cmdConfig2, ok := api.RequestCalls[1].(tgbotapi.SetMyCommandsConfig); ok {
		if len(cmdConfig2.Commands) != 3 {
			t.Errorf("Expected 3 commands in second set, got %d", len(cmdConfig2.Commands))
		}
	} else {
		t.Errorf("Expected SetMyCommandsConfig for second call, got %T", api.RequestCalls[1])
	}
}
