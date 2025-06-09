package main

import (
	"log"

	teleflow "github.com/kslamph/teleflow/core"
)

// registerTemplates registers all message templates
func registerTemplates() {
	templates := map[string]struct {
		text      string
		parseMode teleflow.ParseMode
	}{
		"start_message_template": {
			text: `🤖 *Welcome to the Teleflow Demo Bot!*

This bot showcases various features like multi\-step flows, dynamic keyboards, and template\-based messages\.

✨ *Features:*
• Account management
• Fund transfers
• Mock order placement
• Dynamic image generation
• Click\-to\-copy functionality

Use the menu buttons below to get started\!`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
		"help_message_template": {
			text: `📋 *Available Commands & Menu Options*

*Reply Keyboard:*
• 💼 *Account Info* \- View your accounts or add a new one
• 💸 *Transfer Funds* \- Move funds between your accounts
• 🛒 *Place Order* \- Browse items and simulate an order

*Commands:*
• ` + "`/start`" + ` \- Shows welcome message
• ` + "`/help`" + ` \- Displays this help information
• ` + "`/cancel`" + ` \- Exits any current operation or flow

*Tips:*
• Use the buttons for easy navigation
• Click on IDs and amounts to copy them
• All data is stored temporarily for demo purposes`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
		"account_info": {
			text: `👤 *Account Information*

*User ID:* ` + "`{{.UserID}}`" + `
{{if .Username}}*Username:* ` + "`{{.Username}}`" + `{{end}}
{{if .FirstName}}*Name:* {{.FirstName}}{{if .LastName}} {{.LastName}}{{end}}{{end}}

💼 *Your Accounts:*
{{range .Accounts}}• *{{.Name}}* \(ID: ` + "`{{.AccountID}}`" + `\): *${{printf "%.2f" .Balance}}*
{{end}}`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
		"transfer_success": {
			text: `✅ *Transfer Successful\!*

💸 Transfer of ` + "`${{printf \"%.2f\" .Amount | escape}}`" + ` from account ` + "`{{.FromAccountID|escape}}`" + ` to account ` + "`{{.ToAccountID|escape}}`" + ` completed successfully\.

🧾 *Transaction Details:*
• *From:* {{.FromAccountName | escape}}
• *To:* {{.ToAccountName | escape}}
• *Amount:* ` + "`${{printf \"%.2f\" .Amount | escape}}`" + `
• *Date:* {{.Date | escape}}`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
		"order_summary": {
			text: `🧾 *Order Summary*

*Order ID:* ` + "`{{.OrderID}}`" + `

📦 *Item Details:*
• *Product:* {{.ItemName}}
• *Quantity:* {{.Quantity}}
• *Unit Price:* ` + "`${{printf \"%.2f\" .ItemPrice}}`" + `

🚚 *Shipping:*
• *Method:* {{.ShippingMethod}}
• *Cost:* ` + "`${{printf \"%.2f\" .ShippingCost | escape}}`" + `

💰 *Total Amount:* ` + "`${{printf \"%.2f\" .TotalAmount | escape}}`" + `

📍 *Delivery Address:*
{{.DeliveryAddress}}`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
		"insufficient_funds": {
			text: `❌ *Insufficient Funds*

The selected account does not have enough balance for this transaction\.

💰 *Available Balance:* ` + "`${{printf \"%.2f\" .AvailableBalance}}`" + `
💸 *Required Amount:* ` + "`${{printf \"%.2f\" .RequiredAmount}}`" + `
💳 *Shortfall:* ` + "`${{printf \"%.2f\" .Shortfall}}`" + `

Please select a different account or reduce the amount\.`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
		"account_creation_success": {
			text: `✅ *Account Created Successfully\!*

🆕 *New Account Details:*
• *Name:* {{.AccountName}}
• *Account ID:* ` + "`{{.AccountID}}`" + `
• *Initial Balance:* ` + "`${{printf \"%.2f\" .InitialBalance}}`" + `

Your new account is now ready to use\!`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
		"order_confirmation": {
			text: `🎉 *Order Confirmed\!*

Your order has been successfully placed and payment processed\.

*Order ID:* ` + "`{{.OrderID}}`" + `
*Total Paid:* ` + "`${{printf \"%.2f\" .TotalAmount}}`" + `

📧 You will receive a confirmation email shortly\.
📦 Estimated delivery: 3\-5 business days

Thank you for your purchase\!`,
			parseMode: teleflow.ParseModeMarkdownV2,
		},
	}

	for name, template := range templates {
		err := teleflow.AddTemplate(name, template.text, template.parseMode)
		if err != nil {
			log.Printf("Failed to register template %s: %v", name, err)
		}
	}

	log.Println("All templates registered successfully")
}
