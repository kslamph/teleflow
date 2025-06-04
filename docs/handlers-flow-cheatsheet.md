# Teleflow Handlers & Flow System Cheat Sheet

## 📋 Quick Reference Guide

### 🎯 Handler Types Overview

| Handler Type | Purpose | Signature | Registration Method |
|-------------|---------|-----------|-------------------|
| **Command** | Handle slash commands (`/start`, `/help`) | `func(ctx *Context, command string, args string) error` | `bot.HandleCommand("start", handler)` |
| **Text** | Handle specific text messages | `func(ctx *Context, messageText string) error` | `bot.HandleText("Hello", handler)` |
| **Default Text** | Handle any unmatched text | `func(ctx *Context, fullMessageText string) error` | `bot.SetDefaultTextHandler(handler)` |
| **Callback** | Handle inline keyboard button presses | `func(ctx *Context, fullData string, extracted string) error` | `bot.RegisterCallback(SimpleCallback("pattern", handler))` |
| **Flow Step** | Handle flow step interactions | Various signatures (see Flow section) | Defined within flow builder |

---

## 🎮 Handler Implementation Examples

### Command Handlers
```go
// Basic command
bot.HandleCommand("start", func(ctx *teleflow.Context, command string, args string) error {
    return ctx.Reply("Welcome! 🎉")
})

// Command with arguments
bot.HandleCommand("greet", func(ctx *teleflow.Context, command string, args string) error {
    if args == "" {
        return ctx.Reply("Usage: /greet <name>")
    }
    return ctx.Reply(fmt.Sprintf("Hello, %s! 👋", args))
})
```

### Text Handlers
```go
// Specific text (reply keyboard buttons)
bot.HandleText("Show Menu", func(ctx *teleflow.Context, messageText string) error {
    keyboard := createMainMenuKeyboard()
    return ctx.Reply("📋 Here's the menu:", keyboard)
})

// Default handler for unmatched text
bot.SetDefaultTextHandler(func(ctx *teleflow.Context, fullMessageText string) error {
    return ctx.Reply("I don't understand. Try /help")
})
```

### Callback Handlers
```go
// Simple callback
bot.RegisterCallback(teleflow.SimpleCallback("close_menu", 
    func(ctx *teleflow.Context, fullData string, extracted string) error {
        return ctx.EditOrReply("Menu closed ✅")
    }))

// Pattern matching with data extraction
bot.RegisterCallback(teleflow.SimpleCallback("user_*", 
    func(ctx *teleflow.Context, fullData string, extracted string) error {
        userID, _ := strconv.ParseInt(extracted, 10, 64)
        user := getUserByID(userID)
        return ctx.EditOrReply(fmt.Sprintf("Selected user: %s", user.Name))
    }))

// Complex pattern with multiple data points
bot.RegisterCallback(teleflow.SimpleCallback("action_transfer_*_to_*", 
    func(ctx *teleflow.Context, fullData string, extracted string) error {
        parts := strings.Split(extracted, "_to_")
        fromID := parts[0]
        toID := parts[1]
        // Handle transfer logic
        return nil
    }))
```

---

## 🌊 Flow System Cheat Sheet

### Flow Builder API
```go
flow := teleflow.NewFlow("registration").
    Step("name").
        OnStart(func(ctx *teleflow.Context) error {
            return ctx.Reply("What's your name? 📝")
        }).
        OnInput(func(ctx *teleflow.Context, input string) error {
            ctx.Set("user_name", input)
            return nil
        }).
        NextStep("email").
    Step("email").
        OnStart(func(ctx *teleflow.Context) error {
            return ctx.Reply("What's your email? 📧")
        }).
        WithValidator(emailValidator).
        OnInput(func(ctx *teleflow.Context, input string) error {
            // Access validated input if validator provided it
            if validatedEmail, ok := ctx.Get("validated_input"); ok {
                ctx.Set("user_email", validatedEmail)
            }
            return nil
        }).
    OnComplete(func(ctx *teleflow.Context, flowData map[string]interface{}) error {
        name := ctx.Get("user_name").(string)
        email := ctx.Get("user_email").(string)
        return ctx.Reply(fmt.Sprintf("Welcome %s! Email: %s ✅", name, email))
    }).
    OnCancel(func(ctx *teleflow.Context, flowData map[string]interface{}) error {
        return ctx.Reply("Registration cancelled ❌")
    }).
    Build()

bot.RegisterFlow(flow)
```

