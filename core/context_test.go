package teleflow

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Mock implementations for testing Context - using unique names to avoid conflicts

type contextMockTelegramClient struct {
	SendCalls    []tgbotapi.Chattable
	RequestCalls []tgbotapi.Chattable
	SendFunc     func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	RequestFunc  func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

func (m *contextMockTelegramClient) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SendCalls = append(m.SendCalls, c)
	if m.SendFunc != nil {
		return m.SendFunc(c)
	}
	return tgbotapi.Message{MessageID: 123}, nil
}

func (m *contextMockTelegramClient) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.RequestCalls = append(m.RequestCalls, c)
	if m.RequestFunc != nil {
		return m.RequestFunc(c)
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *contextMockTelegramClient) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return make(tgbotapi.UpdatesChannel)
}

func (m *contextMockTelegramClient) GetMe() (tgbotapi.User, error) {
	return tgbotapi.User{ID: 123, UserName: "testbot"}, nil
}

type contextMockTemplateManager struct {
	AddTemplateCalls []struct {
		Name, Text string
		ParseMode  ParseMode
	}
	HasTemplateCalls    []string
	GetTemplateCalls    []string
	ListTemplatesCalls  int
	RenderTemplateCalls []struct {
		Name string
		Data map[string]interface{}
	}

	AddTemplateFunc    func(name, templateText string, parseMode ParseMode) error
	HasTemplateFunc    func(name string) bool
	GetTemplateFunc    func(name string) *TemplateInfo
	ListTemplatesFunc  func() []string
	RenderTemplateFunc func(name string, data map[string]interface{}) (string, ParseMode, error)
}

func (m *contextMockTemplateManager) AddTemplate(name, templateText string, parseMode ParseMode) error {
	m.AddTemplateCalls = append(m.AddTemplateCalls, struct {
		Name, Text string
		ParseMode  ParseMode
	}{name, templateText, parseMode})
	if m.AddTemplateFunc != nil {
		return m.AddTemplateFunc(name, templateText, parseMode)
	}
	return nil
}

func (m *contextMockTemplateManager) HasTemplate(name string) bool {
	m.HasTemplateCalls = append(m.HasTemplateCalls, name)
	if m.HasTemplateFunc != nil {
		return m.HasTemplateFunc(name)
	}
	return false
}

func (m *contextMockTemplateManager) GetTemplateInfo(name string) *TemplateInfo {
	m.GetTemplateCalls = append(m.GetTemplateCalls, name)
	if m.GetTemplateFunc != nil {
		return m.GetTemplateFunc(name)
	}
	return nil
}

func (m *contextMockTemplateManager) ListTemplates() []string {
	m.ListTemplatesCalls++
	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc()
	}
	return []string{}
}

func (m *contextMockTemplateManager) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	m.RenderTemplateCalls = append(m.RenderTemplateCalls, struct {
		Name string
		Data map[string]interface{}
	}{name, data})
	if m.RenderTemplateFunc != nil {
		return m.RenderTemplateFunc(name, data)
	}
	return "rendered", ParseModeNone, nil
}

type contextMockFlowOperations struct {
	SetUserFlowDataCalls []struct {
		UserID int64
		Key    string
		Value  interface{}
	}
	GetUserFlowDataCalls []struct {
		UserID int64
		Key    string
	}
	StartFlowCalls []struct {
		UserID   int64
		FlowName string
		Ctx      *Context
	}
	IsUserInFlowCalls []int64
	CancelFlowCalls   []int64

	SetUserFlowDataFunc func(userID int64, key string, value interface{}) error
	GetUserFlowDataFunc func(userID int64, key string) (interface{}, bool)
	StartFlowFunc       func(userID int64, flowName string, ctx *Context) error
	IsUserInFlowFunc    func(userID int64) bool
	CancelFlowFunc      func(userID int64)
}

func (m *contextMockFlowOperations) setUserFlowData(userID int64, key string, value interface{}) error {
	m.SetUserFlowDataCalls = append(m.SetUserFlowDataCalls, struct {
		UserID int64
		Key    string
		Value  interface{}
	}{userID, key, value})
	if m.SetUserFlowDataFunc != nil {
		return m.SetUserFlowDataFunc(userID, key, value)
	}
	return nil
}

func (m *contextMockFlowOperations) getUserFlowData(userID int64, key string) (interface{}, bool) {
	m.GetUserFlowDataCalls = append(m.GetUserFlowDataCalls, struct {
		UserID int64
		Key    string
	}{userID, key})
	if m.GetUserFlowDataFunc != nil {
		return m.GetUserFlowDataFunc(userID, key)
	}
	return nil, false
}

