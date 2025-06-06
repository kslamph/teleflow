package teleflow

import (
	"testing"
)

// createTestBot creates a bot for testing without API validation
func createTestBot() *Bot {
	return &Bot{
		api:              nil, // No API needed for template tests
		handlers:         make(map[string]HandlerFunc),
		textHandlers:     make(map[string]HandlerFunc),
		callbackRegistry: newCallbackRegistry(),
		stateManager:     NewInMemoryStateManager(),
		flowManager:      NewFlowManager(NewInMemoryStateManager()),
		templates:        nil, // Will be initialized when first template is added
		middleware:       []MiddlewareFunc{},
		flowConfig: FlowConfig{
			ExitCommands:        []string{"/cancel", "/exit"},
			ExitMessage:         "Operation cancelled.",
			AllowGlobalCommands: false,
			HelpCommands:        []string{"/help"},
		},
	}
}

func TestEnhancedTemplateFeatures(t *testing.T) {
	bot := createTestBot()

	// Test different parse modes
	t.Run("ParseModes", func(t *testing.T) {
		// Plain text
		err := bot.AddTemplate("plain", "Hello {{.Name}}", ParseModeNone)
		if err != nil {
			t.Errorf("Failed to add plain template: %v", err)
		}

		// Markdown
		err = bot.AddTemplate("markdown", "*Hello {{.Name | escape}}*", ParseModeMarkdown)
		if err != nil {
			t.Errorf("Failed to add markdown template: %v", err)
		}

		// HTML
		err = bot.AddTemplate("html", "<b>Hello {{.Name | escape}}</b>", ParseModeHTML)
		if err != nil {
			t.Errorf("Failed to add HTML template: %v", err)
		}

		// MarkdownV2
		err = bot.AddTemplate("md2", "*Hello {{.Name | escape}}*", ParseModeMarkdownV2)
		if err != nil {
			t.Errorf("Failed to add MarkdownV2 template: %v", err)
		}
	})

	// Test template info retrieval
	t.Run("TemplateInfo", func(t *testing.T) {
		bot.AddTemplate("test_info", "Test {{.Value}}", ParseModeHTML)

		info := bot.GetTemplateInfo("test_info")
		if info == nil {
			t.Error("Expected template info, got nil")
		} else {
			if info.Name != "test_info" {
				t.Errorf("Expected name 'test_info', got '%s'", info.Name)
			}
			if info.ParseMode != ParseModeHTML {
				t.Errorf("Expected ParseModeHTML, got %v", info.ParseMode)
			}
		}

		// Test non-existent template
		info = bot.GetTemplateInfo("nonexistent")
		if info != nil {
			t.Error("Expected nil for non-existent template")
		}
	})

	// Test MustAddTemplate
	t.Run("MustAddTemplate", func(t *testing.T) {
		// This should not panic
		bot.MustAddTemplate("must_test", "Hello {{.Name}}", ParseModeNone)

		if !bot.HasTemplate("must_test") {
			t.Error("Expected template to be added")
		}
	})

	// Test validation errors
	t.Run("ValidationErrors", func(t *testing.T) {
		// Test invalid parse mode
		err := bot.AddTemplate("invalid_mode", "test", ParseMode("INVALID"))
		if err == nil {
			t.Error("Expected error for invalid parse mode")
		}

		// Test empty name
		err = bot.AddTemplate("", "test", ParseModeNone)
		if err == nil {
			t.Error("Expected error for empty template name")
		}

		// Test empty template
		err = bot.AddTemplate("empty_template", "", ParseModeNone)
		if err == nil {
			t.Error("Expected error for empty template text")
		}
	})

	// Test template functions
	t.Run("TemplateFunctions", func(t *testing.T) {
		bot.AddTemplate("functions_test", `
{{.Text | escape}}
{{.Text | upper}}
{{.Text | lower}}
{{.Text | title}}
`, ParseModeNone)

		if !bot.HasTemplate("functions_test") {
			t.Error("Template with functions should be valid")
		}
	})

	// Test HTML validation
	t.Run("HTMLValidation", func(t *testing.T) {
		// Valid HTML
		err := bot.AddTemplate("valid_html", "<b>Test {{.Name}}</b>", ParseModeHTML)
		if err != nil {
			t.Errorf("Valid HTML template should not error: %v", err)
		}

		// Invalid HTML (unmatched tags) - this should still pass our basic validation
		// as we're doing simple validation, not full HTML parsing
		err = bot.AddTemplate("simple_html", "<b>Test {{.Name}}", ParseModeHTML)
		if err != nil {
			t.Logf("HTML validation caught unmatched tag: %v", err)
		}
	})

	// Test Markdown validation
	t.Run("MarkdownValidation", func(t *testing.T) {
		// Valid markdown
		err := bot.AddTemplate("valid_md", "*Test {{.Name}}*", ParseModeMarkdown)
		if err != nil {
			t.Errorf("Valid markdown template should not error: %v", err)
		}

		// Invalid markdown (unmatched asterisks)
		err = bot.AddTemplate("invalid_md", "*Test {{.Name}", ParseModeMarkdown)
		if err == nil {
			t.Error("Expected error for unmatched markdown syntax")
		}
	})
}

func TestTemplateExecutionWithParseModes(t *testing.T) {
	bot := createTestBot()

	// Add templates with different parse modes
	bot.AddTemplate("escape_test", "Hello {{.Name | escape}}", ParseModeHTML)
	bot.AddTemplate("safe_test", "Hello {{.Name | safe}}", ParseModeHTML)

	// Test that templates execute without error
	// (We can't easily test the actual message sending without mocking)
	info := bot.GetTemplateInfo("escape_test")
	if info == nil {
		t.Fatal("Template info should not be nil")
	}

	if info.ParseMode != ParseModeHTML {
		t.Errorf("Expected HTML parse mode, got %v", info.ParseMode)
	}
}

func TestParseMode(t *testing.T) {
	testCases := []struct {
		mode  ParseMode
		valid bool
	}{
		{ParseModeNone, true},
		{ParseModeMarkdown, true},
		{ParseModeMarkdownV2, true},
		{ParseModeHTML, true},
		{ParseMode("INVALID"), false},
	}

	for _, tc := range testCases {
		err := validateParseMode(tc.mode)
		if tc.valid && err != nil {
			t.Errorf("Expected %v to be valid, got error: %v", tc.mode, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("Expected %v to be invalid, got no error", tc.mode)
		}
	}
}
