package teleflow

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Context struct {
	bot    *Bot
	update tgbotapi.Update
	data   map[string]interface{}

	userID    int64
	chatID    int64
	isGroup   bool
	isChannel bool
}

func newContext(bot *Bot, update tgbotapi.Update) *Context {
	ctx := &Context{
		bot:    bot,
		update: update,
		data:   make(map[string]interface{}),
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

	return c.bot.flowManager.setUserFlowData(c.UserID(), key, value)
}

func (c *Context) GetFlowData(key string) (interface{}, bool) {
	if !c.isUserInFlow() {
		return nil, false
	}

	return c.bot.flowManager.getUserFlowData(c.UserID(), key)
}

func (c *Context) StartFlow(flowName string) error {

	return c.bot.flowManager.startFlow(c.UserID(), flowName, c)
}

func (c *Context) isUserInFlow() bool {
	return c.bot.flowManager.isUserInFlow(c.UserID())
}

func (c *Context) CancelFlow() {
	c.bot.flowManager.cancelFlow(c.UserID())
}

func (c *Context) SendPrompt(prompt *PromptConfig) error {
	if c.bot.promptComposer == nil {
		return fmt.Errorf("PromptComposer not initialized - this should not happen as initialization is automatic")
	}

	return c.bot.promptComposer.composeAndSend(c, &PromptConfig{
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
	return defaultTemplateManager.AddTemplate(name, templateText, parseMode)
}

func (c *Context) GetTemplateInfo(name string) *TemplateInfo {
	return defaultTemplateManager.GetTemplateInfo(name)
}

func (c *Context) ListTemplates() []string {
	return defaultTemplateManager.ListTemplates()
}

func (c *Context) HasTemplate(name string) bool {
	return defaultTemplateManager.HasTemplate(name)
}

func (c *Context) RenderTemplate(name string, data map[string]interface{}) (string, ParseMode, error) {
	return defaultTemplateManager.RenderTemplate(name, data)
}

func (c *Context) TemplateManager() TemplateManager {
	return defaultTemplateManager
}

func (c *Context) IsGroup() bool {
	return c.isGroup
}

func (c *Context) IsChannel() bool {
	return c.isChannel
}

func (c *Context) getPermissionContext() *PermissionContext {
	if c.bot.accessManager != nil {
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
	_, err := c.bot.api.Request(cb)
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
	_, err := c.bot.api.Send(msg)
	return err
}
