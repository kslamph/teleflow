package teleflow

import (
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Mock implementations for testing

type MockTelegramClient struct {
	RequestFunc         func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	RequestCalls        []tgbotapi.Chattable
	SendFunc            func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	SendCalls           []tgbotapi.Chattable
	GetUpdatesChanFunc  func(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	GetUpdatesChanCalls []tgbotapi.UpdateConfig
	GetMeFunc           func() (tgbotapi.User, error)
	GetMeCalls          int
}

func NewMockTelegramClient() *MockTelegramClient {
	return &MockTelegramClient{
		RequestCalls:        make([]tgbotapi.Chattable, 0),
		SendCalls:           make([]tgbotapi.Chattable, 0),
		GetUpdatesChanCalls: make([]tgbotapi.UpdateConfig, 0),
	}
}

func (m *MockTelegramClient) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.RequestCalls = append(m.RequestCalls, c)
	if m.RequestFunc != nil {
		return m.RequestFunc(c)
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *MockTelegramClient) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SendCalls = append(m.SendCalls, c)
	if m.SendFunc != nil {
		return m.SendFunc(c)
	}
	return tgbotapi.Message{MessageID: 123}, nil
}

func (m *MockTelegramClient) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	m.GetUpdatesChanCalls = append(m.GetUpdatesChanCalls, config)
	if m.GetUpdatesChanFunc != nil {
		return m.GetUpdatesChanFunc(config)
	}
	return make(chan tgbotapi.Update)
}

func (m *MockTelegramClient) GetMe() (tgbotapi.User, error) {
	m.GetMeCalls++
	if m.GetMeFunc != nil {
		return m.GetMeFunc()
	}
	return tgbotapi.User{ID: 12345, UserName: "TestBot"}, nil
}

type MockTemplateManager struct {
	AddTemplateFunc    func(name, templateText string, parseMode ParseMode) error
	HasTemplateFunc    func(name string) bool
	GetTemplateFunc    func(name string) *TemplateInfo
	ListTemplatesFunc  func() []string
	RenderTemplateFunc func(name string, data map[string]interface{}) (string, ParseMode, error)

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
}

func NewMockTemplateManager() *MockTemplateManager {
	return &MockTemplateManager{
		AddTemplateCalls: make([]struct {
			Name, Text string
			ParseMode  ParseMode
		}, 0),
		HasTemplateCalls: make([]string, 0),
		GetTemplateCalls: make([]string, 0),
		RenderTemplateCalls: make([]struct {
			Name string
			Data map[string]interface{}
		}, 0),
	}
}

func (m *MockTemplateManager) AddTemplate(name, templateText string, parseMode ParseMode) error {
	m.AddTemplateCalls = append(m.AddTemplateCalls, struct {
		Name, Text string
		ParseMode  ParseMode
	}{name, templateText, parseMode})
	if m.AddTemplateFunc != nil {
		return m.AddTemplateFunc(name, templateText, parseMode)
	}
	return nil
}

func (m *MockTemplateManager) HasTemplate(name string) bool {
	m.HasTemplateCalls = append(m.HasTemplateCalls, name)
	if m.HasTemplateFunc != nil {
		return m.HasTemplateFunc(name)
	}
	return false
}

func (m *MockTemplateManager) GetTemplateInfo(name string) *TemplateInfo {
	m.GetTemplateCalls = append(m.GetTemplateCalls, name)
	if m.GetTemplateFunc != nil {
		return m.GetTemplateFunc(name)
	}
	return nil
}

func (m *MockTemplateManager) ListTemplates() []string {
	m.ListTemplatesCalls++
	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc()
	}
	return []string{}
}

func (m *MockTemplateManager) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	m.RenderTemplateCalls = append(m.RenderTemplateCalls, struct {
		Name string
		Data map[string]interface{}
	}{name, data})
	if m.RenderTemplateFunc != nil {
		return m.RenderTemplateFunc(name, data)
	}
	return "rendered", ParseModeNone, nil
}

type MockAccessManager struct {
	CheckPermissionFunc  func(ctx *PermissionContext) error
	GetReplyKeyboardFunc func(ctx *PermissionContext) *ReplyKeyboard

	CheckPermissionCalls  []*PermissionContext
	GetReplyKeyboardCalls []*PermissionContext
}

