# Advanced User Management Bot

A comprehensive demonstration of the teleflow framework's capabilities through a practical user management system. This bot showcases advanced features like multi-step flows, pattern-based callbacks, dynamic keyboards, template-driven content, and automatic UI management.

## 🎯 Features Demonstrated

### Core Teleflow Framework Features
- ✅ **Advanced Flow Management** - Multi-step conversations with validation and branching
- ✅ **Pattern-Based Callbacks** - Sophisticated callback handling with data extraction
- ✅ **Template System** - HTML-formatted messages with custom functions
- ✅ **Dynamic Keyboards** - Both inline and reply keyboards with smart generation
- ✅ **Middleware Usage** - Service injection and state management
- ✅ **Automatic UI Management** - Clean interface management and access control
- ✅ **Menu Button Integration** - Command registration with descriptions

### User Management Capabilities
- 👥 **User Listing** - View all users with status and balance information
- ✏️ **Name Changes** - Multi-step flow with validation and confirmation
- 🔄 **Status Toggle** - Enable/disable user accounts with immediate feedback
- 💰 **Balance Transfer** - Complex transfer flow with amount validation and recipient selection
- 🎯 **Role-Based Access** - Permission checking (educational demonstration)

## 🗂️ Project Structure

```
examples/advanced-user-manager/
├── main.go                          # Entry point and bot configuration
├── README.md                        # This documentation
├── .env.example                     # Environment variable template
├── handlers/
│   ├── commands.go                  # Command handlers (/start, /help, /cancel)
│   ├── callbacks.go                 # Callback pattern matching and handlers
│   ├── flows.go                     # Flow definitions and step handlers
│   └── keyboards.go                 # Dynamic keyboard generation utilities
├── models/
│   ├── user.go                      # User data structure and validation
│   └── mock_data.go                 # Sample user dataset (8 users)
├── services/
│   ├── user_service.go              # User CRUD operations and business logic
│   └── access_manager.go            # Role-based UI management and permissions
└── templates/
    ├── templates.go                 # All message templates with HTML formatting
    └── functions.go                 # Custom template functions and helpers
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- A Telegram Bot Token (get one from [@BotFather](https://t.me/BotFather))

### Installation & Setup

1. **Clone the repository and navigate to the example:**
   ```bash
   cd examples/advanced-user-manager
   ```

2. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env and add your BOT_TOKEN
   ```

3. **Install dependencies:**
   ```bash
   go mod tidy
   ```

4. **Run the bot:**
   ```bash
   BOT_TOKEN=your_bot_token_here go run main.go
   ```

   Or with environment file:
   ```bash
   export $(cat .env | xargs) && go run main.go
   ```

## 🎮 How to Use

### Getting Started
1. Start a chat with your bot on Telegram
2. Send `/start` to see the welcome message
3. Use the **👥 User Manager** button to access user management
4. Use the **❓ Help** button to see available commands

### User Management Features

#### 📋 View Users
- Click **👥 User Manager** to see all users
- Each user shows: Name, Status (✅/❌), and Balance
- Select any user to see available actions

#### ✏️ Change User Name
1. Select a user from the list
2. Click **✏️ Change Name**
3. Enter the new name (2-50 characters)
4. Confirm the change
5. See success message and return to list

#### 🔄 Toggle User Status
1. Select a user from the list
2. Click **🔄 Enable/Disable**
3. Status toggles immediately
4. See confirmation and return to list

#### 💰 Transfer Balance
1. Select a user with balance > $0
2. Click **💰 Transfer**
3. Enter transfer amount (validated against balance)
4. Select recipient from available users
5. Confirm transfer details
6. See success message with updated balances

### Navigation
- Use keyboard buttons for main navigation
- Use inline buttons for specific actions
- **⬅️ Back to List** returns to user management
- **❌ Close Menu** returns to main menu
- `/cancel` cancels any active operation

## 🧪 Sample Data

The bot includes 8 sample users with realistic data:

| ID | Name           | Status | Balance |
|----|----------------|--------|---------|
| 1  | Alice Smith    | ✅     | $150.50 |
| 2  | Bob Johnson    | ✅     | $75.25  |
| 3  | Carol Williams | ❌     | $200.00 |
| 4  | Dave Brown     | ✅     | $0.00   |
| 5  | Eve Davis      | ✅     | $325.75 |
| 6  | Frank Wilson   | ❌     | $50.00  |
| 7  | Grace Miller   | ✅     | $500.00 |
| 8  | Henry Taylor   | ✅     | $12.50  |

