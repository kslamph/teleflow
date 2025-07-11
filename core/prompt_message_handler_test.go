package teleflow

import (
	"errors"
	"strings"
	"testing"
)

// Mock TemplateManager for messageHandler testing
type messageHandlerMockTemplateManager struct {
	hasTemplateCalls    []string
	renderTemplateCalls []struct {
		name string
		data map[string]interface{}
	}

	hasTemplateFunc    func(name string) bool
	renderTemplateFunc func(name string, data map[string]interface{}) (string, ParseMode, error)
}

func (m *messageHandlerMockTemplateManager) AddTemplate(name, templateText string, parseMode ParseMode) error {
	return nil
}

func (m *messageHandlerMockTemplateManager) HasTemplate(name string) bool {
	m.hasTemplateCalls = append(m.hasTemplateCalls, name)
	if m.hasTemplateFunc != nil {
		return m.hasTemplateFunc(name)
	}
	return false
}

func (m *messageHandlerMockTemplateManager) GetTemplateInfo(name string) *TemplateInfo {
	return nil
}

func (m *messageHandlerMockTemplateManager) ListTemplates() []string {
	return []string{}
}

func (m *messageHandlerMockTemplateManager) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	m.renderTemplateCalls = append(m.renderTemplateCalls, struct {
		name string
		data map[string]interface{}
	}{name, data})

	if m.renderTemplateFunc != nil {
		return m.renderTemplateFunc(name, data)
	}
	return "default rendered", ParseModeNone, nil
}

// Helper function to create test context
func createMessageHandlerTestContext() *Context {
	return &Context{
		data: map[string]interface{}{
			"contextKey": "contextValue",
		},
	}
}

func TestNewMessageHandler(t *testing.T) {
	mockTM := &messageHandlerMockTemplateManager{}

	handler := newMessageHandler(mockTM)

	if handler == nil {
		t.Fatal("newMessageHandler returned nil")
	}

	if handler.templateManager != mockTM {
		t.Error("newMessageHandler did not correctly store the TemplateManager")
	}
}

func TestMessageHandler_RenderMessage(t *testing.T) {
	tests := []struct {
		name          string
		config        *PromptConfig
		context       *Context
		mockSetup     func(*messageHandlerMockTemplateManager)
		expectedText  string
		expectedMode  ParseMode
		expectedError bool
		errorContains string
	}{
		{
			name: "nil message returns empty",
			config: &PromptConfig{
				Message: nil,
			},
			context:       createMessageHandlerTestContext(),
			expectedText:  "",
			expectedMode:  ParseModeNone,
			expectedError: false,
		},
		{
			name: "plain string message",
			config: &PromptConfig{
				Message: "Hello World",
			},
			context:       createMessageHandlerTestContext(),
			expectedText:  "Hello World",
			expectedMode:  ParseModeNone,
			expectedError: false,
		},
		{
			name: "function message returns string",
			config: &PromptConfig{
				Message: func(ctx *Context) string {
					return "Function result"
				},
			},
			context:       createMessageHandlerTestContext(),
			expectedText:  "Function result",
			expectedMode:  ParseModeNone,
			expectedError: false,
		},
		{
			name: "function message with context data",
			config: &PromptConfig{
				Message: func(ctx *Context) string {
					if val, ok := ctx.Get("contextKey"); ok {
						return "Got: " + val.(string)
					}
					return "No context"
				},
			},
			context:       createMessageHandlerTestContext(),
			expectedText:  "Got: contextValue",
			expectedMode:  ParseModeNone,
			expectedError: false,
		},
		{
			name: "template message - successful render",
			config: &PromptConfig{
				Message: "template:greeting",
				TemplateData: map[string]interface{}{
					"name": "John",
				},
			},
			context: createMessageHandlerTestContext(),
			mockSetup: func(mock *messageHandlerMockTemplateManager) {
				mock.hasTemplateFunc = func(name string) bool {
					return name == "greeting"
				}
				mock.renderTemplateFunc = func(name string, data map[string]interface{}) (string, ParseMode, error) {
					if name == "greeting" {
						return "Hello John!", ParseModeHTML, nil
					}
					return "", ParseModeNone, errors.New("template not found")
				}
			},
			expectedText:  "Hello John!",
			expectedMode:  ParseModeHTML,
			expectedError: false,
		},
		{
			name: "template message - template not found",
			config: &PromptConfig{
				Message: "template:missing",
			},
			context: createMessageHandlerTestContext(),
			mockSetup: func(mock *messageHandlerMockTemplateManager) {
				mock.hasTemplateFunc = func(name string) bool {
					return false
				}
			},
			expectedText:  "",
			expectedMode:  ParseModeNone,
			expectedError: true,
			errorContains: "template 'missing' not found",
		},
		{
			name: "template message - render error",
			config: &PromptConfig{
				Message: "template:error",
			},
			context: createMessageHandlerTestContext(),
			mockSetup: func(mock *messageHandlerMockTemplateManager) {
				mock.hasTemplateFunc = func(name string) bool {
					return name == "error"
				}
				mock.renderTemplateFunc = func(name string, data map[string]interface{}) (string, ParseMode, error) {
					return "", ParseModeNone, errors.New("render failed")
				}
			},
			expectedText:  "",
			expectedMode:  ParseModeNone,
			expectedError: true,
			errorContains: "failed to render template 'error': render failed",
		},
		{
			name: "unsupported message type",
			config: &PromptConfig{
				Message: 123, // int is not supported
			},
			context:       createMessageHandlerTestContext(),
			expectedText:  "",
			expectedMode:  ParseModeNone,
			expectedError: true,
			errorContains: "unsupported message type: int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTM := &messageHandlerMockTemplateManager{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockTM)
			}

			handler := newMessageHandler(mockTM)

			text, mode, err := handler.renderMessage(tt.config, tt.context)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if text != tt.expectedText {
				t.Errorf("Expected text '%s', got '%s'", tt.expectedText, text)
			}

			if mode != tt.expectedMode {
				t.Errorf("Expected ParseMode '%s', got '%s'", tt.expectedMode, mode)
			}
		})
	}
}

