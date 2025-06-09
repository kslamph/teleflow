package teleflow

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(ctx *Context) error

type CommandHandlerFunc func(ctx *Context, command string, args string) error

type TextHandlerFunc func(ctx *Context, text string) error

type DefaultHandlerFunc func(ctx *Context, text string) error

type BotOption func(*Bot)

type PermissionContext struct {
	UserID    int64
	ChatID    int64
	Command   string
	Arguments []string
	IsGroup   bool
	IsChannel bool
	MessageID int
	Update    *tgbotapi.Update
}

type AccessManager interface {
	CheckPermission(ctx *PermissionContext) error

	GetReplyKeyboard(ctx *PermissionContext) *ReplyKeyboard
}

type Bot struct {
	api  TelegramClient
	self tgbotapi.User

	handlers           map[string]HandlerFunc
	textHandlers       map[string]HandlerFunc
	defaultTextHandler HandlerFunc

	flowManager           *flowManager
	promptKeyboardHandler PromptKeyboardActions
	promptComposer        *PromptComposer
	templateManager       TemplateManager

	middleware []MiddlewareFunc

	accessManager AccessManager
	flowConfig    FlowConfig
}

func newBotInternal(client TelegramClient, botUser tgbotapi.User, options ...BotOption) (*Bot, error) {
	b := &Bot{
		api:                   client,
		self:                  botUser,
		handlers:              make(map[string]HandlerFunc),
		textHandlers:          make(map[string]HandlerFunc),
		promptKeyboardHandler: newPromptKeyboardHandler(),
		templateManager:       GetDefaultTemplateManager(),
		middleware:            make([]MiddlewareFunc, 0),
		flowConfig: FlowConfig{
			ExitCommands:        []string{"/cancel"},
			ExitMessage:         "🚫 Operation cancelled.",
			AllowGlobalCommands: false,
			HelpCommands:        []string{"/help"},
			OnProcessAction:     ProcessKeepMessage,
		},
	}

	msgHandler := newMessageHandler(b.templateManager)
	imageHandler := newImageHandler()
	b.promptComposer = newPromptComposer(b.api, msgHandler, imageHandler, b.promptKeyboardHandler.(*PromptKeyboardHandler))

	for _, opt := range options {
		opt(b)
	}

	// Initialize flowManager with its new dependencies
	b.flowManager = newFlowManager(&b.flowConfig, b.promptComposer, b.promptKeyboardHandler, b)
	return b, nil
}

func NewBot(token string, options ...BotOption) (*Bot, error) {
	realAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	botUser, err := realAPI.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	return newBotInternal(realAPI, botUser, options...)
}

func WithFlowConfig(config FlowConfig) BotOption {
	return func(b *Bot) {
		b.flowConfig = config
	}
}

func WithAccessManager(accessManager AccessManager) BotOption {
	return func(b *Bot) {
		b.accessManager = accessManager
		b.UseMiddleware(AuthMiddleware(accessManager))
	}
}

func (b *Bot) UseMiddleware(m MiddlewareFunc) {
	b.middleware = append(b.middleware, m)
}

func (b *Bot) HandleCommand(commandName string, handler CommandHandlerFunc) {

	wrappedHandler := func(ctx *Context) error {

		command := commandName
		args := ""
		if ctx.update.Message != nil && len(ctx.update.Message.Text) > len(command)+1 {
			args = ctx.update.Message.Text[len(command)+1:]
		}
		return handler(ctx, command, args)
	}
	b.handlers[commandName] = b.applyMiddleware(wrappedHandler)
}

func (b *Bot) HandleText(textToMatch string, handler TextHandlerFunc) {

	wrappedHandler := func(ctx *Context) error {
		return handler(ctx, textToMatch)
	}
	b.textHandlers[textToMatch] = b.applyMiddleware(wrappedHandler)
}

func (b *Bot) DefaultHandler(handler DefaultHandlerFunc) {

	wrappedHandler := func(ctx *Context) error {
		var text string
		if ctx.update.Message != nil {
			text = ctx.update.Message.Text
		}
		return handler(ctx, text)
	}
	b.defaultTextHandler = b.applyMiddleware(wrappedHandler)
}

func (b *Bot) RegisterFlow(flow *Flow) {
	b.flowManager.registerFlow(flow)
}

func (b *Bot) GetPromptKeyboardHandler() PromptKeyboardActions {
	return b.promptKeyboardHandler
}

func (b *Bot) applyMiddleware(handler HandlerFunc) HandlerFunc {
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}
	return handler
}

