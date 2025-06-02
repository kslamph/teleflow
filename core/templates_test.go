package teleflow

import (
	"testing"
)

func TestAddTemplate(t *testing.T) {
	// Create a bot instance for testing
	bot := &Bot{
		templates: nil, // Start with nil to test initialization
	}

	// Test adding a simple template
	templateText := "Hello {{.Name}}, welcome to {{.BotName}}!"
	err := bot.AddTemplate("welcome", templateText, ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	// Verify template was added
	if !bot.HasTemplate("welcome") {
		t.Error("Template 'welcome' was not found after adding")
	}

	// Test adding template with empty name
	err = bot.AddTemplate("", "some text", ParseModeNone)
	if err == nil {
		t.Error("Expected error when adding template with empty name")
	}

	// Test adding template with empty text
	err = bot.AddTemplate("empty", "", ParseModeNone)
	if err == nil {
		t.Error("Expected error when adding template with empty text")
	}

	// Test adding invalid template syntax
	err = bot.AddTemplate("invalid", "Hello {{.Name", ParseModeNone)
	if err == nil {
		t.Error("Expected error when adding template with invalid syntax")
	}
}

func TestListTemplates(t *testing.T) {
	bot := &Bot{
		templates: nil,
	}

	// Initially should be empty
	templates := bot.ListTemplates()
	if len(templates) != 0 {
		t.Errorf("Expected 0 templates initially, got %d", len(templates))
	}

	// Add some templates
	bot.AddTemplate("template1", "Hello {{.Name}}", ParseModeNone)
	bot.AddTemplate("template2", "Goodbye {{.Name}}", ParseModeNone)

	templates = bot.ListTemplates()
	if len(templates) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(templates))
	}

	// Check template names are present
	hasTemplate1 := false
	hasTemplate2 := false
	for _, name := range templates {
		if name == "template1" {
			hasTemplate1 = true
		}
		if name == "template2" {
			hasTemplate2 = true
		}
	}

	if !hasTemplate1 || !hasTemplate2 {
		t.Error("Not all expected templates found in list")
	}
}

func TestGetTemplate(t *testing.T) {
	bot := &Bot{
		templates: nil,
	}

	// Test getting non-existent template
	tmpl := bot.GetTemplate("nonexistent")
	if tmpl != nil {
		t.Error("Expected nil for non-existent template")
	}

	// Add a template and test retrieval
	bot.AddTemplate("test", "Test template", ParseModeNone)
	tmpl = bot.GetTemplate("test")
	if tmpl == nil {
		t.Error("Expected to find template 'test'")
	}
}

func TestHasTemplate(t *testing.T) {
	bot := &Bot{
		templates: nil,
	}

	// Test non-existent template
	if bot.HasTemplate("nonexistent") {
		t.Error("Expected false for non-existent template")
	}

	// Add template and test
	bot.AddTemplate("exists", "Template content", ParseModeNone)
	if !bot.HasTemplate("exists") {
		t.Error("Expected true for existing template")
	}
}