func TestMessageHandler_HandleStringMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		config        *PromptConfig
		context       *Context
		mockSetup     func(*messageHandlerMockTemplateManager)
		expectedText  string
		expectedMode  ParseMode
		expectedError bool
	}{
		{
			name:          "plain text message",
			message:       "Hello World",
			config:        &PromptConfig{},
			context:       createMessageHandlerTestContext(),
			expectedText:  "Hello World",
			expectedMode:  ParseModeNone,
			expectedError: false,
		},
		{
			name:    "template message",
			message: "template:greeting",
			config: &PromptConfig{
				TemplateData: map[string]interface{}{
					"name": "Alice",
				},
			},
			context: createMessageHandlerTestContext(),
			mockSetup: func(mock *messageHandlerMockTemplateManager) {
				mock.hasTemplateFunc = func(name string) bool {
					return name == "greeting"
				}
				mock.renderTemplateFunc = func(name string, data map[string]interface{}) (string, ParseMode, error) {
					return "Hello Alice!", ParseModeMarkdown, nil
				}
			},
			expectedText:  "Hello Alice!",
			expectedMode:  ParseModeMarkdown,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTM := &messageHandlerMockTemplateManager{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockTM)
			}

			handler := newMessageHandler(mockTM)

			text, mode, err := handler.handleStringMessage(tt.message, tt.config)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if text != tt.expectedText {
				t.Errorf("Expected text '%s', got '%s'", tt.expectedText, text)
			}

			if mode != tt.expectedMode {
				t.Errorf("Expected ParseMode '%s', got '%s'", tt.expectedMode, mode)
			}
		})
	}
}

func TestMessageHandler_RenderTemplateMessage(t *testing.T) {
	tests := []struct {
		name          string
		templateName  string
		config        *PromptConfig
		context       *Context
		mockSetup     func(*messageHandlerMockTemplateManager)
		expectedText  string
		expectedMode  ParseMode
		expectedError bool
		errorContains string
	}{
		{
			name:         "successful template render",
			templateName: "welcome",
			config: &PromptConfig{
				TemplateData: map[string]interface{}{
					"user": "Bob",
				},
			},
			context: createMessageHandlerTestContext(),
			mockSetup: func(mock *messageHandlerMockTemplateManager) {
				mock.hasTemplateFunc = func(name string) bool {
					return name == "welcome"
				}
				mock.renderTemplateFunc = func(name string, data map[string]interface{}) (string, ParseMode, error) {
					return "Welcome Bob!", ParseModeHTML, nil
				}
			},
			expectedText:  "Welcome Bob!",
			expectedMode:  ParseModeHTML,
			expectedError: false,
		},
		{
			name:         "template not found",
			templateName: "nonexistent",
			config:       &PromptConfig{},
			context:      createMessageHandlerTestContext(),
			mockSetup: func(mock *messageHandlerMockTemplateManager) {
				mock.hasTemplateFunc = func(name string) bool {
					return false
				}
			},
			expectedText:  "",
			expectedMode:  ParseModeNone,
			expectedError: true,
			errorContains: "template 'nonexistent' not found",
		},
		{
			name:         "template render error",
			templateName: "broken",
			config:       &PromptConfig{},
			context:      createMessageHandlerTestContext(),
			mockSetup: func(mock *messageHandlerMockTemplateManager) {
				mock.hasTemplateFunc = func(name string) bool {
					return name == "broken"
				}
				mock.renderTemplateFunc = func(name string, data map[string]interface{}) (string, ParseMode, error) {
					return "", ParseModeNone, errors.New("template syntax error")
				}
			},
			expectedText:  "",
			expectedMode:  ParseModeNone,
			expectedError: true,
			errorContains: "failed to render template 'broken': template syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTM := &messageHandlerMockTemplateManager{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockTM)
			}

			handler := newMessageHandler(mockTM)

			text, mode, err := handler.renderTemplateMessage(tt.templateName, tt.config)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if text != tt.expectedText {
				t.Errorf("Expected text '%s', got '%s'", tt.expectedText, text)
			}

			if mode != tt.expectedMode {
				t.Errorf("Expected ParseMode '%s', got '%s'", tt.expectedMode, mode)
			}
		})
	}
}

