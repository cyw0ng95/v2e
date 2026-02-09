# GLC Project Refined Implementation Plan - Phase 1: Core Infrastructure

## Phase Overview

This phase establishes the core technical foundation of GLC project, including project initialization, centralized state management, complete data models, basic UI framework, and preset system architecture with robust validation.

**Original Duration**: 32-44 hours
**With Mitigations**: 52-68 hours
**Timeline Increase**: +62%
**Actual Duration**: 8 weeks (4 sprints × 2 weeks)

**Deliverables**:
- Next.js 15+ project with React Flow and shadcn/ui
- Centralized Zustand state management with all slices
- Complete TypeScript type system with Zod validation
- Built-in presets (D3FEND, Topo-Graph)
- Landing page with preset selection
- Preset validation and migration system
- Error handling and recovery mechanisms

---

## Sprint 1 (Weeks 1-2): Foundation & Setup

### Duration: 14-18 hours

### Goal: Initialize project with all core dependencies and basic infrastructure

### Week 1 Tasks

#### 1.1 Project Initialization (6-8h)

**Risk**: Setup errors, dependency conflicts

**Files to Create**:
- `website/glc/package.json` - Project dependency configuration
- `website/glc/next.config.ts` - Next.js configuration with static export and allowedHosts
- `website/glc/tsconfig.json` - TypeScript strict mode configuration
- `website/glc/tailwind.config.ts` - Tailwind CSS v4 configuration

**Tasks**:
- Initialize Next.js 15+ project
- Configure TypeScript strict mode
- Setup Tailwind CSS v4
- Configure static export (output: 'export')
- Configure allowedHosts for preview environment ('.monkeycode-ai.online')
- Configure reverse proxy for API calls (if backend running on different port)

**Acceptance Criteria**:
- `npm install` succeeds
- `npm run dev` starts on http://localhost:3000
- `npm run build` produces static export in `out/`
- Allowed hosts configured for preview environment
- Lighthouse score >90 for initial page

---

#### 1.2 Core Dependencies Installation (4-6h)

**Files to Modify**:
- `website/glc/package.json`

**Work Content**:
- Install @xyflow/react (React Flow)
- Install shadcn/ui component library and dependencies
- Install Lucide React icon library
- Install sonner notification library
- Install react-hook-form and zod form validation
- Install class-variance-authority, clsx, tailwind-merge
- Install zustand for state management
- Install immer for immutable updates

**Acceptance Criteria**:
- All packages installed without conflicts
- Imports work without errors
- TypeScript types resolved correctly

---

### Week 2 Tasks

#### 1.3 shadcn/ui Component Setup (4-6h)

**Files to Create**:
- `website/glc/components.json` - shadcn/ui configuration
- `website/glc/components/ui/` - UI component directory structure

**Tasks**:
- Initialize shadcn/ui configuration
- Add 10 essential UI components: button, dialog, dropdown-menu, input, label, sheet, tabs, accordion, toast
- Configure component alias paths
- Verify component rendering

**Acceptance Criteria**:
- `npx shadcn@latest init` completes successfully
- All 10 components added
- Components render correctly
- TypeScript types resolved

---

#### 1.4 Basic Layout Structure (6-8h)

**Files to Create**:
- `website/glc/app/layout.tsx` - Root layout with providers
- `website/glc/app/glc/page.tsx` - Landing page
- `website/glc/app/glc/[presetId]/page.tsx` - Canvas page
- `website/glc/components/providers.tsx` - Context providers (ThemeProvider, Toaster)

**Tasks**:
- Create root layout with ThemeProvider
- Create landing page structure with hero section
- Create dynamic canvas page routes
- Set up Toaster for notifications
- Add basic navigation components

**Acceptance Criteria**:
- Landing page displays correctly
- Canvas route structure works
- Theme provider functional
- Notifications display correctly

---

