package teleflow

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Helper function to capture log output
// Helper function to capture log output
func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	f()
	return buf.String()
}

// Mock handler for testing
type mockHandler struct {
	called    bool
	err       error
	sleepTime time.Duration
}

func (m *mockHandler) Handle(ctx *Context) error {
	m.called = true
	if m.sleepTime > 0 {
		time.Sleep(m.sleepTime)
	}
	return m.err
}

// Helper function to create a test context with different update types
func createMiddlewareTestContext(updateType string, userID int64) *Context {
	var update tgbotapi.Update

	switch updateType {
	case "message":
		update = tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From:      &tgbotapi.User{ID: userID},
				Chat:      &tgbotapi.Chat{ID: userID},
				Text:      "Hello world",
			},
		}
	case "command":
		update = tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From:      &tgbotapi.User{ID: userID},
				Chat:      &tgbotapi.Chat{ID: userID},
				Text:      "/start test args",
				Entities: []tgbotapi.MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 6},
				},
			},
		}
	case "callback":
		update = tgbotapi.Update{
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID:   "callback123",
				From: &tgbotapi.User{ID: userID},
				Data: "button_clicked",
			},
		}
	case "long_text":
		longText := strings.Repeat("This is a very long message that should be truncated. ", 10)
		update = tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From:      &tgbotapi.User{ID: userID},
				Chat:      &tgbotapi.Chat{ID: userID},
				Text:      longText,
			},
		}
	}

	return newContext(
		update,
		&contextMockTelegramClient{},
		&contextMockTemplateManager{},
		&contextMockFlowOperations{},
		&contextMockPromptSender{},
		nil,
	)
}

