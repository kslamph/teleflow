# Refactoring Bot for Improved Testability

This document outlines a plan to refactor the `Bot` struct and its initialization in `core/bot.go` to enhance unit testability while preserving the simplicity of the public API. The primary goal is to allow injection of a mock Telegram API client.

## Problem Statement

Currently, `NewBot` directly creates a `*tgbotapi.BotAPI` instance, making it difficult to mock API interactions for unit testing `Bot` methods. The existing tests for `Bot.SetBotCommands` in `core/bot_test.go` use a `TestableBot` wrapper that re-implements the method, which is not ideal as it doesn't test the production code directly.

## Proposed Solution

1.  **Define `BotAPIRequester` Interface:**
    Create an interface that specifies the methods from `*tgbotapi.BotAPI` that the `Bot` struct (and its directly initialized components like `PromptComposer`) actually use.

    *Example (`core/bot_api_requester.go` or similar):*
    ```go
    package teleflow

    import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    type BotAPIRequester interface {
        Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
        Send(c tgbotapi.Chattable) (tgbotapi.Message, error) // If Bot.Send is used
        GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
        // Potentially: GetSelf() (*tgbotapi.User, error) if bot.api.Self is used
    }
    ```
    *   `*tgbotapi.BotAPI` implicitly satisfies this.
    *   `TestableBotAPI` from `core/bot_test.go` will be updated to satisfy this.

2.  **Modify `Bot` Struct:**
    Change the `api` field in `Bot` to use this interface.

    *Example (`core/bot.go`):*
    ```go
    type Bot struct {
        api BotAPIRequester // Changed from *tgbotapi.BotAPI
        // ... other fields
    }
    ```

3.  **Adapt `NewBot()` and Add Internal Constructor:**
    *   The public `NewBot(token string, options ...BotOption)` signature **remains unchanged**. It will create the real `*tgbotapi.BotAPI` and then call a new, unexported constructor.
    *   Add `newBotWithRequester(requester BotAPIRequester, options ...BotOption) (*Bot, error)`.

    *Example (`core/bot.go`):*
    ```go
    func NewBot(token string, options ...BotOption) (*Bot, error) {
        realAPI, err := tgbotapi.NewBotAPI(token)
        if err != nil {
            return nil, err
        }
        return newBotWithRequester(realAPI, options...)
    }

    func newBotWithRequester(requester BotAPIRequester, options ...BotOption) (*Bot, error) {
        b := &Bot{
            api:                   requester,
            handlers:              make(map[string]HandlerFunc),
            // ... other initializations ...
            promptKeyboardHandler: newPromptKeyboardHandler(),
        }

        // Adapt PromptComposer initialization
        // Option 1 (preferred): newPromptComposer accepts BotAPIRequester
        // messageRenderer := newMessageRenderer()
        // imageHandler := newImageHandler()
        // b.promptComposer = newPromptComposer(b.api, messageRenderer, imageHandler, b.promptKeyboardHandler)

        for _, opt := range options {
            opt(b)
        }
        b.flowManager.initialize(b) // Or adapt flowManager initialization if it also needs BotAPIRequester directly
        return b, nil
    }
    ```
    *   **Note:** `PromptComposer`'s initialization (`newPromptComposer`) will need to be adapted to accept `BotAPIRequester` if it makes direct API calls, or be designed to use methods on the `Bot` instance which then use `bot.api`.

4.  **Update `TestableBotAPI` and Tests:**
    *   Ensure `TestableBotAPI` in `core/bot_test.go` implements `BotAPIRequester`.
    *   Remove the `TestableBot` wrapper struct and its re-implemented methods.
    *   Tests will call the unexported `newBotWithRequester` (or a test helper) to inject `TestableBotAPI` into a real `Bot` instance.

    *Example (`core/bot_test.go`):*
    ```go
    // ...
    testAPI := NewTestableBotAPI() // Implements BotAPIRequester
    bot, err := newBotWithRequester(testAPI /*, options... */)
    if err != nil {
        t.Fatalf("Test setup: %v", err)
    }
    // Now test bot.SetBotCommands(...) or other Bot methods
    // ...
    ```

## Benefits

*   **Preserves User API:** `NewBot()` remains simple for library users.
*   **True Unit Testing:** Allows testing of the actual `Bot` struct's logic with a mocked API.
*   **Removes Code Duplication:** Eliminates the need for the `TestableBot` wrapper in tests.
*   **Improved Code Clarity:** Clearer separation of API interaction concerns.

## Considerations

*   **Interface Completeness:** The `BotAPIRequester` interface must include all methods from `*tgbotapi.BotAPI` that `Bot` (or its directly initialized components like `PromptComposer`) relies upon.
*   **Component Dependencies:** Components like `PromptComposer` or `flowManager` (if they use `bot.api` directly rather than going through `Bot` methods or `Context`) need to be updated to work with the `BotAPIRequester` interface or receive dependencies in a testable way.