package teleflow

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Mock implementations for testing

type mockPromptSender struct {
	composeAndSendCalls []PromptConfig
	composeAndSendError error
	errorOnce           bool // Only return error on first call
	mu                  sync.Mutex
}

func (m *mockPromptSender) ComposeAndSend(ctx *Context, config *PromptConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.composeAndSendCalls = append(m.composeAndSendCalls, *config)

	if m.composeAndSendError != nil {
		if m.errorOnce {
			// Return error only once, then clear it
			err := m.composeAndSendError
			m.composeAndSendError = nil
			return err
		}
		return m.composeAndSendError
	}
	return nil
}

func (m *mockPromptSender) getComposeAndSendCalls() []PromptConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]PromptConfig, len(m.composeAndSendCalls))
	copy(calls, m.composeAndSendCalls)
	return calls
}

func (m *mockPromptSender) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.composeAndSendCalls = nil
	m.composeAndSendError = nil
	m.errorOnce = false
}

func (m *mockPromptSender) setErrorOnce(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.composeAndSendError = err
	m.errorOnce = true
}

type mockPromptKeyboardActions struct {
	buildKeyboardCalls   []KeyboardFunc
	buildKeyboardReturns []interface{}
	buildKeyboardError   error
	callbackData         map[int64]map[string]interface{}
	cleanupUserCalls     []int64
	mu                   sync.Mutex
}

func (m *mockPromptKeyboardActions) BuildKeyboard(ctx *Context, keyboardFunc KeyboardFunc) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buildKeyboardCalls = append(m.buildKeyboardCalls, keyboardFunc)
	if len(m.buildKeyboardReturns) > 0 {
		result := m.buildKeyboardReturns[0]
		m.buildKeyboardReturns = m.buildKeyboardReturns[1:]
		return result, m.buildKeyboardError
	}
	return nil, m.buildKeyboardError
}

func (m *mockPromptKeyboardActions) GetCallbackData(userID int64, uuid string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if userMappings, exists := m.callbackData[userID]; exists {
		if data, found := userMappings[uuid]; found {
			return data, true
		}
	}
	return nil, false
}

func (m *mockPromptKeyboardActions) CleanupUserMappings(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupUserCalls = append(m.cleanupUserCalls, userID)
	delete(m.callbackData, userID)
}

func (m *mockPromptKeyboardActions) setCallbackData(userID int64, uuid string, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.callbackData == nil {
		m.callbackData = make(map[int64]map[string]interface{})
	}
	if m.callbackData[userID] == nil {
		m.callbackData[userID] = make(map[string]interface{})
	}
	m.callbackData[userID][uuid] = data
}

// reset method removed as it's unused

type mockMessageCleaner struct {
	deleteMessageCalls          []messageCall
	editMessageReplyMarkupCalls []editMarkupCall
	deleteMessageError          error
	editMessageReplyMarkupError error
	mu                          sync.Mutex
}

type messageCall struct {
	userID    int64
	messageID int
}

type editMarkupCall struct {
	userID      int64
	messageID   int
	replyMarkup interface{}
}

func (m *mockMessageCleaner) DeleteMessage(ctx *Context, messageID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteMessageCalls = append(m.deleteMessageCalls, messageCall{
		userID:    ctx.UserID(),
		messageID: messageID,
	})
	return m.deleteMessageError
}

func (m *mockMessageCleaner) EditMessageReplyMarkup(ctx *Context, messageID int, replyMarkup interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.editMessageReplyMarkupCalls = append(m.editMessageReplyMarkupCalls, editMarkupCall{
		userID:      ctx.UserID(),
		messageID:   messageID,
		replyMarkup: replyMarkup,
	})
	return m.editMessageReplyMarkupError
}

func (m *mockMessageCleaner) getDeleteMessageCalls() []messageCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]messageCall, len(m.deleteMessageCalls))
	copy(calls, m.deleteMessageCalls)
	return calls
}

// getEditMessageReplyMarkupCalls method removed as it's unused

func (m *mockMessageCleaner) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteMessageCalls = nil
	m.editMessageReplyMarkupCalls = nil
	m.deleteMessageError = nil
	m.editMessageReplyMarkupError = nil
}

// Test helper functions

