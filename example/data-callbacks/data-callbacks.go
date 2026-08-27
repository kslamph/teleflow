package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	teleflow "github.com/kslamph/teleflow/core"
)

// Employee represents our complex data structure.
type Employee struct {
	ID   int
	Name string
	Role string
}

// FlowData holds the state for our example flow.
type FlowData struct {
	// No persistent state needed for this simple example.
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Sample complex data. Imagine this comes from a database.
	employees := map[string][]Employee{
		"Engineering": {
			{ID: 101, Name: "Alice", Role: "Backend Developer"},
			{ID: 102, Name: "Bob", Role: "Frontend Developer"},
		},
		"Sales": {
			{ID: 201, Name: "Charlie", Role: "Account Executive"},
			{ID: 202, Name: "Diana", Role: "Sales Manager"},
		},
	}

	dataFlow := teleflow.NewFlow[FlowData]()

	viewStep := teleflow.Step[FlowData]{
		Name: "ViewDepartments",
		Prompt: func(ctx *teleflow.Context, data *FlowData) error {
			keyboard := teleflow.NewInlineKeyboard()
			prompt := "Select a department to view its employees:\n"

			// For each department, register its employee slice and create a button.
			for deptName, employeeSlice := range employees {
				// The large/complex data (employeeSlice) is stored on the server.
				// Only the short, unique ID is sent to the user in the callback.
				dataID := ctx.RegisterData(employeeSlice)
				keyboard.AddRow(teleflow.ButtonCallback(deptName, dataID))
				prompt += fmt.Sprintf("- %s\n", deptName)
			}

			return ctx.ReplyWithKeyboard(prompt, keyboard.Build())
		},
		Process: func(ctx *teleflow.Context, data *FlowData) (teleflow.FlowAction, error) {
			// Retrieve the data associated with the button press.
			retrievedData, ok := ctx.GetData()
			if !ok {
				ctx.ReplyText("Sorry, that data has expired or is invalid. Please try again.")
				return teleflow.Retry(), nil
			}

			// Perform a type assertion to get our original data structure back.
			employeeSlice, ok := retrievedData.([]Employee)
			if !ok {
				log.Printf("Type assertion failed. Expected []Employee, got %T", retrievedData)
				ctx.ReplyText("An internal error occurred.")
				return teleflow.End(), nil
			}

			// Now we can use the complex data.
			var builder strings.Builder
			builder.WriteString("Employees:\n")
			for _, emp := range employeeSlice {
				builder.WriteString(fmt.Sprintf("- %s (%s)\n", emp.Name, emp.Role))
			}

			ctx.AnswerCallback()
			ctx.EditMessageText(ctx.Update().CallbackQuery.Message.MessageID, builder.String())

			// End the flow. In a real app, you might go back to the start.
			return teleflow.End(), nil
		},
	}

	if _, err := dataFlow.AddStep(viewStep); err != nil {
		log.Fatal(err)
	}

	router := teleflow.NewRouter()
	router.Handle("data", dataFlow)
	router.Default(teleflow.Func(func(ctx *teleflow.Context) error {
		if dataFlow.IsUserInFlow(ctx.UserID()) {
			return dataFlow.HandleUpdate(ctx)
		}
		return ctx.ReplyText("Unknown command. Use /data to start.")
	}))

	bot, err := teleflow.NewBot(token, router)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.SetBotCommands(map[string]string{
		"data": "Demonstrate handling of large callback data",
	})

	log.Println("Data callback bot starting...")
	log.Fatal(bot.Start())
}