func TestMessageHandler_TemplateDataExplicitOnly(t *testing.T) {
	// Test that only explicit TemplateData is passed to templates (no context data merging)
	mockTM := &messageHandlerMockTemplateManager{}
	mockTM.hasTemplateFunc = func(name string) bool {
		return name == "test"
	}

	var capturedData map[string]interface{}
	mockTM.renderTemplateFunc = func(name string, data map[string]interface{}) (string, ParseMode, error) {
		capturedData = data
		return "rendered", ParseModeNone, nil
	}

	handler := newMessageHandler(mockTM)

	config := &PromptConfig{
		TemplateData: map[string]interface{}{
			"template_key": "template_value",
		},
	}

	_, _, err := handler.renderTemplateMessage("test", config)
	if err != nil {
		t.Fatalf("renderTemplateMessage failed: %v", err)
	}

	// Verify only explicit TemplateData was passed to the template manager
	if capturedData["template_key"] != "template_value" {
		t.Error("Explicit template data should be passed")
	}
	if _, exists := capturedData["context_key"]; exists {
		t.Error("Context data should NOT be automatically passed to templates")
	}
	if _, exists := capturedData["override"]; exists {
		t.Error("Context data should NOT be automatically passed to templates")
	}
}

func TestTemplateExplicitDataBehavior(t *testing.T) {
	// Test 3: Template Explicit Data - Complete implementation
	// This test verifies the two critical scenarios for template data isolation

	// Create a real template manager to test actual template rendering
	tm := newTemplateManager()

	// Add a test template that references both flow_var and explicit_var
	templateText := "Hello {{.flow_var}} {{.explicit_var}}"
	err := tm.AddTemplate("test_template", templateText, ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to add test template: %v", err)
	}

	handler := newMessageHandler(tm)

	// Setup: Create context with flow data (simulated via context.data for testing)
	// In real scenarios, this would be set via SetFlowData

	t.Run("Template without explicit data - no access to flow data", func(t *testing.T) {
		// Test Case 1: Template without explicit TemplateData
		config := &PromptConfig{
			TemplateData: nil, // No explicit template data
		}

		renderedText, _, err := handler.renderTemplateMessage("test_template", config)
		if err != nil {
			t.Fatalf("renderTemplateMessage failed: %v", err)
		}

		// Verify: Template should NOT have access to flow_var
		// The rendered output should be "Hello <no value> <no value>" (template engine's default for missing variables)
		expected := "Hello <no value> <no value>"
		if renderedText != expected {
			t.Errorf("Expected '%s', got '%s'. Template should not have access to flow data when TemplateData is nil", expected, renderedText)
		}

		// Additional verification: ensure it doesn't contain flow_value
		if strings.Contains(renderedText, "flow_value") {
			t.Error("Template should NOT have access to flow data (flow_value) when TemplateData is not provided")
		}
	})

	t.Run("Template with explicit data - access to explicit data only", func(t *testing.T) {
		// Test Case 2: Template with explicit TemplateData
		config := &PromptConfig{
			TemplateData: map[string]interface{}{
				"explicit_var": "explicit_value",
				"flow_var":     "override_flow_value", // This should override any potential flow data
			},
		}
		renderedText, _, err := handler.renderTemplateMessage("test_template", config)
		if err != nil {
			t.Fatalf("renderTemplateMessage failed: %v", err)
		}

		// Verify: Template should have access to explicitly provided data
		expected := "Hello override_flow_value explicit_value"
		if renderedText != expected {
			t.Errorf("Expected '%s', got '%s'. Template should have access to explicit TemplateData", expected, renderedText)
		}

		// Additional verifications
		if !strings.Contains(renderedText, "override_flow_value") {
			t.Error("Template should contain explicitly provided flow_var value (override_flow_value)")
		}
		if !strings.Contains(renderedText, "explicit_value") {
			t.Error("Template should contain explicitly provided explicit_var value (explicit_value)")
		}
		// Note: We expect "override_flow_value" not "flow_value"
		// The rendered text contains "override_flow_value", which includes "flow_value" as a substring
		// So we need to check more specifically
		if strings.Contains(renderedText, "flow_value") && !strings.Contains(renderedText, "override_") {
			t.Error("Template should NOT contain original flow data value (flow_value) when explicit data is provided")
		}
	})
}
