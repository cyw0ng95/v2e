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
|  |  Attack Complexity: [L ▼]     |  |  (Animated Progress)         |   |
|  |  ...                          |  |                             |   |
|  |                               |  +-----------------------------+   |
|  +-------------------------------+  Vector String:                |
|                                   CVSS:3.1/AV:N/AC:L/PR:N/...   |
|                                                                   |
|  +-------------------------------+  +-----------------------------+   |
|  |  Temporal Metrics (Optional)  |  [Reset] [Copy] [Export]     |   |
|  |  (Collapsible)                 |  (with tooltip hints)         |   |
|  |  ...                          |                                   |
|  +-------------------------------+                                   |
+----------------------------------------------------------------+
```

### Modern Component Design

#### 1. Animated Score Display

```typescript
// components/cvss/score-display.tsx
import { motion } from 'framer-motion';

export function ScoreDisplay({ score, severity }: { score: number, severity: string }) {
  return (
    <motion.div
      initial={{ scale: 0.9, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      transition={{
        type: 'spring',
        stiffness: 300,
        damping: 20
      }}
      className="relative"
    >
      <motion.div
        className="text-6xl font-bold"
        animate={{
          color: severityColors[severity]
        }}
        transition={{ duration: 0.3 }}
      >
        {score.toFixed(1)}
      </motion.div>
      <motion.span
        className="text-lg font-medium"
        initial={{ y: 10, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ delay: 0.2 }}
      >
        {severity}
      </motion.span>
    </motion.div>
  );
}
```

#### 2. Interactive Metric Selector

```typescript
// components/cvss/metric-selector.tsx
export function MetricSelector({
  label,
  value,
  options,
  onChange
}: MetricSelectorProps) {
  return (
    <motion.div
      className="space-y-2"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <label className="text-sm font-medium text-gray-700">{label}</label>
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger className="w-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <motion.div
              key={option.value}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
            >
              <SelectItem value={option.value}>
                {option.label}
              </SelectItem>
            </motion.div>
          ))}
        </SelectContent>
      </Select>
    </motion.div>
  );
}
```

#### 3. Severity Gauge Component

```typescript
// components/cvss/severity-gauge.tsx
export function SeverityGauge({ score }: { score: number }) {
  const percentage = (score / 10) * 100;
  const severity = getSeverity(score);

  return (
    <div className="relative w-full h-4 bg-gray-200 rounded-full overflow-hidden">
      <motion.div
        className="absolute h-full rounded-full"
        style={{
          width: `${percentage}%`,
          backgroundColor: severityColors[severity]
        }}
        initial={{ width: 0 }}
        animate={{ width: `${percentage}%` }}
        transition={{ duration: 0.8, ease: 'easeOut' }}
      />
      <motion.div
        className="absolute top-0 right-0 h-full w-0.5 bg-white"
        style={{ left: `${percentage}%` }}
        animate={{ opacity: [0, 1, 0] }}
        transition={{ duration: 1.5, repeat: Infinity }}
      />
    </div>
  );
}
```

#### 4. Vector String with Copy Animation

```typescript
// components/cvss/vector-string.tsx
export function VectorString({ vector }: { vector: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(vector);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <motion.div
      className="flex items-center gap-2"
      whileHover={{ scale: 1.01 }}
    >
      <code className="flex-1 bg-gray-100 px-3 py-2 rounded text-sm font-mono">
        {vector}
      </code>
      <Button
        variant="outline"
        size="sm"
        onClick={handleCopy}
        className="relative"
      >
        {copied ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.5 }}
            animate={{ opacity: 1, scale: 1 }}
            className="absolute inset-0 flex items-center justify-center bg-green-500 text-white rounded"
          >
            <Check className="w-4 h-4" />
          </motion.div>
        ) : (
          <Copy className="w-4 h-4" />
        )}
      </Button>
    </motion.div>
  );
}
```

### Transition Effects

#### 1. Page Transitions

```typescript
// app/cvss/layout.tsx
import { motion } from 'framer-motion';

