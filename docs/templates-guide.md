# Teleflow Templates Guide

Teleflow's template system allows you to create dynamic and personalized messages using Go's built-in `text/template` package. This enables you to separate message content and presentation from your bot's logic, making your code cleaner and messages more flexible.

## Table of Contents

- [What are Templates?](#what-are-templates)
- [Adding Templates to Your Bot](#adding-templates-to-your-bot)
  - [`bot.AddTemplate()`](#botaddtemplate)
  - [`bot.MustAddTemplate()`](#botmustaddtemplate)
- [Parse Modes](#parse-modes)
  - [`ParseModeNone`](#parsemodenone)
  - [`ParseModeMarkdown`](#parsemodeMarkdown)
  - [`ParseModeMarkdownV2`](#parsemodeMarkdownv2)
  - [`ParseModeHTML`](#parsemodehtml)
- [Using Templates in Handlers](#using-templates-in-handlers)
  - [`ctx.ReplyTemplate()`](#ctxreplytemplate)
  - [`ctx.EditOrReplyTemplate()`](#ctxeditoreplytemplate)
- [Template Syntax (Go `text/template`)](#template-syntax-go-texttemplate)
  - [Variables (`{{.FieldName}}`)](#variables-fieldname)
  - [Conditionals (`{{if .Condition}}...{{else}}...{{end}}`)](#conditionals-if-conditionelseend)
  - [Loops (`{{range .Items}}...{{end}}`)](#loops-range-itemsend)
  - [Pipelines (`{{.Value | function}}`)](#pipelines-value--function)
- [Built-in Template Functions](#built-in-template-functions)
  - [`escape`](#escape)
  - [`safe`](#safe)
  - [`title`](#title)
  - [`upper`](#upper)
  - [`lower`](#lower)
- [Template Validation](#template-validation)
- [Managing Templates](#managing-templates)
  - [`bot.GetTemplateInfo(name string)`](#botgettemplateinfoname-string)
  - [`bot.GetTemplate(name string)`](#botgettemplatename-string)
  - [`bot.ListTemplates()`](#botlisttemplates)
  - [`bot.HasTemplate(name string)`](#bothastemplatename-string)
- [Example: Welcome Message Template](#example-welcome-message-template)
- [Example: Order Summary Template](#example-order-summary-template)
- [Best Practices for Templates](#best-practices-for-templates)
- [Next Steps](#next-steps)

## What are Templates?
Templates in Teleflow are strings containing text mixed with "actions" – special commands enclosed in `{{` and `}}` – that tell the template engine how to insert dynamic data or control the rendering flow. This is powered by Go's `text/template` package.

Benefits:
- **Personalization**: Tailor messages to individual users (e.g., "Hello {{.UserName}}!").
- **Dynamic Content**: Display lists, conditional information, and formatted data.
- **Separation of Concerns**: Keep message text out of your Go code.
- **Reusability**: Define a template once and use it in multiple places.

## Adding Templates to Your Bot
You register templates with your bot instance, giving each a unique name.

### `bot.AddTemplate()`
`AddTemplate(name string, templateText string, parseMode teleflow.ParseMode) error`
This method registers a template.
```go
import teleflow "github.com/kslamph/teleflow/core"

err := bot.AddTemplate(
    "welcome_message",
    "Hello {{.FirstName}}! Welcome to our bot. You have {{.UnreadCount}} unread messages.",
    teleflow.ParseModeNone, // Or ParseModeMarkdownV2, ParseModeHTML
)
if err != nil {
    log.Fatalf("Failed to add template: %v", err)
}
```

### `bot.MustAddTemplate()`
`MustAddTemplate(name string, templateText string, parseMode teleflow.ParseMode)`
Similar to `AddTemplate`, but panics if an error occurs during template parsing or registration. Useful for templates defined at startup that are critical for the bot.
```go
bot.MustAddTemplate(
    "profile_info",
    "**User Profile**:\nName: `{{.User.FirstName}} {{.User.LastName}}`\nUsername: `@{{.User.UserName}}`",
    teleflow.ParseModeMarkdownV2,
)
```

## Parse Modes
When adding a template, you specify a `ParseMode`. This tells Telegram how to interpret formatting in the *rendered output* of the template. Teleflow also uses this to determine which escaping functions are appropriate within the template.

Supported parse modes (defined in `core/templates.go`):
- `teleflow.ParseModeNone`: Plain text, no special formatting.
- `teleflow.ParseModeMarkdown`: [Markdown (legacy)](https://core.telegram.org/bots/api#markdown-style).
- `teleflow.ParseModeMarkdownV2`: [MarkdownV2](https://core.telegram.org/bots/api#markdownv2-style). Stricter, requires escaping of special characters like `_`, `*`, `[`, `]`, `(`, `)`, `~`, `` ` ``, `>`, `#`, `+`, `-`, `=`, `|`, `{`, `}`, `.`, `!`.
- `teleflow.ParseModeHTML`: [HTML-style](https://core.telegram.org/bots/api#html-style). Supports a subset of HTML tags.

The `escape` template function (see below) behaves differently based on the template's `ParseMode`.

## Using Templates in Handlers
Once templates are added, you can use them to send messages from your handlers.

### `ctx.ReplyTemplate()`
`ReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error`
Renders the specified template with the provided `data` and sends it as a reply.
```go
bot.HandleCommand("welcome", func(ctx *teleflow.Context) error {
    userData := struct {
        FirstName   string
        UnreadCount int
    }{
        FirstName:   ctx.Update.Message.From.FirstName,
        UnreadCount: 5, // Example: get this from a database
    }
    return ctx.ReplyTemplate("welcome_message", userData)
})
```
The `data` argument can be any Go type (struct, map, etc.) that `text/template` can work with. Fields or map keys in `data` become accessible in your template.

### `ctx.EditOrReplyTemplate()`
`EditOrReplyTemplate(templateName string, data interface{}, keyboard ...interface{}) error`
Similar to `ReplyTemplate`, but attempts to edit the current message (if the update is a callback query) or sends a new reply if editing isn't possible.

## Template Syntax (Go `text/template`)
Teleflow uses Go's standard `text/template` syntax. Here are some common elements:

### Variables (`{{.FieldName}}`)
Access fields of the `data` struct/map passed to the template.
- `{{.}}`: The data object itself.
- `{{.FieldName}}`: Access a field named `FieldName`.
- `{{.MapKey}}`: Access a value from a map using `MapKey`.
Example: If `data` is `map[string]interface{}{"Name": "Alice", "Age": 30}`
  - `Hello {{.Name}}, you are {{.Age}}.` -> "Hello Alice, you are 30."

### Conditionals (`{{if .Condition}}...{{else}}...{{end}}`)
Render content conditionally.
```html
{{if .IsAdmin}}
Welcome, Admin!
{{else if .IsMember}}
Hello Member.
{{else}}
Hello Guest.
{{end}}

{{if gt .MessageCount 0}}
You have {{.MessageCount}} new messages.
{{else}}
No new messages.
{{end}}
```
Common comparison functions like `eq`, `ne`, `lt`, `le`, `gt`, `ge` are available.

### Loops (`{{range .Items}}...{{end}}`)
Iterate over slices, arrays, or maps.
```html
Users:
{{range .UserList}}
- Name: {{.Name}}, Email: {{.Email}}
{{else}}
No users found.
{{end}}
```
Inside the `range` block, `{{.}}` refers to the current item.

### Pipelines (`{{.Value | function}}`)
Pass the output of one expression as an argument to a function.
```html
{{.RawText | escape}}
Welcome, {{.Username | title}}!
```

## Built-in Template Functions
Teleflow provides several built-in functions accessible within your templates (see `core/templates.go`):

### `escape`
Escapes a string according to the template's `ParseMode` to prevent unintended formatting or injection issues.
- For `ParseModeMarkdownV2`, it escapes characters like `_`, `*`, `.`, `!`, etc.
- For `ParseModeHTML`, it performs HTML escaping.
- For `ParseModeMarkdown` (legacy), it escapes fewer characters.
```html
Your input was: {{.UserInput | escape}}
```
**This is crucial for rendering user-provided content safely.**

### `safe`
Returns the string as-is, without any escaping. **Use with extreme caution**, only for content you know is already safe for the target `ParseMode`.
```html
This is <b>{{.BoldText | safe}}</b>. <!-- Assuming .BoldText is already valid HTML -->
```

### `title`
Converts a string to title case (e.g., "hello world" -> "Hello World"). Uses `golang.org/x/text/cases`.
```html
{{.Name | title}}
```

### `upper`
Converts a string to uppercase.
```html
{{.Code | upper}}
```

### `lower`
Converts a string to lowercase.
```html
{{.Tag | lower}}
```

## Template Validation
When you add a template, Teleflow performs some basic validation:
1.  **Parse Mode Validation**: Checks if the provided `ParseMode` is supported.
2.  **Template Integrity Validation**:
    - For `ParseModeMarkdown`: Checks for unmatched `*`, `_`, `` ` ``, `[`, `]`.
    - For `ParseModeMarkdownV2`: A simplified check for potentially unescaped special characters. True MarkdownV2 validation is complex.
    - For `ParseModeHTML`: Checks for basic unmatched HTML tags.
    - For `ParseModeNone`: No specific integrity validation.
3.  **Go Template Parsing**: `text/template.Parse()` is called, which catches syntax errors in your template actions.

If validation fails, `bot.AddTemplate()` returns an error, and `bot.MustAddTemplate()` panics.

## Managing Templates

Teleflow provides a few methods to inspect registered templates:

### `bot.GetTemplateInfo(name string) *TemplateInfo`
Retrieves metadata about a template, including its `ParseMode` and the parsed `*template.Template` object.
```go
info := bot.GetTemplateInfo("welcome_message")
if info != nil {
    log.Printf("Template '%s' uses ParseMode: %s", info.Name, info.ParseMode)
}
```

### `bot.GetTemplate(name string) *template.Template`
Retrieves the parsed `*template.Template` object directly. Useful for debugging or advanced use cases.

### `bot.ListTemplates() []string`
Returns a slice of names of all registered templates.

### `bot.HasTemplate(name string) bool`
Checks if a template with the given name exists.

## Example: Welcome Message Template
```go
// In your bot setup:
bot.MustAddTemplate("user_greeting",
    "Hello {{.UserName | title}}! 👋\n" +
    "{{if .IsNewUser}}" +
    "Welcome to our amazing bot! We're glad to have you." +
    "{{else}}" +
    "Welcome back! You have {{.NotificationCount}} new notifications." +
    "{{end}}",
    teleflow.ParseModeMarkdownV2, // Use MarkdownV2 for rich formatting
)

// In a handler (e.g., /start):
bot.HandleCommand("start", func(ctx *teleflow.Context) error {
    data := map[string]interface{}{
        "UserName":          ctx.Update.Message.From.FirstName,
        "IsNewUser":         true, // Replace with actual logic
        "NotificationCount": 0,    // Replace with actual logic
    }
    if !data["IsNewUser"].(bool) { // Example: if not a new user
        data["NotificationCount"] = 5 
    }
    return ctx.ReplyTemplate("user_greeting", data)
})
```

## Example: Order Summary Template
```go
// In bot setup:
bot.MustAddTemplate("order_summary",
    "🛍️ *Order Summary* 🛍️\n\n" +
    "Order ID: `{{.OrderID | escape}}`\n" +
    "Date: {{.OrderDate | escape}}\n\n" +
    "Items:\n" +
    "{{range .Items}}" +
    "• {{.Name | escape}} (x{{.Quantity}}) - ${{printf \"%.2f\" .Price}}\n" +
    "{{else}}" +
    "No items in this order.\n" +
    "{{end}}\n" +
    "{{if .Discount}}" +
    "Discount: -${{printf \"%.2f\" .Discount}}\n" +
    "{{end}}" +
    "Subtotal: ${{printf \"%.2f\" .Subtotal}}\n" +
    "Tax:      ${{printf \"%.2f\" .Tax}}\n" +
    "*Total:   ${{printf \"%.2f\" .Total}}*\n\n" +
    "Status: _{{.Status | escape}}_",
    teleflow.ParseModeMarkdownV2,
)

// In a handler:
// orderData := ... fetch order details ...
// return ctx.ReplyTemplate("order_summary", orderData)
```

## Best Practices for Templates
- **Escape User Input**: Always use the `{{.UserInput | escape}}` pattern when rendering data that originated from users or external sources to prevent formatting issues and potential injection attacks (though Telegram sanitizes heavily).
- **Choose Appropriate ParseMode**: Select `ParseModeMarkdownV2` or `ParseModeHTML` if you need rich formatting. Be aware of their specific escaping rules.
- **Keep Templates Readable**: For complex templates, consider breaking them into smaller, manageable pieces if Go's template system allows (e.g., using `{{template "name"}}` if you define sub-templates, though Teleflow's current `AddTemplate` adds them as distinct named templates).
- **Test Your Templates**: Render templates with various data inputs to ensure they look correct and handle edge cases (e.g., empty lists, missing optional data).
- **Data Structure**: Pass well-structured data (structs or maps with clear field names) to your templates for easier access.

## Next Steps
- [Handlers Guide](handlers-guide.md): Learn where to use `ctx.ReplyTemplate()`.
- [API Reference](api-reference.md): For details on `ParseMode`, `TemplateInfo`, and template-related methods.
- Refer to Go's `text/template` package documentation for the full syntax and capabilities of the templating engine.