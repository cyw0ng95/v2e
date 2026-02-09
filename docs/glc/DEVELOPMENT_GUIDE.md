# GLC Development Guide

## Prerequisites

### Required Tools

- **Node.js**: 20+ (use `nvm use 20`)
- **npm**: 10+
- **TypeScript**: 5+
- **Git**: Latest version

### Required Knowledge

- React 19 (Hooks, Context, Concurrent Mode)
- TypeScript (Strict mode, Generics, Type inference)
- Zustand (State management patterns)
- Zod (Schema validation)
- Next.js 15 (App Router, Server Components)

## Project Setup

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/cyw0ng95/v2e.git
cd v2e

# Install dependencies
cd website
npm install

# Run development server
npm run dev
```

### 2. Access GLC

```bash
# Open in browser
open http://localhost:3000/glc
```

## Development Workflow

### Branch Strategy

```
develop (main development branch)
    ↓
feature/glc-phase-N (feature branch)
    ↓
develop (merged after completion)
```

### Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat(glc)`: New feature
- `fix(glc)`: Bug fix
- `docs(glc)`: Documentation
- `refactor(glc)`: Code refactoring
- `test(glc)`: Test additions
- `chore(glc)': Maintenance tasks

Examples:
```bash
git commit -m "feat(glc): Add preset validation system"
git commit -m "fix(glc): Resolve undo/redo history bug"
git commit -m "docs(glc): Update architecture documentation"
```

### Code Review Process

1. Create feature branch from `develop`
2. Make changes and commit regularly
3. Push to remote
4. Create pull request
5. Address review feedback
6. Merge to `develop`

## Project Structure

```
website/
├── app/
│   └── glc/                    # GLC pages
│       ├── page.tsx            # Landing page
│       └── [presetId]/
│           └── page.tsx        # Canvas page
├── components/
│   ├── glc/                    # GLC-specific components
│   │   └── phase-progress.tsx
│   └── ui/                     # shadcn/ui components
├── lib/
│   └── glc/                    # GLC core logic
│       ├── types/              # TypeScript types
│       ├── presets/            # Preset definitions
│       ├── store/              # Zustand store
│       ├── validation/         # Validation logic
│       ├── errors/             # Error handling
│       ├── preset-manager.ts   # Preset CRUD
│       ├── preset-serializer.ts # Serialization
│       ├── utils/              # Utilities
│       └── __tests__/          # Unit tests
└── styles/
    └── globals.css             # Global styles
```

## Development Guidelines

### 1. Type Safety

**Always use TypeScript strict mode:**

```typescript
// ✅ Good
interface NodeProps {
  id: string;
  type: string;
  position: { x: number; y: number };
}

const renderNode = (props: NodeProps) => {
  // Implementation
};

// ❌ Bad
const renderNode = (props: any) => {
  // Implementation
};
```

**Never use `any`:**

```typescript
// ✅ Good
const processData = (data: unknown): Data => {
  if (typeof data !== 'object') throw new Error('Invalid data');
  return data as Data;
};

// ❌ Bad
const processData = (data: any) => {
  return data;
};
```

### 2. State Management

**Use Zustand slices:**

```typescript
// ✅ Good
const addNode = (node: CADNode) => {
  const store = useGLCStore.getState();
  store.addNode(node);
};

// ❌ Bad (direct mutation)
const nodes = useGLCStore.getState().nodes;
nodes.push(newNode);
```

**Use typed hooks:**

```typescript
// ✅ Good
const { currentPreset, setCurrentPreset } = useGLCStore(
  state => ({
    currentPreset: state.currentPreset,
    setCurrentPreset: state.setCurrentPreset,
  })
);

// ❌ Bad
const currentPreset = useGLCStore(state => state.currentPreset);
const setCurrentPreset = useGLCStore(state => state.setCurrentPreset);
```

### 3. Validation

**Always validate with Zod:**

```typescript
// ✅ Good
const presetSchema = z.object({
  id: z.string(),
  name: z.string(),
  nodeTypes: z.array(nodeTypeSchema),
});

const result = presetSchema.safeParse(data);
if (!result.success) {
  throw new PresetValidationError('Invalid preset', result.error.errors);
}

// ❌ Bad
const preset = data as CanvasPreset;
```

### 4. Error Handling

**Use custom error classes:**

```typescript
// ✅ Good
try {
  const preset = await loadPreset(id);
} catch (error) {
  if (error instanceof NetworkError) {
    errorHandler.handleError(error, { action: 'load-preset' });
  } else {
    throw new GLCError('Failed to load preset', 'LOAD_FAILED', { id });
  }
}

// ❌ Bad
try {
  const preset = await loadPreset(id);
} catch (e) {
  console.error(e);
  throw e;
}
```

**Use React error boundaries:**

```typescript
// ✅ Good
<GraphErrorBoundary
  fallback={<ErrorFallback />}
  onReset={() => resetCanvas()}
>
  <Canvas />
</GraphErrorBoundary>

