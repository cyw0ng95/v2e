# GLC Architecture Document

## Overview

GLC (Graphized Learning Canvas) is a React-based interactive graph modeling platform built on Next.js 15, following a broker-first microservices architecture compatible with the v2e vulnerability management system.

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Browser / Client                       │
│                                                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │              Next.js 15 (Frontend)              │   │
│  │                                                 │   │
│  │  ┌──────────────┐  ┌─────────────────────┐    │   │
│  │  │  Pages       │  │   Components        │    │   │
│  │  │  - Landing   │  │   - GLC Components  │    │   │
│  │  │  - Canvas    │  │   - UI Components    │    │   │
│  │  └──────────────┘  └─────────────────────┘    │   │
│  │                                                 │   │
│  │  ┌──────────────────────────────────────────┐  │   │
│  │  │         State Management (Zustand)      │  │   │
│  │  │  ┌──────────┬──────────┬──────────┐  │  │   │
│  │  │  │ Preset   │ Graph    │ Canvas   │  │  │   │
│  │  │  │ Slice    │ Slice    │ Slice    │  │  │   │
│  │  │  └──────────┴──────────┴──────────┘  │  │   │
│  │  └──────────────────────────────────────────┘  │   │
│  │                                                 │   │
│  │  ┌──────────────────────────────────────────┐  │   │
│  │  │              Business Logic            │  │   │
│  │  │  - Preset Manager                     │  │   │
│  │  │  - Validation System                   │  │   │
│  │  │  - Error Handling                     │  │   │
│  │  └──────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────┘   │
│                                                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │         Local Storage (Browser)                │   │
│  │  - Presets                                   │   │
│  │  - Graphs                                    │   │
│  │  - Settings                                  │   │
│  │  - Error Logs                                │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                    │
                    │ (Future Phase 6)
                    ▼
        ┌──────────────────────┐
        │   v2e Backend       │
        │   (GLC Service)     │
        │   - SQLite DB       │
        │   - RPC API         │
        └──────────────────────┘
```

### Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Framework** | Next.js 15 (App Router) | React framework with SSG |
| **UI Library** | React 19 | Core UI framework |
| **Language** | TypeScript 5 | Type-safe development |
| **Styling** | Tailwind CSS v4 | Utility-first CSS |
| **Components** | shadcn/ui | Pre-built UI components |
| **State Management** | Zustand | Lightweight state manager |
| **Graph Library** | @xyflow/react (React Flow) | Interactive canvas |
| **Validation** | Zod | Schema validation |
| **Forms** | react-hook-form | Form handling |
| **Icons** | Lucide React | Icon library |
| **Notifications** | Sonner | Toast notifications |

## Core Modules

### 1. Type System (`lib/glc/types/`)

**Purpose**: Centralized type definitions for all GLC entities

**Key Types**:
- `CanvasPreset`: Complete preset definition
- `NodeTypeDefinition`: Node type specification
- `RelationshipDefinition`: Relationship type specification
- `CADNode`: Canvas node instance
- `CADEdge`: Canvas edge instance
- `Graph`: Complete graph structure
- `GraphMetadata`: Graph metadata

**Design Principles**:
- No `any` types
- Full TypeScript strict mode
- Runtime validation with Zod

### 2. State Management (`lib/glc/store/`)

**Purpose**: Centralized state management with slice-based architecture

**Architecture**:
```
Store
├── Preset Slice (preset management)
├── Graph Slice (nodes, edges, metadata)
├── Canvas Slice (canvas interactions)
├── UI Slice (theme, panels, modals)
└── UndoRedo Slice (history, undo/redo)
```

**Middleware**:
- `devtools`: Redux DevTools integration
- `persist`: localStorage persistence

**Key Features**:
- Typed actions
- Immer for immutable updates
- History stack for undo/redo

### 3. Validation System (`lib/glc/validation/`)

**Purpose**: Ensure data integrity and provide clear error messages

**Components**:
- `validators.ts`: Zod schemas and validation logic
- `migrations.ts`: Version migration system

**Key Features**:
- Preset validation (nodes, edges, styling)
- Graph validation (node IDs, edge connections)
- Version migrations (0.9.0 → 1.0.0)
- Detailed error reporting

### 4. Error Handling (`lib/glc/errors/`)

**Purpose**: Robust error handling with recovery mechanisms

**Components**:
- `error-types.ts`: Custom error classes
- `error-handler.ts`: Centralized error processing
- `error-boundaries.tsx`: React error boundaries

**Error Types**:
- `GLCError`: Base error class
- `PresetValidationError`: Preset validation errors
- `GraphValidationError`: Graph validation errors
- `StateError`: State management errors
- `NetworkError`: Network-related errors

**Features**:
- Error logging (localStorage + console)
- Toast notifications
- Error log export
- React error boundaries

### 5. Preset System (`lib/glc/presets/` + `preset-manager.ts`)

**Purpose**: Manage canvas presets and user customizations

**Built-in Presets**:
- `D3FEND_PRESET`: MITRE D3FEND ontology (9 node types, 8 relationships)
- `TOPO_PRESET`: General-purpose topology (8 node types, 8 relationships)

**Preset Manager Features**:
- CRUD operations for user presets
- Import/export (JSON)
- Backup system (automatic snapshots)
- Version migration

### 6. Graph Management (`lib/glc/store/slices/graph.ts`)

**Purpose**: Manage graph state and operations

**Operations**:
- Node CRUD (add, update, delete)
- Edge CRUD (add, update, delete)
- Selection management
- Viewport management
- Metadata management

**Key Features**:
- Cascading deletes (deleting node removes edges)
- Position validation
- Type validation

## Data Flow

### Preset Loading Flow

```
User selects preset
    ↓