### Flow Handler Types
| Handler | Purpose | Signature | When Called |
|---------|---------|-----------|-------------|
| **OnStart** | Send prompts, setup step | `func(ctx *Context) error` | User enters step |
| **OnInput** | Process user input | `func(ctx *Context, input string) error` | After validation passes |
| **WithValidator** | Validate input | `func(input string) (bool, string, interface{}, error)` | Before OnInput |
| **OnComplete** | Flow completion | `func(ctx *Context, flowData map[string]interface{}) error` | Flow finishes |
| **OnCancel** | Flow cancellation | `func(ctx *Context, flowData map[string]interface{}) error` | Flow cancelled |

### Flow Data Management
```go
// In any flow handler
ctx.Set("key", value)           // Store data
value := ctx.Get("key")         // Retrieve data
ctx.SetState("key", value)      // Persist beyond flow
```

### Flow Control
```go
// Start a flow
ctx.StartFlow("registration")

// Cancel current flow
ctx.CancelFlow()

// Check if user is in flow
if ctx.IsInFlow() {
    // Handle accordingly
}
```

---

## 🛡️ Middleware System

### General Middleware (Recommended)
```go
// Apply to ALL handler types at once
func loggingMiddleware(next teleflow.HandlerFunc) teleflow.HandlerFunc {
    return func(ctx *teleflow.Context) error {
        log.Printf("Handler started")
        err := next(ctx)
        log.Printf("Handler completed")
        return err
    }
}

bot.Use(loggingMiddleware)  // Applies to commands, text, callbacks, flows

// Multiple general middleware
bot.Use(teleflow.RecoveryMiddleware())
bot.Use(teleflow.LoggingMiddleware())
bot.Use(teleflow.AuthMiddleware(allowedUsers))
```

### Type-Specific Middleware
```go
// Command middleware (when you need command-specific logic)
bot.UseCommandMiddleware(func(next teleflow.CommandHandlerFunc) teleflow.CommandHandlerFunc {
    return func(ctx *teleflow.Context, command string, args string) error {
        log.Printf("Command: /%s with args: %s", command, args)
        return next(ctx, command, args)
    }
})

// Text middleware (when you need text-specific logic)
bot.UseTextMiddleware(func(next teleflow.TextHandlerFunc) teleflow.TextHandlerFunc {
    return func(ctx *teleflow.Context, messageText string) error {
        log.Printf("Text received: %s", messageText)
        return next(ctx, messageText)
    }
})

// Callback middleware (when you need callback-specific logic)
bot.UseCallbackMiddleware(func(next teleflow.CallbackHandlerFunc) teleflow.CallbackHandlerFunc {
    return func(ctx *teleflow.Context, fullData string, extracted string) error {
        log.Printf("Callback: %s -> %s", fullData, extracted)
        return next(ctx, fullData, extracted)
    }
})

// Flow step input middleware
bot.UseFlowStepInputMiddleware(func(next teleflow.FlowStepInputMiddlewareFunc) teleflow.FlowStepInputMiddlewareFunc {
    return func(ctx *teleflow.Context, input string) error {
        log.Printf("Flow input: %s", input)
        return next(ctx, input)
    }
})
```

### Built-in Middleware Examples
```go
// ✅ NEW: Use general middleware (applies everywhere)
bot.Use(teleflow.RecoveryMiddleware())
bot.Use(teleflow.LoggingMiddleware())
bot.Use(teleflow.AuthMiddleware(allowedUsers))

// ⚠️ OLD: Type-specific adapters (still supported)
bot.UseCommandMiddleware(adaptToCommand(teleflow.RecoveryMiddleware()))
bot.UseTextMiddleware(adaptToText(teleflow.LoggingMiddleware()))
```

---

## 🎹 Keyboard Helpers

### Reply Keyboards
```go
keyboard := teleflow.NewReplyKeyboard().
    AddRow("📋 Show Menu", "👥 Users").
    AddRow("⚙️ Settings", "ℹ️ Help").
    SetResizeKeyboard(true).
    SetOneTimeKeyboard(false)

return ctx.Reply("Choose an option:", keyboard)
```

### Inline Keyboards
```go
keyboard := teleflow.NewInlineKeyboard().
    AddButton("User 1", "user_select_1").
    AddButton("User 2", "user_select_2").NewRow().
    AddButton("❌ Close", "close_menu")

return ctx.Reply("Select a user:", keyboard)
```

