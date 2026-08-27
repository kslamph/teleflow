package teleflow

import (
	"fmt"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// testContext creates a Context with a minimal update for the given user/chat.
func testContext(chatID, userID int64) *Context {
	return NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: userID},
			Chat: &tgbotapi.Chat{ID: chatID},
		},
	}, nil, nil)
}

// testContextWithText creates a Context with a message containing text.
func testContextWithText(chatID, userID int64, text string) *Context {
	return NewContext(tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: userID},
			Chat: &tgbotapi.Chat{ID: chatID},
			Text: text,
		},
	}, nil, nil)
}

// simpleStep builds a Step that records its calls and chain.
type stepRecorder struct {
	mu    sync.Mutex
	calls []string
}

func (r *stepRecorder) add(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, s)
}

func (r *stepRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = nil
}

func simpleStep[T any](name string, rec *stepRecorder, process func(data *T) FlowAction) Step[T] {
	return Step[T]{
		Name: name,
		Prompt: func(ctx *Context, data *T) error {
			rec.add("prompt:" + name)
			return nil
		},
		Process: func(ctx *Context, data *T) (FlowAction, error) {
			rec.add("process:" + name)
			if process != nil {
				return process(data), nil
			}
			return Next(), nil
		},
	}
}

// TestFlowStart creates a flow and processes the first user update.
func TestFlowStart(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(simpleStep[string]("s1", rec, nil))

	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "prompt:s1" {
		t.Fatalf("expected prompt:s1, got %v", rec.calls)
	}
}

// TestFlowNext advances through steps with Next().
func TestFlowNext(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(simpleStep[string]("s1", rec, nil))
	flow.AddStep(simpleStep[string]("s2", rec, nil))

	// First call: start flow → prompt s1
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first HandleUpdate: %v", err)
	}
	rec.calls = nil

	// Second call: process s1 → Next → prompt s2
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("second HandleUpdate: %v", err)
	}
	if len(rec.calls) != 2 || rec.calls[0] != "process:s1" || rec.calls[1] != "prompt:s2" {
		t.Fatalf("expected process:s1 → prompt:s2, got %v", rec.calls)
	}
}

// TestFlowNextComplete completes the last step and calls OnComplete.
func TestFlowNextComplete(t *testing.T) {
	rec := &stepRecorder{}
	completed := false
	flow := NewFlow[string]()
	flow.AddStep(simpleStep[string]("s1", rec, nil))
	flow.OnComplete(func(ctx *Context, data string) error {
		completed = true
		return nil
	})

	// Start flow
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first HandleUpdate: %v", err)
	}
	rec.calls = nil

	// Process last step → Next → OnComplete
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("second HandleUpdate: %v", err)
	}
	if !completed {
		t.Fatal("OnComplete not called")
	}
	if len(rec.calls) != 1 || rec.calls[0] != "process:s1" {
		t.Fatalf("expected process:s1, got %v", rec.calls)
	}
	// State should be gone
	if flow.IsUserInFlow(1, 1) {
		t.Fatal("user should no longer be in flow")
	}
}

// TestFlowRetry re-prompts the current step.
func TestFlowRetry(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name: "s1",
		Prompt: func(ctx *Context, data *string) error {
			rec.add("prompt:s1")
			return nil
		},
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			rec.add("process:s1")
			return Retry(), nil
		},
	})

	// Start
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first: %v", err)
	}
	rec.calls = nil

	// Process → Retry → re-prompt same step
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(rec.calls) != 2 || rec.calls[0] != "process:s1" || rec.calls[1] != "prompt:s1" {
		t.Fatalf("expected process:s1 → prompt:s1, got %v", rec.calls)
	}
}

// TestFlowEnd ends the flow and calls OnComplete.
func TestFlowEnd(t *testing.T) {
	rec := &stepRecorder{}
	completed := false
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name: "s1",
		Prompt: func(ctx *Context, data *string) error {
			rec.add("prompt:s1")
			return nil
		},
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			rec.add("process:s1")
			return End(), nil
		},
	})
	flow.OnComplete(func(ctx *Context, data string) error {
		completed = true
		return nil
	})

	// Start
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first: %v", err)
	}
	rec.calls = nil

	// Process → End → OnComplete
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("second: %v", err)
	}
	if !completed {
		t.Fatal("OnComplete not called")
	}
	if flow.IsUserInFlow(1, 1) {
		t.Fatal("user should not be in flow after End")
	}
}

