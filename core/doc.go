// Package teleflow provides a comprehensive framework for building Telegram bots with advanced flow management capabilities.
//
// Teleflow simplifies the creation of sophisticated Telegram bots by offering:
//   - Multi-step conversation flows with state management
//   - Template-based message rendering with multiple parse modes
//   - Middleware support for logging, authentication, and custom processing
//   - Comprehensive keyboard handling (inline and reply keyboards)
//   - Image processing and attachment capabilities
//   - Error handling strategies for robust bot behavior
//   - Concurrent-safe flow operations
//
// # Quick Start
//
// Creating a basic bot:
//
//	bot, err := teleflow.NewBot("YOUR_BOT_TOKEN")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bot.HandleCommand("start", func(ctx *teleflow.Context, command, args string) error {
//		return ctx.SendPromptText("Welcome to the bot!")
//	})
//
//	bot.Start()
//
// # Flow Management
//
// Teleflow's core strength lies in managing multi-step conversations:
//
//	flow := teleflow.NewFlow("registration").
//		Step("ask_name").
//		Prompt("What's your name?").
//		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//			ctx.SetFlowData("name", input)
//			return teleflow.NextStep()
//		}).
//		Step("ask_age").
//		Prompt("How old are you?").
//		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
//			return teleflow.CompleteFlow()
//		}).
//		Build()
//
//	bot.RegisterFlow(flow)
//
// # Templates
//
// Use templates for dynamic message generation:
//
//	ctx.AddTemplate("welcome", "Hello {{.name}}! You are {{.age}} years old.", teleflow.ParseModeMarkdown)
//	ctx.SendPromptWithTemplate("welcome", map[string]interface{}{
//		"name": "John",
//		"age":  25,
//	})
//
// # Middleware
//
// Add middleware for cross-cutting concerns:
//
//	bot.UseMiddleware(teleflow.LoggingMiddleware())
//	bot.UseMiddleware(teleflow.RecoveryMiddleware())
//	bot.UseMiddleware(teleflow.RateLimitMiddleware(10))
//
// # Error Handling
//
// Configure error handling strategies for flows:
//
//	flow := teleflow.NewFlow("example").
//		OnError(teleflow.OnErrorRetry("Please try again.")).
//		// ... define steps
//		Build()
//
// # Keyboards
//
// Create interactive keyboards:
//
//	keyboard := teleflow.NewPromptKeyboard().
//		ButtonCallback("Option 1", "opt1").
//		ButtonCallback("Option 2", "opt2").
//		Build()
//
// This package is designed to handle complex conversational flows while maintaining
// clean, readable code structure and providing extensive customization options.
package teleflow