Store.setCurrentPreset(preset)
    ↓
Validate preset (Zod)
    ↓
Migrate if needed
    ↓
Update store state
    ↓
Persist to localStorage
    ↓
Trigger UI re-render
```

### Graph Operation Flow

```
User action (add node/edge)
    ↓
Action dispatched to store
    ↓
Validate operation
    ↓
Push to history stack
    ↓
Update store state (immer)
    ↓
Persist to localStorage
    ↓
Trigger UI re-render
```

### Error Handling Flow

```
Error occurs
    ↓
Error caught
    ↓
Classify error type
    ↓
Log to console
    ↓
Save to localStorage
    ↓
Show toast notification
    ↓
Offer recovery options
```

## Performance Considerations

### Optimizations

1. **State Updates**: Immer for efficient immutable updates
2. **Persistence**: Selective persistence (only user data)
3. **Validation**: Cached validation results
4. **Rendering**: React 19 concurrent features

### Limits

- Max nodes per graph: 1000 (configurable)
- Max edges per graph: 2000 (configurable)
- History stack: 50 items
- Error log: 100 entries

## Security Considerations

### Input Validation

- All user inputs validated with Zod schemas
- Preset files validated before import
- Graph data validated on load

### Data Storage

- No sensitive data in localStorage
- Backup system limits retention
- Error logs exclude sensitive context

## Integration Points

### v2e Backend (Phase 6)

```
Frontend (GLC)
    ↓ HTTP/RPC
v2e Access Gateway
    ↓ IPC/UDS
v2e Broker
    ↓ IPC/UDS
GLC Service
    ↓
SQLite Database
```

### External APIs (Phase 3)

- D3FEND ontology (MITRE)
- NVD API (CVE data)
- STIX 2.1 (import/export)

## Future Enhancements

### Phase 2: Core Canvas Features
- React Flow integration
- Node palette
- Canvas interactions
- Mini-map and controls

### Phase 3: Advanced Features
- D3FEND inference
- STIX import
- Custom preset editor
- Graph export (PNG, SVG, PDF)

### Phase 6: Backend Integration
- GLC service (Go)
- SQLite database
- RPC API
- Graph sharing

## Testing Strategy

### Unit Tests
- Store operations
- Validation logic
- Utility functions
- Error handling

### Integration Tests
- Preset loading/switching
- Graph operations
- Error recovery
- Persistence

### E2E Tests (Phase 4)
- User flows
- Canvas interactions
- Import/export
- Error scenarios

---

**Document Version**: 1.0
**Last Updated**: 2026-02-09
**Status**: Phase 1 Complete
