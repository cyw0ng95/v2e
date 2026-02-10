# GLC (Graphized Learning Canvas) Implementation Plan

## Overview

**GLC** is a graph-based modeling platform for the v2e vulnerability management system, supporting multiple customizable canvas presets.

**Initial Presets:**
1. **D3FEND Canvas** - Cyber attack/defense modeling using MITRE D3FEND ontology
2. **Topo-Graph Canvas** - General-purpose graph and topology diagramming

## Technology Stack

- **Frontend**: Next.js 15+, React 19, TypeScript (strict mode), Tailwind CSS v4, shadcn/ui
- **Canvas**: @xyflow/react (React Flow)
- **State**: Zustand with slices (graph, preset, canvas, ui, undo-redo)
- **Validation**: Zod schemas
- **Backend**: Existing `v2local` service with SQLite (Phase 6) - no new subprocess

## Project Timeline

| Phase | Focus | Duration | Status |
|-------|-------|----------|--------|
| 1 | Core Infrastructure | 8 weeks | ✅ COMPLETE |
| 2 | Core Canvas Features | 8 weeks | ✅ COMPLETE |
| 3 | Advanced Features | 8 weeks | ⏳ PARTIAL (50%) |
| 4 | UI Polish & Testing | 12 weeks | ⏳ PARTIAL (60%) |
| 5 | Backend Integration | 12 weeks | ⏳ PARTIAL (40%) |
| **Total** | | **48 weeks** | **~60% Complete** |

---

## Phase 1: Core Infrastructure ✅ COMPLETE

**Location**: `website/glc/`
**Duration**: 8 weeks

### Jobs

#### J1.1 Project Initialization ✅ COMPLETE
**Files**: `package.json`, `next.config.ts`, `tsconfig.json`, `tailwind.config.ts`
- ✅ Next.js 15+ with static export (`output: 'export'`)
- ✅ TypeScript strict mode
- ✅ Tailwind CSS v4
- ✅ Dependencies: @xyflow/react, shadcn/ui, zustand, zod, lucide-react, sonner

#### J1.2 shadcn/ui Setup ✅ COMPLETE
**Files**: `components.json`, `components/ui/*.tsx`
- ✅ shadcn/ui initialized
- ✅ Components: button, dialog, dropdown-menu, input, label, sheet, tabs, accordion, toast

#### J1.3 Layout Structure ✅ COMPLETE
**Files**: `app/glc/layout.tsx`, `app/glc/page.tsx`, `app/glc/[presetId]/page.tsx`
- ✅ Root layout with ReactFlowProvider
- ✅ Landing page with hero section and preset cards
- ✅ Dynamic canvas page routes

#### J1.4 State Management ✅ COMPLETE
**Files**: `lib/glc/store/index.ts`, `lib/glc/store/slices/*.ts`
- ✅ Zustand store with 5 slices:
  - `preset` - currentPreset, builtInPresets, userPresets
  - `graph` - nodes, edges, metadata, viewport
  - `canvas` - selection, zoom, pan
  - `ui` - theme, sidebar, modals
  - `undo-redo` - history stack
- ✅ Persistence middleware (localStorage)
- ✅ Devtools middleware

#### J1.5 Type System ✅ COMPLETE
**Files**: `lib/glc/types/index.ts`
- ✅ CanvasPreset, NodeTypeDefinition, RelationshipDefinition
- ✅ CADNode, CADEdge, GraphMetadata, Graph
- ✅ Property, Reference interfaces
- ✅ Store slice types

#### J1.6 Built-in Presets ✅ COMPLETE
**Files**: `lib/glc/presets/d3fend-preset.ts`, `lib/glc/presets/topo-preset.ts`
- ✅ D3FEND preset: 9 node types, relationships, dark theme
- ✅ Topo-Graph preset: 8 node types, 8 relationships, light theme

#### J1.7 Validation System ✅ COMPLETE
**Files**: `lib/glc/validation/index.ts`
- ✅ Zod schemas for all preset types
- ✅ Preset migration system (version compatibility)
- ✅ `validatePreset()`, `validateGraph()`