func TestLoggingMiddleware_BasicFlow(t *testing.T) {
	// Create middleware
	middleware := LoggingMiddleware()

	// Create mock handler
	mockHandler := &mockHandler{}

	// Create test context
	ctx := createMiddlewareTestContext("message", 123)

	// Wrap handler with middleware
	wrappedHandler := middleware(mockHandler.Handle)

	// Capture log output
	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify log content
	if !strings.Contains(logOutput, "[INFO][123]") {
		t.Errorf("Expected INFO log for user 123, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "text: Hello world") {
		t.Errorf("Expected message text in log, got: %s", logOutput)
	}
}

func TestLoggingMiddleware_CommandUpdate(t *testing.T) {
	middleware := LoggingMiddleware()
	mockHandler := &mockHandler{}
	ctx := createMiddlewareTestContext("command", 456)

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify command logging
	if !strings.Contains(logOutput, "[INFO][456]") {
		t.Errorf("Expected INFO log for user 456, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "command: start") {
		t.Errorf("Expected command name in log, got: %s", logOutput)
	}
}

func TestLoggingMiddleware_CallbackUpdate(t *testing.T) {
	middleware := LoggingMiddleware()
	mockHandler := &mockHandler{}
	ctx := createMiddlewareTestContext("callback", 789)

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify callback logging
	if !strings.Contains(logOutput, "[INFO][789]") {
		t.Errorf("Expected INFO log for user 789, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "callback: button_clicked") {
		t.Errorf("Expected callback data in log, got: %s", logOutput)
	}
}

func TestLoggingMiddleware_LongTextTruncation(t *testing.T) {
	middleware := LoggingMiddleware()
	mockHandler := &mockHandler{}
	ctx := createMiddlewareTestContext("long_text", 999)

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Verify text is truncated
	if !strings.Contains(logOutput, "...") {
		t.Errorf("Expected text to be truncated with '...', got: %s", logOutput)
	}

	// Verify it doesn't exceed 100 characters + "text: " prefix + "..."
	logLines := strings.Split(logOutput, "\n")
	for _, line := range logLines {
		if strings.Contains(line, "text:") && strings.Contains(line, "...") {
			textPart := strings.Split(line, "text: ")[1]
			if len(textPart) > 103 { // 100 + "..."
				t.Errorf("Text not properly truncated, length: %d", len(textPart))
			}
		}
	}
}

func TestLoggingMiddleware_HandlerError(t *testing.T) {
	middleware := LoggingMiddleware()
	expectedError := errors.New("handler failed")
	mockHandler := &mockHandler{err: expectedError}
	ctx := createMiddlewareTestContext("message", 111)

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != expectedError {
			t.Errorf("Expected error to be returned, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify error logging
	if !strings.Contains(logOutput, "[ERROR][111]") {
		t.Errorf("Expected ERROR log for user 111, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Handler failed") {
		t.Errorf("Expected 'Handler failed' in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "handler failed") {
		t.Errorf("Expected error message in log, got: %s", logOutput)
	}
}

func TestLoggingMiddleware_ProcessingTime(t *testing.T) {
	middleware := LoggingMiddleware()
	sleepDuration := 10 * time.Millisecond
	mockHandler := &mockHandler{sleepTime: sleepDuration}
	ctx := createMiddlewareTestContext("message", 222)

	wrappedHandler := middleware(mockHandler.Handle)

	start := time.Now()
	_ = captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
	actualDuration := time.Since(start)

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Check that actual execution took at least the sleep time
	if actualDuration < sleepDuration {
		t.Errorf("Expected execution to take at least %v, took %v", sleepDuration, actualDuration)
	}

	// For error case (since no error occurred), there should be no completion log with default settings
	// The middleware only logs completion time on error or debug mode
}

func TestLoggingMiddleware_DebugMode(t *testing.T) {
	middleware := LoggingMiddleware()
	mockHandler := &mockHandler{}
	ctx := createMiddlewareTestContext("message", 333)

	// Set debug mode
	ctx.Set("debug", true)

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify debug logging
	if !strings.Contains(logOutput, "[DEBUG][333]") {
		t.Errorf("Expected DEBUG log for user 333, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Handler completed") {
		t.Errorf("Expected completion log in debug mode, got: %s", logOutput)
	}
}

func TestLoggingMiddleware_DebugLogLevel(t *testing.T) {
	middleware := LoggingMiddleware()
	mockHandler := &mockHandler{}
	ctx := createMiddlewareTestContext("message", 444)

	// Set log level to debug
	ctx.Set("logLevel", "debug")

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify debug logging via logLevel
	if !strings.Contains(logOutput, "[DEBUG][444]") {
		t.Errorf("Expected DEBUG log for user 444, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Handler completed") {
		t.Errorf("Expected completion log in debug mode, got: %s", logOutput)
	}
}

func TestLoggingMiddleware_ErrorWithDebugMode(t *testing.T) {
	middleware := LoggingMiddleware()
	expectedError := errors.New("test error")
	mockHandler := &mockHandler{err: expectedError}
	ctx := createMiddlewareTestContext("message", 555)

	// Set debug mode
	ctx.Set("debug", true)

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != expectedError {
			t.Errorf("Expected error to be returned, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify both debug and error logging
	if !strings.Contains(logOutput, "[DEBUG][555]") {
		t.Errorf("Expected DEBUG log for user 555, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "[ERROR][555]") {
		t.Errorf("Expected ERROR log for user 555, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "test error") {
		t.Errorf("Expected error message in log, got: %s", logOutput)
	}
}

func TestLoggingMiddleware_UnknownUpdateType(t *testing.T) {
	middleware := LoggingMiddleware()
	mockHandler := &mockHandler{}

	// Create context with no message or callback
	update := tgbotapi.Update{
		UpdateID: 1,
		// No Message or CallbackQuery
	}
	ctx := newContext(
		update,
		&contextMockTelegramClient{},
		&contextMockTemplateManager{},
		&contextMockFlowOperations{},
		&contextMockPromptSender{},
		nil,
	)

	wrappedHandler := middleware(mockHandler.Handle)

	logOutput := captureLogOutput(func() {
		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Verify handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called")
	}

	// Verify unknown update type logging
	if !strings.Contains(logOutput, "unknown") {
		t.Errorf("Expected 'unknown' update type in log, got: %s", logOutput)
	}
}

// Mock AccessManager for AuthMiddleware testing
type mockAccessManager struct {
	CheckPermissionCalls []*PermissionContext
	CheckPermissionFunc  func(ctx *PermissionContext) error
}

func (m *mockAccessManager) CheckPermission(ctx *PermissionContext) error {
	m.CheckPermissionCalls = append(m.CheckPermissionCalls, ctx)
	if m.CheckPermissionFunc != nil {
		return m.CheckPermissionFunc(ctx)
	}
	return nil
}

func (m *mockAccessManager) GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard {
	return nil
}

// Helper function to create context with mock access manager for middleware testing
func createAuthMiddlewareTestContext(updateType string, userID int64, chatID int64) (*Context, *contextMockTelegramClient) {
	var update tgbotapi.Update

	switch updateType {
	case "message":
		update = tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From:      &tgbotapi.User{ID: userID},
				Chat:      &tgbotapi.Chat{ID: chatID, Type: "private"},
				Text:      "Hello world",
			},
		}
	case "group_message":
		update = tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From:      &tgbotapi.User{ID: userID},
				Chat:      &tgbotapi.Chat{ID: chatID, Type: "group"},
				Text:      "Hello group",
			},
		}
	case "command":
		update = tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From:      &tgbotapi.User{ID: userID},
				Chat:      &tgbotapi.Chat{ID: chatID, Type: "private"},
				Text:      "/start test args",
				Entities: []tgbotapi.MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 6},
				},
			},
		}
	case "callback":
		update = tgbotapi.Update{
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID:   "callback123",
				From: &tgbotapi.User{ID: userID},
				Data: "button_clicked",
				Message: &tgbotapi.Message{
					MessageID: 1,
					Chat:      &tgbotapi.Chat{ID: chatID, Type: "private"},
				},
			},
		}
	case "callback_group":
		update = tgbotapi.Update{
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID:   "callback456",
				From: &tgbotapi.User{ID: userID},
				Data: "group_button",
				Message: &tgbotapi.Message{
					MessageID: 2,
					Chat:      &tgbotapi.Chat{ID: chatID, Type: "group"},
				},
			},
		}
	}

	mockClient := &contextMockTelegramClient{}
	ctx := newContext(
		update,
		mockClient,
		&contextMockTemplateManager{},
		&contextMockFlowOperations{},
		&contextMockPromptSender{},
		&mockAccessManager{},
	)

	return ctx, mockClient
}

func TestAuthMiddleware_Constructor(t *testing.T) {
	mockAM := &mockAccessManager{}
	middleware := AuthMiddleware(mockAM)

	if middleware == nil {
		t.Error("Expected AuthMiddleware to return a middleware function")
	}

	// Create a test handler to verify the middleware works
	testHandler := func(ctx *Context) error {
		return nil
	}

	wrappedHandler := middleware(testHandler)
	if wrappedHandler == nil {
		t.Error("Expected middleware to wrap the handler function")
	}
}

func TestAuthMiddleware_AccessAllowed_MessageUpdate(t *testing.T) {
	mockAM := &mockAccessManager{}
	middleware := AuthMiddleware(mockAM)

	mockHandler := &mockHandler{}
	ctx, mockClient := createAuthMiddlewareTestContext("message", 123, 456)

	wrappedHandler := middleware(mockHandler.Handle)
	err := wrappedHandler(ctx)

	// Verify no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify next handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called when access is allowed")
	}

	// Verify CheckPermission was called
	if len(mockAM.CheckPermissionCalls) != 1 {
		t.Errorf("Expected CheckPermission to be called once, got %d calls", len(mockAM.CheckPermissionCalls))
	}

	// Verify PermissionContext contents
	permCtx := mockAM.CheckPermissionCalls[0]
	if permCtx.UserID != 123 {
		t.Errorf("Expected UserID 123, got %d", permCtx.UserID)
	}
	if permCtx.ChatID != 456 {
		t.Errorf("Expected ChatID 456, got %d", permCtx.ChatID)
	}
	if permCtx.IsGroup {
		t.Error("Expected IsGroup to be false for private chat")
	}
	if permCtx.MessageID != 1 {
		t.Errorf("Expected MessageID 1, got %d", permCtx.MessageID)
	}

	// Verify no rejection message was sent
	if len(mockClient.SendCalls) != 0 {
		t.Errorf("Expected no messages to be sent when access is allowed, got %d", len(mockClient.SendCalls))
	}
}

func TestAuthMiddleware_AccessAllowed_CommandUpdate(t *testing.T) {
	mockAM := &mockAccessManager{}
	middleware := AuthMiddleware(mockAM)

	mockHandler := &mockHandler{}
	ctx, mockClient := createAuthMiddlewareTestContext("command", 789, 101112)

	wrappedHandler := middleware(mockHandler.Handle)
	err := wrappedHandler(ctx)

	// Verify no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify next handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called when access is allowed")
	}

	// Verify CheckPermission was called with command info
	if len(mockAM.CheckPermissionCalls) != 1 {
		t.Errorf("Expected CheckPermission to be called once, got %d calls", len(mockAM.CheckPermissionCalls))
	}

	// Verify PermissionContext contains command information
	permCtx := mockAM.CheckPermissionCalls[0]
	if permCtx.Command != "start" {
		t.Errorf("Expected Command 'start', got '%s'", permCtx.Command)
	}
	if len(permCtx.Arguments) != 1 || permCtx.Arguments[0] != "test args" {
		t.Errorf("Expected Arguments ['test args'], got %v", permCtx.Arguments)
	}

	// Verify no rejection message was sent
	if len(mockClient.SendCalls) != 0 {
		t.Errorf("Expected no messages to be sent when access is allowed, got %d", len(mockClient.SendCalls))
	}
}

func TestAuthMiddleware_AccessAllowed_GroupMessage(t *testing.T) {
	mockAM := &mockAccessManager{}
	middleware := AuthMiddleware(mockAM)

	mockHandler := &mockHandler{}
	ctx, _ := createAuthMiddlewareTestContext("group_message", 555, 666)

	wrappedHandler := middleware(mockHandler.Handle)
	err := wrappedHandler(ctx)

	// Verify no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify next handler was called
	if !mockHandler.called {
		t.Error("Expected next handler to be called when access is allowed")
	}

	// Verify PermissionContext has IsGroup set correctly
	if len(mockAM.CheckPermissionCalls) != 1 {
		t.Errorf("Expected CheckPermission to be called once, got %d calls", len(mockAM.CheckPermissionCalls))
	}

	permCtx := mockAM.CheckPermissionCalls[0]
	if !permCtx.IsGroup {
		t.Error("Expected IsGroup to be true for group chat")
	}
}

func TestAuthMiddleware_AccessDenied_MessageUpdate(t *testing.T) {
	mockAM := &mockAccessManager{
		CheckPermissionFunc: func(ctx *PermissionContext) error {
			return errors.New("Access denied")
		},
	}
	middleware := AuthMiddleware(mockAM)

	mockHandler := &mockHandler{}
	ctx, mockClient := createAuthMiddlewareTestContext("message", 123, 456)

	wrappedHandler := middleware(mockHandler.Handle)
	err := wrappedHandler(ctx)

	// Verify no error occurred (error is handled by sending rejection message)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify next handler was NOT called
	if mockHandler.called {
		t.Error("Expected next handler NOT to be called when access is denied")
	}

	// Verify CheckPermission was called
	if len(mockAM.CheckPermissionCalls) != 1 {
		t.Errorf("Expected CheckPermission to be called once, got %d calls", len(mockAM.CheckPermissionCalls))
	}

	// Verify rejection message was sent
	if len(mockClient.SendCalls) != 1 {
		t.Errorf("Expected 1 message to be sent for rejection, got %d", len(mockClient.SendCalls))
	}

	// Verify the rejection message content
	sentMessage, ok := mockClient.SendCalls[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Errorf("Expected MessageConfig to be sent, got %T", mockClient.SendCalls[0])
	} else {
		expectedText := "🚫 Access denied"
		if sentMessage.Text != expectedText {
			t.Errorf("Expected rejection text '%s', got '%s'", expectedText, sentMessage.Text)
		}
		if sentMessage.ChatID != 456 {
			t.Errorf("Expected message to be sent to chat 456, got %d", sentMessage.ChatID)
		}
	}
}

func TestAuthMiddleware_AccessDenied_CallbackQuery(t *testing.T) {
	mockAM := &mockAccessManager{
		CheckPermissionFunc: func(ctx *PermissionContext) error {
			return errors.New("Insufficient permissions")
		},
	}
	middleware := AuthMiddleware(mockAM)

	mockHandler := &mockHandler{}
	ctx, mockClient := createAuthMiddlewareTestContext("callback", 789, 101112)

	wrappedHandler := middleware(mockHandler.Handle)
	err := wrappedHandler(ctx)

	// Verify no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify next handler was NOT called
	if mockHandler.called {
		t.Error("Expected next handler NOT to be called when access is denied")
	}

	// Verify CheckPermission was called
	if len(mockAM.CheckPermissionCalls) != 1 {
		t.Errorf("Expected CheckPermission to be called once, got %d calls", len(mockAM.CheckPermissionCalls))
	}

	// Verify PermissionContext for callback query
	permCtx := mockAM.CheckPermissionCalls[0]
	if permCtx.UserID != 789 {
		t.Errorf("Expected UserID 789, got %d", permCtx.UserID)
	}
	if permCtx.ChatID != 101112 {
		t.Errorf("Expected ChatID 101112, got %d", permCtx.ChatID)
	}
	if permCtx.MessageID != 1 {
		t.Errorf("Expected MessageID 1, got %d", permCtx.MessageID)
	}

	// Verify rejection message was sent
	if len(mockClient.SendCalls) != 1 {
		t.Errorf("Expected 1 message to be sent for rejection, got %d", len(mockClient.SendCalls))
	}

	// Verify the rejection message content
	sentMessage, ok := mockClient.SendCalls[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Errorf("Expected MessageConfig to be sent, got %T", mockClient.SendCalls[0])
	} else {
		expectedText := "🚫 Insufficient permissions"
		if sentMessage.Text != expectedText {
			t.Errorf("Expected rejection text '%s', got '%s'", expectedText, sentMessage.Text)
		}
	}
}

func TestAuthMiddleware_AccessDenied_GroupCallback(t *testing.T) {
	mockAM := &mockAccessManager{
		CheckPermissionFunc: func(ctx *PermissionContext) error {
			return errors.New("Group access denied")
		},
	}
	middleware := AuthMiddleware(mockAM)

	mockHandler := &mockHandler{}
	ctx, mockClient := createAuthMiddlewareTestContext("callback_group", 999, 888)

	wrappedHandler := middleware(mockHandler.Handle)
	err := wrappedHandler(ctx)

	// Verify no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify next handler was NOT called
	if mockHandler.called {
		t.Error("Expected next handler NOT to be called when access is denied")
	}

	// Verify PermissionContext has correct group information
	if len(mockAM.CheckPermissionCalls) != 1 {
		t.Errorf("Expected CheckPermission to be called once, got %d calls", len(mockAM.CheckPermissionCalls))
	}

	permCtx := mockAM.CheckPermissionCalls[0]
	if !permCtx.IsGroup {
		t.Error("Expected IsGroup to be true for group callback")
	}
	if permCtx.MessageID != 2 {
		t.Errorf("Expected MessageID 2, got %d", permCtx.MessageID)
	}

	// Verify rejection message was sent
	if len(mockClient.SendCalls) != 1 {
		t.Errorf("Expected 1 message to be sent for rejection, got %d", len(mockClient.SendCalls))
	}
}

func TestAuthMiddleware_PermissionContextFormation(t *testing.T) {
	mockAM := &mockAccessManager{}
	middleware := AuthMiddleware(mockAM)

	testCases := []struct {
		name          string
		updateType    string
		userID        int64
		chatID        int64
		expectedGroup bool
		expectedMsgID int
	}{
		{
			name:          "private message",
			updateType:    "message",
			userID:        111,
			chatID:        222,
			expectedGroup: false,
			expectedMsgID: 1,
		},
		{
			name:          "group message",
			updateType:    "group_message",
			userID:        333,
			chatID:        444,
			expectedGroup: true,
			expectedMsgID: 1,
		},
		{
			name:          "private callback",
			updateType:    "callback",
			userID:        555,
			chatID:        666,
			expectedGroup: false,
			expectedMsgID: 1,
		},
		{
			name:          "group callback",
			updateType:    "callback_group",
			userID:        777,
			chatID:        888,
			expectedGroup: true,
			expectedMsgID: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock for each test
			mockAM.CheckPermissionCalls = nil

			mockHandler := &mockHandler{}
			ctx, _ := createAuthMiddlewareTestContext(tc.updateType, tc.userID, tc.chatID)

			wrappedHandler := middleware(mockHandler.Handle)
			_ = wrappedHandler(ctx)

			// Verify CheckPermission was called
			if len(mockAM.CheckPermissionCalls) != 1 {
				t.Errorf("Expected CheckPermission to be called once, got %d calls", len(mockAM.CheckPermissionCalls))
				return
			}

			permCtx := mockAM.CheckPermissionCalls[0]

			// Verify all PermissionContext fields
			if permCtx.UserID != tc.userID {
				t.Errorf("Expected UserID %d, got %d", tc.userID, permCtx.UserID)
			}
			if permCtx.ChatID != tc.chatID {
				t.Errorf("Expected ChatID %d, got %d", tc.chatID, permCtx.ChatID)
			}
			if permCtx.IsGroup != tc.expectedGroup {
				t.Errorf("Expected IsGroup %v, got %v", tc.expectedGroup, permCtx.IsGroup)
			}
			if permCtx.MessageID != tc.expectedMsgID {
				t.Errorf("Expected MessageID %d, got %d", tc.expectedMsgID, permCtx.MessageID)
			}
		})
	}
}

func TestAuthMiddleware_NilAccessManager(t *testing.T) {
	// Test behavior when nil AccessManager is passed
	middleware := AuthMiddleware(nil)

	mockHandler := &mockHandler{}
	ctx, _ := createAuthMiddlewareTestContext("message", 123, 456)

	wrappedHandler := middleware(mockHandler.Handle)

	// This should panic or cause an error since we're calling CheckPermission on nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when AccessManager is nil, but no panic occurred")
		}
	}()

	_ = wrappedHandler(ctx)
}
