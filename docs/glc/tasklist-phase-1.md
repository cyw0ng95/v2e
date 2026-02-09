# GLC Project Implementation Plan - Phase 1: Core Infrastructure

## Phase Overview

This phase establishes the core technical foundation of the GLC project, including project initialization, core data models, basic UI framework, and preset system architecture. This is the foundation for all subsequent feature development.

## Task 1.1: Project Initialization and Infrastructure

### Change Estimation (File Level)
- New files: 15-20
- Modified files: 3-5
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~1,500-2,000
- Configuration files: ~300
- Documentation: ~200

### Detailed Work Items

#### 1.1.1 Next.js Project Setup
**File List**:
- `website/glc/package.json` - Project dependency configuration
- `website/glc/next.config.ts` - Next.js configuration
- `website/glc/tsconfig.json` - TypeScript configuration
- `website/glc/tailwind.config.ts` - Tailwind CSS v4 configuration

**Work Content**:
- Initialize Next.js 15+ project
- Configure TypeScript strict mode
- Setup Tailwind CSS v4
- Configure static export (output: 'export')

**Acceptance Criteria**:
1. WHEN executing `npm install`, project SHALL successfully install all dependencies
2. WHEN executing `npm run dev`, development server SHALL start at http://localhost:3000
3. WHEN executing `npm run build`, project SHALL successfully build to `out/` directory
4. WHEN visiting http://localhost:3000, page SHALL display GLC welcome message
5. WHEN running `npm run lint`, code SHALL pass ESLint checks (zero errors)

#### 1.1.2 Core Dependencies Installation
**Work Content**:
- Install @xyflow/react (React Flow)
- Install shadcn/ui component library and dependencies
- Install Lucide React icon library
- Install sonner notification library
- Install react-hook-form and zod form validation
- Install class-variance-authority, clsx, tailwind-merge

**Acceptance Criteria**:
1. WHEN checking generated `package.json`, SHALL contain all specified dependency packages
2. WHEN running `npm list`, SHALL display all dependencies installed without conflicts
3. WHEN importing `@xyflow/react`, code SHALL have no errors

#### 1.1.3 shadcn/ui Component Initialization
**File List**:
- `website/glc/components.json` - shadcn/ui configuration
- `website/glc/components/ui/` - UI component directory structure

**Work Content**:
- Initialize shadcn/ui configuration
- Create basic UI component directories
- Configure component alias paths

**Acceptance Criteria**:
1. WHEN executing `npx shadcn@latest init`, command SHALL complete successfully
2. WHEN checking generated `components.json`, SHALL contain correct configuration
3. WHEN checking `components/ui/` directory, SHALL have basic directory structure

---

## Task 1.2: Core Data Model Definition

### Change Estimation (File Level)
- New files: 6-8
- Modified files: 0
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~800-1,200

### Detailed Work Items

#### 1.2.1 TypeScript Type Definitions
**File List**:
- `website/glc/lib/types/preset.ts` - Preset-related type definitions
- `website/glc/lib/types/node.ts` - Node-related type definitions
- `website/glc/lib/types/edge.ts` - Edge-related type definitions
- `website/glc/lib/types/graph.ts` - Graph-related type definitions
- `website/glc/lib/types/index.ts` - Unified export of all types

**Work Content**:
Define complete TypeScript type system, including:
- CanvasPreset interface
- NodeTypeDefinition interface
- RelationshipDefinition interface
- PresetStyling interface
- PresetBehavior interface
- CADNode interface
- CADEdge interface
- GraphMetadata interface
- CADGraph interface

**Acceptance Criteria**:
1. WHEN TypeScript compiler runs, SHALL have no type errors
2. WHEN importing type definitions, code SHALL be able to use all interfaces correctly
3. WHEN checking `preset.ts`, SHALL contain complete CanvasPreset interface definition (including all required fields)
4. WHEN checking `node.ts`, SHALL contain complete CADNode interface definition
5. WHEN checking `edge.ts`, SHALL contain complete CADEdge interface definition

#### 1.2.2 Built-in Preset Data Definition
**File List**:
- `website/glc/data/presets/d3fend-preset.ts` - D3FEND preset data
- `website/glc/data/presets/topo-graph-preset.ts` - Topo-Graph preset data
- `website/glc/data/presets/index.ts` - Preset data export