func (m *contextMockFlowOperations) startFlow(userID int64, flowName string, ctx *Context) error {
	m.StartFlowCalls = append(m.StartFlowCalls, struct {
		UserID   int64
		FlowName string
		Ctx      *Context
	}{userID, flowName, ctx})
	if m.StartFlowFunc != nil {
		return m.StartFlowFunc(userID, flowName, ctx)
	}
	return nil
}

func (m *contextMockFlowOperations) isUserInFlow(userID int64) bool {
	m.IsUserInFlowCalls = append(m.IsUserInFlowCalls, userID)
	if m.IsUserInFlowFunc != nil {
		return m.IsUserInFlowFunc(userID)
	}
	return false
}

func (m *contextMockFlowOperations) cancelFlow(userID int64) {
	m.CancelFlowCalls = append(m.CancelFlowCalls, userID)
	if m.CancelFlowFunc != nil {
		m.CancelFlowFunc(userID)
	}
}

type contextMockPromptSender struct {
	ComposeAndSendCalls []struct {
		Ctx    *Context
		Config *PromptConfig
	}
	ComposeAndSendFunc func(ctx *Context, config *PromptConfig) error
}

func (m *contextMockPromptSender) ComposeAndSend(ctx *Context, config *PromptConfig) error {
	m.ComposeAndSendCalls = append(m.ComposeAndSendCalls, struct {
		Ctx    *Context
		Config *PromptConfig
	}{ctx, config})
	if m.ComposeAndSendFunc != nil {
		return m.ComposeAndSendFunc(ctx, config)
	}
	return nil
}

type contextMockAccessManager struct {
	CheckPermissionCalls  []*PermissionContext
	GetReplyKeyboardCalls []*PermissionContext
	CheckPermissionFunc   func(ctx *PermissionContext) error
	GetReplyKeyboardFunc  func(ctx *PermissionContext) *ReplyKeyboard
}

func (m *contextMockAccessManager) CheckPermission(ctx *PermissionContext) error {
	m.CheckPermissionCalls = append(m.CheckPermissionCalls, ctx)
	if m.CheckPermissionFunc != nil {
		return m.CheckPermissionFunc(ctx)
	}
	return nil
}

func (m *contextMockAccessManager) GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard {
	m.GetReplyKeyboardCalls = append(m.GetReplyKeyboardCalls, ctx)
	if m.GetReplyKeyboardFunc != nil {
		return m.GetReplyKeyboardFunc(ctx)
	}
	return nil
}

// Helper function to create test context with mocks
func createContextTestInstance(update tgbotapi.Update) (*Context, *contextMockTelegramClient, *contextMockTemplateManager, *contextMockFlowOperations, *contextMockPromptSender, *contextMockAccessManager) {
	mockClient := &contextMockTelegramClient{}
	mockTM := &contextMockTemplateManager{}
	mockFlowOps := &contextMockFlowOperations{}
	mockPS := &contextMockPromptSender{}
	mockAM := &contextMockAccessManager{}

	ctx := newContext(update, mockClient, mockTM, mockFlowOps, mockPS, mockAM)
	return ctx, mockClient, mockTM, mockFlowOps, mockPS, mockAM
}