#### J1.8 Error Handling ✅ COMPLETE
**Files**: `lib/glc/errors/index.tsx`
- ✅ Error boundaries: GraphErrorBoundary, PresetErrorBoundary
- ✅ Error recovery mechanisms with reset functionality

---

## Phase 2: Core Canvas Features ✅ COMPLETE

**Duration**: 8 weeks

### Jobs

#### J2.1 React Flow Canvas ✅ COMPLETE
**Files**: `components/glc/canvas/*`, `app/glc/[presetId]/page.tsx`
- ✅ React Flow configured with preset settings
- ✅ Background grid, controls, mini-map
- ✅ Node/edge change handlers

#### J2.2 Dynamic Node/Edge Components ✅ COMPLETE
**Files**: `components/glc/canvas/dynamic-node.tsx`, `components/glc/canvas/dynamic-edge.tsx`
- ✅ Preset-aware node rendering (color, icon, properties)
- ✅ Preset-aware edge rendering (style, label, arrow)
- ✅ React.memo optimization

#### J2.3 Node Palette ✅ COMPLETE
**Files**: `components/glc/palette/node-palette.tsx`
- ✅ Group node types by category (accordion)
- ✅ Search/filter functionality
- ✅ Drag handle for each node type

#### J2.4 Drag-and-Drop ✅ COMPLETE
**Files**: `app/glc/[presetId]/page.tsx`
- ✅ Cross-browser drag support (native HTML5 drag)
- ✅ Position calculation on drop
- ✅ Create node with preset defaults

#### J2.5 Node/Edge Editing ✅ COMPLETE
**Files**: `components/glc/canvas/node-details-sheet.tsx`, `components/glc/canvas/edge-details-sheet.tsx`
- ✅ Node: label, properties, colors
- ✅ Edge: relationship type, label
- ✅ Sheet component for editing

#### J2.6 Performance Optimization ✅ COMPLETE
**Files**: `lib/glc/performance/monitoring.ts`
- ✅ FPS monitoring
- ✅ Debounce/throttle utilities
- ✅ Batch updates via requestAnimationFrame
- ✅ Intersection observer for lazy loading

#### J2.7 Keyboard Shortcuts ✅ COMPLETE
**Files**: `lib/glc/shortcuts/use-shortcuts.ts`, `lib/glc/shortcuts/shortcuts-dialog.tsx`
- ✅ Delete, Ctrl+Z (undo), Ctrl+Shift+Z (redo)
- ✅ Ctrl+C/V (copy/paste), Escape (clear)
- ✅ F (fit view), +/- (zoom), ? (help)

#### J2.8 Context Menus ✅ COMPLETE
**Files**: `components/glc/context-menu/canvas-context-menu.tsx`
- ✅ Canvas menu: fit view, undo, redo
- ⏳ TODO: Node menu: duplicate, edit, delete, color, D3FEND inferences
- ⏳ TODO: Edge menu: edit, delete, reverse

#### J2.9 Undo/Redo System ✅ COMPLETE
**Files**: `lib/glc/store/undo-redo-slice.ts`, `components/glc/toolbar/canvas-toolbar.tsx`
- ✅ Save state snapshots on changes
- ✅ Limit history size (100 max)
- ✅ Ctrl+Z / Ctrl+Shift+Z shortcuts

#### J2.10 Canvas Toolbar ✅ COMPLETE
**Files**: `components/glc/toolbar/canvas-toolbar.tsx`
- ✅ Undo/Redo, Zoom In/Out, Fit View
- ✅ Toggle Grid, Toggle Mini-map
- ✅ Save, Share, Export, Help buttons
- ✅ Preset indicator badge

---

## Phase 3: Advanced Features ⏳ PARTIAL (50%)

**Duration**: 8 weeks

### Jobs

