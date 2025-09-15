package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	teleflow "github.com/kslamph/teleflow/core"
)

// 1. Define a struct for your flow's data for type safety.
type RegistrationData struct {
	Name             string
	Age              int
	FavoriteLanguage string
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// 2. Create a new Flow with your data type.
	registrationFlow := teleflow.NewFlow[RegistrationData]()

	// 3. Define steps with clean, type-safe functions.
	welcomeStep := teleflow.Step[RegistrationData]{
		Name: "Welcome",
		Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
			photo := tgbotapi.NewPhoto(ctx.ChatID(), tgbotapi.FileURL("https://www.vhv.rs/dpng/d/525-5254072_go-language-logo-png-transparent-png.png"))
			photo.Caption = "Welcome! What is your name?"
			_, err := ctx.API().Send(photo)
			return err
		},
		Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
			data.Name = ctx.Text()
			return teleflow.Next(), nil
		},
	}

	ageStep := teleflow.Step[RegistrationData]{
		Name: "Age",
		Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
			return ctx.ReplyText(fmt.Sprintf("Nice to meet you, %s! How old are you?", data.Name))
		},
		Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
			age, err := strconv.Atoi(ctx.Text())
			if err != nil {
				ctx.ReplyText("Please enter a valid number.")
				return teleflow.Retry(), nil
			}
			data.Age = age
			return teleflow.Next(), nil
		},
	}

	languageStep := teleflow.Step[RegistrationData]{
		Name: "Language",
		Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
			keyboard := teleflow.NewInlineKeyboard().AddRow(
				teleflow.ButtonCallback("Go", "lang_go"),
				teleflow.ButtonCallback("Python", "lang_python"),
				teleflow.ButtonCallback("Rust", "lang_rust"),
			).Build()
			return ctx.ReplyWithKeyboard("What is your favorite programming language?", keyboard)
		},
		Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
			// We expect a callback query here.
			lang := ""
			switch ctx.Text() {
			case "lang_go":
				lang = "Go"
			case "lang_python":
				lang = "Python"
			case "lang_rust":
				lang = "Rust"
			default:
				// If the user types something instead of clicking.
				ctx.ReplyText("Please choose one of the options.")
				return teleflow.Retry(), nil
			}
			data.FavoriteLanguage = lang
			ctx.AnswerCallback() // Acknowledge the button press.
			return teleflow.Next(), nil
		},
	}

	confirmationStep := teleflow.Step[RegistrationData]{
		Name: "Confirmation",
		Prompt: func(ctx *teleflow.Context, data *RegistrationData) error {
			keyboard := teleflow.NewInlineKeyboard().AddRow(
				teleflow.ButtonCallback("✅ Yes, Correct", "confirm"),
				teleflow.ButtonCallback("❌ No, Restart", "restart"),
			).Build()

			promptText := fmt.Sprintf(
				"Please confirm your details:\nName: %s\nAge: %d\nFavorite Language: %s",
				data.Name, data.Age, data.FavoriteLanguage,
			)
			return ctx.ReplyWithKeyboard(promptText, keyboard)
		},
		Process: func(ctx *teleflow.Context, data *RegistrationData) (teleflow.FlowAction, error) {
			switch ctx.Text() {
			case "confirm":
				ctx.AnswerCallback()
				return teleflow.End(), nil
			case "restart":
				ctx.AnswerCallback()
				ctx.EditMessageText(ctx.Update().CallbackQuery.Message.MessageID, "Okay, let's start over.")
				return teleflow.GoTo("Welcome"), nil
			default:
				// Handle case where user types instead of clicking buttons.
				ctx.ReplyText("Please use the buttons to confirm or restart.")
				return teleflow.Retry(), nil
			}
		},
	}

	// 4. Add steps to the flow.
	registrationFlow.AddStep(welcomeStep).AddStep(ageStep).AddStep(languageStep).AddStep(confirmationStep)

	// 5. Set a completion handler.
	registrationFlow.OnComplete(func(ctx *teleflow.Context, data RegistrationData) error {
		// The final `data` is type-safe.
		return ctx.ReplyText(fmt.Sprintf("Thanks, %s! Registration complete. Age: %d, Favorite Language: %s", data.Name, data.Age, data.FavoriteLanguage))
	})

	// 6. Set up the router and bot.
	router := teleflow.NewRouter()
	router.Handle("register", registrationFlow)

	router.Handle("cancel", teleflow.Func(func(ctx *teleflow.Context) error {
		if registrationFlow.IsUserInFlow(ctx.UserID()) {
			registrationFlow.CancelFlow(ctx.UserID())
			return ctx.ReplyText("Flow cancelled.")
		}
		return ctx.ReplyText("No active flow to cancel.")
	}))

	router.Handle("hello", teleflow.Func(func(ctx *teleflow.Context) error {
		return ctx.ReplyText("Hello!")
	}))

	router.Handle("broadcast", teleflow.Func(func(ctx *teleflow.Context) error {
		msg := tgbotapi.NewMessage(-640141289, "This is a broadcast message.")
		if _, err := ctx.API().Send(msg); err != nil {
			log.Printf("Failed to send broadcast message: %v", err)
			return ctx.ReplyText("Sorry, I couldn't send the broadcast message.")
		}
		return ctx.ReplyText("Broadcast message sent successfully!")
	}))

	router.Handle("menu", teleflow.Func(func(ctx *teleflow.Context) error {
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/hello"),
				tgbotapi.NewKeyboardButton("/broadcast"),
				tgbotapi.NewKeyboardButton("/register"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/cancel"),
				tgbotapi.NewKeyboardButton("/close"),
			),
		)
		keyboard.ResizeKeyboard = true
		return ctx.ReplyWithKeyboard("Menu:", keyboard)
	}))

	router.Handle("close", teleflow.Func(func(ctx *teleflow.Context) error {
		return ctx.ReplyWithKeyboard("Closing menu.", tgbotapi.NewRemoveKeyboard(true))
	}))

	router.Default(teleflow.Func(func(ctx *teleflow.Context) error {
		// If the user is already in a flow, pass the update to the flow.
		if registrationFlow.IsUserInFlow(ctx.UserID()) {
			return registrationFlow.HandleUpdate(ctx)
		}
		return ctx.ReplyText("Unknown command. Use /register to begin.")
	}))

	bot, err := teleflow.NewBot(token, router)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.SetBotCommands(map[string]string{
		"register":  "Start the user registration flow",
		"cancel":    "Cancel the current flow",
		"hello":     "Say hello",
		"broadcast": "Send a broadcast message",
		"menu":      "Show the main menu",
		"close":     "Close the main menu",
	})

	log.Println("Bot starting...")
	log.Fatal(bot.Start())
}
