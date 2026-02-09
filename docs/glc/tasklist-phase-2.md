# GLC Project Implementation Plan - Phase 2: Core Canvas Features

## Phase Overview

This phase implements the core canvas functionality including the React Flow integration, node palette, canvas interactions, and basic node/edge operations. This brings the canvas to life with interactive capabilities.

## Task 2.1: React Flow Canvas Integration

### Change Estimation (File Level)
- New files: 6-8
- Modified files: 2-3
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~600-900

### Detailed Work Items

#### 2.1.1 React Flow Canvas Setup
**File List**:
- `website/glc/components/glc/canvas/react-flow-canvas.tsx` - Main React Flow wrapper
- `website/glc/lib/hooks/use-react-flow.ts` - React Flow instance hook

**Work Content**:
- Integrate @xyflow/react (React Flow) into the canvas
- Configure pan and zoom controls
- Setup mini-map component
- Configure background grid (preset-aware)
- Handle canvas resize

**Acceptance Criteria**:
1. WHEN the canvas loads, React Flow SHALL initialize with the configured grid
2. WHEN user drags the middle mouse button, canvas SHALL pan
3. WHEN user scrolls the mouse wheel, canvas SHALL zoom
4. WHEN mini-map is visible, it SHALL reflect the current canvas state
5. WHEN viewport changes, position/zoom state SHALL update correctly

#### 2.1.2 Canvas Controls Toolbar
**File List**:
- `website/glc/components/glc/canvas/canvas-controls.tsx` - Floating toolbar
- `website/glc/components/glc/canvas/zoom-controls.tsx` - Zoom controls

**Work Content**:
- Implement zoom in/out buttons
- Implement fit-to-screen button
- Implement zoom percentage display
- Implement lock/unlock toggle
- Implement grid visibility toggle
- Configure control panel positioning

**Acceptance Criteria**:
1. WHEN user clicks zoom-in button, zoom SHALL increase by 20%
2. WHEN user clicks zoom-out button, zoom SHALL decrease by 20%
3. WHEN user clicks fit-to-screen, all nodes SHALL fit in viewport
4. WHEN zoom changes, percentage display SHALL update in real-time
5. WHEN user toggles lock, canvas interactions SHALL be enabled/disabled
6. WHEN user toggles grid, grid SHALL show/hide

---

## Task 2.2: Node Palette Implementation

### Change Estimation (File Level)
- New files: 8-10
- Modified files: 3-4
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~800-1,200

### Detailed Work Items

#### 2.2.1 Draggable Node Components
**File List**:
- `website/glc/components/glc/palette/draggable-node.tsx` - Draggable node template
- `website/glc/components/glc/palette/node-category.tsx` - Node category accordion
- `website/glc/components/glc/palette/node-palette-sidebar.tsx` - Full sidebar

**Work Content**:
- Implement draggable node templates from preset data
- Create category-based accordion layout
- Add node preview with icon and label
- Implement drag-and-drop to canvas
- Add search/filter functionality
- Apply preset-specific colors and icons

**Acceptance Criteria**:
1. WHEN palette loads, SHALL display all node types from current preset
2. WHEN user drags a node, drag preview SHALL show with icon
3. WHEN user drops node on canvas, node SHALL be created at drop position
4. WHEN user types in search box, palette SHALL filter node types
5. WHEN user expands/collapses category, SHALL show/hide node types
6. WHEN preset switches, palette SHALL update to show new preset's nodes

#### 2.2.2 Node Type Preview
**File List**:
- `website/glc/components/glc/palette/node-preview.tsx` - Node preview card

**Work Content**:
- Create node preview component
- Show node icon, color, and label
- Display node type description
- Show example properties
- Add tooltip on hover

**Acceptance Criteria**:
1. WHEN user hovers over node in palette, preview SHALL show tooltip
2. WHEN tooltip displays, SHALL contain node type name and description
3. WHEN node has example properties, preview SHALL show property list
4. WHEN tooltip closes, SHALL happen smoothly without visual glitches

---

