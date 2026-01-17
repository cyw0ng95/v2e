# v2e Website

This is the frontend website for the v2e CVE (Common Vulnerabilities and Exposures) Management System, built with Next.js 15+ and designed for static site generation (SSG).

## Architecture

- **Framework**: Next.js 15+ with App Router
- **Output Strategy**: Static Site Generation (SSG) via `output: 'export'`
- **Styling**: Tailwind CSS + shadcn/ui (Radix UI components)
- **Icons**: Lucide React
- **Data Fetching**: TanStack Query (React Query) v5
- **Form Validation**: react-hook-form + zod
- **Notifications**: Sonner (toast notifications)

## Development

### Prerequisites

- Node.js 20.x or higher
- npm 10.x or higher

### Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Configure environment (optional):
   ```bash
   cp .env.local.example .env.local
   ```
   
   Edit `.env.local` to configure:
   - `NEXT_PUBLIC_USE_MOCK_DATA`: Set to `true` for mock mode (frontend development without backend)
   - `NEXT_PUBLIC_API_BASE_URL`: API endpoint (default: http://localhost:8080)

3. Run development server:
   ```bash
   npm run dev
   ```
   
   Open [http://localhost:3000](http://localhost:3000) in your browser.

### Mock Mode

Mock mode allows frontend development without the Go backend running:

- Set `NEXT_PUBLIC_USE_MOCK_DATA=true` in `.env.local`
- The RPC client will return simulated data with realistic delays
- Perfect for UI development and testing

### Build

Build for production:

```bash
npm run build
```

This generates a static export in the `out/` directory containing only HTML, CSS, JS, and assets.

### Lint

```bash
npm run lint
```

## Integration with Go Backend

The website is designed to be served by the Go `access` service:

1. Build the website:
   ```bash
   npm run build
   ```

2. The `out/` directory can be copied to `.build/package/` for the access service to serve as static assets

3. All assets use relative paths to ensure they load correctly when served from a Go sub-route

## RPC Client Architecture

The RPC client implements a Service-Consumer pattern:

- **Client Factory**: `lib/rpc-client.ts` handles HTTP requests to the access service
- **Type Mirroring**: TypeScript interfaces in `lib/types.ts` mirror Go structs from the backend
- **Case Conversion**: Automatic conversion between Go's PascalCase/snake_case and TypeScript's camelCase
- **Mock Mode**: Built-in mock data support for development without backend

### Example RPC Call

```typescript
import { rpcClient } from '@/lib/rpc-client';

// Get CVE data
const response = await rpcClient.getCVE('CVE-2021-44228');
if (response.retcode === 0) {
  console.log(response.payload.cve);
}
```

### React Query Hooks

Use pre-built hooks for data fetching:

```typescript
import { useCVE, useCVEList, useSessionStatus } from '@/lib/hooks';

function MyComponent() {
  const { data: cveData, isLoading } = useCVE('CVE-2021-44228');
  const { data: cveList } = useCVEList(0, 10);
  const { data: sessionStatus } = useSessionStatus();
  
  // ...
}
```

## UI Components

### shadcn/ui Components Used

- `button`, `card`, `input`, `table`, `form`, `label`, `select`
- `dialog`, `sonner`, `badge`, `separator`, `dropdown-menu`
- `sidebar`, `skeleton`, `sheet`, `tooltip`

### Custom Components

- `CVETable`: Displays CVE data with pagination and filtering
- `SessionControl`: Manages job session control (start, stop, pause, resume)

## Pages

- `/` - Dashboard with CVE table and session management

## API Endpoints

The frontend communicates with the Go backend via:

- `GET /restful/health` - Health check
- `POST /restful/rpc` - Generic RPC endpoint for all backend operations

Example RPC request:
```json
{
  "method": "RPCGetCVE",
  "target": "cve-meta",
  "params": {
    "cve_id": "CVE-2021-44228"
  }
}
```

Example RPC response:
```json
{
  "retcode": 0,
  "message": "success",
  "payload": {
    "cve": { ... },
    "source": "local"
  }
}
```

## Static Export Configuration

The website is configured for static export in `next.config.ts`:

```typescript
{
  output: 'export',
  images: { unoptimized: true },
  basePath: '',
  trailingSlash: true
}
```

This ensures:
- No Node.js runtime features (no SSR, no API routes)
- All assets use relative paths
- Compatible with static hosting or Go HTTP server

## Directory Structure

```
website/
├── app/                    # Next.js app directory
│   ├── layout.tsx         # Root layout
│   └── page.tsx           # Dashboard page
├── components/            # React components
│   ├── ui/               # shadcn/ui components
│   ├── cve-table.tsx     # CVE table component
│   └── session-control.tsx # Session control component
├── lib/                   # Library code
│   ├── hooks.ts          # React Query hooks
│   ├── providers.tsx     # React providers
│   ├── rpc-client.ts     # RPC client
│   ├── types.ts          # TypeScript types
│   └── utils.ts          # Utility functions
├── public/               # Static assets
├── .env.local.example    # Environment variables example
├── next.config.ts        # Next.js configuration
├── package.json          # Dependencies
└── tsconfig.json         # TypeScript configuration
```

## Troubleshooting

### Build Errors

If you encounter build errors:

1. Clean the build cache:
   ```bash
   rm -rf .next out
   npm run build
   ```

2. Verify TypeScript types:
   ```bash
   npx tsc --noEmit
   ```

### API Connection Issues

If the frontend cannot connect to the backend:

1. Verify the backend is running:
   ```bash
   curl http://localhost:8080/restful/health
   ```

2. Check `NEXT_PUBLIC_API_BASE_URL` in `.env.local`

3. Enable mock mode for development:
   ```bash
   NEXT_PUBLIC_USE_MOCK_DATA=true
   ```