---

## 📝 Template System

### Template Registration
```go
templates := teleflow.NewTemplateManager()
templates.AddTemplate("user_details", `
**User Details** 👤
Name: {{.User.Name}}
Status: {{if .User.IsActive}}✅ Active{{else}}❌ Inactive{{end}}
Balance: ${{.User.Balance}}
`)

// Use in handler
return ctx.ReplyTemplate("user_details", map[string]interface{}{
    "User": user,
}, keyboard)
```

---

## 🔍 Context Utilities

### Essential Context Methods
```go
// Responses
ctx.Reply("message")                                    // Send message
ctx.ReplyTemplate("template", data)                     // Send templated message
ctx.EditOrReply("message")                             // Edit if callback, reply if message
ctx.EditOrReplyTemplate("template", data, keyboard)    // Edit/reply with template

// User & Chat Info
userID := ctx.UserID()                                 // Get user ID
chatID := ctx.ChatID()                                 // Get chat ID
username := ctx.Username()                             // Get username

// Data Management
ctx.Set("key", value)                                  // Request-scoped data
value := ctx.Get("key")                                // Get request data
ctx.SetState("key", value)                             // Persistent user state
value := ctx.GetState("key")                           // Get persistent state
ctx.ClearState()                                       // Clear all user state

// Flow Control
ctx.StartFlow("flow_name")                             // Start a flow
ctx.CancelFlow()                                       // Cancel current flow
ctx.IsInFlow()                                         // Check if in flow

// Access Control
if !ctx.HasPermission("admin") {                       // Check permission
    return ctx.Reply("Access denied")
}
```

---

## 🚀 Quick Start Patterns

### Simple Bot Structure
```go
func main() {
    bot, _ := teleflow.NewBot(token)
    
    // ✅ NEW: General middleware (applies to all handlers)
    bot.Use(teleflow.RecoveryMiddleware())
    bot.Use(teleflow.LoggingMiddleware())
    bot.Use(authMiddleware)
    
    // Commands
    bot.HandleCommand("start", startHandler)
    bot.HandleCommand("help", helpHandler)
    
    // Text handlers
    bot.HandleText("Menu", menuHandler)
    bot.SetDefaultTextHandler(defaultHandler)
    
    // Callbacks
    bot.RegisterCallback(teleflow.SimpleCallback("action_*", actionHandler))
    
    // Flows
    bot.RegisterFlow(createRegistrationFlow())
    
    // Optional: Type-specific middleware for granular control
    bot.UseCommandMiddleware(commandSpecificMiddleware)
    
    bot.Start()
}
```

### Error Handling Pattern
```go
func handler(ctx *teleflow.Context, data string) error {
    // Always handle errors gracefully
    if err := someOperation(); err != nil {
        log.Printf("Error: %v", err)
        return ctx.Reply("❌ Something went wrong. Please try again.")
    }
    return ctx.Reply("✅ Operation successful!")
}
```

### Service Injection Pattern
```go
// In middleware
func ServiceInjectMiddleware(userService UserService) teleflow.MiddlewareFunc {
    return func(next teleflow.HandlerFunc) teleflow.HandlerFunc {
        return func(ctx *teleflow.Context) error {
            ctx.Set("userService", userService)
            return next(ctx)
        }
    }
}

// In handler
func handler(ctx *teleflow.Context) error {
    userService := ctx.Get("userService").(UserService)
    // Use service...
}
```

---

## ⚡ Performance Tips

1. **Use specific handlers** instead of checking message type manually
2. **Implement proper validation** in flows to prevent bad data
3. **Use middleware** for cross-cutting concerns (auth, logging, etc.)
4. **Cache frequently used data** in context or state
5. **Handle errors gracefully** to prevent bot crashes
6. **Use callback patterns** for complex UI interactions
7. **Leverage templates** for consistent message formatting

---

## 🔗 Quick Links

- [Handlers Guide](handlers-guide.md) - Complete handler documentation
- [Flow Guide](flow-guide.md) - Comprehensive flow system guide  
- [Middleware Guide](middleware-guide.md) - Middleware patterns and examples
- [Keyboards Guide](keyboards-guide.md) - Interactive keyboard creation
- [Templates Guide](templates-guide.md) - Dynamic message templating
- [API Reference](api-reference.md) - Complete API documentation

---

*This cheat sheet covers the essential patterns for building Telegram bots with Teleflow. For complete examples, see the `examples/` directory.*