func createTestFlowManager() (*flowManager, *mockPromptSender, *mockPromptKeyboardActions, *mockMessageCleaner) {
	config := &FlowConfig{
		ExitCommands:        []string{"/exit", "/cancel"},
		ExitMessage:         "Flow cancelled",
		AllowGlobalCommands: false,
		HelpCommands:        []string{"/help"},
		OnProcessAction:     ProcessKeepMessage,
	}

	mockSender := &mockPromptSender{}
	mockKeyboard := &mockPromptKeyboardActions{callbackData: make(map[int64]map[string]interface{})}
	mockCleaner := &mockMessageCleaner{}

	fm := newFlowManager(config, mockSender, mockKeyboard, mockCleaner)
	return fm, mockSender, mockKeyboard, mockCleaner
}

func createTestFlow() *Flow {
	return &Flow{
		Name: "test-flow",
		Steps: map[string]*flowStep{
			"step1": {
				Name: "step1",
				PromptConfig: &PromptConfig{
					Message: "Enter your name:",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					if input == "" {
						return Retry().WithPrompt("Please enter a valid name")
					}
					ctx.Set("name", input)
					return NextStep()
				},
			},
			"step2": {
				Name: "step2",
				PromptConfig: &PromptConfig{
					Message: "Enter your age:",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					if input == "back" {
						return GoToStep("step1")
					}
					ctx.Set("age", input)
					return CompleteFlow()
				},
			},
		},
		Order: []string{"step1", "step2"},
		OnComplete: func(ctx *Context) error {
			return nil
		},
		OnError: OnErrorCancel("Test flow error"),
		Timeout: time.Minute * 10,
	}
}

func createFlowTestContext(userID int64, messageText string) *Context {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			MessageID: 123,
			From:      &tgbotapi.User{ID: userID},
			Date:      int(time.Now().Unix()),
			Chat:      &tgbotapi.Chat{ID: userID, Type: "private"},
			Text:      messageText,
		},
	}

	// Include mock telegram client to prevent nil pointer panics
	mockClient := &flowTestTelegramClient{}

	return &Context{
		telegramClient: mockClient,
		update:         update,
		data:           make(map[string]interface{}),
		userID:         userID,
		chatID:         userID,
	}
}

func createFlowTestCallbackContext(userID int64, callbackData string) *Context {
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback123",
			From: &tgbotapi.User{ID: userID},
			Data: callbackData,
			Message: &tgbotapi.Message{
				MessageID: 456,
				From:      &tgbotapi.User{ID: 123456789}, // Bot user
				Date:      int(time.Now().Unix()),
				Chat:      &tgbotapi.Chat{ID: userID, Type: "private"},
				Text:      "Previous message",
			},
		},
	}
	// Create a mock telegram client that doesn't panic
	mockClient := &flowTestTelegramClient{}

	ctx := &Context{
		telegramClient: mockClient,
		update:         update,
		data:           make(map[string]interface{}),
		userID:         userID,
		chatID:         userID,
	}

	return ctx
}

// Mock telegram client for flow tests
type flowTestTelegramClient struct{}

func (m *flowTestTelegramClient) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return tgbotapi.Message{}, nil
}

func (m *flowTestTelegramClient) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *flowTestTelegramClient) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return make(tgbotapi.UpdatesChannel)
}

func (m *flowTestTelegramClient) GetMe() (tgbotapi.User, error) {
	return tgbotapi.User{ID: 123456789, UserName: "test_bot"}, nil
}

// Test cases

func TestNewFlowManager(t *testing.T) {
	config := &FlowConfig{
		ExitCommands:        []string{"/exit"},
		ExitMessage:         "Goodbye",
		AllowGlobalCommands: true,
	}

	mockSender := &mockPromptSender{}
	mockKeyboard := &mockPromptKeyboardActions{}
	mockCleaner := &mockMessageCleaner{}

	fm := newFlowManager(config, mockSender, mockKeyboard, mockCleaner)

	if fm == nil {
		t.Fatal("newFlowManager returned nil")
	}

	if fm.flows == nil {
		t.Error("flows map not initialized")
	}

	if fm.userFlows == nil {
		t.Error("userFlows map not initialized")
	}

	if fm.flowConfig != config {
		t.Error("flowConfig not set correctly")
	}

	if fm.promptSender != mockSender {
		t.Error("promptSender not set correctly")
	}

	if fm.keyboardAccess != mockKeyboard {
		t.Error("keyboardAccess not set correctly")
	}

	if fm.messageCleaner != mockCleaner {
		t.Error("messageCleaner not set correctly")
	}
}

