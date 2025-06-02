# Enhanced Template System

The Teleflow framework has been enhanced with a powerful template system that supports multiple parsing modes, content validation, and secure handling of user data.

## Key Features

### 🎯 Parse Mode Support
- **ParseModeNone** - Plain text without formatting
- **ParseModeMarkdown** - Basic Markdown formatting
- **ParseModeMarkdownV2** - Advanced MarkdownV2 with strict escaping
- **ParseModeHTML** - HTML formatting support

### 🛡️ Security & Validation
- **Template Integrity Checking** - Validates templates against their parse mode
- **Content Escaping** - Automatic escaping of user data based on parse mode
- **Syntax Validation** - Ensures templates are syntactically correct

### 🚀 Enhanced API
- **MustAddTemplate** - Development-friendly method that panics on errors
- **EditOrReplyTemplate** - Context method for editing or replying with templates
- **Template Metadata** - Retrieve information about registered templates

## API Reference

### Bot Methods

#### AddTemplate(name, text, parseMode)
```go
func (b *Bot) AddTemplate(name, templateText string, parseMode ParseMode) error
```
Registers a template with validation and parse mode support.

#### MustAddTemplate(name, text, parseMode)
```go
func (b *Bot) MustAddTemplate(name, templateText string, parseMode ParseMode)
```
Registers a template and panics if it fails. Ideal for development.

#### GetTemplateInfo(name)
```go
func (b *Bot) GetTemplateInfo(name string) *TemplateInfo
```
Retrieves metadata about a registered template.

### Context Methods

#### ReplyTemplate(name, data, keyboard...)
```go
func (c *Context) ReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error
```
Sends a reply using a template with the appropriate parse mode.

#### EditOrReplyTemplate(name, data, keyboard...)
```go
func (c *Context) EditOrReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error
```
Attempts to edit the current message using a template, or sends a new message if editing fails.

## Template Functions

The enhanced template system includes built-in functions for safe content handling:

### escape
Automatically escapes content based on the template's parse mode:
```go
// HTML template
<b>Hello {{.Name | escape}}</b>

// Markdown template  
*Hello {{.Name | escape}}*

// MarkdownV2 template
*Hello {{.Name | escape}}*
```

### safe
Returns unescaped content (use with caution):
```go
{{.TrustedHTML | safe}}
```

### Text Transformation
- `title` - Title case
- `upper` - Uppercase
- `lower` - Lowercase

## Examples

### Basic Usage

```go
// Plain text template
bot.AddTemplate("welcome", `
Welcome {{.Name}}!
Your status: {{.Status}}
`, teleflow.ParseModeNone)

// Markdown template with escaping
bot.AddTemplate("info", `
*Hello {{.Name | escape}}!*
You have {{.Count}} new messages.
`, teleflow.ParseModeMarkdown)

// HTML template with formatting
bot.MustAddTemplate("report", `
<b>Daily Report</b>
<i>Date: {{.Date | escape}}</i>

<u>Statistics:</u>
• Users: <code>{{.UserCount}}</code>
• Messages: <code>{{.MessageCount}}</code>
`, teleflow.ParseModeHTML)
```

### Using Templates in Handlers

```go
bot.HandleCommand("/start", func(ctx *teleflow.Context) error {
    data := map[string]interface{}{
        "Name":   ctx.UserID(), // Will be safely escaped
        "Status": "Active",
    }
    return ctx.ReplyTemplate("welcome", data)
})

// Edit or reply pattern for callbacks
bot.RegisterCallback(&teleflow.CallbackHandler{
    Pattern: "refresh_*",
    Handler: func(ctx *teleflow.Context) error {
        data := getUpdatedData()
        // This will edit the message if possible, otherwise send new
        return ctx.EditOrReplyTemplate("report", data)
    },
})
```

### Complex Templates with Loops

```go
bot.MustAddTemplate("user_list", `
📋 <b>Active Users</b>
{{range .Users}}👤 {{.Name | escape}} - <i>{{.Status | escape}}</i>
{{end}}
{{if eq (len .Users) 0}}<i>No users found</i>{{end}}

<b>Total:</b> {{len .Users}} users
`, teleflow.ParseModeHTML)
```

### Conditional Logic

```go
bot.MustAddTemplate("order_status", `
🛒 <b>Order #{{.OrderID | escape}}</b>

{{if eq .Status "pending"}}⏳ <i>Payment pending</i>
{{else if eq .Status "paid"}}✅ <i>Payment confirmed</i>
{{else if eq .Status "shipped"}}🚚 <i>Order shipped</i>
{{else}}❓ <i>Unknown status</i>
{{end}}

<b>Total:</b> ${{printf "%.2f" .Total}}
`, teleflow.ParseModeHTML)
```

## Migration Guide

### From Old API

The template API has been enhanced but maintains backward compatibility concepts:

```go
// Old (not supported)
bot.AddTemplate("welcome", "Hello {{.Name}}")

// New
bot.AddTemplate("welcome", "Hello {{.Name}}", teleflow.ParseModeNone)

// For development
bot.MustAddTemplate("welcome", "Hello {{.Name}}", teleflow.ParseModeNone)
```

### Security Considerations

Always use the `escape` function for user-provided data:

```go
// ❌ Unsafe - could allow injection
bot.AddTemplate("unsafe", "<b>{{.UserInput}}</b>", teleflow.ParseModeHTML)

// ✅ Safe - content is escaped
bot.AddTemplate("safe", "<b>{{.UserInput | escape}}</b>", teleflow.ParseModeHTML)
```

## Error Handling

The enhanced system provides detailed error messages for common issues:

```go
// Template validation errors
err := bot.AddTemplate("invalid", "*unmatched asterisk", teleflow.ParseModeMarkdown)
// Returns: template integrity validation failed for 'invalid': unmatched bold/italic markers in markdown

// Parse mode errors
err = bot.AddTemplate("test", "content", ParseMode("INVALID"))
// Returns: invalid parse mode: unsupported parse mode: INVALID
```

## Testing

Use the test helper for unit tests:

```go
func TestMyTemplates(t *testing.T) {
    bot := &teleflow.Bot{
        // Minimal bot setup for testing
        templates: nil,
    }
    
    err := bot.AddTemplate("test", "Hello {{.Name}}", teleflow.ParseModeNone)
    if err != nil {
        t.Errorf("Template should be valid: %v", err)
    }
}
```

The enhanced template system provides a robust, secure, and flexible foundation for building sophisticated Telegram bots with rich message formatting.