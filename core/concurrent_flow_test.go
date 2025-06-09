package teleflow

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentFlowDataOperations tests that flow data operations don't deadlock
// when performed concurrently during flow processing
func TestConcurrentFlowDataOperations(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()

	// Create a flow that will call SetFlowData during processing
	flow := &Flow{
		Name: "concurrent-test-flow",
		Steps: map[string]*flowStep{
			"step1": {
				Name: "step1",
				PromptConfig: &PromptConfig{
					Message: "Test step",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					// This simulates the deadlock scenario: SetFlowData being called
					// from within a step's ProcessFunc while HandleUpdate holds a lock
					err := ctx.SetFlowData("concurrent_key", "concurrent_value")
					if err != nil {
						t.Errorf("SetFlowData failed: %v", err)
					}

					// Also test GetFlowData
					value, exists := ctx.GetFlowData("concurrent_key")
					if !exists || value != "concurrent_value" {
						t.Errorf("GetFlowData failed: expected 'concurrent_value', got %v (exists: %v)", value, exists)
					}

					return ProcessResult{Action: actionCompleteFlow}
				},
			},
		},
		Order: []string{"step1"},
		OnComplete: func(ctx *Context) error {
			return nil
		},
	}

	fm.registerFlow(flow)

	userID := int64(12345)

	// Start the flow
	ctx := createFlowTestContext(userID, "test input")
	// Set the flowOps so SetFlowData/GetFlowData work
	ctx.flowOps = fm
	err := fm.startFlow(userID, "concurrent-test-flow", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	// Test that HandleUpdate (which holds flowStateMutex) can call ProcessFunc
	// which in turn calls SetFlowData/GetFlowData (which use flowDataMutex)
	// This would previously deadlock with a single mutex

	// Create a new context for HandleUpdate
	updateCtx := createFlowTestContext(userID, "test input")
	updateCtx.flowOps = fm

	handled, err := fm.HandleUpdate(updateCtx)

	if err != nil {
		t.Fatalf("HandleUpdate failed: %v", err)
	}

	if !handled {
		t.Error("Expected update to be handled")
	}

	// Verify the flow completed (user should no longer be in flow)
	if fm.isUserInFlow(userID) {
		t.Error("User should not be in flow after completion")
	}
}

// TestConcurrentFlowDataAccess tests concurrent access to flow data from multiple goroutines
func TestConcurrentFlowDataAccess(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()

	// Create a simple flow for testing
	flow := &Flow{
		Name: "data-access-test",
		Steps: map[string]*flowStep{
			"step1": {
				Name:         "step1",
				PromptConfig: &PromptConfig{Message: "Test"},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					return ProcessResult{Action: actionNextStep}
				},
			},
		},
		Order: []string{"step1"},
	}

	fm.registerFlow(flow)

	userID := int64(12345)
	ctx := createFlowTestContext(userID, "test")

	// Start the flow
	err := fm.startFlow(userID, "data-access-test", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	// Test concurrent data access
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // For both setters and getters

	// Start multiple goroutines that set data
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", index, j)
				value := fmt.Sprintf("value_%d_%d", index, j)
				err := fm.setUserFlowData(userID, key, value)
				if err != nil {
					t.Errorf("setUserFlowData failed: %v", err)
				}
			}
		}(i)
	}

	// Start multiple goroutines that get data
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", index, j)
				// Try to get the data (might not exist yet, which is fine)
				_, _ = fm.getUserFlowData(userID, key)
			}
		}(i)
	}

	// Wait for all operations to complete
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	// Set a timeout to detect deadlocks
	select {
	case <-done:
		// All operations completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Concurrent data access test timed out - possible deadlock")
	}
}

// TestMixedMutexOperations tests that we can safely mix flow state and data operations
func TestMixedMutexOperations(t *testing.T) {
	fm, _, _, _ := createTestFlowManager()

	flow := &Flow{
		Name: "mixed-ops-test",
		Steps: map[string]*flowStep{
			"step1": {
				Name:         "step1",
				PromptConfig: &PromptConfig{Message: "Test"},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					return ProcessResult{Action: actionNextStep}
				},
			},
		},
		Order: []string{"step1"},
	}

	fm.registerFlow(flow)

	userID := int64(12345)
	ctx := createFlowTestContext(userID, "test")

	err := fm.startFlow(userID, "mixed-ops-test", ctx)
	if err != nil {
		t.Fatalf("Failed to start flow: %v", err)
	}

	// Test that we can interleave flow state and data operations without deadlock
	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			// Mix different types of operations
			for j := 0; j < 50; j++ {
				switch j % 4 {
				case 0:
					// Flow state operation
					_ = fm.isUserInFlow(userID)
				case 1:
					// Flow data operation
					key := fmt.Sprintf("mixed_key_%d_%d", index, j)
					_ = fm.setUserFlowData(userID, key, "value")
				case 2:
					// Flow data operation
					key := fmt.Sprintf("mixed_key_%d_%d", index, j-1)
					_, _ = fm.getUserFlowData(userID, key)
				case 3:
					// Flow state operation (reading user flows map)
					_ = fm.isUserInFlow(userID)
				}
			}
		}(i)
	}

	// Wait for completion with timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Mixed mutex operations test timed out - possible deadlock")
	}
}
