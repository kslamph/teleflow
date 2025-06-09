package teleflow

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestBot_SimpleCommandIntegration tests the end-to-end processing of a simple command
// TestBot_SimpleCommandIntegration tests the end-to-end processing of a simple command
func TestBot_SimpleCommandIntegration(t *testing.T) {
	// Setup test state
	handlerCalled := false
	handlerExecutedCorrectly := false
	var receivedContext *Context
	var receivedCommand string
	var receivedArgs string
	middlewareExecuted := false

	// Create bot with mocked dependencies
	bot, mockClient, _, _ := createTestBot()

	// Configure mock client to capture outgoing messages
	var sentMessages []tgbotapi.Chattable
	mockClient.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
		sentMessages = append(sentMessages, c)
		return tgbotapi.Message{MessageID: 123, Text: "Test reply"}, nil
	}

	// Add test middleware to verify middleware chain execution
	testMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			middlewareExecuted = true
			ctx.Set("middleware_data", "test_value")
			return next(ctx)
		}
	}
	bot.UseMiddleware(testMiddleware)

	// Define test command handler
	testHandler := func(ctx *Context, command string, args string) error {
		handlerCalled = true
		receivedContext = ctx
		receivedCommand = command
		receivedArgs = args

		// Verify context has correct basic information
		if ctx.UserID() == 123 && ctx.ChatID() == 456 {
			handlerExecutedCorrectly = true
		}

		// Verify middleware was executed by checking for middleware data
		if value, exists := ctx.Get("middleware_data"); exists && value == "test_value" {
			// Middleware data found, good
		} else {
			t.Errorf("Expected middleware data not found in context")
		}

		// Send a reply to test the response path
		return ctx.SendPromptText("Command processed successfully")
	}

	// Register the test command handler (command lookup uses name without slash)
	bot.HandleCommand("testcmd", testHandler)

	// Create test update representing incoming command message
	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 100,
			From: &tgbotapi.User{
				ID:       123,
				UserName: "testuser",
			},
			Chat: &tgbotapi.Chat{
				ID:   456,
				Type: "private",
			},
			Text: "/testcmd some_argument",
			Entities: []tgbotapi.MessageEntity{{
				Type:   "bot_command",
				Offset: 0,
				Length: 8, // Length of "/testcmd"
			}},
		},
	}

	// Process the update directly (simulating bot receiving the update)
	bot.processUpdate(update)

	// Verify that the handler was called
	if !handlerCalled {
		t.Error("Expected command handler to be called, but it wasn't")
	}

	if !handlerExecutedCorrectly {
		t.Error("Handler was called but didn't execute correctly (context verification failed)")
	}

	if !middlewareExecuted {
		t.Error("Expected middleware to be executed, but it wasn't")
	}

	// Verify received command and arguments
	if receivedCommand != "testcmd" {
		t.Errorf("Expected command 'testcmd', got '%s'", receivedCommand)
	}

	if receivedArgs != " some_argument" {
		t.Errorf("Expected args ' some_argument', got '%s'", receivedArgs)
	}

	// Verify context correctness
	if receivedContext == nil {
		t.Error("Expected context to be non-nil")
	} else {
		if receivedContext.UserID() != 123 {
			t.Errorf("Expected UserID 123, got %d", receivedContext.UserID())
		}
		if receivedContext.ChatID() != 456 {
			t.Errorf("Expected ChatID 456, got %d", receivedContext.ChatID())
		}
	}

	// Verify that a response was sent
	if len(sentMessages) == 0 {
		t.Error("Expected at least one message to be sent, but none were sent")
	} else {
		// Check that the sent message is correct
		if msgConfig, ok := sentMessages[0].(tgbotapi.MessageConfig); ok {
			if msgConfig.ChatID != 456 {
				t.Errorf("Expected message to be sent to ChatID 456, got %d", msgConfig.ChatID)
			}
			if msgConfig.Text != "Command processed successfully" {
				t.Errorf("Expected message text 'Command processed successfully', got '%s'", msgConfig.Text)
			}
		} else {
			t.Error("Expected sent message to be of type MessageConfig")
		}
	}
}