// TestFlowGoTo jumps to a named step.
func TestFlowGoTo(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(simpleStep[string]("s1", rec, nil))
	flow.AddStep(simpleStep[string]("s2", rec, nil))
	flow.AddStep(Step[string]{
		Name: "s3",
		Prompt: func(ctx *Context, data *string) error {
			rec.add("prompt:s3")
			return nil
		},
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			rec.add("process:s3")
			return End(), nil
		},
	})

	// Start
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first: %v", err)
	}
	rec.calls = nil

	// Process s1 → GoTo("s3") → prompt s3
	// We need a custom Process for s1
	flow.steps[0].Process = func(ctx *Context, data *string) (FlowAction, error) {
		rec.add("process:s1")
		return GoTo("s3"), nil
	}
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(rec.calls) != 2 || rec.calls[0] != "process:s1" || rec.calls[1] != "prompt:s3" {
		t.Fatalf("expected process:s1 → prompt:s3, got %v", rec.calls)
	}
}

// TestFlowGoToUnknown returns an error for a missing step name.
func TestFlowGoToUnknown(t *testing.T) {
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name:   "s1",
		Prompt: func(ctx *Context, data *string) error { return nil },
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			return GoTo("nonexistent"), nil
		},
	})

	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first: %v", err)
	}
	err := flow.HandleUpdate(testContext(1, 1))
	if err == nil {
		t.Fatal("expected error for unknown GoTo target")
	}
}

// TestFlowNoSteps returns an error when no steps are defined.
func TestFlowNoSteps(t *testing.T) {
	flow := NewFlow[string]()
	if err := flow.HandleUpdate(testContext(1, 1)); err == nil {
		t.Fatal("expected error for flow with no steps")
	}
}

// TestFlowNoProcess returns an error when a step has no Process function.
func TestFlowNoProcess(t *testing.T) {
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name: "s1",
		Prompt: func(ctx *Context, data *string) error {
			return nil
		},
		// Process is nil
	})
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first: %v", err)
	}
	err := flow.HandleUpdate(testContext(1, 1))
	if err == nil {
		t.Fatal("expected error for nil Process")
	}
}

// TestFlowProcessError removes state when Process returns an error.
func TestFlowProcessError(t *testing.T) {
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name:   "s1",
		Prompt: func(ctx *Context, data *string) error { return nil },
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			return Next(), fmt.Errorf("oops")
		},
	})
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first: %v", err)
	}
	err := flow.HandleUpdate(testContext(1, 1))
	if err == nil {
		t.Fatal("expected error")
	}
	if flow.IsUserInFlow(1, 1) {
		t.Fatal("user should be removed after Process error")
	}
}

// TestFlowAddStepDuplicate returns an error.
func TestFlowAddStepDuplicate(t *testing.T) {
	flow := NewFlow[string]()
	step := Step[string]{Name: "s1"}
	if _, err := flow.AddStep(step); err != nil {
		t.Fatalf("first AddStep: %v", err)
	}
	_, err := flow.AddStep(step)
	if err == nil {
		t.Fatal("expected error for duplicate step name")
	}
}

// TestFlowIsUserInFlow checks user state.
func TestFlowIsUserInFlow(t *testing.T) {
	flow := NewFlow[string]()
	flow.AddStep(simpleStep[string]("s1", &stepRecorder{}, nil))
	if flow.IsUserInFlow(1, 1) {
		t.Fatal("user should not be in flow initially")
	}
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if !flow.IsUserInFlow(1, 1) {
		t.Fatal("user should be in flow after start")
	}
	if flow.IsUserInFlow(2, 1) {
		t.Fatal("different chat should not be in flow")
	}
}

// TestFlowCancelFlow removes user state.
func TestFlowCancelFlow(t *testing.T) {
	flow := NewFlow[string]()
	flow.AddStep(simpleStep[string]("s1", &stepRecorder{}, nil))
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	flow.CancelFlow(1, 1)
	if flow.IsUserInFlow(1, 1) {
		t.Fatal("user should be removed after CancelFlow")
	}
}

