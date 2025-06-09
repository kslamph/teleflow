package teleflow

import (
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Mock implementations for dependencies

// mockTelegramClient implements TelegramClient interface for testing
type mockTelegramClient struct {
	sendFunc     func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	requestFunc  func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	sentMessages []tgbotapi.Chattable
}

func (m *mockTelegramClient) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)
	if m.sendFunc != nil {
		return m.sendFunc(c)
	}
	return tgbotapi.Message{MessageID: 123}, nil
}

func (m *mockTelegramClient) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	if m.requestFunc != nil {
		return m.requestFunc(c)
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *mockTelegramClient) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return make(tgbotapi.UpdatesChannel)
}

func (m *mockTelegramClient) GetMe() (tgbotapi.User, error) {
	return tgbotapi.User{ID: 123, UserName: "testbot"}, nil
}

// mockTemplateManager implements TemplateManager interface for testing
type mockTemplateManager struct {
	renderFunc func(name string, data map[string]interface{}) (string, ParseMode, error)
	hasFunc    func(name string) bool
}

func (m *mockTemplateManager) AddTemplate(name, templateText string, parseMode ParseMode) error {
	return nil
}

func (m *mockTemplateManager) HasTemplate(name string) bool {
	if m.hasFunc != nil {
		return m.hasFunc(name)
	}
	return true
}

func (m *mockTemplateManager) GetTemplateInfo(name string) *TemplateInfo {
	return &TemplateInfo{Name: name, ParseMode: ParseModeHTML}
}

func (m *mockTemplateManager) ListTemplates() []string {
	return []string{"test"}
}

func (m *mockTemplateManager) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	if m.renderFunc != nil {
		return m.renderFunc(name, data)
	}
	return "rendered: " + name, ParseModeHTML, nil
}

// Test helper to create a basic context
func createTestContext() *Context {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890},
		},
	}

	return &Context{
		update: update,
		data:   make(map[string]interface{}),
		userID: 12345,
		chatID: 67890,
	}
}

// Helper function to create a testable PromptComposer with real dependencies but controlled behavior
func createTestPromptComposer(client TelegramClient, templateMgr TemplateManager) *PromptComposer {
	msgHandler := newMessageHandler(templateMgr)
	imgHandler := newImageHandler()
	kbdHandler := newPromptKeyboardHandler()

	return newPromptComposer(client, msgHandler, imgHandler, kbdHandler)
}

func TestNewPromptComposer(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)

	if composer == nil {
		t.Fatal("createTestPromptComposer returned nil")
	}

	if composer.botAPI != mockClient {
		t.Error("botAPI not set correctly")
	}

	if composer.messageRenderer == nil {
		t.Error("messageRenderer not set")
	}

	if composer.imageHandler == nil {
		t.Error("imageHandler not set")
	}

	if composer.keyboardHandler == nil {
		t.Error("keyboardHandler not set")
	}
}

