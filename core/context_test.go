package teleflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// fakeBot constructs a *Bot with a fake Telegram API server that records
// every request and returns the given responder. The server is cleaned up
// automatically when the test finishes.
func fakeBot(t *testing.T, responder func(method string, params url.Values) (any, bool)) *Bot {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the URL path to extract the method: /bot<token>/<method>
		parts := strings.SplitN(r.URL.Path, "/", 3)
		if len(parts) < 3 {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		method := parts[2]

		// Parse the form body
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		// getMe is called during NewBotAPIWithClient construction; always
		// answer it with a valid User so construction succeeds, then hand
		// all other methods to the responder.
		if method == "getMe" {
			json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"result": map[string]any{
					"id":         1,
					"is_bot":     true,
					"first_name": "Test",
					"username":   "testbot",
				},
			})
			return
		}

		// Call the responder
		result, ok := responder(method, r.Form)
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]any{
				"ok":          false,
				"error_code":  403,
				"description": "forbidden",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"result": result,
		})
	}))
	t.Cleanup(srv.Close)

	api, err := tgbotapi.NewBotAPIWithClient("test-token", srv.URL+"/bot%s/%s", srv.Client())
	if err != nil {
		t.Fatalf("NewBotAPIWithClient: %v", err)
	}

	bot := &Bot{
		api:              api,
		self:             tgbotapi.User{ID: 1, UserName: "testbot", FirstName: "Test"},
		handler:          NewRouter(),
		callbackRegistry: make(map[string]registryEntry[func(ctx *Context) error]),
		dataRegistry:     make(map[string]registryEntry[interface{}]),
	}
	return bot
}

// TestContextUserID extracts UserID from a message update.
func TestContextUserID(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{From: &tgbotapi.User{ID: 42}},
	}, nil, nil)
	if id := ctx.UserID(); id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

// TestContextUserIDFromCallback extracts UserID from a callback query.
func TestContextUserIDFromCallback(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: 99}},
	}, nil, nil)
	if id := ctx.UserID(); id != 99 {
		t.Fatalf("expected 99, got %d", id)
	}
}

// TestContextUserIDFromNil returns 0 when no user is present.
func TestContextUserIDFromNil(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{}, nil, nil)
	if id := ctx.UserID(); id != 0 {
		t.Fatalf("expected 0, got %d", id)
	}
}

// TestContextChatID extracts ChatID from a message.
func TestContextChatID(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 77}},
	}, nil, nil)
	if id := ctx.ChatID(); id != 77 {
		t.Fatalf("expected 77, got %d", id)
	}
}

// TestContextChatIDFromCallback extracts ChatID from a callback query.
func TestContextChatIDFromCallback(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 88}},
		},
	}, nil, nil)
	if id := ctx.ChatID(); id != 88 {
		t.Fatalf("expected 88, got %d", id)
	}
}

// TestContextTextFromMessage returns the message text.
func TestContextTextFromMessage(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "hello"},
	}, nil, nil)
	if text := ctx.Text(); text != "hello" {
		t.Fatalf("expected 'hello', got %q", text)
	}
}

// TestContextTextFromCallback returns the callback data.
func TestContextTextFromCallback(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: "cb_data"},
	}, nil, nil)
	if text := ctx.Text(); text != "cb_data" {
		t.Fatalf("expected 'cb_data', got %q", text)
	}
}

// TestContextTextFromNil returns empty.
func TestContextTextFromNil(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{}, nil, nil)
	if text := ctx.Text(); text != "" {
		t.Fatalf("expected empty, got %q", text)
	}
}

// TestContextUpdate returns the raw update.
func TestContextUpdate(t *testing.T) {
	upd := tgbotapi.Update{UpdateID: 123}
	ctx := NewContext(upd, nil, nil)
	if ctx.Update().UpdateID != 123 {
		t.Fatalf("expected 123, got %d", ctx.Update().UpdateID)
	}
}

// TestContextSetGet stores and retrieves request-scoped data.
func TestContextSetGet(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{}, nil, nil)
	ctx.Set("key", "value")
	val, ok := ctx.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val.(string) != "value" {
		t.Fatalf("expected 'value', got %v", val)
	}
}

// TestContextGetMissing returns false for missing keys.
func TestContextGetMissing(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{}, nil, nil)
	_, ok := ctx.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing key")
	}
}

