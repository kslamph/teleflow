# Flow Bot Example

This example demonstrates the complete Teleflow flow system with a multi-step money transfer conversation.

## Features

- **Multi-step Transfer Flow**: Amount → Recipient → Confirmation → Complete
- **Input Validation**: Number validation for amounts, choice validation for confirmation
- **Flow Cancellation**: Users can cancel at any step with `/cancel`
- **Complete Middleware Stack**: Logging, Auth, and Recovery middleware
- **Inline Keyboards**: Interactive confirmation buttons
- **Error Handling**: Proper validation and user feedback

## Transfer Flow Implementation

The bot implements a 3-step transfer flow:

1. **Amount Step**: 
   - Prompts user for transfer amount
   - Uses `NumberValidator()` for input validation
   - Only accepts valid numeric input

2. **Recipient Step**:
   - Prompts for recipient username
   - Stores amount from previous step
   - No validation (accepts any text)

3. **Confirmation Step**:
   - Shows transfer summary with inline keyboard
   - Uses `ChoiceValidator([]string{"yes", "no"})` 
   - Processes confirmation or cancellation

## Commands

- `/start` - Welcome message with transfer flow option
- `/transfer` - Start the transfer flow
- `/cancel` - Cancel current flow operation
- `/help` - Show help information

## Keyboard Buttons

- 💸 **Transfer** - Start transfer flow (same as `/transfer`)
- 📊 **Balance** - Show current balance (demo)
- ❓ **Help** - Show help information
- ⚙️ **Settings** - Settings placeholder

## Usage

1. Set environment variable:
   ```bash
   export TOKEN="your_bot_token_here"
   ```

2. Run the bot:
   ```bash
   go run main.go
   ```

3. Start a conversation with your bot and try:
   - Send `/start` to see the welcome message
   - Send `/transfer` or tap "💸 Transfer" to start the flow
   - Follow the prompts: enter amount → enter recipient → confirm
   - Use `/cancel` to cancel the flow at any step

## Code Structure

### Middleware Stack
```go
bot.Use(teleflow.LoggingMiddleware())     // Request logging
bot.Use(teleflow.AuthMiddleware(checker)) // Authorization
bot.Use(teleflow.RecoveryMiddleware())    // Panic recovery
```

### Flow Definition
```go
transferFlow := teleflow.NewFlow("transfer").
    Step("amount").OnInput(handler).WithValidator(teleflow.NumberValidator()).
    Step("recipient").OnInput(handler).
    Step("confirm").OnInput(handler).WithValidator(teleflow.ChoiceValidator([]string{"yes", "no"})).
    OnComplete(completionHandler).
    OnCancel(cancellationHandler).
    Build()
```

### Flow Registration
```go
bot.RegisterFlow(transferFlow)
```

## Key Concepts Demonstrated

1. **Flow Builder DSL**: Fluent interface for building multi-step conversations
2. **Input Validation**: Built-in validators for common input types
3. **Flow State Management**: Automatic state tracking between steps
4. **Inline Keyboards**: Interactive confirmation with callback handling
5. **Middleware Integration**: Complete middleware stack for production use
6. **Error Handling**: Proper validation and user feedback
7. **Flow Cancellation**: User-initiated flow termination

## Flow System Features

- ✅ **Step-by-step progression** with automatic state management
- ✅ **Input validation** with custom validators
- ✅ **Data persistence** between steps using context
- ✅ **Inline keyboards** for interactive confirmations
- ✅ **Flow cancellation** at any step
- ✅ **Completion handlers** for success and cancellation
- ✅ **Middleware integration** for logging, auth, and recovery
- ✅ **Error handling** with user-friendly messages

This example provides a complete demonstration of the Teleflow flow system and serves as a foundation for building complex conversational bots with multi-step workflows.