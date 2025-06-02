# Automatic UI Management System

Teleflow provides a powerful automatic UI management system that seamlessly handles both reply keyboards and native menu buttons without requiring developers to manually specify them. The system intelligently applies appropriate user interfaces based on user context, permissions, and roles.

## Overview

The automatic UI management system works through:

- **Enhanced AccessManager Interface** - Provides both reply keyboards and menu buttons based on user context
- **Per-Message Automatic Application** - UI elements are set automatically on every message
- **Context-Aware Decision Making** - Different users see different interfaces based on their permissions
- **Zero Developer Intervention** - Clean, simple API with no manual UI management

## Core Concept

```go
// Developers write clean, simple handlers
bot.HandleCommand("start", func(ctx *teleflow.Context) error {
    return ctx.Reply("Welcome to the bot!")
    // ✨ Bot automatically applies:
    // - Appropriate reply keyboard for this user
    // - Correct menu button for this chat
    // - Based on user permissions and context
})
```

## Enhanced AccessManager Interface

The `AccessManager` interface provides comprehensive context-aware UI management:

```go
type AccessManager interface {
    // Permission checking
    CheckPermission(ctx *PermissionContext) error
    
    // Reply keyboard management
    GetReplyKeyboard(ctx *MenuContext) *ReplyKeyboard
    
    // Menu button management (NEW)
    GetMenuButton(ctx *MenuContext) *MenuButtonConfig
}
```

### MenuContext Structure

```go
type MenuContext struct {
    UserID    int64  // User requesting the interface
    ChatID    int64  // Chat where interface will be shown
    IsGroup   bool   // Whether this is a group chat
    IsChannel bool   // Whether this is a channel
}
```

## Implementation Examples

### Basic Role-Based UI Management

```go
type RoleBasedAccessManager struct {
    userRoles map[int64]string
}

func (am *RoleBasedAccessManager) GetReplyKeyboard(ctx *MenuContext) *ReplyKeyboard {
    role := am.userRoles[ctx.UserID]
    
    switch role {
    case "admin":
        return teleflow.NewReplyKeyboard(
            []teleflow.ReplyKeyboardButton{
                {Text: "👑 Admin Panel"}, {Text: "📊 Analytics"},
            },
            []teleflow.ReplyKeyboardButton{
                {Text: "👥 Users"}, {Text: "⚙️ Settings"},
            },
        )
    case "moderator":
        return teleflow.NewReplyKeyboard(
            []teleflow.ReplyKeyboardButton{
                {Text: "🛡️ Moderation"}, {Text: "📊 Reports"},
            },
            []teleflow.ReplyKeyboardButton{
                {Text: "ℹ️ Help"}, {Text: "⚙️ Settings"},
            },
        )
    default: // regular user
        return teleflow.NewReplyKeyboard(
            []teleflow.ReplyKeyboardButton{
                {Text: "🏠 Home"}, {Text: "ℹ️ Help"},
            },
            []teleflow.ReplyKeyboardButton{
                {Text: "👤 Profile"}, {Text: "⚙️ Settings"},
            },
        )
    }
}

func (am *RoleBasedAccessManager) GetMenuButton(ctx *MenuContext) *MenuButtonConfig {
    role := am.userRoles[ctx.UserID]
    
    switch role {
    case "admin":
        return &teleflow.MenuButtonConfig{
            Type: teleflow.MenuButtonTypeCommands,
            Items: []teleflow.MenuButtonItem{
                {Text: "👑 Admin Panel", Command: "/admin"},
                {Text: "📊 Analytics", Command: "/analytics"},
                {Text: "👥 User Management", Command: "/users"},
                {Text: "🔧 System Config", Command: "/config"},
                {Text: "📝 Logs", Command: "/logs"},
            },
        }
    case "moderator":
        return &teleflow.MenuButtonConfig{
            Type: teleflow.MenuButtonTypeCommands,
            Items: []teleflow.MenuButtonItem{
                {Text: "🛡️ Moderation", Command: "/moderate"},
                {Text: "📊 Reports", Command: "/reports"},
                {Text: "👥 User List", Command: "/userlist"},
                {Text: "ℹ️ Help", Command: "/help"},
            },
        }
    default: // regular user
        return &teleflow.MenuButtonConfig{
            Type: teleflow.MenuButtonTypeCommands,
            Items: []teleflow.MenuButtonItem{
                {Text: "🏠 Home", Command: "/start"},
                {Text: "ℹ️ Help", Command: "/help"},
                {Text: "👤 Profile", Command: "/profile"},
                {Text: "⚙️ Settings", Command: "/settings"},
            },
        }
    }
}
```

### Context-Aware Dynamic UI