#### J3.1 D3FEND Ontology Integration ✅ PARTIAL
**Files**: `lib/glc/d3fend/ontology.ts`
- ✅ Simplified D3FEND class hierarchy
- ✅ Search, ancestors, children helpers
- ⏳ TODO: Lazy loading full D3FEND data from assets
- ⏳ TODO: Virtualized tree for class browser component
- ⏳ TODO: Class picker UI component
- ❌ TODO: Inference engine (sensors, defensive techniques, weakness)

#### J3.2 Graph Save/Load ✅ COMPLETE
**Files**: `lib/glc/io/graph-io.ts`
- ✅ JSON format with version field
- ✅ Save to localStorage and file
- ✅ Load from localStorage and file
- ✅ Auto-save with debouncing
- ✅ Crash recovery from localStorage

#### J3.3 Graph Export ✅ COMPLETE
**Files**: `lib/glc/io/exporters.ts`, `components/glc/export/export-dialog.tsx`
- ✅ PNG export (html-to-image)
- ✅ SVG export (html-to-image)
- ⏳ TODO: PDF export (jsPDF) - not implemented
- ✅ Export dialog with options

#### J3.4 STIX 2.1 Import ❌ NOT STARTED
**Files**: `lib/glc/stix/*.ts`, `components/glc/stix/stix-import-dialog.tsx`
- ❌ TODO: Parse STIX JSON
- ❌ TODO: Validate objects and relationships
- ❌ TODO: Map to GLC graph structure
- ❌ TODO: Map to D3FEND ontology

#### J3.5 Custom Preset Editor ❌ NOT STARTED
**Files**: `components/glc/preset-editor/*.tsx`
- ❌ TODO: 5-step wizard: Basic Info → Node Types → Relationships → Styling → Behavior
- ❌ TODO: Live preview
- ❌ TODO: Auto-save drafts
- ❌ TODO: Icon/color pickers

#### J3.6 Example Graphs ❌ NOT STARTED
**Files**: `assets/examples/example-graphs.json`, `components/glc/examples/*.tsx`
- ❌ TODO: D3FEND examples: attack chains, defense strategies
- ❌ TODO: Topo-Graph examples: network topology, process flows
- ❌ TODO: Example gallery with thumbnails

#### J3.7 Smart Edge Routing ❌ NOT STARTED
**Files**: `lib/glc/routing/*.ts`
- ❌ TODO: Obstacle detection
- ❌ TODO: A* pathfinding
- ❌ TODO: Bezier curve generation
- ❌ TODO: User toggle (on/off)

#### J3.8 Share & Embed ✅ COMPLETE
**Files**: `lib/glc/share/share.ts`, `components/glc/share/share-dialog.tsx`
- ✅ URL encoding with compression (lz-string)
- ✅ Share dialog with copy link
- ✅ Embed code generator (iframe)

---

## Phase 4: UI Polish & Production ⏳ PARTIAL (60%)

**Duration**: 12 weeks

### Jobs

#### J4.1 Visual Design System ✅ COMPLETE
**Files**: `lib/glc/theme/design-tokens.ts`
- ✅ Design tokens: colors (light/dark/high-contrast)
- ✅ Color manipulation utilities (lighten, darken, withAlpha)
- ✅ Contrast validation functions

#### J4.2 Dark/Light Mode ✅ COMPLETE
**Files**: `lib/glc/theme/design-tokens.ts`
- ✅ Theme definitions: light, dark, high-contrast
- ✅ System preference detection (`getSystemTheme`)
- ✅ Contrast validation (WCAG AA: 4.5:1)
- ⏳ TODO: Theme provider with persistence (using preset themes instead)
- ⏳ TODO: Theme toggle UI (Ctrl+Shift+T shortcut)

#### J4.3 Responsive Design ❌ NOT STARTED
**Files**: `lib/glc/responsive/*.ts`, `components/glc/responsive/*.tsx`
- ❌ TODO: Breakpoints: xs(0-639), sm(640-767), md(768-1023), lg(1024-1279), xl(1280+)
- ❌ TODO: Mobile: drawer palette, collapsed mini-map, touch targets 44px
- ❌ TODO: Tablet/Desktop: optimized layouts