**Work Content**:
Create complete data definitions for D3FEND and Topo-Graph built-in presets, including:
- Node type definitions (icons, colors, categories, properties)
- Relationship type definitions (directions, styles, restrictions)
- Visual style configuration (background, grid, node styles)
- Behavior configuration (zoom limits, edit permissions, etc.)

**Acceptance Criteria**:
1. WHEN importing d3fendPreset, SHALL return complete preset object
2. WHEN checking d3fendPreset.nodeTypes, SHALL contain at least 9 node types
3. WHEN checking d3fendPreset.relationshipTypes, SHALL contain relationship type definitions
4. WHEN checking topoGraphPreset.nodeTypes, SHALL contain at least 8 node types
5. WHEN running TypeScript type check, all preset data SHALL conform to CanvasPreset interface

---

## Task 1.3: Basic UI Component Development

### Change Estimation (File Level)
- New files: 12-15
- Modified files: 2-3
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~1,200-1,800

### Detailed Work Items

#### 1.3.1 Landing Page Components
**File List**:
- `website/glc/app/glc/page.tsx` - Landing page
- `website/glc/components/glc/preset-card.tsx` - Preset card component
- `website/glc/components/glc/preset-grid.tsx` - Preset grid component

**Work Content**:
Implement preset selection landing page:
- GLC title and description
- Preset card grid (D3FEND, Topo-Graph)
- Create custom preset button placeholder
- Recent graphs panel placeholder
- Responsive layout

**Acceptance Criteria**:
1. WHEN visiting `/glc`, page SHALL display "Graphized Learning Canvas" title
2. WHEN checking D3FEND preset card, SHALL display preset name, description, and "Open Canvas" button
3. WHEN checking Topo-Graph preset card, SHALL display preset name, description, and "Open Canvas" button
4. WHEN clicking "Open Canvas" button, SHALL navigate to `/glc/d3fend` or `/glc/topo-graph`
5. WHEN viewing on mobile device, layout SHALL be responsive

#### 1.3.2 Canvas Page Layout
**File List**:
- `website/glc/app/glc/[presetId]/page.tsx` - Canvas page route
- `website/glc/components/glc/layout/canvas-layout.tsx` - Canvas layout component
- `website/glc/components/glc/layout/node-palette-sidebar.tsx` - Node palette sidebar (skeleton)

**Work Content**:
Create basic layout structure for canvas page:
- Top navigation bar (preset switcher, menu)
- Left node palette placeholder
- Center canvas area placeholder
- Responsive layout

**Acceptance Criteria**:
1. WHEN visiting `/glc/d3fend`, page SHALL load canvas layout
2. WHEN visiting `/glc/topo-graph`, page SHALL load canvas layout
3. WHEN checking page structure, SHALL contain top navigation bar, left sidebar, center canvas area
4. WHEN viewing on mobile device, layout SHALL be responsive

#### 1.3.3 Basic shadcn/ui Component Installation
**File List**:
- `website/glc/components/ui/button.tsx`
- `website/glc/components/ui/dialog.tsx`
- `website/glc/components/ui/dropdown-menu.tsx`
- `website/glc/components/ui/input.tsx`
- `website/glc/components/ui/label.tsx`
- `website/glc/components/ui/tooltip.tsx`
- `website/glc/components/ui/sheet.tsx`

**Work Content**:
Install required UI components using shadcn CLI

**Acceptance Criteria**:
1. WHEN checking `components/ui/` directory, SHALL contain all listed component files
2. WHEN importing Button component, code SHALL have no errors
3. WHEN rendering Button component, SHALL display correct styles
4. WHEN running TypeScript check, all components SHALL have no type errors

---

## Task 1.4: Preset System Architecture

### Change Estimation (File Level)
- New files: 5-6
- Modified files: 2-3
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~600-900

### Detailed Work Items

#### 1.4.1 Preset Context and Hook
**File List**:
- `website/glc/lib/context/preset-context.tsx` - Preset context
- `website/glc/lib/hooks/use-preset.ts` - Preset Hook

**Work Content**:
Create foundation for preset management system:
- PresetContext provider
- usePreset custom Hook
- Preset loading and switching logic
- Preset data validation

**Acceptance Criteria**:
1. WHEN using usePreset Hook, SHALL return current preset data
2. WHEN switching preset, preset data SHALL update correctly
3. WHEN providing invalid preset ID, system SHALL display error message
4. WHEN multiple components use usePreset, SHALL share same preset state
5. WHEN TypeScript checking, type definitions SHALL be correct

