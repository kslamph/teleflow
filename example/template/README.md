# MarkdownV2 Template Showcase

This example demonstrates the powerful template system in TeleFlow with comprehensive MarkdownV2 formatting features.

## Features Demonstrated

### 1. **Basic Formatting**
- **Bold text** with `*text*`
- _Italic text_ with `_text_`
- __Underlined text__ with `__text__`
- ~Strikethrough text~ with `~text~`
- Combined formatting like *_bold italic_*

### 2. **Code Examples**
- `Inline code` with backticks
- Multi-line code blocks with syntax highlighting
- Template variables within code examples

### 3. **Links and Mentions**
- [External links](https://example.com) 
- User mentions and inline user links
- Proper escaping for MarkdownV2 compatibility

### 4. **Lists and Organization**
- Numbered lists with template iteration
- Bullet point lists
- Structured content organization

### 5. **Spoilers and Special Formatting**
- ||Hidden spoiler text||
- `Monospace text`
- Mixed formatting combinations

### 6. **Complex Templates**
- User profile with conditional content
- Notification system with priority levels
- Product showcase with dynamic pricing
- Rich data binding and iteration

## Template System Features

### **Template Registration**
```go
bot.AddTemplate("template_name", "Template content with {{.Variables}}", teleflow.ParseModeMarkdownV2)
```

### **Template Usage in Flows**
```go
.Prompt("template:template_name", templateData, keyboardFunc)
```

### **Fluent Inline Keyboard Builder**
```go
func(ctx *teleflow.Context) *teleflow.InlineKeyboardBuilder {
    // Complex callback data - can be any interface{}
    userData := map[string]interface{}{
        "user_id": ctx.UserID(),
        "action":  "profile",
        "timestamp": time.Now(),
    }
    
    return teleflow.NewInlineKeyboard().
        ButtonCallback("📝 Basic", "basic").
        ButtonCallback("💻 Code", "code").
        Row().
        ButtonCallback("🔗 Links", "links").
        ButtonCallback("📋 Lists", "lists").
        Row().
        ButtonCallback("👤 Profile", userData).
        ButtonUrl("🌐 Website", "https://example.com")
}
```

### **Enhanced Callback Data**
```go
// Handle different data types in ProcessFunc
func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
    switch data := buttonClick.Data.(type) {
    case map[string]interface{}:
        userID := data["user_id"].(int64)
        action := data["action"].(string)
        // Handle complex data
    case string:
        // Handle simple string data
    case MyCustomStruct:
        // Handle custom struct data
    }
}
```

### **Convenience Methods**
```go
// Direct template reply
ctx.ReplyTemplate("template_name", templateData)

// Template-based prompts
ctx.SendPromptWithTemplate("template_name", templateData)
```

### **Data Precedence**
- **Context Data**: Automatically available from `ctx.Set(key, value)`
- **Template Data**: Explicit data passed via `TemplateData` field (takes precedence)
- **Clean Merging**: Template data overrides context data for same keys

## Running the Example

1. **Set your bot token:**
   ```bash
   export TELEGRAM_BOT_TOKEN="your_bot_token_here"
   ```

2. **Run the showcase:**
   ```bash
   go run example/template/markdownv2-showcase.go
   ```

3. **Available commands:**
   - `/start` - Start the interactive showcase
   - `/showcase` - Same as /start
   - `/demo_basic` - Show basic formatting example
   - `/demo_code` - Show code examples
   - `/demo_profile` - Show user profile template
   - `/help` - Show help information

## Template Examples

### Basic Formatting Template
```go
bot.AddTemplate("basic_formatting", 
    "*Bold text* and _italic text_ and __underline__ and ~strikethrough~\n\n"+
    "You can also combine *_bold italic_* and *__bold underline__*",
    teleflow.ParseModeMarkdownV2)
```

### Dynamic User Profile Template
```go
bot.AddTemplate("user_profile",
    "*👤 User Profile*\n\n"+
    "*Name:* {{.Name}}\n"+
    "*Role:* `{{.Role}}`\n"+
    "*Status:* {{if .IsActive}}✅ _Active_{{else}}❌ _Inactive_{{end}}\n\n"+
    "*Recent Activity:*\n"+
    "{{range $i, $activity := .Activities}}"+
    "{{.Index}}\\. {{.Activity}}\n"+
    "{{end}}",
    teleflow.ParseModeMarkdownV2)
```

### Usage with Data
```go
ctx.ReplyTemplate("user_profile", map[string]interface{}{
    "Name": "John Doe",
    "Role": "Developer", 
    "IsActive": true,
    "Activities": []map[string]interface{}{
        {"Index": "1", "Activity": "Implemented templates"},
        {"Index": "2", "Activity": "Fixed bugs"},
    },
})
```

## MarkdownV2 Escaping

The template system automatically handles MarkdownV2 escaping requirements:

- **Reserved Characters**: `_*[]()~`>#+-=|{}.!` are properly escaped
- **Template Functions**: Built-in `escape` and `safe` functions available
- **Context-Aware**: Escaping applied based on parse mode

## Interactive Showcase Flow

The example includes a complete interactive flow that demonstrates:

1. **Main Menu**: Template-based menu with inline keyboard
2. **Category Selection**: Different template categories to explore
3. **Live Examples**: Real-time template rendering with sample data
4. **Navigation**: Seamless flow between examples with back buttons
5. **Help System**: Comprehensive command and feature documentation

## Key Benefits

- **Developer Friendly**: Simple, intuitive template syntax
- **Type Safe**: Strong typing with Go's template system
- **Parse Mode Aware**: Automatic parse mode application
- **Data Flexible**: Multiple data sources with clear precedence
- **Performance Optimized**: Templates compiled once at registration
- **Backwards Compatible**: Works alongside existing TeleFlow features

This showcase demonstrates how TeleFlow's template system makes it easy to create rich, dynamic, and maintainable Telegram bot interfaces with advanced MarkdownV2 formatting.