package teleflow

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestPromptKeyboardHandler_BuildKeyboard_ValidKeyboardFunc(t *testing.T) {
	handler := newPromptKeyboardHandler()
	ctx := createTestContext()

	keyboardFunc := func(ctx *Context) *PromptKeyboardBuilder {
		return NewPromptKeyboard().
			ButtonCallback("Button 1", "data1").
			ButtonCallback("Button 2", "data2").
			Row().
			ButtonCallback("Button 3", "data3")
	}

	result, err := handler.BuildKeyboard(ctx, keyboardFunc)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
		return
	}

	keyboard, ok := result.(tgbotapi.InlineKeyboardMarkup)
	if !ok {
		t.Errorf("Expected tgbotapi.InlineKeyboardMarkup, got %T", result)
		return
	}

	if len(keyboard.InlineKeyboard) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(keyboard.InlineKeyboard))
	}

	if len(keyboard.InlineKeyboard[0]) != 2 {
		t.Errorf("Expected 2 buttons in first row, got %d", len(keyboard.InlineKeyboard[0]))
	}

	if len(keyboard.InlineKeyboard[1]) != 1 {
		t.Errorf("Expected 1 button in second row, got %d", len(keyboard.InlineKeyboard[1]))
	}

	userMappings := handler.userUUIDMappings[ctx.UserID()]
	if userMappings == nil {
		t.Error("Expected user mappings to be created")
		return
	}

	if len(userMappings) != 3 {
		t.Errorf("Expected 3 UUID mappings, got %d", len(userMappings))
	}

	if keyboard.InlineKeyboard[0][0].Text != "Button 1" {
		t.Errorf("Expected 'Button 1', got '%s'", keyboard.InlineKeyboard[0][0].Text)
	}
	if keyboard.InlineKeyboard[0][1].Text != "Button 2" {
		t.Errorf("Expected 'Button 2', got '%s'", keyboard.InlineKeyboard[0][1].Text)
	}
	if keyboard.InlineKeyboard[1][0].Text != "Button 3" {
		t.Errorf("Expected 'Button 3', got '%s'", keyboard.InlineKeyboard[1][0].Text)
	}
}

func TestPromptKeyboardHandler_BuildKeyboard_NilKeyboardFunc(t *testing.T) {
	handler := newPromptKeyboardHandler()
	ctx := createTestContext()

	result, err := handler.BuildKeyboard(ctx, nil)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	userMappings := handler.userUUIDMappings[ctx.UserID()]
	if userMappings != nil {
		t.Error("Expected no user mappings to be created")
	}
}

func TestPromptKeyboardHandler_BuildKeyboard_KeyboardFuncReturnsNil(t *testing.T) {
	handler := newPromptKeyboardHandler()
	ctx := createTestContext()

	keyboardFunc := func(ctx *Context) *PromptKeyboardBuilder {
		return nil
	}

	result, err := handler.BuildKeyboard(ctx, keyboardFunc)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	userMappings := handler.userUUIDMappings[ctx.UserID()]
	if userMappings != nil {
		t.Error("Expected no user mappings to be created")
	}
}

