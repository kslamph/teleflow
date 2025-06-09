package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	teleflow "github.com/kslamph/teleflow/core"
)

// registerFlows registers all the business flows
func registerFlows(bot *teleflow.Bot, businessService *BusinessService) {
	// Register Account Info Flow
	accountInfoFlow, err := createAccountInfoFlow(businessService)
	if err != nil {
		log.Fatal("Failed to create account info flow:", err)
	}
	bot.RegisterFlow(accountInfoFlow)

	// Register Transfer Funds Flow
	transferFundsFlow, err := createTransferFundsFlow(businessService)
	if err != nil {
		log.Fatal("Failed to create transfer funds flow:", err)
	}
	bot.RegisterFlow(transferFundsFlow)

	// Register Place Order Flow
	placeOrderFlow, err := createPlaceOrderFlow(businessService)
	if err != nil {
		log.Fatal("Failed to create place order flow:", err)
	}
	bot.RegisterFlow(placeOrderFlow)

	log.Println("All flows registered successfully")
}

// createAccountInfoFlow creates the account information flow
func createAccountInfoFlow(businessService *BusinessService) (*teleflow.Flow, error) {
	return teleflow.NewFlow("account_info").
		OnError(teleflow.OnErrorCancel("❌ An error occurred in account management.")).
		Step("account_actions").
		Prompt("💼 Account Management - What would you like to do?").
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			return AccountActionsKeyboard()
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				action := buttonClick.Data.(string)
				switch action {
				case "view_accounts":
					return teleflow.NextStep()
				case "add_account":
					return teleflow.GoToStep("add_account_name")
				}
			}
			return teleflow.Retry().WithPrompt("Please select an option from the buttons above.")
		}).
		Step("view_accounts").
		Prompt(func(ctx *teleflow.Context) string {
			accounts := businessService.GetAccounts(ctx.UserID())
			templateData := map[string]interface{}{
				"UserID":   ctx.UserID(),
				"Accounts": accounts,
			}

			// Debug: Check if template exists
			if !teleflow.HasTemplate("account_info") {
				log.Printf("DEBUG: Template 'account_info' NOT FOUND in registry")
				log.Printf("DEBUG: Available templates: %v", teleflow.ListTemplates())
			} else {
				log.Printf("DEBUG: Template 'account_info' EXISTS in registry")
			}

			// Use template: prefix to properly identify this as a template
			ctx.SendPrompt(&teleflow.PromptConfig{
				Message:      "template:account_info",
				TemplateData: templateData,
			})

			return "Account information displayed above. ✅"
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			return teleflow.CompleteFlow()
		}).
		Step("add_account_name").
		Prompt("💳 Enter a name for your new account:").
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if strings.TrimSpace(input) == "" {
				return teleflow.Retry().WithPrompt("Please enter a valid account name:")
			}

			ctx.SetFlowData("new_account_name", input)
			return teleflow.NextStep()
		}).
		Step("add_account_balance").
		Prompt(func(ctx *teleflow.Context) string {
			accountName, _ := ctx.GetFlowData("new_account_name")
			return fmt.Sprintf("💰 Enter initial balance for '%s' account:", accountName)
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			balance, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
			if err != nil || balance < 0 {
				return teleflow.Retry().WithPrompt("Please enter a valid amount (numbers only, e.g., 100.50):")
			}

			accountName, _ := ctx.GetFlowData("new_account_name")
			err = businessService.AddAccount(ctx.UserID(), accountName.(string), balance)
			if err != nil {
				return teleflow.Retry().WithPrompt("Error creating account. Please try again:")
			}

			return teleflow.CompleteFlow()
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			accountName, _ := ctx.GetFlowData("new_account_name")

			// Get the newly created account to show its ID
			accounts := businessService.GetAccounts(ctx.UserID())
			var newAccount *UserAccount
			for _, acc := range accounts {
				if acc.Name == accountName.(string) {
					newAccount = &acc
					break
				}
			}

			if newAccount != nil {
				templateData := map[string]interface{}{
					"AccountName":    newAccount.Name,
					"AccountID":      newAccount.AccountID,
					"InitialBalance": newAccount.Balance,
				}

				return ctx.SendPrompt(&teleflow.PromptConfig{
					Message:      "account_creation_success",
					TemplateData: templateData,
				})
			}

			return ctx.SendPromptText("✅ Account created successfully!")
		}).
		Build()
}

