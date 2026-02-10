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

| Phase | Focus | Duration |
|-------|-------|----------|
| 1 | Core Infrastructure | 8 weeks |
| 2 | Core Canvas Features | 8 weeks |
| 3 | Advanced Features | 8 weeks |
| 4 | UI Polish & Production | 12 weeks |
| 5 | Documentation & Handoff | 4 weeks |
| 6 | Backend Integration | 12 weeks |
| **Total** | | **52 weeks** |

---

## Phase 1: Core Infrastructure

**Location**: `website/glc/`
**Duration**: 8 weeks

### Jobs

#### J1.1 Project Initialization
**Files**: `package.json`, `next.config.ts`, `tsconfig.json`, `tailwind.config.ts`
- Initialize Next.js 15+ with static export (`output: 'export'`)
- Configure TypeScript strict mode
- Setup Tailwind CSS v4
- Install dependencies: @xyflow/react, shadcn/ui, zustand, immer, zod, lucide-react, sonner

#### J1.2 shadcn/ui Setup
**Files**: `components.json`, `components/ui/*.tsx`
- Initialize shadcn/ui
- Add components: button, dialog, dropdown-menu, input, label, sheet, tabs, accordion, toast

#### J1.3 Layout Structure
**Files**: `app/layout.tsx`, `app/glc/page.tsx`, `app/glc/[presetId]/page.tsx`, `components/providers.tsx`
- Root layout with ThemeProvider
- Landing page with hero section
- Dynamic canvas page routes

#### J1.4 State Management
**Files**: `lib/store/index.ts`, `lib/store/slices/*.ts`
- Zustand store with 5 slices:
  - `preset` - currentPreset, builtInPresets, userPresets
  - `graph` - nodes, edges, metadata, viewport
  - `canvas` - selection, zoom, pan
  - `ui` - theme, sidebar, modals
  - `undo-redo` - history stack
- Add persistence middleware (localStorage)
- Add devtools middleware

#### J1.5 Type System
**Files**: `lib/types/*.ts`
- CanvasPreset, NodeTypeDefinition, RelationshipDefinition
- CADNode, CADEdge, GraphMetadata, Graph
- Property, Reference interfaces
- Zod validation schemas for all types

#### J1.6 Built-in Presets
**Files**: `lib/presets/d3fend-preset.ts`, `lib/presets/topo-preset.ts`
- D3FEND preset: 9 node types, 200+ relationships, dark theme
- Topo-Graph preset: 8 node types, 8 relationships, light theme
- Validate with Zod schemas

#### J1.7 Validation System
**Files**: `lib/validation/*.ts`
- Zod schemas for all preset types
- Preset migration system (version compatibility)
- `validatePreset()`, `validateGraph()`

#### J1.8 Error Handling
**Files**: `lib/errors/*.tsx`
- Error boundaries: GraphErrorBoundary, PresetErrorBoundary
- Custom error classes: GLCError, PresetValidationError, GraphValidationError
- Error recovery mechanisms

---

## Phase 2: Core Canvas Features

**Duration**: 8 weeks

### Jobs

#### J2.1 React Flow Canvas
**Files**: `components/canvas/canvas-wrapper.tsx`, `lib/canvas/canvas-config.ts`
- Configure React Flow with preset settings
- Add background grid, controls, mini-map
- Implement node/edge change handlers

#### J2.2 Dynamic Node/Edge Components
**Files**: `components/canvas/dynamic-node.tsx`, `components/canvas/dynamic-edge.tsx`, `components/canvas/*-factory.tsx`
- Preset-aware node rendering (color, icon, properties)
- Preset-aware edge rendering (style, label, arrow)
- React.memo optimization

#### J2.3 Node Palette
**Files**: `components/palette/node-palette.tsx`, `components/palette/node-type-card.tsx`
- Group node types by category (accordion)
- Search/filter functionality
- Drag handle for each node type

#### J2.4 Drag-and-Drop
**Files**: `lib/canvas/drag-drop.ts`, `components/canvas/drop-zone.tsx`
- Cross-browser drag support (react-dnd)
- Position calculation on drop
- Create node with preset defaults

#### J2.5 Node/Edge Editing
**Files**: `components/canvas/node-details-sheet.tsx`, `components/canvas/edge-details-sheet.tsx`
- Node: label, D3FEND class, properties, colors
- Edge: relationship type, label
- Save/Cancel with validation