func (b *Bot) processUpdate(update tgbotapi.Update) {
	ctx := newContext(update, b.api, b.templateManager, b.flowManager, b.promptComposer, b.accessManager)
	var err error

	if b.flowManager.isUserInFlow(ctx.UserID()) {
		if ctx.update.Message != nil && b.isGlobalExitCommand(ctx.update.Message.Text) {
			b.flowManager.cancelFlow(ctx.UserID())
			err = ctx.sendSimpleText(b.flowConfig.ExitMessage)
			if err != nil {
				log.Printf("Error sending flow exit message: %v", err)
			}
			return
		} else if b.flowConfig.AllowGlobalCommands && ctx.update.Message != nil && ctx.update.Message.IsCommand() {
			commandName := ctx.update.Message.Command()
			if cmdHandler := b.resolveGlobalCommandHandler(commandName); cmdHandler != nil {

				err = cmdHandler(ctx)
				if err != nil {
					log.Printf("Global command handler error for UserID %d, command '%s': %v", ctx.UserID(), commandName, err)

				}
				return
			}
		}
	}

	if handledByFlow, flowErr := b.flowManager.HandleUpdate(ctx); handledByFlow {
		if flowErr != nil {
			log.Printf("Flow handler error for UserID %d: %v", ctx.UserID(), flowErr)

		}
		return
	}

	if update.Message != nil {
		if update.Message.IsCommand() {
			commandName := update.Message.Command()
			if cmdHandler, ok := b.handlers[commandName]; ok {
				err = cmdHandler(ctx)
			} else {

				if b.defaultTextHandler != nil {
					err = b.defaultTextHandler(ctx)
				}
			}
		} else {
			text := update.Message.Text
			if textHandler, ok := b.textHandlers[text]; ok {
				err = textHandler(ctx)
			} else if b.defaultTextHandler != nil {
				err = b.defaultTextHandler(ctx)
			}

		}
	} else if update.CallbackQuery != nil {

		if answerErr := ctx.answerCallbackQuery(""); answerErr != nil {

			log.Printf("Failed to answer callback query for UserID %d: %v", ctx.UserID(), answerErr)
		}
	}

	if err != nil {
		log.Printf("Handler error for UserID %d: %v", ctx.UserID(), err)
		if replyErr := ctx.sendSimpleText("An error occurred. Please try again."); replyErr != nil {
			log.Printf("Failed to send error reply to UserID %d: %v", ctx.UserID(), replyErr)
		}
	}
}

func (b *Bot) isGlobalExitCommand(text string) bool {
	for _, cmd := range b.flowConfig.ExitCommands {
		if text == cmd {
			return true
		}
	}
	return false
}

func (b *Bot) resolveGlobalCommandHandler(commandName string) HandlerFunc {

	for _, helpCmd := range b.flowConfig.HelpCommands {

		normalizedCmd := "/" + commandName
		if normalizedCmd == helpCmd || commandName == helpCmd {
			if handler, ok := b.handlers[commandName]; ok {
				return handler
			}
		}
	}
	return nil
}

func (b *Bot) SetBotCommands(commands map[string]string) error {
	if b.api == nil {
		return fmt.Errorf("bot API not initialized")
	}
	if len(commands) == 0 {

		clearCmdCfg := tgbotapi.NewSetMyCommands()
		_, err := b.api.Request(clearCmdCfg)
		if err != nil {
			log.Printf("Warning: Failed to clear bot commands: %v", err)
			return fmt.Errorf("failed to clear bot commands: %w", err)
		}
		log.Printf("Bot commands cleared.")
		return nil
	}

	var tgCommands []tgbotapi.BotCommand
	for cmd, desc := range commands {
		tgCommands = append(tgCommands, tgbotapi.BotCommand{Command: cmd, Description: desc})
	}
	cmdCfg := tgbotapi.NewSetMyCommands(tgCommands...)
	_, err := b.api.Request(cmdCfg)
	if err != nil {
		log.Printf("Warning: Failed to set bot commands: %v", err)
		return fmt.Errorf("failed to set bot commands: %w", err)
	}

	return nil
}

// DeleteMessage deletes a specific message using the context and message ID.
func (b *Bot) DeleteMessage(ctx *Context, messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(ctx.ChatID(), messageID)
	_, err := b.api.Send(deleteMsg)
	return err
}

// EditMessageReplyMarkup edits the reply markup of a specific message
// using the context, message ID, and new reply markup.
// To remove a keyboard, 'replyMarkup' can be nil.
func (b *Bot) EditMessageReplyMarkup(ctx *Context, messageID int, replyMarkup interface{}) error {
	var editMsg tgbotapi.EditMessageReplyMarkupConfig

	if replyMarkup == nil {
		// Remove keyboard by setting empty markup
		editMsg = tgbotapi.NewEditMessageReplyMarkup(ctx.ChatID(), messageID, tgbotapi.InlineKeyboardMarkup{})
	} else {
		// Assert to expected type
		keyboard, ok := replyMarkup.(tgbotapi.InlineKeyboardMarkup)
		if !ok {
			return fmt.Errorf("replyMarkup must be of type tgbotapi.InlineKeyboardMarkup")
		}
		editMsg = tgbotapi.NewEditMessageReplyMarkup(ctx.ChatID(), messageID, keyboard)
	}

	_, err := b.api.Send(editMsg)
	return err
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.processUpdate(update)
	}

	return nil
}