#### J4.4 Animation Optimization ✅ COMPLETE
**Files**: `lib/glc/a11y/focus-management.ts`
- ✅ Reduced motion support (`prefersReducedMotion`)
- ✅ Animation duration helper (`getAnimationDuration`)

#### J4.5 Keyboard Navigation ✅ PARTIAL
**Files**: `lib/glc/a11y/focus-management.ts`
- ✅ Focus trap for modals (`createFocusTrap`)
- ✅ Get focusable elements helper
- ⏳ TODO: Tab through all interactive elements (needs testing)
- ⏳ TODO: Skip links

#### J4.6 Screen Reader Support ✅ PARTIAL
**Files**: `lib/glc/a11y/focus-management.ts`
- ✅ Live regions for announcements (`announce` function)
- ✅ ARIA ID generation (`generateAriaId`)
- ⏳ TODO: Full ARIA attributes on all components
- ❌ TODO: Test with NVDA, VoiceOver, TalkBack

#### J4.7 Performance Optimization ⏳ PARTIAL
**Files**: `lib/glc/performance/monitoring.ts`
- ✅ FPS monitoring
- ✅ Debounce/throttle utilities
- ✅ Intersection observer for lazy loading
- ⏳ TODO: Code splitting by route/feature
- ⏳ TODO: Bundle target: <500KB
- ⏳ TODO: FCP <2s (landing), LCP <2.5s

#### J4.8 Testing ❌ NOT STARTED
**Files**: `__tests__/**/*.test.{ts,tsx}`, `tests/glc/*.spec.ts`
- ❌ TODO: Unit tests: store, types, validation, utilities
- ❌ TODO: Component tests: all UI components
- ❌ TODO: Integration tests: user workflows
- ❌ TODO: E2E tests: critical journeys (Playwright)
- ❌ TODO: Target: >80% coverage

---

## Phase 5: Backend Integration

**Duration**: 12 weeks
**Service**: `cmd/v2local` (existing - no new subprocess)
**Status**: In Progress (J5.1-J5.3 Complete)

### Overview
GLC backend integration uses the existing `v2local` service, which already handles SQLite storage for CVE/CWE/CAPEC/ATT&CK data and memory cards. No new `v2glc` subprocess is needed.

### Jobs

#### J5.1 Backend Handlers in v2local ✅ COMPLETE
**Files**: `cmd/v2local/glc_handlers.go`, `cmd/v2local/service.md`, `pkg/glc/`
- ✅ Created `pkg/glc/` package with models, store, and migrations
- ✅ Added RPC handlers to v2local:
  - `RPCGLCGraphCreate`, `RPCGLCGraphGet`, `RPCGLCGraphUpdate`, `RPCGLCGraphDelete`, `RPCGLCGraphList`, `RPCGLCGraphListRecent`
  - `RPCGLCVersionGet`, `RPCGLCVersionList`, `RPCGLCVersionRestore`
  - `RPCGLCPresetCreate`, `RPCGLCPresetGet`, `RPCGLCPresetUpdate`, `RPCGLCPresetDelete`, `RPCGLCPresetList`
  - `RPCGLCShareCreateLink`, `RPCGLCShareGetShared`, `RPCGLCShareGetEmbedData`
- ✅ Updated `cmd/v2local/service.md` with API specification (methods 101-117)

#### J5.2 Database Schema in v2local ✅ COMPLETE
**Files**: `pkg/glc/models.go`, `pkg/glc/migration.go`
- ✅ SQLite tables via GORM AutoMigrate:
  - `glc_graphs` - graph metadata and content
  - `glc_graph_versions` - version history for undo/restore
  - `glc_user_presets` - user-defined presets
  - `glc_share_links` - share/embed links
- ✅ Reuses v2local's existing SQLite connection (bookmark.db)

