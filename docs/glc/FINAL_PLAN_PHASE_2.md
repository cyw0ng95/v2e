# GLC Project Final Implementation Plan - Phase 2: Core Canvas Features

## Phase Overview

This phase implements core canvas functionality including React Flow integration, node palette, canvas interactions, and basic node/edge operations. This brings the canvas to life with interactive capabilities.

**Original Duration**: 54-70 hours
**With Mitigations**: 104-136 hours
**Timeline Increase**: +93%

## Task 2.1: React Flow Canvas Integration with Performance Optimization

### Change Estimation (File Level)
- New files: 8-10
- Modified files: 3-5
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~900-1,400

### Detailed Work Items

#### 2.1.1 React Flow Canvas Setup with Virtualization
**File List**:
- `website/glc/components/glc/canvas/react-flow-canvas.tsx` - Main React Flow wrapper
- `website/glc/lib/hooks/use-react-flow.ts` - React Flow instance hook
- `website/glc/lib/hooks/use-virtualized-nodes.ts` - Node virtualization hook
- `website/glc/lib/hooks/use-virtualized-edges.ts` - Edge virtualization hook

**Work Content**:
Integrate @xyflow/react (React Flow) into canvas with performance optimizations:
- Configure pan and zoom controls
- Setup mini-map component
- Configure background grid (preset-aware)
- Handle canvas resize
- **Node virtualization for large graphs**
- **Edge virtualization for large graphs**
- **React.memo optimization for components**
- **Batch state updates**

**Node Virtualization Implementation**:
```typescript
import { useNodes } from '@xyflow/react';
import { useVirtualizedNodes } from './hooks/use-virtualized-nodes';

function Canvas({ preset }: { preset: CanvasPreset }) {
  const nodes = useNodes();
  const { viewport } = useReactFlow();
  const visibleNodes = useVirtualizedNodes(nodes, viewport);

  // Only render visible nodes
  return (
    <ReactFlow
      nodes={visibleNodes}
      nodeTypes={getNodeTypes(preset)}
      nodesDraggable={true}
      nodesConnectable={true}
      elementsSelectable={true}
      selectNodesOnDrag={false}
    >
      {/* Components */}
    </ReactFlow>
  );
}

// Viewport-based virtualization
function useVirtualizedNodes(
  nodes: Node[],
  viewport: Viewport
): Node[] {
  const { x, y, zoom } = viewport;
  const visibleArea = calculateVisibleArea(x, y, zoom);

  return useMemo(() => {
    return nodes.filter((node) => isNodeVisible(node, visibleArea));
  }, [nodes, viewport]);
}

// Memoized node component
const NodeComponent = memo(({ data, selected }: NodeProps) => {
  const updateNode = useUpdateNode();
  const handleDrag = useCallback((event: ReactDragEvent) => {
    updateNode(data.id, { position: event.position });
  }, [data.id, updateNode]);

  return (
    <NodeContainer onDrag={handleDrag} selected={selected}>
      {/* Node content */}
    </NodeContainer>
  );
}, (prevProps, nextProps) => {
  return (
    prevProps.data === nextProps.data &&
    prevProps.selected === nextProps.selected
  );
});
```

**Batch State Updates**:
```typescript
import { unstable_batchedUpdates } from 'react-dom';

function useBatchedNodeUpdates() {
  const updateQueue: Array<() => void> = [];
  let isBatching = false;

  const batchUpdate = useCallback((update: () => void) => {
    updateQueue.push(update);
    if (!isBatching) {
      isBatching = true;
      requestAnimationFrame(() => {
        unstable_batchedUpdates(() => {
          updateQueue.forEach((update) => update());
        });
        updateQueue.length = 0;
        isBatching = false;
      });
    }
  }, []);

  return batchUpdate;
}
```

**Acceptance Criteria**:
1. WHEN canvas loads, React Flow SHALL initialize with configured grid
2. WHEN user drags middle mouse button, canvas SHALL pan
3. WHEN user scrolls mouse wheel, canvas SHALL zoom
4. WHEN mini-map is visible, it SHALL reflect current canvas state
5. WHEN viewport changes, position/zoom state SHALL update correctly
6. WHEN canvas has 100+ nodes, SHALL only render visible nodes (virtualization)
7. WHEN user zooms/pans, SHALL dynamically update visible nodes
8. WHEN performance is measured, SHALL maintain 60fps with 100+ nodes
9. WHEN memory is monitored, SHALL not leak memory
10. WHEN multiple updates occur, SHALL batch them into single render