// Test newContext initialization
func TestNewContext(t *testing.T) {
	tests := []struct {
		name     string
		update   tgbotapi.Update
		expected struct {
			userID    int64
			chatID    int64
			isGroup   bool
			isChannel bool
		}
	}{
		{
			name: "message update",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{ID: 12345},
					Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
				},
			},
			expected: struct {
				userID    int64
				chatID    int64
				isGroup   bool
				isChannel bool
			}{12345, 67890, false, false},
		},
		{
			name: "callback query update",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 54321},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 98765, Type: "private"},
					},
				},
			},
			expected: struct {
				userID    int64
				chatID    int64
				isGroup   bool
				isChannel bool
			}{54321, 98765, false, false},
		},
		{
			name: "group message update",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{ID: 11111},
					Chat: &tgbotapi.Chat{ID: 22222, Type: "group"},
				},
			},
			expected: struct {
				userID    int64
				chatID    int64
				isGroup   bool
				isChannel bool
			}{11111, 22222, true, false},
		},
		{
			name: "channel message update",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{ID: 33333},
					Chat: &tgbotapi.Chat{ID: 44444, Type: "channel"},
				},
			},
			expected: struct {
				userID    int64
				chatID    int64
				isGroup   bool
				isChannel bool
			}{33333, 44444, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, mockClient, mockTM, mockFlowOps, mockPS, mockAM := createContextTestInstance(tt.update)

			// Test that all dependencies are properly set
			if ctx.telegramClient != mockClient {
				t.Error("TelegramClient not properly initialized")
			}
			if ctx.templateManager != mockTM {
				t.Error("TemplateManager not properly initialized")
			}
			if ctx.flowOps != mockFlowOps {
				t.Error("ContextFlowOperations not properly initialized")
			}
			if ctx.promptSender != mockPS {
				t.Error("PromptSender not properly initialized")
			}
			if ctx.accessManager != mockAM {
				t.Error("AccessManager not properly initialized")
			}

			// Test extracted values
			if ctx.UserID() != tt.expected.userID {
				t.Errorf("Expected UserID %d, got %d", tt.expected.userID, ctx.UserID())
			}
			if ctx.ChatID() != tt.expected.chatID {
				t.Errorf("Expected ChatID %d, got %d", tt.expected.chatID, ctx.ChatID())
			}
			if ctx.IsGroup() != tt.expected.isGroup {
				t.Errorf("Expected IsGroup %v, got %v", tt.expected.isGroup, ctx.IsGroup())
			}
			if ctx.IsChannel() != tt.expected.isChannel {
				t.Errorf("Expected IsChannel %v, got %v", tt.expected.isChannel, ctx.IsChannel())
			}

			// Test that data map is initialized
			if ctx.data == nil {
				t.Error("Context data map not initialized")
			}

			// Test that update is stored (compare values, not addresses)
			if ctx.update.UpdateID != tt.update.UpdateID {
				t.Error("Update not properly stored in context")
			}
		})
	}
}

// Test getter methods for update details
func TestContext_GetterMethods(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}

	ctx, _, _, _, _, _ := createContextTestInstance(update)

	if ctx.UserID() != 12345 {
		t.Errorf("Expected UserID 12345, got %d", ctx.UserID())
	}

	if ctx.ChatID() != 67890 {
		t.Errorf("Expected ChatID 67890, got %d", ctx.ChatID())
	}
}

// Test context data management (Set, Get)
func TestContext_DataManagement(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}

	ctx, _, _, _, _, _ := createContextTestInstance(update)

	// Test Set and Get
	ctx.Set("testKey", "testValue")
	value, exists := ctx.Get("testKey")

	if !exists {
		t.Error("Expected key to exist")
	}

	if value != "testValue" {
		t.Errorf("Expected 'testValue', got %v", value)
	}

	// Test Get for non-existent key
	_, exists = ctx.Get("nonExistentKey")
	if exists {
		t.Error("Expected key to not exist")
	}

	// Test setting different data types
	ctx.Set("intKey", 42)
	ctx.Set("boolKey", true)
	ctx.Set("structKey", struct{ Name string }{"test"})

	intVal, _ := ctx.Get("intKey")
	if intVal != 42 {
		t.Errorf("Expected 42, got %v", intVal)
	}

	boolVal, _ := ctx.Get("boolKey")
	if boolVal != true {
		t.Errorf("Expected true, got %v", boolVal)
	}
}

// Test flow operation wrappers
func TestContext_FlowOperations(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}

	ctx, _, _, mockFlowOps, _, _ := createContextTestInstance(update)

	// Test StartFlow
	err := ctx.StartFlow("testFlow")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mockFlowOps.StartFlowCalls) != 1 {
		t.Errorf("Expected 1 StartFlow call, got %d", len(mockFlowOps.StartFlowCalls))
	}

	call := mockFlowOps.StartFlowCalls[0]
	if call.UserID != 12345 || call.FlowName != "testFlow" || call.Ctx != ctx {
		t.Error("StartFlow called with incorrect parameters")
	}

	// Test CancelFlow
	ctx.CancelFlow()
	if len(mockFlowOps.CancelFlowCalls) != 1 {
		t.Errorf("Expected 1 CancelFlow call, got %d", len(mockFlowOps.CancelFlowCalls))
	}

	if mockFlowOps.CancelFlowCalls[0] != 12345 {
		t.Error("CancelFlow called with incorrect userID")
	}

	// Test SetFlowData when user is not in flow
	err = ctx.SetFlowData("key", "value")
	if err == nil {
		t.Error("Expected error when user is not in flow")
	}

	// Test SetFlowData when user is in flow
	mockFlowOps.IsUserInFlowFunc = func(userID int64) bool { return true }
	err = ctx.SetFlowData("key", "value")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mockFlowOps.SetUserFlowDataCalls) != 1 {
		t.Errorf("Expected 1 SetUserFlowData call, got %d", len(mockFlowOps.SetUserFlowDataCalls))
	}

	// Test GetFlowData when user is not in flow
	mockFlowOps.IsUserInFlowFunc = func(userID int64) bool { return false }
	_, exists := ctx.GetFlowData("key")
	if exists {
		t.Error("Expected no data when user is not in flow")
	}

	// Test GetFlowData when user is in flow
	mockFlowOps.IsUserInFlowFunc = func(userID int64) bool { return true }
	mockFlowOps.GetUserFlowDataFunc = func(userID int64, key string) (interface{}, bool) {
		return "flowValue", true
	}

	value, exists := ctx.GetFlowData("key")
	if !exists {
		t.Error("Expected flow data to exist")
	}
	if value != "flowValue" {
		t.Errorf("Expected 'flowValue', got %v", value)
	}
}