func TestRegisterFlow(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	flow := createTestFlow()

	fm.registerFlow(flow)

	if len(fm.flows) != 1 {
		t.Errorf("Expected 1 flow, got %d", len(fm.flows))
	}

	storedFlow, exists := fm.flows["test-flow"]
	if !exists {
		t.Error("Flow not found after registration")
	}

	if storedFlow != flow {
		t.Error("Stored flow is not the same as registered flow")
	}
}

func TestStartFlow(t *testing.T) {
	tests := []struct {
		name          string
		flowName      string
		userID        int64
		ctx           *Context
		expectedError bool
		errorContains string
	}{
		{
			name:          "valid flow with context",
			flowName:      "test-flow",
			userID:        12345,
			ctx:           createFlowTestContext(12345, ""),
			expectedError: false,
		},
		{
			name:          "valid flow without context",
			flowName:      "test-flow",
			userID:        12345,
			ctx:           nil,
			expectedError: false,
		},
		{
			name:          "non-existent flow",
			flowName:      "non-existent",
			userID:        12345,
			ctx:           createFlowTestContext(12345, ""),
			expectedError: true,
			errorContains: "flow non-existent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, mockSender, _, _ := createTestFlowManager()
			flow := createTestFlow()
			fm.registerFlow(flow)

			mockSender.reset()

			err := fm.startFlow(tt.userID, tt.flowName, tt.ctx)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check user flow state was created
			if !fm.isUserInFlow(tt.userID) {
				t.Error("User should be in flow after startFlow")
			}

			// Check if prompt was sent when context provided
			if tt.ctx != nil {
				calls := mockSender.getComposeAndSendCalls()
				if len(calls) != 1 {
					t.Errorf("Expected 1 prompt call, got %d", len(calls))
				}
			}
		})
	}
}

func TestStartFlowWithEmptySteps(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()

	// Create flow with no steps
	emptyFlow := &Flow{
		Name:  "empty-flow",
		Steps: map[string]*flowStep{},
		Order: []string{}, // Empty order
	}
	fm.registerFlow(emptyFlow)

	ctx := createFlowTestContext(12345, "")
	err := fm.startFlow(12345, "empty-flow", ctx)

	if err == nil {
		t.Error("Expected error for flow with no steps")
	}

	if !contains(err.Error(), "has no steps") {
		t.Errorf("Expected error about no steps, got: %v", err)
	}
}

