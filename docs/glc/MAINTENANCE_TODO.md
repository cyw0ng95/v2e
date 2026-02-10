# MAINTENANCE TODO

## Priority 1: HIGH - Fix STIX Import Test Failures

### Task: Fix validateSTIX function to handle all test scenarios correctly

**Status**: IN PROGRESS
**Estimated Time**: 15 minutes
**Dependencies**: None

**Steps**:
1. Analyze failing tests
2. Fix validateSTIX to handle all edge cases
3. Ensure statistics tracking works correctly
4. Commit and push

---

## Priority 2: HIGH - Fix Dark Mode Toggle

### Task: Implement dark mode toggle button functionality

**Status**: PENDING
**Estimated Time**: 30 minutes
**Dependencies**: Theme system

**Steps**:
1. Add theme toggle to canvas toolbar
2. Persist theme preference
3. Update components to respect theme
4. Test theme switching
5. Commit and push

---

## Priority 3: MEDIUM - Increase Test Coverage to 90%

### Task: Add component tests for missing coverage

**Status**: PENDING
**Estimated Time**: 2 hours
**Dependencies**: @testing-library/react

**Steps**:
1. Install @testing-library/react
2. Add component tests for DynamicNode, NodePalette, CanvasToolbar, InferencePanel
3. Fix component test issues
4. Increase coverage to >90%
5. Commit and push

---

## Priority 4: LOW - Add Example Graphs

### Task: Create example graphs for D3FEND and Topo-Graph

**Status**: PENDING
**Estimated Time**: 1 hour

**Steps**:
1. Create D3FEND example graph
2. Create Topo-Graph example graph
3. Add to example gallery
4. Commit and push

---

## Completed Work (2026-02-10)

### Features Implemented
- ✅ D3FEND Inference Engine (541 lines)
- ✅ STIX 2.1 Import (1,759 lines)
- ✅ Component Tests (402 lines)
- ✅ TypeScript Errors Fixed (all resolved)
- ✅ ESLint Warnings Fixed (all resolved)
- ✅ Documentation Updated (IMPLEMENTATION.md, TESTING_SUMMARY)

### Remaining Work

1. **STIX Import Tests** (2 failing tests)
   - Fix `validateSTIX` to handle all test scenarios correctly
   - Ensure result structure matches test expectations

2. **Dark Mode Toggle** (30 minutes)
   - Implement theme toggle mechanism
   - Persist theme preference
   - Update all components

3. **Test Coverage** (2 hours)
   - Add @testing-library/react
   - Create component tests
   - Increase from 80% to >90%

4. **Example Graphs** (1 hour)
   - Create example graphs
   - Add to example gallery

5. **Smart Edge Routing** (1 hour, LOW priority)
   - A* pathfinding
   - Obstacle detection
   - User toggle

---

## Next Steps

1. Fix STIX import test failures
2. Implement dark mode toggle
3. Add component tests to reach 90% coverage
4. Create example graphs

---

**Note**: All work should be done incrementally with frequent commits. Don't batch large changes.
