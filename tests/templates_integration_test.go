package tests

import (
	"bytes"
	"testing"
	"text/template"
)

func TestReplyTemplateIntegration(t *testing.T) {
	// Create a bot with initialized templates
	bot := &Bot{
		templates: template.New("botMessages"),
	}

	// Add a template
	err := bot.AddTemplate("greeting", "Hello {{.Name}}! You have {{.Count}} messages.", ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	// Test template execution directly (simulating what ReplyTemplate does)
	data := map[string]interface{}{
		"Name":  "John",
		"Count": 5,
	}

	var buf bytes.Buffer
	err = bot.templates.ExecuteTemplate(&buf, "greeting", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := "Hello John! You have 5 messages."
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

func TestTemplateWithComplexData(t *testing.T) {
	bot := &Bot{
		templates: template.New("botMessages"),
	}

	// Add a template with more complex logic
	templateText := `Welcome {{.User.Name}}!
{{if .User.IsAdmin}}You have admin privileges.{{end}}
You have {{len .Messages}} messages:
{{range .Messages}}- {{.}}
{{end}}`

	err := bot.AddTemplate("welcome_complex", templateText, ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to add complex template: %v", err)
	}

	// Test with complex data
	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":    "Alice",
			"IsAdmin": true,
		},
		"Messages": []string{"Hello", "How are you?", "See you later"},
	}

	var buf bytes.Buffer
	err = bot.templates.ExecuteTemplate(&buf, "welcome_complex", data)
	if err != nil {
		t.Fatalf("Failed to execute complex template: %v", err)
	}

	expected := `Welcome Alice!
You have admin privileges.
You have 3 messages:
- Hello
- How are you?
- See you later
`

	if buf.String() != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, buf.String())
	}
}

func TestTemplateErrorHandling(t *testing.T) {
	bot := &Bot{
		templates: template.New("botMessages"),
	}

	// Add a template that references undefined fields
	err := bot.AddTemplate("error_template", "Hello {{.UndefinedField}}!", ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	// Test execution with missing data
	data := map[string]interface{}{
		"Name": "John",
	}

	var buf bytes.Buffer
	err = bot.templates.ExecuteTemplate(&buf, "error_template", data)

	// This should not error in Go templates - undefined fields just render as empty
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	// The result should be "Hello <no value>!" or "Hello !"
	result := buf.String()
	if result != "Hello <no value>!" && result != "Hello !" {
		t.Errorf("Unexpected result for undefined field: %q", result)
	}
}

func TestTemplateOverwrite(t *testing.T) {
	bot := &Bot{
		templates: template.New("botMessages"),
	}

	// Add initial template
	err := bot.AddTemplate("test", "Original: {{.Value}}", ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to add initial template: %v", err)
	}

	// Overwrite with new template
	err = bot.AddTemplate("test", "Updated: {{.Value}}", ParseModeNone)
	if err != nil {
		t.Fatalf("Failed to overwrite template: %v", err)
	}

	// Test that the new template is used
	data := map[string]interface{}{"Value": "hello"}
	var buf bytes.Buffer
	err = bot.templates.ExecuteTemplate(&buf, "test", data)
	if err != nil {
		t.Fatalf("Failed to execute overwritten template: %v", err)
	}

	expected := "Updated: hello"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}