// Test SendPrompt wrapper
func TestContext_SendPrompt(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}

	ctx, _, _, _, mockPS, _ := createContextTestInstance(update)

	promptConfig := &PromptConfig{
		Message:      "test message",
		TemplateData: map[string]interface{}{"key": "value"},
	}

	err := ctx.SendPrompt(promptConfig)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mockPS.ComposeAndSendCalls) != 1 {
		t.Errorf("Expected 1 ComposeAndSend call, got %d", len(mockPS.ComposeAndSendCalls))
	}

	call := mockPS.ComposeAndSendCalls[0]
	if call.Ctx != ctx {
		t.Error("SendPrompt called with incorrect context")
	}
	if call.Config.Message != "test message" {
		t.Error("SendPrompt called with incorrect config")
	}
}

// Test SendPromptText wrapper
func TestContext_SendPromptText(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}

	ctx, mockClient, _, _, _, _ := createContextTestInstance(update)

	err := ctx.SendPromptText("Hello World")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mockClient.SendCalls) != 1 {
		t.Errorf("Expected 1 Send call, got %d", len(mockClient.SendCalls))
	}

	// Verify the message sent
	msg, ok := mockClient.SendCalls[0].(tgbotapi.MessageConfig)
	if !ok {
		t.Error("Expected MessageConfig")
	}
	if msg.Text != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", msg.Text)
	}
	if msg.ChatID != 67890 {
		t.Errorf("Expected ChatID 67890, got %d", msg.ChatID)
	}
}

// Test SendPromptWithTemplate wrapper
func TestContext_SendPromptWithTemplate(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}

	ctx, _, _, _, mockPS, _ := createContextTestInstance(update)

	templateData := map[string]interface{}{"name": "John"}
	err := ctx.SendPromptWithTemplate("greeting", templateData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mockPS.ComposeAndSendCalls) != 1 {
		t.Errorf("Expected 1 ComposeAndSend call, got %d", len(mockPS.ComposeAndSendCalls))
	}

	call := mockPS.ComposeAndSendCalls[0]
	if call.Config.Message != "template:greeting" {
		t.Error("Template message not formatted correctly")
	}
	if call.Config.TemplateData["name"] != "John" {
		t.Error("Template data not passed correctly")
	}
}

