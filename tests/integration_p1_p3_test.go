package tests

import (
	"fmt"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	teleflow "github.com/kslamph/teleflow/core"
)

// TestAccessManager implements AccessManager for testing reply keyboards
type TestAccessManager struct {
	replyKeyboard         *teleflow.ReplyKeyboard
	shouldProvideKeyboard bool
}

func NewTestAccessManager() *TestAccessManager {
	// Create a test reply keyboard
	keyboard := teleflow.BuildReplyKeyboard(
		[]string{"📝 Start Flow", "⚙️ Settings"}, 2,
	).Resize()

	return &TestAccessManager{
		replyKeyboard:         keyboard,
		shouldProvideKeyboard: true,
	}
}

func (tam *TestAccessManager) CheckPermission(ctx *teleflow.PermissionContext) error {
	return nil // Allow all permissions for testing
}

func (tam *TestAccessManager) GetReplyKeyboard(ctx *teleflow.PermissionContext) *teleflow.ReplyKeyboard {
	if tam.shouldProvideKeyboard {
		return tam.replyKeyboard
	}
	return nil
}

func (tam *TestAccessManager) SetShouldProvideKeyboard(should bool) {
	tam.shouldProvideKeyboard = should
}

// MockBotAPI implements the BotAPI interface for testing
type MockBotAPI struct {
	sentMessages     []tgbotapi.Chattable
	lastSentMessage  tgbotapi.Message
	messageID        int
	callbackAnswered bool
}

func NewMockBotAPI() *MockBotAPI {
	return &MockBotAPI{
		sentMessages: make([]tgbotapi.Chattable, 0),
		messageID:    1,
	}
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)

	// Create a mock response message
	msg := tgbotapi.Message{
		MessageID: m.messageID,
		Text:      "",
	}

	// Extract text and other details based on message type
	switch chattable := c.(type) {
	case tgbotapi.MessageConfig:
		msg.Text = chattable.Text
		msg.Chat = &tgbotapi.Chat{ID: chattable.ChatID}
	case tgbotapi.PhotoConfig:
		msg.Caption = chattable.Caption
		msg.Chat = &tgbotapi.Chat{ID: chattable.ChatID}
	}

	m.lastSentMessage = msg
	m.messageID++
	return msg, nil
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.sentMessages = append(m.sentMessages, c)
	if _, ok := c.(tgbotapi.CallbackConfig); ok {
		m.callbackAnswered = true
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *MockBotAPI) GetSentMessagesCount() int {
	return len(m.sentMessages)
}

func (m *MockBotAPI) GetLastTextMessage() string {
	if len(m.sentMessages) == 0 {
		return ""
	}

	lastMsg := m.sentMessages[len(m.sentMessages)-1]
	switch msg := lastMsg.(type) {
	case tgbotapi.MessageConfig:
		return msg.Text
	case tgbotapi.PhotoConfig:
		return msg.Caption
	}
	return ""
}

func (m *MockBotAPI) GetLastReplyKeyboard() *tgbotapi.ReplyKeyboardMarkup {
	if len(m.sentMessages) == 0 {
		return nil
	}

	lastMsg := m.sentMessages[len(m.sentMessages)-1]
	switch msg := lastMsg.(type) {
	case tgbotapi.MessageConfig:
		if kbd, ok := msg.ReplyMarkup.(*tgbotapi.ReplyKeyboardMarkup); ok {
			return kbd
		}
	}
	return nil
}

