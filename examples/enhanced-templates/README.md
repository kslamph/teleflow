# Enhanced Template System

The Teleflow framework now includes an enhanced template system with support for different parsing modes, validation, and secure content handling.

## New Features

### Parse Modes

Templates now support different Telegram parsing modes:

- `ParseModeNone` - Plain text (no formatting)
- `ParseModeMarkdown` - Basic Markdown formatting  
- `ParseModeMarkdownV2` - Advanced MarkdownV2 with stricter rules
- `ParseModeHTML` - HTML formatting

### Template Functions

- `escape` - Safely escapes user content based on parse mode
- `safe` - Returns unescaped content (use with caution)
- `title`, `upper`, `lower` - Text transformation functions

### New Methods

- `AddTemplate(name, text, parseMode)` - Add template with parse mode
- `MustAddTemplate(name, text, parseMode)` - Add template or panic (for development)
- `GetTemplateInfo(name)` - Get template metadata
- `ctx.EditOrReplyTemplate(name, data)` - Edit message or reply with template

## Examples

### Plain Text Template
```go
bot.AddTemplate("welcome", `
Welcome {{.Name}}!
Status: {{.Status}}
`, teleflow.ParseModeNone)
```

### Markdown Template
```go
bot.AddTemplate("info", `
*Hello {{.Name | escape}}!*
_You have {{.Count}} messages_
`, teleflow.ParseModeMarkdown)
```

### HTML Template
```go
bot.MustAddTemplate("report", `
<b>Daily Report</b>
<i>Date: {{.Date | escape}}</i>

<u>Stats:</u>
• Users: <code>{{.UserCount}}</code>
• Messages: <code>{{.MessageCount}}</code>
`, teleflow.ParseModeHTML)
```

### Using Templates in Handlers
```go
bot.HandleCommand("/start", func(ctx *teleflow.Context) error {
    data := map[string]interface{}{
        "Name": "User",
        "Status": "Active",
    }
    return ctx.ReplyTemplate("welcome", data)
})
```

### Edit or Reply Pattern
```go
// This will edit the message if it's from a callback, otherwise send new message
bot.RegisterCallback(handler{
    Pattern: "refresh",
    Handler: func(ctx *teleflow.Context) error {
        return ctx.EditOrReplyTemplate("report", updatedData)
    },
})
```

## Security Features

- **Input Validation**: Templates are validated against their parse mode during registration
- **Content Escaping**: The `escape` function automatically escapes user content to prevent injection
- **Template Integrity**: Ensures templates are syntactically correct for their parse mode

## Migration from Old API

The old `AddTemplate(name, text)` method is now `AddTemplate(name, text, parseMode)`. Update your code:

```go
// Old
bot.AddTemplate("welcome", "Hello {{.Name}}")

// New
bot.AddTemplate("welcome", "Hello {{.Name}}", teleflow.ParseModeNone)
```

For development, use `MustAddTemplate` to catch template errors early:

```go
// This will panic if template is invalid (good for development)
bot.MustAddTemplate("welcome", "Hello {{.Name}}", teleflow.ParseModeNone)