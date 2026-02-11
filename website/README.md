# v2e Website

Next.js 15+ frontend for the v2e (Vulnerabilities Viewer Engine) project using App Router and Static Site Generation.

## CVSS Calculator

The CVSS Calculator is an interactive tool for calculating Common Vulnerability Scoring System (CVSS) scores following FIRST.org specifications.

### Supported Versions

- **CVSS v4.0** - Latest version with enhanced granularity and threat modeling
- **CVSS v3.1** - Refined v3.0 with additional environmental metrics
- **CVSS v3.0** - Current standard with temporal and environmental metrics

### Quick Start

Access the CVSS Calculator at `/cvss` or navigate directly to:
- `/cvss/4.0` - CVSS v4.0 Calculator
- `/cvss/3.1` - CVSS v3.1 Calculator
- `/cvss/3.0` - CVSS v3.0 Calculator

### Features

- **Real-time Calculation**: Instant score updates as you adjust metrics
- **Vector Generation**: Auto-generate CVSS vector strings for vulnerability reporting
- **Export Options**: JSON, CSV, and URL sharing capabilities
- **Version Switching**: Seamless switching between CVSS versions
- **Interactive Help**: Tooltips and guides for each metric

### Score Types

#### Base Score (0-10)
The core severity based on exploitability and impact metrics. Represents the intrinsic characteristics of a vulnerability.

#### Temporal Score (v3.0/v3.1)
Adjustments for exploit maturity, remediation availability, and report confidence. Changes over time as the vulnerability landscape evolves.

#### Environmental Score (v3.0/v3.1)
Customizes the score for your specific security environment and requirements. Takes into account modified base metrics and confidentiality/integrity/availability requirements.

#### Threat Score (v4.0)
Reflects the current threat landscape including exploit activity and threat intelligence.

### Severity Ratings

| Score Range | Severity |
|-------------|----------|
| 0.0 | None |
| 0.1 - 3.9 | Low |
| 4.0 - 6.9 | Medium |
| 7.0 - 8.9 | High |
| 9.0 - 10.0 | Critical |

### External Resources

- [CVSS v4.0 Specification](https://www.first.org/cvss/calculator/4.0)
- [CVSS v3.1 Specification](https://www.first.org/cvss/calculator/3.1)
- [CVSS v3.0 Specification](https://www.first.org/cvss/calculator/3.0)
- [CVSS Calculator Information](https://www.first.org/cvss/calculator-information)
- [FIRST.org](https://www.first.org/)

## Tech Stack

- **Framework**: Next.js 15+ with App Router
- **Styling**: Tailwind CSS v4 + shadcn/ui (Radix UI components)
- **Data Fetching**: TanStack Query v5
- **Icons**: Lucide React
- **Build**: Static Site Generation (`output: 'export'`)

## Development

```bash
# Install dependencies
npm install

# Development server
npm run dev

# Production build
npm run build

# Run linter
npm run lint
```

## Build Output

The production build exports static files to the `out/` directory:
- All routes are pre-rendered as static HTML
- Paths must be relative (no absolute paths)
- Compatible with Go sub-route serving from v2e backend

## Project Structure

```
app/
  cvss/
    page.tsx           # CVSS version selector landing page
    [version]/
      page.tsx         # Dynamic route for CVSS calculators
components/
  cvss/
    calculator.tsx      # Main calculator component
    metric-tooltip.tsx # Help tooltips for metrics
    user-guide.tsx     # Interactive user guide
lib/
  cvss-calculator.ts  # CVSS scoring formulas
  cvss-context.tsx    # React context for calculator state
  types.ts           # TypeScript type definitions
```

## License

MIT
