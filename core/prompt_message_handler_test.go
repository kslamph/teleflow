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

			text, mode, err := handler.handleStringMessage(tt.message, tt.config, tt.context)

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

			text, mode, err := handler.renderTemplateMessage(tt.templateName, tt.config, tt.context)

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

func TestMessageHandler_MergeDataSources(t *testing.T) {
	tests := []struct {
		name         string
		templateData map[string]interface{}
		contextData  map[string]interface{}
		expected     map[string]interface{}
	}{
		{
			name:         "both nil",
			templateData: nil,
			contextData:  nil,
			expected:     map[string]interface{}{},
		},
		{
			name:         "only template data",
			templateData: map[string]interface{}{"key1": "value1"},
			contextData:  nil,
			expected:     map[string]interface{}{"key1": "value1"},
		},
		{
			name:         "only context data",
			templateData: nil,
			contextData:  map[string]interface{}{"key2": "value2"},
			expected:     map[string]interface{}{"key2": "value2"},
		},
		{
			name:         "both with different keys",
			templateData: map[string]interface{}{"tkey": "tvalue"},
			contextData:  map[string]interface{}{"ckey": "cvalue"},
			expected:     map[string]interface{}{"tkey": "tvalue", "ckey": "cvalue"},
		},
		{
			name:         "template data overrides context data",
			templateData: map[string]interface{}{"key": "template_value"},
			contextData:  map[string]interface{}{"key": "context_value"},
			expected:     map[string]interface{}{"key": "template_value"},
		},
		{
			name: "complex merge",
			templateData: map[string]interface{}{
				"name":     "John",
				"override": "template",
			},
			contextData: map[string]interface{}{
				"id":       123,
				"override": "context",
				"extra":    "data",
			},
			expected: map[string]interface{}{
				"name":     "John",
				"id":       123,
				"override": "template",
				"extra":    "data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newMessageHandler(&messageHandlerMockTemplateManager{})

			result := handler.mergeDataSources(tt.templateData, tt.contextData)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected merged data length %d, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected key '%s' not found in merged data", key)
				} else if actualValue != expectedValue {
					t.Errorf("Expected value '%v' for key '%s', got '%v'", expectedValue, key, actualValue)
				}
			}
		})
	}
}

func TestMessageHandler_TemplateDataMergeInRender(t *testing.T) {
	// Test that template data and context data are properly merged when rendering templates
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
			"override":     "from_template",
		},
	}

	context := &Context{
		data: map[string]interface{}{
			"context_key": "context_value",
			"override":    "from_context",
		},
	}

	_, _, err := handler.renderTemplateMessage("test", config, context)
	if err != nil {
		t.Fatalf("renderTemplateMessage failed: %v", err)
	}

	// Verify the merged data was passed to the template manager
	if capturedData["template_key"] != "template_value" {
		t.Error("Template data not properly merged")
	}
	if capturedData["context_key"] != "context_value" {
		t.Error("Context data not properly merged")
	}
	if capturedData["override"] != "from_template" {
		t.Error("Template data should override context data")
	}
}
