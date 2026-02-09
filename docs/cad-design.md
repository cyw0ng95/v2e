# GLC Frontend Design Document

## Executive Summary

This document outlines the design for **Graphized Learning Canvas (GLC)**, a modern, interactive graph-based modeling platform for the v2e vulnerability management system. GLC is designed as a flexible platform supporting multiple customizable canvas presets. The initial release includes two presets:
1. **D3FEND Canvas** - Cyber attack and defense modeling using MITRE D3FEND ontology
2. **Normal Topo-Graph Canvas** - General-purpose graph and topology diagramming

The platform is inspired by MITRE's D3FEND CAD tool but reimagined as a modern, extensible platform with a clean, professional interface built with Next.js, React, and React Flow.

---

## 1. Vision

### 1.1 Project Goal

Create a **Graphized Learning Canvas (GLC)** platform for visual, interactive graph-based modeling of various scenarios using customizable canvas presets. The platform should enable users to:

- **Switch between different canvas presets** tailored for specific use cases
- Visually design and document cyber attack chains (D3FEND preset)
- Create general topology and graph diagrams (Topo-Graph preset)
- Model complex relationships and data flows
- Understand connections between entities
- Export and share graph models
- **Create and customize their own canvas presets** for specialized scenarios

### 1.2 Platform Architecture

The GLC platform is designed as a **multi-tenant canvas system** where:

1. **Core Canvas Engine** - The underlying graph visualization infrastructure (React Flow-based)
2. **Canvas Presets** - Predefined configurations that define:
   - Available node types and their properties
   - Available relationship types
   - Visual styling (colors, layouts, themes)
   - Validation rules
   - Ontology mappings
3. **Preset Management** - Load, create, edit, and share canvas presets
4. **User Graphs** - Individual graph instances created using a specific preset

### 1.3 Available Presets

#### Preset 1: D3FEND Canvas

**Purpose**: Cyber attack and defense modeling using MITRE D3FEND ontology

**Target Users**:
- Security analysts
- Threat hunters
- Incident responders
- Security architects

**Use Cases**:
- Modeling attack chains
- Visualizing defense strategies
- Understanding D3FEND relationships
- Documenting threat models

#### Preset 2: Normal Topo-Graph Canvas

**Purpose**: General-purpose graph and topology diagramming

**Target Users**:
- System architects
- DevOps engineers
- Network planners
- Data analysts
- Anyone needing to visualize relationships

**Use Cases**:
- System architecture diagrams
- Network topologies
- Data flow diagrams
- Process flows
- Mind mapping
- Entity relationship diagrams

### 1.4 Design Philosophy

- **Platform First**: A flexible engine supporting multiple canvas presets
- **Preset-Driven**: Each preset defines a complete modeling language and visual style
- **Visual First**: Graph-based modeling with intuitive drag-and-drop interactions
- **Modern UX**: Clean, minimal interface with focus on the canvas
- **Professional**: Enterprise-grade design suitable for security teams
- **Accessible**: Keyboard navigation, screen reader support, high contrast modes
- **Performant**: Smooth 60fps interactions even with large graphs
- **Extensible**: Easy to add new presets without modifying core code
- **Learning-Focused**: Designed for educational and knowledge visualization purposes

---

## 2. Technology Stack

### 2.1 Core Framework

- **Next.js 15+** with App Router and Static Site Generation
- **React 19** for component architecture
- **TypeScript** for type safety
- **Tailwind CSS v4** for styling

### 2.2 Graph/Canvas Library

- **@xyflow/react (React Flow)** - Primary graph visualization engine
  - Battle-tested for diagram editors
  - Excellent drag-and-drop support
  - Custom node and edge rendering
  - Smooth zoom/pan interactions
  - Mini-map support
  - Background grid patterns

### 2.3 UI Components

- **shadcn/ui (Radix UI primitives)** for:
  - Dialogs and modals
  - Dropdown menus
  - Form inputs
  - Tooltips
  - Sheets/drawers
  - Tabs
  - Alerts

### 2.4 Additional Libraries

- **lucide-react** for consistent iconography
- **sonner** for toast notifications
- **react-hook-form + zod** for form validation
- **class-variance-authority + clsx + tailwind-merge** for styling

---

## 3. User Interface Design

### 3.1 Platform Landing Page

Before entering a specific canvas, users see a landing page where they can:

```
┌─────────────────────────────────────────────────────────────────┐
│  Graphized Learning Canvas (GLC)                               │
│  Visual Learning Platform for Graph-Based Modeling                │
│                                                                │
│  Select a Preset to Start                                     │
│                                                                │
│  ┌─────────────────────┐  ┌─────────────────────┐             │
│  │   D3FEND Canvas    │  │   Topo-Graph       │             │
│  │                    │  │   Canvas            │             │
│  │  Cyber Attack &     │  │                    │             │
│  │  Defense Modeling   │  │  General-Purpose    │             │
│  │                    │  │  Graph &           │             │
│  │  MITRE D3FEND      │  │  Topology          │             │
│  │        ↓           │  │                    │             │
│  │  [Open Canvas]     │  │        ↓           │             │
│  └─────────────────────┘  │  [Open Canvas]     │             │
│                            └─────────────────────┘             │
│                                                                │
│  [+ Create Custom Preset]                                        │
│                                                                │
│  Recent Graphs:                                                │
│  • Malware Analysis (D3FEND)      [Open]                     │
│  • Network Architecture (Topo)      [Open]                     │
│  • Process Flow (Topo)             [Open]                     │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Canvas Page Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  GLC  ▼ D3FEND Preset  File  Edit  View  Preset  Help       │
├─────────┬───────────────────────────────────────────────────────┤
│         │                                                       │
│  Node   │                  Canvas Area                          │
│  Palette│                                                       │
│  (Left) │              [Drag & Drop Area]                       │
│         │                                                       │
│         │                                                       │
│         │                                                       │
├─────────┴───────────────────────────────────────────────────────┤
│  D3FEND Preset  |  d3fend-graph.json  |  Save  Share  Embed  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 Preset Switching

Users can switch presets via a dropdown in menu bar:

```
┌──────────────────────────────────────────────────┐
│ Preset ▼                                       │
├──────────────────────────────────────────────────┤
│  D3FEND Canvas               (Active)           │
│  Topo-Graph Canvas                               │
│  - - - - - - - - - - - - - - - - - - - - -     │
│  + Create Custom Preset...                    │
│  + Manage Presets...                          │
└──────────────────────────────────────────────────┘
```

When switching presets:
- User is prompted to save current graph (if unsaved changes)
- Canvas is reset with new preset configuration
- Node palette updates with new node types
- Relationship picker updates with new relationship types
- Visual styling is applied

### 3.2 Color Palette

#### Dark Mode (Primary)

```css
Background: #09090b (zinc-950)
Surface: #18181b (zinc-900)
Border: #27272a (zinc-800)
Text Primary: #fafafa (zinc-50)
Text Secondary: #a1a1aa (zinc-400)
Accent: #3b82f6 (blue-500)
Accent Hover: #2563eb (blue-600)
```

#### Light Mode (Alternative)

```css
Background: #ffffff
Surface: #f4f4f5 (zinc-100)
Border: #e4e4e7 (zinc-200)
Text Primary: #09090b (zinc-950)
Text Secondary: #71717a (zinc-500)
Accent: #2563eb (blue-600)
```

#### Node Type Colors

- **Attack**: #ef4444 (red-500)
- **Countermeasure**: #22c55e (green-500)
- **Artifact**: #3b82f6 (blue-500)
- **Event**: #a855f7 (purple-500)
- **Agent**: #f59e0b (amber-500)
- **Vulnerability**: #ec4899 (pink-500)
- **Condition**: #6b7280 (gray-500)
- **Note**: #eab308 (yellow-500)
- **Thing**: #14b8a6 (teal-500)

### 3.3 Typography

```css
Font Family: Inter, system-ui, -apple-system, sans-serif
Headings: 600 weight
Body: 400 weight
Small/Meta: 400 weight, smaller size
Monospace: for IDs, technical values
```

---

## 4. Canvas Preset System

### 4.1 Preset Architecture

Each canvas preset is a complete configuration package defining:

```typescript
interface CanvasPreset {
  // Identification
  id: string;                    // Unique preset ID
  name: string;                  // Display name
  description: string;           // Human-readable description
  version: string;               // Semantic version
  author?: string;               // Preset author
  category: PresetCategory;      // Category for organization

  // Node Types
  nodeTypes: NodeTypeDefinition[];

  // Relationship Types
  relationshipTypes: RelationshipDefinition[];

  // Visual Styling
  styling: PresetStyling;

  // Behavior Configuration
  behavior: PresetBehavior;