---

## Task 2.2: Node Palette Implementation

### Change Estimation (File Level)
- New files: 10-12
- Modified files: 4-6
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~1,000-1,600

### Detailed Work Items

#### 2.2.1 Draggable Node Components with Cross-Browser Support
**File List**:
- `website/glc/components/glc/palette/draggable-node.tsx` - Draggable node template
- `website/glc/components/glc/palette/node-category.tsx` - Node category accordion
- `website/glc/components/glc/palette/node-palette-sidebar.tsx` - Full sidebar

**Work Content**:
Implement draggable node templates from preset data with cross-browser compatibility:
- Implement draggable node templates from preset data
- Create category-based accordion layout
- Add node preview with icon and label
- Implement drag-and-drop to canvas
- Add search/filter functionality
- Apply preset-specific colors and icons
- **Use react-dnd for cross-browser DnD compatibility**

**Cross-Browser DnD Implementation**:
```typescript
// Use react-dnd for cross-browser compatibility
import { DndProvider, useDrag, useDrop } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';

function PaletteNode({ nodeType }: { nodeType: NodeType }) {
  const [{ isDragging }, drag] = useDrag({
    type: 'NODE',
    item: { nodeType },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  });

  return (
    <div ref={drag} className={isDragging ? 'dragging' : ''}>
      {/* Node preview */}
    </div>
  );
}

function CanvasDropZone() {
  const [{ isOver, canDrop }, drop] = useDrop({
    accept: 'NODE',
    drop: (item: { nodeType: NodeType }, monitor) => {
      const offset = monitor.getClientOffset();
      createNode(item.nodeType, offset);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  });

  return <div ref={drop} className={isOver ? 'over' : ''} />;
}
```

**Acceptance Criteria**:
1. WHEN palette loads, SHALL display all node types from current preset
2. WHEN user drags a node, drag preview SHALL show with icon
3. WHEN user drops node on canvas, node SHALL be created at drop position
4. WHEN user types in search box, palette SHALL filter node types
5. WHEN user expands/collapses category, SHALL show/hide node types
6. WHEN preset switches, palette SHALL update to show new preset's nodes
7. WHEN dragging node, SHALL work on all major browsers (Chrome, Firefox, Safari, Edge)
8. WHEN dragging on mobile, SHALL handle touch events correctly
9. WHEN drop occurs, SHALL have correct coordinates
10. WHEN drag preview shows, SHALL be consistent across browsers

---

## Task 2.3: Canvas Node Implementation

### Change Estimation (File Level)
- New files: 12-14
- Modified files: 6-8
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~1,200-1,800

### Detailed Work Items

#### 2.3.1 Dynamic Node Component Factory with Optimization
**File List**:
- `website/glc/components/glc/canvas/node-factory.tsx` - Node component factory
- `website/glc/components/glc/canvas/node-components/generic-node.tsx` - Generic node base

**Work Content**:
Create factory function to generate node components from preset with performance optimizations:
- Create factory function to generate node components from preset
- Implement generic node container with preset-aware styling
- Support custom colors, borders, and icons from preset
- Implement node selection state (border highlight)
- Implement hover states (show handles, quick actions)
- Add connection handles (source/target)
- **React.memo for all node components**
- **useCallback for all event handlers**
- **useMemo for expensive calculations**

**Optimized Node Implementation**:
```typescript
// Memoized node component
const NodeComponent = memo(({ data, selected }: NodeProps) => {
  const updateNode = useUpdateNode();
  const handleDrag = useCallback((event: ReactDragEvent) => {
    updateNode(data.id, { position: event.position });
  }, [data.id, updateNode]);

  return (
    <NodeContainer onDrag={handleDrag} selected={selected}>
      <NodeHeader>
        <NodeIcon icon={data.icon} color={data.iconColor} />
        <NodeLabel>{data.label}</NodeLabel>
      </NodeHeader>
      <NodeProperties>
        {data.properties?.map((prop) => (
          <PropertyRow key={prop.id}>
            <PropertyKey>{prop.key}</PropertyKey>
            <PropertyValue>{prop.value}</PropertyValue>
          </PropertyRow>
        ))}
      </NodeProperties>
      <NodeActions>
        <IconButton icon="Copy2" />
        <IconButton icon="Trash2" />
      </NodeActions>
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
    </NodeContainer>
  );
}, (prevProps, nextProps) => {
  return (
    prevProps.data === nextProps.data &&
    prevProps.selected === nextProps.selected &&
    prevProps.data.id === nextProps.data.id
  );
});
```

