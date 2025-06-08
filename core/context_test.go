package teleflow

import (
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TestableContext struct {
	*Context
	testBot *TestBot
}

type TestBot struct {
	api         *TestBotAPI
	flowManager *TestFlowManager
}

type TestFlowManager struct {
	userFlowStates map[int64]*TestUserFlowState
	setDataFunc    func(userID int64, key string, value interface{}) error
	getDataFunc    func(userID int64, key string) (interface{}, bool)
}

type TestUserFlowState struct {
	UserID   int64
	FlowName string
	Data     map[string]interface{}
}

func NewTestFlowManager() *TestFlowManager {
	return &TestFlowManager{
		userFlowStates: make(map[int64]*TestUserFlowState),
	}
}

func (tfm *TestFlowManager) isUserInFlow(userID int64) bool {
	_, exists := tfm.userFlowStates[userID]
	return exists
}

func (tfm *TestFlowManager) setUserFlowData(userID int64, key string, value interface{}) error {
	if tfm.setDataFunc != nil {
		return tfm.setDataFunc(userID, key, value)
	}

	state, exists := tfm.userFlowStates[userID]
	if !exists {
		return errors.New("user not in a flow, cannot set flow data")
	}

	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data[key] = value
	return nil
}

func (tfm *TestFlowManager) getUserFlowData(userID int64, key string) (interface{}, bool) {
	if tfm.getDataFunc != nil {
		return tfm.getDataFunc(userID, key)
	}

	state, exists := tfm.userFlowStates[userID]
	if !exists {
		return nil, false
	}

	if state.Data == nil {
		return nil, false
	}

	value, found := state.Data[key]
	return value, found
}

func (tfm *TestFlowManager) addUserToFlow(userID int64, flowName string) {
	tfm.userFlowStates[userID] = &TestUserFlowState{
		UserID:   userID,
		FlowName: flowName,
		Data:     make(map[string]interface{}),
	}
}

func (tc *TestableContext) SendReplyKeyboard(buttons []string, buttonsPerRow int, options ...ReplyKeyboardOption) error {
	if tc.testBot == nil || tc.testBot.api == nil {
		return errors.New("bot or bot API not available in context for SendReplyKeyboard")
	}

	tempReplyKeyboard := BuildReplyKeyboard(buttons, buttonsPerRow)
	tgAPIReplyKeyboard := tempReplyKeyboard.ToTgbotapi()

	for _, opt := range options {
		opt(&tgAPIReplyKeyboard)
	}

	msg := tgbotapi.NewMessage(tc.ChatID(), "\u200B")
	msg.ReplyMarkup = tgAPIReplyKeyboard
	_, err := tc.testBot.api.Send(msg)
	return err
}

func (tc *TestableContext) SetFlowData(key string, value interface{}) error {
	if !tc.IsUserInFlow() {
		return errors.New("user not in a flow, cannot set flow data")
	}
	return tc.testBot.flowManager.setUserFlowData(tc.UserID(), key, value)
}

func (tc *TestableContext) GetFlowData(key string) (interface{}, bool) {
	if !tc.IsUserInFlow() {
		return nil, false
	}
	return tc.testBot.flowManager.getUserFlowData(tc.UserID(), key)
}

func (tc *TestableContext) IsUserInFlow() bool {
	return tc.testBot.flowManager.isUserInFlow(tc.UserID())
}

func NewTestableContext(botAPI *TestBotAPI, flowManager *TestFlowManager) *TestableContext {
	testBot := &TestBot{
		api:         botAPI,
		flowManager: flowManager,
	}

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123},
			From: &tgbotapi.User{ID: 456},
		},
	}

	baseContext := newContext(&Bot{}, update)

	return &TestableContext{
		Context: baseContext,
		testBot: testBot,
	}
}

