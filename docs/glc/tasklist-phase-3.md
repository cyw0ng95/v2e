# GLC Project Implementation Plan - Phase 3: Advanced Features

## Phase Overview

This phase implements advanced features including D3FEND ontology integration, graph operations (save/load/export), and custom preset creation. This enhances GLC with professional capabilities for security modeling and customization.

## Task 3.1: D3FEND Ontology Integration

### Change Estimation (File Level)
- New files: 10-12
- Modified files: 5-6
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~1,200-1,800

### Detailed Work Items

#### 3.1.1 D3FEND Data Loading
**File List**:
- `website/glc/lib/data/d3fend-ontology.ts` - D3FEND ontology data
- `website/glc/lib/utils/d3fend-loader.ts` - D3FEND data loader
- `website/glc/lib/utils/d3fend-mapper.ts` - D3FEND class mapper

**Work Content**:
- Load D3FEND ontology data (classes, relationships, inferences)
- Create D3FEND class tree structure
- Implement class-to-node type mapping
- Implement relationship-to-edge type mapping
- Cache D3FEND data for performance

**Acceptance Criteria**:
1. WHEN D3FEND preset loads, ontology data SHALL be loaded successfully
2. WHEN viewing D3FEND class tree, SHALL show complete hierarchy
3. WHEN mapping class to node, SHALL return correct node type
4. WHEN mapping relationship to edge, SHALL return correct edge type
5. WHEN D3FEND data is cached, subsequent loads SHALL be faster

#### 3.1.2 D3FEND Class Picker
**File List**:
- `website/glc/components/glc/d3fend/class-picker.tsx` - D3FEND class selector
- `website/glc/components/glc/d3fend/class-tree.tsx` - Class tree component

**Work Content**:
- Create tree view of D3FEND classes
- Implement search/filter functionality
- Add class descriptions on hover
- Implement class selection
- Show class hierarchy (parent/child relationships)

**Acceptance Criteria**:
1. WHEN class picker opens, SHALL display D3FEND class tree
2. WHEN user searches class, tree SHALL filter matching classes
3. WHEN user clicks class, SHALL select that class
4. WHEN user hovers class, SHALL show class description
5. WHEN class has children, SHALL show expandable subtree

#### 3.1.3 D3FEND Inference Capabilities
**File List**:
- `website/glc/lib/utils/d3fend-inferences.ts` - Inference engine
- `website/glc/components/glc/d3fend/inference-dialog.tsx` - Inference selection dialog

**Work Content**:
- Implement inference logic for:
  - Add Sensors to Artifacts
  - Add Defensive Techniques to Attacks
  - Add Offensive Techniques to Countermeasures
  - Add Weakness to Artifacts
  - Explode (Parts, Control, Events)
- Create inference result display
- Apply selected inferences to graph

**Acceptance Criteria**:
1. WHEN user right-clicks Artifact node, SHALL show "Add Sensors" option
2. WHEN user selects "Add Sensors", SHALL show available sensor classes
3. WHEN user selects sensors and confirms, sensor nodes SHALL be added to graph
4. WHEN sensors are added, edges SHALL be created to Artifact
5. WHEN inference results are large, SHALL show pagination or scrollable list

#### 3.1.4 STIX 2.1 Import
**File List**:
- `website/glc/lib/utils/stix-importer.ts` - STIX file parser
- `website/glc/components/glc/d3fend/stix-import-dialog.tsx` - Import dialog

**Work Content**:
- Parse STIX 2.1 JSON files
- Map STIX objects to D3FEND nodes
- Map STIX relationships to D3FEND edges
- Handle STIX extensions and custom properties
- Display import summary (nodes/edges created)

**Acceptance Criteria**:
1. WHEN user uploads STIX file, SHALL parse file contents
2. WHEN STIX object is parsed, SHALL create corresponding node
3. WHEN STIX relationship is parsed, SHALL create corresponding edge
4. WHEN import completes, SHALL show summary of created nodes/edges
5. WHEN STIX file has errors, SHALL display error message

