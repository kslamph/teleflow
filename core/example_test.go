package teleflow

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ExampleNewFlow demonstrates creating a Flow, adding steps, and driving
// it through Next/Retry/End/GoTo actions.
func ExampleNewFlow() {
	type Profile struct {
		Name string
		Age  int
	}

	flow := NewFlow[Profile]()
	if _, err := flow.AddStep(Step[Profile]{
		Name: "name",
		Prompt: func(ctx *Context, data *Profile) error {
			return ctx.ReplyText("What is your name?")
		},
		Process: func(ctx *Context, data *Profile) (FlowAction, error) {
			data.Name = ctx.Text()
			return Next(), nil
		},
	}); err != nil {
		panic(err)
	}
	if _, err := flow.AddStep(Step[Profile]{
		Name: "age",
		Prompt: func(ctx *Context, data *Profile) error {
			return ctx.ReplyText("How old are you?")
		},
		Process: func(ctx *Context, data *Profile) (FlowAction, error) {
			return End(), nil
		},
	}); err != nil {
		panic(err)
	}
	flow.OnComplete(func(ctx *Context, data Profile) error {
		return ctx.ReplyText(fmt.Sprintf("Hello %s, age %d!", data.Name, data.Age))
	})

	// The flow has 2 steps and a completion handler.
	_ = flow
}

// ExampleNewFlow_addStepDuplicate demonstrates that AddStep returns an error
// for duplicate step names instead of panicking.
func ExampleNewFlow_addStepDuplicate() {
	type Data struct{}
	flow := NewFlow[Data]()
	step := Step[Data]{Name: "a"}

	if _, err := flow.AddStep(step); err != nil {
		panic(err)
	}
	_, err := flow.AddStep(step) // duplicate name
	fmt.Println(err != nil)
	// Output:
	// true
}

// ExampleNewFlow_goTo demonstrates non-linear flow control with GoTo.
func ExampleNewFlow_goTo() {
	type Data struct{}
	flow := NewFlow[Data]()
	if _, err := flow.AddStep(Step[Data]{
		Name:    "start",
		Prompt:  func(ctx *Context, data *Data) error { return nil },
		Process: func(ctx *Context, data *Data) (FlowAction, error) { return GoTo("end"), nil },
	}); err != nil {
		panic(err)
	}
	if _, err := flow.AddStep(Step[Data]{
		Name:    "end",
		Prompt:  func(ctx *Context, data *Data) error { return nil },
		Process: func(ctx *Context, data *Data) (FlowAction, error) { return End(), nil },
	}); err != nil {
		panic(err)
	}
	// Flow has start -> end via GoTo.
	fmt.Println(len(flow.steps) == 2)
	// Output:
	// true
}

// ExampleNewInlineKeyboard demonstrates building an inline keyboard.
func ExampleNewInlineKeyboard() {
	kb := NewInlineKeyboard().
		AddRow(ButtonCallback("Yes", "yes"), ButtonCallback("No", "no")).
		AddRow(ButtonCallback("Maybe", "maybe"))
	markup := kb.Build()
	fmt.Println(len(markup.InlineKeyboard) == 2)
	fmt.Println(len(markup.InlineKeyboard[0]) == 2)
	fmt.Println(len(markup.InlineKeyboard[1]) == 1)
	// Output:
	// true
	// true
	// true
}

// ExampleNewRouter demonstrates routing commands and a default handler.
func ExampleNewRouter() {
	router := NewRouter()

	var handled []string
	router.Handle("start", Func(func(ctx *Context) error {
		handled = append(handled, "start")
		return nil
	}))
	router.Handle("help", Func(func(ctx *Context) error {
		handled = append(handled, "help")
		return nil
	}))
	router.Default(Func(func(ctx *Context) error {
		handled = append(handled, "default")
		return nil
	}))

	fmt.Println(len(router.routes) == 2)
	fmt.Println(router.defaultRoute != nil)
	// Output:
	// true
	// true
}

// ExampleRouter_HandleUpdate demonstrates that only commands are routed,
// while non-command messages fall through to the default handler.
func ExampleRouter_HandleUpdate() {
	var result strings.Builder
	router := NewRouter()
	router.Handle("hi", Func(func(ctx *Context) error {
		result.WriteString("hi:")
		return nil
	}))
	router.Default(Func(func(ctx *Context) error {
		result.WriteString("default")
		return nil
	}))

	// A command update is routed to the matching handler.
	cmdCtx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text:     "/hi",
			Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 3}},
			From:     &tgbotapi.User{ID: 1, UserName: "alice"},
			Chat:     &tgbotapi.Chat{ID: 1},
		},
	}, nil, nil)
	if err := router.HandleUpdate(cmdCtx); err != nil {
		panic(err)
	}

	// A non-command update hits the default route.
	textCtx := NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "hello", From: &tgbotapi.User{ID: 1}, Chat: &tgbotapi.Chat{ID: 1}},
	}, nil, nil)
	if err := router.HandleUpdate(textCtx); err != nil {
		panic(err)
	}

	fmt.Println(result.String())
	// Output:
	// hi:default
}

// ExampleFunc demonstrates the Func adapter type.
func ExampleFunc() {
	var called bool
	f := Func(func(ctx *Context) error {
		called = true
		return nil
	})
	_ = f
	fmt.Println(!called) // not called yet
	// Output:
	// true
}