## Task 2.3: Canvas Node Implementation

### Change Estimation (File Level)
- New files: 10-12
- Modified files: 4-5
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~1,000-1,500

### Detailed Work Items

#### 2.3.1 Dynamic Node Component Factory
**File List**:
- `website/glc/components/glc/canvas/node-factory.tsx` - Node component factory
- `website/glc/components/glc/canvas/node-components/generic-node.tsx` - Generic node base

**Work Content**:
- Create factory function to generate node components from preset
- Implement generic node container with preset-aware styling
- Support custom colors, borders, and icons from preset
- Implement node selection state (border highlight)
- Implement hover states (show handles, quick actions)
- Add connection handles (source/target)

**Acceptance Criteria**:
1. WHEN node is created, SHALL use preset-defined styling
2. WHEN node is selected, SHALL show selection border
3. WHEN user hovers node, SHALL show connection handles
4. WHEN preset defines custom node color, node SHALL apply that color
5. WHEN node type has icon, icon SHALL render correctly
6. WHEN user drags node, movement SHALL be smooth

#### 2.3.2 Node Structure and Content
**File List**:
- `website/glc/components/glc/canvas/node-components/node-header.tsx` - Node header
- `website/glc/components/glc/canvas/node-components/node-label.tsx` - Node label
- `website/glc/components/glc/canvas/node-components/node-properties.tsx` - Node properties display
- `website/glc/components/glc/canvas/node-components/node-actions.tsx` - Quick action buttons

**Work Content**:
- Implement node header with icon and label
- Add inline label editing (double-click)
- Display node properties (key-value pairs)
- Add quick action buttons (duplicate, delete)
- Add D3FEND class indicator (for D3FEND preset)

**Acceptance Criteria**:
1. WHEN node renders, SHALL show icon and label in header
2. WHEN user double-clicks label, SHALL become editable input
3. WHEN user saves label change, SHALL update node data
4. WHEN node has properties, SHALL display them below header
5. WHEN user hovers node, SHALL show duplicate and delete buttons
6. WHEN node has D3FEND class, SHALL display class badge

#### 2.3.3 Node Details Sheet
**File List**:
- `website/glc/components/glc/canvas/node-details-sheet.tsx` - Node details panel

**Work Content**:
- Create slide-out sheet for node details
- Implement form fields for ID, label, D3FEND class
- Add property editor (add/edit/remove properties)
- Implement form validation
- Add save/cancel actions

**Acceptance Criteria**:
1. WHEN user selects node, details sheet SHALL open
2. WHEN user changes label and saves, node SHALL update
3. WHEN user adds property, property SHALL appear in node details
4. WHEN user deletes property, property SHALL be removed
5. WHEN form is invalid, SHALL show validation error
6. WHEN user cancels changes, form SHALL revert to original values

---

## Task 2.4: Edge (Relationship) Implementation

### Change Estimation (File Level)
- New files: 8-10
- Modified files: 3-4
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~700-1,000

### Detailed Work Items

#### 2.4.1 Dynamic Edge Component
**File List**:
- `website/glc/components/glc/canvas/edge-factory.tsx` - Edge component factory
- `website/glc/components/glc/canvas/edge-components/generic-edge.tsx` - Generic edge base

**Work Content**:
- Create factory function to generate edge components from preset
- Implement preset-aware edge styling (color, width, line style)
- Support different arrow styles (default, open, filled)
- Implement edge labels
- Handle edge selection state
- Apply preset-defined relationship styles

**Acceptance Criteria**:
1. WHEN edge is created, SHALL use preset-defined style
2. WHEN relationship type has custom color, edge SHALL use that color
3. WHEN relationship type is dashed, edge SHALL render dashed line
4. WHEN relationship defines arrow style, edge SHALL use correct arrow
5. WHEN edge has label, label SHALL display centered on edge
6. WHEN edge is selected, SHALL highlight with accent color

#### 2.4.2 Edge Creation Workflow
**File List**:
- `website/glc/lib/hooks/use-edge-creation.ts` - Edge creation hook
- `website/glc/components/glc/canvas/relationship-picker.tsx` - Relationship picker dialog