#### J2.6 Performance Optimization
**Files**: `lib/performance/*.ts`
- Node virtualization (render only visible nodes)
- React.memo, useCallback, useMemo throughout
- Batched state updates
- FPS monitoring

#### J2.7 Keyboard Shortcuts
**Files**: `lib/shortcuts/*.ts`, `components/shortcuts/shortcuts-dialog.tsx`
- Delete, Ctrl+Z (undo), Ctrl+Shift+Z (redo)
- Ctrl+C/V (copy/paste), Escape (clear)
- F (fit view), +/- (zoom)

#### J2.8 Context Menus
**Files**: `components/context-menu/*.tsx`
- Node menu: duplicate, edit, delete, color, D3FEND inferences
- Edge menu: edit, delete, reverse
- Canvas menu: paste, undo, redo, reset view

#### J2.9 Undo/Redo System
**Files**: `lib/undo-redo/*.ts`, `components/toolbar/undo-redo-controls.tsx`
- Save state snapshots on changes
- Limit history size (from preset)
- Ctrl+Z / Ctrl+Shift+Z shortcuts

#### J2.10 Canvas Toolbar
**Files**: `components/toolbar/canvas-toolbar.tsx`
- Undo/Redo, Zoom In/Out, Fit View
- Toggle Grid, Toggle Mini-map
- Save, Share, Export, Help buttons

---

## Phase 3: Advanced Features

**Duration**: 8 weeks

### Jobs

#### J3.1 D3FEND Ontology Integration
**Files**: `lib/d3fend/*.ts`, `components/d3fend/*.tsx`
- Lazy loading D3FEND data
- Virtualized tree for class browser
- Class picker with search
- Inference engine (sensors, defensive techniques, weakness)

#### J3.2 Graph Save/Load
**Files**: `lib/io/graph-io.ts`, `lib/io/graph-schema.ts`
- JSON format with version field
- Save to localStorage and file
- Load from localStorage and file
- Auto-save with debouncing

#### J3.3 Graph Export
**Files**: `lib/io/exporters/*.ts`
- PNG export (html2canvas)
- SVG export (React Flow toSvg)
- PDF export (jsPDF)
- Export dialog with options

#### J3.4 STIX 2.1 Import
**Files**: `lib/stix/*.ts`, `components/stix/stix-import-dialog.tsx`
- Parse STIX JSON
- Validate objects and relationships
- Map to GLC graph structure
- Map to D3FEND ontology

#### J3.5 Custom Preset Editor
**Files**: `components/preset-editor/*.tsx`
- 5-step wizard: Basic Info → Node Types → Relationships → Styling → Behavior
- Live preview
- Auto-save drafts
- Icon/color pickers

#### J3.6 Example Graphs
**Files**: `assets/examples/example-graphs.json`, `components/examples/*.tsx`
- D3FEND examples: attack chains, defense strategies
- Topo-Graph examples: network topology, process flows
- Example gallery with thumbnails

#### J3.7 Smart Edge Routing
**Files**: `lib/routing/*.ts`
- Obstacle detection
- A* pathfinding
- Bezier curve generation
- User toggle (on/off)

#### J3.8 Share & Embed
**Files**: `lib/share/*.ts`, `components/share/*.tsx`
- URL encoding with compression
- Share dialog with copy link
- Embed code generator (iframe)

---

## Phase 4: UI Polish & Production

**Duration**: 12 weeks

### Jobs

#### J4.1 Visual Design System
**Files**: `lib/theme/*.ts`
- Design tokens: colors, spacing, typography, shadows
- Component variants: buttons, inputs, cards, badges
- Animation utilities: fade, slide, scale

#### J4.2 Dark/Light Mode
**Files**: `lib/theme/dark-mode.ts`, `lib/theme/light-mode.ts`, `lib/theme/contrast-validator.ts`
- Theme provider with persistence
- Theme toggle (Ctrl+Shift+T)
- Automated contrast validation (WCAG AA: 4.5:1)
- High contrast mode

#### J4.3 Responsive Design
**Files**: `lib/responsive/*.ts`, `components/responsive/*.tsx`
- Breakpoints: xs(0-639), sm(640-767), md(768-1023), lg(1024-1279), xl(1280+)
- Mobile: drawer palette, collapsed mini-map, touch targets 44px
- Tablet/Desktop: optimized layouts

#### J4.4 Animation Optimization
**Files**: `lib/animations/*.ts`
- Reduced motion support (prefers-reduced-motion)
- Hardware acceleration (transform3d)
- FPS monitoring

