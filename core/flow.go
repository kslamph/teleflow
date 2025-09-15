package teleflow

import (
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- Generic Flow System ---

type actionType int

const (
	actionNext actionType = iota
	actionRetry
	actionEnd
	actionRestart
	actionGoTo
)

// FlowAction encapsulates the action to take after a step and its potential target.
type FlowAction struct {
	Type       actionType
	TargetStep string
}

// Helper functions for clean, declarative flow control.
func Next() FlowAction    { return FlowAction{Type: actionNext} }
func Retry() FlowAction   { return FlowAction{Type: actionRetry} }
func End() FlowAction     { return FlowAction{Type: actionEnd} }
func Restart() FlowAction { return FlowAction{Type: actionRestart} }
func GoTo(stepName string) FlowAction {
	return FlowAction{Type: actionGoTo, TargetStep: stepName}
}

// Step defines a single, type-safe, named interaction within a flow.
type Step[T any] struct {
	Name    string
	Prompt  func(ctx *Context, data *T) error
	Process func(ctx *Context, data *T) (action FlowAction, err error)
}

type userFlowState[T any] struct {
	currentStep int
	flowData    T
}

// Flow is a thread-safe, stateful handler for complex, non-linear conversations.
type Flow[T any] struct {
	steps         []Step[T]
	stepNameIndex map[string]int // For fast GoTo lookups
	userStates    map[int64]*userFlowState[T]
	mu            sync.Mutex
	onComplete    func(ctx *Context, data T) error
}

func NewFlow[T any]() *Flow[T] {
	return &Flow[T]{
		steps:         []Step[T]{},
		stepNameIndex: make(map[string]int),
		userStates:    make(map[int64]*userFlowState[T]),
	}
}

func (f *Flow[T]) AddStep(step Step[T]) *Flow[T] {
	if _, exists := f.stepNameIndex[step.Name]; exists {
		panic(fmt.Sprintf("step with name '%s' already exists", step.Name))
	}
	f.stepNameIndex[step.Name] = len(f.steps)
	f.steps = append(f.steps, step)
	return f
}

func (f *Flow[T]) OnComplete(handler func(ctx *Context, data T) error) *Flow[T] {
	f.onComplete = handler
	return f
}

func (f *Flow[T]) HandleUpdate(ctx *Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	userID := ctx.UserID()
	state, inFlow := f.userStates[userID]

	if !inFlow {
		if len(f.steps) == 0 {
			return fmt.Errorf("flow has no steps")
		}
		var initialData T
		state = &userFlowState[T]{
			currentStep: 0,
			flowData:    initialData,
		}
		f.userStates[userID] = state
		return f.steps[0].Prompt(ctx, &state.flowData)
	}

	step := f.steps[state.currentStep]
	if step.Process == nil {
		return fmt.Errorf("step '%s' has no process function", step.Name)
	}

	action, err := step.Process(ctx, &state.flowData)
	if err != nil {
		delete(f.userStates, userID)
		return fmt.Errorf("error processing step '%s': %w", step.Name, err)
	}

	switch action.Type {
	case actionNext:
		state.currentStep++
		if state.currentStep >= len(f.steps) {
			delete(f.userStates, userID)
			if f.onComplete != nil {
				return f.onComplete(ctx, state.flowData)
			}
			return nil
		}
	case actionRetry:
		// Do nothing, just re-prompt at the end.
	case actionEnd:
		delete(f.userStates, userID)
		if f.onComplete != nil {
			return f.onComplete(ctx, state.flowData)
		}
		return nil
	case actionRestart:
		delete(f.userStates, userID)
		// Let the next entry into HandleUpdate start the flow fresh.
		return f.HandleUpdate(ctx)
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

func (f *Flow[T]) IsUserInFlow(userID int64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.userStates[userID]
	return ok
}

func (f *Flow[T]) CancelFlow(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.userStates, userID)
}

// --- Router and Helpers ---
type Router struct {
	routes       map[string]Handler
	defaultRoute Handler
}

func NewRouter() *Router {
	return &Router{routes: make(map[string]Handler)}
}

func (r *Router) Handle(command string, handler Handler) {
	r.routes[command] = handler
}

func (r *Router) Default(handler Handler) {
	r.defaultRoute = handler
}

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

type Func func(ctx *Context) error

func (f Func) HandleUpdate(ctx *Context) error { return f(ctx) }

// --- Keyboard Builders ---
type InlineKeyboard struct {
	keyboard [][]tgbotapi.InlineKeyboardButton
}

func NewInlineKeyboard() *InlineKeyboard { return &InlineKeyboard{} }

func (kb *InlineKeyboard) AddRow(buttons ...tgbotapi.InlineKeyboardButton) *InlineKeyboard {
	kb.keyboard = append(kb.keyboard, buttons)
	return kb
}

func ButtonCallback(text, data string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, data)
}

func (kb *InlineKeyboard) Build() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(kb.keyboard...)
}