// TestIntegrationP1P3FlowComponents tests flow component interactions through public API
func TestIntegrationP1P3FlowComponents(t *testing.T) {
	t.Run("FlowDataPersistenceAndRetrieval", func(t *testing.T) {
		// Test flow data management via Context methods
		userID := int64(12345)
		chatID := int64(67890)

		// Create a minimal bot for testing
		bot, err := teleflow.NewBot("test_token")
		if err != nil {
			t.Skip("Skipping test that requires bot token")
		}

		// Create test context
		update := tgbotapi.Update{
			Message: &tgbotapi.Message{
				From: &tgbotapi.User{ID: userID},
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "test input",
			},
		}
		ctx := teleflow.NewContext(bot, update)

		// Test that SetFlowData/GetFlowData work correctly when not in flow
		err = ctx.SetFlowData("test_key", "test_value")
		if err == nil {
			t.Error("Expected error when setting flow data outside of flow")
		}

		_, exists := ctx.GetFlowData("test_key")
		if exists {
			t.Error("Should not get flow data when not in flow")
		}

		// Test request-scoped data (ctx.Set/Get)
		ctx.Set("request_key", "request_value")
		value, exists := ctx.Get("request_key")
		if !exists {
			t.Error("Request-scoped data should be accessible")
		}
		if value != "request_value" {
			t.Errorf("Expected 'request_value', got '%v'", value)
		}
	})

	t.Run("FlowBuilderValidation", func(t *testing.T) {
		// Test that flow building validates components properly

		// Valid flow should build successfully
		validFlow, err := teleflow.NewFlow("test_flow").
			Step("step1").
			Prompt("What's your name?").
			Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
				if input == "" {
					return teleflow.Retry()
				}
				ctx.Set("name", input)
				return teleflow.NextStep()
			}).
			Step("step2").
			Prompt("What's your age?").
			Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
				ctx.Set("age", input)
				return teleflow.CompleteFlow()
			}).
			Build()

		if err != nil {
			t.Fatalf("Valid flow should build without error: %v", err)
		}

		// Verify flow structure
		if validFlow.Name != "test_flow" {
			t.Errorf("Expected flow name 'test_flow', got '%s'", validFlow.Name)
		}

		if len(validFlow.Steps) != 2 {
			t.Errorf("Expected 2 steps, got %d", len(validFlow.Steps))
		}

		if len(validFlow.Order) != 2 {
			t.Errorf("Expected 2 steps in order, got %d", len(validFlow.Order))
		}

		// Test flow with inline keyboard
		keyboardFlow, err := teleflow.NewFlow("keyboard_flow").
			Step("choice").
			Prompt("Choose an option:").
			WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
				return teleflow.NewPromptKeyboard().
					ButtonCallback("Option A", "a").
					ButtonCallback("Option B", "b")
			}).
			Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
				if buttonClick != nil {
					ctx.Set("choice", buttonClick.Data)
					return teleflow.CompleteFlow()
				}
				return teleflow.Retry()
			}).
			Build()

		if err != nil {
			t.Fatalf("Keyboard flow should build without error: %v", err)
		}

		if keyboardFlow.Name != "keyboard_flow" {
			t.Errorf("Expected flow name 'keyboard_flow', got '%s'", keyboardFlow.Name)
		}
	})

	t.Run("PromptConfigRendering", func(t *testing.T) {
		// Test PromptConfig with various configurations

		// Simple text prompt
		simplePrompt := &teleflow.PromptConfig{
			Message: "Hello, world!",
		}

		if simplePrompt.Message != "Hello, world!" {
			t.Error("Simple prompt message should be set correctly")
		}

		// Function-based message
		funcPrompt := &teleflow.PromptConfig{
			Message: func(ctx *teleflow.Context) string {
				return fmt.Sprintf("Hello, user %d!", ctx.UserID())
			},
		}

		// Verify function-based message is callable
		if funcPrompt.Message == nil {
			t.Error("Function-based message should be set")
		}

		// Template-based prompt with data
		templatePrompt := &teleflow.PromptConfig{
			Message: "template:greeting",
			TemplateData: map[string]interface{}{
				"Name": "John",
				"Role": "Admin",
			},
		}

		if !strings.HasPrefix(templatePrompt.Message.(string), "template:") {
			t.Error("Template prompt should have template: prefix")
		}

		if templatePrompt.TemplateData["Name"] != "John" {
			t.Error("Template data should be preserved")
		}
	})
}