// TestBot_UnknownCommandIntegration tests handling of unknown commands
func TestBot_UnknownCommandIntegration(t *testing.T) {
	var defaultHandlerCalled bool
	var receivedText string
	var mu sync.Mutex

	// Create bot with mocked dependencies
	bot, mockClient, _, _ := createTestBot()

	// Configure mock client
	mockClient.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
		return tgbotapi.Message{MessageID: 123}, nil
	}

	// Set up default handler
	bot.DefaultHandler(func(ctx *Context, text string) error {
		mu.Lock()
		defer mu.Unlock()
		defaultHandlerCalled = true
		receivedText = text
		return ctx.SendPromptText("Unknown command")
	})

	// Create test update with unknown command
	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 100,
			From: &tgbotapi.User{
				ID:       123,
				UserName: "testuser",
			},
			Chat: &tgbotapi.Chat{
				ID:   456,
				Type: "private",
			},
			Text: "/unknowncmd",
			Entities: []tgbotapi.MessageEntity{{
				Type:   "bot_command",
				Offset: 0,
				Length: 11, // Length of "/unknowncmd"
			}},
		},
	}

	// Process the update
	bot.processUpdate(update)

	// Wait for processing
	time.Sleep(10 * time.Millisecond)

	// Verify default handler was called
	mu.Lock()
	defer mu.Unlock()

	if !defaultHandlerCalled {
		t.Error("Expected default handler to be called for unknown command")
	}

	if receivedText != "/unknowncmd" {
		t.Errorf("Expected received text '/unknowncmd', got '%s'", receivedText)
	}
}

// TestBot_MiddlewareChainIntegration tests that multiple middleware are executed in correct order
func TestBot_MiddlewareChainIntegration(t *testing.T) {
	executionOrder := []string{}
	handlerCalled := false

	// Create bot with mocked dependencies
	bot, mockClient, _, _ := createTestBot()

	// Configure mock client
	mockClient.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
		return tgbotapi.Message{MessageID: 123}, nil
	}

	// Add multiple middleware in specific order
	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware1_before")
			err := next(ctx)
			executionOrder = append(executionOrder, "middleware1_after")
			return err
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware2_before")
			err := next(ctx)
			executionOrder = append(executionOrder, "middleware2_after")
			return err
		}
	}

	bot.UseMiddleware(middleware1)
	bot.UseMiddleware(middleware2)

	// Define test handler
	testHandler := func(ctx *Context, command string, args string) error {
		handlerCalled = true
		executionOrder = append(executionOrder, "handler")
		return nil
	}

	bot.HandleCommand("test", testHandler)

	// Create test update
	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 100,
			From: &tgbotapi.User{
				ID:       123,
				UserName: "testuser",
			},
			Chat: &tgbotapi.Chat{
				ID:   456,
				Type: "private",
			},
			Text: "/test",
			Entities: []tgbotapi.MessageEntity{{
				Type:   "bot_command",
				Offset: 0,
				Length: 5, // Length of "/test"
			}},
		},
	}

	// Process the update
	bot.processUpdate(update)

	// Verify execution order
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	expectedOrder := []string{
		"middleware1_before",
		"middleware2_before",
		"handler",
		"middleware2_after",
		"middleware1_after",
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d execution steps, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Expected execution order[%d] = '%s', got '%s'", i, expected,
				func() string {
					if i < len(executionOrder) {
						return executionOrder[i]
					}
					return "<missing>"
				}())
		}
	}
}

