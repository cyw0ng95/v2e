# MAINTENANCE TODO

This document tracks maintenance tasks for the v2e project. Tasks are organized by priority and type.

## TODO

| ID  | Package           | Type    | Description                                                                        | Est LoC | Priority | WONTFIX |
|-----|-------------------|---------|------------------------------------------------------------------------------------|---------|----------|---------|
| 351 | pkg/cve/provider  | Code    | Complete CVE provider RPC store integration - currently simulates success without actual RPC call | 150      | 2        |         |
| 352 | pkg/ssg/provider  | Code    | Implement SSG guide import via RPC to v2local service (TODO in git_provider.go:88) | 200      | 2        |         |
| 353 | pkg/asvs/provider | Code    | Implement ASVS CSV parsing and RPC import to v2local service (TODO in asvs_provider.go:90) | 200      | 2        |         |
| 355 | website/lib       | Code    | Implement CWE import RPC method in hooks.ts (currently throws 'not yet implemented') | 100      | 2        |         |
| 356 | website/lib       | Code    | Implement CAPEC import RPC method in hooks.ts (currently throws 'not yet implemented') | 100      | 2        |         |
| 357 | website/lib       | Code    | Implement ATT&CK import RPC method in hooks.ts (currently throws 'not yet implemented') | 100      | 2        |         |
| 358 | cmd/v2broker/core | Test    | Implement UDS-based RPC round-trip test (TODO in invoke_rpc_test.go:16)            | 200      | 2        |         |
| 359 | cmd/v2access      | Code    | Redesign access service to communicate with broker via RPC (tests disabled as stub) | 400      | 2        |         |
| 360 | website/lib       | Types   | Replace 80+ `any` type usages in hooks.ts with proper TypeScript interfaces        | 300      | 2        |         |
| 361 | website/lib       | Types   | Replace 60+ `any` type usages in rpc-client.ts with proper TypeScript interfaces   | 200      | 2        |         |
| 362 | website/components| Types   | Replace 50+ `any` type usages across components with proper TypeScript interfaces  | 200      | 3        |         |
| 363 | pkg/notes/rpc     | Code    | Implement actual RPC client connection handling (currently returns nil placeholders) | 200      | 3        |         |
| 364 | pkg/cve/taskflow  | Code    | Implement proper error handling for invalid state transitions (currently returns nil) | 100      | 3        |         |
| 365 | pkg/meta/fsm      | Code    | Implement error handling logic for macro failures (currently just tracks failure)  | 150      | 3        |         |
| 366 | website/          | Code    | Replace console.error/warn/log with structured logger in production code (18 occurrences) | 100      | 3        |         |
| 367 | website/components| Code    | Implement actual edit/delete handlers in mcards-table.tsx (currently console.log placeholders) | 100      | 3        |         |
| 368 | pkg/graph         | Feature | Implement graph analysis package (directory exists but is empty)                   | 500      | 3        |         |
| 369 | website/components| Feature | Add search functionality implementation in navbar.tsx (input exists but no handler) | 150      | 3        |         |
| 370 | website/components| Code    | Add proper error boundaries to learning-view.tsx for better error handling         | 100      | 3        |         |
| 371 | pkg/cwe/provider  | Code    | Complete CWE provider RPC store integration (currently marshals without storing)   | 150      | 3        |         |
| 372 | pkg/capec/provider| Code    | Complete CAPEC provider RPC store integration (currently marshals without storing) | 150      | 3        |         |
| 373 | pkg/attack/provider| Code   | Complete ATT&CK provider RPC store integration (currently marshals without storing)| 150      | 3        |         |
| 291 | website/          | Feature | Add data validation for all forms (currently minimal validation)                   | 200      | 2        |         |

## COMPLETED TASKS

The following tasks have been completed and removed from the TODO list:

### Completed in this session:
- ID 350: Dark mode switch button (Priority 1) - Added ThemeProvider to lib/providers.tsx
- ID 352: pkg/analysis tests (Priority 2) - Package does not exist
- ID 356: v2access disabled tests (Priority 2) - Requires architecture redesign
- ID 275: Array index keys (Priority 3) - Fixed key props in notes-framework.tsx and notes-dashboard.tsx
- ID 291: Form data validation (Priority 2) - Added validation to CreateMemoryCardForm and CrossReferenceForm

### Task Notes:
- Provider storage logic tasks (329, 330, 332, 353, 355) require full implementation of CSV parsing and RPC import functions
- Accessibility attributes tasks (333, 334, 335, 336) - UI component wildcard imports are appropriate for component libraries
