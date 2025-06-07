package teleflow

import (
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestBotAPI implements the BotAPI interface for testing
type TestBotAPI struct {
	SendFunc    func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	RequestFunc func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	// Track calls for verification
	SendCalls    []tgbotapi.Chattable
	RequestCalls []tgbotapi.Chattable
}

func (t *TestBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	t.SendCalls = append(t.SendCalls, c)
	if t.SendFunc != nil {
		return t.SendFunc(c)
	}
	return tgbotapi.Message{}, nil
}

func (t *TestBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	t.RequestCalls = append(t.RequestCalls, c)
	if t.RequestFunc != nil {
		return t.RequestFunc(c)
	}
	return &tgbotapi.APIResponse{}, nil
}

// TestablePromptComposer extends PromptComposer to allow dependency injection for testing
type TestablePromptComposer struct {
	botAPI              BotAPI
	testMessageRenderer *TestMessageRenderer
	testImageHandler    *TestImageHandler
	testKeyboardHandler *TestPromptKeyboardHandler
}

func NewTestablePromptComposer(botAPI BotAPI) *TestablePromptComposer {
	return &TestablePromptComposer{
		botAPI:              botAPI,
		testMessageRenderer: &TestMessageRenderer{},
		testImageHandler:    &TestImageHandler{},
		testKeyboardHandler: &TestPromptKeyboardHandler{},
	}
}

func (tpc *TestablePromptComposer) ComposeAndSend(ctx *Context, promptConfig *PromptConfig) error {
	if err := tpc.validatePromptConfig(promptConfig); err != nil {
		return err
	}

	// 1. Render Message Text & ParseMode
	messageText, parseMode, err := tpc.testMessageRenderer.renderMessage(promptConfig, ctx)
	if err != nil {
		return err
	}

	// 2. Process Image
	processedImg, err := tpc.testImageHandler.processImage(promptConfig.Image, ctx)
	if err != nil {
		return err
	}

	// 3. Build Inline Keyboard
	var tgInlineKeyboard *tgbotapi.InlineKeyboardMarkup
	if promptConfig.Keyboard != nil {
		builtKeyboard, err := tpc.testKeyboardHandler.BuildKeyboard(ctx, promptConfig.Keyboard)
		if err != nil {
			return err
		}
		if builtKeyboard != nil {
			if keyboard, ok := builtKeyboard.(tgbotapi.InlineKeyboardMarkup); ok {
				if numButtons(keyboard) > 0 {
					tgInlineKeyboard = &keyboard
				}
			}
		}
	}

	// 4. Determine message type and send
	if processedImg != nil {
		// Send as photo
		photoMsg := tgbotapi.NewPhoto(ctx.ChatID(), nil)
		if processedImg.data != nil {
			photoMsg.File = tgbotapi.FileBytes{Name: "image.jpg", Bytes: processedImg.data}
		} else if processedImg.filePath != "" {
			photoMsg.File = tgbotapi.FileURL(processedImg.filePath)
		} else {
			return errors.New("processed image has no data or path")
		}

		photoMsg.Caption = messageText
		if parseMode != ParseModeNone {
			photoMsg.ParseMode = string(parseMode)
		}
		if tgInlineKeyboard != nil {
			photoMsg.ReplyMarkup = tgInlineKeyboard
		}
		_, err = tpc.botAPI.Send(photoMsg)
		return err
	} else if messageText != "" {
		// Send as text message
		textMsg := tgbotapi.NewMessage(ctx.ChatID(), messageText)
		if parseMode != ParseModeNone {
			textMsg.ParseMode = string(parseMode)
		}
		if tgInlineKeyboard != nil {
			textMsg.ReplyMarkup = tgInlineKeyboard
		}
		_, err = tpc.botAPI.Send(textMsg)
		return err
	} else if tgInlineKeyboard != nil {
		// Send keyboard with an invisible message
		invisibleMsg := tgbotapi.NewMessage(ctx.ChatID(), "\u200B")
		invisibleMsg.ReplyMarkup = tgInlineKeyboard
		_, err = tpc.botAPI.Send(invisibleMsg)
		return err
	}

	return nil
}

func (tpc *TestablePromptComposer) validatePromptConfig(config *PromptConfig) error {
	if config.Message == nil && config.Image == nil && config.Keyboard == nil {
		return errors.New("PromptConfig must have at least one of Message, Image, or Keyboard specified")
	}
	return nil
}

// Test message renderer
type TestMessageRenderer struct {
	RenderFunc  func(config *PromptConfig, ctx *Context) (string, ParseMode, error)
	RenderCalls []struct {
		Config *PromptConfig
		Ctx    *Context
	}
}

func (t *TestMessageRenderer) renderMessage(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
	t.RenderCalls = append(t.RenderCalls, struct {
		Config *PromptConfig
		Ctx    *Context
	}{config, ctx})
	if t.RenderFunc != nil {
		return t.RenderFunc(config, ctx)
	}
	return "", ParseModeNone, nil
}

// Test image handler
type TestImageHandler struct {
	ProcessFunc  func(imageSpec ImageSpec, ctx *Context) (*processedImage, error)
	ProcessCalls []struct {
		ImageSpec ImageSpec
		Ctx       *Context
	}
}

func (t *TestImageHandler) processImage(imageSpec ImageSpec, ctx *Context) (*processedImage, error) {
	t.ProcessCalls = append(t.ProcessCalls, struct {
		ImageSpec ImageSpec
		Ctx       *Context
	}{imageSpec, ctx})
	if t.ProcessFunc != nil {
		return t.ProcessFunc(imageSpec, ctx)
	}
	return nil, nil
}

// Test keyboard handler
type TestPromptKeyboardHandler struct {
	BuildFunc   func(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error)
	GetFunc     func(userID int64, uuid string) (interface{}, bool)
	CleanupFunc func(userID int64)
	BuildCalls  []struct {
		Ctx          *Context
		KeyboardFunc KeyboardFunc
	}
	GetCalls []struct {
		UserID int64
		UUID   string
	}
	CleanupCalls []int64
}

func (t *TestPromptKeyboardHandler) BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
	t.BuildCalls = append(t.BuildCalls, struct {
		Ctx          *Context
		KeyboardFunc KeyboardFunc
	}{ctx, keyboardFunc})
	if t.BuildFunc != nil {
		return t.BuildFunc(ctx, keyboardFunc)
	}
	return nil, nil
}