---

## Task 3.2: Graph Operations

### Change Estimation (File Level)
- New files: 8-10
- Modified files: 6-7
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~900-1,400

### Detailed Work Items

#### 3.2.1 Save and Load Graph
**File List**:
- `website/glc/lib/utils/graph-io.ts` - Graph I/O utilities
- `website/glc/components/glc/file/save-dialog.tsx` - Save file dialog
- `website/glc/components/glc/file/load-dialog.tsx` - Load file dialog

**Work Content**:
- Implement graph serialization to JSON
- Implement graph deserialization from JSON
- Validate graph structure on load
- Handle different file versions
- Support file browser dialog

**Acceptance Criteria**:
1. WHEN user saves graph, SHALL write valid JSON file
2. WHEN user loads graph, SHALL parse JSON and recreate nodes/edges
3. WHEN graph has invalid structure, SHALL show validation error
4. WHEN loading old version graph, SHALL migrate to current version
5. WHEN save completes, SHALL show success notification

#### 3.2.2 Graph Metadata Editor
**File List**:
- `website/glc/components/glc/file/metadata-dialog.tsx` - Metadata editor
- `website/glc/lib/utils/metadata-validator.ts` - Metadata validation

**Work Content**:
- Create dialog for editing graph metadata
- Implement form fields for:
  - Title
  - Authors
  - Organization
  - Description
  - References (add/remove multiple)
  - Tags
- Add form validation

**Acceptance Criteria**:
1. WHEN user opens metadata dialog, SHALL show current metadata
2. WHEN user edits title and saves, metadata SHALL be updated
3. WHEN user adds reference, reference SHALL appear in list
4. WHEN user deletes reference, reference SHALL be removed
5. WHEN form is invalid, SHALL show validation errors

#### 3.2.3 Export Graph
**File List**:
- `website/glc/lib/utils/graph-exporter.ts` - Export utilities
- `website/glc/components/glc/file/export-dialog.tsx` - Export options dialog

**Work Content**:
- Implement JSON export (native format)
- Implement PNG export (canvas screenshot)
- Implement SVG export (vector format)
- Implement PDF export (requires html2pdf)
- Add export options (size, quality, include metadata)

**Acceptance Criteria**:
1. WHEN user exports as JSON, SHALL download .json file
2. WHEN user exports as PNG, SHALL download .png file
3. WHEN user exports as SVG, SHALL download .svg file
4. WHEN user exports as PDF, SHALL download .pdf file
5. WHEN export fails, SHALL show error message

#### 3.2.4 Share and Embed
**File List**:
- `website/glc/lib/utils/share-utils.ts` - Share utilities
- `website/glc/components/glc/file/share-dialog.tsx` - Share dialog
- `website/glc/components/glc/file/embed-dialog.tsx` - Embed code dialog

**Work Content**:
- Generate shareable URL (encoded graph data)
- Create embed code generator (iframe)
- Add QR code generation (optional)
- Support URL length limits
- Handle share link expiration

**Acceptance Criteria**:
1. WHEN user clicks share, SHALL generate shareable URL
2. WHEN user opens share URL, SHALL load and display graph
3. WHEN user requests embed code, SHALL generate iframe code
4. WHEN graph is too large for URL, SHALL show warning
5. WHEN share link expires, SHALL show error message

#### 3.2.5 Example Graphs
**File List**:
- `website/glc/data/examples/d3fend-malware-analysis.json` - Example D3FEND graph
- `website/glc/data/examples/topo-network-architecture.json` - Example Topo graph
- `website/glc/components/glc/file/example-graphs-dropdown.tsx` - Example selector

**Work Content**:
- Create example D3FEND graph (malware analysis scenario)
- Create example Topo-Graph (network architecture)
- Create example selector dropdown
- Load example graphs on selection

