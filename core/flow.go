package teleflow

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// flowStateTTL is how long an idle user flow state is kept before eviction.
// States are purged lazily on access.
const flowStateTTL = 24 * time.Hour

// --- Flow Control ---

type actionType int

const (
	actionNext actionType = iota
	actionRetry
	actionEnd
	actionGoTo
)

// FlowAction encapsulates the action to take after a step's Process function
// completes. It is returned by the Process function to control the conversation's
// progression.
type FlowAction struct {
	// Type is the kind of transition to perform (Next, Retry, End, or GoTo).
	Type actionType
	// TargetStep is the name of the destination step when Type is actionGoTo;
	// empty otherwise.
	TargetStep string
}

// Next moves the flow to the next step in the sequence.
func Next() FlowAction { return FlowAction{Type: actionNext} }

// Retry re-prompts the current step. Useful for input validation.
func Retry() FlowAction { return FlowAction{Type: actionRetry} }

// End terminates the flow for the user. The OnComplete handler will be called.
func End() FlowAction { return FlowAction{Type: actionEnd} }

// GoTo jumps to a specific step by its unique name. This enables the creation of
// non-linear conversations, loops, and branches.
func GoTo(stepName string) FlowAction {
	return FlowAction{Type: actionGoTo, TargetStep: stepName}
}

// Step defines a single, logical part of a conversation. Each step must have a
// unique Name for non-linear control via GoTo.
// The Prompt function sends a message to the user (e.g., asks a question).
// The Process function handles the user's response and returns a FlowAction to
// direct the conversation.
type Step[T any] struct {
	// Name is a unique identifier for this step, used by GoTo for non-linear
	// transitions.
	Name string
	// Prompt sends a message to the user (e.g., asks a question). It is called
	// when the step becomes active.
	Prompt func(ctx *Context, data *T) error
	// Process handles the user's response and returns a FlowAction to direct the
	// conversation.
	Process func(ctx *Context, data *T) (action FlowAction, err error)
}

// flowStateKey is a composite key that uniquely identifies a user's state
// within a specific chat, preventing state leakage across group chats.
type flowStateKey struct {
	ChatID, UserID int64
}

type userFlowState[T any] struct {
	mu          sync.Mutex // Serializes HandleUpdate for this user; held across callbacks.
	currentStep int
	flowData    T
	lastActive  time.Time
}

// Flow is a stateful handler that manages a multi-step conversation for multiple
// users concurrently. It is parameterized with a generic type `T` which defines the
// structure of the data that will be collected and maintained throughout the flow.
// It is thread-safe.
type Flow[T any] struct {
	steps         []Step[T]
	stepNameIndex map[string]int // For fast GoTo lookups
	userStates    map[flowStateKey]*userFlowState[T]
	mu            sync.Mutex
	onComplete    func(ctx *Context, data T) error
}

// NewFlow creates and initializes a new Flow for a given data type `T`.
func NewFlow[T any]() *Flow[T] {
	return &Flow[T]{
		steps:         []Step[T]{},
		stepNameIndex: make(map[string]int),
		userStates:    make(map[flowStateKey]*userFlowState[T]),
	}
}

// AddStep adds a new step to the flow's sequence. Steps are executed in the order
// they are added, unless controlled otherwise by a FlowAction like GoTo.
// Returns an error if a step with the same name already exists.
func (f *Flow[T]) AddStep(step Step[T]) (*Flow[T], error) {
	if _, exists := f.stepNameIndex[step.Name]; exists {
		return nil, fmt.Errorf("step with name '%s' already exists", step.Name)
	}
	f.stepNameIndex[step.Name] = len(f.steps)
	f.steps = append(f.steps, step)
	return f, nil
}

// OnComplete sets a final handler that is called only when the flow finishes
// successfully (e.g., via the End() action or by completing the last step).
// The handler receives the final, populated data from the flow.
func (f *Flow[T]) OnComplete(handler func(ctx *Context, data T) error) *Flow[T] {
	f.onComplete = handler
	return f
}

