package templates

import (
	teleflow "github.com/kslamph/teleflow/core"
)

// RegisterTemplates registers all message templates with the bot
func RegisterTemplates(bot *teleflow.Bot) {
	// Welcome and navigation templates (HTML formatting)
	bot.MustAddTemplate("welcome", `
🤖 <b>Advanced User Management Bot</b>

Welcome to the comprehensive user management demonstration! This bot showcases:

✨ <b>Advanced Features:</b>
• Multi-step conversation flows
• Dynamic inline keyboards
• Pattern-based callback handling
• Template-driven messages
• Automatic UI management

👥 <b>User Management:</b>
• View and select users
• Change user names with validation
• Toggle user enable/disable status
• Transfer balances between users

🎯 <b>Educational Purpose:</b>
This bot demonstrates the full capabilities of the teleflow framework through practical, real-world user management scenarios.

<i>Use the buttons below to get started!</i>
`, teleflow.ParseModeHTML)

	// Help template
	bot.MustAddTemplate("help", `
❓ <b>Help - Advanced User Management Bot</b>

<b>Available Commands:</b>
/start - Start the bot and show welcome message
/help - Show this help message
/cancel - Cancel current operation

<b>Main Features:</b>

👥 <b>User Manager</b>
Access the user management system to:
• View all users with their status and balance
• Select users to perform actions on them
• Change user names with real-time validation
• Toggle user enable/disable status
• Transfer balances between users

<b>Navigation:</b>
• Use the <b>👥 User Manager</b> button to access user management
• Use the <b>❓ Help</b> button to show this help
• Use keyboard buttons for easy navigation
• Use inline buttons for specific actions

<b>Flows Demonstrated:</b>
• <b>Change Name Flow:</b> Multi-step name change with validation
• <b>Toggle Status Flow:</b> Simple enable/disable functionality
• <b>Transfer Balance Flow:</b> Complex transfer with amount validation

This bot showcases the teleflow framework's capabilities in a practical, educational manner.
`, teleflow.ParseModeHTML)

	// User list template with dynamic rendering
	bot.MustAddTemplate("user_list", `
👥 <b>User Management System</b>

<b>Total Users:</b> {{len .Users}} | <b>Active:</b> {{.ActiveCount}}

{{range .Users}}<b>{{.Name | html}}</b> {{if .Enabled}}✅{{else}}❌{{end}}
💰 Balance: ${{printf "%.2f" .Balance}}
{{end}}

<i>Select a user to manage their account:</i>
`, teleflow.ParseModeHTML)

	// User details template
	bot.MustAddTemplate("user_details", `
👤 <b>User Details</b>

<b>Name:</b> {{.User.Name | html}}
<b>ID:</b> <code>{{.User.ID}}</code>
<b>Status:</b> {{if .User.Enabled}}<b>✅ Enabled</b>{{else}}<b>❌ Disabled</b>{{end}}
<b>Balance:</b> ${{printf "%.2f" .User.Balance}}

<i>Choose an action:</i>
`, teleflow.ParseModeHTML)

	// Flow step templates
	bot.MustAddTemplate("current_name", `
✏️ <b>Change Name</b>

<b>Current name:</b> {{.User.Name | html}}

Please enter the new name:
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("confirm_name_change", `
✏️ <b>Confirm Name Change</b>

<b>New name:</b> {{.NewName | html}}

Are you sure you want to change this user's name?
`, teleflow.ParseModeHTML)

	// Transfer flow templates
	bot.MustAddTemplate("request_transfer_amount", `
💰 <b>Transfer Balance</b>

<b>From:</b> {{.Sender.Name | html}} (${{printf "%.2f" .Sender.Balance}})

Please enter the amount to transfer (maximum: ${{printf "%.2f" .Sender.Balance}}):
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("select_receiver", `
💰 <b>Transfer ${{printf "%.2f" .Amount}}</b>

<b>From:</b> {{.Sender.Name | html}}

Please select the recipient:
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("confirm_transfer", `
💰 <b>Confirm Transfer</b>

<b>From:</b> {{.Sender.Name | html}} (${{printf "%.2f" .Sender.Balance}})
<b>To:</b> {{.Receiver.Name | html}} (${{printf "%.2f" .Receiver.Balance}})
<b>Amount:</b> ${{printf "%.2f" .Amount}}

<b>After transfer:</b>
{{.Sender.Name | html}}: ${{printf "%.2f" .SenderBalanceAfter}}
{{.Receiver.Name | html}}: ${{printf "%.2f" .ReceiverBalanceAfter}}

Proceed with transfer?
`, teleflow.ParseModeHTML)

	// Success and error templates
	bot.MustAddTemplate("transfer_success", `
✅ <b>Transfer Completed</b>

The balance transfer has been processed successfully!

<b>Details:</b>
{{.Sender.Name | html}} → {{.Receiver.Name | html}}
Amount: ${{printf "%.2f" .Amount}}
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("name_change_success", `
✅ <b>Name Updated</b>

The user's name has been changed successfully!
<b>New name:</b> {{.NewName | html}}
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("status_toggle_success", `
✅ <b>Status Updated</b>

{{.User.Name | html}} is now {{if .User.Enabled}}<b>✅ Enabled</b>{{else}}<b>❌ Disabled</b>{{end}}
`, teleflow.ParseModeHTML)

	// Error templates
	bot.MustAddTemplate("error_insufficient_balance", `
❌ <b>Insufficient Balance</b>

{{.User.Name | html}} has insufficient balance for this transfer.
Current balance: ${{printf "%.2f" .User.Balance}}
Required: ${{printf "%.2f" .Amount}}
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("error_invalid_user", `
❌ <b>Invalid User</b>

The selected user could not be found. Please try again.
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("error_user_not_found", `
❌ <b>User Not Found</b>

The requested user does not exist in the system.
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("error_invalid_amount", `
❌ <b>Invalid Amount</b>

Please enter a valid amount greater than 0.
{{if .MaxAmount}}Maximum allowed: ${{printf "%.2f" .MaxAmount}}{{end}}
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("error_same_user_transfer", `
❌ <b>Invalid Transfer</b>

You cannot transfer money to the same user.
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("operation_cancelled", `
❌ <b>Operation Cancelled</b>

The current operation has been cancelled.
`, teleflow.ParseModeHTML)

	bot.MustAddTemplate("menu_closed", `
✅ <b>Menu Closed</b>

Use the buttons below to access features.
`, teleflow.ParseModeHTML)
}