// createTransferFundsFlow creates the fund transfer flow
func createTransferFundsFlow(businessService *BusinessService) (*teleflow.Flow, error) {
	return teleflow.NewFlow("transfer_funds").
		OnError(teleflow.OnErrorCancel("❌ An error occurred during the transfer.")).
		Step("select_from_account").
		Prompt("💸 Select the account to transfer FROM:").
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			accounts := businessService.GetAccounts(ctx.UserID())
			return AccountSelectionKeyboard(accounts, "from_account")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				if data, ok := buttonClick.Data.(map[string]interface{}); ok {
					if accountID, exists := data["account_id"]; exists {
						ctx.SetFlowData("from_account_id", accountID)
						return teleflow.NextStep()
					}
				}
			}
			return teleflow.Retry().WithPrompt("Please select an account from the buttons above.")
		}).
		Step("select_to_account").
		Prompt("💰 Select the account to transfer TO:").
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			accounts := businessService.GetAccounts(ctx.UserID())
			fromAccountID, _ := ctx.GetFlowData("from_account_id")

			// Filter out the from account
			var availableAccounts []UserAccount
			for _, acc := range accounts {
				if acc.AccountID != fromAccountID.(string) {
					availableAccounts = append(availableAccounts, acc)
				}
			}

			return AccountSelectionKeyboard(availableAccounts, "to_account")
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				if data, ok := buttonClick.Data.(map[string]interface{}); ok {
					if accountID, exists := data["account_id"]; exists {
						ctx.SetFlowData("to_account_id", accountID)
						return teleflow.NextStep()
					}
				}
			}
			return teleflow.Retry().WithPrompt("Please select an account from the buttons above.")
		}).
		Step("enter_amount").
		Prompt(func(ctx *teleflow.Context) string {
			fromAccountID, _ := ctx.GetFlowData("from_account_id")
			balance, _ := businessService.GetAccountBalance(ctx.UserID(), fromAccountID.(string))
			return fmt.Sprintf("💵 Enter transfer amount (Available balance: $%.2f):", balance)
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			amount, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
			if err != nil || amount <= 0 {
				return teleflow.Retry().WithPrompt("Please enter a valid amount (numbers only, e.g., 50.00):")
			}

			fromAccountID, _ := ctx.GetFlowData("from_account_id")
			toAccountID, _ := ctx.GetFlowData("to_account_id")

			// Perform the transfer
			err = businessService.TransferFunds(ctx.UserID(), fromAccountID.(string), toAccountID.(string), amount)
			if err != nil {
				if strings.Contains(err.Error(), "insufficient funds") {
					balance, _ := businessService.GetAccountBalance(ctx.UserID(), fromAccountID.(string))
					templateData := map[string]interface{}{
						"AvailableBalance": balance,
						"RequiredAmount":   amount,
						"Shortfall":        amount - balance,
					}

					ctx.SendPrompt(&teleflow.PromptConfig{
						Message:      "template:insufficient_funds",
						TemplateData: templateData,
					})
					return teleflow.Retry().WithPrompt("Please enter a smaller amount:")
				}
				return teleflow.Retry().WithPrompt("Transfer failed. Please try again:")
			}

			ctx.SetFlowData("transfer_amount", amount)
			return teleflow.CompleteFlow()
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			fromAccountID, _ := ctx.GetFlowData("from_account_id")
			toAccountID, _ := ctx.GetFlowData("to_account_id")
			amount, _ := ctx.GetFlowData("transfer_amount")

			// Get account names
			accounts := businessService.GetAccounts(ctx.UserID())
			var fromAccountName, toAccountName string
			for _, acc := range accounts {
				if acc.AccountID == fromAccountID.(string) {
					fromAccountName = acc.Name
				}
				if acc.AccountID == toAccountID.(string) {
					toAccountName = acc.Name
				}
			}

			templateData := map[string]interface{}{
				"Amount":          amount,
				"FromAccountID":   fromAccountID,
				"ToAccountID":     toAccountID,
				"FromAccountName": fromAccountName,
				"ToAccountName":   toAccountName,
				"Date":            time.Now().Format("2006-01-02 15:04:05"),
			}

			return ctx.SendPrompt(&teleflow.PromptConfig{
				Message:      "template:transfer_success",
				TemplateData: templateData,
			})
		}).
		Build()
}

