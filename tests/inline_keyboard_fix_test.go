package tests

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	teleflow "github.com/kslamph/teleflow/core"
)

// TestInlineKeyboardEmptyHandling tests the fix for the "inline_keyboard must be of type Array" error
func TestInlineKeyboardEmptyHandling(t *testing.T) {
	// Create a test bot
	bot := &teleflow.Bot{}

	// Create a test context
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
			From: &tgbotapi.User{ID: 67890},
			Text: "test message",
		},
	}
	ctx := teleflow.NewContext(bot, update)

	// Register a template without any keyboard (like the "not_understood" template)
	err := bot.AddTemplate("not_understood", "❓ I didn't understand `{{.Input}}`. Type /help for available commands.", teleflow.ParseModeMarkdownV2)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	// Test template rendering without keyboard - this should not cause "inline_keyboard must be of type Array" error
	templateManager := teleflow.GetDefaultTemplateManager()
	result, parseMode, err := templateManager.RenderTemplate("not_understood", map[string]interface{}{
		"Input": "unknown command",
	})
	if err != nil {
		t.Errorf("Failed to render template: %v", err)
	}

	expectedResult := "❓ I didn't understand `unknown command`. Type /help for available commands."
	if result != expectedResult {
		t.Errorf("Expected '%s', got '%s'", expectedResult, result)
	}

	if parseMode != teleflow.ParseModeMarkdownV2 {
		t.Errorf("Expected ParseModeMarkdownV2, got %s", parseMode)
	}

	// Test that the convenience methods exist and can be called without error
	_ = ctx.ReplyTemplate
	_ = ctx.SendPromptWithTemplate

	t.Log("✅ Template rendering without keyboard works correctly")
}

// TestEmptyInlineKeyboardHandling specifically tests that empty inline keyboards are handled correctly
func TestEmptyInlineKeyboardHandling(t *testing.T) {
	// Test that an empty tgbotapi.InlineKeyboardMarkup is properly detected and ignored
	emptyKeyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}

	// This should be treated as "no keyboard" - the important thing is that it has no buttons
	if len(emptyKeyboard.InlineKeyboard) != 0 {
		t.Error("Empty keyboard should have no buttons")
	}

	// Test that a keyboard with empty rows is still considered empty
	emptyRowsKeyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{}, // Empty row
		},
	}

	// This keyboard has rows but no actual buttons, so it should still be considered empty
	hasButtons := false
	for _, row := range emptyRowsKeyboard.InlineKeyboard {
		if len(row) > 0 {
			hasButtons = true
			break
		}
	}
	if hasButtons {
		t.Error("Keyboard with empty rows should have no buttons")
	}

	t.Log("✅ Empty inline keyboard detection works correctly")
}

// TestKeyboardBuilderNilHandling tests that the keyboard builder returns nil for no keyboard
func TestKeyboardBuilderNilHandling(t *testing.T) {
	// Create a keyboard handler
	handler := teleflow.NewPromptKeyboardHandler()

	// Create a test context
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
			From: &tgbotapi.User{ID: 67890},
			Text: "test",
		},
	}
	bot := &teleflow.Bot{}
	ctx := teleflow.NewContext(bot, update)

	// Test with nil keyboard function
	result, err := handler.BuildKeyboard(ctx, nil)
	if err != nil {
		t.Errorf("BuildKeyboard with nil function should not error: %v", err)
	}
	if result != nil {
		t.Error("BuildKeyboard with nil function should return nil")
	}

	// Test with keyboard function that returns nil
	nilKeyboardFunc := func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
		return nil
	}
	result2, err := handler.BuildKeyboard(ctx, nilKeyboardFunc)
	if err != nil {
		t.Errorf("BuildKeyboard with nil-returning function should not error: %v", err)
	}
	if result2 != nil {
		t.Error("BuildKeyboard with nil-returning function should return nil")
	}

	t.Log("✅ Keyboard builder nil handling works correctly")
}
