# Middleware Demonstration Bot

A comprehensive demonstration bot showcasing all Teleflow middleware types in action. This bot provides interactive examples of middleware functionality including logging, authentication, rate limiting, and panic recovery.

## 🛡️ Middleware Stack

The bot implements a complete middleware stack in the correct order:

1. **RecoveryMiddleware()** - Catches panics and prevents crashes (outermost)
2. **LoggingMiddleware()** - Logs all requests and execution times
3. **RateLimitMiddleware(2)** - Limits requests to 2 per minute per user
4. **AuthMiddleware()** - Checks user authorization (innermost)

## 📋 Features

### Commands
- `/start` - Welcome message explaining middleware demo
- `/admin` - Admin-only command (requires admin role)
- `/spam` - Command to test rate limiting
- `/panic` - Command that triggers panic to test recovery
- `/help` - Help information about middleware features

### Interactive Buttons
- 🏠 **Home** - Return to main menu with status
- ℹ️ **Help** - Quick middleware guide
- ⚡ **Spam Test** - Interactive rate limiting test
- 💥 **Panic Test** - Controlled panic recovery demonstration
- 👑 **Admin Panel** - Admin-only features (when authorized)

## 🔐 Permission System

### Admin Users
Configure admin users via the `ADMIN_USERS` environment variable:
```bash
export ADMIN_USERS="123456789,987654321"
```

### Permission Levels
- **Admin Users**: Can access all commands including `/admin` and admin panel
- **Regular Users**: Can access basic commands but are blocked from admin features

## 🚀 Setup and Usage

### Prerequisites
- Go 1.19+
- Telegram Bot Token

### Environment Variables
```bash
# Required
export TOKEN="your_bot_token_here"

# Optional - Admin users (comma-separated user IDs)
export ADMIN_USERS="123456789,987654321"
```

### Running the Bot
```bash
cd examples/middleware-bot
go run main.go
```

### Getting Your User ID
1. Start the bot with `/start`
2. Your user ID will be displayed in the welcome message
3. Add it to `ADMIN_USERS` environment variable to become an admin

## 🧪 Testing Middleware

### 1. Rate Limiting Test
- Use `/spam` command or "Spam Test" button
- Send multiple requests quickly
- After 2 requests per minute, you'll be rate limited
- Clear feedback shows when you're blocked

### 2. Authorization Test
- Try `/admin` command or "Admin Panel" button
- Non-admin users will be denied access
- Admin users see full admin features
- Clear feedback about permission checks

### 3. Panic Recovery Test
- Use `/panic` command or "Panic Test" button
- Deliberately triggers a panic in the handler
- RecoveryMiddleware catches it gracefully
- Bot continues running normally
- Check server logs for panic details

### 4. Logging Test
- All actions are automatically logged
- Check server console for detailed logs
- Includes timing, user IDs, and status information

## 📊 Example Output

### Rate Limiting in Action
```
⏳ Please wait before sending another message.
```

### Authorization Denial
```
🚫 Access Denied

This command requires admin privileges.
• Your user ID: 123456789
• Required permission: admin_access
• Result: DENIED
```

### Panic Recovery
```
An unexpected error occurred. Please try again.
```

### Server Logs
```
[123456789] Processing command: start
[123456789] Handler completed in 2.3ms
[123456789] Processing command: spam
[123456789] Handler completed in 1.1ms
Panic in handler for user 123456789: DEMONSTRATION PANIC
```

## 🏗️ Implementation Details

### Custom Permission Checker
```go
type AdminPermissionChecker struct {
    adminUsers map[int64]bool
}

func (c *AdminPermissionChecker) CanExecute(userID int64, action string) bool {
    switch action {
    case "admin_access":
        return c.adminUsers[userID]
    case "basic_access":
        return true
    default:
        return false
    }
}
```

### Middleware Order
The middleware stack is carefully ordered:
- Recovery middleware is outermost to catch all panics
- Logging tracks all requests including failed ones
- Rate limiting prevents abuse
- Authorization is innermost for final access control

### Thread Safety
- Rate limiting uses mutex for thread-safe user tracking
- Permission checker uses read-only maps after initialization
- All middleware is designed for concurrent use

## 🎯 Educational Value

This bot demonstrates:
- **Proper middleware stacking and order**
- **Real-world middleware behavior**
- **Custom permission checker implementation**
- **Interactive middleware testing**
- **Clear user feedback about middleware actions**
- **Comprehensive error handling**
- **Production-ready patterns**

## 🔧 Customization

### Adding New Permissions
Extend the permission checker:
```go
case "moderator_access":
    return c.moderatorUsers[userID]
```

### Adjusting Rate Limits
Change the rate limit parameter:
```go
bot.Use(teleflow.RateLimitMiddleware(5)) // 5 requests per minute
```

### Custom Middleware
Add your own middleware to the stack:
```go
bot.Use(YourCustomMiddleware())
```

## 📝 Notes

- All middleware effects are clearly demonstrated to users
- Server logs provide detailed technical information
- Rate limiting uses low threshold (2/min) for easy testing
- Panic recovery includes detailed logging for debugging
- Authorization system supports role-based access control
- Environment variables provide flexible configuration