// createPlaceOrderFlow creates the order placement flow
func createPlaceOrderFlow(businessService *BusinessService) (*teleflow.Flow, error) {
	return teleflow.NewFlow("place_order").
		OnError(teleflow.OnErrorCancel("❌ An error occurred while placing your order.")).
		Step("select_category").
		Prompt("🛒 Welcome to our store! Select a product category:").
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			categories := []string{"📱 Electronics", "👕 Clothing", "📚 Books", "🏠 Home & Garden"}
			return CategorySelectionKeyboard(categories)
		}).
		WithImage(func(ctx *teleflow.Context) []byte {
			imageBytes, err := GeneratePromoImage("Tech Gadgets", 600, 200)
			if err != nil {
				log.Printf("Failed to generate promo image: %v", err)
				return nil
			}
			return imageBytes
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				category := buttonClick.Data.(string)
				ctx.SetFlowData("selected_category", category)
				return teleflow.NextStep()
			}
			return teleflow.Retry().WithPrompt("Please select a category from the buttons above.")
		}).
		Step("select_merchandise").
		Prompt(func(ctx *teleflow.Context) string {
			category, _ := ctx.GetFlowData("selected_category")
			return fmt.Sprintf("📦 Select an item from %s:", category)
		}).
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			category, _ := ctx.GetFlowData("selected_category")
			items := getMerchandiseForCategory(category.(string))
			return MerchandiseSelectionKeyboard(items)
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				itemID := buttonClick.Data.(string)
				ctx.SetFlowData("selected_item_id", itemID)
				return teleflow.NextStep()
			}
			return teleflow.Retry().WithPrompt("Please select an item from the buttons above.")
		}).
		Step("enter_quantity").
		Prompt("🔢 Enter the quantity you want to order:").
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			quantity, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || quantity <= 0 {
				return teleflow.Retry().WithPrompt("Please enter a valid quantity (positive number):")
			}

			ctx.SetFlowData("quantity", quantity)
			return teleflow.NextStep()
		}).
		Step("enter_address").
		Prompt("📍 Please type your full delivery address:").
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if strings.TrimSpace(input) == "" {
				return teleflow.Retry().WithPrompt("Please enter a valid delivery address:")
			}

			ctx.SetFlowData("delivery_address", input)
			return teleflow.NextStep()
		}).
		Step("confirm_location").
		Prompt(func(ctx *teleflow.Context) string {
			address, _ := ctx.GetFlowData("delivery_address")
			return fmt.Sprintf("📍 Confirm delivery address: %s", address)
		}).
		WithImage(func(ctx *teleflow.Context) []byte {
			address, _ := ctx.GetFlowData("delivery_address")
			imageBytes, err := GenerateMapImage(address.(string), 600, 400)
			if err != nil {
				log.Printf("Failed to generate map image: %v", err)
				return nil
			}
			return imageBytes
		}).
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			return LocationConfirmationKeyboard()
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				action := buttonClick.Data.(string)
				switch action {
				case "confirm_location":
					return teleflow.NextStep()
				case "reenter_address":
					return teleflow.GoToStep("enter_address")
				}
			}
			return teleflow.Retry().WithPrompt("Please choose an option from the buttons above.")
		}).
		Step("select_shipping").
		Prompt("🚚 Select shipping method:").
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			shippingOptions := getShippingOptions()
			return ShippingSelectionKeyboard(shippingOptions)
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				shippingID := buttonClick.Data.(string)
				ctx.SetFlowData("shipping_id", shippingID)
				return teleflow.NextStep()
			}
			return teleflow.Retry().WithPrompt("Please select a shipping method from the buttons above.")
		}).
		Step("select_payment").
		Prompt("💳 Select payment account:").
		WithPromptKeyboard(func(ctx *teleflow.Context) *teleflow.PromptKeyboardBuilder {
			accounts := businessService.GetAccounts(ctx.UserID())
			return PaymentAccountSelectionKeyboard(accounts)
		}).
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			if buttonClick != nil {
				accountID := buttonClick.Data.(string)
				ctx.SetFlowData("payment_account_id", accountID)
				return teleflow.NextStep()
			}
			return teleflow.Retry().WithPrompt("Please select a payment account from the buttons above.")
		}).
		Step("process_payment").
		Prompt("💰 Processing your order and payment...").
		Process(func(ctx *teleflow.Context, input string, buttonClick *teleflow.ButtonClick) teleflow.ProcessResult {
			// Calculate total amount
			itemID, _ := ctx.GetFlowData("selected_item_id")
			quantity, _ := ctx.GetFlowData("quantity")
			shippingID, _ := ctx.GetFlowData("shipping_id")
			paymentAccountID, _ := ctx.GetFlowData("payment_account_id")

			item := getMerchandiseByID(itemID.(string))
			shipping := getShippingByID(shippingID.(string))

			if item == nil || shipping == nil {
				return teleflow.Retry().WithPrompt("Error calculating order total. Please try again.")
			}

			totalAmount := (item.Price * float64(quantity.(int))) + shipping.Cost

			// Process payment
			err := businessService.ProcessPayment(ctx.UserID(), paymentAccountID.(string), totalAmount)
			if err != nil {
				if strings.Contains(err.Error(), "insufficient funds") {
					balance, _ := businessService.GetAccountBalance(ctx.UserID(), paymentAccountID.(string))
					templateData := map[string]interface{}{
						"AvailableBalance": balance,
						"RequiredAmount":   totalAmount,
						"Shortfall":        totalAmount - balance,
					}

					ctx.SendPrompt(&teleflow.PromptConfig{
						Message:      "insufficient_funds",
						TemplateData: templateData,
					})
					return teleflow.GoToStep("select_payment").WithPrompt("Please select a different payment account:")
				}
				return teleflow.Retry().WithPrompt("Payment failed. Please try again.")
			}

			// Store order details
			orderID := uuid.New().String()
			ctx.SetFlowData("order_id", orderID)
			ctx.SetFlowData("total_amount", totalAmount)
			ctx.SetFlowData("item_name", item.Name)
			ctx.SetFlowData("item_price", item.Price)
			ctx.SetFlowData("shipping_method", shipping.Method)
			ctx.SetFlowData("shipping_cost", shipping.Cost)

			return teleflow.CompleteFlow()
		}).
		OnComplete(func(ctx *teleflow.Context) error {
			// Get all order details
			orderID, _ := ctx.GetFlowData("order_id")
			totalAmount, _ := ctx.GetFlowData("total_amount")
			itemName, _ := ctx.GetFlowData("item_name")
			itemPrice, _ := ctx.GetFlowData("item_price")
			quantity, _ := ctx.GetFlowData("quantity")
			shippingMethod, _ := ctx.GetFlowData("shipping_method")
			shippingCost, _ := ctx.GetFlowData("shipping_cost")
			deliveryAddress, _ := ctx.GetFlowData("delivery_address")

			templateData := map[string]interface{}{
				"OrderID":         orderID,
				"ItemName":        itemName,
				"ItemPrice":       itemPrice,
				"Quantity":        quantity,
				"ShippingMethod":  shippingMethod,
				"ShippingCost":    shippingCost,
				"TotalAmount":     totalAmount,
				"DeliveryAddress": deliveryAddress,
			}

			// Send order summary first
			ctx.SendPrompt(&teleflow.PromptConfig{
				Message:      "order_summary",
				TemplateData: templateData,
			})

			// Then send confirmation
			return ctx.SendPrompt(&teleflow.PromptConfig{
				Message:      "order_confirmation",
				TemplateData: templateData,
			})
		}).
		Build()
}

