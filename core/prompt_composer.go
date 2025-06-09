package teleflow

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type PromptComposer struct {
	botAPI TelegramClient

	messageRenderer *messageHandler

	imageHandler *imageHandler

	keyboardHandler *PromptKeyboardHandler
}

func newPromptComposer(botAPI TelegramClient, msgRenderer *messageHandler, imgHandler *imageHandler, kbdHandler *PromptKeyboardHandler) *PromptComposer {
	return &PromptComposer{
		botAPI:          botAPI,
		messageRenderer: msgRenderer,
		imageHandler:    imgHandler,
		keyboardHandler: kbdHandler,
	}
}

func (pc *PromptComposer) ComposeAndSend(ctx *Context, promptConfig *PromptConfig) error {
	if err := pc.validatePromptConfig(promptConfig); err != nil {
		return fmt.Errorf("invalid PromptConfig: %w", err)
	}

	messageText, parseMode, err := pc.messageRenderer.renderMessage(promptConfig, ctx)
	if err != nil {
		return fmt.Errorf("message rendering failed: %w", err)
	}

	processedImg, err := pc.imageHandler.processImage(promptConfig.Image, ctx)
	if err != nil {
		return fmt.Errorf("image processing failed: %w", err)
	}

	var tgInlineKeyboard *tgbotapi.InlineKeyboardMarkup
	if promptConfig.Keyboard != nil {
		builtKeyboard, err := pc.keyboardHandler.BuildKeyboard(ctx, promptConfig.Keyboard)
		if err != nil {
			return fmt.Errorf("keyboard building failed: %w", err)
		}
		if builtKeyboard != nil {

			if keyboard, ok := builtKeyboard.(tgbotapi.InlineKeyboardMarkup); ok {
				if numButtons(keyboard) > 0 {
					tgInlineKeyboard = &keyboard
				}
			}
		}
	}

	if processedImg != nil {

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
		} else if ctx.pendingReplyKeyboard != nil {
			// Attach pending reply keyboard if no inline keyboard is present
			photoMsg.ReplyMarkup = ctx.pendingReplyKeyboard.ToTgbotapi()
			ctx.pendingReplyKeyboard = nil // Clear after use
		}
		// Log before sending photo message
		logChattable("Sending photo message", photoMsg)
		_, err = pc.botAPI.Send(photoMsg)
		return err
	} else if messageText != "" {

		textMsg := tgbotapi.NewMessage(ctx.ChatID(), messageText)
		if parseMode != ParseModeNone {
			textMsg.ParseMode = string(parseMode)
		}
		if tgInlineKeyboard != nil {
			textMsg.ReplyMarkup = tgInlineKeyboard
		} else if ctx.pendingReplyKeyboard != nil {
			// Attach pending reply keyboard if no inline keyboard is present
			textMsg.ReplyMarkup = ctx.pendingReplyKeyboard.ToTgbotapi()
			ctx.pendingReplyKeyboard = nil // Clear after use
		}
		// Log before sending text message
		logChattable("Sending text message", textMsg)
		_, err = pc.botAPI.Send(textMsg)
		return err
	} else if tgInlineKeyboard != nil {

		invisibleMsg := tgbotapi.NewMessage(ctx.ChatID(), "\u200B") // Zero-width space
		invisibleMsg.ReplyMarkup = tgInlineKeyboard
		// Log before sending invisible message for keyboard
		logChattable("Sending invisible message for keyboard", invisibleMsg)
		_, err = pc.botAPI.Send(invisibleMsg)
		return err
	} else if ctx.pendingReplyKeyboard != nil {
		// Send invisible message with pending reply keyboard if no other content
		invisibleMsg := tgbotapi.NewMessage(ctx.ChatID(), "\u200B")
		invisibleMsg.ReplyMarkup = ctx.pendingReplyKeyboard.ToTgbotapi()
		ctx.pendingReplyKeyboard = nil // Clear after use
		// Log before sending invisible message for pending reply keyboard
		logChattable("Sending invisible message for pending reply keyboard", invisibleMsg)
		_, err = pc.botAPI.Send(invisibleMsg)
		return err
	}

	return nil
}

func (pc *PromptComposer) validatePromptConfig(config *PromptConfig) error {
	if config.Message == nil && config.Image == nil && config.Keyboard == nil {
		return fmt.Errorf("PromptConfig must have at least one of Message, Image, or Keyboard specified")
	}
	return nil
}

// logChattable is a helper function to log tgbotapi.Chattable objects.
func logChattable(description string, chattable tgbotapi.Chattable) {
	jsonData, err := json.MarshalIndent(chattable, "", "  ")
	if err != nil {
		log.Printf("DEBUG: Error marshaling chattable for logging (%s): %v", description, err)
		return
	}
	log.Printf("DEBUG: %s: %s", description, string(jsonData))
}

func numButtons(keyboard tgbotapi.InlineKeyboardMarkup) int {
	count := 0
	for _, row := range keyboard.InlineKeyboard {
		count += len(row)
	}
	return count
}