  // Validation Rules
  validation?: ValidationRules;

  // Ontology Mappings (for intelligent features)
  ontologyMappings?: OntologyMapping[];
}

type PresetCategory =
  | 'cyber-security'
  | 'general-graph'
  | 'network'
  | 'process'
  | 'data-flow'
  | 'custom';

interface NodeTypeDefinition {
  id: string;
  name: string;
  icon: string;
  category: string;
  defaultLabel: string;

  // Properties
  properties?: PropertyDefinition[];

  // D3FEND/Ontology mapping
  d3fendClass?: string;

  // Visual style
  color: string;
  borderColor?: string;
  backgroundColor?: string;
  iconColor?: string;

  // Behavior
  allowMultiple?: boolean;
  allowEdges?: boolean;
  edgeRestrictions?: {
    canConnectTo?: string[];     // Node types this can connect to
    canReceiveFrom?: string[];   // Node types that can connect to this
  };

  // Inference capabilities (D3FEND only)
  inferences?: InferenceCapability[];
}

interface RelationshipDefinition {
  id: string;
  name: string;
  category: string;
  description: string;

  // Alternative labels
  altLabels?: string[];

  // Visual style
  color?: string;
  lineStyle?: 'solid' | 'dashed' | 'dotted';
  arrowStyle?: 'default' | 'open' | 'filled';

  // D3FEND/Ontology mapping
  d3fendProperty?: string;

  // Directionality
  direction?: 'directed' | 'undirected' | 'bidirectional';

  // Restrictions
  fromNodeTypes?: string[];       // Can start from these node types
  toNodeTypes?: string[];         // Can end at these node types
}

interface PresetStyling {
  // Canvas
  backgroundColor?: string;
  grid?: {
    enabled: boolean;
    type: 'dots' | 'lines' | 'cross';
    size: number;
    color: string;
  };

  // Nodes
  nodeBorderRadius?: number;
  nodePadding?: number;
  nodeShadow?: boolean;

  // Edges
  edgeColor?: string;
  edgeWidth?: number;
  selectedEdgeColor?: string;

  // Labels
  fontFamily?: string;
  fontSize?: number;
  labelColor?: string;

  // Color palette for node types
  colorPalette?: Record<string, string>;
}

interface PresetBehavior {
  // Canvas
  allowPan?: boolean;
  allowZoom?: boolean;
  minZoom?: number;
  maxZoom?: number;
  snapToGrid?: boolean;
  gridSize?: number;

  // Editing
  allowNodeCreation?: boolean;
  allowNodeDeletion?: boolean;
  allowEdgeCreation?: boolean;
  allowEdgeDeletion?: boolean;
  allowLabelEditing?: boolean;

  // Multi-select
  allowMultiSelect?: boolean;
  allowDragSelection?: boolean;

  // Undo/Redo
  enableUndo?: boolean;
  maxHistorySize?: number;
}
```

### 4.2 Built-in Presets

#### 4.2.1 D3FEND Canvas Preset

**Purpose**: Cyber attack and defense modeling using MITRE D3FEND ontology

**Target Users**: Security analysts, threat hunters, incident responders

**Node Types**:
- Event (Cyber events)
- Remote Command (ATT&CK techniques)
- Countermeasure (Defensive tactics)
- Artifact (Digital artifacts)
- Agent (Threat actors)
- Vulnerability (CVEs, CWEs)
- Condition (States, conditions)
- Note (Annotations)
- Thing (Custom entities)

**Relationship Types**: 200+ D3FEND relationships (accesses, creates, detects, counters, etc.)

**Styling**: Dark theme with color-coded node types (Attack=red, Countermeasure=green, etc.)

**Special Features**:
- D3FEND class picker with ontology tree
- Inference capabilities (Add Sensors, Defensive Techniques, etc.)
- STIX 2.1 import

#### 4.2.2 Normal Topo-Graph Canvas Preset

**Purpose**: General-purpose graph and topology diagramming

**Target Users**: System architects, DevOps engineers, network planners, data analysts

**Node Types**:
- Entity (General-purpose node for any concept)
- Process (Processes, workflows, actions)
- Data (Data items, information, records)
- Resource (Computing resources, systems, infrastructure)
- Group (Logical grouping, containers, categories)
- Decision (Decision points, branches, conditions)
- Start/End (Flowchart start and end points)
- Note (Annotations and comments)

**Relationship Types**:
- connects (General connection between entities)
- contains (Entity contains another)
- depends-on (Entity depends on another)
- flows-to (Data or process flows to another)
- related-to (General relationship)
- controls (Entity controls another)
- owns (Entity owns another)
- implements (Entity implements another)

**Styling**: Clean, minimal design with color-coded node types:
- Entity: Blue (#3b82f6)
- Process: Purple (#a855f7)
- Data: Green (#22c55e)
- Resource: Amber (#f59e0b)
- Group: Gray (#6b7280)
- Decision: Pink (#ec4899)
- Start/End: Slate (#475569)
- Note: Yellow (#eab308)

**Special Features**:
- Auto-layout algorithms (force-directed, hierarchical, circular)
- Swimlane/grouping support
- Export to various formats (PNG, SVG, PDF)
- Template library for common diagrams (flowcharts, mind maps, org charts)

### 4.3 Custom Preset Creation

Users can create custom presets via a preset editor:

```
┌────────────────────────────────────────────────────────────┐
│ Custom Preset Editor                                      │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  1. Basic Information                                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Name: [My Custom Preset]                            │   │
│  │ Description: [A canvas for modeling...]               │   │
│  │ Category: [custom ▼]                                 │   │
│  │ Version: [1.0.0]                                     │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
│  2. Node Types                                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Node Type   | Icon | Color        | Actions          │   │
│  ├──────────────────────────────────────────────────────┤   │
│  │ Server      | [🖥️] | #3b82f6      | [Edit] [Delete] │   │
│  │ Database    | [🗄️] | #22c55e      | [Edit] [Delete] │   │
│  │ User        | [👤] | #f59e0b      | [Edit] [Delete] │   │
│  │                                             [+ Add] │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
│  3. Relationship Types                                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Relationship | Direction | Color      | Actions       │   │
│  ├──────────────────────────────────────────────────────┤   │
│  │ connects     | Directed   | #a1a1aa    | [Edit] [Del]│   │
│  │ contains     | Directed   | #a1a1aa    | [Edit] [Del]│   │
│  │                                         [+ Add]     │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
│  4. Visual Styling                                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Theme: [Dark ▼]                                     │   │
│  │ Grid: [Enabled ▼]                                    │   │
│  │ Node Radius: [8 px]                                   │   │
│  │ Edge Width: [2 px]                                   │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
│  5. Behavior Rules                                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ ☑ Allow Pan    ☑ Allow Zoom    ☑ Snap to Grid      │   │
│  │ ☑ Undo/Redo    Max History: [50]                    │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
├────────────────────────────────────────────────────────────┤
│                      [Cancel]  [Save Preset]              │
└────────────────────────────────────────────────────────────┘
```

### 4.4 Preset Management

**Features**:
- List all available presets (built-in + custom)
- Preview preset before loading
- Import preset from JSON file
- Export preset to JSON file
- Share preset via URL (for small presets)
- Duplicate preset to create variations
- Delete custom presets
- Reset to built-in presets

**Built-in Presets**:
- D3FEND Canvas (read-only)
- Topo-Graph Canvas (read-only)

**Custom Presets**:
- Fully editable
- Can be created from scratch or duplicated from existing presets
- Can be shared with other users
- Stored in browser localStorage (future: cloud storage)

---

## 5. Core Features

**Node Categories**:

1. **Action Nodes**
   - Event (Cyber events)
   - Remote Command (Attack steps)
   - Countermeasure (Defensive tactics)

2. **Object Nodes**
   - Artifact (Digital artifacts, files, etc.)
   - Agent (Threat actors, persons, entities)

3. **State Nodes**
   - Vulnerability (CVEs, CWEs)
   - Condition (States, conditions)

4. **Miscellaneous**
   - Note (Annotations)
   - Thing (Custom entities)

**UI Design**:
- Vertical sidebar on the left
- Accordion-style expandable categories
- Drag handles on each node type
- Node previews showing icons and labels
- Hover effects with tooltips
- Search/filter input at top

```tsx
// Node Palette Component Structure
<NodePaletteSidebar>
  <SearchInput placeholder="Search nodes..." />
  <Accordion>
    <AccordionItem title="Actions">
      <DraggableNode type="event" icon="Zap" label="Event" />
      <DraggableNode type="remote-command" icon="Terminal" label="Remote Command" />
      <DraggableNode type="countermeasure" icon="Shield" label="Countermeasure" />
    </AccordionItem>
    <AccordionItem title="Objects">
      <DraggableNode type="artifact" icon="File" label="Artifact" />
      <DraggableNode type="agent" icon="User" label="Agent" />
    </AccordionItem>
    // ... more categories
  </Accordion>