// Helper functions for merchandise and shipping

func getMerchandiseForCategory(category string) []Merchandise {
	switch category {
	case "📱 Electronics":
		return []Merchandise{
			{ID: "phone1", Name: "Smartphone Pro", Price: 999.99},
			{ID: "laptop1", Name: "Gaming Laptop", Price: 1499.99},
			{ID: "tablet1", Name: "Tablet Air", Price: 599.99},
		}
	case "👕 Clothing":
		return []Merchandise{
			{ID: "shirt1", Name: "Premium T-Shirt", Price: 29.99},
			{ID: "jeans1", Name: "Designer Jeans", Price: 89.99},
			{ID: "jacket1", Name: "Winter Jacket", Price: 149.99},
		}
	case "📚 Books":
		return []Merchandise{
			{ID: "book1", Name: "Programming Guide", Price: 39.99},
			{ID: "book2", Name: "Business Strategy", Price: 24.99},
			{ID: "book3", Name: "Self Development", Price: 19.99},
		}
	case "🏠 Home & Garden":
		return []Merchandise{
			{ID: "plant1", Name: "Indoor Plant", Price: 15.99},
			{ID: "lamp1", Name: "LED Desk Lamp", Price: 45.99},
			{ID: "chair1", Name: "Office Chair", Price: 199.99},
		}
	default:
		return []Merchandise{}
	}
}

func getMerchandiseByID(id string) *Merchandise {
	allCategories := []string{"📱 Electronics", "👕 Clothing", "📚 Books", "🏠 Home & Garden"}
	for _, category := range allCategories {
		items := getMerchandiseForCategory(category)
		for _, item := range items {
			if item.ID == id {
				return &item
			}
		}
	}
	return nil
}

func getShippingOptions() []ShippingOption {
	return []ShippingOption{
		{ID: "standard", Name: "Standard Shipping", Cost: 5.99, Method: "Standard (5-7 days)"},
		{ID: "express", Name: "Express Shipping", Cost: 12.99, Method: "Express (2-3 days)"},
		{ID: "overnight", Name: "Overnight Shipping", Cost: 24.99, Method: "Overnight (1 day)"},
	}
}

func getShippingByID(id string) *ShippingOption {
	options := getShippingOptions()
	for _, option := range options {
		if option.ID == id {
			return &option
		}
	}
	return nil
}
