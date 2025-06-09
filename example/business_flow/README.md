# Business Flow Example

This example demonstrates a comprehensive business bot built with the Teleflow package, showcasing advanced features like multi-step flows, dynamic keyboards, template-based messages, and image generation.

## Features

- **Account Management**: View accounts, add new accounts with custom names and initial balances
- **Fund Transfers**: Transfer money between user accounts with validation and confirmation
- **Order Placement**: Complete e-commerce flow with product selection, shipping, and payment
- **Dynamic Images**: Generated images for welcome messages, product promotions, and delivery maps
- **Click-to-Copy**: Template-based messages with copyable account IDs and amounts
- **Interactive Keyboards**: Reply keyboards for main navigation and inline keyboards for choices
- **Error Handling**: Comprehensive error handling with user-friendly messages
- **Data Persistence**: In-memory business service for demo purposes

## Bot Commands

- `/start` - Welcome message with intro image
- `/help` - Help information about available features
- `/cancel` - Cancel current operation (global exit command)

## Main Menu Options

- **💼 Account Info** - View existing accounts or create new ones
- **💸 Transfer Funds** - Transfer money between accounts
- **🛒 Place Order** - Complete product ordering flow

## Flow Details

### Account Info Flow
1. Choose between viewing accounts or adding a new account
2. If viewing: Display all accounts with balances and copyable IDs
3. If adding: Enter account name and initial balance
4. Confirmation with new account details

### Transfer Funds Flow
1. Select source account from available accounts
2. Select destination account (excluding source)
3. Enter transfer amount with balance validation
4. Process transfer with success confirmation
5. Display transaction details with copyable information

### Place Order Flow
1. **Category Selection**: Choose from Electronics, Clothing, Books, or Home & Garden
2. **Product Selection**: Select specific items with prices
3. **Quantity Input**: Enter desired quantity
4. **Address Entry**: Type full delivery address
5. **Location Confirmation**: View generated map and confirm address
6. **Shipping Selection**: Choose shipping method and cost
7. **Payment Selection**: Select payment account
8. **Order Processing**: Process payment and generate order
9. **Confirmation**: Display order summary with copyable order ID

## Technical Features

### Templates
- MarkdownV2 formatted messages with click-to-copy elements
- Dynamic data injection for personalized content
- Error messages for insufficient funds and validation

### Image Generation
- Welcome images with custom text
- Product promotion banners
- Mock delivery maps with address markers
- All images generated dynamically using Go's image package

### Keyboard Management
- Reply keyboards for main navigation
- Inline keyboards for step-by-step choices
- Dynamic keyboard generation based on available data
- Account selection with balance display

### Business Service
- In-memory data storage for demo purposes
- Thread-safe operations with proper mutex handling
- Account management (create, view, balance operations)
- Fund transfer validation and processing
- Payment processing for orders
- Uses read-write locks for optimal concurrent access

## Running the Example

1. Set your Telegram bot token:
   ```bash
   export TELEGRAM_BOT_TOKEN="your_bot_token_here"
   ```

2. Run the example:
   ```bash
   go run .
   ```

3. Start a conversation with your bot and use `/start` to begin

## Testing and Safety

This example includes proper concurrent programming practices:

- **Race Condition Detection**: Test with the race detector enabled:
  ```bash
  go run -race .
  ```

- **Thread-Safe Operations**: The business service uses read-write mutexes properly to avoid deadlocks and race conditions

- **Double-Checked Locking**: The account initialization uses the standard double-checked locking pattern to prevent duplicate initialization in concurrent scenarios

## File Structure

- `main.go` - Bot initialization and command handlers
- `business_service.go` - Mock business logic and data management
- `flows.go` - Flow definitions for all business processes
- `keyboards.go` - Keyboard helper functions
- `templates.go` - Message template definitions
- `image_generator.go` - Dynamic image generation utilities

## Example Interactions

### Account Creation
```
User: 💼 Account Info
Bot: Account Management - What would you like to do?
User: [Clicks "Add New Account"]
Bot: Enter a name for your new account:
User: Investment Account
Bot: Enter initial balance for 'Investment Account' account:
User: 2500.00
Bot: ✅ Account created successfully!
     Name: Investment Account
     Account ID: `abc123-def456-ghi789`
     Initial Balance: `$2500.00`
```

### Fund Transfer
```
User: 💸 Transfer Funds
Bot: Select the account to transfer FROM:
User: [Clicks "Main Account - $1000.00"]
Bot: Select the account to transfer TO:
User: [Clicks "Savings Account - $5000.00"]
Bot: Enter transfer amount (Available balance: $1000.00):
User: 250.00
Bot: ✅ Transfer Successful!
     Transfer of `$250.00` from account `main123` to account `save456` completed successfully.
```

### Order Placement
```
User: 🛒 Place Order
Bot: [Shows promotional image] Welcome to our store! Select a product category:
User: [Clicks "📱 Electronics"]
Bot: Select an item from 📱 Electronics:
User: [Clicks "Smartphone Pro - $999.99"]
Bot: Enter the quantity you want to order:
User: 1
Bot: Please type your full delivery address:
User: 123 Main St, New York, NY 10001
Bot: [Shows map image] Confirm delivery address: 123 Main St, New York, NY 10001
User: [Clicks "✅ Yes, Confirm"]
Bot: Select shipping method:
User: [Clicks "Express (2-3 days) - $12.99"]
Bot: Select payment account:
User: [Clicks "Main Account - $750.00"]
Bot: 🎉 Order Confirmed!
     Order ID: `ord_789abc123def`
     Total Paid: `$1012.98`
```

This example demonstrates the full capabilities of the Teleflow package in a realistic business scenario, showcasing how to build complex, interactive Telegram bots with professional-grade features.