func TestContext_SendReplyKeyboard_BasicFunctionality(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	buttons := []string{"Button 1", "Button 2", "Button 3"}
	buttonsPerRow := 2

	err := ctx.SendReplyKeyboard(buttons, buttonsPerRow)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(botAPI.SendCalls) != 1 {
		t.Errorf("Expected 1 Send call, got %d", len(botAPI.SendCalls))
	}

	if msg, ok := botAPI.SendCalls[0].(tgbotapi.MessageConfig); ok {
		if msg.Text != "\u200B" {
			t.Errorf("Expected invisible character, got '%s'", msg.Text)
		}
		if msg.ChatID != ctx.ChatID() {
			t.Errorf("Expected ChatID %d, got %d", ctx.ChatID(), msg.ChatID)
		}

		if kb, ok := msg.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup); ok {
			if len(kb.Keyboard) != 2 {
				t.Errorf("Expected 2 rows, got %d", len(kb.Keyboard))
			}
			if len(kb.Keyboard[0]) != 2 {
				t.Errorf("Expected 2 buttons in first row, got %d", len(kb.Keyboard[0]))
			}
			if len(kb.Keyboard[1]) != 1 {
				t.Errorf("Expected 1 button in second row, got %d", len(kb.Keyboard[1]))
			}
		} else {
			t.Error("Expected ReplyKeyboardMarkup")
		}
	} else {
		t.Error("Expected MessageConfig type")
	}
}

func TestContext_SendReplyKeyboard_WithOptions(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	buttons := []string{"Yes", "No"}
	buttonsPerRow := 2

	err := ctx.SendReplyKeyboard(buttons, buttonsPerRow, WithResize(), WithOneTime(), WithPlaceholder("Choose an option"))

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if msg, ok := botAPI.SendCalls[0].(tgbotapi.MessageConfig); ok {
		if kb, ok := msg.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup); ok {
			if !kb.ResizeKeyboard {
				t.Error("Expected ResizeKeyboard to be true")
			}
			if !kb.OneTimeKeyboard {
				t.Error("Expected OneTimeKeyboard to be true")
			}
			if kb.InputFieldPlaceholder != "Choose an option" {
				t.Errorf("Expected placeholder 'Choose an option', got '%s'", kb.InputFieldPlaceholder)
			}
		} else {
			t.Error("Expected ReplyKeyboardMarkup")
		}
	} else {
		t.Error("Expected MessageConfig type")
	}
}

func TestContext_SendReplyKeyboard_EmptyButtons(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	buttons := []string{}
	buttonsPerRow := 2

	err := ctx.SendReplyKeyboard(buttons, buttonsPerRow)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if msg, ok := botAPI.SendCalls[0].(tgbotapi.MessageConfig); ok {
		if kb, ok := msg.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup); ok {
			if len(kb.Keyboard) != 0 {
				t.Errorf("Expected 0 rows for empty buttons, got %d", len(kb.Keyboard))
			}
		} else {
			t.Error("Expected ReplyKeyboardMarkup")
		}
	} else {
		t.Error("Expected MessageConfig type")
	}
}

func TestContext_SendReplyKeyboard_BotAPIError(t *testing.T) {

	botAPI := &TestBotAPI{
		SendFunc: func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, errors.New("API error")
		},
	}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	buttons := []string{"Test"}
	buttonsPerRow := 1

	err := ctx.SendReplyKeyboard(buttons, buttonsPerRow)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "API error" {
		t.Errorf("Expected 'API error', got '%s'", err.Error())
	}
}

func TestContext_SendReplyKeyboard_NilBot(t *testing.T) {

	ctx := NewTestableContext(nil, NewTestFlowManager())

	buttons := []string{"Test"}
	buttonsPerRow := 1

	err := ctx.SendReplyKeyboard(buttons, buttonsPerRow)

	if err == nil {
		t.Error("Expected error for nil bot, got nil")
	}
	if err != nil && err.Error() != "bot or bot API not available in context for SendReplyKeyboard" {
		t.Errorf("Expected specific error message, got '%s'", err.Error())
	}
}

func TestContext_SetFlowData_UserInFlow(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	flowManager.addUserToFlow(ctx.UserID(), "test_flow")

	err := ctx.SetFlowData("test_key", "test_value")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	state := flowManager.userFlowStates[ctx.UserID()]
	if state == nil {
		t.Error("Expected user flow state to exist")
		return
	}

	if state.Data["test_key"] != "test_value" {
		t.Errorf("Expected 'test_value', got %v", state.Data["test_key"])
	}
}

func TestContext_SetFlowData_UserNotInFlow(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	err := ctx.SetFlowData("test_key", "test_value")

	if err == nil {
		t.Error("Expected error for user not in flow, got nil")
	}
	if err != nil && err.Error() != "user not in a flow, cannot set flow data" {
		t.Errorf("Expected specific error message, got '%s'", err.Error())
	}
}