#### J4.5 Keyboard Navigation
**Files**: `lib/a11y/keyboard-navigation.ts`, `lib/a11y/focus-management.ts`
- Tab through all interactive elements
- Focus trap in modals
- Skip links

#### J4.6 Screen Reader Support
**Files**: `lib/a11y/*.ts`, `components/a11y/*.tsx`
- ARIA attributes on all components
- Live regions for announcements
- Semantic HTML (headings, lists, labels)
- Test with NVDA, VoiceOver, TalkBack

#### J4.7 Performance Optimization
**Files**: `lib/performance/*.ts`
- Code splitting by route/feature
- Lazy loading heavy components
- Bundle target: <500KB
- FCP <2s (landing), LCP <2.5s, 60fps with 100+ nodes

#### J4.8 Testing
**Files**: `__tests__/**/*.test.{ts,tsx}`, `.github/workflows/test.yml`
- Unit tests: store, types, validation, utilities
- Component tests: all UI components
- Integration tests: user workflows
- E2E tests: critical journeys (Playwright)
- Target: >80% coverage

#### J4.9 Production Deployment
**Files**: `scripts/build.sh`, `scripts/deploy.sh`, `lib/monitoring/*.ts`
- Production build optimization
- Analytics tracking
- Error tracking
- Health check endpoint
- Monitoring dashboard

---

## Phase 5: Documentation & Handoff

**Duration**: 4 weeks

### Jobs

#### J5.1 User Documentation
**Files**: `docs/user-guide.md`, `docs/quick-start.md`, `docs/feature-guide/*.md`
- User guide covering all features
- Quick start guide (5 min)
- Feature-specific guides with screenshots

#### J5.2 Tutorials
**Files**: `docs/tutorials/*.md`
- Tutorial 1: First Graph
- Tutorial 2: D3FEND Modeling
- Tutorial 3: Custom Preset
- Tutorial 4: Advanced Features

#### J5.3 Developer Documentation
**Files**: `docs/developer/*.md`
- Architecture overview
- API reference (store, components, utilities)
- Component library docs
- Contributing guide

#### J5.4 Deployment Documentation
**Files**: `docs/deployment/*.md`
- Deployment guide
- Troubleshooting guide
- Maintenance guide
- Runbooks

#### J5.5 Training Materials
**Files**: `docs/training/*.md`
- Video scripts
- Hands-on exercises
- Practice datasets
- Trainer's guide

#### J5.6 Future Roadmap
**Files**: `docs/roadmap/*.md`
- Short-term (3-6 months)
- Medium-term (6-12 months)
- Long-term (12+ months)
- Known issues and workarounds

---

## Phase 6: Backend Integration

**Duration**: 12 weeks
**Service**: `cmd/v2local` (existing - no new subprocess)
**Status**: In Progress

### Overview
GLC backend integration uses the existing `v2local` service, which already handles SQLite storage for CVE/CWE/CAPEC/ATT&CK data and memory cards. No new `v2glc` subprocess is needed.

### Jobs

#### J6.1 Backend Handlers in v2local ✅ COMPLETE
**Files**: `cmd/v2local/glc_handlers.go`, `cmd/v2local/service.md`, `pkg/glc/`
- ✅ Created `pkg/glc/` package with models, store, and migrations
- ✅ Added RPC handlers to v2local:
  - `RPCGLCGraphCreate`, `RPCGLCGraphGet`, `RPCGLCGraphUpdate`, `RPCGLCGraphDelete`, `RPCGLCGraphList`, `RPCGLCGraphListRecent`
  - `RPCGLCVersionGet`, `RPCGLCVersionList`, `RPCGLCVersionRestore`
  - `RPCGLCPresetCreate`, `RPCGLCPresetGet`, `RPCGLCPresetUpdate`, `RPCGLCPresetDelete`, `RPCGLCPresetList`
  - `RPCGLCShareCreateLink`, `RPCGLCShareGetShared`, `RPCGLCShareGetEmbedData`
- ✅ Updated `cmd/v2local/service.md` with API specification (methods 101-117)

#### J6.2 Database Schema in v2local ✅ COMPLETE
**Files**: `pkg/glc/models.go`, `pkg/glc/migration.go`
- ✅ SQLite tables via GORM AutoMigrate:
  - `glc_graphs` - graph metadata and content
  - `glc_graph_versions` - version history for undo/restore
  - `glc_user_presets` - user-defined presets
  - `glc_share_links` - share/embed links