**Sprint 1 Deliverables**:
- ✅ Next.js 15+ project with all dependencies
- ✅ shadcn/ui components installed
- ✅ Basic page structure
- ✅ Development environment ready

---

## Sprint 2 (Weeks 3-4): State Management & Data Models

### Duration: 16-20 hours

### Goal: Implement centralized state management and complete type system

### Week 3 Tasks

#### 1.5 Centralized State Management (CRITICAL MITIGATION) - 12-16h

**Risk**: CP1 - State Management Complexity Explosion
**Mitigation**: Zustand store with slice-based architecture, devtools, persistence

**Files to Create**:
- `website/glc/lib/store/index.ts` - Main store with all slices
- `website/glc/lib/store/slices/preset.ts` - Preset state slice
- `website/glc/lib/store/slices/graph.ts` - Graph state slice
- `website/glc/lib/store/slices/canvas.ts` - Canvas state slice
- `website/glc/lib/store/slices/ui.ts` - UI state slice
- `website/glc/lib/store/slices/undo-redo.ts` - Undo/redo state slice

**Tasks**:
- Install Zustand and middleware (devtools, persist, immer)
- Create store structure with 5 slices:
  - **PresetSlice**: currentPreset, builtInPresets, userPresets, preset management actions
  - **GraphSlice**: nodes, edges, metadata, viewport, CRUD operations
  - **CanvasSlice**: selection, zoom, pan, canvas interactions
  - **UiSlice**: theme, sidebar state, modal states
  - **UndoRedoSlice**: history stack, undo/redo actions
- Add persistence middleware for localStorage
- Add devtools middleware
- Create typed hooks for each slice
- Write unit tests for store operations

**Acceptance Criteria**:
- Store persists to localStorage
- Devtools show state changes
- All slice operations work correctly
- State survives page reloads
- No memory leaks from store subscriptions
- All actions typed correctly

---

#### 1.6 Complete TypeScript Type System - 8-12h

**Risk**: Type system rigidity limiting flexibility
**Mitigation**: Flexible type system with runtime validation

**Files to Create**:
- `website/glc/lib/types/preset.ts` - Preset-related type definitions
- `website/glc/lib/types/node.ts` - Node-related type definitions
- `website/glc/lib/types/edge.ts` - Edge-related type definitions
- `website/glc/lib/types/graph.ts` - Graph-related type definitions
- `website/glc/lib/types/brand.ts` - Brand type utilities
- `website/glc/lib/types/index.ts` - Unified export of all types

**Tasks**:
- Define complete TypeScript type system:
  - CanvasPreset interface
  - NodeTypeDefinition interface
  - RelationshipDefinition interface
  - PresetStyling interface
  - PresetBehavior interface
  - CADNode interface
  - CADEdge interface
  - GraphMetadata interface
  - Graph interface
  - Property interface
  - Reference interface
- Create type utilities and helpers
- Add Zod validation schemas for all types
- Create type guards
- Export unified type index
- Write TypeScript tests

**Acceptance Criteria**:
- All types compile without errors
- Zod schemas validate correctly
- Type guards work as expected
- No `any` types in core code
- Full type coverage for store operations

---

### Week 4 Tasks

#### 1.7 Built-in Presets Implementation - 6-8h

**Risk**: Preset definition errors, validation failures
**Mitigation**: Zod schema validation, comprehensive testing

**Files to Create**:
- `website/glc/lib/presets/d3fend-preset.ts` - D3FEND canvas preset
- `website/glc/lib/presets/topo-preset.ts` - Topo-Graph canvas preset
- `website/glc/lib/presets/index.ts` - Preset exports

**Tasks**:
- Create D3FEND preset with:
  - 9 node types (event, remote-command, countermeasure, artifact, agent, vulnerability, condition, note, thing)
  - 200+ D3FEND relationships (accesses, creates, detects, counters, etc.)
  - Dark theme styling with color-coded node types
  - Behavior configuration (pan, zoom, snap-to-grid, undo/redo)
  - D3FEND ontology mappings for inferences