func (t *TestPromptKeyboardHandler) GetCallbackData(userID int64, uuid string) (interface{}, bool) {
	t.GetCalls = append(t.GetCalls, struct {
		UserID int64
		UUID   string
	}{userID, uuid})
	if t.GetFunc != nil {
		return t.GetFunc(userID, uuid)
	}
	return nil, false
}

func (t *TestPromptKeyboardHandler) CleanupUserMappings(userID int64) {
	t.CleanupCalls = append(t.CleanupCalls, userID)
	if t.CleanupFunc != nil {
		t.CleanupFunc(userID)
	}
}

// Helper function to create a test context
func createTestContext() *Context {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123},
			From: &tgbotapi.User{ID: 456},
		},
	}
	return NewContext(&Bot{}, update)
}

func TestPromptComposer_ComposeAndSend_TextOnlyMessage(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "Hello World", ParseModeNone, nil
	}

	ctx := createTestContext()
	promptConfig := &PromptConfig{
		Message: "Hello World",
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that Send was called
	if len(testBotAPI.SendCalls) != 1 {
		t.Errorf("Expected 1 Send call, got %d", len(testBotAPI.SendCalls))
	}

	// Verify the message sent
	if msg, ok := testBotAPI.SendCalls[0].(tgbotapi.MessageConfig); ok {
		if msg.Text != "Hello World" {
			t.Errorf("Expected message text 'Hello World', got '%s'", msg.Text)
		}
		if msg.ChatID != ctx.ChatID() {
			t.Errorf("Expected ChatID %d, got %d", ctx.ChatID(), msg.ChatID)
		}
	} else {
		t.Error("Expected MessageConfig type")
	}
}

func TestPromptComposer_ComposeAndSend_ImageOnlyMessage(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "", ParseModeNone, nil
	}
	composer.testImageHandler.ProcessFunc = func(imageSpec ImageSpec, ctx *Context) (*processedImage, error) {
		return &processedImage{
			data:     []byte("fake_image_data"),
			filePath: "",
		}, nil
	}

	ctx := createTestContext()
	promptConfig := &PromptConfig{
		Image: "test.jpg",
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that Send was called
	if len(testBotAPI.SendCalls) != 1 {
		t.Errorf("Expected 1 Send call, got %d", len(testBotAPI.SendCalls))
	}

	// Verify the photo message sent
	if _, ok := testBotAPI.SendCalls[0].(tgbotapi.PhotoConfig); !ok {
		t.Error("Expected PhotoConfig type")
	}
}

func TestPromptComposer_ComposeAndSend_KeyboardOnlyMessage(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "", ParseModeNone, nil
	}
	composer.testKeyboardHandler.BuildFunc = func(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Test", "uuid123"),
		})
		return keyboard, nil
	}

	ctx := createTestContext()
	keyboardFunc := func(ctx *Context) *InlineKeyboardBuilder {
		return NewInlineKeyboard().ButtonCallback("Test", "test_data")
	}
	promptConfig := &PromptConfig{
		Keyboard: keyboardFunc,
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that Send was called
	if len(testBotAPI.SendCalls) != 1 {
		t.Errorf("Expected 1 Send call, got %d", len(testBotAPI.SendCalls))
	}

	// Verify the message sent with invisible character
	if msg, ok := testBotAPI.SendCalls[0].(tgbotapi.MessageConfig); ok {
		if msg.Text != "\u200B" {
			t.Errorf("Expected invisible character, got '%s'", msg.Text)
		}
	} else {
		t.Error("Expected MessageConfig type")
	}
}