func NewMockAccessManager() *MockAccessManager {
	return &MockAccessManager{
		CheckPermissionCalls:  make([]*PermissionContext, 0),
		GetReplyKeyboardCalls: make([]*PermissionContext, 0),
	}
}

func (m *MockAccessManager) CheckPermission(ctx *PermissionContext) error {
	m.CheckPermissionCalls = append(m.CheckPermissionCalls, ctx)
	if m.CheckPermissionFunc != nil {
		return m.CheckPermissionFunc(ctx)
	}
	return nil
}

func (m *MockAccessManager) GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard {
	m.GetReplyKeyboardCalls = append(m.GetReplyKeyboardCalls, ctx)
	if m.GetReplyKeyboardFunc != nil {
		return m.GetReplyKeyboardFunc(ctx)
	}
	return nil
}

type MockPromptComposer struct {
	ComposeAndSendFunc  func(ctx *Context, config *PromptConfig) error
	ComposeAndSendCalls []struct {
		Ctx    *Context
		Config *PromptConfig
	}
}

func NewMockPromptComposer() *MockPromptComposer {
	return &MockPromptComposer{
		ComposeAndSendCalls: make([]struct {
			Ctx    *Context
			Config *PromptConfig
		}, 0),
	}
}

func (m *MockPromptComposer) ComposeAndSend(ctx *Context, config *PromptConfig) error {
	m.ComposeAndSendCalls = append(m.ComposeAndSendCalls, struct {
		Ctx    *Context
		Config *PromptConfig
	}{ctx, config})
	if m.ComposeAndSendFunc != nil {
		return m.ComposeAndSendFunc(ctx, config)
	}
	return nil
}

type MockPromptKeyboardActions struct {
	BuildKeyboardFunc       func(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error)
	GetCallbackDataFunc     func(userID int64, uuid string) (interface{}, bool)
	CleanupUserMappingsFunc func(userID int64)

	BuildKeyboardCalls []struct {
		Ctx          *Context
		KeyboardFunc KeyboardFunc
	}
	GetCallbackDataCalls []struct {
		UserID int64
		UUID   string
	}
	CleanupUserMappingsCalls []int64
}

func NewMockPromptKeyboardActions() *MockPromptKeyboardActions {
	return &MockPromptKeyboardActions{
		BuildKeyboardCalls: make([]struct {
			Ctx          *Context
			KeyboardFunc KeyboardFunc
		}, 0),
		GetCallbackDataCalls: make([]struct {
			UserID int64
			UUID   string
		}, 0),
		CleanupUserMappingsCalls: make([]int64, 0),
	}
}

func (m *MockPromptKeyboardActions) BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
	m.BuildKeyboardCalls = append(m.BuildKeyboardCalls, struct {
		Ctx          *Context
		KeyboardFunc KeyboardFunc
	}{ctx, keyboardFunc})
	if m.BuildKeyboardFunc != nil {
		return m.BuildKeyboardFunc(ctx, keyboardFunc)
	}
	return tgbotapi.InlineKeyboardMarkup{}, nil
}

func (m *MockPromptKeyboardActions) GetCallbackData(userID int64, uuid string) (interface{}, bool) {
	m.GetCallbackDataCalls = append(m.GetCallbackDataCalls, struct {
		UserID int64
		UUID   string
	}{userID, uuid})
	if m.GetCallbackDataFunc != nil {
		return m.GetCallbackDataFunc(userID, uuid)
	}
	return nil, false
}

func (m *MockPromptKeyboardActions) CleanupUserMappings(userID int64) {
	m.CleanupUserMappingsCalls = append(m.CleanupUserMappingsCalls, userID)
	if m.CleanupUserMappingsFunc != nil {
		m.CleanupUserMappingsFunc(userID)
	}
}

