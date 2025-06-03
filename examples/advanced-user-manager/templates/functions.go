package templates

import teleflow "github.com/kslamph/teleflow/core"

// RegisterTemplateFunctions registers custom template functions with the bot
// Note: The current teleflow implementation uses built-in template functions
// This function serves as a placeholder for future extensibility
func RegisterTemplateFunctions(bot *teleflow.Bot) {
	// Template functions are currently built-in to teleflow
	// Available functions include:
	// - escape: HTML/Markdown escaping
	// - safe: Mark content as safe (no escaping)
	// - printf: String formatting
	// - And standard Go template functions

	// Future versions of teleflow may support custom template functions
	// This function is provided for forward compatibility
}