func TestPromptComposer_ValidatePromptConfig(t *testing.T) {
	composer := &PromptComposer{}

	tests := []struct {
		name    string
		config  *PromptConfig
		wantErr bool
	}{
		{
			name: "valid config with message",
			config: &PromptConfig{
				Message: "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with image",
			config: &PromptConfig{
				Image: "test.jpg",
			},
			wantErr: false,
		},
		{
			name: "valid config with keyboard",
			config: &PromptConfig{
				Keyboard: func(ctx *Context) *PromptKeyboardBuilder { return nil },
			},
			wantErr: false,
		},
		{
			name: "invalid config - all nil",
			config: &PromptConfig{
				Message:  nil,
				Image:    nil,
				Keyboard: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := composer.validatePromptConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePromptConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptComposer_ComposeAndSend_TextOnly(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Message: "Hello World!",
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	msgConfig, ok := mockClient.sentMessages[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Fatalf("Expected MessageConfig, got %T", mockClient.sentMessages[0])
	}

	if msgConfig.Text != "Hello World!" {
		t.Errorf("Expected text 'Hello World!', got '%s'", msgConfig.Text)
	}

	if msgConfig.ChatID != ctx.ChatID() {
		t.Errorf("Expected ChatID %d, got %d", ctx.ChatID(), msgConfig.ChatID)
	}
}

func TestPromptComposer_ComposeAndSend_WithTemplate(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{
		renderFunc: func(name string, data map[string]interface{}) (string, ParseMode, error) {
			if name == "greeting" {
				if nameVal, ok := data["name"].(string); ok {
					return "Hello " + nameVal + "!", ParseModeHTML, nil
				}
				return "Hello World!", ParseModeHTML, nil
			}
			return "Unknown template", ParseModeNone, nil
		},
	}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Message: "template:greeting",
		TemplateData: map[string]interface{}{
			"name": "John",
		},
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	msgConfig, ok := mockClient.sentMessages[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Fatalf("Expected MessageConfig, got %T", mockClient.sentMessages[0])
	}

	if msgConfig.Text != "Hello John!" {
		t.Errorf("Expected text 'Hello John!', got '%s'", msgConfig.Text)
	}

	if msgConfig.ParseMode != string(ParseModeHTML) {
		t.Errorf("Expected ParseMode '%s', got '%s'", ParseModeHTML, msgConfig.ParseMode)
	}
}

func TestPromptComposer_ComposeAndSend_WithImage_ByteData(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	// Test with byte data
	imageData := []byte("fake image data")
	config := &PromptConfig{
		Message: "Image caption",
		Image:   imageData,
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	photoConfig, ok := mockClient.sentMessages[0].(tgbotapi.PhotoConfig)
	if !ok {
		t.Fatalf("Expected PhotoConfig, got %T", mockClient.sentMessages[0])
	}

	if photoConfig.Caption != "Image caption" {
		t.Errorf("Expected caption 'Image caption', got '%s'", photoConfig.Caption)
	}

	if photoConfig.ChatID != ctx.ChatID() {
		t.Errorf("Expected ChatID %d, got %d", ctx.ChatID(), photoConfig.ChatID)
	}
}

func TestPromptComposer_ComposeAndSend_WithImage_URL(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Message: "URL Image caption",
		Image:   "https://example.com/image.jpg",
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	photoConfig, ok := mockClient.sentMessages[0].(tgbotapi.PhotoConfig)
	if !ok {
		t.Fatalf("Expected PhotoConfig, got %T", mockClient.sentMessages[0])
	}

	if photoConfig.Caption != "URL Image caption" {
		t.Errorf("Expected caption 'URL Image caption', got '%s'", photoConfig.Caption)
	}
}

func TestPromptComposer_ComposeAndSend_WithKeyboard(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Message: "Choose an option:",
		Keyboard: func(ctx *Context) *PromptKeyboardBuilder {
			return NewPromptKeyboard().
				ButtonCallback("Option 1", "opt1").
				ButtonCallback("Option 2", "opt2")
		},
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	msgConfig, ok := mockClient.sentMessages[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Fatalf("Expected MessageConfig, got %T", mockClient.sentMessages[0])
	}

	if msgConfig.Text != "Choose an option:" {
		t.Errorf("Expected text 'Choose an option:', got '%s'", msgConfig.Text)
	}

	if msgConfig.ReplyMarkup == nil {
		t.Error("Expected ReplyMarkup to be set")
	}
}

func TestPromptComposer_ComposeAndSend_KeyboardOnly(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Keyboard: func(ctx *Context) *PromptKeyboardBuilder {
			return NewPromptKeyboard().
				ButtonCallback("Only Button", "only")
		},
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	msgConfig, ok := mockClient.sentMessages[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Fatalf("Expected MessageConfig, got %T", mockClient.sentMessages[0])
	}

	// Should send invisible character when only keyboard
	if msgConfig.Text != "\u200B" {
		t.Errorf("Expected invisible character, got '%s'", msgConfig.Text)
	}

	if msgConfig.ReplyMarkup == nil {
		t.Error("Expected ReplyMarkup to be set")
	}
}

func TestPromptComposer_ComposeAndSend_ImageWithKeyboard(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Message: "Image with buttons",
		Image:   []byte("fake image"),
		Keyboard: func(ctx *Context) *PromptKeyboardBuilder {
			return NewPromptKeyboard().
				ButtonCallback("Action", "action")
		},
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	photoConfig, ok := mockClient.sentMessages[0].(tgbotapi.PhotoConfig)
	if !ok {
		t.Fatalf("Expected PhotoConfig, got %T", mockClient.sentMessages[0])
	}

	if photoConfig.Caption != "Image with buttons" {
		t.Errorf("Expected caption 'Image with buttons', got '%s'", photoConfig.Caption)
	}

	if photoConfig.ReplyMarkup == nil {
		t.Error("Expected ReplyMarkup to be set")
	}
}

func TestPromptComposer_ComposeAndSend_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupClient   func() *mockTelegramClient
		setupTM       func() *mockTemplateManager
		config        *PromptConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "telegram send error",
			setupClient: func() *mockTelegramClient {
				return &mockTelegramClient{
					sendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
						return tgbotapi.Message{}, errors.New("send failed")
					},
				}
			},
			setupTM: func() *mockTemplateManager {
				return &mockTemplateManager{}
			},
			config: &PromptConfig{
				Message: "test",
			},
			expectError:   true,
			errorContains: "send failed",
		},
		{
			name: "template rendering error",
			setupClient: func() *mockTelegramClient {
				return &mockTelegramClient{}
			},
			setupTM: func() *mockTemplateManager {
				return &mockTemplateManager{
					renderFunc: func(name string, data map[string]interface{}) (string, ParseMode, error) {
						return "", ParseModeNone, errors.New("template render failed")
					},
				}
			},
			config: &PromptConfig{
				Message: "template:bad",
			},
			expectError:   true,
			errorContains: "message rendering failed",
		},
		{
			name: "invalid prompt config",
			setupClient: func() *mockTelegramClient {
				return &mockTelegramClient{}
			},
			setupTM: func() *mockTemplateManager {
				return &mockTemplateManager{}
			},
			config: &PromptConfig{
				// All fields nil
			},
			expectError:   true,
			errorContains: "invalid PromptConfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupClient()
			mockTM := tt.setupTM()
			composer := createTestPromptComposer(mockClient, mockTM)
			ctx := createTestContext()

			err := composer.ComposeAndSend(ctx, tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestPromptComposer_ComposeAndSend_FunctionMessage(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Message: func(ctx *Context) string {
			return "Dynamic message for user " + string(rune(ctx.UserID()))
		},
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	msgConfig, ok := mockClient.sentMessages[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Fatalf("Expected MessageConfig, got %T", mockClient.sentMessages[0])
	}

	// Should contain dynamic content
	if msgConfig.Text == "" {
		t.Error("Expected non-empty dynamic message")
	}
}

func TestPromptComposer_ComposeAndSend_FunctionImage(t *testing.T) {
	mockClient := &mockTelegramClient{}
	mockTM := &mockTemplateManager{}

	composer := createTestPromptComposer(mockClient, mockTM)
	ctx := createTestContext()

	config := &PromptConfig{
		Message: "Dynamic image",
		Image: func(ctx *Context) []byte {
			return []byte("dynamic image data")
		},
	}

	err := composer.ComposeAndSend(ctx, config)
	if err != nil {
		t.Fatalf("ComposeAndSend failed: %v", err)
	}

	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockClient.sentMessages))
	}

	photoConfig, ok := mockClient.sentMessages[0].(tgbotapi.PhotoConfig)
	if !ok {
		t.Fatalf("Expected PhotoConfig, got %T", mockClient.sentMessages[0])
	}

	if photoConfig.Caption != "Dynamic image" {
		t.Errorf("Expected caption 'Dynamic image', got '%s'", photoConfig.Caption)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