func TestContext_SetFlowData_FlowManagerError(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	flowManager.setDataFunc = func(userID int64, key string, value interface{}) error {
		return errors.New("flow manager error")
	}
	ctx := NewTestableContext(botAPI, flowManager)

	flowManager.addUserToFlow(ctx.UserID(), "test_flow")

	err := ctx.SetFlowData("test_key", "test_value")

	if err == nil {
		t.Error("Expected error from flow manager, got nil")
	}
	if err != nil && err.Error() != "flow manager error" {
		t.Errorf("Expected 'flow manager error', got '%s'", err.Error())
	}
}

func TestContext_GetFlowData_UserInFlow(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	flowManager.addUserToFlow(ctx.UserID(), "test_flow")
	flowManager.userFlowStates[ctx.UserID()].Data["existing_key"] = "existing_value"

	value, found := ctx.GetFlowData("existing_key")

	if !found {
		t.Error("Expected to find flow data")
	}
	if value != "existing_value" {
		t.Errorf("Expected 'existing_value', got %v", value)
	}
}

func TestContext_GetFlowData_UserNotInFlow(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	value, found := ctx.GetFlowData("test_key")

	if found {
		t.Error("Expected not to find flow data for user not in flow")
	}
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}
}

func TestContext_GetFlowData_NonExistingKey(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	flowManager.addUserToFlow(ctx.UserID(), "test_flow")

	value, found := ctx.GetFlowData("non_existing_key")

	if found {
		t.Error("Expected not to find non-existing flow data key")
	}
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}
}

func TestContext_GetFlowData_FlowManagerCustomLogic(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	flowManager.getDataFunc = func(userID int64, key string) (interface{}, bool) {
		if key == "special_key" {
			return "special_value", true
		}
		return nil, false
	}
	ctx := NewTestableContext(botAPI, flowManager)

	flowManager.addUserToFlow(ctx.UserID(), "test_flow")

	value, found := ctx.GetFlowData("special_key")

	if !found {
		t.Error("Expected to find special key")
	}
	if value != "special_value" {
		t.Errorf("Expected 'special_value', got %v", value)
	}
}

func TestContext_FlowDataIntegration(t *testing.T) {

	botAPI := &TestBotAPI{}
	flowManager := NewTestFlowManager()
	ctx := NewTestableContext(botAPI, flowManager)

	flowManager.addUserToFlow(ctx.UserID(), "test_flow")

	testData := map[string]interface{}{
		"string_key": "string_value",
		"int_key":    42,
		"bool_key":   true,
		"map_key":    map[string]string{"nested": "value"},
	}

	for key, value := range testData {
		err := ctx.SetFlowData(key, value)
		if err != nil {
			t.Errorf("Failed to set flow data for key '%s': %v", key, err)
		}
	}

	for key, expectedValue := range testData {
		actualValue, found := ctx.GetFlowData(key)
		if !found {
			t.Errorf("Expected to find flow data for key '%s'", key)
			continue
		}

		switch expected := expectedValue.(type) {
		case map[string]string:
			if actual, ok := actualValue.(map[string]string); ok {
				if len(actual) != len(expected) {
					t.Errorf("Map length mismatch for key '%s': expected %d, got %d", key, len(expected), len(actual))
				}
				for k, v := range expected {
					if actual[k] != v {
						t.Errorf("Map value mismatch for key '%s': expected %s, got %s", key, v, actual[k])
					}
				}
			} else {
				t.Errorf("Type mismatch for key '%s': expected map[string]string, got %T", key, actualValue)
			}
		default:
			if actualValue != expectedValue {
				t.Errorf("Value mismatch for key '%s': expected %v, got %v", key, expectedValue, actualValue)
			}
		}
	}

	err := ctx.SetFlowData("string_key", "new_string_value")
	if err != nil {
		t.Errorf("Failed to overwrite flow data: %v", err)
	}

	newValue, found := ctx.GetFlowData("string_key")
	if !found {
		t.Error("Expected to find overwritten flow data")
	}
	if newValue != "new_string_value" {
		t.Errorf("Expected 'new_string_value', got %v", newValue)
	}
}