```go
type DynamicAccessManager struct {
    userService    UserService
    featureFlags   FeatureFlags
    subscriptions  SubscriptionService
}

func (am *DynamicAccessManager) GetReplyKeyboard(ctx *MenuContext) *ReplyKeyboard {
    user := am.userService.GetUser(ctx.UserID)
    subscription := am.subscriptions.GetSubscription(ctx.UserID)
    
    keyboard := teleflow.NewReplyKeyboard()
    
    // Basic features for all users
    keyboard.AddButton("🏠 Home").AddButton("ℹ️ Help").AddRow()
    
    // Premium features for subscribers
    if subscription.IsPremium() {
        keyboard.AddButton("⭐ Premium Features").AddRow()
    }
    
    // Admin features
    if user.IsAdmin() {
        keyboard.AddButton("👑 Admin").AddRow()
    }
    
    // Feature flag controlled features
    if am.featureFlags.IsEnabled("beta_features", ctx.UserID) {
        keyboard.AddButton("🧪 Beta Features").AddRow()
    }
    
    // Group-specific features
    if ctx.IsGroup {
        keyboard.AddButton("👥 Group Tools").AddRow()
    }
    
    return keyboard.Resize()
}

func (am *DynamicAccessManager) GetMenuButton(ctx *MenuContext) *MenuButtonConfig {
    user := am.userService.GetUser(ctx.UserID)
    
    // Web app for premium users
    if user.IsPremium() {
        return &teleflow.MenuButtonConfig{
            Type: teleflow.MenuButtonTypeWebApp,
            Text: "🚀 Premium Dashboard",
            WebApp: &teleflow.WebAppInfo{
                URL: "https://premium.mybot.com/dashboard",
            },
        }
    }
    
    // Commands menu for regular users
    menu := &teleflow.MenuButtonConfig{
        Type: teleflow.MenuButtonTypeCommands,
        Items: []teleflow.MenuButtonItem{
            {Text: "🏠 Home", Command: "/start"},
            {Text: "ℹ️ Help", Command: "/help"},
        },
    }
    
    // Add context-specific commands
    if ctx.IsGroup {
        menu.Items = append(menu.Items, 
            teleflow.MenuButtonItem{Text: "👥 Group Settings", Command: "/groupsettings"})
    } else {
        menu.Items = append(menu.Items, 
            teleflow.MenuButtonItem{Text: "👤 Profile", Command: "/profile"})
    }
    
    return menu
}
```

## How It Works Internally

### Automatic Application Flow

1. **User sends message** to bot
2. **Handler processes** message using `ctx.Reply()`
3. **Context automatically calls** `applyAutomaticMenuButton()`
4. **AccessManager consulted** for both keyboard and menu button
5. **Reply keyboard applied** to the message
6. **Menu button set** for the chat
7. **Message sent** with complete UI

### Per-Message Processing

```go
// Every ctx.Reply() triggers automatic UI management
func (c *Context) send(text string, keyboard ...interface{}) error {
    // 1. Automatic menu button management
    c.applyAutomaticMenuButton() // Sets menu button for chat
    
    // 2. Message creation
    msg := tgbotapi.NewMessage(c.ChatID(), text)
    
    // 3. Keyboard application (manual override or automatic)
    if len(keyboard) > 0 {
        // Developer provided explicit keyboard
        msg.ReplyMarkup = keyboard[0] 
    } else {
        // Automatic keyboard from AccessManager
        if userMenu := c.Bot.accessManager.GetReplyKeyboard(menuContext); userMenu != nil {
            msg.ReplyMarkup = userMenu.ToTgbotapi()
        }
    }
    
    // 4. Send message with complete UI
    return c.Bot.api.Send(msg)
}
```

## Bot Setup

### Simple Setup

```go
// Create access manager
accessManager := &MyAccessManager{}

// Create bot with automatic UI management
bot, err := teleflow.NewBot(token, 
    teleflow.WithAccessManager(accessManager))

// Register handlers - no keyboard management needed!
bot.HandleCommand("start", func(ctx *teleflow.Context) error {
    return ctx.Reply("Welcome! Your interface is automatically customized.")
})
```

### Advanced Setup with Multiple Features

```go
// Comprehensive access manager
type ComprehensiveAccessManager struct {
    userDB        *UserDatabase
    permissions   *PermissionSystem
    subscriptions *SubscriptionService
    features      *FeatureFlags
}

func (am *ComprehensiveAccessManager) CheckPermission(ctx *teleflow.PermissionContext) error {
    return am.permissions.CheckAccess(ctx.UserID, ctx.Command)
}

func (am *ComprehensiveAccessManager) GetReplyKeyboard(ctx *teleflow.MenuContext) *teleflow.ReplyKeyboard {
    user := am.userDB.GetUser(ctx.UserID)
    return am.buildKeyboardForUser(user, ctx)
}

func (am *ComprehensiveAccessManager) GetMenuButton(ctx *teleflow.MenuContext) *teleflow.MenuButtonConfig {
    user := am.userDB.GetUser(ctx.UserID)
    return am.buildMenuButtonForUser(user, ctx)
}

// Create bot with comprehensive system
accessManager := &ComprehensiveAccessManager{
    userDB:        userDatabase,
    permissions:   permissionSystem,
    subscriptions: subscriptionService,
    features:      featureFlags,
}

bot, err := teleflow.NewBot(token,
    teleflow.WithAccessManager(accessManager))
```

## Benefits