export default function CVSSLayout({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.3 }}
      className="min-h-screen"
    >
      {children}
    </motion.div>
  );
}
```

#### 2. Metric Group Animations

```typescript
// components/cvss/metric-group.tsx
export function MetricGroup({ title, children, collapsed }: MetricGroupProps) {
  return (
    <motion.div
      className="border rounded-lg overflow-hidden"
      initial={{ opacity: 0, height: 0 }}
      animate={{
        opacity: 1,
        height: collapsed ? 0 : 'auto'
      }}
      transition={{
        duration: 0.4,
        ease: [0.04, 0.62, 0.23, 0.98]
      }}
    >
      <motion.div
        className="px-4 py-3 bg-gray-50 border-b"
        whileHover={{ backgroundColor: '#f9fafb' }}
      >
        <h3 className="font-semibold text-gray-900">{title}</h3>
      </motion.div>
      <motion.div
        className="p-4 space-y-4"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.1 }}
      >
        {children}
      </motion.div>
    </motion.div>
  );
}
```

#### 3. Score Update Animations

```typescript
// components/cvss/score-update.tsx
export function ScoreUpdate({ prevScore, newScore }: { prevScore: number, newScore: number }) {
  const diff = newScore - prevScore;

  return (
    <motion.div
      className="text-sm"
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      exit={{ y: 20, opacity: 0 }}
    >
      {diff !== 0 && (
        <motion.span
          className={`font-medium ${diff > 0 ? 'text-red-600' : 'text-green-600'}`}
          animate={{
            scale: [1, 1.2, 1]
          }}
          transition={{ duration: 0.3 }}
        >
          {diff > 0 ? '+' : ''}{diff.toFixed(1)}
        </motion.span>
      )}
    </motion.div>
  );
}
```

#### 4. Hover Effects for Metrics

```typescript
// components/cvss/metric-card.tsx
export function MetricCard({ label, value, description }: MetricCardProps) {
  return (
    <motion.div
      className="p-4 bg-white rounded-lg border shadow-sm"
      whileHover={{
        y: -2,
        boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.1)'
      }}
      transition={{
        type: 'spring',
        stiffness: 400,
        damping: 17
      }}
    >
      <div className="flex items-start justify-between">
        <div>
          <h4 className="font-medium text-gray-900">{label}</h4>
          <p className="text-sm text-gray-600 mt-1">{description}</p>
        </div>
        <motion.div
          className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium"
          whileHover={{ scale: 1.1 }}
        >
          {value}
        </motion.div>
      </div>
    </motion.div>
  );
}
```

### Responsive Design

#### Mobile Layout

```typescript
// components/cvss/mobile-view.tsx
export function MobileView({ metrics, score }: { metrics: any, score: CVSSScore }) {
  return (
    <div className="flex flex-col gap-4">
      {/* Score Display - Sticky Header */}
      <div className="sticky top-0 z-10 bg-white shadow-md">
        <ScoreDisplay score={score.baseScore} severity={score.severity} />
      </div>

      {/* Metric Groups - Accordion Style */}
      <Accordion type="single" collapsible>
        <AccordionItem value="base">
          <AccordionTrigger>
            <span>Base Metrics</span>
            <Badge>Required</Badge>
          </AccordionTrigger>
          <AccordionContent>
            <BaseMetrics metrics={metrics.base} />
          </AccordionContent>
        </AccordionItem>

        <AccordionItem value="temporal">
          <AccordionTrigger>
            <span>Temporal Metrics</span>
            <Badge variant="outline">Optional</Badge>
          </AccordionTrigger>
          <AccordionContent>
            <TemporalMetrics metrics={metrics.temporal} />
          </AccordionContent>
        </AccordionItem>
      </Accordion>

      {/* Vector String - Bottom Sheet */}
      <motion.div
        className="fixed bottom-0 left-0 right-0 bg-white border-t shadow-lg"
        initial={{ y: '100%' }}
        animate={{ y: 0 }}
      >
        <VectorString vector={score.vectorString} />
      </motion.div>
    </div>
  );
}
```

#### Desktop Layout

```typescript
// components/cvss/desktop-view.tsx
export function DesktopView({ metrics, score }: { metrics: any, score: CVSSScore }) {
  return (
    <div className="grid grid-cols-3 gap-6">
      {/* Left Column - Metric Groups */}
      <div className="col-span-2 space-y-6">
        <MetricGroup title="Base Metrics">
          <BaseMetrics metrics={metrics.base} />
        </MetricGroup>
        <MetricGroup title="Temporal Metrics">
          <TemporalMetrics metrics={metrics.temporal} />
        </MetricGroup>
      </div>

      {/* Right Column - Score Display */}
      <motion.div
        className="col-span-1"
        initial={{ x: 50, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        transition={{ delay: 0.2 }}
      >
        <Card className="sticky top-4">
          <CardHeader>
            <ScoreDisplay score={score.baseScore} severity={score.severity} />
          </CardHeader>
          <CardContent>
            <VectorString vector={score.vectorString} />
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
```

### Accessibility Enhancements

#### 1. Keyboard Navigation

```typescript
// components/cvss/keyboard-accessible.tsx
export function KeyboardAccessibleMetricSelector({
  label,
  value,
  options,
  onChange
}: MetricSelectorProps) {
  return (
    <div className="space-y-2">
      <label
        htmlFor={`metric-${label}`}
        className="text-sm font-medium text-gray-700"
      >
        {label}
      </label>
      <Select
        value={value}
        onValueChange={onChange}
      >
        <SelectTrigger
          id={`metric-${label}`}
          className="w-full focus:ring-2 focus:ring-blue-500"
        >
          <SelectValue aria-label={`Selected ${label}: ${value}`} />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem
              key={option.value}
              value={option.value}
              aria-label={`${option.label} (${option.value})`}
            >
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
```

#### 2. Screen Reader Support

```typescript
// components/cvss/screen-reader.tsx
export function ScoreAnnouncement({ score, severity }: { score: number, severity: string }) {
  const announcement = `${severity} severity with a base score of ${score.toFixed(1)} out of 10`;

  return (
    <div className="sr-only" role="status" aria-live="polite">
      {announcement}
    </div>
  );
}
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

### Additional Dependencies

#### Animation Libraries
```json
{
  "framer-motion": "^11.0.0"
}
```

#### Additional UI Components
```json
{
  "@radix-ui/react-accordion": "^1.1.2",
  "@radix-ui/react-tooltip": "^1.0.7",
  "@radix-ui/react-slot": "^1.0.2"
}
```

### Configuration Updates

#### Tailwind Animation Extensions

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      animation: {
        'fade-in': 'fadeIn 0.3s ease-in-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.9)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
    },
  },
}
```

### Performance Optimizations

#### 1. Memoization for Expensive Calculations

```typescript
// lib/cvss/calculator.ts
import { useMemo } from 'react';

export function useCVSSCalculator(metrics: CVSSMetrics) {
  const score = useMemo(() => {
    return calculateCVSSScore(metrics);
  }, [metrics]);

  return score;
}
```

#### 2. Debounced Metric Updates

```typescript
// lib/cvss/debounce.ts
import { useCallback, useEffect, useRef } from 'react';

export function useDebouncedCallback<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const timeoutRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return useCallback((...args: Parameters<T>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      callback(...args);
    }, delay);
  }, [callback, delay]) as T;
}
```

#### 3. Lazy Loading Components

```typescript
// components/cvss/calculator.tsx
import dynamic from 'next/dynamic';

const ScoreDisplay = dynamic(() => import('./score-display'), {
  loading: () => <div className="animate-pulse">Loading score...</div>,
});

const VectorString = dynamic(() => import('./vector-string'), {
  loading: () => <div className="animate-pulse">Loading vector...</div>,
});
```

### Testing Strategy (Enhanced)

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

#### 1. Component Animation Tests

```typescript
// components/cvss/__tests__/score-display.test.tsx
import { render, screen } from '@testing-library/react';
import ScoreDisplay from '../score-display';

describe('ScoreDisplay', () => {
  it('renders score with correct severity color', () => {
    render(<ScoreDisplay score={9.8} severity="CRITICAL" />);
    expect(screen.getByText('9.8')).toBeInTheDocument();
  });

  it('applies animation classes', () => {
    const { container } = render(<ScoreDisplay score={9.8} severity="CRITICAL" />);
    expect(container.firstChild).toHaveClass('motion-component');
  });

  it('updates on score change', async () => {
    const { rerender } = render(<ScoreDisplay score={5.0} severity="MEDIUM" />);
    rerender(<ScoreDisplay score={7.5} severity="HIGH" />);
    expect(screen.getByText('7.5')).toBeInTheDocument();
  });
});
```

#### 2. Transition Effect Tests

```typescript
// components/cvss/__tests__/metric-group.test.tsx
import { render, fireEvent, waitFor } from '@testing-library/react';
import MetricGroup from '../metric-group';

describe('MetricGroup', () => {
  it('renders collapsed initially', () => {
    render(<MetricGroup title="Base Metrics" collapsed={true} />);
    expect(screen.queryByText('Attack Vector')).not.toBeInTheDocument();
  });

  it('expands on click', async () => {
    render(<MetricGroup title="Base Metrics" collapsed={true} />);
    const trigger = screen.getByText('Base Metrics');
    fireEvent.click(trigger);

    await waitFor(() => {
      expect(screen.getByText('Attack Vector')).toBeInTheDocument();
    });
  });
});
```

#### 3. Performance Tests

```typescript
// __tests__/performance/calculator-performance.test.ts
import { calculateV3Score } from '../../lib/cvss/v3-calculator';

describe('Calculator Performance', () => {
  it('calculates 1000 scores in under 100ms', () => {
    const metrics = {
      attackVector: 'N',
      attackComplexity: 'L',
      privilegesRequired: 'N',
      userInteraction: 'N',
      scope: 'U',
      confidentiality: 'H',
      integrity: 'H',
      availability: 'H'
    };

    const start = performance.now();
    for (let i = 0; i < 1000; i++) {
      calculateV3Score(metrics);
    }
    const duration = performance.now() - start;

    expect(duration).toBeLessThan(100);
  });
});
```

### Additional UI Components

#### 1. Export Modal with Animations

```typescript
// components/cvss/export-modal.tsx
export function ExportModal({ score, onClose }: ExportModalProps) {
  const [format, setFormat] = useState<'json' | 'csv' | 'url'>('json');

  return (
    <Dialog>
      <DialogContent className="sm:max-w-md">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.2 }}
        >
          <DialogHeader>
            <DialogTitle>Export CVSS Score</DialogTitle>
            <DialogDescription>
              Choose a format to export the CVSS calculation
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <Tabs value={format} onValueChange={(v: any) => setFormat(v)}>
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="json">JSON</TabsTrigger>
                <TabsTrigger value="csv">CSV</TabsTrigger>
                <TabsTrigger value="url">URL</TabsTrigger>
              </TabsList>

              <TabsContent value="json">
                <CodeBlock code={exportToJSON(score)} language="json" />
              </TabsContent>
              <TabsContent value="csv">
                <CodeBlock code={exportToCSV([score])} language="csv" />
              </TabsContent>
              <TabsContent value="url">
                <Input
                  value={generateShareURL(score.version, score.vectorString)}
                  readOnly
                  className="font-mono"
                />
              </TabsContent>
            </Tabs>
          </div>

          <DialogFooter>
            <Button onClick={onClose}>Close</Button>
          </DialogFooter>
        </motion.div>
      </DialogContent>
    </Dialog>
  );
}
```

#### 2. Tooltip Enhancement

```typescript
// components/cvss/metric-tooltip.tsx
export function MetricTooltip({ description, formula }: MetricTooltipProps) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Info className="w-4 h-4 text-gray-400 cursor-help" />
        </TooltipTrigger>
        <TooltipContent>
          <motion.div
            initial={{ opacity: 0, y: 5 }}
            animate={{ opacity: 1, y: 0 }}
            className="max-w-xs"
          >
            <p className="text-sm">{description}</p>
            {formula && (
              <p className="text-xs text-gray-600 mt-2 font-mono">
                {formula}
              </p>
            )}
          </motion.div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