- ✅ Reuses v2local's existing SQLite connection (bookmark.db)

#### J6.3 Frontend RPC Client ✅ COMPLETE
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

#### J6.4 Optimistic UI
**Files**: `lib/optimistic/*.ts`, `components/optimistic/*.tsx`
- Immediate UI updates
- Version conflict detection
- Conflict resolution dialog
- Rollback on failure

#### J6.5 Graph Browser UI
**Files**: `app/my-graphs/page.tsx`, `components/graph-browser/*.tsx`
- Grid/list view with thumbnails
- Search and filters (preset, date, tags)
- Pagination
- Actions: open, duplicate, delete, share

#### J6.6 Versioning & Recovery
**Files**: `lib/versioning/*.ts`, `components/versioning/*.tsx`
- Auto-save (debounced, on idle, before unload)
- Version history with diff
- Restore previous version
- Crash recovery from localStorage

#### J6.7 Offline Support
**Files**: `lib/offline/*.ts`, `components/offline/*.tsx`
- Offline detection
- Operation queueing
- Sync on reconnect
- Conflict resolution

#### J6.8 Testing
**Files**: `__tests__/e2e/backend-integration.spec.ts`, `cmd/v2local/glc_handlers_test.go`
- RPC integration tests via `/restful/rpc` endpoint
- Backend unit tests in v2local
- E2E tests for full lifecycle
- Load tests (100 concurrent users)

#### J6.9 Production Deployment
**Files**: `tool/migrations/` (if schema migration needed)
- Database migrations via existing migration system
- Health checks through v2local
- Monitoring through existing v2sysmon

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
website/glc/
├── app/
│   ├── layout.tsx
│   ├── glc/
│   │   ├── page.tsx              # Landing
│   │   └── [presetId]/
│   │       └── page.tsx          # Canvas
│   └── my-graphs/
│       └── page.tsx              # Graph browser (Phase 6)
├── components/
│   ├── ui/                       # shadcn/ui components
│   ├── canvas/                   # Canvas components
│   ├── palette/                  # Node palette
│   ├── toolbar/                  # Canvas toolbar
│   ├── context-menu/             # Context menus
│   ├── d3fend/                   # D3FEND components
│   ├── preset-editor/            # Custom preset wizard
│   ├── graph-browser/            # Graph list (Phase 6)
│   ├── shortcuts/                # Shortcuts dialog
│   ├── a11y/                     # Accessibility components
│   └── providers.tsx             # Context providers
├── lib/
│   ├── store/                    # Zustand store slices
│   ├── types/                    # TypeScript types
│   ├── presets/                  # Built-in presets
│   ├── validation/               # Zod schemas
│   ├── errors/                   # Error handling
│   ├── canvas/                   # Canvas utilities
│   ├── shortcuts/                # Keyboard shortcuts
│   ├── undo-redo/                # Undo/redo system
│   ├── d3fend/                   # D3FEND integration
│   ├── io/                       # Save/load/export
│   ├── stix/                     # STIX import
│   ├── routing/                  # Edge routing
│   ├── share/                    # Share/embed
│   ├── theme/                    # Theming
│   ├── responsive/               # Responsive utilities
│   ├── a11y/                     # Accessibility utilities
│   ├── performance/              # Performance optimization
│   ├── rpc/                      # RPC client (Phase 6)
│   ├── optimistic/               # Optimistic updates (Phase 6)
│   ├── versioning/               # Versioning (Phase 6)
│   └── offline/                  # Offline support (Phase 6)
├── assets/
│   ├── d3fend/                   # D3FEND ontology data
│   └── examples/                 # Example graphs
└── __tests__/                    # Test files

cmd/v2local/                      # Existing service - GLC handlers added
├── glc_handlers.go               # ✅ GLC RPC handlers
├── service.md                    # ✅ Updated with GLC API spec (methods 101-117)
└── main.go                       # ✅ Updated to register GLC handlers

pkg/glc/                          # ✅ NEW: GLC backend package
├── models.go                     # GraphModel, GraphVersionModel, UserPresetModel, ShareLinkModel
├── store.go                      # CRUD operations for graphs/versions/presets/links
└── migration.go                  # GORM AutoMigrate helper
```

---

**Document Version**: 4.0
**Last Updated**: 2026-02-10
**Status**: Phase 6 In Progress (J6.1-J6.3 Complete)