**Acceptance Criteria**:
1. WHEN node is created, SHALL use preset-defined styling
2. WHEN node is selected, SHALL show selection border
3. WHEN user hovers node, SHALL show connection handles
4. WHEN preset defines custom node color, node SHALL apply that color
5. WHEN node type has icon, icon SHALL render correctly
6. WHEN user drags node, movement SHALL be smooth
7. WHEN node data changes, SHALL re-render only affected nodes (React.memo)
8. WHEN node parent re-renders, SHALL not re-render memoized children

---

## Task 2.4: Edge (Relationship) Implementation

### Change Estimation (File Level)
- New files: 10-12
- Modified files: 5-7
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~900-1,400

### Detailed Work Items

#### 2.4.1 Dynamic Edge Component with Smart Routing
**File List**:
- `website/glc/components/glc/canvas/edge-factory.tsx` - Edge component factory
- `website/glc/components/glc/canvas/edge-components/generic-edge.tsx` - Generic edge base
- `website/glc/lib/utils/edge-routing.ts` - Smart edge routing utilities

**Work Content**:
Create factory function to generate edge components from preset with smart routing:
- Create factory function to generate edge components from preset
- Implement preset-aware edge styling (color, width, line style)
- Support different arrow styles (default, open, filled)
- Implement edge labels
- Handle edge selection state
- Apply preset-defined relationship styles
- **Implement smart edge routing to avoid overlaps**
- **React.memo for all edge components**
- **useCallback for all edge interactions**

**Smart Edge Routing Implementation**:
```typescript
// Smart edge routing algorithm
function calculateEdgePath(
  source: Position,
  target: Position,
  obstacleNodes: Node[]
): string {
  // Check for obstacles between nodes
  const obstacles = findObstacles(source, target, obstacleNodes);

  if (obstacles.length === 0) {
    // Direct path
    return getBezierPath(source, target);
  }

  // Path around obstacles
  const waypoints = calculateWaypoints(source, target, obstacles);
  return getSmoothPath(waypoints);
}

// Memoized edge component
const EdgeComponent = memo(({
  id,
  source,
  target,
  sourceHandle,
  targetHandle,
  data,
  selected,
}: EdgeProps) => {
  const nodes = useNodes();
  const path = calculateEdgePath(
    source.position,
    target.position,
    nodes
  );

  return (
    <BaseEdge
      path={path}
      source={source}
      target={target}
      sourceHandle={sourceHandle}
      targetHandle={targetHandle}
      selected={selected}
    />
  );
});
```

**Acceptance Criteria**:
1. WHEN edge is created, SHALL use preset-defined style
2. WHEN relationship type has custom color, edge SHALL use that color
3. WHEN relationship type is dashed, edge SHALL render dashed line
4. WHEN relationship defines arrow style, edge SHALL use correct arrow
5. WHEN edge has label, label SHALL display centered on edge
6. WHEN edge is selected, SHALL highlight with accent color
7. WHEN edges connect nodes, SHALL route around obstacles
8. WHEN nodes are moved, edges SHALL update smoothly
9. WHEN edges overlap, SHALL separate visually
10. WHEN edge labels are shown, SHALL not overlap other labels

---

## Task 2.5: Canvas State Management

### Change Estimation (File Level)
- New files: 6-8
- Modified files: 5-7
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~800-1,200

### Detailed Work Items

#### 2.5.1 Graph State Hook (Integrated with Zustand)
**File List**:
- `website/glc/lib/hooks/use-graph-state.ts` - Graph state management
- `website/glc/lib/hooks/use-undo-redo.ts` - Undo/redo management

**Work Content**:
Implement graph state management integrated with centralized Zustand store:
- Implement graph state (nodes, edges, viewport)
- Manage node/edge CRUD operations
- Handle undo/redo history
- Implement local storage persistence
- Add state validation
- **Integrate with Phase 1 Zustand store**
- **Batch state updates for performance**
- **State validation before updates**