// TestContextReplyText sends a text message and verifies the API call.
func TestContextReplyText(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "sendMessage" {
			t.Fatalf("expected sendMessage, got %s", method)
		}
		if params.Get("chat_id") != "1" {
			t.Fatalf("expected chat_id=1, got %s", params.Get("chat_id"))
		}
		if params.Get("text") != "Hello" {
			t.Fatalf("expected text=Hello, got %s", params.Get("text"))
		}
		// Return a minimal message
		return map[string]any{"message_id": 10, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 1},
			From: &tgbotapi.User{ID: 1},
		},
	}, bot.api, bot)

	if err := ctx.ReplyText("Hello"); err != nil {
		t.Fatalf("ReplyText: %v", err)
	}
}

// TestContextReplyTextWithParseMode includes parse_mode in the request.
func TestContextReplyTextWithParseMode(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "sendMessage" {
			t.Fatalf("expected sendMessage, got %s", method)
		}
		if params.Get("parse_mode") != "HTML" {
			t.Fatalf("expected parse_mode=HTML, got %s", params.Get("parse_mode"))
		}
		return map[string]any{"message_id": 1, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	if err := ctx.ReplyTextWithParseMode("<b>bold</b>", ParseModeHTML); err != nil {
		t.Fatalf("ReplyTextWithParseMode: %v", err)
	}
}

// TestContextReplyWithKeyboard includes reply_markup in the request.
func TestContextReplyWithKeyboard(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "sendMessage" {
			t.Fatalf("expected sendMessage, got %s", method)
		}
		if params.Get("reply_markup") == "" {
			t.Fatal("expected reply_markup to be present")
		}
		return map[string]any{"message_id": 1, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	kb := NewInlineKeyboard().AddRow(ButtonCallback("OK", "ok")).Build()
	if err := ctx.ReplyWithKeyboard("Choose", kb); err != nil {
		t.Fatalf("ReplyWithKeyboard: %v", err)
	}
}

// TestContextEditMessageText calls editMessageText.
func TestContextEditMessageText(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "editMessageText" {
			t.Fatalf("expected editMessageText, got %s", method)
		}
		if params.Get("message_id") != "5" {
			t.Fatalf("expected message_id=5, got %s", params.Get("message_id"))
		}
		return map[string]any{"message_id": 5, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	if err := ctx.EditMessageText(5, "new text"); err != nil {
		t.Fatalf("EditMessageText: %v", err)
	}
}

// TestContextDeleteMessage calls deleteMessage.
func TestContextDeleteMessage(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "deleteMessage" {
			t.Fatalf("expected deleteMessage, got %s", method)
		}
		return true, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	if err := ctx.DeleteMessage(3); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
}

// TestContextAnswerCallback calls answerCallbackQuery.
func TestContextAnswerCallback(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "answerCallbackQuery" {
			t.Fatalf("expected answerCallbackQuery, got %s", method)
		}
		if params.Get("callback_query_id") != "cb_123" {
			t.Fatalf("expected callback_query_id=cb_123, got %s", params.Get("callback_query_id"))
		}
		return true, true
	})

	ctx := NewContext(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{ID: "cb_123", From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	if err := ctx.AnswerCallback(); err != nil {
		t.Fatalf("AnswerCallback: %v", err)
	}
}

// TestContextAnswerCallbackNotCallback returns an error when the update is not a callback.
func TestContextAnswerCallbackNotCallback(t *testing.T) {
	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, nil, nil)

	if err := ctx.AnswerCallback(); err == nil {
		t.Fatal("expected error for non-callback update")
	}
}

// TestContextRegisterDataAndGetData stores and retrieves data through the bot.
func TestContextRegisterDataAndGetData(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})
	bot.handler = NewRouter()

	// Bot has a handler — we need to connect it properly
	// RegisterData through Context
	ctx1 := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	dataID := ctx1.RegisterData(map[string]string{"action": "approve", "amount": "100"})
	if dataID == "" {
		t.Fatal("expected non-empty data ID")
	}

	// GetData through a callback context
	ctx2 := NewContext(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Data: dataID,
			From: &tgbotapi.User{ID: 1},
		},
	}, bot.api, bot)

	val, ok := ctx2.GetData()
	if !ok {
		t.Fatal("expected data to be found")
	}
	m := val.(map[string]string)
	if m["action"] != "approve" {
		t.Fatalf("expected action=approve, got %v", m["action"])
	}
}

// TestContextGetDataConsumedOnce verifies data is deleted after retrieval.
func TestContextGetDataConsumedOnce(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	ctx1 := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	dataID := ctx1.RegisterData("secret")

	// First retrieval succeeds
	ctx2 := NewContext(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: dataID, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)
	val, ok := ctx2.GetData()
	if !ok || val.(string) != "secret" {
		t.Fatal("first GetData should succeed")
	}

	// Second retrieval fails
	ctx3 := NewContext(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: dataID, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)
	_, ok = ctx3.GetData()
	if ok {
		t.Fatal("second GetData should fail (data was consumed)")
	}
}

