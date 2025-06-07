# Process Message Actions - Flow-Level Keyboard Management

This example demonstrates the new flow-level message handling options that solve the scroll-back keyboard issue in Telegram bot flows.

## The Problem

Previously, when users scrolled back in chat history and clicked old inline keyboard buttons, they could trigger unexpected behavior because UUID mappings were cleaned up too early. This created a confusing user experience where old buttons might not work as expected.

## The Solution

New flow-level configuration options that give developers control over how previous messages and keyboards are handled when users interact with buttons:

### 1. `OnProcessDeleteMessage()`
- **Behavior**: Completely deletes the message containing clicked button
- **Use Case**: Clean, minimal UX where only the current step is visible
- **UUID Mapping**: Cleaned up when messages are deleted
- **Target**: The specific message that contained the clicked button

```go
flow := teleflow.NewFlow("example").
    OnProcessDeleteMessage(). // Enable message deletion
    Step("menu").
    Prompt("Choose an option:").
    // ... rest of flow
```

### 2. `OnProcessDeleteKeyboard()`
- **Behavior**: Removes only keyboard from the message containing clicked button
- **Use Case**: Preserve conversation history but disable old interactive elements
- **UUID Mapping**: Cleaned up when keyboards are removed
- **Target**: The specific message that contained the clicked button

```go
flow := teleflow.NewFlow("example").
    OnProcessDeleteKeyboard(). // Enable keyboard removal
    Step("menu").
    Prompt("Choose an option:").
    // ... rest of flow
```

### 3. Default Behavior (No method called)
- **Behavior**: Keep all messages and keyboards untouched
- **Use Case**: Traditional behavior where all keyboards remain functional
- **UUID Mapping**: Persisted until flow ends (solves scroll-back issue)

```go
flow := teleflow.NewFlow("example").
    // No OnProcess* method called - default behavior
    Step("menu").
    Prompt("Choose an option:").
    // ... rest of flow
```

## Implementation Details

### Current Implementation Strategy
The implementation deletes/modifies the **specific message that contains the clicked button**, identified from the Telegram callback query. This provides immediate visual feedback and prevents confusion from multiple identical keyboards.

### Data Mapping Strategy
- **When deletion is enabled**: UUID mappings are cleaned up when messages/keyboards are deleted
- **When deletion is disabled**: UUID mappings persist until the flow ends, solving the scroll-back issue
- **Automatic cleanup**: Mappings are always cleaned up when flows complete or are cancelled

### Technical Changes
- Added `ProcessMessageAction` enum with three options
- Enhanced `FlowConfig` with `OnProcessAction` field
- Added flow builder methods for declarative configuration
- Modified message handling logic to respect flow-level configuration
- Conditional UUID mapping cleanup based on deletion settings

## Running the Example

1. Set your bot token in the example code
2. Run the example:
   ```bash
   go run example.go
   ```
3. Use these commands to test different behaviors:
   - `/delete_messages` - Test complete message deletion
   - `/delete_keyboards` - Test keyboard-only removal
   - `/keep_messages` - Test default behavior with persistent keyboards
   - `/help` - Show help with all available commands

## Benefits

- **Solves scroll-back issue**: Old keyboards work correctly when deletion is disabled
- **Developer control**: Explicit, declarative configuration per flow
- **Clean UX options**: Choose between minimal (delete) or historical (keep) interfaces
- **Performance**: Efficient memory management with proper cleanup
- **Backward compatible**: Default behavior unchanged

## Use Cases

- **Delete Messages**: Wizards, forms, clean step-by-step processes
- **Delete Keyboards**: Support conversations, help systems, documentation flows  
- **Keep Everything**: Complex decision trees, multi-path flows, reference materials

This feature transforms the scroll-back problem from a technical issue into a deliberate UX design choice.