**Acceptance Criteria**:
1. WHEN user opens example dropdown, SHALL show available examples
2. WHEN user selects D3FEND example, SHALL load malware analysis graph
3. WHEN user selects Topo example, SHALL load network architecture graph
4. WHEN example loads, SHALL display graph with all nodes/edges
5. WHEN example is modified, user SHALL be prompted to save changes

---

## Task 3.3: Custom Preset Creation

### Change Estimation (File Level)
- New files: 12-15
- Modified files: 8-10
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~1,400-2,000

### Detailed Work Items

#### 3.3.1 Preset Editor Layout
**File List**:
- `website/glc/app/glc/preset-editor/page.tsx` - Preset editor page
- `website/glc/components/glc/preset-editor/editor-layout.tsx` - Editor layout

**Work Content**:
- Create preset editor page structure
- Implement multi-step wizard (5 steps):
  1. Basic Information
  2. Node Types
  3. Relationship Types
  4. Visual Styling
  5. Behavior Rules
- Add navigation between steps
- Implement progress indicator

**Acceptance Criteria**:
1. WHEN user opens preset editor, SHALL show step 1 (Basic Information)
2. WHEN user completes step and clicks Next, SHALL advance to next step
3. WHEN user clicks Previous, SHALL go back to previous step
4. WHEN progress indicator displays, SHALL show current step
5. WHEN user navigates, form data SHALL be preserved

#### 3.3.2 Basic Information Form
**File List**:
- `website/glc/components/glc/preset-editor/basic-info-form.tsx` - Step 1 form

**Work Content**:
- Create form fields for:
  - Preset name
  - Description
  - Category (dropdown)
  - Version
  - Author (optional)
- Add form validation
- Implement save draft functionality

**Acceptance Criteria**:
1. WHEN user enters preset name, SHALL show real-time validation
2. WHEN user selects category, SHALL update preset metadata
3. WHEN form is valid, Next button SHALL be enabled
4. WHEN user saves draft, data SHALL persist to localStorage
5. WHEN user reloads page, draft SHALL be restored

#### 3.3.3 Node Types Editor
**File List**:
- `website/glc/components/glc/preset-editor/node-types-editor.tsx` - Step 2 editor
- `website/glc/components/glc/preset-editor/node-type-form.tsx` - Node type form
- `website/glc/components/glc/preset-editor/node-type-list.tsx` - Node type list

**Work Content**:
- Create node type list with add/edit/delete
- Implement node type form with fields:
  - Node type ID
  - Name
  - Icon (icon picker)
  - Category
  - Default label
  - Color (color picker)
  - Border color
  - Background color
  - Icon color
- Add edge restrictions configuration
- Support inference capabilities (D3FEND)

**Acceptance Criteria**:
1. WHEN user adds node type, SHALL appear in list
2. WHEN user edits node type, form SHALL populate with current values
3. WHEN user deletes node type, SHALL confirm before removal
4. WHEN user picks icon, SHALL show icon picker dialog
5. WHEN user selects color, SHALL show color picker dialog
6. WHEN node type has edge restrictions, SHALL show restriction form

#### 3.3.4 Relationship Types Editor
**File List**:
- `website/glc/components/glc/preset-editor/relationship-types-editor.tsx` - Step 3 editor
- `website/glc/components/glc/preset-editor/relationship-form.tsx` - Relationship form
- `website/glc/components/glc/preset-editor/relationship-list.tsx` - Relationship list

**Work Content**:
- Create relationship type list with add/edit/delete
- Implement relationship form with fields:
  - Relationship ID
  - Name
  - Category
  - Description
  - Alternative labels (add/remove)
  - Color
  - Line style (solid/dashed/dotted)
  - Arrow style (default/open/filled)
  - Direction (directed/undirected/bidirectional)
- Add source/target node type restrictions