// TestContextRegisterCallback verifies callback handler registration.
func TestContextRegisterCallback(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	var called bool
	cbID := ctx.RegisterCallback(func(ctx *Context) error {
		called = true
		return nil
	})
	if cbID == "" {
		t.Fatal("expected non-empty callback ID")
	}

	// Fire the callback via processUpdate
	cbCtx := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: cbID, From: &tgbotapi.User{ID: 1}},
	}
	bot.processUpdate(cbCtx)
	if !called {
		t.Fatal("callback handler should have been called")
	}

	// Second call should not fire (one-time)
	called = false
	bot.processUpdate(cbCtx)
	if called {
		t.Fatal("callback handler should NOT be called again (one-time)")
	}
}

// TestContextEditMessageReplyMarkup includes value-type (no nil panic).
func TestContextEditMessageReplyMarkup(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "editMessageReplyMarkup" {
			t.Fatalf("expected editMessageReplyMarkup, got %s", method)
		}
		return map[string]any{"message_id": 5, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	kb := NewInlineKeyboard().AddRow(ButtonCallback("OK", "ok")).Build()
	if err := ctx.EditMessageReplyMarkup(5, kb); err != nil {
		t.Fatalf("EditMessageReplyMarkup: %v", err)
	}
}

// TestContextAPI returns the underlying API.
func TestContextAPI(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	if ctx.API() != bot.api {
		t.Fatal("API() returned different bot")
	}
}

// TestContextEditMessageTextWithParseMode tests edit with parse mode.
func TestContextEditMessageTextWithParseMode(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "editMessageText" {
			t.Fatalf("expected editMessageText, got %s", method)
		}
		if params.Get("parse_mode") != "Markdown" {
			t.Fatalf("expected parse_mode=Markdown, got %s", params.Get("parse_mode"))
		}
		return map[string]any{"message_id": 5, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	if err := ctx.EditMessageTextWithParseMode(5, "*bold*", ParseModeMarkdown); err != nil {
		t.Fatalf("EditMessageTextWithParseMode: %v", err)
	}
}

func TestContextReplyWithKeyboardAndParseMode(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "sendMessage" {
			t.Fatalf("expected sendMessage, got %s", method)
		}
		if params.Get("parse_mode") != "MarkdownV2" {
			t.Fatalf("expected parse_mode=MarkdownV2, got %s", params.Get("parse_mode"))
		}
		if params.Get("reply_markup") == "" {
			t.Fatal("expected reply_markup to be present")
		}
		return map[string]any{"message_id": 1, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	kb := NewInlineKeyboard().AddRow(ButtonCallback("Yes", "yes")).Build()
	if err := ctx.ReplyWithKeyboardAndParseMode("*text*", kb, ParseModeMarkdownV2); err != nil {
		t.Fatalf("ReplyWithKeyboardAndParseMode: %v", err)
	}
}

// TestContextEditMessageMedia tests editing media.
func TestContextEditMessageMedia(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "editMessageMedia" {
			t.Fatalf("expected editMessageMedia, got %s", method)
		}
		return map[string]any{"message_id": 5, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL("https://example.com/photo.jpg"))
	if err := ctx.EditMessageMedia(5, media); err != nil {
		t.Fatalf("EditMessageMedia: %v", err)
	}
}

// TestContextRegisterCallbackOneTime verifies the callback is removed after execution.
func TestContextRegisterCallbackOneTime(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	var count int
	cbID := ctx.RegisterCallback(func(ctx *Context) error {
		count++
		return nil
	})

	bot.processUpdate(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: cbID, From: &tgbotapi.User{ID: 1}},
	})
	bot.processUpdate(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: cbID, From: &tgbotapi.User{ID: 1}},
	})
	if count != 1 {
		t.Fatalf("expected callback to fire exactly once, got %d", count)
	}
}

// TestContextRegisterDataOneTime verifies data is consumed once.
func TestContextRegisterDataOneTime(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	// Register some data
	bot.RegisterData("test-value")

	// GetData through the Bot directly
	val, ok := bot.GetData("nonexistent")
	if ok {
		t.Fatal("expected nonexistent key to not be found")
	}
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}

	_ = ctx // ctx used for setup
}

// TestBotRegisterDataAndGetData via Bot directly.
func TestBotRegisterDataAndGetData(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	id := bot.RegisterData(42)
	val, ok := bot.GetData(id)
	if !ok {
		t.Fatal("expected data to be found")
	}
	if val.(int) != 42 {
		t.Fatalf("expected 42, got %d", val.(int))
	}

	// Second get fails
	_, ok = bot.GetData(id)
	if ok {
		t.Fatal("expected data to be consumed")
	}
}

