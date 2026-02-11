# CVSS Calculator Feature Design

## Overview

A CVSS (Common Vulnerability Scoring System) calculator for v2e that supports CVSS versions 3.0, 3.1, and 4.0. The feature provides both standalone calculation capabilities and integration with the existing CVE workflow.

## Requirements

### Functional Requirements

1. **CVSS Version Support**: Implement calculators for CVSS v3.0, v3.1, and v4.0
2. **Dual Mode Operation**:
   - Standalone calculator for researchers to score new vulnerabilities
   - Integration with existing CVE records
3. **Export Functionality**: Support JSON, CSV, Vector String, and shareable URL formats
4. **UI Flexibility**: Mixed mode with v2e design (default) and traditional FIRST.org-style view

### Non-Functional Requirements

1. **Client-Side Calculation**: All CVSS math executes in the browser
2. **Accuracy**: Results must match FIRST.org official calculators exactly
3. **Performance**: Real-time score updates as user changes metrics
4. **Accessibility**: Follow v2e accessibility standards

## Architecture

### Route Structure

```
/cvss                    # CVSS homepage with version selector
/cvss/3.0               # CVSS v3.0 calculator
/cvss/3.1               # CVSS v3.1 calculator
/cvss/4.0               # CVSS v4.0 calculator
```

### Component Hierarchy

```
app/cvss/
├── page.tsx                      # CVSS homepage
└── [version]/
    └── page.tsx                  # Dynamic route for version-specific calculators

components/cvss/
├── cvss-calculator.tsx           # Main calculator component
├── metric-selector.tsx            # Reusable metric dropdown/radio
├── score-display.tsx              # Visual score display with severity colors
├── vector-string.tsx              # Vector string display and copy
├── view-toggle.tsx                # v2e/Traditional view switcher
├── export-menu.tsx                # Export options dropdown
├── version-specific/
│   ├── v3-metrics.tsx            # v3.0/v3.1 metric groups
│   └── v4-metrics.tsx            # v4.0 metric groups
└── recent-calculations.tsx        # History of recent calculations

lib/cvss/
├── calculator.ts                   # Shared calculator interface
├── v3-calculator.ts              # v3.0/v3.1 implementation
├── v4-calculator.ts              # v4.0 implementation
├── metrics.ts                     # Metric definitions and weights
├── utils.ts                      # Vector parsing/formatting
└── constants.ts                  # Version-specific constants
```

## Data Structures

### Core Interfaces

```typescript
// Common score result interface
interface CVSSScore {
  baseScore: number;
  temporalScore?: number;
  environmentalScore?: number;
  severity: 'NONE' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  vectorString: string;
  version: '3.0' | '3.1' | '4.0';
}

// CVSS v3.x metrics
interface V3Metrics {
  // Base Score Metrics
  attackVector: 'N' | 'A' | 'L' | 'P';          // AV
  attackComplexity: 'L' | 'H';                    // AC
  privilegesRequired: 'N' | 'L' | 'H';             // PR
  userInteraction: 'N' | 'R';                      // UI
  scope: 'U' | 'C';                                // S
  confidentiality: 'H' | 'L' | 'N';               // C
  integrity: 'H' | 'L' | 'N';                      // I
  availability: 'H' | 'L' | 'N';                  // A

  // Temporal Metrics (optional)
  exploitCodeMaturity?: 'X' | 'H' | 'F' | 'P' | 'U';  // E
  remediationLevel?: 'X' | 'O' | 'W' | 'U';             // RL
  reportConfidence?: 'X' | 'C' | 'R' | 'U';              // RC

  // Environmental Metrics (optional)
  confidentialityRequirement?: 'X' | 'H' | 'M' | 'L';    // CR
  integrityRequirement?: 'X' | 'H' | 'M' | 'L';         // IR
  availabilityRequirement?: 'X' | 'H' | 'M' | 'L';      // AR
  modifiedAttackVector?: 'X' | 'N' | 'A' | 'L' | 'P';  // MAV
  modifiedAttackComplexity?: 'X' | 'L' | 'H';           // MAC
  modifiedPrivilegesRequired?: 'X' | 'N' | 'L' | 'H';    // MPR
  modifiedUserInteraction?: 'X' | 'N' | 'R';             // MUI
  modifiedScope?: 'X' | 'U' | 'C';                      // MS
  modifiedConfidentiality?: 'X' | 'H' | 'L' | 'N';       // MC
  modifiedIntegrity?: 'X' | 'H' | 'L' | 'N';             // MI
  modifiedAvailability?: 'X' | 'H' | 'L' | 'N';          // MA
}

// CVSS v4.0 metrics (significantly different from v3.x)
interface V4Metrics {
  // Base Score Metrics
  attackVector: 'N' | 'A' | 'L' | 'P';          // AV
  attackComplexity: 'L' | 'H';                    // AC
  attackRequirements: 'I' | 'R' | 'X';             // AT
  privilegesRequired: 'N' | 'L' | 'H';             // PR
  userInteraction: 'N' | 'P' | 'A';                // UI
  vulnerableSystemIntegrity: 'H' | 'L' | 'N';       // VI

  // Vulnerability Impact Metrics
  vulnerabilityImpactEfficiency?: 'X' | 'H' | 'M' | 'L'; // VIE

  // Security Requirements (CIA triad)
  confidentialityRequirement?: 'X' | 'H' | 'M' | 'L';  // CR
  integrityRequirement?: 'X' | 'H' | 'M' | 'L';         // IR
  availabilityRequirement?: 'X' | 'H' | 'M' | 'L';      // AR

  // Security Response Metrics
  securityResponseEfficiency?: 'X' | 'H' | 'M' | 'L';  // SRE

  // Provider Urgency (required for CVSS v4.0)
  providerUrgency: 'X' | 'H' | 'M' | 'L' | 'U';  // PU
}
```

