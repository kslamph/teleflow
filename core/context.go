package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Context struct {
	telegramClient  TelegramClient
	templateManager TemplateManager
	flowOps         ContextFlowOperations
	promptSender    PromptSender
	accessManager   AccessManager

	update tgbotapi.Update
	data   map[string]interface{}

	userID    int64
	chatID    int64
	isGroup   bool
	isChannel bool
}

func newContext(
	update tgbotapi.Update,
	client TelegramClient,
	tm TemplateManager,
	fo ContextFlowOperations,
	ps PromptSender,
	am AccessManager,
) *Context {
	ctx := &Context{
		telegramClient:  client,
		templateManager: tm,
		flowOps:         fo,
		promptSender:    ps,
		accessManager:   am,
		update:          update,
		data:            make(map[string]interface{}),
	}

	ctx.userID = ctx.extractUserID(update)
	ctx.chatID = ctx.extractChatID(update)
	ctx.isGroup = update.Message != nil && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup())
	ctx.isChannel = update.Message != nil && update.Message.Chat.IsChannel()

	return ctx
}

func (c *Context) UserID() int64 {
	return c.userID
}

func (c *Context) ChatID() int64 {
	return c.chatID
}

func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

func (c *Context) SetFlowData(key string, value interface{}) error {
	if !c.isUserInFlow() {
		return fmt.Errorf("user not in a flow, cannot set flow data")
	}

	return c.flowOps.setUserFlowData(c.UserID(), key, value)
}

func (c *Context) GetFlowData(key string) (interface{}, bool) {
	if !c.isUserInFlow() {
		return nil, false
	}

	return c.flowOps.getUserFlowData(c.UserID(), key)
}

func (c *Context) StartFlow(flowName string) error {

	return c.flowOps.startFlow(c.UserID(), flowName, c)
}

func (c *Context) isUserInFlow() bool {
	return c.flowOps.isUserInFlow(c.UserID())
}

func (c *Context) CancelFlow() {
	c.flowOps.cancelFlow(c.UserID())
}

func (c *Context) SendPrompt(prompt *PromptConfig) error {
	if c.promptSender == nil {
		return fmt.Errorf("PromptSender not initialized - this should not happen as initialization is automatic")
	}

	return c.promptSender.ComposeAndSend(c, &PromptConfig{
		Message:      prompt.Message,
		Image:        prompt.Image,
		TemplateData: prompt.TemplateData,
	})
}

func (c *Context) SendPromptText(text string) error {
	return c.sendSimpleText(text)
}

func (c *Context) SendPromptWithTemplate(templateName string, data map[string]interface{}) error {
	return c.SendPrompt(&PromptConfig{
		Message:      "template:" + templateName,
		TemplateData: data,
	})
}

// Template management methods - providing access to TemplateManager through Context
func (c *Context) AddTemplate(name, templateText string, parseMode ParseMode) error {
	return c.templateManager.AddTemplate(name, templateText, parseMode)
}

func (c *Context) GetTemplateInfo(name string) *TemplateInfo {
	return c.templateManager.GetTemplateInfo(name)
}

func (c *Context) ListTemplates() []string {
	return c.templateManager.ListTemplates()
}

func (c *Context) HasTemplate(name string) bool {
	return c.templateManager.HasTemplate(name)
}

func (c *Context) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	return c.templateManager.RenderTemplate(name, data)
}

func (c *Context) TemplateManager() TemplateManager {
	return c.templateManager
}

func (c *Context) IsGroup() bool {
	return c.isGroup
}

func (c *Context) IsChannel() bool {
	return c.isChannel
}

func (c *Context) getPermissionContext() *PermissionContext {
	if c.accessManager != nil {
		return &PermissionContext{
			UserID:    c.UserID(),
			ChatID:    c.ChatID(),
			IsGroup:   c.isGroup,
			IsChannel: c.isChannel,
		}

	}
	return nil
}

func (c *Context) extractUserID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

func (c *Context) answerCallbackQuery(text string) error {
	if c.update.CallbackQuery == nil {
		return nil
	}

	cb := tgbotapi.NewCallback(c.update.CallbackQuery.ID, text)
	_, err := c.telegramClient.Request(cb)
	return err
}

func (c *Context) extractChatID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func (c *Context) sendSimpleText(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID(), text)
	msg.DisableWebPagePreview = true
	_, err := c.telegramClient.Send(msg)
	return err
}