type MockFlowManager struct {
	RegisterFlowFunc    func(flow *Flow)
	IsUserInFlowFunc    func(userID int64) bool
	CancelFlowFunc      func(userID int64)
	HandleUpdateFunc    func(ctx *Context) (bool, error)
	StartFlowFunc       func(userID int64, flowName string, ctx *Context) error
	SetUserFlowDataFunc func(userID int64, key string, value interface{}) error
	GetUserFlowDataFunc func(userID int64, key string) (interface{}, bool)

	RegisterFlowCalls []*Flow
	IsUserInFlowCalls []int64
	CancelFlowCalls   []int64
	HandleUpdateCalls []*Context
	StartFlowCalls    []struct {
		UserID   int64
		FlowName string
		Ctx      *Context
	}
	SetUserFlowDataCalls []struct {
		UserID int64
		Key    string
		Value  interface{}
	}
	GetUserFlowDataCalls []struct {
		UserID int64
		Key    string
	}
}

func NewMockFlowManager() *MockFlowManager {
	return &MockFlowManager{
		RegisterFlowCalls: make([]*Flow, 0),
		IsUserInFlowCalls: make([]int64, 0),
		CancelFlowCalls:   make([]int64, 0),
		HandleUpdateCalls: make([]*Context, 0),
		StartFlowCalls: make([]struct {
			UserID   int64
			FlowName string
			Ctx      *Context
		}, 0),
		SetUserFlowDataCalls: make([]struct {
			UserID int64
			Key    string
			Value  interface{}
		}, 0),
		GetUserFlowDataCalls: make([]struct {
			UserID int64
			Key    string
		}, 0),
	}
}

func (m *MockFlowManager) isUserInFlow(userID int64) bool {
	m.IsUserInFlowCalls = append(m.IsUserInFlowCalls, userID)
	if m.IsUserInFlowFunc != nil {
		return m.IsUserInFlowFunc(userID)
	}
	return false
}

func (m *MockFlowManager) cancelFlow(userID int64) {
	m.CancelFlowCalls = append(m.CancelFlowCalls, userID)
	if m.CancelFlowFunc != nil {
		m.CancelFlowFunc(userID)
	}
}

func (m *MockFlowManager) HandleUpdate(ctx *Context) (bool, error) {
	m.HandleUpdateCalls = append(m.HandleUpdateCalls, ctx)
	if m.HandleUpdateFunc != nil {
		return m.HandleUpdateFunc(ctx)
	}
	return false, nil
}

func (m *MockFlowManager) startFlow(userID int64, flowName string, ctx *Context) error {
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

func (m *MockFlowManager) setUserFlowData(userID int64, key string, value interface{}) error {
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

func (m *MockFlowManager) getUserFlowData(userID int64, key string) (interface{}, bool) {
	m.GetUserFlowDataCalls = append(m.GetUserFlowDataCalls, struct {
		UserID int64
		Key    string
	}{userID, key})
	if m.GetUserFlowDataFunc != nil {
		return m.GetUserFlowDataFunc(userID, key)
	}
	return nil, false
}

// Helper function to create a bot with mocked dependencies
func createTestBot(options ...BotOption) (*Bot, *MockTelegramClient, *MockTemplateManager, *MockAccessManager) {
	mockClient := NewMockTelegramClient()
	mockTemplateManager := NewMockTemplateManager()
	mockAccessManager := NewMockAccessManager()
	mockUser := tgbotapi.User{ID: 12345, UserName: "TestBot"}

	// Create bot options that inject mocked dependencies
	allOptions := append(options,
		func(b *Bot) {
			b.templateManager = mockTemplateManager
			b.accessManager = mockAccessManager
		},
	)

	bot, _ := newBotInternal(mockClient, mockUser, allOptions...)

	return bot, mockClient, mockTemplateManager, mockAccessManager
}

// Test NewBot constructor with options
func TestNewBot_DefaultInitialization(t *testing.T) {
	mockClient := NewMockTelegramClient()
	mockUser := tgbotapi.User{ID: 12345, UserName: "TestBot"}

	bot, err := newBotInternal(mockClient, mockUser)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if bot.api != mockClient {
		t.Error("Expected telegram client to be set")
	}

	if bot.self != mockUser {
		t.Error("Expected bot user to be set")
	}

	if bot.handlers == nil {
		t.Error("Expected handlers map to be initialized")
	}

	if bot.textHandlers == nil {
		t.Error("Expected text handlers map to be initialized")
	}

	if bot.middleware == nil {
		t.Error("Expected middleware slice to be initialized")
	}

	if len(bot.flowConfig.ExitCommands) == 0 {
		t.Error("Expected default exit commands to be set")
	}

	if bot.flowConfig.ExitMessage == "" {
		t.Error("Expected default exit message to be set")
	}
}

func TestNewBot_WithFlowConfig(t *testing.T) {
	customConfig := FlowConfig{
		ExitCommands:        []string{"/quit", "/exit"},
		ExitMessage:         "Custom exit message",
		AllowGlobalCommands: true,
		HelpCommands:        []string{"/help", "/info"},
		OnProcessAction:     ProcessDeleteMessage,
	}

	mockClient := NewMockTelegramClient()
	mockUser := tgbotapi.User{ID: 12345, UserName: "TestBot"}

	bot, err := newBotInternal(mockClient, mockUser, WithFlowConfig(customConfig))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(bot.flowConfig.ExitCommands) != 2 {
		t.Errorf("Expected 2 exit commands, got %d", len(bot.flowConfig.ExitCommands))
	}

	if bot.flowConfig.ExitMessage != "Custom exit message" {
		t.Errorf("Expected custom exit message, got: %s", bot.flowConfig.ExitMessage)
	}

	if !bot.flowConfig.AllowGlobalCommands {
		t.Error("Expected AllowGlobalCommands to be true")
	}
}

func TestNewBot_WithAccessManager(t *testing.T) {
	mockAccessManager := NewMockAccessManager()
	mockClient := NewMockTelegramClient()
	mockUser := tgbotapi.User{ID: 12345, UserName: "TestBot"}

	bot, err := newBotInternal(mockClient, mockUser, WithAccessManager(mockAccessManager))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if bot.accessManager != mockAccessManager {
		t.Error("Expected access manager to be set")
	}

	// WithAccessManager should also add auth middleware
	if len(bot.middleware) == 0 {
		t.Error("Expected auth middleware to be added")
	}
}

// Test middleware functionality
func TestBot_UseMiddleware(t *testing.T) {
	bot, _, _, _ := createTestBot()

	var executionOrder []string

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware1")
			return next(ctx)
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware2")
			return next(ctx)
		}
	}

	bot.UseMiddleware(middleware1)
	bot.UseMiddleware(middleware2)

	if len(bot.middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(bot.middleware))
	}
}