</NodePaletteSidebar>
```

### 5.2 Canvas (Main Workspace)

**Description**: Infinite canvas for creating and arranging nodes

**Features**:
- Drag-and-drop from palette
- Pan with middle mouse or space+drag
- Zoom with mouse wheel
- Mini-map in corner
- Grid background (adjustable by preset)
- Snap-to-grid option (preset-defined)
- Multiple selection (shift+click or drag selection)
- Copy/paste nodes
- Undo/redo support

**Preset-Aware Behavior**:
- Grid pattern varies by preset (dots/lines/cross)
- Zoom limits defined by preset
- Snap-to-grid size defined by preset
- Default viewport settings from preset

**Canvas Controls**:
- Floating toolbar with:
  - Zoom in/out buttons
  - Fit to screen button
  - Zoom percentage display
  - Lock/unlock toggle
  - Grid toggle
  - Help button

```tsx
// Canvas Component Structure (Preset-Aware)
<CanvasContainer preset={currentPreset}>
  <ReactFlow
    nodes={nodes}
    edges={edges}
    nodeTypes={getNodeTypes(currentPreset)}  // Dynamic node types from preset
    edgeTypes={getEdgeTypes(currentPreset)}  // Dynamic edge types from preset
    onNodesChange={handleNodesChange}
    onEdgesChange={handleEdgesChange}
    onConnect={handleConnect}
    onDrop={handleDrop}
    onDragOver={handleDragOver}
    minZoom={currentPreset.behavior.minZoom}
    maxZoom={currentPreset.behavior.maxZoom}
    defaultViewport={currentPreset.behavior.defaultViewport}
    snapToGrid={currentPreset.behavior.snapToGrid}
    snapGrid={[currentPreset.behavior.gridSize, currentPreset.behavior.gridSize]}
    fitView
    attributionPosition="bottom-left"
  >
    <Background
      variant={currentPreset.styling.grid?.type || 'dots'}
      gap={currentPreset.styling.grid?.size || 20}
      size={1}
      color={currentPreset.styling.grid?.color}
    />
    <Controls />
    <MiniMap
      nodeColor={currentPreset.styling.minMapNodeColor}
      maskColor="rgba(0, 0, 0, 0.8)"
    />
    <Panel position="top-right">
      <CanvasActions />
    </Panel>
  </ReactFlow>
</CanvasContainer>
```

### 5.3 Node Types

Each node type is defined by the active preset, with specific visual characteristics and editable properties.

#### 5.3.1 Preset-Based Node Rendering

Nodes are dynamically rendered based on the preset's node type definition:

```typescript
// Node component factory - generates components based on preset
const createNodeComponent = (nodeType: NodeTypeDefinition) => {
  return memo(({ data, selected }: NodeProps) => (
    <NodeContainer
      style={{
        backgroundColor: data.customColor || nodeType.backgroundColor,
        borderColor: data.customBorderColor || nodeType.borderColor,
        borderRadius: nodeType.borderRadius || 8,
        boxShadow: selected ? `0 0 0 2px ${nodeType.color}` : undefined,
      }}
    >
      <NodeHeader>
        <NodeIcon
          icon={nodeType.icon}
          color={nodeType.iconColor}
        />
        <NodeLabel>{data.label}</NodeLabel>
      </NodeHeader>

      {/* D3FEND class indicator (if applicable) */}
      {nodeType.d3fendClass && data.d3fendClass && (
        <NodeClassIndicator>{data.d3fendClass}</NodeClassIndicator>
      )}

      {/* Custom properties from preset */}
      {data.properties && data.properties.length > 0 && (
        <NodeProperties>
          {data.properties.map(prop => (
            <PropertyRow key={prop.id}>
              <PropertyKey>{prop.key}</PropertyKey>
              <PropertyValue>{prop.value}</PropertyValue>
            </PropertyRow>
          ))}
        </NodeProperties>
      )}

      <NodeActions>
        <IconButton icon="Copy2" />
        <IconButton icon="Trash2" />
      </NodeActions>

      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
    </NodeContainer>
  ));
};