#### J5.3 Frontend RPC Client ✅ COMPLETE
**Files**: `website/lib/types.ts`, `website/lib/rpc-client.ts`
- ✅ Added GLC types: GLCGraph, GLCGraphVersion, GLCUserPreset, GLCShareLink
- ✅ Added request/response types for all operations
- ✅ Added RPC methods:
  - Graph: createGLCGraph, getGLCGraph, updateGLCGraph, deleteGLCGraph, listGLCGraphs, listRecentGLCGraphs
  - Versions: getGLCVersion, listGLCVersions, restoreGLCVersion
  - Presets: createGLCPreset, getGLCPreset, updateGLCPreset, deleteGLCPreset, listGLCPresets
  - Share: createGLCShareLink, getGLCSharedGraph, getGLCShareEmbedData
- ⏳ TODO: Retry with exponential backoff
- ⏳ TODO: Offline queue (localStorage)
- ⏳ TODO: Network status monitoring

#### J5.4 Optimistic UI ⏳ NOT STARTED
**Files**: `website/lib/glc/optimistic/*.ts`, `website/components/glc/optimistic/*.tsx`
- ⏳ TODO: Immediate UI updates before server confirmation
- ⏳ TODO: Version conflict detection (compare client/server versions)
- ⏳ TODO: Conflict resolution dialog (merge or overwrite options)
- ⏳ TODO: Rollback on failure with toast notification

#### J5.5 Graph Browser UI ⏳ NOT STARTED
**Files**: `website/app/glc/my-graphs/page.tsx`, `website/components/glc/graph-browser/*.tsx`
- ⏳ TODO: Grid/list view toggle with thumbnails (generate thumbnails on save)
- ⏳ TODO: Search by name and filter by preset, date range, tags
- ⏳ TODO: Pagination with configurable page size
- ⏳ TODO: Actions: open, duplicate, delete, share, export
- ⏳ TODO: Enable "My Graphs" button on landing page (currently disabled)

#### J5.6 Versioning & Recovery ⏳ NOT STARTED
**Files**: `website/lib/glc/versioning/*.ts`, `website/components/glc/versioning/*.tsx`
- ⏳ TODO: Auto-save with debouncing (500ms default), on idle, before unload
- ⏳ TODO: Version history panel with diff visualization (node/edge changes)
- ⏳ TODO: Restore previous version with confirmation
- ⏳ TODO: Crash recovery from localStorage (recover unsaved changes)
- ⏳ TODO: Version limit enforcement (prune old versions)

#### J5.7 Offline Support ⏳ NOT STARTED
**Files**: `website/lib/glc/offline/*.ts`, `website/components/glc/offline/*.tsx`
- ⏳ TODO: Offline detection (navigator.onLine + online/offline events)
- ⏳ TODO: Operation queueing in IndexedDB/localStorage
- ⏳ TODO: Sync on reconnect with conflict resolution
- ⏳ TODO: Offline indicator in UI (banner + status icon)

#### J5.8 Testing ⏳ NOT STARTED
**Files**: `tests/glc/*.spec.ts` (Playwright), `pkg/glc/*_test.go`
- ⏳ TODO: Backend unit tests for GLC handlers in v2local
- ⏳ TODO: RPC integration tests via `/restful/rpc` endpoint
- ⏳ TODO: Frontend component tests for GLC components
- ⏳ TODO: E2E tests for full lifecycle (create → edit → save → share)
- ⏳ TODO: Load tests (100 concurrent users) - optional, depends on scale requirements

---

## Quality Standards

### Performance Targets
| Metric | Target |
|--------|--------|
| Bundle size | <500KB |
| FCP (landing) | <2s |
| LCP | <2.5s |
| Rendering | 60fps with 100+ nodes |
| Operations | <100ms |

### Accessibility
- WCAG AA compliance (4.5:1 contrast ratio)
- Full keyboard navigation
- Screen reader support (ARIA)
- Reduced motion support

### Testing
- >80% code coverage
- Zero TypeScript errors
- Zero ESLint errors
- All tests pass

---

## Critical Risks & Mitigations

| Risk | Level | Mitigation |
|------|-------|------------|
| State management complexity | CRITICAL | Zustand with slices, persistence, devtools |
| React Flow performance | CRITICAL | Virtualization, React.memo, batched updates |
| D3FEND data overload | CRITICAL | Lazy loading, virtualized tree |
| Concurrent edit conflicts | HIGH | Optimistic UI, version checking, resolution dialog |
| Accessibility gaps | HIGH | WCAG audit, keyboard nav, screen reader testing |
| Dark mode contrast | MEDIUM | Automated validation, high contrast mode |