// TestBotSetBotCommandsOrdering verifies deterministic ordering.
func TestBotSetBotCommandsOrdering(t *testing.T) {
	var capturedCommands []string
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method == "setMyCommands" {
			// Parse the commands JSON from the body
			commands := params.Get("commands")
			if commands != "" {
				var cmds []map[string]string
				if err := json.Unmarshal([]byte(commands), &cmds); err == nil {
					for _, c := range cmds {
						capturedCommands = append(capturedCommands, c["command"])
					}
				}
			}
		}
		return true, true
	})

	err := bot.SetBotCommands(map[string]string{
		"zzz": "last command",
		"aaa": "first command",
		"mmm": "middle command",
	})
	if err != nil {
		t.Fatalf("SetBotCommands: %v", err)
	}

	expected := []string{"aaa", "mmm", "zzz"}
	if len(capturedCommands) != 3 {
		t.Fatalf("expected 3 commands, got %v", capturedCommands)
	}
	for i, c := range capturedCommands {
		if c != expected[i] {
			t.Fatalf("position %d: expected %s, got %s", i, expected[i], c)
		}
	}
}

// TestBotProcessUpdateCallbackDispatch dispatches a registered callback handler.
func TestBotProcessUpdateCallbackDispatch(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	var called bool
	cbID := bot.registerCallback(func(ctx *Context) error {
		called = true
		return nil
	})

	bot.processUpdate(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: cbID, From: &tgbotapi.User{ID: 1}},
	})
	if !called {
		t.Fatal("callback handler should have been called")
	}
}

// TestBotProcessUpdateUnmatchedCallback falls through to the router.
func TestBotProcessUpdateUnmatchedCallback(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})
	router := NewRouter()
	var defaultCalled bool
	router.Default(Func(func(ctx *Context) error {
		defaultCalled = true
		return nil
	}))
	bot.handler = router

	bot.processUpdate(tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{Data: "nonexistent", From: &tgbotapi.User{ID: 1}},
	})
	if !defaultCalled {
		t.Fatal("unmatched callback should fall through to default handler")
	}
}

// TestBotStoreRegistry verifies the shared store helper.
func TestBotStoreRegistry(t *testing.T) {
	reg := make(map[string]registryEntry[string])
	id := storeRegistry(reg, "hello")
	if len(reg) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(reg))
	}
	entry, ok := reg[id]
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if entry.value != "hello" {
		t.Fatalf("expected 'hello', got %q", entry.value)
	}
	if entry.expiresAt.IsZero() {
		t.Fatal("expected expiry to be set")
	}
}

// TestBotGetDataExpired verifies that expired entries return ok=false.
func TestBotGetDataExpired(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	// Insert an expired entry directly
	bot.mu.Lock()
	bot.dataRegistry["expired"] = registryEntry[interface{}]{
		value:     "stale",
		expiresAt: time.Now().Add(-time.Hour),
	}
	bot.mu.Unlock()

	val, ok := bot.GetData("expired")
	if ok {
		t.Fatal("expected expired entry to not be found")
	}
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}
}

// TestBotNewRouter verifies the router is created correctly.
func TestBotNewRouter(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	if bot.handler == nil {
		t.Fatal("expected handler to be set")
	}
}

// TestNewBotWithAPINilHandler returns an error for a nil handler.
func TestNewBotWithAPINilHandler(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	if _, err := newBotWithAPI(bot.api, nil); err == nil {
		t.Fatal("expected error for nil handler")
	}
}

// TestNewBotWithAPISuccess builds a Bot from a working API client.
func TestNewBotWithAPISuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"result": map[string]any{
				"id":         1,
				"is_bot":     true,
				"first_name": "Test",
				"username":   "testbot",
			},
		})
	}))
	t.Cleanup(srv.Close)

	api, err := tgbotapi.NewBotAPIWithClient("test-token", srv.URL+"/bot%s/%s", srv.Client())
	if err != nil {
		t.Fatalf("NewBotAPIWithClient: %v", err)
	}

	handler := NewRouter()
	bot, err := newBotWithAPI(api, handler)
	if err != nil {
		t.Fatalf("newBotWithAPI: %v", err)
	}
	if bot.handler != handler {
		t.Fatal("handler not wired correctly")
	}
	if bot.self.UserName != "testbot" {
		t.Fatalf("expected self username testbot, got %q", bot.self.UserName)
	}
}