## 🔧 Technical Implementation

### Flow System
The bot demonstrates three different flow types:

1. **Simple Toggle Flow** - Direct action with immediate result
2. **Validation Flow** - Name change with input validation
3. **Complex Multi-Step Flow** - Transfer with amount input, recipient selection, and confirmation

### Callback Patterns
Advanced pattern matching extracts data from callback data:
- `user_select_*` → Extract user ID
- `action_changename_*` → Extract user ID for name change
- `receiver_*_*` → Extract sender and receiver IDs
- `confirm_*_*` → Extract action type and related data

### Template Features
- **HTML Formatting** - Rich text with bold, italic, code blocks
- **Custom Functions** - Mathematical operations, formatting helpers
- **Dynamic Content** - User data, calculations, conditional display
- **Parse Modes** - HTML mode for rich formatting

### Architecture Benefits
- **Clean Separation** - Models, services, handlers, templates
- **Educational Code** - Well-commented for learning purposes
- **Extensible Design** - Easy to add new features
- **Type Safety** - Proper Go interfaces and error handling

## 📚 Learning Objectives

This example teaches developers how to:

1. **Structure a Bot Application** - Clean architecture patterns
2. **Implement Complex Flows** - Multi-step user interactions
3. **Handle Callbacks Effectively** - Pattern matching and data extraction
4. **Create Dynamic UIs** - Responsive keyboard generation
5. **Use Templates Properly** - Rich content with custom functions
6. **Manage State** - Flow state and context management
7. **Handle Errors Gracefully** - Validation and error messaging
8. **Implement Permissions** - Access control patterns

## 🔍 Code Highlights

### Flow Definition Example
```go
func createChangeNameFlow() *teleflow.Flow {
    return teleflow.NewFlow("change_name").
        Step("show_current").
        OnInput(showCurrentNameHandler).
        Step("request_new_name").
        OnInput(requestNewNameHandler).
        WithValidator(nameValidator()).
        OnComplete(nameChangeCompleteHandler).
        Build()
}
```

### Callback Pattern Matching
```go
bot.RegisterCallback(teleflow.SimpleCallback("user_select_*", func(ctx *teleflow.Context, userID string) error {
    // Extract userID from callback data
    id, _ := strconv.ParseInt(userID, 10, 64)
    user, _ := userService.GetUserByID(id)
    
    return ctx.EditOrReplyTemplate("user_details", map[string]interface{}{
        "User": user,
    }, createUserActionKeyboard(id))
}))
```

### Dynamic Keyboard Generation
```go
func createUserListKeyboard(users []models.User) *teleflow.InlineKeyboardMarkup {
    var keyboard [][]teleflow.InlineKeyboardButton
    
    for _, user := range users {
        button := teleflow.NewInlineKeyboardButton(
            fmt.Sprintf("👤 %s ($%.0f) %s", user.Name, user.Balance, statusIcon(user.Enabled)),
            fmt.Sprintf("user_select_%d", user.ID),
        )
        keyboard = append(keyboard, []teleflow.InlineKeyboardButton{button})
    }
    
    return &teleflow.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}
```

## 🌟 Advanced Features

### Validation System
- **Input Validation** - Name length, character restrictions
- **Business Logic Validation** - Balance checks, user status
- **Error Handling** - Graceful error messages with recovery

### Template Functions
- **Formatting Helpers** - currency, percentage, truncate
- **Conditional Logic** - ifelse, status helpers
- **HTML Safety** - Automatic escaping for user content

### UI Management
- **Automatic Cleanup** - Message editing and replacement
- **Context-Aware Navigation** - Smart back button behavior
- **Responsive Design** - Keyboard layouts adapt to content
- **Permission Integration** - UI elements respect access controls

## 🤝 Contributing

This example is designed to be educational and extensible. Consider these enhancements:

- **Database Integration** - Replace mock data with real persistence
- **User Authentication** - Add real user roles and permissions
- **Audit Logging** - Track all user actions
- **Advanced Validation** - More sophisticated business rules
- **Internationalization** - Multi-language support
- **Admin Panel** - Web interface for user management

## 📄 License

This example is part of the teleflow framework and follows the same license terms.

## 🙋‍♂️ Support

For questions about this example or the teleflow framework:
- Check the [main documentation](../../docs/)
- Review other [examples](../)
- Submit issues for bugs or enhancements

---

**Note**: This is a demonstration bot with mock data. In a production environment, implement proper user authentication, data persistence, and security measures.