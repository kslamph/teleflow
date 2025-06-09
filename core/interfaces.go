package teleflow

// PromptSender defines the interface for composing and sending prompts.
type PromptSender interface {
	// ComposeAndSend composes a prompt based on the given configuration and sends it.
	// It takes a context and a prompt configuration, and returns an error if any occurs.
	ComposeAndSend(ctx *Context, config *PromptConfig) error
}

// MessageCleaner defines the interface for managing messages,
// such as deleting them or editing their reply markup.
type MessageCleaner interface {
	// DeleteMessage deletes a specific message using the context and message ID.
	DeleteMessage(ctx *Context, messageID int) error
	// EditMessageReplyMarkup edits the reply markup of a specific message
	// using the context, message ID, and new reply markup.
	// To remove a keyboard, 'replyMarkup' can be nil.
	EditMessageReplyMarkup(ctx *Context, messageID int, replyMarkup interface{}) error
}

// ContextFlowOperations defines methods for interacting with user flows from the context.
type ContextFlowOperations interface {
	// SetUserFlowData sets flow-specific data for a user.
	setUserFlowData(userID int64, key string, value interface{}) error
	// GetUserFlowData retrieves flow-specific data for a user.
	getUserFlowData(userID int64, key string) (interface{}, bool)
	// StartFlow starts a flow for a user.
	startFlow(userID int64, flowName string, ctx *Context) error
	// IsUserInFlow checks if a user is currently in a flow.
	isUserInFlow(userID int64) bool
	// CancelFlow cancels the current flow for a user.
	cancelFlow(userID int64)
}