### Metric Definitions

```typescript
// V3.0/V3.1 Metric Options
export const V3_BASE_METRICS = {
  AV: {
    N: { label: 'Network', value: 0.85, weight: 1.0 },
    A: { label: 'Adjacent', value: 0.62, weight: 1.0 },
    L: { label: 'Local', value: 0.55, weight: 1.0 },
    P: { label: 'Physical', value: 0.2, weight: 1.0 }
  },
  AC: {
    L: { label: 'Low', value: 0.77, weight: 1.0 },
    H: { label: 'High', value: 0.44, weight: 1.0 }
  },
  PR: {
    N: { label: 'None', value: 0.85, weight: 1.0 },
    L: { label: 'Low', value: 0.68, weight: 1.0 },
    H: { label: 'High', value: 0.50, weight: 1.0 }
  },
  UI: {
    N: { label: 'None', value: 0.85, weight: 1.0 },
    R: { label: 'Required', value: 0.62, weight: 1.0 }
  },
  S: {
    U: { label: 'Unchanged', value: 1.0, changed: false },
    C: { label: 'Changed', value: 1.0, changed: true }
  },
  C_I_A: {
    H: { label: 'High', value: 0.56, weight: 1.0 },
    L: { label: 'Low', value: 0.22, weight: 1.0 },
    N: { label: 'None', value: 0.0, weight: 1.0 }
  }
};

// V4.0 Metric Options (simplified example)
export const V4_BASE_METRICS = {
  AV: { N: 0.85, A: 0.62, L: 0.55, P: 0.2 },
  AC: { L: 0.77, H: 0.44 },
  AT: { I: 1.0, R: 0.5, X: 1.0 },
  // ... additional V4 metrics
};
```

## UI/UX Design

### Layout Structure

```
+----------------------------------------------------------------+
|  CVSS Calculator                                    [v2e ▼][Export] |
+----------------------------------------------------------------+
|                                                                   |
|  +-------------------------------+  +-----------------------------+   |
|  |  Base Metrics (Required)      |  |  Score: 9.8 (CRITICAL)   |   |
|  |                               |  |                             |   |
|  |  Attack Vector: [N ▼]        |  |  [Severity Meter]            |   |
|  |  Attack Complexity: [L ▼]     |  |                             |   |
|  |  ...                          |  +-----------------------------+   |
|  |                               |                                   |
|  +-------------------------------+  Vector String:                |
|                                   CVSS:3.1/AV:N/AC:L/PR:N/...   |
|                                                                   |
|  +-------------------------------+  +-----------------------------+   |
|  |  Temporal Metrics (Optional)  |  [Reset] [Copy] [Export]     |   |
|  |  ...                          |                                   |
|  +-------------------------------+                                   |
+----------------------------------------------------------------+
```

### View Modes

**v2e View (Default)**:
- Uses shadcn/ui Card components
- Compact metric selectors with dropdown menus
- Color-coded severity indicators (matching v2e theme)
- Animated score updates

**Traditional View**:
- Table-based layout similar to FIRST.org
- Arrow flow indicators between metric groups
- Detailed formula explanations
- Text-heavy information density

### Severity Color Mapping

| Score Range | Severity | Color (v2e) | Color (Traditional) |
|------------|----------|----------------|-------------------|
| 0.0 | NONE | Gray | Gray |
| 0.1 - 3.9 | LOW | Blue | Green |
| 4.0 - 6.9 | MEDIUM | Yellow | Yellow |
| 7.0 - 8.9 | HIGH | Orange | Orange |
| 9.0 - 10.0 | CRITICAL | Red | Red |

## Implementation Details

### Calculator Engine

