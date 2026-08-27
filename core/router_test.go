package teleflow

import (
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestRouterHandleUpdateRoutesCommands routes a command to its handler.
func TestRouterHandleUpdateRoutesCommands(t *testing.T) {
	router := NewRouter()
	called := ""
	router.Handle("start", Func(func(ctx *Context) error {
		called = "start"
		return nil
	}))

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text:     "/start",
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}},
			From:     &tgbotapi.User{ID: 1},
			Chat:     &tgbotapi.Chat{ID: 1},
		},
	}, nil, nil)

	if err := router.HandleUpdate(ctx); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if called != "start" {
		t.Fatalf("expected 'start' handler called, got %q", called)
	}
}

// TestRouterHandleUpdateUnknownCommandFallsToDefault routes unknown commands
// to the default handler.
func TestRouterHandleUpdateUnknownCommandFallsToDefault(t *testing.T) {
	router := NewRouter()
	called := ""
	router.Handle("start", Func(func(ctx *Context) error {
		called = "start"
		return nil
	}))
	router.Default(Func(func(ctx *Context) error {
		called = "default"
		return nil
	}))

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text:     "/unknown",
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 8}},
			From:     &tgbotapi.User{ID: 1},
			Chat:     &tgbotapi.Chat{ID: 1},
		},
	}, nil, nil)

	if err := router.HandleUpdate(ctx); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if called != "default" {
		t.Fatalf("expected default handler called, got %q", called)
	}
}

// TestRouterHandleUpdateNonCommandFallsToDefault routes plain messages
// to the default handler.
func TestRouterHandleUpdateNonCommandFallsToDefault(t *testing.T) {
	router := NewRouter()
	called := ""
	router.Handle("start", Func(func(ctx *Context) error {
		called = "start"
		return nil
	}))
	router.Default(Func(func(ctx *Context) error {
		called = "default"
		return nil
	}))

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "hello",
			From: &tgbotapi.User{ID: 1},
			Chat: &tgbotapi.Chat{ID: 1},
		},
	}, nil, nil)

	if err := router.HandleUpdate(ctx); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if called != "default" {
		t.Fatalf("expected default handler called, got %q", called)
	}
}

// TestRouterHandleUpdateNoDefaultIsNoop: without a default handler, unmatched
// updates are silently dropped.
func TestRouterHandleUpdateNoDefaultIsNoop(t *testing.T) {
	router := NewRouter()
	called := false
	router.Handle("start", Func(func(ctx *Context) error {
		called = true
		return nil
	}))

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "hello",
			From: &tgbotapi.User{ID: 1},
			Chat: &tgbotapi.Chat{ID: 1},
		},
	}, nil, nil)

	if err := router.HandleUpdate(ctx); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if called {
		t.Fatal("handler should not be called")
	}
}

// TestRouterHandleUpdatePropagatesErrors verifies handler errors propagate.
func TestRouterHandleUpdatePropagatesErrors(t *testing.T) {
	router := NewRouter()
	wantErr := errors.New("boom")
	router.Handle("start", Func(func(ctx *Context) error {
		return wantErr
	}))

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text:     "/start",
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}},
			From:     &tgbotapi.User{ID: 1},
			Chat:     &tgbotapi.Chat{ID: 1},
		},
	}, nil, nil)

	if err := router.HandleUpdate(ctx); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

// TestRouterOverwriteHandle replaces an existing command handler.
func TestRouterOverwriteHandle(t *testing.T) {
	router := NewRouter()
	called := ""
	router.Handle("start", Func(func(ctx *Context) error {
		called = "first"
		return nil
	}))
	router.Handle("start", Func(func(ctx *Context) error {
		called = "second"
		return nil
	}))

	ctx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text:     "/start",
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}},
			From:     &tgbotapi.User{ID: 1},
			Chat:     &tgbotapi.Chat{ID: 1},
		},
	}, nil, nil)

	if err := router.HandleUpdate(ctx); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if called != "second" {
		t.Fatalf("expected 'second' handler, got %q", called)
	}
}

// TestFuncAdapter verifies the Func adapter implements Handler.
func TestFuncAdapter(t *testing.T) {
	var called bool
	f := Func(func(ctx *Context) error {
		called = true
		return nil
	})
	var h Handler = f
	if err := h.HandleUpdate(nil); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if !called {
		t.Fatal("Func not called")
	}
}

// TestInlineKeyboard verifies the keyboard builder.
func TestInlineKeyboard(t *testing.T) {
	kb := NewInlineKeyboard().
		AddRow(ButtonCallback("Yes", "yes"), ButtonCallback("No", "no")).
		AddRow(ButtonCallback("Maybe", "maybe"))
	markup := kb.Build()

	if len(markup.InlineKeyboard) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(markup.InlineKeyboard))
	}
	if len(markup.InlineKeyboard[0]) != 2 {
		t.Fatalf("expected 2 buttons in row 0, got %d", len(markup.InlineKeyboard[0]))
	}
	if markup.InlineKeyboard[0][0].Text != "Yes" || markup.InlineKeyboard[0][0].CallbackData == nil || *markup.InlineKeyboard[0][0].CallbackData != "yes" {
		t.Fatalf("unexpected first button: %+v", markup.InlineKeyboard[0][0])
	}
	if len(markup.InlineKeyboard[1]) != 1 {
		t.Fatalf("expected 1 button in row 1, got %d", len(markup.InlineKeyboard[1]))
	}
}

// TestButtonCallback verifies the button helper.
func TestButtonCallback(t *testing.T) {
	btn := ButtonCallback("Label", "data")
	if btn.Text != "Label" || btn.CallbackData == nil || *btn.CallbackData != "data" {
		t.Fatalf("unexpected button: %+v", btn)
	}
}

// TestInlineKeyboardEmpty verifies an empty builder.
func TestInlineKeyboardEmpty(t *testing.T) {
	markup := NewInlineKeyboard().Build()
	if len(markup.InlineKeyboard) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(markup.InlineKeyboard))
	}
}