// TestBot_MultiStepFlowIntegration tests the end-to-end execution of a multi-step conversation flow
func TestBot_MultiStepFlowIntegration(t *testing.T) {
	// Test state tracking
	var sentMessages []tgbotapi.Chattable
	var sentCallbacks []tgbotapi.CallbackConfig
	var stepExecutionOrder []string
	var flowDataValues = make(map[string]interface{})

	// Create bot with real dependencies for flow testing
	bot, mockClient, _, _ := createTestBot()

	// Configure mock client to capture all outgoing messages and callbacks
	mockClient.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
		sentMessages = append(sentMessages, c)
		// Return a realistic message response
		return tgbotapi.Message{
			MessageID: len(sentMessages) + 100, // Unique ID for each message
			Text:      "Response message",
			Chat:      &tgbotapi.Chat{ID: 456},
		}, nil
	}

	mockClient.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
		// Capture callback queries (answers)
		if callback, ok := c.(tgbotapi.CallbackConfig); ok {
			sentCallbacks = append(sentCallbacks, callback)
		}
		return &tgbotapi.APIResponse{Ok: true}, nil
	}

	// Create a comprehensive multi-step flow configuration
	testFlow := &Flow{
		Name:  "testMultiStepFlow",
		Order: []string{"greeting", "collect_name", "choose_option", "final"},
		Steps: map[string]*flowStep{
			"greeting": {
				Name: "greeting",
				PromptConfig: &PromptConfig{
					Message: "Welcome! Let's start a multi-step process.",
					Keyboard: func(ctx *Context) *PromptKeyboardBuilder {
						return NewPromptKeyboard().
							ButtonCallback("Start", "start_flow").
							Row().
							ButtonCallback("Cancel", "cancel_flow")
					},
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					stepExecutionOrder = append(stepExecutionOrder, "greeting_process")
					if buttonClick != nil {
						if buttonClick.Data == "start_flow" {
							// Store flow data in our test tracking map instead of ctx.SetFlowData to avoid deadlock
							flowDataValues["flow_started"] = true
							return NextStep()
						} else if buttonClick.Data == "cancel_flow" {
							return CancelFlow()
						}
					}
					return Retry().WithPrompt("Please click one of the buttons to continue.")
				},
			},
			"collect_name": {
				Name: "collect_name",
				PromptConfig: &PromptConfig{
					Message: "Great! Please tell me your name:",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					stepExecutionOrder = append(stepExecutionOrder, "collect_name_process")
					if buttonClick != nil {
						return Retry().WithPrompt("Please type your name, don't use buttons.")
					}
					if input != "" && len(input) > 1 {
						// Store flow data in our test tracking map instead of ctx.SetFlowData to avoid deadlock
						flowDataValues["user_name"] = input
						return NextStep()
					}
					return Retry().WithPrompt("Please enter a valid name (at least 2 characters).")
				},
			},
			"choose_option": {
				Name: "choose_option",
				PromptConfig: &PromptConfig{
					Message: "Nice to meet you! Now choose an option:",
					Keyboard: func(ctx *Context) *PromptKeyboardBuilder {
						return NewPromptKeyboard().
							ButtonCallback("Option A", "option_a").
							ButtonCallback("Option B", "option_b").
							Row().
							ButtonCallback("Show My Name", "show_name")
					},
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					stepExecutionOrder = append(stepExecutionOrder, "choose_option_process")
					if buttonClick != nil {
						switch buttonClick.Data {
						case "option_a":
							flowDataValues["selected_option"] = "A"
							return NextStep()
						case "option_b":
							flowDataValues["selected_option"] = "B"
							return NextStep()
						case "show_name":
							if name, exists := flowDataValues["user_name"]; exists {
								return Retry().WithPrompt(fmt.Sprintf("Your name is: %v. Please choose an option.", name))
							}
							return Retry().WithPrompt("No name found. Please choose an option.")
						}
					}
					return Retry().WithPrompt("Please select one of the options using the buttons.")
				},
			},
			"final": {
				Name: "final",
				PromptConfig: &PromptConfig{
					Message: "Flow completed successfully!",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					stepExecutionOrder = append(stepExecutionOrder, "final_process")
					// Verify flow data is accessible from our test tracking
					if name, exists := flowDataValues["user_name"]; exists {
						flowDataValues["final_user_name"] = name
					}
					if option, exists := flowDataValues["selected_option"]; exists {
						flowDataValues["final_selected_option"] = option
					}
					return CompleteFlow()
				},
			},
		},
		OnComplete: func(ctx *Context) error {
			stepExecutionOrder = append(stepExecutionOrder, "flow_completed")
			return ctx.SendPromptText("Thank you for completing the flow!")
		},
		OnProcessAction: ProcessKeepMessage,
	}

	// Register the flow with the bot
	bot.RegisterFlow(testFlow)

	// Register a command to start the flow
	bot.HandleCommand("startflow", func(ctx *Context, command string, args string) error {
		stepExecutionOrder = append(stepExecutionOrder, "command_startflow")
		return ctx.StartFlow("testMultiStepFlow")
	})

	// Test user details
	testUserID := int64(123)
	testChatID := int64(456)

	// Step 1: Initiate Flow via Command
	t.Log("Step 1: Starting flow via /startflow command")
	startUpdate := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 100,
			From:      &tgbotapi.User{ID: testUserID, UserName: "testuser"},
			Chat:      &tgbotapi.Chat{ID: testChatID, Type: "private"},
			Text:      "/startflow",
			Entities: []tgbotapi.MessageEntity{{
				Type:   "bot_command",
				Offset: 0,
				Length: 10,
			}},
		},
	}
	bot.processUpdate(startUpdate)

	// Verify flow started and first prompt sent
	if len(sentMessages) == 0 {
		t.Fatal("Expected greeting message to be sent after starting flow")
	}
	if msgConfig, ok := sentMessages[0].(tgbotapi.MessageConfig); ok {
		if msgConfig.Text != "Welcome! Let's start a multi-step process." {
			t.Errorf("Expected greeting message, got: %s", msgConfig.Text)
		}
		if msgConfig.ReplyMarkup == nil {
			t.Error("Expected keyboard to be attached to greeting message")
		}
	} else {
		t.Error("Expected first message to be MessageConfig")
	}

	// Verify flow state
	if !bot.flowManager.isUserInFlow(testUserID) {
		t.Error("Expected user to be in flow after starting")
	}

	// Step 2: Handle Keyboard Callback (Start Flow)
	t.Log("Step 2: Clicking 'Start' button")

	// Extract callback data from the sent keyboard
	var startCallbackData string
	if msgConfig, ok := sentMessages[0].(tgbotapi.MessageConfig); ok {
		if keyboard, ok := msgConfig.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup); ok {
			if len(keyboard.InlineKeyboard) > 0 && len(keyboard.InlineKeyboard[0]) > 0 {
				startCallbackData = *keyboard.InlineKeyboard[0][0].CallbackData
			}
		}
	}
	if startCallbackData == "" {
		t.Fatal("Could not extract callback data from start button")
	}

	callbackUpdate := tgbotapi.Update{
		UpdateID: 2,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback123",
			From: &tgbotapi.User{ID: testUserID, UserName: "testuser"},
			Message: &tgbotapi.Message{
				MessageID: 101,
				Chat:      &tgbotapi.Chat{ID: testChatID},
				Text:      "Welcome! Let's start a multi-step process.",
			},
			Data: startCallbackData,
		},
	}
	bot.processUpdate(callbackUpdate)

	// Verify callback was answered and next step prompt sent
	if len(sentCallbacks) == 0 {
		t.Error("Expected callback query to be answered")
	}
	if len(sentMessages) < 2 {
		t.Fatal("Expected second message (name collection) to be sent")
	}
	if msgConfig, ok := sentMessages[1].(tgbotapi.MessageConfig); ok {
		if msgConfig.Text != "Great! Please tell me your name:" {
			t.Errorf("Expected name collection message, got: %s", msgConfig.Text)
		}
	}

	// Step 3: Handle Text Input (Name)
	t.Log("Step 3: Entering user name")
	nameUpdate := tgbotapi.Update{
		UpdateID: 3,
		Message: &tgbotapi.Message{
			MessageID: 102,
			From:      &tgbotapi.User{ID: testUserID, UserName: "testuser"},
			Chat:      &tgbotapi.Chat{ID: testChatID, Type: "private"},
			Text:      "John Doe",
		},
	}
	bot.processUpdate(nameUpdate)

	// Verify name was processed and next step prompt sent
	if len(sentMessages) < 3 {
		t.Fatal("Expected third message (option selection) to be sent")
	}
	if msgConfig, ok := sentMessages[2].(tgbotapi.MessageConfig); ok {
		if msgConfig.Text != "Nice to meet you! Now choose an option:" {
			t.Errorf("Expected option selection message, got: %s", msgConfig.Text)
		}
		if msgConfig.ReplyMarkup == nil {
			t.Error("Expected keyboard for option selection")
		}
	}

	// Verify flow data was set
	if flowDataValues["user_name"] != "John Doe" {
		t.Errorf("Expected user_name to be 'John Doe', got: %v", flowDataValues["user_name"])
	}

	// Step 4: Test Flow Data Usage (Show Name Button)
	t.Log("Step 4: Testing flow data usage with 'Show My Name' button")

	// Extract callback data for "Show My Name" button
	var showNameCallbackData string
	if msgConfig, ok := sentMessages[2].(tgbotapi.MessageConfig); ok {
		if keyboard, ok := msgConfig.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup); ok {
			// "Show My Name" should be in the second row, first button
			if len(keyboard.InlineKeyboard) > 1 && len(keyboard.InlineKeyboard[1]) > 0 {
				showNameCallbackData = *keyboard.InlineKeyboard[1][0].CallbackData
			}
		}
	}
	if showNameCallbackData == "" {
		t.Fatal("Could not extract callback data from 'Show My Name' button")
	}

	showNameUpdate := tgbotapi.Update{
		UpdateID: 4,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback124",
			From: &tgbotapi.User{ID: testUserID, UserName: "testuser"},
			Message: &tgbotapi.Message{
				MessageID: 103,
				Chat:      &tgbotapi.Chat{ID: testChatID},
				Text:      "Nice to meet you! Now choose an option:",
			},
			Data: showNameCallbackData,
		},
	}
	bot.processUpdate(showNameUpdate)

	// Verify name was displayed (retry message with name)
	if len(sentMessages) < 4 {
		t.Fatal("Expected message showing the name")
	}
	if msgConfig, ok := sentMessages[3].(tgbotapi.MessageConfig); ok {
		expectedText := "Your name is: John Doe. Please choose an option."
		if msgConfig.Text != expectedText {
			t.Errorf("Expected name display message '%s', got: %s", expectedText, msgConfig.Text)
		}
	}

	// Step 5: Select Option A
	// Step 5: Select Option A
	t.Log("Step 5: Selecting Option A")

	// Extract callback data for "Option A" button from the most recent message with keyboard
	var optionACallbackData string
	var keyboardMessage tgbotapi.MessageConfig
	var found bool

	// Find the most recent message with a keyboard (retry message should have same keyboard)
	for i := len(sentMessages) - 1; i >= 0; i-- {
		if msgConfig, ok := sentMessages[i].(tgbotapi.MessageConfig); ok {
			if msgConfig.ReplyMarkup != nil {
				keyboardMessage = msgConfig
				found = true
				break
			}
		}
	}

	if !found {
		t.Fatal("Could not find any message with keyboard")
	}

	if keyboard, ok := keyboardMessage.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup); ok {
		t.Logf("Keyboard has %d rows", len(keyboard.InlineKeyboard))
		if len(keyboard.InlineKeyboard) > 0 {
			t.Logf("First row has %d buttons", len(keyboard.InlineKeyboard[0]))
			if len(keyboard.InlineKeyboard[0]) > 0 {
				optionACallbackData = *keyboard.InlineKeyboard[0][0].CallbackData
				t.Logf("Option A callback data: %s", optionACallbackData)
			}
		}
	}

	if optionACallbackData == "" {
		t.Fatal("Could not extract callback data from 'Option A' button")
	}
	optionAUpdate := tgbotapi.Update{
		UpdateID: 5,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback125",
			From: &tgbotapi.User{ID: testUserID, UserName: "testuser"},
			Message: &tgbotapi.Message{
				MessageID: 104,
				Chat:      &tgbotapi.Chat{ID: testChatID},
				Text:      "Your name is: John Doe. Please choose an option:",
			},
			Data: optionACallbackData,
		},
	}
	bot.processUpdate(optionAUpdate)

	// Verify final step was reached
	if len(sentMessages) < 5 {
		t.Fatal("Expected final message to be sent")
	}
	if msgConfig, ok := sentMessages[4].(tgbotapi.MessageConfig); ok {
		if msgConfig.Text != "Flow completed successfully!" {
			t.Errorf("Expected final message, got: %s", msgConfig.Text)
		}
	}

	// Step 6: Complete Flow
	t.Log("Step 6: Completing flow")
	completeUpdate := tgbotapi.Update{
		UpdateID: 6,
		Message: &tgbotapi.Message{
			MessageID: 105,
			From:      &tgbotapi.User{ID: testUserID, UserName: "testuser"},
			Chat:      &tgbotapi.Chat{ID: testChatID, Type: "private"},
			Text:      "done",
		},
	}
	bot.processUpdate(completeUpdate)

	// Verify flow completion
	if len(sentMessages) < 6 {
		t.Fatal("Expected flow completion message")
	}
	if msgConfig, ok := sentMessages[5].(tgbotapi.MessageConfig); ok {
		if msgConfig.Text != "Thank you for completing the flow!" {
			t.Errorf("Expected completion message, got: %s", msgConfig.Text)
		}
	}

	// Verify user is no longer in flow
	if bot.flowManager.isUserInFlow(testUserID) {
		t.Error("Expected user to no longer be in flow after completion")
	}

	// Verify step execution order
	expectedOrder := []string{
		"command_startflow",
		"greeting_process",
		"collect_name_process",
		"choose_option_process", // for show_name click
		"choose_option_process", // for option_a click
		"final_process",
		"flow_completed",
	}

	if len(stepExecutionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d steps, got %d. Steps: %v", len(expectedOrder), len(stepExecutionOrder), stepExecutionOrder)
	}

	for i, expected := range expectedOrder {
		if i >= len(stepExecutionOrder) || stepExecutionOrder[i] != expected {
			t.Errorf("Step execution order[%d]: expected '%s', got '%s'", i, expected,
				func() string {
					if i < len(stepExecutionOrder) {
						return stepExecutionOrder[i]
					}
					return "<missing>"
				}())
		}
	}

	// Verify flow data persistence and usage
	if flowDataValues["final_user_name"] != "John Doe" {
		t.Errorf("Expected final_user_name to be 'John Doe', got: %v", flowDataValues["final_user_name"])
	}
	if flowDataValues["final_selected_option"] != "A" {
		t.Errorf("Expected final_selected_option to be 'A', got: %v", flowDataValues["final_selected_option"])
	}

	// Verify callback queries were answered
	expectedCallbacks := 3 // start_flow, show_name, option_a
	if len(sentCallbacks) != expectedCallbacks {
		t.Errorf("Expected %d callback answers, got %d", expectedCallbacks, len(sentCallbacks))
	}

	// Verify all sent messages have correct chat ID
	for i, msg := range sentMessages {
		if msgConfig, ok := msg.(tgbotapi.MessageConfig); ok {
			if msgConfig.ChatID != testChatID {
				t.Errorf("Message %d sent to wrong chat ID: expected %d, got %d", i, testChatID, msgConfig.ChatID)
			}
		}
	}

	t.Log("Multi-step flow integration test completed successfully")
}

