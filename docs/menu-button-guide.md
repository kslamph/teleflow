# Teleflow Menu Button Guide

The Telegram menu button (often seen as a "/" or a dedicated menu icon in the chat input field) provides users with quick access to a bot's primary functionalities. Teleflow allows you to easily configure this menu button to display a list of commands or open a Web App.

## Table of Contents

- [What is the Menu Button?](#what-is-the-menu-button)
- [Menu Button Types](#menu-button-types)
  - [1. Commands Menu Button (`MenuButtonTypeCommands`)](#1-commands-menu-button-menubuttontypecommands)
  - [2. Web App Menu Button (`MenuButtonTypeWebApp`)](#2-web-app-menu-button-menubuttontypewebapp)
  - [3. Default Menu Button (`MenuButtonTypeDefault`)](#3-default-menu-button-menubuttontypedefault)
- [Configuring the Menu Button](#configuring-the-menu-button)
  - [Using `MenuButtonConfig`](#using-menubuttonconfig)
  - [Creating a Commands Menu Button](#creating-a-commands-menu-button)
  - [Creating a Web App Menu Button](#creating-a-web-app-menu-button)
  - [Creating a Default Menu Button](#creating-a-default-menu-button)
- [Setting the Menu Button for Your Bot](#setting-the-menu-button-for-your-bot)
  - [During Bot Initialization (`WithMenuButton` Option)](#during-bot-initialization-withmenubutton-option)
  - [After Bot Initialization (`bot.WithMenuButton()` Method)](#after-bot-initialization-botwithmenubutton-method)
  - [How Teleflow Applies the Menu Button](#how-teleflow-applies-the-menu-button)
- [Dynamic Menu Buttons with `AccessManager`](#dynamic-menu-buttons-with-accessmanager)
- [Important Notes](#important-notes)
- [Example](#example)
- [Next Steps](#next-steps)

## What is the Menu Button?

The menu button is a standard UI element in Telegram chats with bots. It offers a persistent way for users to discover and access your bot's core features without needing to remember specific commands.

## Menu Button Types

Teleflow supports the main types of menu buttons defined by Telegram:

### 1. Commands Menu Button (`MenuButtonTypeCommands`)
This type configures the menu button to display a list of bot commands. When the user taps the menu button, a panel slides up showing the commands you've defined, along with their descriptions.
- The commands listed are also typically registered with Telegram via `setMyCommands` so they appear as autocomplete suggestions when the user types `/`.
- Teleflow handles this registration automatically when you use `MenuButtonTypeCommands` and add commands to it.

### 2. Web App Menu Button (`MenuButtonTypeWebApp`)
This type configures the menu button to open a [Telegram Web App](https://core.telegram.org/bots/webapps). When tapped, it launches the specified web app.
- You need to provide the text for the button and the URL of your web app.
- **Note**: As of `core/menu_button.go` (line 32), full WebApp menu button *setting via Telegram API* might not be fully implemented in Teleflow's `SetMenuButton` function itself, but the configuration is stored. The `commands` type is fully supported for setting the visual menu.

### 3. Default Menu Button (`MenuButtonTypeDefault`)
This type reverts the menu button to its default state (usually showing a simple "/" icon, indicating that the bot accepts commands).

## Configuring the Menu Button

You configure the menu button using the `teleflow.MenuButtonConfig` struct (defined in `core/keyboards.go`). Helper functions are provided in `core/menu_button.go` to create these configurations.

### Using `MenuButtonConfig`
The `MenuButtonConfig` struct:
```go
type MenuButtonConfig struct {
	Type   MenuButtonType   `json:"type"`
	Text   string           `json:"text,omitempty"`    // For web_app type
	WebApp *WebAppInfo      `json:"web_app,omitempty"` // For web_app type
	Items  []MenuButtonItem `json:"items,omitempty"`   // For commands type
}

type MenuButtonItem struct {
	Text    string `json:"text"`    // Description of the command
	Command string `json:"command"` // The command itself (e.g., "/start")
}

type WebAppInfo struct {
	URL string `json:"url"`
}
```

### Creating a Commands Menu Button
Use `teleflow.NewCommandsMenuButton()` and then `AddCommand(text, command)`:
```go
import teleflow "github.com/kslamph/teleflow/core"

commandsMenu := teleflow.NewCommandsMenuButton().
    AddCommand("🚀 Start Bot", "/start").       // Text is description, Command is the actual command
    AddCommand("ℹ️ Get Help", "/help").
    AddCommand("⚙️ Settings", "/settings")
```
The `text` is what the user sees as the description, and `command` is the actual command string (e.g., "/start"). Teleflow will strip the leading slash if present when registering with `setMyCommands`.

### Creating a Web App Menu Button
Use `teleflow.NewWebAppMenuButton(text, url)`:
```go
webAppMenu := teleflow.NewWebAppMenuButton("🛍️ Open Store", "https://your-store-webapp.com/app")
```

### Creating a Default Menu Button
Use `teleflow.NewDefaultMenuButton()`:
```go
defaultMenu := teleflow.NewDefaultMenuButton()
```

## Setting the Menu Button for Your Bot

Once you have a `*MenuButtonConfig`, you can associate it with your bot.

### During Bot Initialization (`WithMenuButton` Option)
This is the recommended way to set a global default menu button for your bot.
```go
bot, err := teleflow.NewBot(botToken,
    teleflow.WithMenuButton(commandsMenu), // Pass your MenuButtonConfig here
)
```

### After Bot Initialization (`bot.WithMenuButton()` Method)
You can also set or change the menu button after the bot has been created:
```go
bot.WithMenuButton(webAppMenu)
// Note: This updates the bot's internal configuration.
// The actual Telegram menu button is set when bot.Start() calls InitializeMenuButton(),
// or if you use an AccessManager that dynamically sets it.
```

### How Teleflow Applies the Menu Button
- When `bot.Start()` is called, it internally calls `bot.InitializeMenuButton()`.
- `InitializeMenuButton()` then calls `bot.SetDefaultMenuButton()`.
- `SetDefaultMenuButton()` uses the `b.menuButton` configuration.
- For `MenuButtonTypeCommands`, it calls `b.setMyCommands()` which registers the commands with Telegram, making them appear in the menu.
- For other types, it currently logs the action or prepares the configuration. Direct API calls to `setChatMenuButton` for WebApp or Default types per chat might be handled differently or by an `AccessManager`.

The primary effect you'll see immediately with the built-in logic is for `MenuButtonTypeCommands`, which populates the command list.

## Dynamic Menu Buttons with `AccessManager`
For more advanced scenarios, such as having different menu buttons for different users or chats, you can use an `AccessManager`. The `AccessManager` interface includes a method:
```go
GetMenuButton(ctx *teleflow.MenuContext) *teleflow.MenuButtonConfig
```
If an `AccessManager` is configured and this method returns a `MenuButtonConfig`, Teleflow's `Context` methods (like `ctx.Reply`) may attempt to set this chat-specific menu button. This allows for dynamic, context-aware menu buttons. (The exact mechanism for per-chat updates via `AccessManager` would need to be verified in `core/context.go`'s send methods or related logic).

The global menu button set via `WithMenuButton` serves as the default.

## Important Notes
- **Telegram Caching**: Telegram clients can cache the command list. If you change your commands menu, users might need to restart their Telegram client or clear cache to see the changes immediately.
- **Command Registration**: For `MenuButtonTypeCommands`, Teleflow automatically attempts to register the commands with Telegram using `setMyCommands`. This is what makes them appear in the "/" menu.
- **Scope**: The `setMyCommands` API call is global for the bot. Per-chat or per-user command lists are more complex and usually involve filtering commands within your bot's logic rather than changing the Telegram-side menu for every user. `AccessManager` can help simulate this by providing different UI elements.

## Example
```go
package main

import (
	"log"
	"os"
	teleflow "github.com/kslamph/teleflow/core"
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")

	// 1. Define a commands menu button configuration
	myMenu := teleflow.NewCommandsMenuButton().
		AddCommand("🚀 Start", "/start").
		AddCommand("❓ Help", "/help").
		AddCommand("📢 About", "/about")

	// 2. Create the bot and set the menu button using the BotOption
	bot, err := teleflow.NewBot(botToken,
		teleflow.WithMenuButton(myMenu),
	)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// 3. Define handlers for these commands
	bot.HandleCommand("start", func(ctx *teleflow.Context) error {
		return ctx.Reply("Bot started! Welcome!")
	})
	bot.HandleCommand("help", func(ctx *teleflow.Context) error {
		return ctx.Reply("This is the help message. Available commands are in the menu!")
	})
	bot.HandleCommand("about", func(ctx *teleflow.Context) error {
		return ctx.Reply("This bot is built with Teleflow.")
	})

	// 4. Start the bot
	// Upon starting, InitializeMenuButton will be called, which sets up the commands.
	log.Println("Bot starting...")
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
```
After running this bot, users will see a menu button. Tapping it will show "Start", "Help", and "About" with their descriptions.

## Next Steps
- [Handlers Guide](handlers-guide.md): To implement the logic for the commands you add to your menu.
- [Keyboards Guide](keyboards-guide.md): For other types of interactive buttons.
- [API Reference](api-reference.md): For details on `MenuButtonConfig` and related types.
- Explore `core/menu_button.go` and `core/bot.go` (specifically `InitializeMenuButton` and `setMyCommands`) for implementation details.