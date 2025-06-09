# Data Lifecycle in Teleflow Core

This document explains how flow-specific data and prompt keyboard mapped data are managed within the Teleflow core, focusing on their storage and lifecycle. This assumes no external persistent state manager is currently in use for this data.

## Flow Data Management

Flow data pertains to the state and information collected as a user progresses through a defined sequence of steps (a flow).

**Storage:**

1.  **`flowManager` (`core/flow.go`):**
    *   The central component is the `flowManager` struct.
    *   `flows map[string]*Flow`: Stores definitions of all registered flows (static after registration).
    *   `userFlows map[int64]*userFlowState`: This is the primary in-memory store for the active state of each user currently in a flow, keyed by `userID`.

2.  **`userFlowState` (`core/flow.go`):**
    *   `FlowName string`: Name of the active flow.
    *   `CurrentStep string`: The current step the user is on.
    *   `Data map[string]interface{}`: **The main store for user-specific data collected during the flow.** This map is updated as the user provides input at each step.
    *   `StartedAt time.Time`: Timestamp of flow initiation.
    *   `LastActive time.Time`: Timestamp of the last user interaction.
    *   `LastMessageID int`: ID of the last message in the flow.

**Lifecycle:**

*   **Creation:** A new `userFlowState` is created and added to `flowManager.userFlows` when a user starts a flow (via `fm.startFlow()`). Initial data can be passed via the `Context`.
*   **Access & Modification:**
    *   During `fm.HandleUpdate()`, the current `userState.Data` is loaded into the `Context`.
    *   After a step's `ProcessFunc` executes, data from the `Context` is copied back to `userState.Data`.
*   **Deletion:** `userFlowState` (and its `Data`) is deleted from `fm.userFlows` when:
    *   The flow is explicitly cancelled.
    *   The flow completes successfully.
    *   An unrecoverable error occurs during the flow.
    *   The flow or current step definition is not found.

## Prompt Keyboard Mapped Data Management

This refers to the data associated with inline keyboard buttons, specifically their callback data.

**Storage:**

1.  **`PromptKeyboardHandler` (`core/prompt_keyboard_handler.go`):**
    *   Manages data for inline keyboard buttons.
    *   `userUUIDMappings map[int64]map[string]interface{}`: An in-memory nested map.
        *   Outer map keyed by `userID`.
        *   Inner map keyed by a unique `uuid` (string) generated for each callback button.
        *   The `interface{}` value is the custom data associated with that button.

2.  **`PromptKeyboardBuilder` (`core/prompt_keyboard_builder.go`):**
    *   When a callback button is added (`ButtonCallback(text string, data interface{})`):
        *   A `callbackUUID` is generated.
        *   The provided `data` is mapped to this `callbackUUID` in the builder's internal `uuidMapping`.
        *   This `callbackUUID` is sent to Telegram as the button's `callback_data`.

**Lifecycle:**

*   **Creation/Population:**
    *   When a prompt with an inline keyboard is sent, `promptKeyboardHandler.buildKeyboard()` is called.
    *   Mappings from the `PromptKeyboardBuilder`'s `uuidMapping` are copied into `promptKeyboardHandler.userUUIDMappings[userID]`. This happens *before* the keyboard is shown to the user.
*   **Access/Retrieval:**
    *   When a user clicks an inline button, Telegram sends the `callbackUUID`.
    *   `flowManager.extractInputData()` calls `promptKeyboardHandler.getCallbackData()` to retrieve the original custom `data` using the `userID` and `callbackUUID`.
    *   This retrieved `data` is then available to the flow's step processing logic (typically in `ButtonClick.Data`).
*   **Deletion/Clearing:**
    *   Mappings for a user (`promptKeyboardHandler.userUUIDMappings[userID]`) are cleared via `promptKeyboardHandler.cleanupUserMappings()`.
    *   This cleanup occurs when a user's flow completes successfully or is cancelled.

## Overall Data Management Summary

Both flow data and keyboard callback data are stored **in-memory** within the Go application. This means if the application restarts, this transient data is lost.

## Visualization

```mermaid
graph TD
    subgraph Bot
        BAPI[tgbotapi.BotAPI]
        FM[flowManager]
        PKH[PromptKeyboardHandler]
        PC[PromptComposer]
    end

    FM --- BAPI
    PC --- BAPI
    PC --- PKH

    subgraph flowManager
        direction LR
        RegisteredFlows[flows map[string]*Flow]
        ActiveUserFlows[userFlows map[int64]*userFlowState]
    end

    subgraph userFlowState
        direction LR
        UF_FlowName[FlowName string]
        UF_CurrentStep[CurrentStep string]
        UF_Data[Data map[string]interface{}]
        UF_Timestamps[StartedAt, LastActive]
    end
    ActiveUserFlows --> UF_Data

    subgraph PromptKeyboardHandler
        direction LR
        UserUUIDMappings[userUUIDMappings map[int64]map[string]interface{}]
    end
    UserUUIDMappings --> MappedButtonData[interface{}]


    subgraph Context
        direction LR
        Ctx_Update[Update]
        Ctx_BotRef[Bot Reference]
        Ctx_Data[data map[string]interface{}]
    end

    ProcessFuncArgs --> Ctx_Data
    ProcessFuncArgs --> ButtonClickData[ButtonClick.Data]


    UserInteraction[User Interaction] --> Ctx_Update
    Ctx_Update -- CallbackQuery --> FM
    FM -- Retrieves/Updates --> ActiveUserFlows
    FM -- Uses --> PKH

    PKH -- buildKeyboard --> PromptKeyboardBuilderRef[PromptKeyboardBuilder]
    PromptKeyboardBuilderRef -- ButtonCallback --> UUID_to_Data_Mapping[uuidMapping in Builder]
    UUID_to_Data_Mapping -- Copied to --> UserUUIDMappings

    FM -- extractInputData & Callback --> PKH
    PKH -- getCallbackData --> MappedButtonData
    MappedButtonData --> ButtonClickData

    FM -- HandleUpdate copies --> Ctx_Data
    Ctx_Data -- Copied from/to --> UF_Data


    style FM fill:#f9f,stroke:#333,stroke-width:2px
    style PKH fill:#ccf,stroke:#333,stroke-width:2px
    style ActiveUserFlows fill:#lightgreen,stroke:#333,stroke-width:1px
    style UserUUIDMappings fill:#lightblue,stroke:#333,stroke-width:1px
    style UF_Data fill:#palegreen,stroke:#333,stroke-width:1px
    style MappedButtonData fill:#skyblue,stroke:#333,stroke-width:1px