**Work Content**:
- Implement drag-to-connect from node handles
- Show relationship picker on connection creation
- Filter relationships by preset
- Validate edge restrictions (from/to node types)
- Apply relationship type to created edge

**Acceptance Criteria**:
1. WHEN user drags from node handle, SHALL show connection line
2. WHEN user drops on target node, SHALL open relationship picker
3. WHEN user selects relationship, edge SHALL be created with label
4. WHEN relationship has source restrictions, SHALL filter valid source nodes
5. WHEN relationship has target restrictions, SHALL filter valid target nodes
6. WHEN connection is invalid, SHALL show error message

#### 2.4.3 Edge Label Editor
**File List**:
- `website/glc/components/glc/canvas/edge-label-editor.tsx` - Edge label edit dialog

**Work Content**:
- Create dialog for editing edge labels
- Support changing relationship type
- Support custom label text
- Add relationship description display

**Acceptance Criteria**:
1. WHEN user clicks edge label, SHALL open label editor
2. WHEN user changes relationship type, edge SHALL update
3. WHEN user enters custom text, label SHALL use custom text
4. WHEN user saves changes, edge SHALL reflect new settings
5. WHEN user cancels, changes SHALL be reverted

---

## Task 2.5: Canvas State Management

### Change Estimation (File Level)
- New files: 6-8
- Modified files: 4-5
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~600-900

### Detailed Work Items

#### 2.5.1 Graph State Hook
**File List**:
- `website/glc/lib/hooks/use-graph-state.ts` - Graph state management
- `website/glc/lib/stores/graph-store.ts` - Zustand store (optional)

**Work Content**:
- Implement graph state (nodes, edges, viewport)
- Manage node/edge CRUD operations
- Handle undo/redo history
- Implement local storage persistence
- Add state validation

**Acceptance Criteria**:
1. WHEN node is added, SHALL appear in graph state
2. WHEN node is removed, SHALL be removed from graph state
3. WHEN edge is added, SHALL connect correct nodes
4. WHEN user performs undo, SHALL revert last change
5. WHEN user performs redo, SHALL reapply reverted change
6. WHEN graph state changes, SHALL persist to localStorage

#### 2.5.2 Selection and Multi-Select
**File List**:
- `website/glc/lib/hooks/use-selection.ts` - Selection management

**Work Content**:
- Implement single node/edge selection
- Implement multi-select (shift+click)
- Implement drag selection (box select)
- Handle selection state across canvas

**Acceptance Criteria**:
1. WHEN user clicks node, node SHALL be selected
2. WHEN user shift+clicks multiple nodes, all SHALL be selected
3. WHEN user drags selection box, nodes inside SHALL be selected
4. WHEN user clicks empty space, selection SHALL be cleared
5. WHEN selection changes, details sheet SHALL update to reflect selection

---

## Task 2.6: Canvas Interactions

### Change Estimation (File Level)
- New files: 4-5
- Modified files: 3-4
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~400-600

### Detailed Work Items

#### 2.6.1 Keyboard Shortcuts
**File List**:
- `website/glc/lib/hooks/use-keyboard-shortcuts.ts` - Keyboard shortcuts

**Work Content**:
- Implement Delete/Backspace to delete selected
- Implement Ctrl+C/V to copy/paste nodes
- Implement Ctrl+Z/Y for undo/redo
- Implement Ctrl+S to save graph
- Implement Ctrl+A to select all

**Acceptance Criteria**:
1. WHEN user presses Delete, selected nodes/edges SHALL be removed
2. WHEN user presses Ctrl+C and Ctrl+V, nodes SHALL be copied and pasted
3. WHEN user presses Ctrl+Z, last change SHALL be undone
4. WHEN user presses Ctrl+Y, undone change SHALL be redone
5. WHEN user presses Ctrl+S, graph SHALL be saved

