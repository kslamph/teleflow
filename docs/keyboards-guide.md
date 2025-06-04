# Teleflow Keyboards Guide

Telegram keyboards are essential for creating interactive and user-friendly bots. Teleflow provides intuitive abstractions for building both **Reply Keyboards** and **Inline Keyboards**. This guide covers how to create, configure, and use them in your bot.

## Table of Contents

- [Introduction to Telegram Keyboards](#introduction-to-telegram-keyboards)
- [1. Reply Keyboards](#1-reply-keyboards)
  - [What are Reply Keyboards?](#what-are-reply-keyboards)
  - [Creating a Reply Keyboard](#creating-a-reply-keyboard)
  - [Adding Buttons and Rows](#adding-buttons-and-rows)
  - [Configuring Reply Keyboards](#configuring-reply-keyboards)
    - [`Resize()`](#resize)
    - [`OneTime()`](#onetime)
    - [`Placeholder(text string)`](#placeholdertext-string)
  - [Special Reply Keyboard Buttons](#special-reply-keyboard-buttons)
    - [Requesting Contact Information](#requesting-contact-information)
    - [Requesting Location](#requesting-location)
    - [Web App Button (Reply Keyboard)](#web-app-button-reply-keyboard)
  - [Sending a Message with a Reply Keyboard](#sending-a-message-with-a-reply-keyboard)
  - [Removing a Reply Keyboard](#removing-a-reply-keyboard)
  - [Handling Reply Keyboard Button Presses](#handling-reply-keyboard-button-presses)
  - [Example](#example)
- [2. Inline Keyboards](#2-inline-keyboards)
  - [What are Inline Keyboards?](#what-are-inline-keyboards)
  - [Creating an Inline Keyboard](#creating-an-inline-keyboard)
  - [Adding Buttons and Rows](#adding-buttons-and-rows-1)
    - [Callback Buttons](#callback-buttons)
    - [URL Buttons](#url-buttons)
    - [Web App Buttons (Inline Keyboard)](#web-app-buttons-inline-keyboard)
    - [Switch to Inline Query Buttons](#switch-to-inline-query-buttons)
  - [Sending a Message with an Inline Keyboard](#sending-a-message-with-an-inline-keyboard)
  - [Handling Inline Keyboard Button Presses (Callbacks)](#handling-inline-keyboard-button-presses-callbacks)
  - [Example](#example-1)
- [Automatic UI Management with AccessManager](#automatic-ui-management-with-accessmanager)
- [Best Practices for Keyboards](#best-practices-for-keyboards)
- [Next Steps](#next-steps)

## Introduction to Telegram Keyboards

Teleflow simplifies the creation of two main types of Telegram keyboards:
- **Reply Keyboards**: These keyboards replace the standard text input field with custom buttons. They are persistent or one-time and are good for main navigation or common actions.
- **Inline Keyboards**: These keyboards are attached directly to a specific message. Their buttons can trigger callbacks, open URLs, or perform other interactive actions without sending a text message.

Both keyboard types are built using a fluent API provided by `teleflow.NewReplyKeyboard()` and `teleflow.NewInlineKeyboard()`.

## 1. Reply Keyboards

### What are Reply Keyboards?
Reply keyboards (also known as custom keyboards) appear in place of the user's standard keyboard. When a user taps a button on a reply keyboard, the button's text is sent as a regular message to the bot.

### Creating a Reply Keyboard
You start by creating a `ReplyKeyboard` object:
```go
import teleflow "github.com/kslamph/teleflow/core"

kb := teleflow.NewReplyKeyboard()
```

### Adding Buttons and Rows
Use `AddButton(text string)` to add a button to the current row and `AddRow()` to start a new row of buttons.
```go
kb.AddButton("Profile") // Adds "Profile" to the current row
kb.AddButton("Settings") // Adds "Settings" to the same row

kb.AddRow() // Starts a new row

kb.AddButton("Help") // Adds "Help" to the new row
```
Buttons are added left-to-right in a row. `AddRow()` finalizes the current row and prepares for a new one.

### Configuring Reply Keyboards
`ReplyKeyboard` offers several methods to customize its appearance and behavior:

#### `Resize()`
Call `kb.Resize()` to make the keyboard vertically smaller if the buttons allow. This is generally recommended for a cleaner look.
```go
kb.Resize()
```

#### `OneTime()`
Call `kb.OneTime()` to make the keyboard hide automatically after the user presses a button. Useful for quick choices.
```go
kb.OneTime()
```

#### `Placeholder(text string)`
Call `kb.Placeholder("Search...")` to set placeholder text that appears in the input field when the keyboard is active.
```go
kb.Placeholder("Choose an option or type...")
```

### Special Reply Keyboard Buttons

Reply keyboards can also include buttons with special functionalities:

#### Requesting Contact Information
```go
kb.AddRow().AddButton("Share My Contact").RequestContact()
// Note: `RequestContact()` modifies the last added button to request contact.
// It's often clearer to create the button object directly for special types:
contactButton := teleflow.ReplyKeyboardButton{Text: "Share Contact", RequestContact: true}
// Then add it to a row: kb.currentRow = append(kb.currentRow, contactButton)
// Or use the AddRequestContact helper:
kb.AddRow().AddRequestContact() // Adds a button with "Share Contact" text
```
When pressed, the user will be prompted to share their phone number with the bot.

#### Requesting Location
```go
kb.AddRow().AddButton("Share My Location").RequestLocation()
// Or use the AddRequestLocation helper:
kb.AddRow().AddRequestLocation() // Adds a button with "Share Location" text
```
When pressed, the user will be prompted to share their current location.

#### Web App Button (Reply Keyboard)
You can open a [Telegram Web App](https://core.telegram.org/bots/webapps) using a reply keyboard button.
```go
webAppButton := teleflow.ReplyKeyboardButton{
    Text:   "Open Store",
    WebApp: &teleflow.WebAppInfo{URL: "https://your-webapp-url.com"},
}
// Add this button to your keyboard structure.
// Example:
// kb.currentRow = append(kb.currentRow, webAppButton)
// kb.AddRow()
```

### Sending a Message with a Reply Keyboard
Pass the `ReplyKeyboard` object as an argument to `ctx.Reply()` or `ctx.ReplyTemplate()`:
```go
err := ctx.Reply("Please choose an option:", kb)
```

### Removing a Reply Keyboard
To remove a reply keyboard and restore the standard text input, send a message with `tgbotapi.NewRemoveKeyboard(true)`:
```go
import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

removeMarkup := tgbotapi.NewRemoveKeyboard(true) // true to remove selectively for targeted users
msg := tgbotapi.NewMessage(ctx.ChatID(), "Keyboard removed.")
msg.ReplyMarkup = removeMarkup
_, err := ctx.Bot.API.Send(msg)
```
Alternatively, if your `AccessManager` returns `nil` for `GetReplyKeyboard`, the keyboard might be removed on the next reply.

### Handling Reply Keyboard Button Presses
When a reply keyboard button is pressed, its text is sent as a regular message. You handle these using `bot.HandleText()`:
```go
bot.HandleText("Profile", func(ctx *teleflow.Context) error {
    // Logic for when "Profile" button is pressed
    return ctx.Reply("Showing your profile...")
})
```
See the [Handlers Guide](handlers-guide.md#2-text-handlers) for more.

### Example
```go
package main

import (
	"log"
	"os"
	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	bot, _ := teleflow.NewBot(os.Getenv("BOT_TOKEN"))

	bot.HandleCommand("menu", func(ctx *teleflow.Context) error {
		kb := teleflow.NewReplyKeyboard().
			AddButton("Info").AddButton("Help").AddRow().
			AddButton("Settings").AddRow().
			Resize().
			Placeholder("Select an action...")
		return ctx.Reply("Main Menu:", kb)
	})

	bot.HandleText("Info", func(ctx *teleflow.Context) error {
		return ctx.Reply("This bot provides information.")
	})
    // ... other text handlers for "Help", "Settings"

	log.Fatal(bot.Start())
}
```

## 2. Inline Keyboards

### What are Inline Keyboards?
Inline keyboards are attached directly to messages sent by your bot. Their buttons can trigger actions like sending callback data, opening URLs, or switching to inline mode. They do *not* send text messages when pressed.

### Creating an Inline Keyboard
```go
kb := teleflow.NewInlineKeyboard()
```

### Adding Buttons and Rows
Similar to reply keyboards, use `AddButton()` (with specific types) and `AddRow()`.

#### Callback Buttons
These buttons send a `callback_data` string to your bot when pressed. This is the most common type for interactive inline UIs.
```go
// Adds a button with text "Confirm" and callback data "action_confirm_123"
kb.AddButton("Confirm", "action_confirm_123")
```

#### URL Buttons
These buttons open a URL in the user's browser.
```go
kb.AddURL("Visit Website", "https://example.com")
```

#### Web App Buttons (Inline Keyboard)
Open a Telegram Web App.
```go
webAppInfo := teleflow.WebAppInfo{URL: "https://your-webapp-url.com"}
kb.AddWebApp("Open App", webAppInfo)
```

#### Switch to Inline Query Buttons
These buttons prompt the user to choose a chat, then switch to inline mode with the bot's username and an optional query.
```go
// Switch to inline mode in the current chat with "share " query
kb.currentRow = append(kb.currentRow, teleflow.InlineKeyboardButton{
    Text:                         "Share Content",
    SwitchInlineQueryCurrentChat: "share ",
})
kb.AddRow()

// Switch to inline mode, allowing user to pick a chat
kb.currentRow = append(kb.currentRow, teleflow.InlineKeyboardButton{
    Text:              "Share with Friend",
    SwitchInlineQuery: "share_profile",
})
kb.AddRow()
```

### Sending a Message with an Inline Keyboard
Pass the `InlineKeyboard` object to `ctx.Reply()`, `ctx.ReplyTemplate()`, or message editing methods:
```go
err := ctx.Reply("Do you want to proceed?", kb)
```

### Handling Inline Keyboard Button Presses (Callbacks)
Callback button presses are handled by [Callback Handlers](handlers-guide.md#3-callback-handlers-for-inline-keyboards). You register these using `bot.RegisterCallback()`:
```go
bot.RegisterCallback(teleflow.SimpleCallback("action_confirm_*", func(ctx *teleflow.Context, data string) error {
    // 'data' will be the part of callback_data after "action_confirm_"
    // e.g., if callback_data was "action_confirm_123", data is "123"
    ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Confirmed!") // Acknowledge the press
    return ctx.EditOrReply("Action confirmed for ID: " + data)
}))
```
**Important**: Always answer callback queries (e.g., using `ctx.Bot.AnswerCallbackQuery()` or by successfully editing the message with `ctx.EditOrReply()`) to remove the loading state on the button.

### Example
```go
package main

import (
	"log"
	"os"
	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	bot, _ := teleflow.NewBot(os.Getenv("BOT_TOKEN"))

	bot.HandleCommand("task", func(ctx *teleflow.Context) error {
		taskID := "42"
		kb := teleflow.NewInlineKeyboard().
			AddButton("Complete", "complete_task_"+taskID).
			AddButton("Delete", "delete_task_"+taskID).AddRow().
			AddURL("More Info", "https://example.com/task/"+taskID).AddRow()
		return ctx.Reply("Task #"+taskID+": Review report", kb)
	})

	bot.RegisterCallback(teleflow.SimpleCallback("complete_task_*", func(ctx *teleflow.Context, taskID string) error {
		ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Task marked complete")
		return ctx.EditOrReply("Task " + taskID + " marked as complete!")
	}))

	bot.RegisterCallback(teleflow.SimpleCallback("delete_task_*", func(ctx *teleflow.Context, taskID string) error {
		ctx.Bot.AnswerCallbackQuery(ctx.Update.CallbackQuery.ID, "Task deleted")
		return ctx.EditOrReply("Task " + taskID + " deleted.")
	}))

	log.Fatal(bot.Start())
}
```

## Automatic UI Management with AccessManager
If you use an `AccessManager`, Teleflow can automatically provide context-aware reply keyboards. The `AccessManager` interface has a `GetReplyKeyboard(ctx *MenuContext) *ReplyKeyboard` method. If this method returns a keyboard, Teleflow will attempt to apply it to messages sent via `ctx.Reply()` unless a specific keyboard is provided in the `ctx.Reply()` call.

See `core/context.go` (the `send` method) and `core/bot.go` (the `AccessManager` interface) for details. This is an advanced feature useful for role-based UIs.

## Best Practices for Keyboards

- **Clarity**: Keep button text short and clear.
- **Responsiveness**: For inline keyboards, always answer callback queries promptly.
- **Simplicity**: Don't overwhelm users with too many buttons.
- **Consistency**: Maintain a consistent style and layout for your keyboards.
- **Use `Resize()`**: For reply keyboards, `Resize()` usually improves the look.
- **Consider `OneTime()`**: For reply keyboards that present a single choice, `OneTime()` can provide a smoother experience.

## Next Steps

- [Handlers Guide](handlers-guide.md): Understand how to process keyboard interactions.
- [Flow Guide](flow-guide.md): Use keyboards as part of multi-step conversations.
- [Menu Button Guide](menu-button-guide.md): For the main bot menu button.
- [API Reference](api-reference.md): Detailed information on keyboard types and methods.