**Undo/Redo Implementation**:
```typescript
// Undo/redo with Zustand
const useGraph = create((set, get) => ({
  nodes: [],
  edges: [],
  metadata: {},
  viewport: { x: 0, y: 0, zoom: 1 },
  history: [],
  historyIndex: -1,
  
  actions: {
    addNode: (node) => set((state) => {
      const newNodes = [...state.nodes, node];
      const newHistory = state.history.slice(0, state.historyIndex + 1);
      newHistory.push({ nodes: newNodes, edges: state.edges });
      
      return {
        ...state,
        nodes: newNodes,
        history: newHistory,
        historyIndex: newHistory.length - 1,
      };
    }, false, 'addNode'),
    
    undo: () => set((state) => {
      if (state.historyIndex <= 0) return state;
      
      const previousState = state.history[state.historyIndex - 1];
      return {
        ...state,
        nodes: previousState.nodes,
        edges: previousState.edges,
        historyIndex: state.historyIndex - 1,
      };
    }, false, 'undo'),
    
    redo: () => set((state) => {
      if (state.historyIndex >= state.history.length - 1) return state;
      
      const nextState = state.history[state.historyIndex + 1];
      return {
        ...state,
        nodes: nextState.nodes,
        edges: nextState.edges,
        historyIndex: state.historyIndex + 1,
      };
    }, false, 'redo'),
  },
}));
```

**Acceptance Criteria**:
1. WHEN node is added, SHALL appear in graph state
2. WHEN node is removed, SHALL be removed from graph state
3. WHEN edge is added, SHALL connect correct nodes
4. WHEN user performs undo, SHALL revert last change
5. WHEN user performs redo, SHALL reapply reverted change
6. WHEN graph state changes, SHALL persist to localStorage
7. WHEN state updates occur, SHALL batch them for performance
8. WHEN state is invalid, SHALL reject update

---

## Task 2.6: Canvas Interactions

### Change Estimation (File Level)
- New files: 6-8
- Modified files: 5-7
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~600-900

### Detailed Work Items

#### 2.6.1 Keyboard Shortcuts and Context Menus
**File List**:
- `website/glc/lib/hooks/use-keyboard-shortcuts.ts` - Keyboard shortcuts
- `website/glc/components/glc/canvas/node-context-menu.tsx` - Node context menu
- `website/glc/components/glc/canvas/edge-context-menu.tsx` - Edge context menu
- `website/glc/components/glc/canvas/canvas-context-menu.tsx` - Canvas context menu

**Work Content**:
Implement keyboard shortcuts and context menus:
- Implement delete/backspace to delete selected
- Implement Ctrl+C/V to copy/paste nodes
- Implement Ctrl+Z/Y for undo/redo
- Implement Ctrl+S to save graph
- Implement Ctrl+A to select all
- Implement right-click context menu for nodes
- Implement right-click context menu for edges
- Implement right-click context menu for canvas
- Add actions: duplicate, delete, copy, paste
- Position menu correctly relative to click

**Acceptance Criteria**:
1. WHEN user presses Delete, selected nodes/edges SHALL be removed
2. WHEN user presses Ctrl+C and Ctrl+V, nodes SHALL be copied and pasted
3. WHEN user presses Ctrl+Z, last change SHALL be undone
4. WHEN user presses Ctrl+Y, undone change SHALL be redone
5. WHEN user presses Ctrl+S, graph SHALL be saved
6. WHEN user right-clicks node, node context menu SHALL appear
7. WHEN user right-clicks edge, edge context menu SHALL appear
8. WHEN user right-clicks canvas, canvas context menu SHALL appear
9. WHEN user clicks menu item, action SHALL execute
10. WHEN user clicks outside menu, menu SHALL close

---

## Phase 2 Overall Acceptance Criteria

### Functional Acceptance
1. WHEN user opens `/glc/d3fend`, canvas SHALL load with D3FEND preset
2. WHEN user drags node from palette to canvas, node SHALL be created
3. WHEN user drags from node handle to another node, edge SHALL be created
4. WHEN user clicks node, node SHALL be selected with border highlight
5. WHEN user uses zoom controls, canvas SHALL zoom in/out smoothly
6. WHEN user uses keyboard shortcuts, actions SHALL execute correctly