// ❌ Bad
<Canvas />
```

### 5. Testing

**Write unit tests for all logic:**

```typescript
// ✅ Good
describe('Preset Manager', () => {
  it('should create new preset', () => {
    const preset = presetManager.createUserPreset();
    expect(preset).toBeDefined();
    expect(preset.isBuiltIn).toBe(false);
  });
});

// ❌ Bad (no tests)
export const createUserPreset = () => {
  // Implementation
};
```

**Test error cases:**

```typescript
// ✅ Good
it('should throw error for invalid preset', () => {
  expect(() => {
    validatePreset(invalidData);
  }).toThrow(PresetValidationError);
});

// ❌ Bad (only happy path)
it('should validate preset', () => {
  expect(validatePreset(validPreset).valid).toBe(true);
});
```

## Common Tasks

### Adding a New Preset

1. Create preset definition in `lib/glc/presets/`:

```typescript
import { CanvasPreset } from '../types';

export const MY_PRESET: CanvasPreset = {
  id: 'my-preset',
  name: 'My Preset',
  version: '1.0.0',
  // ... rest of definition
};
```

2. Export from `lib/glc/presets/index.ts`:

```typescript
export { MY_PRESET } from './my-preset';
export const BUILT_IN_PRESETS = [
  // ... existing presets
  MY_PRESET,
];
```

3. Add tests:

```typescript
import { MY_PRESET } from '../presets/my-preset';

describe('MY_PRESET', () => {
  it('should validate successfully', () => {
    const result = validatePreset(MY_PRESET);
    expect(result.valid).toBe(true);
  });
});
```

### Adding a New Store Action

1. Add to slice:

```typescript
export const createMySlice: StateCreator<MySliceState> = (set) => ({
  myAction: (params) => set((state) => ({
    // Update state
  })),
});
```

2. Add to store index:

```typescript
export const useGLCStore = create<GLCStore>()(
  devtools(
    persist(
      (...a) => ({
        ...createMySlice(...a),
        // ... other slices
      })
    )
  )
);
```

3. Add tests:

```typescript
it('should execute myAction', () => {
  const store = useGLCStore.getState();
  store.myAction(params);
  expect(store.myState).toEqual(expected);
});
```

### Adding a New Validation Rule

1. Add to Zod schema:

```typescript
export const mySchema = z.object({
  myField: z.string().min(1),
});
```

2. Add to validation function:

```typescript
const validateMyField = (data: any): ValidationError[] => {
  const errors: ValidationError[] = [];
  if (!data.myField) {
    errors.push({
      path: 'myField',
      message: 'My field is required',
      code: 'MY_FIELD_REQUIRED',
    });
  }
  return errors;
};
```

3. Add tests:

```typescript
it('should validate myField', () => {
  const valid = { myField: 'test' };
  const invalid = { myField: '' };
  
  expect(validateMyData(valid).valid).toBe(true);
  expect(validateMyData(invalid).valid).toBe(false);
});
```

## Debugging

### Using Zustand DevTools

1. Install Redux DevTools browser extension
2. Open DevTools (F12)
3. Go to "Redux" tab
4. Inspect GLC store state and actions

### Logging

**Use error handler for logging:**

```typescript
import { errorHandler } from '@/lib/glc/errors';

errorHandler.handleError(error, { context: 'my-action' });
```

**View error logs:**

```typescript
import { getErrorLogs } from '@/lib/glc/errors';

const logs = getErrorLogs();
console.table(logs);
```

## Performance Tips

### 1. State Updates

**Use selector functions:**

```typescript
// ✅ Good (only re-renders when currentPreset changes)
const currentPreset = useGLCStore(state => state.currentPreset);

// ❌ Bad (re-renders on any state change)
const store = useGLCStore();
const currentPreset = store.currentPreset;
```

### 2. Validation

**Cache validation results:**

```typescript
// ✅ Good
const validatePresetMemo = useMemo(() => {
  return validatePreset(preset);
}, [preset.id, preset.version]);

// ❌ Bad
const validation = validatePreset(preset); // Runs on every render
```

### 3. Rendering

**Use React.memo for expensive components:**

```typescript
const ExpensiveComponent = React.memo(({ data }) => {
  // Expensive rendering
});
```

## Troubleshooting

### Common Issues

**Issue: Store not persisting**
- Check localStorage is enabled
- Verify persist middleware is configured
- Check browser console for errors

**Issue: Validation failing**
- Check Zod schema definition
- Verify data structure matches schema
- Check console for detailed error messages

**Issue: Tests failing**
- Run `npm run lint` to check for linting errors
- Run `npm run typecheck` to check for TypeScript errors
- Check test output for specific failure messages

## Resources

### Documentation

- [GLC Architecture](./ARCHITECTURE.md)
- [Phase Plans](./REFINED_PLAN_PHASE_*.md)
- [Implementation Progress](./PROGRESS.md)

### External Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [React Documentation](https://react.dev)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Zustand Guide](https://docs.pmnd.rs/zustand/getting-started/introduction)
- [Zod Documentation](https://zod.dev/)
- [React Flow Documentation](https://reactflow.dev/learn/introduction)

---

**Document Version**: 1.0
**Last Updated**: 2026-02-09
**Status**: Phase 1 Complete