// HandleUpdate is the entry point for processing a user's message. It manages
// the user's state, calls the appropriate Prompt or Process function for the
// current step, and transitions the user between steps based on the returned
// FlowAction.
//
// Callbacks must NOT call HandleUpdate reentrantly on the same flow (i.e.,
// from within a Prompt or Process function). Doing so will deadlock.
func (f *Flow[T]) HandleUpdate(ctx *Context) error {
	key := flowStateKey{ChatID: ctx.ChatID(), UserID: ctx.UserID()}

	// Step 1: Get or create the user state under f.mu (held briefly).
	// Lazy-purge any expired state while holding the lock.
	f.mu.Lock()
	for k, s := range f.userStates {
		if time.Since(s.lastActive) > flowStateTTL {
			delete(f.userStates, k)
		}
	}
	state, inFlow := f.userStates[key]
	if !inFlow {
		if len(f.steps) == 0 {
			f.mu.Unlock()
			return fmt.Errorf("flow has no steps")
		}
		var initialData T
		state = &userFlowState[T]{currentStep: 0, flowData: initialData, lastActive: time.Now()}
		f.userStates[key] = state
	} else {
		state.lastActive = time.Now()
	}
	f.mu.Unlock()

	// Step 2: Acquire per-user mutex (held across callbacks).
	// Lock ordering: f.mu then state.mu; we never hold f.mu here.
	state.mu.Lock()
	defer state.mu.Unlock()

	// Step 3: Run user callbacks without f.mu held.
	if !inFlow {
		return f.steps[0].Prompt(ctx, &state.flowData)
	}

	step := f.steps[state.currentStep]
	if step.Process == nil {
		return fmt.Errorf("step '%s' has no process function", step.Name)
	}

	action, err := step.Process(ctx, &state.flowData)
	if err != nil {
		f.mu.Lock()
		delete(f.userStates, key)
		f.mu.Unlock()
		return fmt.Errorf("error processing step '%s': %w", step.Name, err)
	}

	switch action.Type {
	case actionNext:
		state.currentStep++
		if state.currentStep >= len(f.steps) {
			f.mu.Lock()
			delete(f.userStates, key)
			f.mu.Unlock()
			if f.onComplete != nil {
				return f.onComplete(ctx, state.flowData)
			}
			return nil
		}
	case actionRetry:
		// Do nothing, just re-prompt at the end.
	case actionEnd:
		f.mu.Lock()
		delete(f.userStates, key)
		f.mu.Unlock()
		if f.onComplete != nil {
			return f.onComplete(ctx, state.flowData)
		}
		return nil
	case actionGoTo:
		targetIndex, ok := f.stepNameIndex[action.TargetStep]
		if !ok {
			return fmt.Errorf("step '%s' not found", action.TargetStep)
		}
		state.currentStep = targetIndex
	}

	// Re-prompt the (potentially new) current step.
	return f.steps[state.currentStep].Prompt(ctx, &state.flowData)
}

// IsUserInFlow checks if a user is currently active in this flow.
func (f *Flow[T]) IsUserInFlow(chatID, userID int64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.userStates[flowStateKey{ChatID: chatID, UserID: userID}]
	return ok
}

// CancelFlow immediately removes a user's state from the flow, effectively
// terminating their session.
func (f *Flow[T]) CancelFlow(chatID, userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.userStates, flowStateKey{ChatID: chatID, UserID: userID})
}

// Router is a command-based multiplexer for Handlers. It inspects incoming
// messages for commands (e.g., "/start") and routes the update to the
// corresponding registered Handler.
type Router struct {
	routes       map[string]Handler
	defaultRoute Handler
}

// NewRouter creates and initializes a new Router.
func NewRouter() *Router {
	return &Router{routes: make(map[string]Handler)}
}

// Handle registers a Handler for a specific command. The command should be
// provided without the leading slash (e.g., "start" for "/start").
func (r *Router) Handle(command string, handler Handler) {
	r.routes[command] = handler
}

// Default sets a fallback Handler that is called when an incoming message
// is not a command or does not match any registered command.
func (r *Router) Default(handler Handler) {
	r.defaultRoute = handler
}

// HandleUpdate dispatches the update. It routes command messages to the
// matching registered handler; all other updates (including callback queries
// and non-command messages) fall through to the default route.
func (r *Router) HandleUpdate(ctx *Context) error {
	if msg := ctx.Update().Message; msg != nil && msg.IsCommand() {
		if handler, ok := r.routes[msg.Command()]; ok {
			return handler.HandleUpdate(ctx)
		}
	}
	if r.defaultRoute != nil {
		return r.defaultRoute.HandleUpdate(ctx)
	}
	return nil
}

// Func is an adapter to allow the use of ordinary functions as Handlers.
type Func func(ctx *Context) error

// HandleUpdate delegates to the underlying function.
func (f Func) HandleUpdate(ctx *Context) error { return f(ctx) }

// InlineKeyboard is a helper for building Telegram Inline Keyboards using a
// fluent, chainable API.
type InlineKeyboard struct {
	keyboard [][]tgbotapi.InlineKeyboardButton
}

// NewInlineKeyboard creates a new, empty keyboard builder.
func NewInlineKeyboard() *InlineKeyboard { return &InlineKeyboard{} }

// AddRow adds a new row of buttons to the keyboard. Each call to AddRow creates
// a new horizontal line of buttons.
func (kb *InlineKeyboard) AddRow(buttons ...tgbotapi.InlineKeyboardButton) *InlineKeyboard {
	kb.keyboard = append(kb.keyboard, buttons)
	return kb
}

// ButtonCallback creates a standard InlineKeyboardButton with the given text
// and callback data.
func ButtonCallback(text, data string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, data)
}

// Build finalizes the keyboard and returns the `tgbotapi.InlineKeyboardMarkup`
// object that can be sent with a message.
func (kb *InlineKeyboard) Build() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(kb.keyboard...)
}