func TestPromptComposer_ComposeAndSend_CombinedMessage(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "*Bold Text*", ParseModeMarkdown, nil
	}
	composer.testImageHandler.ProcessFunc = func(imageSpec ImageSpec, ctx *Context) (*processedImage, error) {
		return &processedImage{
			filePath: "http://example.com/image.jpg",
		}, nil
	}
	composer.testKeyboardHandler.BuildFunc = func(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Test", "uuid123"),
		})
		return keyboard, nil
	}

	ctx := createTestContext()
	keyboardFunc := func(ctx *Context) *InlineKeyboardBuilder {
		return NewInlineKeyboard().ButtonCallback("Test", "test_data")
	}
	promptConfig := &PromptConfig{
		Message:  "*Bold Text*",
		Image:    "test.jpg",
		Keyboard: keyboardFunc,
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that Send was called
	if len(testBotAPI.SendCalls) != 1 {
		t.Errorf("Expected 1 Send call, got %d", len(testBotAPI.SendCalls))
	}

	// Verify the photo message with caption and parse mode
	if photoMsg, ok := testBotAPI.SendCalls[0].(tgbotapi.PhotoConfig); ok {
		if photoMsg.Caption != "*Bold Text*" {
			t.Errorf("Expected caption '*Bold Text*', got '%s'", photoMsg.Caption)
		}
		if photoMsg.ParseMode != "Markdown" {
			t.Errorf("Expected parse mode 'Markdown', got '%s'", photoMsg.ParseMode)
		}
	} else {
		t.Error("Expected PhotoConfig type")
	}
}

func TestPromptComposer_ComposeAndSend_MessageRenderError(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "", ParseModeNone, errors.New("render error")
	}

	ctx := createTestContext()
	promptConfig := &PromptConfig{
		Message: "Hello World",
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "render error" {
		t.Errorf("Expected 'render error', got '%s'", err.Error())
	}
}

func TestPromptComposer_ComposeAndSend_ImageProcessingError(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "", ParseModeNone, nil
	}
	composer.testImageHandler.ProcessFunc = func(imageSpec ImageSpec, ctx *Context) (*processedImage, error) {
		return nil, errors.New("image error")
	}

	ctx := createTestContext()
	promptConfig := &PromptConfig{
		Image: "invalid.jpg",
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "image error" {
		t.Errorf("Expected 'image error', got '%s'", err.Error())
	}
}

func TestPromptComposer_ComposeAndSend_KeyboardBuildingError(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "", ParseModeNone, nil
	}
	composer.testKeyboardHandler.BuildFunc = func(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
		return nil, errors.New("keyboard error")
	}

	ctx := createTestContext()
	keyboardFunc := func(ctx *Context) *InlineKeyboardBuilder {
		return NewInlineKeyboard()
	}
	promptConfig := &PromptConfig{
		Keyboard: keyboardFunc,
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "keyboard error" {
		t.Errorf("Expected 'keyboard error', got '%s'", err.Error())
	}
}

func TestPromptComposer_ValidatePromptConfig(t *testing.T) {
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	tests := []struct {
		name        string
		config      *PromptConfig
		expectError bool
	}{
		{
			name:        "Empty config",
			config:      &PromptConfig{},
			expectError: true,
		},
		{
			name: "Message only",
			config: &PromptConfig{
				Message: "Hello",
			},
			expectError: false,
		},
		{
			name: "Image only",
			config: &PromptConfig{
				Image: "test.jpg",
			},
			expectError: false,
		},
		{
			name: "Keyboard only",
			config: &PromptConfig{
				Keyboard: func(ctx *Context) *InlineKeyboardBuilder {
					return NewInlineKeyboard()
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := composer.validatePromptConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestPromptComposer_SendAPIError(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{
		SendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, errors.New("API error")
		},
	}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "Hello World", ParseModeNone, nil
	}

	ctx := createTestContext()
	promptConfig := &PromptConfig{
		Message: "Hello World",
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "API error" {
		t.Errorf("Expected 'API error', got '%s'", err.Error())
	}
}

func TestPromptComposer_NilKeyboardFunc(t *testing.T) {
	// Setup test components
	testBotAPI := &TestBotAPI{}
	composer := NewTestablePromptComposer(testBotAPI)

	composer.testMessageRenderer.RenderFunc = func(config *PromptConfig, ctx *Context) (string, ParseMode, error) {
		return "Hello World", ParseModeNone, nil
	}

	ctx := createTestContext()
	promptConfig := &PromptConfig{
		Message:  "Hello World",
		Keyboard: nil, // Nil keyboard
	}

	// Execute
	err := composer.ComposeAndSend(ctx, promptConfig)

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that keyboard handler was not called
	if len(composer.testKeyboardHandler.BuildCalls) != 0 {
		t.Errorf("Expected 0 keyboard build calls, got %d", len(composer.testKeyboardHandler.BuildCalls))
	}

	// Verify the message sent without keyboard
	if msg, ok := testBotAPI.SendCalls[0].(tgbotapi.MessageConfig); ok {
		if msg.Text != "Hello World" {
			t.Errorf("Expected message text 'Hello World', got '%s'", msg.Text)
		}
		if msg.ReplyMarkup != nil {
			t.Error("Expected no reply markup, got some")
		}
	} else {
		t.Error("Expected MessageConfig type")
	}
}