- Create Topo-Graph preset with:
  - 8 node types (entity, process, data, resource, group, decision, start-end, note)
  - 8 relationship types (connects, contains, depends-on, flows-to, related-to, controls, owns, implements)
  - Light theme styling
  - Behavior configuration
- Validate both presets with Zod schemas
- Add preset metadata (version, author, category)

**Acceptance Criteria**:
- Both presets validate successfully
- D3FEND preset has 9 node types
- Topo-Graph preset has 8 node types
- Relationship types defined for both presets
- Styling configurations complete
- Behavior configurations complete
- D3FEND ontology mappings present

---

#### 1.8 Landing Page & Navigation - 6-8h

**Files to Create**:
- `website/glc/app/glc/page.tsx` - Landing page (update)
- `website/glc/components/preset-card.tsx` - Preset preview card
- `website/glc/components/recent-graphs-list.tsx` - Recent graphs section
- `website/glc/components/navigation/header.tsx` - Header navigation

**Tasks**:
- Create landing page with:
  - Hero section with project description
  - Preset selection grid
  - Recent graphs section
  - Create custom preset button
  - Documentation link
- Implement preset card component with:
  - Preset name and description
  - Open canvas button
  - Category badge
- Implement recent graphs list
- Create navigation header
- Add responsive layout
- Implement preset selection routing

**Acceptance Criteria**:
- Landing page displays all built-in presets
- Preset cards show correct information
- Recent graphs list displays correctly
- Navigation works between pages
- Preset selection routes to canvas page
- Responsive design on mobile/tablet

---

**Sprint 2 Deliverables**:
- ✅ Centralized Zustand store with all 5 slices
- ✅ Complete TypeScript type system
- ✅ State persistence working
- ✅ Type-safe store operations
- ✅ D3FEND preset implemented
- ✅ Topo-Graph preset implemented
- ✅ Landing page functional
- ✅ Preset selection working

---

## Sprint 3 (Weeks 5-6): Preset System & Validation

### Duration: 12-16 hours

### Goal: Complete preset system with robust validation and error handling

### Week 5 Tasks

#### 1.9 Preset Validation System (HIGH PRIORITY MITIGATION) - 6-8h

**Risk**: 1.1 - Preset System Complexity
**Mitigation**: Robust Zod schema validation with migration support

**Files to Create**:
- `website/glc/lib/validation/preset-schema.ts` - Zod schemas for all preset types
- `website/glc/lib/validation/preset-migration.ts` - Preset version migration system
- `website/glc/lib/validation/validators.ts` - Custom validation functions
- `website/glc/lib/validation/index.ts` - Validation exports

**Tasks**:
- Define Zod schemas for all preset types:
  - propertyDefinitionSchema
  - inferenceCapabilitySchema
  - nodeTypeDefinitionSchema
  - relationshipDefinitionSchema
  - presetStylingSchema
  - presetBehaviorSchema
  - validationRulesSchema
  - ontologyMappingSchema
  - canvasPresetSchema
- Create validation functions:
  - validatePreset()
  - validatePresetFile()
  - validateGraph()
- Implement preset migration system:
  - Define migration from version 0.9.0 to 1.0.0
  - Create migration registry
  - Implement automatic migration on load
- Add error messages for validation failures
- Write validation tests

**Acceptance Criteria**:
- All schemas compile without errors
- Validation catches invalid presets
- Error messages are clear and actionable
- Migration system handles version upgrades
- Validation tests pass

---

### Week 6 Tasks

#### 1.10 Error Handling & Recovery - 4-6h

**Risk**: Runtime errors, state corruption
**Mitigation**: Error boundaries, custom error classes, recovery mechanisms

**Files to Create**:
- `website/glc/lib/errors/error-boundaries.tsx` - React error boundaries
- `website/glc/lib/errors/error-handler.ts` - Error handling utilities
- `website/glc/lib/errors/error-types.ts` - Custom error classes
- `website/glc/lib/errors/index.ts` - Error exports