### 1. **Zero UI Management Code**

```go
// ❌ Before: Manual keyboard management
bot.HandleCommand("profile", func(ctx *teleflow.Context) error {
    keyboard := createKeyboardForUser(ctx.UserID()) // Manual
    return ctx.Reply("Your profile", keyboard)      // Manual
})

// ✅ After: Automatic UI management
bot.HandleCommand("profile", func(ctx *teleflow.Context) error {
    return ctx.Reply("Your profile") // Automatic keyboard + menu button
})
```

### 2. **Consistent User Experience**

- Users always see appropriate interfaces for their role
- Reply keyboards and menu buttons always match
- Context-aware interface changes (group vs private chat)
- Automatic updates when permissions change

### 3. **Maintainable Code**

- Single point of UI logic (AccessManager)
- No scattered keyboard code throughout handlers
- Easy to modify UI behavior globally
- Separation of concerns (business logic vs UI)

### 4. **Dynamic and Flexible**

- Interface adapts to user role changes in real-time
- Feature flags can control UI elements
- Subscription status affects available features
- Context-aware interfaces (different for groups/private chats)

## Advanced Patterns

### State-Based Interfaces

```go
func (am *StateAwareAccessManager) GetMenuButton(ctx *MenuContext) *MenuButtonConfig {
    userState := am.stateManager.GetUserState(ctx.UserID)
    
    switch userState {
    case "onboarding":
        return am.createOnboardingMenu()
    case "tutorial":
        return am.createTutorialMenu()
    case "premium_trial":
        return am.createTrialMenu()
    default:
        return am.createDefaultMenu(ctx)
    }
}
```

### Multi-Language Support

```go
func (am *LocalizedAccessManager) GetReplyKeyboard(ctx *MenuContext) *ReplyKeyboard {
    lang := am.userLanguage[ctx.UserID]
    t := am.translator.For(lang)
    
    return teleflow.NewReplyKeyboard(
        []teleflow.ReplyKeyboardButton{
            {Text: t.Translate("button.home")},
            {Text: t.Translate("button.help")},
        },
    )
}
```

### Feature Flag Integration

```go
func (am *FeatureFlagAccessManager) GetMenuButton(ctx *MenuContext) *MenuButtonConfig {
    menu := &teleflow.MenuButtonConfig{
        Type: teleflow.MenuButtonTypeCommands,
        Items: []teleflow.MenuButtonItem{
            {Text: "🏠 Home", Command: "/start"},
        },
    }
    
    // Add features based on flags
    if am.features.IsEnabled("new_dashboard", ctx.UserID) {
        menu.Items = append(menu.Items, 
            teleflow.MenuButtonItem{Text: "📊 New Dashboard", Command: "/dashboard"})
    }
    
    if am.features.IsEnabled("ai_assistant", ctx.UserID) {
        menu.Items = append(menu.Items, 
            teleflow.MenuButtonItem{Text: "🤖 AI Assistant", Command: "/ai"})
    }
    
    return menu
}
```

## Best Practices

### 1. **Keep AccessManager Stateless**

```go
// ✅ Good - Uses external services for state
func (am *AccessManager) GetMenuButton(ctx *MenuContext) *MenuButtonConfig {
    user := am.userService.GetUser(ctx.UserID) // External state
    return am.buildMenuForUser(user)
}

// ❌ Avoid - Internal state management
type AccessManager struct {
    userCache map[int64]*User // Internal state
}
```

### 2. **Handle Errors Gracefully**

```go
func (am *AccessManager) GetReplyKeyboard(ctx *MenuContext) *ReplyKeyboard {
    user, err := am.userService.GetUser(ctx.UserID)
    if err != nil {
        // Return default keyboard on error
        return am.createDefaultKeyboard()
    }
    return am.buildKeyboardForUser(user)
}
```

### 3. **Cache Expensive Operations**

```go
type CachedAccessManager struct {
    cache     map[string]*MenuButtonConfig
    cacheTTL  time.Duration
}

func (am *CachedAccessManager) GetMenuButton(ctx *MenuContext) *MenuButtonConfig {
    cacheKey := fmt.Sprintf("menu:%d:%s", ctx.UserID, am.getUserRole(ctx.UserID))
    
    if cached, exists := am.cache[cacheKey]; exists {
        return cached
    }
    
    menu := am.buildMenuButton(ctx)
    am.cache[cacheKey] = menu
    return menu
}
```

### 4. **Test UI Logic**

```go
func TestAccessManagerMenuButton(t *testing.T) {
    am := &MyAccessManager{}
    
    // Test admin user
    ctx := &teleflow.MenuContext{UserID: 123}
    menu := am.GetMenuButton(ctx)
    
    assert.Equal(t, teleflow.MenuButtonTypeCommands, menu.Type)
    assert.Contains(t, menu.Items, teleflow.MenuButtonItem{
        Text: "👑 Admin", Command: "/admin",
    })
}
```

The automatic UI management system transforms Teleflow into a truly intelligent bot framework where developers focus on business logic while the system handles all user interface concerns automatically and contextually.