func TestBot_ApplyMiddleware(t *testing.T) {
	bot, _, _, _ := createTestBot()

	var executionOrder []string

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware1")
			return next(ctx)
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware2")
			return next(ctx)
		}
	}

	baseHandler := func(ctx *Context) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}

	bot.UseMiddleware(middleware1)
	bot.UseMiddleware(middleware2)

	wrappedHandler := bot.applyMiddleware(baseHandler)

	// Create a mock context for testing
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, NewMockTelegramClient(), NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	err := wrappedHandler(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedOrder := []string{"middleware1", "middleware2", "handler"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d executions, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Expected execution order %v, got %v", expectedOrder, executionOrder)
			break
		}
	}
}

// Test command handling
func TestBot_HandleCommand(t *testing.T) {
	bot, _, _, _ := createTestBot()

	var handlerCalled bool
	var receivedCommand, receivedArgs string

	handler := func(ctx *Context, command string, args string) error {
		handlerCalled = true
		receivedCommand = command
		receivedArgs = args
		return nil
	}

	bot.HandleCommand("start", handler)

	if _, exists := bot.handlers["start"]; !exists {
		t.Error("Expected command handler to be registered")
	}

	// Test command execution
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "/start hello world",
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, NewMockTelegramClient(), NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	err := bot.handlers["start"](ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if receivedCommand != "start" {
		t.Errorf("Expected command 'start', got '%s'", receivedCommand)
	}

	if receivedArgs != " hello world" {
		t.Errorf("Expected args ' hello world', got '%s'", receivedArgs)
	}
}

// Test flow registration
func TestBot_RegisterFlow(t *testing.T) {
	bot, _, _, _ := createTestBot()

	flow := &Flow{
		Name: "test-flow",
	}

	// We can't easily mock the flowManager here since it's a concrete type
	// But we can verify the method doesn't panic and completes
	bot.RegisterFlow(flow)

	// This test verifies the method works without error
	// The actual flow registration logic is tested separately in flow_test.go
}

// Test processUpdate logic
func TestBot_ProcessUpdate_CommandMessage(t *testing.T) {
	bot, _, _, _ := createTestBot()

	var handlerCalled bool
	handler := func(ctx *Context, command string, args string) error {
		handlerCalled = true
		return nil
	}

	bot.HandleCommand("test", handler)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text:     "/test",
			From:     &tgbotapi.User{ID: 123},
			Chat:     &tgbotapi.Chat{ID: 456},
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}},
		},
	}

	bot.processUpdate(update)

	if !handlerCalled {
		t.Error("Expected command handler to be called")
	}
}

