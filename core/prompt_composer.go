// Package teleflow provides a Telegram bot framework for building conversational flows.
// This file contains the PromptComposer component which handles the orchestration
// of message composition and delivery to Telegram, including text rendering,
// image processing, and keyboard attachment.
package teleflow

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotAPI defines the subset of tgbotapi.BotAPI methods needed by PromptComposer.
// This interface abstraction allows for easier testing and mocking of the Telegram Bot API.
//
// The interface includes:
//   - Send: for sending messages, photos, and other content to Telegram
//   - Request: for making API requests like AnswerCallbackQuery that don't return messages
type BotAPI interface {
	// Send transmits a Chattable (message, photo, etc.) to Telegram and returns the sent message.
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	// Request makes an API request to Telegram for operations that return API responses rather than messages.
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// PromptComposer orchestrates the composition and delivery of messages to Telegram
// based on PromptConfig specifications. It coordinates three main components:
//   - Message rendering (text content and templates)
//   - Image processing (various image sources and formats)
//   - Keyboard building (inline keyboards with callback handling)
//
// The composer follows a sequential process:
//  1. Validates the PromptConfig has at least one component (message, image, or keyboard)
//  2. Renders message text and determines parse mode (plain, Markdown, HTML)
//  3. Processes image specifications into sendable format
//  4. Builds inline keyboards with proper callback data mapping
//  5. Determines optimal message type (photo with caption, text message, or keyboard-only)
//  6. Sends the composed message to Telegram via BotAPI
type PromptComposer struct {
	// botAPI provides the interface to Telegram Bot API for sending messages and requests
	botAPI BotAPI
	// messageRenderer handles conversion of MessageSpec into final text with parse mode
	messageRenderer *messageRenderer
	// imageHandler processes ImageSpec from various sources into sendable image data
	imageHandler *imageHandler
	// keyboardHandler builds inline keyboards and manages callback data mappings
	keyboardHandler *PromptKeyboardHandler
}

// NewPromptComposer creates and initializes a new PromptComposer with the required dependencies.
//
// Parameters:
//   - botAPI: Interface to Telegram Bot API for message sending
//   - msgRenderer: Component for processing MessageSpec into rendered text
//   - imgHandler: Component for processing ImageSpec into sendable image data
//   - kbdHandler: Component for building keyboards and managing callback mappings
//
// Returns a fully configured PromptComposer ready to process PromptConfig instances.
func NewPromptComposer(botAPI BotAPI, msgRenderer *messageRenderer, imgHandler *imageHandler, kbdHandler *PromptKeyboardHandler) *PromptComposer {
	return &PromptComposer{
		botAPI:          botAPI,
		messageRenderer: msgRenderer,
		imageHandler:    imgHandler,
		keyboardHandler: kbdHandler,
	}
}

// ComposeAndSend is the main method that processes a PromptConfig and sends the corresponding
// Telegram message. This method orchestrates the entire message composition and delivery process.
//
// The method follows a sequential four-step process:
//  1. Validates the PromptConfig has at least one component (message, image, or keyboard)
//  2. Renders message text from MessageSpec and determines appropriate parse mode
//  3. Processes image from ImageSpec into sendable format (bytes, file path, or URL)
//  4. Builds inline keyboard from KeyboardFunc and maps callback data
//  5. Determines optimal message type and sends to Telegram
//
// Message Type Selection Logic:
//   - If image is present: Send as PhotoConfig (with optional caption and keyboard)
//   - Else if text is present: Send as MessageConfig (with optional keyboard)
//   - Else if keyboard is present: Send as MessageConfig with invisible character
//   - Else: Return nil (should be prevented by validation)
//
// Parameters:
//   - ctx: The conversation context containing user/chat information and data
//   - promptConfig: Configuration specifying message content, image, and keyboard
//
// Returns error if any step fails: validation, rendering, processing, building, or sending.
func (pc *PromptComposer) ComposeAndSend(ctx *Context, promptConfig *PromptConfig) error {
	if err := pc.validatePromptConfig(promptConfig); err != nil {
		return fmt.Errorf("invalid PromptConfig: %w", err)
	}

	// 1. Render Message Text & ParseMode
	messageText, parseMode, err := pc.messageRenderer.renderMessage(promptConfig, ctx)
	if err != nil {
		return fmt.Errorf("message rendering failed: %w", err)
	}

	// 2. Process Image
	processedImg, err := pc.imageHandler.processImage(promptConfig.Image, ctx)
	if err != nil {
		return fmt.Errorf("image processing failed: %w", err)
	}

	// 3. Build Inline Keyboard
	var tgInlineKeyboard *tgbotapi.InlineKeyboardMarkup
	if promptConfig.Keyboard != nil {
		builtKeyboard, err := pc.keyboardHandler.BuildKeyboard(ctx, promptConfig.Keyboard)
		if err != nil {
			return fmt.Errorf("keyboard building failed: %w", err)
		}
		if builtKeyboard != nil {
			// Convert interface{} to tgbotapi.InlineKeyboardMarkup
			if keyboard, ok := builtKeyboard.(tgbotapi.InlineKeyboardMarkup); ok {
				if numButtons(keyboard) > 0 {
					tgInlineKeyboard = &keyboard
				}
			}
		}
	}

	// 4. Determine message type and send
	if processedImg != nil {
		// Send as photo with optional caption and keyboard
		photoMsg := tgbotapi.NewPhoto(ctx.ChatID(), nil)
		if processedImg.data != nil {
			photoMsg.File = tgbotapi.FileBytes{Name: "image.jpg", Bytes: processedImg.data}
		} else if processedImg.filePath != "" {
			if strings.HasPrefix(processedImg.filePath, "http") {
				photoMsg.File = tgbotapi.FileURL(processedImg.filePath)
			} else {
				photoMsg.File = tgbotapi.FilePath(processedImg.filePath)
			}
		} else {
			return fmt.Errorf("processed image has no data or path")
		}

		photoMsg.Caption = messageText
		if parseMode != ParseModeNone {
			photoMsg.ParseMode = string(parseMode)
		}
		if tgInlineKeyboard != nil {
			photoMsg.ReplyMarkup = tgInlineKeyboard
		}
		_, err = pc.botAPI.Send(photoMsg)
		return err
	} else if messageText != "" {
		// Send as text message with optional keyboard
		textMsg := tgbotapi.NewMessage(ctx.ChatID(), messageText)
		if parseMode != ParseModeNone {
			textMsg.ParseMode = string(parseMode)
		}
		if tgInlineKeyboard != nil {
			textMsg.ReplyMarkup = tgInlineKeyboard
		}
		_, err = pc.botAPI.Send(textMsg)
		return err
	} else if tgInlineKeyboard != nil {
		// Send keyboard with an invisible message (zero-width space)
		invisibleMsg := tgbotapi.NewMessage(ctx.ChatID(), "\u200B")
		invisibleMsg.ReplyMarkup = tgInlineKeyboard
		_, err = pc.botAPI.Send(invisibleMsg)
		return err
	}

	// This should be prevented by validatePromptConfig, but included for completeness
	return nil
}

// validatePromptConfig ensures that a PromptConfig has at least one component specified.
// This prevents sending empty messages to Telegram which would result in API errors.
//
// A valid PromptConfig must have at least one of:
//   - Message: text content (string, function, or template reference)
//   - Image: image content (file path, URL, base64, or function)
//   - Keyboard: inline keyboard specification (function returning InlineKeyboardBuilder)
//
// Parameters:
//   - config: The PromptConfig to validate
//
// Returns error if all components (Message, Image, Keyboard) are nil.
func (pc *PromptComposer) validatePromptConfig(config *PromptConfig) error {
	if config.Message == nil && config.Image == nil && config.Keyboard == nil {
		return fmt.Errorf("PromptConfig must have at least one of Message, Image, or Keyboard specified")
	}
	return nil
}

// numButtons counts the total number of buttons in an inline keyboard.
// This helper function is used to determine if a keyboard has any buttons
// before attaching it to a message, preventing empty keyboard markup.
//
// Parameters:
//   - keyboard: The Telegram inline keyboard markup to count buttons in
//
// Returns the total number of buttons across all rows in the keyboard.
func numButtons(keyboard tgbotapi.InlineKeyboardMarkup) int {
	count := 0
	for _, row := range keyboard.InlineKeyboard {
		count += len(row)
	}
	return count
}