**Acceptance Criteria**:
1. WHEN user adds relationship, SHALL appear in list
2. WHEN user edits relationship, form SHALL populate with current values
3. WHEN user adds alternative label, SHALL appear in label list
4. WHEN user selects line style, SHALL show preview
5. WHEN user selects arrow style, SHALL show preview
6. WHEN relationship has restrictions, SHALL show restriction form

#### 3.3.5 Visual Styling Editor
**File List**:
- `website/glc/components/glc/preset-editor/styling-editor.tsx` - Step 4 editor

**Work Content**:
- Create form for visual styling configuration:
  - Canvas background color
  - Grid settings (enabled, type, size, color)
  - Node border radius
  - Node padding
  - Node shadow (toggle)
  - Edge color, width
  - Selected edge color
  - Font family, size, label color
  - Color palette for node types
- Live preview of styling changes

**Acceptance Criteria**:
1. WHEN user changes background color, preview SHALL update immediately
2. WHEN user toggles grid, preview SHALL show/hide grid
3. WHEN user changes node radius, preview SHALL reflect change
4. WHEN user selects color palette, nodes SHALL use those colors
5. WHEN user selects font, preview SHALL use that font

#### 3.3.6 Behavior Rules Editor
**File List**:
- `website/glc/components/glc/preset-editor/behavior-editor.tsx` - Step 5 editor

**Work Content**:
- Create form for behavior configuration:
  - Pan/zoom permissions (toggles)
  - Min/max zoom levels (sliders)
  - Snap-to-grid (toggle and size input)
  - Node creation/deletion permissions
  - Edge creation/deletion permissions
  - Label editing permission
  - Multi-select permissions
  - Undo/redo settings (toggle and max history size)
- Live preview of behavior changes

**Acceptance Criteria**:
1. WHEN user enables/disables pan, setting SHALL be saved
2. WHEN user sets min/max zoom, values SHALL be validated (min < max)
3. WHEN user enables snap-to-grid, grid size SHALL be saved
4. WHEN user disables node deletion, delete button SHALL be hidden
5. WHEN user sets max history, value SHALL be validated (1-1000)

#### 3.3.7 Preset Save and Export
**File List**:
- `website/glc/lib/utils/preset-exporter.ts` - Preset export utilities
- `website/glc/components/glc/preset-editor/save-dialog.tsx` - Save preset dialog

**Work Content**:
- Implement preset validation before save
- Implement preset export to JSON file
- Implement preset export to URL (for sharing)
- Save custom presets to localStorage
- Add preset versioning

**Acceptance Criteria**:
1. WHEN user clicks save, SHALL validate preset structure
2. WHEN preset is invalid, SHALL show validation errors
3. WHEN preset is valid, SHALL save to localStorage
4. WHEN user exports preset, SHALL download .json file
5. WHEN user creates URL, SHALL generate shareable link

---

## Task 3.4: Preset Management

### Change Estimation (File Level)
- New files: 6-8
- Modified files: 4-5
- Deleted files: 0

### Cost Estimation (LoC Level)
- Code lines: ~500-800

### Detailed Work Items

#### 3.4.1 Preset Manager
**File List**:
- `website/glc/app/glc/preset-manager/page.tsx` - Preset manager page
- `website/glc/components/glc/preset-manager/preset-list.tsx` - Preset list
- `website/glc/components/glc/preset-manager/preset-card.tsx` - Preset preview card

**Work Content**:
- Create preset manager page
- List all presets (built-in + custom)
- Show preset preview cards
- Add preset actions (edit, duplicate, delete, export)
- Implement preset search and filter

**Acceptance Criteria**:
1. WHEN user opens preset manager, SHALL show all available presets
2. WHEN user searches preset, list SHALL filter matching presets
3. WHEN user hovers over preset card, SHALL show action buttons
4. WHEN user duplicates preset, SHALL create copy with "- Copy" suffix
5. WHEN user deletes custom preset, SHALL confirm before deletion
6. WHEN user tries to delete built-in preset, SHALL show error