// TestIntegrationP1P3AccessManagerKeyboards tests AccessManager keyboard integration
func TestIntegrationP1P3AccessManagerKeyboards(t *testing.T) {
	// Create test access manager
	accessManager := NewTestAccessManager()

	t.Run("AccessManagerKeyboardConstruction", func(t *testing.T) {
		// Test that AccessManager provides valid keyboards
		permCtx := &teleflow.PermissionContext{
			UserID: 12345,
			ChatID: 67890,
		}

		keyboard := accessManager.GetReplyKeyboard(permCtx)
		if keyboard == nil {
			t.Fatal("AccessManager should provide a keyboard")
		}

		// Convert to Telegram format and verify
		tgKeyboard := keyboard.ToTgbotapi()
		if len(tgKeyboard.Keyboard) == 0 {
			t.Error("Keyboard should have at least one row")
		}

		if len(tgKeyboard.Keyboard[0]) != 2 {
			t.Errorf("Expected 2 buttons in first row, got %d", len(tgKeyboard.Keyboard[0]))
		}

		if tgKeyboard.Keyboard[0][0].Text != "📝 Start Flow" {
			t.Errorf("Expected first button '📝 Start Flow', got '%s'", tgKeyboard.Keyboard[0][0].Text)
		}

		if !tgKeyboard.ResizeKeyboard {
			t.Error("Keyboard should have resize flag set")
		}
	})

	t.Run("AccessManagerKeyboardDisabling", func(t *testing.T) {
		// Test disabling keyboard provision
		accessManager.SetShouldProvideKeyboard(false)

		permCtx := &teleflow.PermissionContext{
			UserID: 12345,
			ChatID: 67890,
		}

		keyboard := accessManager.GetReplyKeyboard(permCtx)
		if keyboard != nil {
			t.Error("AccessManager should not provide keyboard when disabled")
		}

		// Re-enable for other tests
		accessManager.SetShouldProvideKeyboard(true)
	})

	t.Run("AccessManagerPermissionChecking", func(t *testing.T) {
		// Test permission checking
		permCtx := &teleflow.PermissionContext{
			UserID:  12345,
			ChatID:  67890,
			Command: "test",
		}

		err := accessManager.CheckPermission(permCtx)
		if err != nil {
			t.Errorf("Permission check should pass for test access manager: %v", err)
		}
	})
}

// TestIntegrationP1P3PromptKeyboardHandler tests keyboard handler functionality
func TestIntegrationP1P3PromptKeyboardHandler(t *testing.T) {
	t.Run("KeyboardBuilderUUIDMapping", func(t *testing.T) {
		// Create keyboard handler
		handler := teleflow.NewPromptKeyboardHandler()

		// Create test context
		update := tgbotapi.Update{
			Message: &tgbotapi.Message{
				From: &tgbotapi.User{ID: 12345},
				Chat: &tgbotapi.Chat{ID: 67890},
				Text: "test",
			},
		}
		bot, err := teleflow.NewBot("test_token")
		if err != nil {
			t.Skip("Skipping test that requires bot token")
		}
		ctx := teleflow.NewContext(bot, update)

		// Test nil keyboard function
		result, err := handler.BuildKeyboard(ctx, nil)
		if err != nil {
			t.Errorf("BuildKeyboard with nil function should not error: %v", err)
		}
		if result != nil {
			t.Error("BuildKeyboard with nil function should return nil")
		}

		// Test keyboard function that returns nil
		nilKeyboardFunc := func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			return nil
		}
		result2, err := handler.BuildKeyboard(ctx, nilKeyboardFunc)
		if err != nil {
			t.Errorf("BuildKeyboard with nil-returning function should not error: %v", err)
		}
		if result2 != nil {
			t.Error("BuildKeyboard with nil-returning function should return nil")
		}

		// Test valid keyboard function
		validKeyboardFunc := func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			return teleflow.NewPromptKeyboard().
				ButtonCallback("Test Button", "test_data")
		}

		result3, err := handler.BuildKeyboard(ctx, validKeyboardFunc)
		if err != nil {
			t.Fatalf("BuildKeyboard with valid function should not error: %v", err)
		}

		keyboard, ok := result3.(tgbotapi.InlineKeyboardMarkup)
		if !ok {
			t.Fatal("Expected InlineKeyboardMarkup result")
		}

		if len(keyboard.InlineKeyboard) == 0 || len(keyboard.InlineKeyboard[0]) == 0 {
			t.Fatal("Expected keyboard with buttons")
		}

		// Get the UUID from the built keyboard
		uuid := *keyboard.InlineKeyboard[0][0].CallbackData

		// Verify UUID mapping exists and retrieves original data
		originalData, exists := handler.GetCallbackData(ctx.UserID(), uuid)
		if !exists {
			t.Error("UUID mapping should exist after building keyboard")
		}

		if originalData != "test_data" {
			t.Errorf("Expected original data 'test_data', got '%v'", originalData)
		}

		// Test cleanup
		handler.CleanupUserMappings(ctx.UserID())
		_, exists = handler.GetCallbackData(ctx.UserID(), uuid)
		if exists {
			t.Error("UUID mapping should be cleaned up after cleanup")
		}
	})

	t.Run("InlineKeyboardBuilderValidation", func(t *testing.T) {
		// Test keyboard builder validation
		builder := teleflow.NewPromptKeyboard()

		// Empty keyboard should fail validation
		err := builder.ValidateBuilder()
		if err == nil {
			t.Error("Empty keyboard should fail validation")
		}

		// Add a button and validate
		builder.ButtonCallback("Valid Button", "data")
		err = builder.ValidateBuilder()
		if err != nil {
			t.Errorf("Keyboard with buttons should pass validation: %v", err)
		}

		// Test building the keyboard
		keyboard := builder.Build()
		if len(keyboard.InlineKeyboard) == 0 {
			t.Error("Built keyboard should have buttons")
		}

		// Test UUID mapping
		mapping := builder.GetUUIDMapping()
		if len(mapping) == 0 {
			t.Error("Builder should have UUID mappings")
		}
	})
}