---

## File Structure

```
website/
├── app/glc/
│   ├── layout.tsx                      # ✅ ReactFlowProvider wrapper
│   ├── page.tsx                        # ✅ Landing page
│   └── [presetId]/
│       └── page.tsx                    # ✅ Canvas page
├── components/glc/
│   ├── canvas/
│   │   ├── dynamic-node.tsx            # ✅ Preset-aware node renderer
│   │   ├── dynamic-edge.tsx            # ✅ Preset-aware edge renderer
│   │   ├── node-details-sheet.tsx      # ✅ Node editing sheet
│   │   └── edge-details-sheet.tsx      # ✅ Edge editing sheet
│   ├── palette/
│   │   └── node-palette.tsx            # ✅ Node type palette with search
│   ├── toolbar/
│   │   └── canvas-toolbar.tsx          # ✅ Canvas toolbar
│   ├── context-menu/
│   │   └── canvas-context-menu.tsx     # ✅ Canvas context menu
│   ├── export/
│   │   └── export-dialog.tsx           # ✅ Export dialog
│   └── share/
│       └── share-dialog.tsx            # ✅ Share dialog
├── lib/glc/
│   ├── store/
│   │   ├── index.ts                    # ✅ Zustand store
│   │   ├── preset-slice.ts             # ✅ Preset state
│   │   ├── graph-slice.ts              # ✅ Graph state
│   │   ├── canvas-slice.ts             # ✅ Canvas state
│   │   ├── ui-slice.ts                 # ✅ UI state
│   │   └── undo-redo-slice.ts          # ✅ Undo/redo state
│   ├── types/
│   │   └── index.ts                    # ✅ TypeScript types
│   ├── presets/
│   │   ├── index.ts                    # ✅ Preset exports
│   │   ├── d3fend-preset.ts            # ✅ D3FEND preset
│   │   └── topo-preset.ts              # ✅ Topo-Graph preset
│   ├── validation/
│   │   └── index.ts                    # ✅ Zod schemas
│   ├── errors/
│   │   └── index.tsx                   # ✅ Error boundaries
│   ├── shortcuts/
│   │   ├── index.ts                    # ✅ Shortcut exports
│   │   ├── use-shortcuts.ts            # ✅ Keyboard hooks
│   │   └── shortcuts-dialog.tsx        # ✅ Help dialog
│   ├── io/
│   │   ├── index.ts                    # ✅ IO exports
│   │   ├── graph-io.ts                 # ✅ Save/load/auto-save
│   │   └── exporters.ts                # ✅ PNG/SVG/JSON export
│   ├── share/
│   │   ├── index.ts                    # ✅ Share exports
│   │   └── share.ts                    # ✅ URL encoding/embed
│   ├── d3fend/
│   │   ├── index.ts                    # ✅ D3FEND exports
│   │   └── ontology.ts                 # ✅ Simplified ontology
│   ├── theme/
│   │   ├── index.ts                    # ✅ Theme exports
│   │   └── design-tokens.ts            # ✅ Colors/themes
│   ├── a11y/
│   │   ├── index.ts                    # ✅ A11y exports
│   │   └── focus-management.ts         # ✅ Focus trap/announce
│   └── performance/
│       ├── index.ts                    # ✅ Performance exports
│       └── monitoring.ts               # ✅ FPS/debounce/throttle

cmd/v2local/                            # Existing service - GLC handlers added
├── glc_handlers.go                     # ✅ GLC RPC handlers
├── service.md                          # ✅ Updated with GLC API spec (methods 101-117)
└── main.go                             # ✅ Updated to register GLC handlers

pkg/glc/                                # ✅ NEW: GLC backend package
├── models.go                           # ✅ GraphModel, GraphVersionModel, UserPresetModel, ShareLinkModel
├── store.go                            # ✅ CRUD operations for graphs/versions/presets/links
└── migration.go                        # ✅ GORM AutoMigrate helper
```

