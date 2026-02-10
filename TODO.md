# MAINTENANCE TODO

This document tracks maintenance tasks for the v2e project. Tasks are organized by priority and type.

## TODO

| ID  | Package           | Type    | Description                                                                        | Est LoC | Priority | WONTFIX |
|-----|-------------------|---------|------------------------------------------------------------------------------------|---------|----------|---------|
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