func TestBot_ProcessUpdate_UserInFlow(t *testing.T) {
	// This test is more complex to implement without direct flow manager mocking
	// The processUpdate method creates its own context and calls flowManager methods
	// For now, we verify the method runs without panicking
	bot, _, _, _ := createTestBot()

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "some text",
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}

	// This should not panic
	bot.processUpdate(update)
}

func TestBot_ProcessUpdate_CallbackQuery(t *testing.T) {
	bot, mockClient, _, _ := createTestBot()

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback123",
			From: &tgbotapi.User{ID: 123},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 456},
			},
		},
	}

	bot.processUpdate(update)

	// Check that callback query was answered
	if len(mockClient.RequestCalls) == 0 {
		t.Error("Expected callback query to be answered")
	}
}

func TestBot_ProcessUpdate_ExitCommand(t *testing.T) {
	// This test verifies that exit commands are processed
	// We can't easily mock flow state, but we can verify the method runs
	bot, _, _, _ := createTestBot()

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "/cancel",
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}

	// This should not panic and should handle the exit command logic
	bot.processUpdate(update)
}

// Test MessageCleaner implementation
func TestBot_DeleteMessage(t *testing.T) {
	bot, mockClient, _, _ := createTestBot()

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, mockClient, NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	err := bot.DeleteMessage(ctx, 789)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(mockClient.RequestCalls) != 1 {
		t.Errorf("Expected 1 Request call, got %d", len(mockClient.RequestCalls))
	}

	if deleteMsg, ok := mockClient.RequestCalls[0].(tgbotapi.DeleteMessageConfig); ok {
		if deleteMsg.ChatID != 456 {
			t.Errorf("Expected ChatID 456, got %d", deleteMsg.ChatID)
		}
		if deleteMsg.MessageID != 789 {
			t.Errorf("Expected MessageID 789, got %d", deleteMsg.MessageID)
		}
	} else {
		t.Errorf("Expected DeleteMessageConfig, got %T", mockClient.RequestCalls[0])
	}
}

func TestBot_EditMessageReplyMarkup(t *testing.T) {
	bot, mockClient, _, _ := createTestBot()

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, mockClient, NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "Button", CallbackData: &[]string{"data"}[0]}},
		},
	}

	err := bot.EditMessageReplyMarkup(ctx, 789, keyboard)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(mockClient.RequestCalls) != 1 {
		t.Errorf("Expected 1 Request call, got %d", len(mockClient.RequestCalls))
	}

	if editMsg, ok := mockClient.RequestCalls[0].(tgbotapi.EditMessageReplyMarkupConfig); ok {
		if editMsg.ChatID != 456 {
			t.Errorf("Expected ChatID 456, got %d", editMsg.ChatID)
		}
		if editMsg.MessageID != 789 {
			t.Errorf("Expected MessageID 789, got %d", editMsg.MessageID)
		}
	} else {
		t.Errorf("Expected EditMessageReplyMarkupConfig, got %T", mockClient.SendCalls[0])
	}
}

func TestBot_EditMessageReplyMarkup_RemoveKeyboard(t *testing.T) {
	bot, mockClient, _, _ := createTestBot()

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, mockClient, NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	err := bot.EditMessageReplyMarkup(ctx, 789, nil)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(mockClient.RequestCalls) != 1 {
		t.Errorf("Expected 1 Request call, got %d", len(mockClient.RequestCalls))
	}

	if editMsg, ok := mockClient.RequestCalls[0].(tgbotapi.EditMessageReplyMarkupConfig); ok {
		if len(editMsg.ReplyMarkup.InlineKeyboard) != 0 {
			t.Error("Expected empty keyboard when removing")
		}
	} else {
		t.Errorf("Expected EditMessageReplyMarkupConfig, got %T", mockClient.RequestCalls[0])
	}
}