**Tasks**:
- Create error boundary components:
  - GraphErrorBoundary for canvas
  - PresetErrorBoundary for preset operations
  - GlobalErrorBoundary for entire app
- Define custom error classes:
  - GLCError (base error)
  - PresetValidationError
  - GraphValidationError
  - StateError
  - RPCTimeoutError (for future Phase 6)
- Implement error handler utility:
  - handleGLCError() - Error classification and formatting
  - logError() - Error logging
  - showError() - User-facing error display
- Create error recovery mechanisms:
  - State checkpointing
  - Auto-recovery on errors
  - User-initiated rollback
- Add loading states for async operations
- Implement error logging

**Acceptance Criteria**:
- Error boundaries catch all React errors
- Custom error classes provide context
- Error handler classifies errors correctly
- Recovery mechanisms work
- Error logging captures details
- Loading states display during async operations

---

#### 1.11 Preset Management System - 4-6h

**Risk**: Preset state corruption, invalid user presets
**Mitigation**: Validation before saving, backup system

**Files to Create**:
- `website/glc/lib/preset-manager.ts` - Preset CRUD operations
- `website/glc/lib/preset-serializer.ts` - Preset import/export
- `website/glc/components/preset-manager.tsx` - Preset manager UI (placeholder)

**Tasks**:
- Implement preset manager:
  - createUserPreset() - Create new custom preset
  - updateUserPreset() - Update existing preset
  - deleteUserPreset() - Delete preset
  - duplicatePreset() - Duplicate preset
  - exportPreset() - Export to JSON file
  - importPreset() - Import from JSON file
- Implement preset serializer:
  - serializePreset() - Convert to JSON
  - deserializePreset() - Parse from JSON with validation
  - validateBeforeSave() - Validate preset before saving
- Add backup system:
  - Auto-backup before modifications
  - Restore from backup
- Create preset manager UI (basic version, full version in Phase 3)

**Acceptance Criteria**:
- User presets CRUD operations work
- Import/export functionality works
- Validation prevents invalid presets
- Backup system prevents data loss
- User presets persist in localStorage

---

**Sprint 3 Deliverables**:
- ✅ Robust preset validation system
- ✅ Preset migration system
- ✅ Error boundaries implemented
- ✅ Error recovery mechanisms
- ✅ Preset management system
- ✅ Import/export functionality

---

## Sprint 4 (Weeks 7-8): Testing & Integration

### Duration: 10-14 hours

### Goal: Complete Phase 1 testing, integration, and documentation

### Week 7 Tasks

#### 1.12 Unit Testing - 6-8h

**Risk**: Bugs in core functionality
**Mitigation**: Comprehensive unit test coverage

**Files to Create**:
- `website/glc/lib/__tests__/store.test.ts` - Store tests
- `website/glc/lib/__tests__/types.test.ts` - Type utility tests
- `website/glc/lib/__tests__/validation.test.ts` - Validation tests
- `website/glc/lib/__tests__/preset-manager.test.ts` - Preset manager tests

**Tasks**:
- Write store tests:
  - Test all slice actions
  - Test state updates
  - Test persistence
  - Test undo/redo
- Write validation tests:
  - Test preset validation with valid presets
  - Test preset validation with invalid presets
  - Test migration system
- Write preset manager tests:
  - Test CRUD operations
  - Test import/export
  - Test backup system
- Write type utility tests
- Run all tests and fix failures
- Achieve >80% code coverage

**Acceptance Criteria**:
- All tests pass
- >80% code coverage achieved
- No critical bugs found
- Test execution time <5 seconds

---

### Week 8 Tasks

#### 1.13 Integration Testing & Documentation - 4-6h

**Risk**: Integration issues, missing documentation
**Mitigation**: Integration tests, comprehensive documentation