// Test template management wrappers
func TestContext_TemplateManagement(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}

	ctx, _, mockTM, _, _, _ := createContextTestInstance(update)

	// Test AddTemplate
	err := ctx.AddTemplate("test", "Hello {{.name}}", ParseModeHTML)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mockTM.AddTemplateCalls) != 1 {
		t.Errorf("Expected 1 AddTemplate call, got %d", len(mockTM.AddTemplateCalls))
	}

	// Test HasTemplate
	mockTM.HasTemplateFunc = func(name string) bool { return name == "test" }
	if !ctx.HasTemplate("test") {
		t.Error("Expected template to exist")
	}

	// Test GetTemplateInfo
	mockTM.GetTemplateFunc = func(name string) *TemplateInfo {
		return &TemplateInfo{Name: name, ParseMode: ParseModeHTML}
	}
	info := ctx.GetTemplateInfo("test")
	if info == nil || info.Name != "test" {
		t.Error("Template info not returned correctly")
	}

	// Test ListTemplates
	mockTM.ListTemplatesFunc = func() []string { return []string{"test1", "test2"} }
	templates := ctx.ListTemplates()
	if len(templates) != 2 {
		t.Error("Templates list not returned correctly")
	}

	// Test RenderTemplate
	mockTM.RenderTemplateFunc = func(name string, data map[string]interface{}) (string, ParseMode, error) {
		return "Hello John", ParseModeHTML, nil
	}
	result, parseMode, err := ctx.RenderTemplate("test", map[string]interface{}{"name": "John"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "Hello John" {
		t.Errorf("Expected 'Hello John', got %s", result)
	}
	if parseMode != ParseModeHTML {
		t.Error("Parse mode not returned correctly")
	}

	// Test TemplateManager getter
	if ctx.TemplateManager() != mockTM {
		t.Error("TemplateManager getter not working correctly")
	}
}

// Test answerCallbackQuery method
func TestContext_AnswerCallbackQuery(t *testing.T) {
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback123",
			From: &tgbotapi.User{ID: 12345},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
			},
		},
	}

	ctx, mockClient, _, _, _, _ := createContextTestInstance(update)

	err := ctx.answerCallbackQuery("Response text")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(mockClient.RequestCalls) != 1 {
		t.Errorf("Expected 1 Request call, got %d", len(mockClient.RequestCalls))
	}

	// Verify the callback response
	callback, ok := mockClient.RequestCalls[0].(tgbotapi.CallbackConfig)
	if !ok {
		t.Error("Expected CallbackConfig")
	}
	if callback.CallbackQueryID != "callback123" {
		t.Errorf("Expected callback ID 'callback123', got %s", callback.CallbackQueryID)
	}
	if callback.Text != "Response text" {
		t.Errorf("Expected 'Response text', got %s", callback.Text)
	}

	// Test with no callback query
	updateNoCallback := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "private"},
		},
	}
	ctxNoCallback, mockClientNoCallback, _, _, _, _ := createContextTestInstance(updateNoCallback)

	err = ctxNoCallback.answerCallbackQuery("Response text")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should not make any requests
	if len(mockClientNoCallback.RequestCalls) != 0 {
		t.Error("Expected no Request calls when no callback query")
	}
}

// Test getPermissionContext method
func TestContext_GetPermissionContext(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 12345},
			Chat: &tgbotapi.Chat{ID: 67890, Type: "group"},
		},
	}

	ctx, _, _, _, _, _ := createContextTestInstance(update)

	permCtx := ctx.getPermissionContext()
	if permCtx == nil {
		t.Fatal("Expected permission context to be created")
	}

	if permCtx.UserID != 12345 {
		t.Errorf("Expected UserID 12345, got %d", permCtx.UserID)
	}
	if permCtx.ChatID != 67890 {
		t.Errorf("Expected ChatID 67890, got %d", permCtx.ChatID)
	}
	if !permCtx.IsGroup {
		t.Error("Expected IsGroup to be true")
	}
	if permCtx.IsChannel {
		t.Error("Expected IsChannel to be false")
	}

	// Test with nil access manager
	ctxNilAM := newContext(update, &contextMockTelegramClient{}, &contextMockTemplateManager{}, &contextMockFlowOperations{}, &contextMockPromptSender{}, nil)
	permCtxNil := ctxNilAM.getPermissionContext()
	if permCtxNil != nil {
		t.Error("Expected nil permission context when access manager is nil")
	}
}

// Test extract methods
func TestContext_ExtractMethods(t *testing.T) {
	tests := []struct {
		name           string
		update         tgbotapi.Update
		expectedUserID int64
		expectedChatID int64
	}{
		{
			name: "message update",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{ID: 111},
					Chat: &tgbotapi.Chat{ID: 222},
				},
			},
			expectedUserID: 111,
			expectedChatID: 222,
		},
		{
			name: "callback query update",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 333},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 444},
					},
				},
			},
			expectedUserID: 333,
			expectedChatID: 444,
		},
		{
			name:           "empty update",
			update:         tgbotapi.Update{},
			expectedUserID: 0,
			expectedChatID: 0,
		},
		{
			name: "callback query without message",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 555},
				},
			},
			expectedUserID: 555,
			expectedChatID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _, _, _, _, _ := createContextTestInstance(tt.update)

			if ctx.UserID() != tt.expectedUserID {
				t.Errorf("Expected UserID %d, got %d", tt.expectedUserID, ctx.UserID())
			}
			if ctx.ChatID() != tt.expectedChatID {
				t.Errorf("Expected ChatID %d, got %d", tt.expectedChatID, ctx.ChatID())
			}
		})
	}
}
