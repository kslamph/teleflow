/*
Package teleflow provides a simple, type-safe, and opinionated framework for
building Telegram bots in Go. It is designed to handle complex, stateful
conversations (flows) with ease while still allowing for simple, stateless
commands and middleware.

# Key Features

  - Type-Safe Flows: Uses Go generics (`teleflow.Flow[T]`) to provide
    compile-time safety for your conversational state. No more type assertions.
  - Non-Linear Conversations: Built-in support for complex conversational paths
    using named steps and `GoTo` actions.
  - Simplified State Management: Automatically manages user state within a flow,
    freeing you to focus on your bot's logic.
  - Middleware Support: A standard `Handler` interface allows for easy integration
    of middleware for features like authentication or logging.
  - Large Callback Data: Overcomes Telegram's 64-byte callback data limit by
    providing a built-in mechanism to store and retrieve large data objects
    (`ctx.RegisterData` and `ctx.GetData`).

# Example

Here is a minimal example of a bot with a simple two-step registration flow:

	package main

	import (
		"fmt"
		"log"
		"os"

		"github.com/kslamph/teleflow/core"
	)

	// 1. Define a struct for your flow's data.
	type UserProfile struct {
		Name string
		Age  int
	}

	func main() {
		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		if token == "" {
			log.Fatal("TELEGRAM_BOT_TOKEN is required")
		}

		// 2. Create a new Flow with your data type.
		profileFlow := teleflow.NewFlow[UserProfile]()

		// 3. Define the steps of your conversation.
		nameStep := teleflow.Step[UserProfile]{
			Name: "Name",
			Prompt: func(ctx *teleflow.Context, data *UserProfile) error {
				return ctx.ReplyText("What is your name?")
			},
			Process: func(ctx *teleflow.Context, data *UserProfile) (teleflow.FlowAction, error) {
				data.Name = ctx.Text()
				return teleflow.Next(), nil
			},
		}

		ageStep := teleflow.Step[UserProfile]{
			Name: "Age",
			Prompt: func(ctx *teleflow.Context, data *UserProfile) error {
				return ctx.ReplyText(fmt.Sprintf("Nice to meet you, %s! How old are you?", data.Name))
			},
			Process: func(ctx *teleflow.Context, data *UserProfile) (teleflow.FlowAction, error) {
				// In a real app, you would parse and validate the age.
				return teleflow.End(), nil // End the flow.
			},
		}

		// 4. Add steps to the flow and set a completion handler.
		profileFlow.AddStep(nameStep).AddStep(ageStep)
		profileFlow.OnComplete(func(ctx *teleflow.Context, data UserProfile) error {
			return ctx.ReplyText(fmt.Sprintf("Thanks, %s! Your profile is saved.", data.Name))
		})

		// 5. Set up the router and bot.
		router := teleflow.NewRouter()
		router.Handle("profile", profileFlow)

		bot, err := teleflow.NewBot(token, router)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Bot starting...")
		log.Fatal(bot.Start())
	}
*/
package teleflow
