# Native Menu Button Support

Teleflow provides comprehensive support for Telegram's native menu button - the button that appears beside the attachment icon in the chat input field. This feature allows you to define custom commands or link to mini-apps (Web Apps) for enhanced user interaction.

## Overview

The native menu button provides three types of functionality:

1. **Commands Menu** (`MenuButtonTypeCommands`) - Shows a list of bot commands
2. **Web App Menu** (`MenuButtonTypeWebApp`) - Opens a mini-app/web application
3. **Default Menu** (`MenuButtonTypeDefault`) - Standard Telegram menu behavior

## Basic Usage

### Commands Menu Button

The most common use case is to display bot commands in an organized menu:

```go
// Configure menu button with bot commands
menuButton := &teleflow.MenuButtonConfig{
    Type: teleflow.MenuButtonTypeCommands,
    Items: []teleflow.MenuButtonItem{
        {
            Text:    "📖 Help",
            Command: "/help",
        },
        {
            Text:    "📊 Stats", 
            Command: "/stats",
        },
        {
            Text:    "⚙️ Settings",
            Command: "/settings",
        },
    },
}

// Apply to bot using functional option
bot, err := teleflow.NewBot("YOUR_TOKEN", teleflow.WithMenuButton(menuButton))

// Or set after bot creation
bot.WithMenuButton(menuButton)
```

### Web App Menu Button

For bots that integrate with web applications:

```go
webAppMenu := teleflow.NewWebAppMenuButton("🌐 Open App", "https://your-webapp.com")
bot.WithMenuButton(webAppMenu)
```

### Helper Functions

Teleflow provides convenient helper functions for creating menu buttons:

```go
// Commands menu (auto-shows registered commands)
commandsMenu := teleflow.NewCommandsMenuButton()

// Web app menu
webAppMenu := teleflow.NewWebAppMenuButton("🚀 Launch", "https://app.example.com") 

// Default menu
defaultMenu := teleflow.NewDefaultMenuButton()

// Fluent API for commands
commandsMenu.
    AddCommand("📋 Menu", "/menu").
    AddCommand("📊 Stats", "/stats").
    AddCommand("⚙️ Settings", "/settings")
```

## Configuration Types

### MenuButtonTypeCommands

Displays bot commands in an organized menu format:

```go
&teleflow.MenuButtonConfig{
    Type: teleflow.MenuButtonTypeCommands,
    Items: []teleflow.MenuButtonItem{
        {Text: "🏠 Home", Command: "/start"},
        {Text: "📊 Statistics", Command: "/stats"},
        {Text: "⚙️ Settings", Command: "/settings"},
    },
}
```

**Features:**
- Organizes commands with descriptive text and emojis
- Users can tap to execute commands quickly
- Ideal for bots with multiple features

### MenuButtonTypeWebApp

Opens a web application when pressed:

```go
&teleflow.MenuButtonConfig{
    Type: teleflow.MenuButtonTypeWebApp,
    Text: "🌐 Open Dashboard",
    WebApp: &teleflow.WebAppInfo{
        URL: "https://your-bot-dashboard.com",
    },
}
```

**Features:**
- Seamlessly integrates with web applications
- Supports Telegram Web App APIs
- Perfect for complex interfaces that need HTML/CSS/JS

### MenuButtonTypeDefault

Standard Telegram menu behavior (shows attachment options):

```go
&teleflow.MenuButtonConfig{
    Type: teleflow.MenuButtonTypeDefault,
}
```

## Advanced Usage

### Per-Chat Menu Buttons

You can set different menu buttons for different chats:

```go
// Set for specific chat
err := bot.SetMenuButton(chatID, menuConfig)

// Set default for all chats
err := bot.SetMenuButton(0, menuConfig)

// Get current menu button
currentMenu, err := bot.GetMenuButton(chatID)
```

### Dynamic Menu Configuration

Create context-aware menus based on user permissions:

```go
func createUserMenu(userRole string) *teleflow.MenuButtonConfig {
    menu := teleflow.NewCommandsMenuButton()
    
    // Basic commands for all users
    menu.AddCommand("📋 Help", "/help").
         AddCommand("👤 Profile", "/profile")
    
    // Admin-only commands
    if userRole == "admin" {
        menu.AddCommand("⚙️ Admin Panel", "/admin").
             AddCommand("📊 Analytics", "/analytics")
    }
    
    return menu
}

// In your handler
bot.HandleCommand("start", func(ctx *teleflow.Context) error {
    userRole := getUserRole(ctx.UserID()) // Your logic
    userMenu := createUserMenu(userRole)
    
    // Set menu for this specific chat
    bot.SetMenuButton(ctx.ChatID(), userMenu)
    
    return ctx.Reply("Welcome! Check out the menu button → for quick access to commands.")
})
```

## Integration with Teleflow Features

### Command Registration

The menu button works seamlessly with Teleflow's command system:

```go
// Register commands that appear in menu
bot.HandleCommand("help", helpHandler)
bot.HandleCommand("stats", statsHandler)
bot.HandleCommand("settings", settingsHandler)

// Menu button will show these with friendly names
menuButton := &teleflow.MenuButtonConfig{
    Type: teleflow.MenuButtonTypeCommands,
    Items: []teleflow.MenuButtonItem{
        {Text: "📖 Get Help", Command: "/help"},
        {Text: "📊 View Stats", Command: "/stats"},
        {Text: "⚙️ Bot Settings", Command: "/settings"},
    },
}
```

### Web App Integration

For advanced bots, combine with Web Apps:

```go
// Web app menu button
webAppMenu := &teleflow.MenuButtonConfig{
    Type: teleflow.MenuButtonTypeWebApp,
    Text: "🎮 Play Game",
    WebApp: &teleflow.WebAppInfo{
        URL: "https://your-game.com/telegram-app",
    },
}

bot.WithMenuButton(webAppMenu)

// Handle web app data
bot.HandleWebAppData(func(ctx *teleflow.Context, data *teleflow.WebAppData) error {
    // Process data from web app
    return ctx.Reply(fmt.Sprintf("Received: %s", data.Data))
})
```

## Best Practices

### 1. Use Descriptive Text with Emojis

```go
// ✅ Good - Clear and visually appealing
{Text: "📊 View Statistics", Command: "/stats"}
{Text: "⚙️ Bot Settings", Command: "/settings"}

// ❌ Avoid - Plain and unclear
{Text: "stats", Command: "/stats"}
{Text: "config", Command: "/settings"}
```

### 2. Organize Commands Logically

```go
// ✅ Good - Logical grouping
Items: []teleflow.MenuButtonItem{
    // Main actions first
    {Text: "🏠 Home", Command: "/start"},
    {Text: "📋 Help", Command: "/help"},
    
    // Feature commands
    {Text: "📊 Statistics", Command: "/stats"},
    {Text: "📝 Create Post", Command: "/create"},
    
    // Settings last
    {Text: "⚙️ Settings", Command: "/settings"},
}
```

### 3. Limit Menu Items

Keep menu items to 6-8 maximum for better UX:

```go
// ✅ Good - Focused menu
Items: []teleflow.MenuButtonItem{
    {Text: "🏠 Home", Command: "/start"},
    {Text: "📊 Stats", Command: "/stats"},
    {Text: "📝 New", Command: "/new"},
    {Text: "⚙️ Settings", Command: "/settings"},
}
```

### 4. Update Menu Based on Context

```go
func updateMenuForUserState(bot *teleflow.Bot, chatID int64, userState string) {
    var menu *teleflow.MenuButtonConfig
    
    switch userState {
    case "setup":
        menu = createSetupMenu()
    case "active":
        menu = createActiveUserMenu()
    case "admin":
        menu = createAdminMenu()
    default:
        menu = createDefaultMenu()
    }
    
    bot.SetMenuButton(chatID, menu)
}
```

## Error Handling

```go
// Check for errors when setting menu buttons
if err := bot.SetMenuButton(chatID, menuConfig); err != nil {
    log.Printf("Failed to set menu button: %v", err)
    // Fallback to default menu or notify user
}

// Validate menu configuration
func validateMenuConfig(config *teleflow.MenuButtonConfig) error {
    if config.Type == teleflow.MenuButtonTypeWebApp && config.WebApp == nil {
        return fmt.Errorf("web app config required for web_app menu type")
    }
    
    if config.Type == teleflow.MenuButtonTypeCommands && len(config.Items) == 0 {
        return fmt.Errorf("at least one command item required for commands menu")
    }
    
    return nil
}
```

## Example: Complete Bot with Menu Button

```go
package main

import (
    "log"
    "os"
    
    "github.com/kslamph/teleflow/core"
)

func main() {
    // Create menu button configuration
    menuButton := &teleflow.MenuButtonConfig{
        Type: teleflow.MenuButtonTypeCommands,
        Items: []teleflow.MenuButtonItem{
            {Text: "🏠 Home", Command: "/start"},
            {Text: "📖 Help", Command: "/help"},
            {Text: "📊 Stats", Command: "/stats"},
            {Text: "⚙️ Settings", Command: "/settings"},
        },
    }
    
    // Create bot with menu button
    bot, err := teleflow.NewBot(
        os.Getenv("BOT_TOKEN"),
        teleflow.WithMenuButton(menuButton),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Register command handlers
    bot.HandleCommand("start", func(ctx *teleflow.Context) error {
        return ctx.Reply("🎉 Welcome! Use the menu button → for quick navigation.")
    })
    
    bot.HandleCommand("help", func(ctx *teleflow.Context) error {
        return ctx.Reply("📖 Here's how to use this bot...")
    })
    
    bot.HandleCommand("stats", func(ctx *teleflow.Context) error {
        return ctx.Reply("📊 Your statistics: ...")
    })
    
    bot.HandleCommand("settings", func(ctx *teleflow.Context) error {
        return ctx.Reply("⚙️ Bot settings: ...")
    })
    
    // Start bot (menu button will be automatically set)
    log.Fatal(bot.Start())
}
```

The native menu button enhances user experience by providing quick access to bot functionality without typing commands or navigating through reply keyboards.