#### 3.4.2 Preset Import
**File List**:
- `website/glc/lib/utils/preset-importer.ts` - Preset import utilities
- `website/glc/components/glc/preset-manager/import-dialog.tsx` - Import dialog

**Work Content**:
- Implement preset import from JSON file
- Validate preset structure
- Handle version migration
- Check for duplicate IDs
- Show import summary

**Acceptance Criteria**:
1. WHEN user uploads preset file, SHALL parse JSON
2. WHEN preset is valid, SHALL add to preset list
3. WHEN preset has duplicate ID, SHALL offer to rename
4. WHEN preset has old version, SHALL migrate to current version
5. WHEN import fails, SHALL show error message

---

## Phase 3 Overall Acceptance Criteria

### Functional Acceptance
1. WHEN user right-clicks D3FEND Artifact node, SHALL show inference options
2. WHEN user loads STIX file, SHALL create nodes/edges from STIX data
3. WHEN user saves graph, SHALL export valid JSON file
4. WHEN user creates custom preset, SHALL be usable in canvas
5. WHEN user manages presets, SHALL be able to edit/delete custom presets

### Code Quality Acceptance
1. WHEN running `npm run lint`, code SHALL pass ESLint with zero errors
2. WHEN running TypeScript check, SHALL have no type errors
3. WHEN reviewing code, D3FEND integration SHALL be well-organized
4. WHEN reviewing code, preset editor SHALL follow consistent patterns
5. WHEN reviewing code, graph I/O SHALL handle edge cases

### Performance Acceptance
1. WHEN D3FEND ontology loads, SHALL complete in <2 seconds
2. WHEN STIX file is parsed (100 objects), SHALL complete in <1 second
3. WHEN graph is exported to PNG, SHALL complete in <3 seconds
4. WHEN preset is created, SHALL save in <500ms
5. WHEN preset manager loads (50 presets), SHALL render in <1 second

### Usability Acceptance
1. WHEN user uses D3FEND picker, tree SHALL be easy to navigate
2. WHEN user creates custom preset, workflow SHALL be intuitive
3. WHEN user shares graph, URL SHALL be easy to copy
4. WHEN user exports graph, file SHALL be ready to use
5. WHEN user manages presets, actions SHALL be clear and accessible

---

## Phase 3 Deliverables Checklist

### Code Deliverables
- [ ] D3FEND ontology integration
- [ ] D3FEND class picker
- [ ] D3FEND inference capabilities
- [ ] STIX 2.1 import
- [ ] Graph save/load
- [ ] Graph metadata editor
- [ ] Graph export (JSON, PNG, SVG, PDF)
- [ ] Share and embed functionality
- [ ] Example graphs
- [ ] Custom preset editor (5-step wizard)
- [ ] Preset manager

### Documentation Deliverables
- [ ] Phase 3 implementation plan
- [ ] Phase 3 acceptance criteria checklist

---

## Dependencies

- Phase 2 must be completed before starting Phase 3
- Task 3.1 can be developed in parallel with 3.2
- Task 3.3 must be completed before 3.4
- Task 3.2 depends on Phase 2 Task 2.5 (state management)

---

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| D3FEND ontology data size | High | Lazy load classes, implement caching |
| STIX file complexity varies | Medium | Add robust error handling, show detailed errors |
| Custom preset validation complexity | Medium | Use JSON Schema for validation, provide clear error messages |
| File export cross-browser issues | Low | Test in multiple browsers, use well-tested libraries |

---

## Time Estimation

| Task | Estimated Hours |
|------|-----------------|
| 3.1 D3FEND Ontology Integration | 18-22 |
| 3.2 Graph Operations | 16-20 |
| 3.3 Custom Preset Creation | 24-30 |
| 3.4 Preset Management | 8-12 |
| **Total** | **66-84** |
