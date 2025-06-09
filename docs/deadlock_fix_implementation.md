# Deadlock Fix Implementation - Separate Mutexes Solution

## Problem Statement

During the refactoring process (Task Group 3), we added a single mutex (`muUserFlows`) to the `flowManager` for thread safety. However, this created a deadlock scenario:

1. **Step Processing**: `flowManager.HandleUpdate()` acquires a write lock on the mutex
2. **SetFlowData Call**: Within the same goroutine, `ctx.SetFlowData()` calls `flowManager.setUserFlowData()` which tries to acquire another write lock
3. **Deadlock**: Since the same goroutine already holds the write lock, it deadlocks waiting for itself

## Solution Implemented

We implemented a **separate mutexes strategy** combined with **strategic lock release** to resolve the deadlock:

### 1. Separate Mutexes

Modified the `flowManager` struct to use two distinct mutexes:

```go
type flowManager struct {
    flows          map[string]*Flow
    userFlows      map[int64]*userFlowState
    flowStateMutex sync.RWMutex // For flow state changes (start/cancel/step transitions)
    flowDataMutex  sync.RWMutex // For flow data operations (SetFlowData/GetFlowData)
    flowConfig     *FlowConfig
    // ... other fields
}
```

### 2. Mutex Assignment Strategy

**Flow State Operations** (use `flowStateMutex`):
- `startFlow()`
- `cancelFlow()`
- `isUserInFlow()`
- Step transitions in `HandleUpdate()`

**Flow Data Operations** (use `flowDataMutex`):
- `setUserFlowData()`
- `getUserFlowData()`

### 3. Strategic Lock Release in HandleUpdate

The key insight was that `HandleUpdate()` was holding the `flowStateMutex` during the entire `ProcessFunc` execution, which could call `SetFlowData()`. We modified `HandleUpdate()` to:

1. **Acquire lock** to read flow state and validate
2. **Release lock** before calling `ProcessFunc`
3. **Re-acquire lock** for final state modifications

```go
func (fm *flowManager) HandleUpdate(ctx *Context) (bool, error) {
    // Acquire lock to get flow state info
    fm.flowStateMutex.Lock()
    
    // ... validation and setup ...
    
    // Release the lock before calling ProcessFunc to avoid deadlock
    fm.flowStateMutex.Unlock()

    // Call ProcessFunc without holding any locks
    result := currentStep.ProcessFunc(ctx, input, buttonClick)
    
    // ... handle callbacks ...
    
    // Re-acquire lock for state modifications
    fm.flowStateMutex.Lock()
    defer fm.flowStateMutex.Unlock()
    
    // ... final state updates ...
}
```

### 4. Race Condition Safety

To handle potential race conditions during the unlocked period:
- Re-check that user is still in flow after re-acquiring the lock
- Gracefully handle cases where flow was cancelled during ProcessFunc execution

## Testing

Implemented comprehensive concurrent tests to validate the solution:

### TestConcurrentFlowDataOperations
Tests the specific deadlock scenario:
- `HandleUpdate()` calls `ProcessFunc` which calls `SetFlowData()`
- Verifies no deadlock occurs
- Validates data operations work correctly

### TestConcurrentFlowDataAccess
Tests concurrent access from multiple goroutines:
- 10 goroutines performing 100 data set operations each
- 10 goroutines performing 100 data get operations each
- 5-second timeout to detect deadlocks

### TestMixedMutexOperations
Tests interleaving of flow state and data operations:
- 5 goroutines mixing `isUserInFlow()`, `setUserFlowData()`, and `getUserFlowData()` calls
- Validates both mutex types can be used safely together

## Results

✅ **All tests pass** - No deadlocks detected
✅ **Backward compatibility** - Existing functionality unchanged
✅ **Thread safety** - Concurrent operations work correctly
✅ **Performance** - No significant performance impact

## Implementation Details

### Files Modified
- `core/flow.go` - Updated mutex strategy and HandleUpdate method
- `core/flow_manager_test.go` - Updated test references to new mutex names
- `core/concurrent_flow_test.go` - Added comprehensive concurrent tests

### Key Changes
1. Replaced single `muUserFlows` with `flowStateMutex` and `flowDataMutex`
2. Updated all method locks to use appropriate mutex
3. Modified `HandleUpdate()` to release lock during ProcessFunc execution
4. Added re-validation after lock re-acquisition

## Future Considerations

1. **Monitor Performance**: While current tests show no issues, monitor production performance
2. **Lock Ordering**: Maintain consistent lock ordering if future changes require holding multiple locks
3. **Documentation**: Update flow processing documentation to reflect the new locking strategy

## Conclusion

The separate mutexes solution successfully resolves the deadlock issue while maintaining thread safety and backward compatibility. The strategic lock release in `HandleUpdate()` allows `ProcessFunc` to safely call flow data operations without creating circular lock dependencies.