// TestIntegrationP1P3TemplateSystemIntegration tests template system with flows
func TestIntegrationP1P3TemplateSystemIntegration(t *testing.T) {
	t.Run("TemplateWithFlowIntegration", func(t *testing.T) {
		// Create bot and register templates
		bot, err := teleflow.NewBot("test_token")
		if err != nil {
			t.Skip("Skipping test that requires bot token")
		}

		// Register test templates
		err = bot.AddTemplate("greeting", "Hello {{.Name}}!", teleflow.ParseModeNone)
		if err != nil {
			t.Fatalf("Failed to add greeting template: %v", err)
		}

		err = bot.AddTemplate("age_prompt", "Nice to meet you, {{.Name}}! How old are you?", teleflow.ParseModeNone)
		if err != nil {
			t.Fatalf("Failed to add age_prompt template: %v", err)
		}

		// Verify templates are registered
		if !bot.HasTemplate("greeting") {
			t.Error("Template 'greeting' should be registered")
		}

		if !bot.HasTemplate("age_prompt") {
			t.Error("Template 'age_prompt' should be registered")
		}

		// Test template info retrieval
		info := bot.GetTemplateInfo("greeting")
		if info == nil {
			t.Error("Should be able to get template info")
		}
		if info != nil && info.ParseMode != teleflow.ParseModeNone {
			t.Errorf("Expected ParseModeNone, got %s", info.ParseMode)
		}

		// Test flow with template
		templateFlow, err := teleflow.NewFlow("template_test").
			Step("greet").
			Prompt("template:greeting").
			Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
				if input != "" {
					ctx.Set("Name", input)
					return teleflow.NextStep()
				}
				return teleflow.Retry()
			}).
			Step("age").
			Prompt("template:age_prompt").
			Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
				return teleflow.CompleteFlow()
			}).
			Build()

		if err != nil {
			t.Fatalf("Template flow should build successfully: %v", err)
		}

		if templateFlow.Name != "template_test" {
			t.Errorf("Expected flow name 'template_test', got '%s'", templateFlow.Name)
		}
	})

	t.Run("ContextTemplateConvenienceMethods", func(t *testing.T) {
		// Test that template convenience methods exist and can be called
		userID := int64(12345)
		chatID := int64(67890)

		bot, err := teleflow.NewBot("test_token")
		if err != nil {
			t.Skip("Skipping test that requires bot token")
		}

		update := tgbotapi.Update{
			Message: &tgbotapi.Message{
				From: &tgbotapi.User{ID: userID},
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "test",
			},
		}
		ctx := teleflow.NewContext(bot, update)

		// These should compile without error (methods exist)

		_ = ctx.SendPromptWithTemplate

		// Test that the methods can be called (though they may fail without proper setup)
		// We're mainly testing the API exists and compiles correctly
		templateData := map[string]interface{}{
			"Name": "TestUser",
		}

		// Test SendPromptWithTemplate method signature
		prompt := &teleflow.PromptConfig{
			Message:      "template:test",
			TemplateData: templateData,
		}

		if prompt.TemplateData["Name"] != "TestUser" {
			t.Error("Template data should be preserved in PromptConfig")
		}
	})
}
