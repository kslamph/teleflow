# Enhanced Template System Demo

This example showcases the powerful enhanced template system in Teleflow with support for multiple parsing modes, automatic content escaping, and advanced formatting capabilities.

## Features Demonstrated

### 🎨 Multiple Parsing Modes
- **Plain Text** (`ParseModeNone`) - No formatting
- **Markdown** (`ParseModeMarkdown`) - Basic Markdown formatting
- **MarkdownV2** (`ParseModeMarkdownV2`) - Advanced Telegram MarkdownV2
- **HTML** (`ParseModeHTML`) - Rich HTML formatting

### 🛡️ Security Features
- **Automatic Content Escaping** - User input is safely escaped based on parse mode
- **Template Validation** - Templates are validated for syntax correctness
- **Injection Prevention** - Prevents formatting injection attacks

### ⚡ Enhanced Functionality
- **EditOrReplyTemplate** - Seamlessly edit messages or send new ones
- **Template Functions** - Built-in functions for text processing
- **Dynamic Data Binding** - Real-time data updates with templates

## Available Commands

### Basic Commands
- `/start` - Welcome message with plain text template
- `/help` - Comprehensive help with HTML formatting

### Parsing Mode Demos
- `/markdown` - Markdown formatting demonstration
- `/markdownv2` - MarkdownV2 advanced formatting
- `/html` - HTML rich formatting showcase

### Interactive Features
- `/profile` - User profile with dynamic data and HTML
- `/status` - System status with refresh button (demonstrates EditOrReplyTemplate)
  - **First click**: Updates message in-place with new data
  - **Subsequent clicks**: Continues to update the same message (no duplicates)
  - **Smart editing**: Only updates when content actually changes

### Text Echo
Send any non-command text to see automatic content escaping in action.

## How to Run

1. **Set Bot Token**:
   ```bash
   # Edit main.go and replace "YOUR_BOT_TOKEN" with your actual bot token
   ```

2. **Build and Run**:
   ```bash
   go build main.go
   ./main
   ```

3. **Start Chatting**:
   - Send `/start` to begin
   - Try different commands to see various parsing modes
   - Send regular text to see content escaping
   - Use the refresh button in `/status` to see EditOrReplyTemplate

## Code Highlights

### Template Registration with Parse Modes
```go
// Plain text template
bot.MustAddTemplate("welcome", `
Welcome {{.Name}}!
Status: {{.Status}}
`, teleflow.ParseModeNone)

// HTML template with escaping
bot.MustAddTemplate("profile", `
<b>User: {{.Name | escape}}</b>
<i>Safe content display</i>
`, teleflow.ParseModeHTML)
```

### Secure Content Handling
```go
data := map[string]interface{}{
    "Name": userInput, // Automatically escaped in template
}
return ctx.ReplyTemplate("profile", data)
```

### EditOrReplyTemplate Usage
```go
// This will edit the message if possible, otherwise send new
return ctx.EditOrReplyTemplate("system_status", updatedData, keyboard)
```

### Template Functions
```go
// In templates:
{{.UserInput | escape}}  // Safe escaping
{{.Text | upper}}        // Uppercase
{{.Message | title}}     // Title case
```

## Template Examples

### Markdown Template
```markdown
**Bold Text** and *Italic Text*
[Links](https://example.com)
`Inline code` snippets
```

### MarkdownV2 Template
```markdown
*Bold*, _italic_, ~strikethrough~
||spoiler text||, __underlined__
[Links](https://example.com)
```

### HTML Template
```html
<b>Bold</b>, <i>italic</i>, <u>underlined</u>
<code>code</code>, <pre>preformatted</pre>
<a href="https://example.com">Links</a>
<blockquote>Quotes</blockquote>
```

## Security Best Practices

1. **Always use `escape` function** for user-provided data:
   ```go
   // ✅ Safe
   "Hello {{.UserName | escape}}"
   
   // ❌ Unsafe
   "Hello {{.UserName}}"
   ```

2. **Use `MustAddTemplate` during development** to catch template errors early:
   ```go
   // Will panic if template is invalid
   bot.MustAddTemplate("name", template, parseMode)
   ```

3. **Validate templates** for their intended parse mode:
   ```go
   // Template validation happens automatically
   err := bot.AddTemplate("test", "<b>{{.Text | escape}}</b>", ParseModeHTML)
   ```

## Advanced Features

### Conditional Logic
```go
{{if .IsPremium}}
⭐ Premium User
{{else}}
Standard User  
{{end}}
```

### Loops and Iteration
```go
{{range .Items}}
• {{.Name | escape}} - {{.Price}}
{{end}}
```

### Custom Data Processing
```go
data := map[string]interface{}{
    "Items": []map[string]interface{}{
        {"Name": "Item 1", "Price": 10.99},
        {"Name": "Item 2", "Price": 15.50},
    },
    "Total": calculateTotal(items),
}
```

This demo provides a comprehensive showcase of Teleflow's enhanced template system, demonstrating both the power and security of the framework's template capabilities.