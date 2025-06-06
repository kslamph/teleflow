# Photo Capabilities Showcase - example_new_api.go Update

## 🎯 Update Overview

The example has been enhanced to demonstrate TeleFlow's comprehensive photo capabilities using the Step-Prompt-Process API. The registration flow now includes images at every step, showcasing different image specification types.

## 📸 Photo Features Demonstrated

### 1. Static Image URLs
```go
// Welcome step with static image
Prompt(
    "👋 Welcome! Let's get you registered. What's your name?",
    "https://upload.wikimedia.org/wikipedia/commons/thumb/0/05/Go_Logo_Blue.svg/1200px-Go_Logo_Blue.svg.png", // Static URL
    nil,
)
```

### 2. Dynamic Images Based on Context
```go
// Age step with personalized image
Prompt(
    func(ctx *teleflow.Context) string {
        name, _ := ctx.Get("user_name")
        return fmt.Sprintf("Nice to meet you, %s! How old are you?", name)
    },
    func(ctx *teleflow.Context) string {
        // Dynamic image incorporating user's name
        name, _ := ctx.Get("user_name")
        return fmt.Sprintf("https://via.placeholder.com/400x200/9C27B0/white?text=Hello+%s", name)
    },
    nil,
)
```

### 3. Contextual Error/Success Images
```go
// Different images for validation states
if input == "" {
    return teleflow.RetryWithPrompt(&teleflow.PromptConfig{
        Message: "Please enter your name:",
        Image:   "https://via.placeholder.com/300x150/FF9800/white?text=Name+Required", // Warning
    })
}

// Success state
return teleflow.NextStep().WithPrompt(&teleflow.PromptConfig{
    Message: "✅ Name saved! Moving to the next step...",
    Image:   "https://via.placeholder.com/300x150/2196F3/white?text=Name+Saved", // Success
})
```

### 4. Multi-Data Dynamic Images
```go
// Confirmation step with multiple user data points
Image: func(ctx *teleflow.Context) string {
    name, _ := ctx.Get("user_name")
    age, _ := ctx.Get("user_age")
    return fmt.Sprintf("https://via.placeholder.com/400x300/607D8B/white?text=Confirm:+%s+(%s)", name, age)
},
```

### 5. Completion Celebration Image
```go
OnComplete(func(ctx *teleflow.Context) error {
    // Celebration image for successful completion
    return ctx.SendPrompt(&teleflow.PromptConfig{
        Message: fmt.Sprintf("🎉 Registration complete!\nName: %s\nAge: %s\n\nWelcome to our service!", name, age),
        Image:   "https://via.placeholder.com/500x300/4CAF50/white?text=🎉+Welcome+Aboard!",
    })
})
```

## 🆕 Additional Commands Added

### Demo Command
```bash
/demo
```
Shows a photo capabilities demonstration with an overview image.

### Enhanced Help Command
```bash
/help
```
Now includes an image and comprehensive command list.

## 🎨 Image Specifications Supported

| Type | Example | Use Case |
|------|---------|----------|
| **Static URL** | `"https://example.com/image.jpg"` | Fixed images, logos, banners |
| **Dynamic Function** | `func(ctx) string { return "url" }` | Context-based images |
| **File Path** | `"/path/to/image.jpg"` | Local files (when available) |
| **Base64** | `"data:image/png;base64,..."` | Embedded images |

## 🌈 Color-Coded Flow Experience

The registration flow now uses color psychology:

- **🟢 Green**: Welcome, success states
- **🟣 Purple**: Personal greetings  
- **🔵 Blue**: Progress indicators
- **🟠 Orange**: Warnings/required input
- **🔴 Red**: Errors/validation failures
- **🔷 Blue-Grey**: Confirmation/review

## 🧪 Test Scenarios

### 1. Normal Flow (with photos)
1. `/start` → Welcome image shown
2. Enter name → Success image + personalized greeting image
3. Enter age → Success image + confirmation image with data
4. Click "Yes" → Celebration completion image

### 2. Error Handling (with contextual photos)
1. `/start` → Welcome image
2. Submit empty name → Warning image with retry message
3. Enter valid name → Success image
4. Enter invalid age (>3 digits) → Error image with validation message

### 3. Demo & Help Commands
1. `/demo` → Photo capabilities overview image
2. `/help` → Help menu with command list image

## ✨ Benefits Demonstrated

1. **Visual Engagement**: Every interaction includes relevant imagery
2. **Context Awareness**: Images adapt based on user data and flow state
3. **Error Communication**: Visual feedback for validation states
4. **Professional Polish**: Consistent visual experience throughout
5. **Flexible Implementation**: Multiple image specification types supported

## 🎯 Key Takeaways

- **ImageSpec flexibility**: Supports static URLs, dynamic functions, file paths, and base64
- **Context integration**: Images can incorporate user data and flow state
- **Error handling**: Visual feedback enhances user experience
- **Seamless integration**: Works naturally with Step-Prompt-Process API
- **Professional results**: Easy to create visually appealing bots

The updated example now serves as a comprehensive showcase of TeleFlow's image capabilities while maintaining the clean, intuitive Step-Prompt-Process API! 🎉