// Register node types for React Flow
const nodeTypes: Record<string, React.ComponentType> = {};
preset.nodeTypes.forEach(nodeType => {
  nodeTypes[nodeType.id] = createNodeComponent(nodeType);
});
```

**Base Structure**:
```tsx
<NodeContainer>
  <NodeHeader>
    <NodeIcon type={nodeType} />
    <NodeLabel>{label}</NodeLabel>
  </NodeHeader>
  <NodeClassIndicator>{d3fendClass}</NodeClassIndicator>
  <NodeProperties>
    {properties.map(prop => (
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
```

#### 4.3.2 Node Editing

**Click to Select**: Highlights node with blue border

**Double-Click to Edit Label**: Inline editable text field

**Hover**:
- Shows connection handles (dots on edges)
- Shows quick action buttons (duplicate, delete)

**Right-Click Context Menu**:
- Duplicate
- Delete
- Add Sensors (for Artifacts)
- Add Defensive Techniques (for Attacks)
- Add Offensive Techniques (for Countermeasures)
- Add Weakness (for Artifacts)
- Explode (for Artifacts)

**Node Details Panel**:
- Opens when a node is selected
- Shows in the right sidebar or sheet
- Editable fields:
  - ID (auto-generated, editable)
  - Label
  - D3FEND Class (with picker dialog)
  - Properties (add/edit/remove)
  - Custom relationships

```tsx
// Node Details Sheet
<Sheet open={selectedNode !== null}>
  <SheetContent>
    <SheetHeader>
      <SheetTitle>Node Details</SheetTitle>
    </SheetHeader>
    <Form>
      <FormField>
        <FormLabel>ID</FormLabel>
        <Input value={selectedNode.id} />
      </FormField>
      <FormField>
        <FormLabel>Label</FormLabel>
        <Input value={selectedNode.label} />
      </FormField>
      <FormField>
        <FormLabel>D3FEND Class</FormLabel>
        <ClassPicker onSelect={handleClassSelect} />
      </FormField>
      <FormField>
        <FormLabel>Properties</FormLabel>
        <PropertyList>
          {properties.map(prop => (
            <PropertyRow key={prop.id}>
              <Input value={prop.key} />
              <Input value={prop.value} />
              <Button variant="ghost" onClick={handleDeleteProperty}>
                <Trash2 />
              </Button>
            </PropertyRow>
          ))}
          <Button onClick={handleAddProperty}>
            <Plus /> Add Property
          </Button>
        </PropertyList>
      </FormField>
    </Form>
  </SheetContent>
</Sheet>
```

### 5.4 Relationships (Edges)

**Visual Design**:
- Directed edges with arrowheads (style defined by preset)
- Smooth bezier curves
- Label displayed on edge
- Hover highlights edge
- Selected state with black accent

**Preset-Aware Styling**:
- Edge color defined by preset
- Line style (solid/dashed/dotted) from preset
- Arrow style (default/open/filled) from preset
- Width defined by preset

**Creating Relationships**:
1. Select source node (shows handles)
2. Click and drag from a handle to target node
3. Release to create relationship
4. Relationship picker dialog appears (filtered by preset)
5. Select relationship type
6. Edge is created with label

**Relationship Types**: Defined by preset
- **D3FEND Preset**: 200+ D3FEND relationships (accesses, creates, detects, etc.)
- **ATT&CK Preset**: uses, achieves, targets, compromises, detects
- **Network Preset**: connects-to, routes-to, protects, hosts
- **Custom Preset**: User-defined relationships

**Deleting Relationships**:
- Click to select edge
- Press Backspace/Delete
- Or use context menu option

```tsx
// Edge Component (Preset-Aware)
const createEdgeComponent = (preset: CanvasPreset) => {
  return memo(({
    id,
    source,
    target,
    data,
    selected,
  }: EdgeProps) => {
    const relationshipDef = preset.relationshipTypes.find(
      r => r.id === data?.relationshipId
    );

    return (
      <EdgeWithLabel
        style={{
          stroke: relationshipDef?.color || preset.styling.edgeColor,
          strokeWidth: preset.styling.edgeWidth,
          strokeDasharray:
            relationshipDef?.lineStyle === 'dashed'
              ? '5,5'
              : relationshipDef?.lineStyle === 'dotted'
              ? '1,1'
              : undefined,
        }}
        markerEnd={{
          type: relationshipDef?.arrowStyle === 'open'
            ? MarkerType.ArrowOpen
            : relationshipDef?.arrowStyle === 'filled'
            ? MarkerType.ArrowClosed
            : MarkerType.Arrow,
          color: relationshipDef?.color || preset.styling.edgeColor,
        }}
        label={data?.label || relationshipDef?.name}
        labelStyle={{
          fill: preset.styling.labelColor,
          fontSize: preset.styling.fontSize,
        }}
        selected={selected}
      >
        <EdgeLabelRenderer>
          <EdgeLabel
            style={{
              backgroundColor: selected
                ? preset.styling.selectedEdgeColor
                : preset.styling.backgroundColor,
            }}
          >
            {data?.label || relationshipDef?.name}
            <EdgeLabelMenu onClick={handleLabelClick}>
              {/* Relationship picker filtered by preset */}
              <RelationshipPicker
                preset={preset}
                onSelect={handleRelationshipSelect}
              />
            </EdgeLabelMenu>
          </EdgeLabel>
        </EdgeLabelRenderer>
      </EdgeWithLabel>
    );
  });
};
```

### 5.5 Preset Picker / Relationship Picker

#### 5.5.1 Preset Picker Dialog

Available when switching presets or creating new graphs:

```
┌────────────────────────────────────────────────────────┐
│ Select Canvas Preset                      [×]          │
├────────────────────────────────────────────────────────┤
│                                                        │
│  Search presets...                                     │
│                                                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │ ┌────────────┐  D3FEND Canvas                 │  │
│  │ │            │  Cyber Attack & Defense Modeling  │  │
│  │ │   [icon]   │  MITRE D3FEND ontology         │  │
│  │ └────────────┘                                 │  │
│  │                                              [→] │  │
│  ├──────────────────────────────────────────────────┤  │
│  │ ┌────────────┐  MITRE ATT&CK Canvas           │  │
│  │ │            │  Attack Chain & Technique Graph  │  │
│  │ │   [icon]   │  MITRE ATT&CK framework       │  │
│  │ └────────────┘                                 │  │
│  │                                              [→] │  │
│  ├──────────────────────────────────────────────────┤  │
│  │ ┌────────────┐  Network Topology               │  │
│  │ │            │  Network Infrastructure           │  │
│  │ │   [icon]   │  Diagramming tool               │  │
│  │ └────────────┘                                 │  │
│  │                                              [→] │  │
│  └──────────────────────────────────────────────────┘  │
│                                                        │
│  [+ Create Custom Preset]                               │
│                                                        │
├────────────────────────────────────────────────────────┤
│                   [Cancel]              [Select]       │
└────────────────────────────────────────────────────────┘
```

#### 5.5.2 Relationship Picker (Preset-Aware)

Dialog for selecting relationship types from current preset:

```
┌────────────────────────────────────────────────────────┐
│ Select Relationship                        [×]          │
├────────────────────┬───────────────────────────────────┤
│                    │                                   │
│  Filter/Save:      │  Category Tree                  │
│  [_______________] │                                   │
│                    │  ├─ Access                       │
│  List View         │  │  ├─ accesses                 │
│                    │  │  ├─ may-access               │
│  • accesses        │  │  └─ accessed-by              │
│  • creates         │  ├─ Creation                    │
│  • detects         │  │  ├─ creates                 │
│  • counters        │  │  ├─ may-create               │
│  • ...             │  │  └─ created-by               │
│                    │  ├─ Detection                   │
│                    │  │  ├─ detects                 │
│                    │  │  ├─ may-detect               │
│                    │  │  └─ detected-by              │
│                    │  └─ Defense                    │
│                    │     ├─ counters                 │
│                    │     ├─ neutralizes              │
│                    │     └─ may-counter              │
│                    │                                   │
├────────────────────┴───────────────────────────────────┤
│                   [Cancel]  [Select Relationship]     │
└────────────────────────────────────────────────────────┘
```

**Features**:
- Dual-pane view (list + category tree)
- Hierarchical tree visualization
- Keyword filtering
- Hover shows full definition
- Click to select relationship
- Search highlights matching relationships
- **Preset-specific**: Only shows relationships from active preset

### 5.6 Menu Bar

**File Menu**:
- New Graph (with preset selector)
- Clear (with confirmation)
- Load (file picker)
- Import → STIX 2.1 (D3FEND preset only)
- Example Graphs → Dropdown (preset-specific)
- Save
- Save As → JSON, PNG
- Share → URL encoding
- Embed → Embed code generator

**Edit Menu**:
- Graph Metadata → Dialog
- Undo / Redo

**View Menu**:
- Reset View
- Fullscreen

**Preset Menu**:
- Switch Preset → Opens preset picker
- Manage Presets → Opens preset manager
- Create New Preset → Opens preset editor
- Export Current Preset
- Import Preset

**Help Menu**:
- Documentation (link to user guide)
- Keyboard Shortcuts
- About

### 4.7 Graph Metadata Editor

**Fields**:
- Title
- Authors
- Organization
- Description
- References (add/remove multiple)
- Tags

**UI**: Modal dialog with form validation

### 4.8 Status Bar

**Elements**:
- Current filename (clickable to edit)
- Quick save button
- Share button
- Embed button
- Node count
- Edge count
- Last saved timestamp

---

## 5. Core Features

### 5.1 Node Palette (Sidebar)

**Description**: Draggable node templates defined by current preset

**Dynamic Node Types**: The palette automatically adapts to show node types from the active preset:
- **D3FEND Preset**: Shows Event, Attack, Countermeasure, Artifact, etc.
- **ATT&CK Preset**: Shows Tactic, Technique, Actor, Victim, etc.
- **Network Preset**: Shows Router, Switch, Server, etc.
- **Custom Preset**: Shows user-defined node types

**Node Categories** (D3FEND example):

1. **Action Nodes**
   - Event (Cyber events)
   - Remote Command (Attack steps)
   - Countermeasure (Defensive tactics)

2. **Object Nodes**
   - Artifact (Digital artifacts, files, etc.)
   - Agent (Threat actors, persons, entities)

3. **State Nodes**
   - Vulnerability (CVEs, CWEs)
   - Condition (States, conditions)

4. **Miscellaneous**
   - Note (Annotations)
   - Thing (Custom entities)

**UI Design**:
- Vertical sidebar on the left
- Accordion-style expandable categories
- Drag handles on each node type
- Node previews showing icons and labels
- Hover effects with tooltips
- Search/filter input at top
- Preset-specific colors and icons

```tsx
// Node Palette Component Structure (Preset-Aware)
<NodePaletteSidebar preset={currentPreset}>
  <PresetBadge name={currentPreset.name} icon={currentPreset.icon} />
  <SearchInput placeholder="Search nodes..." />
  <Accordion>
    {currentPreset.nodeTypes.map(category => (
      <AccordionItem key={category.id} title={category.name}>
        {category.types.map(nodeType => (
          <DraggableNode
            key={nodeType.id}
            type={nodeType.id}
            icon={nodeType.icon}
            label={nodeType.name}
            color={nodeType.color}
            preset={currentPreset}
          />
        ))}
      </AccordionItem>
    ))}
  </Accordion>
</NodePaletteSidebar>
```

```typescript
// Handle drag start from palette
const onDragStart = (event: ReactDragEvent, nodeType: string) => {
  event.dataTransfer.setData('application/reactflow', nodeType);
  event.dataTransfer.effectAllowed = 'move';
};

// Handle drop on canvas
const onDrop = (event: React.DragEvent) => {
  event.preventDefault();

  const nodeType = event.dataTransfer.getData('application/reactflow');

  if (!nodeType) return;

  const position = reactFlowInstance.project({
    x: event.clientX - reactFlowBounds.left,
    y: event.clientY - reactFlowBounds.top,
  });

  const newNode = {
    id: `node-${Date.now()}`,
    type: nodeType,
    position,
    data: {
      label: `<${nodeType}>`,
      d3fendClass: '',
      properties: [],
    },
  };

  setNodes((nds) => nds.concat(newNode));
};
```

### 5.2 Reordering Nodes on Canvas

- Built-in React Flow drag functionality
- Smooth dragging with constraints
- Visual feedback during drag
- Snap-to-grid option

### 5.3 Drag to Create Connections

- Click handle on source node
- Drag to target node
- Release to create connection
- Relationship picker appears

---

## 6. D3FEND Ontology Integration

### 6.1 Class Mapping

| Node Type | D3FEND Class Prefix | Description |
|-----------|---------------------|-------------|
| Event | `d3f:Event` | Cyber events |
| Remote Command | `d3f:ATTACKEnterprise` | ATT&CK techniques |
| Countermeasure | All D3FEND tactics | Defensive techniques |
| Artifact | `d3f:Artifact` | Digital artifacts |
| Agent | Custom | Threat actors |
| Vulnerability | `d3f:Weakness` | CVEs, CWEs |
| Condition | Custom | States, conditions |
| Note | Custom | Annotations |
| Thing | Custom | Any entity |

### 6.2 Inference Capabilities

**Right-click on node to access inferences**:

1. **Artifacts** → Shows related artifacts
2. **Defensive Techniques** → Shows countermeasures (for Attacks)
3. **Offensive Techniques** → Shows attacks countered (for Countermeasures)
4. **Add Sensors** → Shows monitoring sensors (for Artifacts)
5. **Add Weakness** → Shows potential CWEs (for Artifacts)
6. **Explode All** → Full class inferences (for Artifacts)
7. **Explode Parts** → Child components (for Artifacts)
8. **Explode Control** → Neighbor inferences (for Artifacts)
9. **Events** → Related events (for Artifacts)

**Inference Dialog**:
```
┌────────────────────────────────────────────────────────┐
│ Add Defensive Techniques to [Node Label]      [×]       │
├────────────────────────────────────────────────────────┤
│                                                        │
│  Available techniques that counter this attack:       │
│                                                        │
│  ☑ d3f:FileSystemSensor    monitors File               │
│  ☐ d3f:NetworkSensor       monitors NetworkArtifact     │
│  ☑ d3f:ProcessSensor       monitors Process             │
│  ☐ d3f:EndpointDetection  detects ProcessInjection    │
│                                                        │
│  [3 selected]                                          │
│                                                        │
├────────────────────────────────────────────────────────┤
│              [Cancel]              [Insert (3)]        │
└────────────────────────────────────────────────────────┘
```

### 6.3 Relationship Types

200+ relationships organized by semantic category:
- Access (accesses, may-access)
- Creation (creates, may-create)
- Detection (detects, may-detect)
- Defense (counters, neutralizes)
- Communication (connects, addresses)
- Modification (modifies, updates)
- Execution (executes, runs, invokes)
- And many more...

**Relationship Picker**:
- Categorized tree view
- Search/filter
- Alternative labels shown
- Definition tooltip on hover

---

### 5.7 Status Bar

**Elements**:
- Preset name (clickable to switch)
- Current filename (clickable to edit)
- Quick save button
- Share button
- Embed button
- Node count
- Edge count
- Last saved timestamp

---

## 6. Data Models (Preset-Aware)

### 6.1 Preset Data Structure

```typescript
// Full preset definition (see section 4.1 for complete interface)
interface CanvasPreset {
  id: string;
  name: string;
  description: string;
  version: string;
  category: PresetCategory;
  nodeTypes: NodeTypeDefinition[];
  relationshipTypes: RelationshipDefinition[];
  styling: PresetStyling;
  behavior: PresetBehavior;
  validation?: ValidationRules;
  ontologyMappings?: OntologyMapping[];
}
```

### 6.2 Node Data Structure

```typescript
interface CADNode {
  id: string;
  type: string;                    // References preset's node type ID
  position: { x: number; y: number };
  data: {
    label: string;
    // D3FEND-specific (optional, depending on preset)
    d3fendClass?: string;
    // Custom properties from preset definition
    properties: Property[];
    // Custom class/relationship (for flexibility)
    customClass?: string;
    // Preset-specific styling overrides
    customColor?: string;
    customBorderColor?: string;
  };
  style?: React.CSSProperties;
}

interface Property {
  id: string;
  key: string;
  value: string;
}
```

### 6.3 Edge Data Structure

```typescript
interface CADEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle?: string;
  targetHandle?: string;
  type?: string;
  animated?: boolean;
  style?: React.CSSProperties;
  data?: {
    label: string;
    // References preset's relationship type ID
    relationshipId?: string;
    // Custom relationship (for flexibility)
    customRelationship?: string;
  };
  markerEnd?: {
    type: MarkerType;
    color?: string;
  };
}
```

### 6.4 Graph Metadata

```typescript
interface GraphMetadata {
  title: string;
  authors: string[];
  organization: string;
  description: string;
  references: Reference[];
  tags: string[];
  createdAt: string;
  updatedAt: string;
  // Preset reference
  presetId: string;
  presetVersion: string;
}

interface Reference {
  title: string;
  url: string;
}
```

### 6.5 Complete Graph Structure

```typescript
interface CADGraph {
  metadata: GraphMetadata;
  nodes: CADNode[];
  edges: CADEdge[];
  viewport?: {
    x: number;
    y: number;
    zoom: number;
  };
}
```

---

## 7. Drag and Drop Implementation

### 7.1 Drag from Palette to Canvas

```typescript
// Handle drag start from palette (preset-aware)
const onDragStart = (event: ReactDragEvent, nodeTypeId: string) => {
  event.dataTransfer.setData('application/reactflow', nodeTypeId);
  event.dataTransfer.setData('preset-id', currentPreset.id);
  event.dataTransfer.effectAllowed = 'move';
};

// Handle drop on canvas (preset-aware)
const onDrop = (event: React.DragEvent) => {
  event.preventDefault();

  const nodeTypeId = event.dataTransfer.getData('application/reactflow');
  const presetId = event.dataTransfer.getData('preset-id');

  if (!nodeTypeId || presetId !== currentPreset.id) {
    toast.error('Cannot mix nodes from different presets');
    return;
  }

  // Get node type definition from preset
  const nodeType = currentPreset.nodeTypes.find(nt => nt.id === nodeTypeId);
  if (!nodeType) {
    toast.error('Invalid node type');
    return;
  }

  const position = reactFlowInstance.project({
    x: event.clientX - reactFlowBounds.left,
    y: event.clientY - reactFlowBounds.top,
  });

  const newNode = {
    id: `node-${Date.now()}`,
    type: nodeTypeId,
    position,
    data: {
      label: nodeType.defaultLabel,
      d3fendClass: nodeType.d3fendClass || '',
      properties: [],
    },
  };

  setNodes((nds) => nds.concat(newNode));
};
```

---

## 8. Component Architecture (Preset-Aware)

### 8.1 Page Structure

```
/app
  /cad
    /page.tsx              // Landing page - preset selection
    /[presetId]
      /page.tsx           // Canvas page for specific preset
  /layout.tsx             // App layout with navbar
```

### 8.2 Component Hierarchy

```
<CADLandingPage>
  <PresetGrid>
    <PresetCard preset={d3fendPreset} />
    <PresetCard preset={attackPreset} />
    <PresetCard preset={networkPreset} />
  </PresetGrid>
  <RecentGraphsList />
  <CreateCustomPresetButton />
</CADLandingPage>

<CADCanvasPage presetId={presetId}>
  <TopNavigation />
  <CADLayout>
    <NodePaletteSidebar preset={currentPreset} />
    <CanvasContainer preset={currentPreset}>
      <ReactFlow preset={currentPreset}>
        <DynamicNode type={nodeType} />
        <DynamicEdge />
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>
      <CanvasToolbar />
    </CanvasContainer>
    <NodeDetailsSheet preset={currentPreset} />
  </CADLayout>
  <MenuBar>
    <PresetMenu />
  </MenuBar>
  <StatusBar preset={currentPreset} />
  <Toaster />
</CADCanvasPage>
```

### 8.3 Key Components

1. **CADLandingPage** - Preset selection landing page
2. **PresetCard** - Preset preview card on landing page
3. **PresetGrid** - Grid of available presets
4. **CADCanvasPage** - Main canvas page for a specific preset
5. **NodePaletteSidebar** - Draggable node templates (preset-aware)
6. **CanvasContainer** - React Flow wrapper (preset-aware)
7. **DynamicNode** - Generic node component that adapts to preset
8. **DynamicEdge** - Generic edge component that adapts to preset
9. **PresetPicker** - Dialog for selecting presets
10. **PresetEditor** - Create/edit custom presets
11. **PresetManager** - Manage presets (import/export/delete)
12. **NodeDetailsSheet** - Right sidebar for node editing
13. **RelationshipPicker** - Relationship selector (preset-aware)
14. **InferenceDialog** - Insert inferences (D3FEND preset only)
15. **CanvasToolbar** - Floating controls (zoom, etc.)
16. **MenuBar** - Top menu with Preset menu
17. **StatusBar** - Bottom status bar with preset indicator
18. **ExampleGraphsDropdown** - Load example graphs (preset-specific)
19. **ShareDialog** - URL sharing
20. **EmbedDialog** - Embed code generator

---

## 9. State Management

### 9.1 React State

```typescript
// Preset management
const [currentPreset, setCurrentPreset] = useState<CanvasPreset>(null);
const [availablePresets, setAvailablePresets] = useState<CanvasPreset[]>([]);

// Graph state
const [nodes, setNodes] = useState<CADNode[]>([]);
const [edges, setEdges] = useState<CADEdge[]>([]);
const [selectedNode, setSelectedNode] = useState<CADNode | null>(null);
const [selectedEdge, setSelectedEdge] = useState<CADEdge | null>(null);

// Metadata
const [metadata, setMetadata] = useState<GraphMetadata>(defaultMetadata);

// Viewport and canvas state
const [viewport, setViewport] = useState({ x: 0, y: 0, zoom: 1 });
const [isLocked, setIsLocked] = useState(false);

// Preset-specific behavior
const [showGrid, setShowGrid] = useState(true);
const [snapToGrid, setSnapToGrid] = useState(false);
```

### 9.2 Refs

```typescript
const reactFlowWrapper = useRef<HTMLDivElement>(null);
const reactFlowInstance = useReactFlow();
```

### 9.3 History (Undo/Redo)

```typescript
const [history, setHistory] = useState<GraphState[]>([]);
const [historyIndex, setHistoryIndex] = useState(-1);

const undo = () => {
  if (historyIndex > 0) {
    const previousState = history[historyIndex - 1];
    setNodes(previousState.nodes);
    setEdges(previousState.edges);
    setHistoryIndex(historyIndex - 1);
  }
};

const redo = () => {
  if (historyIndex < history.length - 1) {
    const nextState = history[historyIndex + 1];
    setNodes(nextState.nodes);
    setEdges(nextState.edges);
    setHistoryIndex(historyIndex + 1);
  }
};
```

### 9.4 Preset Management

```typescript
// Load built-in presets
const loadBuiltInPresets = async () => {
  const presets = await Promise.all([
    import('@/presets/d3fend.json'),
    import('@/presets/attack.json'),
    import('@/presets/network.json'),
    import('@/presets/threat-model.json'),
  ]);
  setAvailablePresets(presets.map(p => p.default));
};

// Load custom presets from localStorage
const loadCustomPresets = () => {
  const saved = localStorage.getItem('cad-custom-presets');
  if (saved) {
    const customPresets = JSON.parse(saved);
    setAvailablePresets(prev => [...prev, ...customPresets]);
  }
};

// Switch preset
const switchPreset = (preset: CanvasPreset) => {
  if (nodes.length > 0 && !confirm('Switching presets will clear current graph. Continue?')) {
    return;
  }

  setCurrentPreset(preset);
  setNodes([]);
  setEdges([]);
  setMetadata(defaultMetadata);
};
```

---

---

## 10. File Operations (Preset-Aware)

### 10.1 Save (JSON)

```typescript
const saveGraph = () => {
  const graph: CADGraph = {
    metadata: {
      ...metadata,
      presetId: currentPreset.id,
      presetVersion: currentPreset.version,
    },
    nodes,
    edges,
    viewport: reactFlowInstance.getViewport(),
  };

  const json = JSON.stringify(graph, null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = `${metadata.title || 'cad-graph'}.json`;
  link.click();

  URL.revokeObjectURL(url);
};
```

### 10.2 Load (JSON)

```typescript
const loadGraph = (file: File) => {
  const reader = new FileReader();

  reader.onload = (event) => {
    try {
      const graph: CADGraph = JSON.parse(event.target?.result as string);

      // Check if preset is available
      const preset = availablePresets.find(
        p => p.id === graph.metadata.presetId
      );

      if (!preset) {
        toast.error(
          `Preset "${graph.metadata.presetId}" not found. Please install the required preset.`
        );
        return;
      }

      // Check version compatibility
      if (graph.metadata.presetVersion !== preset.version) {
        toast.warning(
          `Graph was created with preset version ${graph.metadata.presetVersion}, but current version is ${preset.version}. Some features may not work correctly.`
        );
      }

      setCurrentPreset(preset);
      setMetadata(graph.metadata);
      setNodes(graph.nodes);
      setEdges(graph.edges);

      if (graph.viewport) {
        reactFlowInstance.setViewport(graph.viewport);
      }

      toast.success('Graph loaded successfully');
    } catch (error) {
      toast.error('Failed to load graph file');
    }
  };

  reader.readAsText(file);
};
```

### 10.3 Export PNG

```typescript
const exportPNG = async () => {
  const dataUrl = await reactFlowInstance.toImages({
    format: 'png',
    quality: 1,
  });

  const link = document.createElement('a');
  link.href = dataUrl[0];
  link.download = `${metadata.title || 'd3fend-cad'}.png`;
  link.click();
};
```

### 10.4 Share (URL)

```typescript
const shareGraph = () => {
  const graph: CADGraph = {
    metadata,
    nodes,
    edges,
    viewport: reactFlowInstance.getViewport(),
  };

  const compressed = LZString.compressToEncodedURIComponent(JSON.stringify(graph));
  const url = `${window.location.origin}/cad?graph=${compressed}`;

  navigator.clipboard.writeText(url);
  toast.success('Share URL copied to clipboard');
};
```

### 10.5 Embed Code

```typescript
const generateEmbedCode = (options: EmbedOptions) => {
  const graph: CADGraph = {
    metadata,
    nodes,
    edges,
    viewport: reactFlowInstance.getViewport(),
  };

  const compressed = LZString.compressToEncodedURIComponent(JSON.stringify(graph));
  const embedCode = `
<iframe
  src="${window.location.origin}/cad/embed?graph=${compressed}"
  width="${options.width}"
  height="${options.height}"
  style="border: 1px solid #ccc; border-radius: 8px;"
></iframe>
  `.trim();

  return embedCode;
};
```

---

## 11. Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl + S` | Save graph |
| `Ctrl + Z` | Undo |
| `Ctrl + Y` / `Ctrl + Shift + Z` | Redo |
| `Ctrl + C` | Copy selected |
| `Ctrl + V` | Paste |
| `Ctrl + X` | Cut selected |
| `Delete` / `Backspace` | Delete selected |
| `Ctrl + A` | Select all |
| `Ctrl + D` | Deselect all |
| `Escape` | Deselect / Close dialogs |
| `Ctrl + F` | Search nodes |
| `Ctrl + Plus` | Zoom in |
| `Ctrl + Minus` | Zoom out |
| `Ctrl + 0` | Reset zoom |
| `Space + Drag` | Pan canvas |
| `Middle Click + Drag` | Pan canvas |
| `F` | Fit to screen |
| `L` | Lock/unlock canvas |

---

## 12. Accessibility

### 12.1 Keyboard Navigation

- All UI elements accessible via keyboard
- Tab order: Menu bar → Palette → Canvas → Details panel
- Arrow keys for palette navigation
- Enter/Space to select

### 12.2 Screen Reader Support

- ARIA labels on all interactive elements
- Live regions for notifications
- Descriptive labels for nodes and edges
- Alt text for exported images

### 12.3 High Contrast Mode

- Toggle for increased contrast
- Larger fonts option
- Colorblind-friendly palettes

### 12.4 Focus Indicators

- Visible focus rings on all interactive elements
- Clear selection states for nodes and edges

---

## 13. Performance Optimizations

### 13.1 React Flow Optimizations

```typescript
// Use onNodesChange / onEdgesChange instead of full state updates
const onNodesChange = useCallback(
  (changes: NodeChange[]) => {
    setNodes((nds) => applyNodeChanges(changes, nds));
  },
  [setNodes]
);
```

### 13.2 Virtualization

- Use `@tanstack/react-virtual` for large lists
- Lazy load example graphs
- Pagination for large property lists

### 13.3 Debouncing

- Debounce search input
- Debounce auto-save

### 13.4 Memoization

- Memoize node components
- Memoize expensive computations

---

## 14. Responsive Design

### 14.1 Breakpoints

```css
/* Desktop (default) */
@media (min-width: 1280px) {
  /* Full layout with all panels */
}

/* Tablet */
@media (min-width: 768px) and (max-width: 1279px) {
  /* Collapsible palette, smaller canvas */
}

/* Mobile */
@media (max-width: 767px) {
  /* Full-screen canvas, drawer for palette */
}
```

### 14.2 Adaptations

- **Desktop**: Full palette, canvas, details panel
- **Tablet**: Collapsible palette, canvas, sheet for details
- **Mobile**: Drawer for palette, full-screen canvas, modal for details

---

## 15. Example Graphs

### 15.1 Shadowcat (CTI Report)
Models data from a CTI report demonstrating:
- Attack chain visualization
- Evidence linking
- Timeline representation

### 15.2 Bushwalk (Malware Procedures)
Demonstrates:
- Malware execution flow
- File operations
- Network communications

### 15.3 Disk Formatting (Advanced)
Shows:
- Custom ontology extension
- Complex relationships
- Multiple artifact types

---

## 11. Preset Storage and Management

### 11.1 Built-in Presets

Stored as JSON files in codebase:

```
/assets/presets/
  ├── d3fend-canvas.json       # D3FEND Canvas preset
  └── topo-graph-canvas.json    # Topo-Graph Canvas preset
```

Each preset file follows the `CanvasPreset` interface defined in Section 4.1.

### 11.2 Custom Presets

Stored in browser's localStorage:

```typescript
const saveCustomPreset = (preset: CanvasPreset) => {
  const customPresets = JSON.parse(
    localStorage.getItem('glc-custom-presets') || '[]'
  );
  customPresets.push(preset);
  localStorage.setItem('glc-custom-presets', JSON.stringify(customPresets));
};

const loadCustomPresets = (): CanvasPreset[] => {
  const saved = localStorage.getItem('glc-custom-presets');
  return saved ? JSON.parse(saved) : [];
};

const deleteCustomPreset = (presetId: string) => {
  const customPresets = loadCustomPresets();
  const filtered = customPresets.filter(p => p.id !== presetId);
  localStorage.setItem('glc-custom-presets', JSON.stringify(filtered));
};
```
/assets/presets/
  ├── d3fend-cad.json
  ├── attack-cad.json
  ├── network-topology.json
  └── threat-model.json
```

Each preset file follows the `CanvasPreset` interface defined in Section 4.1.

### 11.2 Custom Presets

Stored in browser's localStorage:

```typescript
const saveCustomPreset = (preset: CanvasPreset) => {
  const customPresets = JSON.parse(
    localStorage.getItem('cad-custom-presets') || '[]'
  );
  customPresets.push(preset);
  localStorage.setItem('cad-custom-presets', JSON.stringify(customPresets));
};

const loadCustomPresets = (): CanvasPreset[] => {
  const saved = localStorage.getItem('cad-custom-presets');
  return saved ? JSON.parse(saved) : [];
};

const deleteCustomPreset = (presetId: string) => {
  const customPresets = loadCustomPresets();
  const filtered = customPresets.filter(p => p.id !== presetId);
  localStorage.setItem('cad-custom-presets', JSON.stringify(filtered));
};
```

### 11.3 Preset Sharing

**Share via URL** (for small presets):
```typescript
const sharePreset = (preset: CanvasPreset) => {
  const compressed = LZString.compressToEncodedURIComponent(
    JSON.stringify(preset)
  );
  const url = `${window.location.origin}/glc/presets/load?data=${compressed}`;
  navigator.clipboard.writeText(url);
  toast.success('Preset share URL copied to clipboard');
};
```

**Export to File**:
```typescript
const exportPreset = (preset: CanvasPreset) => {
  const json = JSON.stringify(preset, null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = `${preset.name.toLowerCase().replace(/\s+/g, '-')}-preset.json`;
  link.click();

  URL.revokeObjectURL(url);
};
```

**Import from File**:
```typescript
const importPreset = (file: File) => {
  const reader = new FileReader();

  reader.onload = (event) => {
    try {
      const preset: CanvasPreset = JSON.parse(event.target?.result as string);

      // Validate preset structure
      if (!preset.id || !preset.name || !preset.nodeTypes) {
        throw new Error('Invalid preset structure');
      }

      // Save as custom preset
      saveCustomPreset(preset);
      setAvailablePresets(prev => [...prev, preset]);
      toast.success(`Preset "${preset.name}" imported successfully`);
    } catch (error) {
      toast.error('Failed to import preset: Invalid format');
    }
  };

  reader.readAsText(file);
};
```

---

## 12. Future Enhancements (Backend Integration)

### 12.1 Storage Backend

- User authentication
- Cloud storage for saved graphs
- **Preset marketplace** - Share and discover community presets
- Graph sharing and collaboration
- Version history

### 12.2 Inference Engine

- Real-time D3FEND ontology queries (D3FEND preset only)
- Automated relationship suggestions
- Conflict detection

### 12.3 Import/Export

- STIX 2.1 import (D3FEND preset only)
- Mitre ATT&CK navigator import (ATT&CK preset only)
- CSV export for analysis

### 12.4 Advanced Features

- Graph templates (preset-specific)
- Component libraries (preset-specific)
- Graph validation (preset-specific rules)
- Automated diagram generation from STIX

---

## 13. Implementation Roadmap (Preset-Aware)

### Phase 0: Platform Foundation
- [ ] Design preset architecture and data models
- [ ] Create preset interface definitions
- [ ] Set up Next.js page structure
- [ ] Build preset landing page
- [ ] Implement preset storage (built-in + custom)

### Phase 1: Core Canvas Engine
- [ ] Implement React Flow canvas wrapper
- [ ] Create preset-aware canvas behavior
- [ ] Build dynamic node rendering system
- [ ] Build dynamic edge rendering system
- [ ] Implement preset-based styling

### Phase 2: Built-in Presets
- [ ] Create D3FEND preset with all node/relationship types
- [ ] Create MITRE ATT&CK preset
- [ ] Create Network Topology preset
- [ ] Create Threat Modeling preset

### Phase 3: Preset Management
- [ ] Build preset picker dialog
- [ ] Implement preset switcher
- [ ] Create preset editor
- [ ] Implement preset import/export
- [ ] Add preset sharing via URL

### Phase 4: Canvas Features
- [ ] Create node palette sidebar (preset-aware)
- [ ] Implement drag-and-drop (preset-aware)
- [ ] Add node selection and editing
- [ ] Implement edge creation (preset-aware)
- [ ] Build relationship picker (preset-aware)
- [ ] Add zoom/pan controls
- [ ] Add mini-map

### Phase 5: Advanced Canvas Features
- [ ] Node details panel
- [ ] Context menus
- [ ] Graph metadata editor
- [ ] Menu bar implementation

### Phase 6: File Operations
- [ ] Save/Load JSON (preset-aware)
- [ ] Export PNG
- [ ] URL sharing
- [ ] Embed code generator
- [ ] Example graphs (preset-specific)

### Phase 7: D3FEND Integration (D3FEND Preset Only)
- [ ] Load D3FEND ontology
- [ ] Implement class picker with tree view
- [ ] Relationship picker with definitions
- [ ] Inference dialogs
- [ ] Custom class/relationship support
- [ ] STIX 2.1 import

### Phase 8: Polish
- [ ] Keyboard shortcuts
- [ ] Undo/redo
- [ ] Accessibility improvements
- [ ] Responsive design
- [ ] Dark/light mode toggle
- [ ] Performance optimization

---

## 14. Design Principles Summary

1. **Platform First**: Flexible engine supporting multiple canvas presets
2. **Preset-Driven**: Each preset defines complete modeling language and style
3. **Visual Clarity**: Clean, uncluttered interface with focus on graph
4. **Discoverability**: Intuitive drag-and-drop, clear visual cues
5. **Efficiency**: Keyboard shortcuts, quick actions, smooth interactions
6. **Professional**: Enterprise-grade design suitable for security teams
7. **Accessible**: WCAG 2.1 AA compliant, keyboard navigation, screen reader support
8. **Performant**: 60fps animations, optimized rendering
9. **Extensible**: Easy to add new presets without modifying core code
10. **Customizable**: Users can create and share their own presets

---

## Appendix: Mockups

### A1. Platform Landing Page
```
┌─────────────────────────────────────────────────────────────────┐
│  Graphized Learning Canvas (GLC)                               │
│  Visual Learning Platform for Graph-Based Modeling                │
│                                                                │
│  Select a Preset to Start                                     │
│                                                                │
│  ┌─────────────────────┐  ┌─────────────────────┐             │
│  │   D3FEND Canvas    │  │   Topo-Graph       │             │
│  │                    │  │   Canvas            │             │
│  │  Cyber Attack &     │  │                    │             │
│  │  Defense Modeling   │  │  General-Purpose    │             │
│  │                    │  │  Graph &           │             │
│  │  MITRE D3FEND      │  │  Topology          │             │
│  │        ↓           │  │                    │             │
│  │  [Open Canvas]     │  │        ↓           │             │
│  └─────────────────────┘  │  [Open Canvas]     │             │
│                            └─────────────────────┘             │
│                                                                │
│  [+ Create Custom Preset]                                        │
│                                                                │
│  Recent Graphs:                                                │
│  • Malware Analysis (D3FEND)      [Open]                     │
│  • Network Architecture (Topo)      [Open]                     │
│  • Process Flow (Topo)             [Open]                     │
└─────────────────────────────────────────────────────────────────┘
```

### A2. Canvas Page (D3FEND Preset)
```
┌─────────────────────────────────────────────────────────────────┐
│  GLC  ▼ D3FEND Preset  File  Edit  View  Preset  Help       │
├─────────┬───────────────────────────────────────────────────────┤
│  🔍    │                                                       │
│  Nodes  │                       ┌──────────────┐                │
│  ────── │                       │   Attack     │                │
│  ▶ Action│                       │  Remote Cmd  │                │
│    ⚡ Event│                      └──────┬───────┘                │
│    💻 Cmd │                             │ executes                │
│    🛡️ Counter│                            ▼                       │
│  ▶ Object │                       ┌──────────────┐                │
│    📄 Artifact │                     │  Artifact    │                │
│    👤 Agent │                       │ Executable   │                │
│  ▶ State  │                       └──────┬───────┘                │
│    🔒 Vuln │                             │ accesses               │
│    ⚙️ Condition│                            ▼                       │
│  ▶ Misc   │                       ┌──────────────┐                │
│    📝 Note │                       │   Event     │                │
│    📦 Thing│                       │  Process    │                │
│         │                       └──────────────┘                │
│         │                                                       │
├─────────┴───────────────────────────────────────────────────────┤
│  D3FEND Preset | d3fend-graph.json | Save | Share | Embed     │
└─────────────────────────────────────────────────────────────────┘
```

### A3. Canvas Page (Topo-Graph Preset)
```
┌─────────────────────────────────────────────────────────────────┐
│  GLC  ▼ Topo-Graph Preset  File  Edit  View  Preset  Help  │
├─────────┬───────────────────────────────────────────────────────┤
│  🔍    │                                                       │
│  Nodes  │                       ┌──────────────┐                │
│  ────── │                       │   Entity    │                │
│  ▶ Basic │                       │  Process    │                │
│    📦 Entity │                      └──────┬───────┘                │
│    ⚙️ Process│                             │ flows-to              │
│    📊 Data  │                             ▼                       │
│  ▶ Struct │                       ┌──────────────┐                │
│    🗂️ Group  │                       │   Data      │                │
│    💡 Decision│                      └──────────────┘                │
│    🏁 Start  │                                                   │
│    🏁 End    │                                                   │
│  ▶ Misc   │                                                   │
│    📝 Note  │                                                   │
│         │                                                   │
│         │                                                   │
├─────────┴───────────────────────────────────────────────────────┤
│  Topo-Graph | my-flowchart.json | Save | Share | Export        │
└─────────────────────────────────────────────────────────────────┘
```

### A4. Node Palette (D3FEND Preset)
```
┌─────────────────────┐
│ D3FEND Preset      │
├─────────────────────┤
│ 🔍 Search nodes...  │
├─────────────────────┤
│ ▶ Actions          │
│   ⚡ Event         │
│   💻 Remote Cmd    │
│   🛡️ Countermeasure│
│ ▶ Objects          │
│   📄 Artifact      │
│   👤 Agent         │
│ ▶ States           │
│   🔒 Vulnerability │
│   ⚙️ Condition     │
│ ▶ Misc             │
│   📝 Note          │
│   📦 Thing         │
└─────────────────────┘
```

### A5. Node Palette (Topo-Graph Preset)
```
┌─────────────────────┐
│ Topo-Graph Preset   │
├─────────────────────┤
│ 🔍 Search nodes...  │
├─────────────────────┤
│ ▶ Basic            │
│   📦 Entity        │
│   ⚙️ Process       │
│   📊 Data          │
│ ▶ Struct           │
│   🗂️ Group         │
│   💡 Decision      │
│   🏁 Start         │
│   🏁 End           │
│ ▶ Misc             │
│   📝 Note          │
└─────────────────────┘
```

### A6. Canvas with Nodes (D3FEND Preset)
```
                      ┌──────────────┐
                      │   Attack     │
                      │  Remote Cmd  │
                      └──────┬───────┘
                             │ executes
                             ▼
                      ┌──────────────┐
                      │  Artifact    │
                      │ Executable   │
                      └──────┬───────┘
                             │ accesses
                             ▼
                      ┌──────────────┐
                      │   Event     │
                      │  Process    │
                      └──────────────┘
```

### A7. Canvas with Nodes (Topo-Graph Preset)
```
         ┌──────────┐
         │   Entity │
         └────┬─────┘
              │ flows-to
              ▼
    ┌─────────────────┐
    │    Process      │
    └────┬──────┬────┘
         │      │
    depends│      │contains
         ▼      ▼
    ┌────────┐ ┌────────┐
    │  Data  │ │ Group  │
    └────────┘ └────────┘
```

### A8. Node Details (D3FEND Preset)
```
┌──────────────────────────────────┐
│ Node Details (D3FEND)      ×   │
├──────────────────────────────────┤
│ ID: node-12345                   │
│ Label: Malware.exe               │
│ Preset: D3FEND Canvas           │
│ Node Type: Artifact              │
│                                │
│ D3FEND Class:                 │
│ [d3f:ExecutableBinary]   Select │
│                                │
│ Properties:                    │
│ MD5: abc123...              [×] │
│ Size: 2.5 MB                [×] │
│ [+ Add Property]               │
├──────────────────────────────────┤
│                [Cancel] [Save]   │
└──────────────────────────────────┘
```

### A9. Node Details (Topo-Graph Preset)
```
┌──────────────────────────────────┐
│ Node Details (Topo-Graph)   ×   │
├──────────────────────────────────┤
│ ID: node-12345                   │
│ Label: Database Server           │
│ Preset: Topo-Graph Canvas       │
│ Node Type: Entity              │
│                                │
│ Properties:                    │
│ Type: MySQL                 [×] │
│ Version: 8.0               [×] │
│ Status: Running              [×] │
│ [+ Add Property]               │
├──────────────────────────────────┤
│                [Cancel] [Save]   │
└──────────────────────────────────┘
```

### A10. Preset Picker Dialog
```
┌────────────────────────────────────────────────────────┐
│ Select Canvas Preset                      [×]          │
├────────────────────────────────────────────────────────┤
│                                                        │
│  🔍 Search presets...                                 │
│                                                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │ ┌────────────┐  D3FEND Canvas                 │  │
│  │ │            │  Cyber Attack & Defense Modeling  │  │
│  │ │   🛡️       │  MITRE D3FEND ontology         │  │
│  │ └────────────┘                                 │  │
│  │                                              [→] │  │
│  ├──────────────────────────────────────────────────┤  │
│  │ ┌────────────┐  Topo-Graph Canvas              │  │
│  │ │            │  General-Purpose Graph &       │  │
│  │ │   📊       │  Topology Diagramming         │  │
│  │ └────────────┘                                 │  │
│  │                                              [→] │  │
│  └──────────────────────────────────────────────────┘  │
│                                                        │
│  [+ Create Custom Preset]                               │
│                                                        │
├────────────────────────────────────────────────────────┤
│                   [Cancel]              [Select]       │
└────────────────────────────────────────────────────────┘
```

### A11. Custom Preset Editor
```
┌────────────────────────────────────────────────────────────┐
│ Custom Preset Editor                                      │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  1. Basic Information                                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Name: [My Custom Graph]                              │   │
│  │ Description: [A canvas for modeling...]              │   │
│  │ Category: [custom ▼]                                 │   │
│  │ Version: [1.0.0]                                     │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
│  2. Node Types                                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Node Type   | Icon | Color        | Actions          │   │
│  ├──────────────────────────────────────────────────────┤   │
│  │ MyEntity   | [📦] | #3b82f6      | [Edit] [Delete] │   │
│  │ MyProcess  | [⚙️]  | #a855f7      | [Edit] [Delete] │   │
│  │ MyData     | [📊] | #22c55e      | [Edit] [Delete] │   │
│  │                                             [+ Add] │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
│  3. Relationship Types                                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Relationship | Direction | Color      | Actions       │   │
│  ├──────────────────────────────────────────────────────┤   │
│  │ connects     | Directed   | #a1a1aa    | [Edit] [Del]│   │
│  │ contains     | Directed   | #a1a1aa    | [Edit] [Del]│   │
│  │                                         [+ Add]     │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
│  4. Visual Styling                                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Theme: [Dark ▼]                                     │   │
│  │ Grid: [Enabled ▼]                                    │   │
│  │ Node Radius: [8 px]                                   │   │
│  │ Edge Width: [2 px]                                   │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                            │
├────────────────────────────────────────────────────────────┤
│                      [Cancel]  [Save Preset]              │
└────────────────────────────────────────────────────────────┘
```

---

**Document Version**: 2.0
**Last Updated**: 2025-02-09
**Status**: Design Complete - Platform Architecture with Customizable Presets
**Changes**: v2.0 - Renamed to Graphized Learning Canvas (GLC), added platform architecture with preset system, 2 built-in presets (D3FEND, Topo-Graph), custom preset creation, preset management