```

#### 3. Progress Animation for Calculations

```typescript
// components/cvss/calculation-progress.tsx
export function CalculationProgress({
  currentStep,
  totalSteps
}: CalculationProgressProps) {
  const progress = (currentStep / totalSteps) * 100;

  return (
    <motion.div
      className="w-full h-2 bg-gray-200 rounded-full overflow-hidden"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
    >
      <motion.div
        className="h-full bg-blue-500 rounded-full"
        initial={{ width: 0 }}
        animate={{ width: `${progress}%` }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
      >
        <motion.div
          className="h-full w-full"
          animate={{
            opacity: [0.5, 1, 0.5]
          }}
          transition={{
            duration: 1,
            repeat: Infinity,
            ease: 'easeInOut'
          }}
        />
      </motion.div>
    </motion.div>
  );
}
```

### Documentation Structure

#### Component Documentation Template

```typescript
/**
 * ScoreDisplay Component
 *
 * @description
 * Displays the CVSS score with animated updates and severity indication.
 * Supports color-coded severity levels based on CVSS scoring ranges.
 *
 * @features
 * - Animated score transitions using framer-motion
 * - Severity color mapping (NONE, LOW, MEDIUM, HIGH, CRITICAL)
 * - Real-time score updates with spring animations
 * - Screen reader support with aria-live announcements
 *
 * @example
 * ```tsx
 * <ScoreDisplay
 *   score={9.8}
 *   severity="CRITICAL"
 *   onUpdate={(newScore) => console.log(newScore)}
 * />
 * ```
 *
 * @accessibility
 * - Role: status
 * - Live region: polite
 * - Announces score changes to screen readers
 *
 * @performance
 * - Memoized calculations to prevent unnecessary re-renders
 * - CSS animations over JavaScript for better performance
 */
```

### Phased Implementation (Updated)

### Phase 1: Foundation
- [ ] Set up routing structure (`/cvss`, `/cvss/[version]`)
- [ ] Create type definitions in `lib/types.ts`
- [ ] Build metric definitions in `lib/cvss/metrics.ts`
- [ ] Install framer-motion and configure Tailwind animations
- [ ] Set up component documentation templates

### Phase 2: CVSS v3.1
- [ ] Implement v3.1 calculator engine
- [ ] Build v3.1 UI components with basic animations
- [ ] Add ScoreDisplay component with spring animations
- [ ] Add MetricSelector with hover effects
- [ ] Add unit tests for v3.1
- [ ] Validate against FIRST.org
- [ ] Add keyboard navigation support

### Phase 3: CVSS v3.0
- [ ] Adapt v3.1 for v3.0 (minor differences)
- [ ] Add v3.0 specific tests
- [ ] Update routing and UI
- [ ] Test cross-version consistency

### Phase 4: CVSS v4.0
- [ ] Implement v4.0 calculator engine (new formula)
- [ ] Build v4.0 specific UI components
- [ ] Add v4.0 specific animations (new metric categories)
- [ ] Add comprehensive v4.0 tests
- [ ] Validate against FIRST.org v4.0 calculator

### Phase 5: UI Enhancements
- [ ] Implement v2e/Traditional view toggle with page transitions
- [ ] Add SeverityGauge component with animated progress
- [ ] Add VectorString component with copy animation
- [ ] Add export functionality (JSON, CSV, URL) with modal animations
- [ ] Create recent calculations history with slide transitions
- [ ] Add accessibility features (ARIA labels, keyboard navigation)
- [ ] Implement responsive mobile layout with accordion components

### Phase 6: Performance & Optimization
- [ ] Add memoization for expensive calculations
- [ ] Implement debounced metric updates
- [ ] Add lazy loading for components
- [ ] Optimize animation performance (use CSS transforms)
- [ ] Add performance tests and benchmarks

### Phase 7: Integration
- [ ] Connect to CVE workflow (attach scores to CVE records)
- [ ] Add bulk calculation mode
- [ ] Implement calculation import/export
- [ ] Add real-time collaboration features (optional)

### Phase 8: Documentation & Polish
- [ ] Complete component documentation
- [ ] Add code examples and usage guides
- [ ] Create user guide for CVSS calculator
- [ ] Add interactive tutorials for CVSS concepts
- [ ] Polish animations and transitions
- [ ] Final accessibility audit

### Phase 9: Testing & Validation
- [ ] Run comprehensive unit tests
- [ ] Perform integration testing with CVE workflow
- [ ] Cross-platform browser testing
- [ ] Performance testing and optimization
- [ ] Accessibility testing (WCAG 2.1 AA compliance)
- [ ] User acceptance testing (UAT)

### New npm Packages
```json
{
  "framer-motion": "^11.0.0"
}
```

### Existing Dependencies
```json
{
  "react": "^18.2.0",
  "next": "^15.0.0",
  "@radix-ui/react-accordion": "^1.1.2",
  "@radix-ui/react-tooltip": "^1.0.7",
  "@radix-ui/react-slot": "^1.0.2",
  "@radix-ui/react-dialog": "^1.0.5",
  "@radix-ui/react-tabs": "^1.0.4",
  "shadcn/ui": "^1.0.0",
  "lucide-react": "^0.344.0",
  "clsx": "^2.1.0",
  "tailwind-merge": "^2.2.0"
}
```

## Animation Guidelines

### Performance Best Practices

1. **Use CSS transforms instead of position changes**
   ```typescript
   // Good
   <motion.div animate={{ x: 100 }} />

   // Bad (causes layout thrashing)
   <motion.div animate={{ left: '100px' }} />
   ```

2. **Prefer GPU-accelerated properties**
   ```typescript
   // GPU-accelerated (fast)
   <motion.div animate={{ scale: 1.1, rotate: 5, opacity: 0.8 }} />

   // CPU-bound (slower)
   <motion.div animate={{ width: 300, height: 200 }} />
   ```

3. **Use `layout` prop for layout animations**
   ```typescript
   <motion.div layout transition={{ duration: 0.3 }}>
     {children}
   </motion.div>
   ```

4. **Avoid nested animations**
   ```typescript
   // Prefer flattened animations
   <motion.div animate={{ opacity: 1, scale: 1 }} />
   ```

### Transition Durations

```typescript
// Constants for consistent animations
export const ANIMATION_DURATIONS = {
  fast: 0.15,      // Button clicks, hovers
  normal: 0.3,     // Modal open/close, page transitions
  slow: 0.5,       // Complex layouts, form inputs
  verySlow: 0.8    // Score updates, gauge animations
} as const;
```

### Easing Functions

```typescript
// Recommended easing functions
export const EASING_FUNCTIONS = {
  spring: {
    type: 'spring',
    stiffness: 300,
    damping: 20
  },
  easeOut: {
    type: 'easeOut',
    duration: ANIMATION_DURATIONS.normal
  },
  bounce: {
    type: 'spring',
    stiffness: 400,
    damping: 10
  }
} as const;
```

## Accessibility Standards

### WCAG 2.1 AA Compliance

1. **Color Contrast**
   - Minimum 4.5:1 for normal text
   - Minimum 3:1 for large text
   - Severity colors must pass WCAG contrast checks

2. **Keyboard Navigation**
   - All interactive elements must be keyboard accessible
   - Tab order must be logical
   - Focus indicators must be visible (2px minimum)

3. **Screen Reader Support**
   - Use `aria-live` for dynamic content updates
   - Provide `aria-label` for form controls
   - Use semantic HTML elements

4. **Motion Preferences**
   - Respect `prefers-reduced-motion` media query
   ```typescript
   const prefersReducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)');

   <motion.div
     animate={prefersReducedMotion ? false : { opacity: 1 }}
   />
   ```

## References

- [FIRST CVSS v3.1 Specification](https://www.first.org/cvss/calculator/3.1)
- [FIRST CVSS v3.0 Specification](https://www.first.org/cvss/calculator/3.0)
- [FIRST CVSS v4.0 Specification](https://www.first.org/cvss/calculator/4.0)
- [CVSS Scoring System Documentation](https://www.first.org/cvss/specification-document)

## Notes

### Technical Notes
- All calculations are client-side; no backend RPC calls required
- Calculator results should be deterministic and reproducible
- Vector strings are the canonical representation - scores are derived
- CVSS v4.0 introduces new metric categories not present in v3.x

### Animation Notes
- Use framer-motion for all animations (consistent API)
- Prioritize CSS transforms over JavaScript for performance
- Test animations with `prefers-reduced-motion` enabled
- Ensure animations don't interfere with keyboard navigation
- Provide immediate visual feedback for all user interactions

### Design Notes
- Follow v2e design system (shadcn/ui components)
- Maintain consistency with existing UI patterns
- Ensure mobile-first responsive design
- Use semantic HTML for accessibility
- Provide clear visual hierarchy and information architecture

### Performance Notes
- Memoize expensive calculations to prevent re-renders
- Use debouncing for rapid metric changes
- Lazy load components where appropriate
- Optimize bundle size (tree-shake unused animations)
- Profile animation performance in development

### Testing Notes
- Test animations with different browser capabilities
- Verify accessibility with screen readers (NVDA, JAWS, VoiceOver)
- Test on various screen sizes and orientations
- Validate against FIRST.org calculators for accuracy
- Include regression tests for animation behavior

### Future Enhancements
- Support for custom metric weights
- Historical score comparison charts
- Bulk import/export of CVSS vectors
- Integration with vulnerability databases
- Machine learning-based severity prediction
- Real-time collaboration features
- CVSS score trend analysis
- Export to PDF with formatted reports