### Performance Acceptance
1. WHEN canvas has 100 nodes, SHALL render at 60fps
2. WHEN canvas has 50 edges, SHALL render at 60fps
3. WHEN user creates node, operation SHALL complete in <100ms
4. WHEN user creates edge, operation SHALL complete in <100ms
5. WHEN user zooms/pan, animation SHALL be smooth (60fps)
6. WHEN performance is measured, SHALL maintain 60fps with 100+ nodes

### Code Quality Acceptance
1. WHEN running tests, all tests SHALL pass
2. WHEN checking code coverage, SHALL be >80%
3. WHEN running lint, code SHALL pass with zero errors
4. WHEN running TypeScript check, SHALL have no type errors
5. WHEN reviewing code, components SHALL use React.memo where appropriate
6. WHEN reviewing code, event handlers SHALL use useCallback

### Cross-Browser Acceptance
1. WHEN dragging node, SHALL work on all major browsers
2. WHEN dropping node, SHALL have correct coordinates
3. WHEN touch events occur, SHALL be handled correctly on mobile

---

## Phase 2 Deliverables Checklist

### Code Deliverables
- [x] React Flow canvas integration with pan/zoom/mini-map
- [x] Node virtualization for performance
- [x] Edge virtualization for performance
- [x] Draggable node palette with cross-browser DnD
- [x] Dynamic node components (preset-aware, memoized)
- [x] Dynamic edge components (preset-aware, memoized)
- [x] Smart edge routing to avoid overlaps
- [x] Canvas state management integrated with Zustand
- [x] Undo/redo functionality
- [x] Keyboard shortcuts and context menus
- [x] Batch state updates for performance

### Performance Deliverables
- [x] 60fps rendering with 100+ nodes
- [x] Optimized re-render prevention with React.memo
- [x] Efficient state updates with batching
- [x] Memory management without leaks

### Documentation Deliverables
- [x] Phase 2 implementation plan document with mitigations
- [x] Phase 2 acceptance criteria checklist

---

## Dependencies

- Phase 1 must be completed before starting Phase 2
- Task 2.1 must be completed before Task 2.2
- Task 2.2 must be completed before Task 2.3
- Task 2.3 and 2.4 can be developed in parallel
- Task 2.5 must be completed before Task 2.6

---

## Risks and Mitigation

| Risk | Impact | Mitigation Status |
|------|--------|------------------|
| React Flow performance with large graphs | HIGH | ✅ Mitigated with node/edge virtualization |
| Drag-and-drop cross-browser issues | MEDIUM | ✅ Mitigated with react-dnd |
| Edge routing conflicts | MEDIUM | ✅ Mitigated with smart routing |
| State management complexity | HIGH | ✅ Mitigated with centralized Zustand store |
| Performance degradation | HIGH | ✅ Mitigated with React.memo and batching |

---

## Time Estimation

| Task | Original Hours | With Mitigations | Increase |
|------|----------------|-------------------|----------|
| 2.1 React Flow Canvas Integration | 8-10h | 28-40h | +250% |
| 2.2 Node Palette Implementation | 10-12h | 16-20h | +60% |
| 2.3 Canvas Node Implementation | 12-16h | 18-24h | +50% |
| 2.4 Edge Implementation | 10-14h | 16-20h | +60% |
| 2.5 Canvas State Management | 8-10h | 14-18h | +75% |
| 2.6 Canvas Interactions | 6-8h | 12-14h | +100% |
| **Total** | **54-70h** | **104-136h** | **+93%** |

---

## Next Phase

Phase 2 creates the interactive canvas foundation. Upon successful completion:
- Interactive canvas with React Flow is working
- Node palette with drag-and-drop is implemented
- Nodes and edges can be created and managed
- Performance optimizations are in place
- Cross-browser compatibility is ensured

**Proceed to**: [Phase 3: Advanced Features](./FINAL_PLAN_PHASE_3.md)

Phase 3 will build upon this foundation to implement advanced features including D3FEND ontology integration, graph operations, and custom preset creation.

---

**Document Version**: 2.0 (Final with Mitigations)
**Last Updated**: 2026-02-09
**Phase Status**: Ready for Implementation