### Missing Directories (Not Yet Created)
```
website/lib/glc/
├── stix/                               # ❌ STIX import (J3.4)
├── routing/                            # ❌ Smart edge routing (J3.7)
└── responsive/                         # ❌ Responsive utilities (J4.3)

website/components/glc/
├── preset-editor/                      # ❌ Custom preset wizard (J3.5)
├── examples/                           # ❌ Example gallery (J3.6)
├── d3fend/                             # ❌ D3FEND class browser (J3.1)
├── graph-browser/                      # ❌ Graph list UI (J6.5)
├── versioning/                         # ❌ Version history UI (J5.6)
├── offline/                            # ❌ Offline indicator (J5.7)
└── optimistic/                         # ❌ Conflict resolution (J6.4)

website/app/glc/
└── my-graphs/                          # ❌ Graph browser page (J6.5)

website/assets/
├── d3fend/                             # ❌ Full D3FEND ontology data
└── examples/                           # ❌ Example graphs
```

---

**Document Version**: 5.0
**Last Updated**: 2026-02-10
**Status**: Phase 5 In Progress (J5.1-J5.3 Complete, J5.4-J5.8 Pending)

---

## Open TODO Items

### Phase 3 Remaining Work (J3.1, J3.4-J3.7)

| Job | Priority | Effort | Description |
|-----|----------|--------|-------------|
| J3.1 | HIGH | 2 weeks | D3FEND full ontology loading, class browser UI, inference engine |
| J3.4 | MEDIUM | 2 weeks | STIX 2.1 import with validation and mapping |
| J3.5 | LOW | 3 weeks | Custom preset editor wizard (5-step) |
| J3.6 | LOW | 1 week | Example graphs and gallery |
| J3.7 | LOW | 1 week | Smart edge routing (A* pathfinding) |

### Phase 4 Remaining Work (J4.3, J4.8)

| Job | Priority | Effort | Description |
|-----|----------|--------|-------------|
| J4.3 | MEDIUM | 1 week | Responsive design (mobile/tablet layouts) |
| J4.8 | HIGH | 2 weeks | Unit/component/E2E tests (>80% coverage) |

### Phase 5 Remaining Work (J5.4-J5.8)

| Job | Priority | Effort | Dependencies |
|-----|----------|--------|--------------|
| J5.4 | HIGH | 2 weeks | J5.3 |
| J5.5 | HIGH | 2 weeks | J5.3 |
| J5.6 | MEDIUM | 2 weeks | J5.3, J5.4 |
| J5.7 | MEDIUM | 2 weeks | J5.4, J5.6 |
| J5.8 | HIGH | 2 weeks | J5.4-J5.7 |

### Technical Debt & Enhancements

1. **Phase 2 Context Menus** (J2.8 partial):
   - Node menu: duplicate, edit, delete, color, D3FEND inferences
   - Edge menu: edit, delete, reverse

2. **Phase 3 D3FEND** (J3.1 partial):
   - Lazy load full D3FEND ontology from assets
   - Virtualized class browser component
   - Inference engine for suggesting relationships

3. **Phase 4 Testing** (J4.8):
   - Backend unit tests for GLC handlers (`pkg/glc/*_test.go`)
   - Frontend unit tests for store/validation
   - Component tests for UI
   - E2E tests via Playwright

4. **Phase 5 RPC Client** (J5.3 follow-up):
   - Implement retry with exponential backoff
   - Add offline queue in localStorage
   - Add network status monitoring

5. **Landing Page** (`website/app/glc/page.tsx`):
   - "My Graphs" button is disabled - needs J5.5

6. **Performance Considerations**:
   - Graph thumbnail generation on save (for browser UI)
   - Virtualized list for large graph collections
   - Version history pruning strategy

7. **Accessibility**:
   - Verify WCAG AA compliance for all components
   - Keyboard navigation for graph browser
   - Screen reader testing (NVDA, VoiceOver, TalkBack)