#### 2.6.2 Context Menus
**File List**:
- `website/glc/components/glc/canvas/node-context-menu.tsx` - Node context menu
- `website/glc/components/glc/canvas/edge-context-menu.tsx` - Edge context menu
- `website/glc/components/glc/canvas/canvas-context-menu.tsx` - Canvas context menu

**Work Content**:
- Implement right-click context menu for nodes
- Implement right-click context menu for edges
- Implement right-click context menu for canvas
- Add actions: duplicate, delete, copy, paste
- Position menu correctly relative to click

**Acceptance Criteria**:
1. WHEN user right-clicks node, node context menu SHALL appear
2. WHEN user right-clicks edge, edge context menu SHALL appear
3. WHEN user right-clicks canvas, canvas context menu SHALL appear
4. WHEN user clicks menu item, action SHALL execute
5. WHEN user clicks outside menu, menu SHALL close

---

## Phase 2 Overall Acceptance Criteria

### Functional Acceptance
1. WHEN user opens `/glc/d3fend`, canvas SHALL load with D3FEND preset
2. WHEN user drags node from palette to canvas, node SHALL be created
3. WHEN user drags from node handle to another node, edge SHALL be created
4. WHEN user clicks node, node SHALL be selected with border highlight
5. WHEN user uses zoom controls, canvas SHALL zoom in/out smoothly

### Code Quality Acceptance
1. WHEN running `npm run lint`, code SHALL pass ESLint with zero errors
2. WHEN running TypeScript check, SHALL have no type errors
3. WHEN reviewing code, components SHALL follow React best practices
4. WHEN reviewing code, hooks SHALL follow React Hooks rules
5. WHEN reviewing code, state management SHALL be consistent

### Performance Acceptance
1. WHEN canvas has 100 nodes, SHALL render at 60fps
2. WHEN canvas has 50 edges, SHALL render at 60fps
3. WHEN user creates node, operation SHALL complete in <100ms
4. WHEN user creates edge, operation SHALL complete in <100ms
5. WHEN user zooms, animation SHALL be smooth (60fps)

### Usability Acceptance
1. WHEN user drags node, movement SHALL follow mouse precisely
2. WHEN user creates edge, connection line SHALL show clear path
3. WHEN user selects relationship, edge SHALL be created instantly
4. WHEN user deletes node, deletion SHALL be smooth
5. WHEN user zooms/pan, controls SHALL be responsive

---

## Phase 2 Deliverables Checklist

### Code Deliverables
- [ ] React Flow canvas integration with pan/zoom/mini-map
- [ ] Node palette with drag-and-drop
- [ ] Dynamic node components (preset-aware)
- [ ] Node details sheet
- [ ] Dynamic edge components (preset-aware)
- [ ] Relationship picker dialog
- [ ] Canvas state management
- [ ] Keyboard shortcuts
- [ ] Context menus

### Documentation Deliverables
- [ ] Phase 2 implementation plan
- [ ] Phase 2 acceptance criteria checklist

---

## Dependencies

- Phase 1 must be completed before starting Phase 2
- Task 2.1 must be completed before Task 2.2
- Task 2.2 must be completed before Task 2.3
- Task 2.3 and 2.4 can be developed in parallel
- Task 2.5 must be completed before Task 2.6

---

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| React Flow performance with many nodes | High | Implement virtualization, limit node count for MVP |
| Drag-and-drop cross-browser issues | Medium | Test in multiple browsers, use React Flow's built-in DnD |
| State management complexity | Medium | Use proven pattern (Zustand or Context), keep simple |
| Edge routing conflicts | Low | Use React Flow's built-in routing algorithms |

---

## Time Estimation

| Task | Estimated Hours |
|------|-----------------|
| 2.1 React Flow Canvas Integration | 8-10 |
| 2.2 Node Palette Implementation | 10-12 |
| 2.3 Canvas Node Implementation | 12-16 |
| 2.4 Edge (Relationship) Implementation | 10-14 |
| 2.5 Canvas State Management | 8-10 |
| 2.6 Canvas Interactions | 6-8 |
| **Total** | **54-70** |