func TestIsUserInFlow(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	flow := createTestFlow()
	fm.registerFlow(flow)

	userID := int64(12345)

	// User not in flow initially
	if fm.isUserInFlow(userID) {
		t.Error("User should not be in flow initially")
	}

	// Start flow
	ctx := createFlowTestContext(userID, "")
	err := fm.startFlow(userID, "test-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	// User should be in flow now
	if !fm.isUserInFlow(userID) {
		t.Error("User should be in flow after startFlow")
	}

	// Different user should not be in flow
	if fm.isUserInFlow(67890) {
		t.Error("Different user should not be in flow")
	}
}

func TestCancelFlow(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	flow := createTestFlow()
	fm.registerFlow(flow)

	userID := int64(12345)
	ctx := createFlowTestContext(userID, "")

	// Start flow
	err := fm.startFlow(userID, "test-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	// Verify user is in flow
	if !fm.isUserInFlow(userID) {
		t.Error("User should be in flow")
	}

	// Cancel flow
	fm.cancelFlow(userID)

	// Verify user is no longer in flow
	if fm.isUserInFlow(userID) {
		t.Error("User should not be in flow after cancellation")
	}
}

func TestSetUserFlowData(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	flow := createTestFlow()
	fm.registerFlow(flow)

	userID := int64(12345)
	ctx := createFlowTestContext(userID, "")

	// Test setting data for user not in flow
	err := fm.setUserFlowData(userID, "key1", "value1")
	if err == nil {
		t.Error("Expected error when setting data for user not in flow")
	}

	// Start flow
	err = fm.startFlow(userID, "test-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	// Test setting data for user in flow
	err = fm.setUserFlowData(userID, "key1", "value1")
	if err != nil {
		t.Errorf("Unexpected error setting flow data: %v", err)
	}

	// Verify data was set
	value, exists := fm.getUserFlowData(userID, "key1")
	if !exists {
		t.Error("Flow data should exist after setting")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got '%v'", value)
	}
}

func TestGetUserFlowData(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	flow := createTestFlow()
	fm.registerFlow(flow)

	userID := int64(12345)

	// Test getting data for user not in flow
	value, exists := fm.getUserFlowData(userID, "key1")
	if exists {
		t.Error("Should not find data for user not in flow")
	}
	if value != nil {
		t.Error("Value should be nil for user not in flow")
	}

	// Start flow
	ctx := createFlowTestContext(userID, "")
	err := fm.startFlow(userID, "test-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	// Test getting non-existent key
	_, exists = fm.getUserFlowData(userID, "non-existent")
	if exists {
		t.Error("Should not find non-existent key")
	}

	// Set and get data
	err = fm.setUserFlowData(userID, "testkey", "testvalue")
	if err != nil {
		t.Fatalf("Failed to set flow data: %v", err)
	}

	value, exists = fm.getUserFlowData(userID, "testkey")
	if !exists {
		t.Error("Should find existing key")
	}
	if value != "testvalue" {
		t.Errorf("Expected 'testvalue', got '%v'", value)
	}
}

func TestHandleUpdateUserNotInFlow(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	ctx := createFlowTestContext(12345, "test message")

	handled, err := fm.HandleUpdate(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if handled {
		t.Error("Should not handle update for user not in flow")
	}
}

func TestHandleUpdateUserInFlow(t *testing.T) {
	fm, mockSender, _, _ := createTestFlowManager()
	flow := createTestFlow()
	fm.registerFlow(flow)

	userID := int64(12345)
	ctx := createFlowTestContext(userID, "")

	// Start flow
	err := fm.startFlow(userID, "test-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	mockSender.reset()

	// Test valid input
	ctx = createFlowTestContext(userID, "John Doe")
	handled, err := fm.HandleUpdate(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !handled {
		t.Error("Should handle update for user in flow")
	}

	// Check if prompt was sent for next step
	calls := mockSender.getComposeAndSendCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 prompt call for next step, got %d", len(calls))
	}
}

func TestHandleUpdateWithCallback(t *testing.T) {
	fm, mockSender, mockKeyboard, _ := createTestFlowManager()

	// Create flow with callback handling
	flow := &Flow{
		Name: "callback-flow",
		Steps: map[string]*flowStep{
			"step1": {
				Name: "step1",
				PromptConfig: &PromptConfig{
					Message: "Choose option:",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					if buttonClick != nil {
						if buttonClick.Data == "option1" {
							return CompleteFlow()
						}
					}
					return Retry()
				},
			},
		},
		Order:      []string{"step1"},
		OnComplete: func(ctx *Context) error { return nil },
	}
	fm.registerFlow(flow)

	userID := int64(12345)
	ctx := createFlowTestContext(userID, "")

	// Start flow
	err := fm.startFlow(userID, "callback-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	mockSender.reset()

	// Set up callback data
	mockKeyboard.setCallbackData(userID, "callback123", "option1")

	// Test callback update
	ctx = createFlowTestCallbackContext(userID, "callback123")
	handled, err := fm.HandleUpdate(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !handled {
		t.Error("Should handle callback update")
	}

	// User should no longer be in flow (completed)
	if fm.isUserInFlow(userID) {
		t.Error("User should not be in flow after completion")
	}
}

func TestHandleUpdateFlowNotFound(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()

	userID := int64(12345)

	// Manually add user flow state with non-existent flow
	fm.muUserFlows.Lock()
	fm.userFlows[userID] = &userFlowState{
		FlowName:    "non-existent-flow",
		CurrentStep: "step1",
		Data:        make(map[string]interface{}),
		StartedAt:   time.Now(),
		LastActive:  time.Now(),
	}
	fm.muUserFlows.Unlock()

	ctx := createFlowTestContext(userID, "test")
	handled, err := fm.HandleUpdate(ctx)

	if err == nil {
		t.Error("Expected error for non-existent flow")
	}

	if handled {
		t.Error("Should return handled=false when flow not found")
	}

	// User should be removed from flow
	if fm.isUserInFlow(userID) {
		t.Error("User should be removed from flow when flow not found")
	}
}

func TestHandleUpdateStepNotFound(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	flow := createTestFlow()
	fm.registerFlow(flow)

	userID := int64(12345)

	// Manually add user flow state with non-existent step
	fm.muUserFlows.Lock()
	fm.userFlows[userID] = &userFlowState{
		FlowName:    "test-flow",
		CurrentStep: "non-existent-step",
		Data:        make(map[string]interface{}),
		StartedAt:   time.Now(),
		LastActive:  time.Now(),
	}
	fm.muUserFlows.Unlock()

	ctx := createFlowTestContext(userID, "test")
	handled, err := fm.HandleUpdate(ctx)

	if err == nil {
		t.Error("Expected error for non-existent step")
	}

	if handled {
		t.Error("Should return handled=false when step not found")
	}

	// User should be removed from flow
	if fm.isUserInFlow(userID) {
		t.Error("User should be removed from flow when step not found")
	}
}

func TestConcurrentAccess(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()
	flow := createTestFlow()
	fm.registerFlow(flow)

	const numGoroutines = 10
	const userIDBase = 10000

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*2)

	// Test concurrent setUserFlowData and getUserFlowData
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)
		userID := int64(userIDBase + i)

		// Start flow for each user
		ctx := createFlowTestContext(userID, "")
		err := fm.startFlow(userID, "test-flow", ctx)
		if err != nil {
			t.Fatalf("Failed to start flow for user %d: %v", userID, err)
		}

		// Goroutine for setting data
		go func(uid int64, index int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", index)
			value := fmt.Sprintf("value%d", index)
			if err := fm.setUserFlowData(uid, key, value); err != nil {
				errors <- err
			}
		}(userID, i)

		// Goroutine for getting data
		go func(uid int64, index int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", index)
			// Add small delay to allow setter to run first
			time.Sleep(time.Millisecond)
			_, _ = fm.getUserFlowData(uid, key)
			// Getting data shouldn't error, even if key doesn't exist yet
		}(userID, i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestMessageCleanerInteractions(t *testing.T) {
	fm, _, _, mockCleaner := createTestFlowManager()

	// Create flow with message deletion action
	flow := &Flow{
		Name: "cleanup-flow",
		Steps: map[string]*flowStep{
			"step1": {
				Name: "step1",
				PromptConfig: &PromptConfig{
					Message: "Click button:",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					return CompleteFlow()
				},
			},
		},
		Order:           []string{"step1"},
		OnComplete:      func(ctx *Context) error { return nil },
		OnProcessAction: ProcessDeleteMessage,
	}
	fm.registerFlow(flow)

	userID := int64(12345)
	ctx := createFlowTestContext(userID, "")

	// Start flow
	err := fm.startFlow(userID, "cleanup-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	mockCleaner.reset()

	// Test callback that should trigger message deletion
	ctx = createFlowTestCallbackContext(userID, "test-callback")
	handled, err := fm.HandleUpdate(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !handled {
		t.Error("Should handle callback update")
	}

	// Check if DeleteMessage was called
	deleteCalls := mockCleaner.getDeleteMessageCalls()
	if len(deleteCalls) != 1 {
		t.Errorf("Expected 1 delete message call, got %d", len(deleteCalls))
	}

	if len(deleteCalls) > 0 && deleteCalls[0].messageID != 456 {
		t.Errorf("Expected message ID 456 to be deleted, got %d", deleteCalls[0].messageID)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name             string
		errorStrategy    *ErrorConfig
		senderError      error
		expectUserInFlow bool
	}{
		{
			name:             "cancel on error",
			errorStrategy:    OnErrorCancel("Flow cancelled due to error"),
			senderError:      errors.New("prompt send failed"),
			expectUserInFlow: false,
		},
		{
			name:             "retry on error",
			errorStrategy:    OnErrorRetry("Retrying step"),
			senderError:      errors.New("prompt send failed"),
			expectUserInFlow: true,
		},
		{
			name:             "ignore on error",
			errorStrategy:    OnErrorIgnore("Ignoring error"),
			senderError:      errors.New("prompt send failed"),
			expectUserInFlow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, mockSender, _, _ := createTestFlowManager()

			flow := createTestFlow()
			flow.OnError = tt.errorStrategy
			fm.registerFlow(flow)

			userID := int64(12345)
			ctx := createFlowTestContext(userID, "")

			// Set sender to return error
			if tt.errorStrategy.Action == errorStrategyIgnore {
				// For ignore strategy, only error once to prevent infinite loop
				mockSender.setErrorOnce(tt.senderError)
			} else {
				mockSender.composeAndSendError = tt.senderError
			}

			// Start flow (this should trigger the error)
			_ = fm.startFlow(userID, "test-flow", ctx)

			// Check if user is still in flow based on error strategy
			if fm.isUserInFlow(userID) != tt.expectUserInFlow {
				t.Errorf("Expected user in flow: %v, got: %v", tt.expectUserInFlow, fm.isUserInFlow(userID))
			}

			// Reset sender error for cleanup
			mockSender.reset()
		})
	}
}

// Helper function