#### 1.4.2 Preset Management Utility Functions
**File List**:
- `website/glc/lib/utils/preset-manager.ts` - Preset management utilities
- `website/glc/lib/utils/preset-validator.ts` - Preset validation utilities

**Work Content**:
Implement preset management utility functions:
- Preset loading (from data files or localStorage)
- Preset validation (structure integrity check)
- Preset serialization/deserialization
- Preset merging and overriding

**Acceptance Criteria**:
1. WHEN loading D3FEND preset, function SHALL return complete preset object
2. WHEN validating valid preset data, function SHALL return true
3. WHEN validating invalid preset data, function SHALL return false and error message
4. WHEN serializing preset object, function SHALL return valid JSON string
5. WHEN deserializing JSON string, function SHALL return preset object

#### 1.4.3 Preset Selector Component
**File List**:
- `website/glc/components/glc/preset/preset-picker.tsx` - Preset selector dialog
- `website/glc/components/glc/preset/preset-badge.tsx` - Preset badge component

**Work Content**:
Implement preset selection and display components:
- Preset selection dialog
- Preset card display
- Preset search and filtering
- Preset switching confirmation

**Acceptance Criteria**:
1. WHEN opening preset selector, SHALL display all available presets
2. WHEN searching preset, SHALL filter and display matching presets
3. WHEN selecting preset and confirming, system SHALL switch to new preset
4. WHEN switching preset, SHALL show confirmation dialog (if there are unsaved changes)
5. WHEN preset switching completes, page SHALL update to display new preset information

---

## Phase 1 Overall Acceptance Criteria

### Functional Acceptance
1. WHEN visiting `/glc`, user SHALL see preset selection landing page
2. WHEN clicking D3FEND preset card, user SHALL navigate to `/glc/d3fend`
3. WHEN visiting `/glc/d3fend` or `/glc/topo-graph`, page SHALL display Canvas layout
4. WHEN using preset selector, user SHALL be able to switch between different presets
5. WHEN loading page, preset data SHALL load and apply correctly

### Code Quality Acceptance
1. WHEN running `npm run build`, project SHALL build successfully without errors
2. WHEN running `npm run lint`, code SHALL pass ESLint checks
3. WHEN running TypeScript type check, SHALL have no type errors
4. WHEN reviewing code, all components SHALL use TypeScript strict typing
5. WHEN reviewing code, components SHALL follow React best practices

### Performance Acceptance
1. WHEN visiting Landing Page, page SHALL complete First Contentful Paint (FCP) within 2 seconds
2. WHEN visiting Canvas Page, page SHALL complete First Contentful Paint (FCP) within 3 seconds
3. WHEN running Lighthouse performance test, performance score SHALL be greater than 90
4. WHEN switching preset, operation SHALL complete within 500ms

### Accessibility Acceptance
1. WHEN using keyboard Tab key, focus SHALL traverse all interactive elements in logical order
2. WHEN using screen reader, page SHALL correctly announce content
3. WHEN page contains interactive elements, SHALL have appropriate ARIA labels
4. WHEN using high contrast mode, all text SHALL maintain readability

---

## Phase 1 Deliverables Checklist

### Code Deliverables
- [x] Next.js 15+ project configuration files
- [x] Complete TypeScript type definitions
- [x] Built-in preset data (D3FEND, Topo-Graph)
- [x] Landing Page components
- [x] Canvas Page basic layout
- [x] Preset system architecture (Context, Hooks, utility functions)
- [x] Preset selector components
- [x] Basic shadcn/ui components

### Documentation Deliverables
- [x] Phase 1 implementation plan document
- [x] Phase 1 acceptance criteria checklist

---

## Dependencies

- Phase 1 is a prerequisite for all subsequent phases
- Task 1.1 must be completed before Task 1.2
- Task 1.2 must be completed before Task 1.3
- Task 1.4 can be developed in parallel with Task 1.3

---

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Next.js 15+ compatibility issues | High | Use stable versions, update dependencies promptly |
| High complexity of TypeScript type definitions | Medium | Reference official documentation, gradually improve types |
| Large preset data volume, complex management | Medium | Use utility functions for automated management |
| UI component library version compatibility | Low | Lock major dependency versions |

---

## Time Estimation

| Task | Estimated Hours |
|------|-----------------|
| 1.1 Project Initialization and Infrastructure | 8-12 |
| 1.2 Core Data Model Definition | 6-8 |
| 1.3 Basic UI Component Development | 10-14 |
| 1.4 Preset System Architecture | 8-10 |
| **Total** | **32-44** |