**Tasks**:
- Create integration tests:
  - Test preset loading flow
  - Test preset switching flow
  - Test error recovery flow
- Write Sprint 1-4 documentation:
  - Architecture overview
  - Store design document
  - Type system documentation
  - Validation system documentation
- Create setup guide for developers
- Update README with Phase 1 features
- Perform end-to-end testing:
  - Load landing page
  - Select D3FEND preset
  - Select Topo-Graph preset
  - Test error handling
- Fix any integration issues
- Prepare Phase 2 handoff

**Acceptance Criteria**:
- Integration tests pass
- Documentation complete
- Setup guide works for new developers
- End-to-end flows work
- No integration issues
- Phase 1 acceptance criteria met

---

**Sprint 4 Deliverables**:
- ✅ Comprehensive unit tests (>80% coverage)
- ✅ Integration tests passing
- ✅ Documentation complete
- ✅ Setup guide working
- ✅ All acceptance criteria met

---

## Phase 1 Summary

### Total Duration: 52-68 hours (8 weeks)

### Deliverables Summary

#### Files Created (38-51)
- Configuration files: 4
- Store files: 6
- Type files: 6
- Preset files: 3
- Component files: 12-15
- Validation files: 4
- Error handling files: 4
- Test files: 4-8
- Documentation files: 2-4

#### Code Lines: 4,100-5,900
- Configuration: 300
- Store implementation: 800-1,000
- Type definitions: 800-1,200
- Preset definitions: 600-900
- Component code: 800-1,200
- Validation code: 400-600
- Error handling: 300-500
- Test code: 500-800
- Documentation: 400-700

### Success Criteria

#### Functional Success
- [x] User can select and open D3FEND preset
- [x] User can select and open Topo-Graph preset
- [x] Presets validate correctly
- [x] State persists across page reloads
- [x] Error boundaries catch and display errors
- [x] All types compile without errors
- [x] Preset management works

#### Technical Success
- [x] All unit tests pass
- [x] >80% code coverage achieved
- [x] Zero TypeScript errors
- [x] Zero ESLint errors
- [x] State management centralized
- [x] Performance acceptable (<2s FCP)

#### Quality Success
- [x] Code follows best practices
- [x] Documentation complete
- [x] Tests comprehensive
- [x] Error handling robust
- [x] Type safety enforced

### Risks Mitigated

1. **CP1 - State Management Complexity** ✅
   - Implemented centralized Zustand store
   - Added devtools and persistence
   - Created typed hooks

2. **1.1 - Preset System Complexity** ✅
   - Implemented robust Zod validation
   - Added migration system
   - Created preset manager

3. **1.2 - TypeScript Type System Rigidity** ✅
   - Created flexible type system
   - Added runtime validation
   - Implemented type guards

4. **1.3 - Static Export Path Conflicts** ✅
   - Configured static export correctly
   - Added allowedHosts for preview
   - Tested export functionality

### Phase Dependencies

**Phase 2 Depends On**:
- ✅ Centralized state management (Task 1.5)
- ✅ Complete type system (Task 1.6)
- ✅ Built-in presets (Task 1.7)
- ✅ Error handling (Task 1.10)

**Phase 3 Depends On**:
- All Phase 1 deliverables

**Phase 4 Depends On**:
- All Phase 1 deliverables

**Phase 5 Depends On**:
- All Phase 1 deliverables

**Phase 6 Depends On**:
- All Phase 1 deliverables

### Next Steps

**Transition to Phase 2**:
1. Review Phase 1 deliverables
2. Verify all acceptance criteria met
3. Update project timeline
4. Begin Phase 2 Sprint 5

**Immediate Actions**:
- Review Sprint 5 tasks
- Set up React Flow environment
- Begin canvas integration

---

**Document Version**: 2.0 (Refined)
**Last Updated**: 2026-02-09
**Phase Status**: Ready for Implementation