**Pure Functions Strategy**:
```typescript
// lib/cvss/v3-calculator.ts
export function calculateV3Score(metrics: V3Metrics): CVSSScore {
  // 1. Calculate Impact Subscore (ISS)
  const ISS = calculateImpact(metrics);

  // 2. Calculate Impact
  const I = calculateI(ISS, metrics.scope);

  // 3. Calculate Exploitability
  const Exp = calculateExploitability(metrics);

  // 4. Calculate Base Score
  let baseScore;
  if (metrics.scope === 'U') {
    baseScore = Math.min(10, I + Exp);
  } else {
    baseScore = Math.min(10, 1.08 * (I + Exp));
  }

  // 5. Round to 1 decimal place
  baseScore = Math.round(baseScore * 10) / 10;

  // 6. Determine severity
  const severity = determineSeverity(baseScore);

  return { baseScore, severity, vectorString: formatVector(metrics) };
}
```

### URL Sharing

Parse vector from URL query parameter:
```typescript
// app/cvss/[version]/page.tsx
export default function CVSSCalculatorPage({ searchParams }: PageProps) {
  const vectorParam = searchParams.vector;
  const initialMetrics = useMemo(() =>
    vectorParam ? parseVector(vectorParam, version) : defaultMetrics
  , [vectorParam, version]);

  return <CVSSCalculator initialMetrics={initialMetrics} />;
}
```

### Export Functionality

```typescript
// lib/cvss/utils.ts
export function exportToJSON(score: CVSSScore): string {
  return JSON.stringify(score, null, 2);
}

export function exportToCSV(scores: CVSSScore[]): string {
  const headers = ['version', 'baseScore', 'severity', 'vectorString'];
  const rows = scores.map(s => [s.version, s.baseScore, s.severity, s.vectorString]);
  return [headers, ...rows].map(row => row.join(',')).join('\n');
}

export function generateShareURL(version: string, vector: string): string {
  const baseUrl = typeof window !== 'undefined'
    ? window.location.origin
    : 'http://localhost:3000';
  return `${baseUrl}/cvss/${version}?vector=${encodeURIComponent(vector)}`;
}
```

## Testing Strategy

### Unit Tests

**Calculator Logic Tests** (`lib/cvss/__tests__/`):
- Test each CVSS version with known vectors from FIRST.org
- Verify base, temporal, and environmental score calculations
- Edge cases: all "NONE", all "HIGH", undefined metrics
- Vector string parsing and formatting

**Test Vectors** (examples from FIRST.org):
```
CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H → Base: 9.8 (CRITICAL)
CVSS:3.1/AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H → Base: 8.8 (HIGH)
CVSS:3.1/AV:N/AC:H/PR:H/UI:N/S:U/C:L/I:N/A:N → Base: 3.1 (LOW)
```

### Integration Tests

**Component Testing**:
- Render each version's calculator
- Simulate user metric selections
- Verify vector string updates correctly
- Test export functionality (JSON, CSV, URL)
- View toggle between v2e and traditional modes

### Validation

**Cross-Reference with FIRST.org**:
- Test vectors must match official FIRST.org calculator results exactly
- Use sample vectors from FIRST CVSS specification documents

## Phased Implementation

### Phase 1: Foundation
- [ ] Set up routing structure (`/cvss`, `/cvss/[version]`)
- [ ] Create type definitions in `lib/types.ts`
- [ ] Build metric definitions in `lib/cvss/metrics.ts`

### Phase 2: CVSS v3.1
- [ ] Implement v3.1 calculator engine
- [ ] Build v3.1 UI components
- [ ] Add unit tests for v3.1
- [ ] Validate against FIRST.org

### Phase 3: CVSS v3.0
- [ ] Adapt v3.1 for v3.0 (minor differences)
- [ ] Add v3.0 specific tests
- [ ] Update routing and UI

### Phase 4: CVSS v4.0
- [ ] Implement v4.0 calculator engine (new formula)
- [ ] Build v4.0 specific UI components
- [ ] Add comprehensive v4.0 tests

### Phase 5: UI Enhancements
- [ ] Implement v2e/Traditional view toggle
- [ ] Add export functionality (JSON, CSV, URL)
- [ ] Create recent calculations history
- [ ] Add accessibility features (ARIA labels, keyboard navigation)

### Phase 6: Integration
- [ ] Connect to CVE workflow (attach scores to CVE records)
- [ ] Add bulk calculation mode
- [ ] Implement calculation import/export

## Dependencies

### New npm Packages
None required - pure TypeScript/JavaScript implementation

### Existing Dependencies
- `react`, `next`: Core framework
- `shadcn/ui`: UI components (Card, Select, Tabs, Button)
- `lucide-react`: Icons
- `clsx`, `tailwind-merge`: Styling utilities

## References

- [FIRST CVSS v3.1 Specification](https://www.first.org/cvss/calculator/3.1)
- [FIRST CVSS v3.0 Specification](https://www.first.org/cvss/calculator/3.0)
- [FIRST CVSS v4.0 Specification](https://www.first.org/cvss/calculator/4.0)
- [CVSS Scoring System Documentation](https://www.first.org/cvss/specification-document)

## Notes

- All calculations are client-side; no backend RPC calls required
- Calculator results should be deterministic and reproducible
- Vector strings are the canonical representation - scores are derived
- CVSS v4.0 introduces new metric categories not present in v3.x
