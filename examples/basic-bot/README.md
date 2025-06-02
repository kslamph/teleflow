# Basic Bot Example

This example demonstrates basic Teleflow API usage with a simple bot that responds to commands and keyboard interactions.

## Features

### Commands
- `/start` - Welcome message with reply keyboard
- `/help` - Detailed help information  
- `/ping` - Simple ping response

### Reply Keyboard Buttons
- "🏠 Home" - Shows main menu
- "ℹ️ Info" - Shows bot information
- "❓ Help" - Shows help text

### Demonstrated API Usage
- Basic bot creation with `NewBot()`
- Command handling with `HandleCommand()`
- Text message handling with `HandleText()`
- Logging middleware with `LoggingMiddleware()`
- Reply keyboard creation and usage
- Proper error handling patterns
- Environment variable configuration

## Setup and Usage

### Prerequisites
- Go 1.24 or later
- A Telegram bot token (get one from [@BotFather](https://t.me/BotFather))

### Running the Bot

1. **Set your bot token as an environment variable:**
   ```bash
   export TOKEN="your_bot_token_here"
   ```

2. **Run the bot:**
   ```bash
   go run examples/basic-bot/main.go
   ```

3. **Test the bot:**
   - Start a chat with your bot on Telegram
   - Send `/start` to see the welcome message
   - Use the keyboard buttons or type commands
   - Try `/ping` and `/help` commands

## Code Structure

The example is organized into clear functions:

- `main()` - Bot initialization and startup
- `registerCommands()` - Sets up all command handlers
- `registerTextHandlers()` - Sets up text message routing
- `handleHomeButton()`, `handleInfoButton()`, `handleHelpButton()` - Keyboard button handlers
- `handleUnknownText()` - Handles unrecognized text input
- `createMainKeyboard()` - Creates the reply keyboard

## Learning Points

This example shows:

1. **Bot Lifecycle**: Create → Configure → Start
2. **Middleware Usage**: Adding logging to track all interactions
3. **Handler Registration**: Both commands and text messages
4. **Keyboard Integration**: Creating and using reply keyboards
5. **Error Handling**: Proper error responses and logging
6. **Code Organization**: Clean separation of concerns

## Next Steps

After understanding this basic example, you can explore:

- More complex keyboard layouts
- Inline keyboards with callbacks
- Flow-based conversations
- State management
- Custom middleware
- Template usage for dynamic messages

This bot serves as the foundation for building more sophisticated Telegram bots with the Teleflow framework.