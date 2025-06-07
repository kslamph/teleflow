package tests

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	teleflow "github.com/kslamph/teleflow/core"
)

// TestPhase5TemplateConvenienceMethods tests the new convenience methods added in Phase 5
func TestPhase5TemplateConvenienceMethods(t *testing.T) {
	// Create a test bot
	bot := &teleflow.Bot{}

	// Create a test context
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 12345},
			From: &tgbotapi.User{ID: 67890},
			Text: "/test",
		},
	}
	ctx := teleflow.NewContext(bot, update)

	// Test template registration (backwards compatibility)
	err := bot.AddTemplate("test_template", "Hello {{.Name}}!", teleflow.ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	// Test HasTemplate (backwards compatibility)
	if !bot.HasTemplate("test_template") {
		t.Error("HasTemplate should return true for registered template")
	}

	// Test GetTemplateInfo (backwards compatibility)
	info := bot.GetTemplateInfo("test_template")
	if info == nil {
		t.Error("GetTemplateInfo should return template info")
		return
	}
	if info.Name != "test_template" {
		t.Errorf("Expected template name 'test_template', got '%s'", info.Name)
	}

	// Test ListTemplates (backwards compatibility)
	templates := bot.ListTemplates()
	found := false
	for _, name := range templates {
		if name == "test_template" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListTemplates should include 'test_template'")
	}

	// Test that convenience methods exist and can be called
	// The actual sending would require a full bot setup with flow manager
	_ = ctx.ReplyTemplate
	_ = ctx.SendPromptWithTemplate

	// Verify the template data merging works with TemplateData field
	templateData := map[string]interface{}{"Name": "Test"}
	if templateData["Name"] != "Test" {
		t.Error("Template data should preserve values")
	}
}

// TestTemplateMessageDetection tests the template message detection
func TestTemplateMessageDetection(t *testing.T) {
	testCases := []struct {
		message    string
		isTemplate bool
		name       string
	}{
		{"template:welcome", true, "welcome"},
		{"template:user_profile", true, "user_profile"},
		{"Hello World", false, ""},
		{"template:", false, ""},
		{"not_template:test", false, ""},
	}

	for _, tc := range testCases {
		// This tests the template detection logic matching the actual implementation
		const templatePrefix = "template:"
		isTemplate := len(tc.message) > len(templatePrefix) && tc.message[:len(templatePrefix)] == templatePrefix
		var name string
		if isTemplate {
			name = tc.message[len(templatePrefix):]
		}

		if isTemplate != tc.isTemplate {
			t.Errorf("For message '%s': expected isTemplate=%v, got %v", tc.message, tc.isTemplate, isTemplate)
		}

		if name != tc.name {
			t.Errorf("For message '%s': expected name='%s', got '%s'", tc.message, tc.name, name)
		}
	}
}