// TestFlowUserIsolation ensures different users don't interfere.
func TestFlowUserIsolation(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name: "counter",
		Prompt: func(ctx *Context, data *string) error {
			rec.add("prompt:" + fmt.Sprint(ctx.ChatID()))
			return nil
		},
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			rec.add("process:" + fmt.Sprint(ctx.ChatID()))
			return End(), nil
		},
	})

	// Two users start
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("user1 start: %v", err)
	}
	if err := flow.HandleUpdate(testContext(2, 2)); err != nil {
		t.Fatalf("user2 start: %v", err)
	}
	rec.calls = nil

	// Both process
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("user1 process: %v", err)
	}
	if err := flow.HandleUpdate(testContext(2, 2)); err != nil {
		t.Fatalf("user2 process: %v", err)
	}
	if len(rec.calls) != 2 ||
		rec.calls[0] != "process:1" || rec.calls[1] != "process:2" {
		t.Fatalf("expected process:1 then process:2, got %v", rec.calls)
	}
}

// TestFlowSameUserDifferentChat verifies the chat+user composite key.
func TestFlowSameUserDifferentChat(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name: "s1",
		Prompt: func(ctx *Context, data *string) error {
			rec.add("prompt:" + fmt.Sprint(ctx.ChatID()))
			return nil
		},
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			rec.add("process:" + fmt.Sprint(ctx.ChatID()))
			return Next(), nil
		},
	})
	flow.AddStep(simpleStep[string]("s2", rec, nil))

	// Same user in two different chats should have independent states
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("chat1 start: %v", err)
	}
	if err := flow.HandleUpdate(testContext(2, 1)); err != nil {
		t.Fatalf("chat2 start: %v", err)
	}
	rec.calls = nil

	// Process chat1 → advances to s2
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("chat1 process: %v", err)
	}
	if len(rec.calls) != 2 || rec.calls[0] != "process:1" || rec.calls[1] != "prompt:s2" {
		t.Fatalf("chat1 should advance to s2, got %v", rec.calls)
	}
}

// TestFlowOnCompleteData verifies OnComplete receives the final data.
func TestFlowOnCompleteData(t *testing.T) {
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name:   "s1",
		Prompt: func(ctx *Context, data *string) error { return nil },
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			*data = "hello"
			return End(), nil
		},
	})
	var finalData string
	flow.OnComplete(func(ctx *Context, data string) error {
		finalData = data
		return nil
	})

	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("second: %v", err)
	}
	if finalData != "hello" {
		t.Fatalf("expected 'hello', got %q", finalData)
	}
}

// TestFlowPurgeExpired verifies that idle states beyond TTL are evicted.
func TestFlowPurgeExpired(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(simpleStep[string]("s1", rec, nil))

	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("HandleUpdate: %v", err)
	}
	if !flow.IsUserInFlow(1, 1) {
		t.Fatal("user should be in flow after start")
	}
	rec.calls = nil

	// Manually set lastActive far in the past
	flow.mu.Lock()
	key := flowStateKey{ChatID: 1, UserID: 1}
	if s, ok := flow.userStates[key]; ok {
		s.lastActive = time.Now().Add(-(flowStateTTL + time.Hour))
	}
	flow.mu.Unlock()

	// Next call purges the expired state and starts a fresh flow (prompt again).
	if err := flow.HandleUpdate(testContext(1, 1)); err != nil {
		t.Fatalf("HandleUpdate after expiry: %v", err)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "prompt:s1" {
		t.Fatalf("expected a fresh prompt:s1 after purge, got %v", rec.calls)
	}
}

// TestFlowConcurrent runs two users in parallel.
func TestFlowConcurrent(t *testing.T) {
	rec := &stepRecorder{}
	flow := NewFlow[string]()
	flow.AddStep(Step[string]{
		Name: "s1",
		Prompt: func(ctx *Context, data *string) error {
			rec.add("prompt:" + fmt.Sprint(ctx.ChatID()))
			return nil
		},
		Process: func(ctx *Context, data *string) (FlowAction, error) {
			rec.add("process:" + fmt.Sprint(ctx.ChatID()))
			return End(), nil
		},
	})

	var wg sync.WaitGroup
	for i := int64(1); i <= 10; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			if err := flow.HandleUpdate(testContext(id, id)); err != nil {
				t.Errorf("user %d start: %v", id, err)
			}
			if err := flow.HandleUpdate(testContext(id, id)); err != nil {
				t.Errorf("user %d process: %v", id, err)
			}
		}(i)
	}
	wg.Wait()

	// All 10 users should have completed
	for i := int64(1); i <= 10; i++ {
		if flow.IsUserInFlow(i, i) {
			t.Errorf("user %d should not be in flow after End", i)
		}
	}
}