func TestPromptKeyboardHandler_BuildKeyboard_InvalidKeyboard(t *testing.T) {
	handler := newPromptKeyboardHandler()
	ctx := createTestContext()

	keyboardFunc := func(ctx *Context) *PromptKeyboardBuilder {
		return NewPromptKeyboard()
	}

	result, err := handler.BuildKeyboard(ctx, keyboardFunc)

	if err == nil {
		t.Error("Expected error for invalid keyboard, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	userMappings := handler.userUUIDMappings[ctx.UserID()]
	if userMappings != nil {
		t.Error("Expected no user mappings to be created for invalid keyboard")
	}
}

func TestPromptKeyboardHandler_GetCallbackData_ExistingUUID(t *testing.T) {
	handler := newPromptKeyboardHandler()
	ctx := createTestContext()
	userID := ctx.UserID()

	keyboardFunc := func(ctx *Context) *PromptKeyboardBuilder {
		return NewPromptKeyboard().
			ButtonCallback("Test Button", "test_data").
			ButtonCallback("Another Button", map[string]string{"key": "value"})
	}

	result, err := handler.BuildKeyboard(ctx, keyboardFunc)
	if err != nil {
		t.Fatalf("Failed to build keyboard: %v", err)
	}

	keyboard := result.(tgbotapi.InlineKeyboardMarkup)
	uuid1 := *keyboard.InlineKeyboard[0][0].CallbackData
	uuid2 := *keyboard.InlineKeyboard[0][1].CallbackData

	data1, found1 := handler.GetCallbackData(userID, uuid1)
	if !found1 {
		t.Error("Expected to find callback data for uuid1")
	}
	if data1 != "test_data" {
		t.Errorf("Expected 'test_data', got %v", data1)
	}

	data2, found2 := handler.GetCallbackData(userID, uuid2)
	if !found2 {
		t.Error("Expected to find callback data for uuid2")
	}
	if mapData, ok := data2.(map[string]string); ok {
		if mapData["key"] != "value" {
			t.Errorf("Expected map with key='value', got %v", mapData)
		}
	} else {
		t.Errorf("Expected map[string]string, got %T", data2)
	}
}

func TestPromptKeyboardHandler_GetCallbackData_NonExistingUUID(t *testing.T) {
	handler := newPromptKeyboardHandler()
	userID := int64(12345)

	data, found := handler.GetCallbackData(userID, "non-existing-uuid")

	if found {
		t.Error("Expected not to find callback data for non-existing UUID")
	}
	if data != nil {
		t.Errorf("Expected nil data, got %v", data)
	}
}

func TestPromptKeyboardHandler_GetCallbackData_NonExistingUser(t *testing.T) {
	handler := newPromptKeyboardHandler()
	ctx := createTestContext()

	keyboardFunc := func(ctx *Context) *PromptKeyboardBuilder {
		return NewPromptKeyboard().ButtonCallback("Test", "data")
	}

	result, err := handler.BuildKeyboard(ctx, keyboardFunc)
	if err != nil {
		t.Fatalf("Failed to build keyboard: %v", err)
	}

	keyboard := result.(tgbotapi.InlineKeyboardMarkup)
	uuid := *keyboard.InlineKeyboard[0][0].CallbackData

	differentUserID := ctx.UserID() + 1000
	data, found := handler.GetCallbackData(differentUserID, uuid)

	if found {
		t.Error("Expected not to find callback data for different user")
	}
	if data != nil {
		t.Errorf("Expected nil data, got %v", data)
	}
}

func TestPromptKeyboardHandler_CleanupUserMappings(t *testing.T) {
	handler := newPromptKeyboardHandler()
	ctx := createTestContext()
	userID := ctx.UserID()

	keyboardFunc := func(ctx *Context) *PromptKeyboardBuilder {
		return NewPromptKeyboard().
			ButtonCallback("Button 1", "data1").
			ButtonCallback("Button 2", "data2")
	}

	_, err := handler.BuildKeyboard(ctx, keyboardFunc)
	if err != nil {
		t.Fatalf("Failed to build keyboard: %v", err)
	}

	userMappings := handler.userUUIDMappings[userID]
	if userMappings == nil || len(userMappings) != 2 {
		t.Fatalf("Expected 2 UUID mappings, got %d", len(userMappings))
	}

	handler.CleanupUserMappings(userID)

	userMappings = handler.userUUIDMappings[userID]
	if userMappings != nil {
		t.Error("Expected user mappings to be removed")
	}
}

func TestPromptKeyboardHandler_CleanupUserMappings_NonExistingUser(t *testing.T) {
	handler := newPromptKeyboardHandler()

	nonExistingUserID := int64(99999)

	handler.CleanupUserMappings(nonExistingUserID)

	if handler.userUUIDMappings == nil {
		t.Error("Expected userUUIDMappings to remain initialized")
	}
}

func TestPromptKeyboardHandler_MultipleUsers(t *testing.T) {
	handler := newPromptKeyboardHandler()

	update1 := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 123},
			From: &tgbotapi.User{ID: 456},
		},
	}
	ctx1 := newContext(&Bot{}, update1)

	update2 := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 789},
			From: &tgbotapi.User{ID: 101112},
		},
	}
	ctx2 := newContext(&Bot{}, update2)

	keyboardFunc1 := func(ctx *Context) *PromptKeyboardBuilder {
		return NewPromptKeyboard().ButtonCallback("User1 Button", "user1_data")
	}
	keyboardFunc2 := func(ctx *Context) *PromptKeyboardBuilder {
		return NewPromptKeyboard().ButtonCallback("User2 Button", "user2_data")
	}

	result1, err1 := handler.BuildKeyboard(ctx1, keyboardFunc1)
	if err1 != nil {
		t.Fatalf("Failed to build keyboard for user1: %v", err1)
	}

	result2, err2 := handler.BuildKeyboard(ctx2, keyboardFunc2)
	if err2 != nil {
		t.Fatalf("Failed to build keyboard for user2: %v", err2)
	}

	keyboard1 := result1.(tgbotapi.InlineKeyboardMarkup)
	uuid1 := *keyboard1.InlineKeyboard[0][0].CallbackData

	keyboard2 := result2.(tgbotapi.InlineKeyboardMarkup)
	uuid2 := *keyboard2.InlineKeyboard[0][0].CallbackData

	data1, found1 := handler.GetCallbackData(ctx1.UserID(), uuid1)
	if !found1 || data1 != "user1_data" {
		t.Errorf("User1 should access their own data, got %v, found=%v", data1, found1)
	}

	data2, found2 := handler.GetCallbackData(ctx2.UserID(), uuid2)
	if !found2 || data2 != "user2_data" {
		t.Errorf("User2 should access their own data, got %v, found=%v", data2, found2)
	}

	_, found1Cross := handler.GetCallbackData(ctx1.UserID(), uuid2)
	if found1Cross {
		t.Error("User1 should not access User2's data")
	}

	_, found2Cross := handler.GetCallbackData(ctx2.UserID(), uuid1)
	if found2Cross {
		t.Error("User2 should not access User1's data")
	}

	handler.CleanupUserMappings(ctx1.UserID())

	_, found1After := handler.GetCallbackData(ctx1.UserID(), uuid1)
	if found1After {
		t.Error("User1's data should be cleaned up")
	}

	data2After, found2After := handler.GetCallbackData(ctx2.UserID(), uuid2)
	if !found2After || data2After != "user2_data" {
		t.Error("User2's data should still exist after User1 cleanup")
	}
}
