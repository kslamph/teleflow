package teleflow

import (
	"testing"
)

// TestComprehensiveMiddlewareDataIsolation tests that middleware context data
// and flow data remain properly isolated throughout flow execution
func TestComprehensiveMiddlewareDataIsolation(t *testing.T) {
	// Create test components using the standard pattern
	fm, _, _, _ := createTestFlowManager()

	// Track values to verify isolation
	var stepInitialKeyFromFlow, stepRuntimeKeyFromFlow string
	var stepInitialKeyFromContext, stepRuntimeKeyFromContext string
	var middlewareUpdatedInitial, middlewareSetRuntime bool

	// Create a test middleware that sets initial context data and updates it during flow
	testMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			// Set initial context data before flow processing
			ctx.Set("middleware_initial_key", "initial_value")

			// Call the next handler (which will process the flow)
			err := next(ctx)

			// Simulate middleware updating context data during/after flow processing
			ctx.Set("middleware_initial_key", "updated_context_value")
			ctx.Set("middleware_runtime_key", "runtime_value")
			middlewareUpdatedInitial = true
			middlewareSetRuntime = true

			return err
		}
	}

	// Create a flow that tests data isolation
	flow := &Flow{
		Name: "middleware-data-isolation-test",
		Steps: map[string]*flowStep{
			"test_step": {
				Name: "test_step",
				PromptConfig: &PromptConfig{
					Message: "Testing middleware data isolation",
				},
				ProcessFunc: func(ctx *Context, input string, buttonClick *ButtonClick) ProcessResult {
					// Simulate middleware updating context data during flow execution
					ctx.Set("middleware_initial_key", "updated_context_value")
					ctx.Set("middleware_runtime_key", "runtime_value")

					// Test 1: Access initial key via GetFlowData - should show seeded value
					if val, exists := ctx.GetFlowData("middleware_initial_key"); exists {
						stepInitialKeyFromFlow = val.(string)
					}

					// Test 2: Access runtime key via GetFlowData - should NOT exist
					if val, exists := ctx.GetFlowData("middleware_runtime_key"); exists {
						stepRuntimeKeyFromFlow = val.(string)
					}

					// Test 3: Access data via context Get() - should show current context values
					if val, exists := ctx.Get("middleware_initial_key"); exists {
						stepInitialKeyFromContext = val.(string)
					}

					if val, exists := ctx.Get("middleware_runtime_key"); exists {
						stepRuntimeKeyFromContext = val.(string)
					}

					return CompleteFlow()
				},
			},
		},
		Order: []string{"test_step"},
	}

	fm.registerFlow(flow)

	userID := int64(123)

	// Create context using the standard test pattern
	ctx := createFlowTestContext(userID, "test input", nil)
	ctx.flowOps = fm // Set flow operations

	// Create wrapped handler that starts flow and processes it
	flowHandler := func(ctx *Context) error {
		// Start the flow (this seeds initial context data into flow data)
		if err := ctx.StartFlow("middleware-data-isolation-test"); err != nil {
			return err
		}

		// Process the step
		_, err := fm.HandleUpdate(ctx)
		return err
	}

	// Wrap with middleware
	wrappedHandler := testMiddleware(flowHandler)

	// Execute the wrapped handler
	err := wrappedHandler(ctx)
	if err != nil {
		t.Fatalf("Handler execution failed: %v", err)
	}

	// Verify middleware executed its updates
	if !middlewareUpdatedInitial {
		t.Error("Middleware should have updated initial key")
	}
	if !middlewareSetRuntime {
		t.Error("Middleware should have set runtime key")
	}

	// **CRITICAL ASSERTIONS for Data Isolation**

	// Assertion 1: GetFlowData("middleware_initial_key") should return the seeded value,
	// NOT the middleware's updated context value
	expectedFlowInitial := "initial_value" // Value seeded when flow started
	if stepInitialKeyFromFlow != expectedFlowInitial {
		t.Errorf("GetFlowData('middleware_initial_key') should return seeded value '%s', got '%s'",
			expectedFlowInitial, stepInitialKeyFromFlow)
	}

	// Assertion 2: GetFlowData("middleware_runtime_key") should NOT find the key
	// (runtime context data set by middleware should not leak into flow data)
	if stepRuntimeKeyFromFlow != "" {
		t.Errorf("GetFlowData('middleware_runtime_key') should not find the key, but got '%s'",
			stepRuntimeKeyFromFlow)
	}

	// Assertion 3: Get("middleware_initial_key") should return the middleware's updated value
	expectedContextInitial := "updated_context_value"
	if stepInitialKeyFromContext != expectedContextInitial {
		t.Errorf("Get('middleware_initial_key') should return updated context value '%s', got '%s'",
			expectedContextInitial, stepInitialKeyFromContext)
	}

	// Assertion 4: Get("middleware_runtime_key") should return the middleware's runtime value
	expectedContextRuntime := "runtime_value"
	if stepRuntimeKeyFromContext != expectedContextRuntime {
		t.Errorf("Get('middleware_runtime_key') should return runtime context value '%s', got '%s'",
			expectedContextRuntime, stepRuntimeKeyFromContext)
	}

	t.Logf("SUCCESS: Data isolation verified")
	t.Logf("  - Flow data 'middleware_initial_key': '%s' (seeded value, isolated from context updates)", stepInitialKeyFromFlow)
	t.Logf("  - Flow data 'middleware_runtime_key': '%s' (empty, runtime context data didn't leak)", stepRuntimeKeyFromFlow)
	t.Logf("  - Context data 'middleware_initial_key': '%s' (updated by middleware)", stepInitialKeyFromContext)
	t.Logf("  - Context data 'middleware_runtime_key': '%s' (set by middleware)", stepRuntimeKeyFromContext)
}
