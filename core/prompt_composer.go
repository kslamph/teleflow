package teleflow

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotAPI defines the subset of tgbotapi.BotAPI methods needed by PromptComposer.
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) // For things like AnswerCallbackQuery
}

// PromptComposer is responsible for taking a PromptConfig, rendering its components,
// and sending the final message(s) to Telegram.
type PromptComposer struct {
	botAPI          BotAPI                 // Interface for tgbotapi.BotAPI for testability
	messageRenderer *messageRenderer       // Changed from MessageRenderer to messageRenderer (lowercase)
	imageHandler    *imageHandler          // Changed from ImageHandler to imageHandler (lowercase)
	keyboardHandler *PromptKeyboardHandler // The renamed one from keyboard_builder.go
}

// NewPromptComposer creates a new PromptComposer.
func NewPromptComposer(botAPI BotAPI, msgRenderer *messageRenderer, imgHandler *imageHandler, kbdHandler *PromptKeyboardHandler) *PromptComposer {
	return &PromptComposer{
		botAPI:          botAPI,
		messageRenderer: msgRenderer,
		imageHandler:    imgHandler,
		keyboardHandler: kbdHandler,
	}
}

// ComposeAndSend processes a PromptConfig and sends the corresponding Telegram message.
func (pc *PromptComposer) ComposeAndSend(ctx *Context, promptConfig *PromptConfig) error {
	if err := pc.validatePromptConfig(promptConfig); err != nil {
		return fmt.Errorf("invalid PromptConfig: %w", err)
	}

	// 1. Render Message Text & ParseMode
	messageText, parseMode, err := pc.messageRenderer.renderMessage(promptConfig, ctx)
	if err != nil {
		// log.Printf("Error rendering message for prompt: %v", err)
		return fmt.Errorf("message rendering failed: %w", err)
	}

	// 2. Process Image
	processedImg, err := pc.imageHandler.processImage(promptConfig.Image, ctx)
	if err != nil {
		// log.Printf("Error processing image for prompt: %v", err)
		return fmt.Errorf("image processing failed: %w", err)
	}

	// 3. Build Inline Keyboard
	var tgInlineKeyboard *tgbotapi.InlineKeyboardMarkup
	if promptConfig.Keyboard != nil {
		builtKeyboard, err := pc.keyboardHandler.BuildKeyboard(ctx, promptConfig.Keyboard)
		if err != nil {
			// log.Printf("Error building keyboard for prompt: %v", err)
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
		// Send as photo
		photoMsg := tgbotapi.NewPhoto(ctx.ChatID(), nil) // FileBytes or FileURL set below
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
		// Send as text message
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
		// Send keyboard with an invisible message
		invisibleMsg := tgbotapi.NewMessage(ctx.ChatID(), "\u200B") // Zero-width space
		invisibleMsg.ReplyMarkup = tgInlineKeyboard
		_, err = pc.botAPI.Send(invisibleMsg)
		return err
	}

	// If nothing to send (should be caught by validatePromptConfig)
	// log.Printf("PromptComposer: Nothing to send for prompt (UID: %d, CID: %d)", ctx.UserID(), ctx.ChatID())
	return nil // Or an error indicating nothing was sent
}

func (pc *PromptComposer) validatePromptConfig(config *PromptConfig) error {
	if config.Message == nil && config.Image == nil && config.Keyboard == nil {
		return fmt.Errorf("PromptConfig must have at least one of Message, Image, or Keyboard specified")
	}
	return nil
}

func numButtons(keyboard tgbotapi.InlineKeyboardMarkup) int {
	count := 0
	for _, row := range keyboard.InlineKeyboard {
		count += len(row)
	}
	return count
}