func TestBot_EditMessageReplyMarkup_InvalidType(t *testing.T) {
	bot, _, _, _ := createTestBot()

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, NewMockTelegramClient(), NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	err := bot.EditMessageReplyMarkup(ctx, 789, "invalid")

	if err == nil {
		t.Error("Expected error for invalid reply markup type")
	}

	if err.Error() != "replyMarkup must be of type tgbotapi.InlineKeyboardMarkup" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// Test error handling in MessageCleaner methods
func TestBot_DeleteMessage_Error(t *testing.T) {
	bot, mockClient, _, _ := createTestBot()

	mockClient.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
		return nil, errors.New("API error")
	}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, mockClient, NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	err := bot.DeleteMessage(ctx, 789)

	if err == nil {
		t.Error("Expected error from API")
	}

	if err.Error() != "API error" {
		t.Errorf("Expected 'API error', got: %v", err)
	}
}

func TestBot_EditMessageReplyMarkup_Error(t *testing.T) {
	bot, mockClient, _, _ := createTestBot()

	mockClient.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
		return nil, errors.New("API error")
	}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123},
			Chat: &tgbotapi.Chat{ID: 456},
		},
	}
	ctx := newContext(update, mockClient, NewMockTemplateManager(), NewMockFlowManager(), NewMockPromptComposer(), NewMockAccessManager())

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	err := bot.EditMessageReplyMarkup(ctx, 789, keyboard)

	if err == nil {
		t.Error("Expected error from API")
	}

	if err.Error() != "API error" {
		t.Errorf("Expected 'API error', got: %v", err)
	}
}

// Integration test with multiple components
func TestBot_Integration_CommandWithMiddleware(t *testing.T) {
	bot, _, _, _ := createTestBot()

	var middlewareExecuted, handlerExecuted bool

	// Add middleware
	bot.UseMiddleware(func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			middlewareExecuted = true
			return next(ctx)
		}
	})

	// Add command handler
	bot.HandleCommand("test", func(ctx *Context, command string, args string) error {
		handlerExecuted = true
		return nil
	})

	// Process update
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text:     "/test",
			From:     &tgbotapi.User{ID: 123},
			Chat:     &tgbotapi.Chat{ID: 456},
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}},
		},
	}

	bot.processUpdate(update)

	if !middlewareExecuted {
		t.Error("Expected middleware to be executed")
	}

	if !handlerExecuted {
		t.Error("Expected handler to be executed")
	}
}

// Test table-driven scenarios for processUpdate
func TestBot_ProcessUpdate_Scenarios(t *testing.T) {
	tests := []struct {
		name           string
		update         tgbotapi.Update
		userInFlow     bool
		flowHandles    bool
		expectFlowCall bool
		expectCmdCall  bool
	}{
		{
			name: "Command when not in flow",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text:     "/start",
					From:     &tgbotapi.User{ID: 123},
					Chat:     &tgbotapi.Chat{ID: 456},
					Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}},
				},
			},
			userInFlow:     false,
			flowHandles:    false,
			expectFlowCall: true,
			expectCmdCall:  true,
		},
		{
			name: "Text when in flow",
			update: tgbotapi.Update{
				Message: &tgbotapi.Message{
					Text: "some text",
					From: &tgbotapi.User{ID: 123},
					Chat: &tgbotapi.Chat{ID: 456},
				},
			},
			userInFlow:     true,
			flowHandles:    true,
			expectFlowCall: true,
			expectCmdCall:  false,
		},
		{
			name: "Callback query",
			update: tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					ID:   "callback123",
					From: &tgbotapi.User{ID: 123},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 456},
					},
				},
			},
			userInFlow:     false,
			flowHandles:    true,
			expectFlowCall: true,
			expectCmdCall:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, _, _, _ := createTestBot()

			var cmdHandlerCalled bool
			bot.HandleCommand("start", func(ctx *Context, command string, args string) error {
				cmdHandlerCalled = true
				return nil
			})

			bot.processUpdate(tt.update)

			// For command messages, we can verify the handler was called
			if tt.expectCmdCall && !cmdHandlerCalled {
				t.Error("Expected command handler to be called")
			}

			if !tt.expectCmdCall && cmdHandlerCalled {
				t.Error("Expected command handler not to be called")
			}
		})
	}
}