// TestNewBotWithAPIGetMeError returns an error when GetMe fails.
func TestNewBotWithAPIGetMeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate an API failure
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":          false,
			"error_code":  401,
			"description": "Unauthorized",
		})
	}))
	t.Cleanup(srv.Close)

	api, err := tgbotapi.NewBotAPIWithClient("test-token", srv.URL+"/bot%s/%s", srv.Client())
	if err != nil {
		// NewBotAPIWithClient itself calls GetMe, so it should fail too.
		if err == nil {
			t.Fatal("expected GetMe error")
		}
		return
	}

	if _, err := newBotWithAPI(api, NewRouter()); err == nil {
		t.Fatal("expected GetMe error")
	}
}

// TestProcessUpdates dispatches updates until the channel closes.
func TestProcessUpdates(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	var called bool
	bot.handler = Func(func(ctx *Context) error {
		called = true
		return nil
	})

	updates := make(chan tgbotapi.Update)
	go func() {
		updates <- tgbotapi.Update{
			Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
		}
		close(updates)
	}()

	bot.processUpdates(updates)

	// processUpdates spawns goroutines; give the handler a moment to run.
	deadline := time.Now().Add(2 * time.Second)
	for !called && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !called {
		t.Fatal("handler should have been called")
	}
}

// TestNewBotNilHandler checks that NewBot rejects nil without network.
func TestNewBotNilHandler(t *testing.T) {
	_, err := NewBot("test-token", nil)
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
}

// TestSweepRegistriesOnce removes expired entries from both registries.
func TestSweepRegistriesOnce(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		return true, true
	})

	now := time.Now()
	bot.mu.Lock()
	bot.callbackRegistry["expired-cb"] = registryEntry[func(ctx *Context) error]{
		value:     func(ctx *Context) error { return nil },
		expiresAt: now.Add(-time.Minute),
	}
	bot.callbackRegistry["fresh-cb"] = registryEntry[func(ctx *Context) error]{
		value:     func(ctx *Context) error { return nil },
		expiresAt: now.Add(time.Hour),
	}
	bot.dataRegistry["expired-data"] = registryEntry[interface{}]{
		value:     "stale",
		expiresAt: now.Add(-time.Minute),
	}
	bot.dataRegistry["fresh-data"] = registryEntry[interface{}]{
		value:     "fresh",
		expiresAt: now.Add(time.Hour),
	}
	bot.mu.Unlock()

	bot.sweepRegistriesOnce(now)

	bot.mu.Lock()
	defer bot.mu.Unlock()
	if _, ok := bot.callbackRegistry["expired-cb"]; ok {
		t.Error("expired callback entry should have been removed")
	}
	if _, ok := bot.callbackRegistry["fresh-cb"]; !ok {
		t.Error("fresh callback entry should remain")
	}
	if _, ok := bot.dataRegistry["expired-data"]; ok {
		t.Error("expired data entry should have been removed")
	}
	if _, ok := bot.dataRegistry["fresh-data"]; !ok {
		t.Error("fresh data entry should remain")
	}
}

// TestFlowActionHelpers verifies the helper constructors.
func TestFlowActionHelpers(t *testing.T) {
	if Next().Type != actionNext {
		t.Fatal("Next() should have type actionNext")
	}
	if Retry().Type != actionRetry {
		t.Fatal("Retry() should have type actionRetry")
	}
	if End().Type != actionEnd {
		t.Fatal("End() should have type actionEnd")
	}
	if g := GoTo("target"); g.Type != actionGoTo || g.TargetStep != "target" {
		t.Fatal("GoTo() should have type actionGoTo and correct target")
	}
}

// TestContextReplyWithKeyboardParseMode verifies parse mode with keyboard.
func TestContextReplyWithKeyboardParseMode(t *testing.T) {
	bot := fakeBot(t, func(method string, params url.Values) (any, bool) {
		if method != "sendMessage" {
			t.Fatalf("expected sendMessage, got %s", method)
		}
		if params.Get("parse_mode") != "Markdown" {
			t.Fatalf("expected parse_mode=Markdown, got %s", params.Get("parse_mode"))
		}
		return map[string]any{"message_id": 1, "chat": map[string]any{"id": 1, "type": "private"}}, true
	})

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, From: &tgbotapi.User{ID: 1}},
	}, bot.api, bot)

	kb := NewInlineKeyboard().AddRow(ButtonCallback("OK", "ok")).Build()
	if err := ctx.ReplyWithKeyboardAndParseMode("text", kb, ParseModeMarkdown); err != nil {
		t.Fatalf("ReplyWithKeyboardAndParseMode: %v", err)
	}
}