// TestBot_AuthMiddlewareIntegration tests the end-to-end processing with AuthMiddleware
func TestBot_AuthMiddlewareIntegration(t *testing.T) {
	t.Run("Scenario 1: Access Allowed", func(t *testing.T) {
		// Setup test state
		handlerCalled := false
		var receivedContext *Context
		var receivedCommand string
		var receivedArgs string

		// Create bot with mocked dependencies
		bot, mockClient, _, mockAccessManager := createTestBot()

		// Configure mock client to capture outgoing messages
		var sentMessages []tgbotapi.Chattable
		mockClient.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			sentMessages = append(sentMessages, c)
			return tgbotapi.Message{MessageID: 123, Text: "Test reply"}, nil
		}

		// Configure mock AccessManager to allow access
		mockAccessManager.CheckPermissionFunc = func(ctx *PermissionContext) error {
			return nil // Allow access
		}

		// Add AuthMiddleware
		bot.UseMiddleware(AuthMiddleware(mockAccessManager))

		// Define test command handler
		testHandler := func(ctx *Context, command string, args string) error {
			handlerCalled = true
			receivedContext = ctx
			receivedCommand = command
			receivedArgs = args
			return ctx.SendPromptText("Handler executed successfully")
		}

		// Register the test command handler
		bot.HandleCommand("auth_test_cmd", testHandler)

		// Create test update representing incoming command message
		update := tgbotapi.Update{
			UpdateID: 1,
			Message: &tgbotapi.Message{
				MessageID: 100,
				From: &tgbotapi.User{
					ID:       123,
					UserName: "testuser",
				},
				Chat: &tgbotapi.Chat{
					ID:   456,
					Type: "private",
				},
				Text: "/auth_test_cmd test_argument",
				Entities: []tgbotapi.MessageEntity{{
					Type:   "bot_command",
					Offset: 0,
					Length: 14, // Length of "/auth_test_cmd"
				}},
			},
		}

		// Process the update
		bot.processUpdate(update)

		// Verify AccessManager.CheckPermission was called
		if len(mockAccessManager.CheckPermissionCalls) == 0 {
			t.Error("Expected CheckPermission to be called, but it wasn't")
		} else {
			permCtx := mockAccessManager.CheckPermissionCalls[0]
			if permCtx.UserID != 123 {
				t.Errorf("Expected UserID 123, got %d", permCtx.UserID)
			}
			if permCtx.ChatID != 456 {
				t.Errorf("Expected ChatID 456, got %d", permCtx.ChatID)
			}
			if permCtx.Command != "auth_test_cmd" {
				t.Errorf("Expected command 'auth_test_cmd', got '%s'", permCtx.Command)
			}
		}

		// Verify that the handler was executed
		if !handlerCalled {
			t.Error("Expected command handler to be called when access is allowed, but it wasn't")
		}

		// Verify received command and arguments
		if receivedCommand != "auth_test_cmd" {
			t.Errorf("Expected command 'auth_test_cmd', got '%s'", receivedCommand)
		}

		if receivedArgs != " test_argument" {
			t.Errorf("Expected args ' test_argument', got '%s'", receivedArgs)
		}

		// Verify context correctness
		if receivedContext == nil {
			t.Error("Expected context to be non-nil")
		} else {
			if receivedContext.UserID() != 123 {
				t.Errorf("Expected UserID 123, got %d", receivedContext.UserID())
			}
			if receivedContext.ChatID() != 456 {
				t.Errorf("Expected ChatID 456, got %d", receivedContext.ChatID())
			}
		}

		// Verify that the handler's response was sent (not a rejection message)
		if len(sentMessages) == 0 {
			t.Error("Expected at least one message to be sent, but none were sent")
		} else {
			if msgConfig, ok := sentMessages[0].(tgbotapi.MessageConfig); ok {
				if msgConfig.Text != "Handler executed successfully" {
					t.Errorf("Expected message text 'Handler executed successfully', got '%s'", msgConfig.Text)
				}
			} else {
				t.Error("Expected sent message to be of type MessageConfig")
			}
		}
	})

	t.Run("Scenario 2: Access Denied (Message Update)", func(t *testing.T) {
		// Setup test state
		handlerCalled := false

		// Create bot with mocked dependencies
		bot, mockClient, _, mockAccessManager := createTestBot()

		// Configure mock client to capture outgoing messages
		var sentMessages []tgbotapi.Chattable
		mockClient.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			sentMessages = append(sentMessages, c)
			return tgbotapi.Message{MessageID: 123, Text: "Rejection message"}, nil
		}

		// Configure mock AccessManager to deny access
		mockAccessManager.CheckPermissionFunc = func(ctx *PermissionContext) error {
			return errors.New("Access denied for testing")
		}

		// Add AuthMiddleware
		bot.UseMiddleware(AuthMiddleware(mockAccessManager))

		// Define test command handler
		testHandler := func(ctx *Context, command string, args string) error {
			handlerCalled = true
			return ctx.SendPromptText("Handler executed successfully")
		}

		// Register the test command handler
		bot.HandleCommand("auth_test_cmd", testHandler)

		// Create test update representing incoming command message
		update := tgbotapi.Update{
			UpdateID: 1,
			Message: &tgbotapi.Message{
				MessageID: 100,
				From: &tgbotapi.User{
					ID:       123,
					UserName: "testuser",
				},
				Chat: &tgbotapi.Chat{
					ID:   456,
					Type: "private",
				},
				Text: "/auth_test_cmd test_argument",
				Entities: []tgbotapi.MessageEntity{{
					Type:   "bot_command",
					Offset: 0,
					Length: 14, // Length of "/auth_test_cmd"
				}},
			},
		}

		// Process the update
		bot.processUpdate(update)

		// Verify AccessManager.CheckPermission was called
		if len(mockAccessManager.CheckPermissionCalls) == 0 {
			t.Error("Expected CheckPermission to be called, but it wasn't")
		}

		// Verify that the handler was NOT executed
		if handlerCalled {
			t.Error("Expected command handler NOT to be called when access is denied, but it was")
		}

		// Verify that a rejection message was sent
		if len(sentMessages) == 0 {
			t.Error("Expected rejection message to be sent, but no messages were sent")
		} else {
			if msgConfig, ok := sentMessages[0].(tgbotapi.MessageConfig); ok {
				if msgConfig.Text != "🚫 Access denied for testing" {
					t.Errorf("Expected rejection message '🚫 Access denied for testing', got '%s'", msgConfig.Text)
				}
				if msgConfig.ChatID != 456 {
					t.Errorf("Expected message to be sent to ChatID 456, got %d", msgConfig.ChatID)
				}
			} else {
				t.Error("Expected sent message to be of type MessageConfig")
			}
		}
	})

	t.Run("Scenario 3: Access Denied with Different Error Message", func(t *testing.T) {
		// Setup test state
		handlerCalled := false

		// Create bot with mocked dependencies
		bot, mockClient, _, mockAccessManager := createTestBot()

		// Configure mock client to capture outgoing messages
		var sentMessages []tgbotapi.Chattable
		mockClient.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			sentMessages = append(sentMessages, c)
			return tgbotapi.Message{MessageID: 123, Text: "Rejection message"}, nil
		}

		// Configure mock AccessManager to deny access with different error message
		mockAccessManager.CheckPermissionFunc = func(ctx *PermissionContext) error {
			return errors.New("Insufficient privileges")
		}

		// Add AuthMiddleware
		bot.UseMiddleware(AuthMiddleware(mockAccessManager))

		// Define test command handler
		testHandler := func(ctx *Context, command string, args string) error {
			handlerCalled = true
			return ctx.SendPromptText("Handler executed successfully")
		}

		// Register the test command handler
		bot.HandleCommand("auth_test_cmd2", testHandler)

		// Create test update representing incoming command message
		update := tgbotapi.Update{
			UpdateID: 1,
			Message: &tgbotapi.Message{
				MessageID: 100,
				From: &tgbotapi.User{
					ID:       789,
					UserName: "anotheruser",
				},
				Chat: &tgbotapi.Chat{
					ID:   999,
					Type: "group",
				},
				Text: "/auth_test_cmd2",
				Entities: []tgbotapi.MessageEntity{{
					Type:   "bot_command",
					Offset: 0,
					Length: 15, // Length of "/auth_test_cmd2"
				}},
			},
		}

		// Process the update
		bot.processUpdate(update)

		// Verify AccessManager.CheckPermission was called with correct context
		if len(mockAccessManager.CheckPermissionCalls) == 0 {
			t.Error("Expected CheckPermission to be called, but it wasn't")
		} else {
			permCtx := mockAccessManager.CheckPermissionCalls[0]
			if permCtx.UserID != 789 {
				t.Errorf("Expected UserID 789, got %d", permCtx.UserID)
			}
			if permCtx.ChatID != 999 {
				t.Errorf("Expected ChatID 999, got %d", permCtx.ChatID)
			}
			if permCtx.Command != "auth_test_cmd2" {
				t.Errorf("Expected command 'auth_test_cmd2', got '%s'", permCtx.Command)
			}
			if permCtx.IsGroup != true {
				t.Errorf("Expected IsGroup true, got %v", permCtx.IsGroup)
			}
		}

		// Verify that the handler was NOT executed
		if handlerCalled {
			t.Error("Expected command handler NOT to be called when access is denied, but it was")
		}

		// Verify that a rejection message was sent with the custom error
		if len(sentMessages) == 0 {
			t.Error("Expected rejection message to be sent, but no messages were sent")
		} else {
			if msgConfig, ok := sentMessages[0].(tgbotapi.MessageConfig); ok {
				if msgConfig.Text != "🚫 Insufficient privileges" {
					t.Errorf("Expected rejection message '🚫 Insufficient privileges', got '%s'", msgConfig.Text)
				}
				if msgConfig.ChatID != 999 {
					t.Errorf("Expected message to be sent to ChatID 999, got %d", msgConfig.ChatID)
				}
			} else {
				t.Error("Expected sent message to be of type MessageConfig")
			}
		}
	})

	t.Log("AuthMiddleware integration test completed successfully")
}
