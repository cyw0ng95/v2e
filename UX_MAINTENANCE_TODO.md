# UX/HI/Vision Design Modernization Tasks

This document lists design and UX improvements for the v2e frontend. Each task is designed to be completed by a single agent.

## Current Design Assessment

### Visual Design Issues
- Lack of cohesive modern design system
- Color system uses OKLCH but lacks visual hierarchy
- Missing unified design token management
- Excessive white/gray colors lacking visual impact
- Inconsistent spacing and typography scales

### UX Issues
- Complex navigation: 11 primary tabs, hard to discover
- View/Learn mode switching unclear
- No breadcrumb navigation
- Overly complex tables and dialogs
- Missing onboarding/guided tours
- No empty states or loading states
- Inconsistent error handling

### Component Architecture Issues
- Large components (>300 lines): notes-framework.tsx (724 lines), ssg-views.tsx (939 lines), ui/sidebar.tsx (726 lines)
- Missing component reuse patterns
- Scattered state management

### Accessibility Issues
- Only 19 aria-label attributes (55 components total)
- Only 12 role attributes
- Missing keyboard navigation support
- Missing screen reader support

### Performance Issues
- Dynamic imports without proper code splitting strategy
- Missing virtual scrolling (especially for tables)
- Missing memo optimization

## Design Tasks (Agent-Ready)

Each task below is self-contained and can be executed by a single agent independently.

---

### Task 1: Design System Foundation
**ID**: UX-001
**Priority**: 1 (Critical)
**Est LoC**: 300
**Package**: website/lib

**Description**: Create a unified design system with design tokens for spacing, typography, colors, and effects.

**Closure Details**:
```typescript
// Create website/lib/design-system.ts with:
export const designTokens = {
  // Spacing scale (0-12)
  spacing: { 0: '0', 1: '0.25rem', 2: '0.5rem', 3: '0.75rem', 4: '1rem', 5: '1.25rem', 6: '1.5rem', 8: '2rem', 10: '2.5rem', 12: '3rem' },
  // Typography scale (xs-9xl)
  typography: {
    xs: { fontSize: '0.75rem', lineHeight: '1rem', fontWeight: 400 },
    sm: { fontSize: '0.875rem', lineHeight: '1.25rem', fontWeight: 400 },
    base: { fontSize: '1rem', lineHeight: '1.5rem', fontWeight: 400 },
    // ... complete to 9xl
  },
  // Border radius scale (sm-3xl)
  radius: { sm: '0.375rem', md: '0.5rem', lg: '0.625rem', xl: '0.75rem', '2xl': '1rem', '3xl': '1.5rem' },
  // Shadows (xs-2xl)
  shadows: { xs: '0 1px 2px', sm: '0 1px 3px', md: '0 4px 6px', lg: '0 10px 15px', xl: '0 20px 25px', '2xl': '0 25px 50px' },
  // Transitions
  transitions: { fast: '150ms ease-out', normal: '300ms ease-out', slow: '500ms ease-out' },
  // Z-index scale (0-50)
  zIndex: { dropdown: 1000, sticky: 1020, fixed: 1030, modalBackdrop: 1040, modal: 1050, popover: 1060, tooltip: 1070 }
};

// Export utility functions
export const getSpacing = (key: keyof typeof designTokens.spacing) => designTokens.spacing[key];
export const getTypography = (key: keyof typeof designTokens.typography) => designTokens.typography[key];
// ... similar utilities for radius, shadows, transitions, zIndex

// Update globals.css to use these design tokens via CSS variables
```

**Success Criteria**:
- All spacing, typography, radius, and shadow values are centralized
- CSS variables are added to globals.css
- Existing components reference design tokens
- Design tokens follow 8pt grid system
- Typography scale uses major third (1.2) or major second (1.5) ratios

---

### Task 2: Modern Color Palette
**ID**: UX-002
**Priority**: 1 (Critical)
**Est LoC**: 200
**Package**: website/lib

**Description**: Redesign color palette with modern, accessible colors and proper semantic mapping.

**Closure Details**:
```typescript
// Update website/lib/design-system.ts color section:
export const colors = {
  // Primary - Modern indigo/violet gradient
  primary: {
    50: '#eef2ff',
    100: '#e0e7ff',
    200: '#c7d2fe',
    300: '#a5b4fc',
    400: '#818cf8',
    500: '#6366f1', // Base
    600: '#4f46e5',
    700: '#4338ca',
    800: '#3730a3',
    900: '#312e81',
  },
  // Secondary - Neutral with subtle warmth
  secondary: {
    50: '#fafafa',
    100: '#f4f4f5',
    200: '#e4e4e7',
    300: '#d4d4d8',
    400: '#a1a1aa',
    500: '#71717a', // Base
    600: '#52525b',
    700: '#3f3f46',
    800: '#27272a',
    900: '#18181b',
  },
  // Success - Modern green with blue undertone
  success: {
    50: '#f0fdf4',
    100: '#dcfce7',
    200: '#bbf7d0',
    300: '#86efac',
    400: '#4ade80',
    500: '#22c55e', // Base
    600: '#16a34a',
    700: '#15803d',
    800: '#166534',
    900: '#14532d',
  },
  // Warning - Modern amber/orange
  warning: {
    50: '#fffbeb',
    100: '#fef3c7',
    200: '#fde68a',
    300: '#fcd34d',
    400: '#fbbf24',
    500: '#f59e0b', // Base
    600: '#d97706',
    700: '#b45309',
    800: '#92400e',
    900: '#78350f',
  },
  // Error - Modern red with orange undertone
  error: {
    50: '#fef2f2',
    100: '#fee2e2',
    200: '#fecaca',
    300: '#fca5a5',
    400: '#f87171',
    500: '#ef4444', // Base
    600: '#dc2626',
    700: '#b91c1c',
    800: '#991b1b',
    900: '#7f1d1d',
  },
  // Semantic colors for specific use cases
  semantic: {
    info: '#6366f1',     // Primary blue-violet
    positive: '#10b981', // Emerald
    negative: '#ef4444',  // Red
    warning: '#f59e0b',  // Amber
    neutral: '#6b7280',  // Gray
  },
  // Neutral scales for surfaces
  neutral: {
    0: '#ffffff',
    50: '#fafafa',
    100: '#f4f4f5',
    200: '#e4e4e7',
    300: '#d4d4d8',
    400: '#a1a1aa',
    500: '#71717a',
    600: '#52525b',
    700: '#3f3f46',
    800: '#27272a',
    900: '#18181b',
    950: '#09090b',
  }
};

// Export color utilities
export const getColor = (color: string, shade: number = 500) => {
  const [colorName] = color.split('-');
  return colors[colorName]?.[shade] || color;
};

// Update globals.css to use these colors via CSS variables
// Ensure WCAG AA contrast ratios (4.5:1 for text, 3:1 for large text)
```

**Success Criteria**:
- Color palette follows modern design trends (indigo/violet primary)
- All color scales (50-900) are complete
- WCAG AA contrast ratios met (verify with automated tools)
- Dark mode colors properly inverted
- Semantic colors map to appropriate base colors
- CSS variables updated in globals.css

---

### Task 3: Modern Typography System
**ID**: UX-003
**Priority**: 1 (Critical)
**Est LoC**: 150
**Package**: website/lib

**Description**: Create a comprehensive typography system with font weights, sizes, and line heights.

**Closure Details**:
```typescript
// Update website/lib/design-system.ts typography section:
export const typography = {
  fontFamily: {
    sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
    mono: ['JetBrains Mono', 'Fira Code', 'Consolas', 'Monaco', 'monospace'],
  },
  fontSize: {
    // Display sizes for headings
    display: {
      xs: { fontSize: '2.5rem', lineHeight: '1.2', fontWeight: 700, letterSpacing: '-0.02em' },
      sm: { fontSize: '3rem', lineHeight: '1.2', fontWeight: 700, letterSpacing: '-0.02em' },
      md: { fontSize: '3.75rem', lineHeight: '1.2', fontWeight: 700, letterSpacing: '-0.02em' },
      lg: { fontSize: '4.5rem', lineHeight: '1.1', fontWeight: 700, letterSpacing: '-0.02em' },
      xl: { fontSize: '6rem', lineHeight: '1', fontWeight: 700, letterSpacing: '-0.03em' },
    },
    // Heading sizes
    heading: {
      1: { fontSize: '2.25rem', lineHeight: '1.25', fontWeight: 700, letterSpacing: '-0.01em' },
      2: { fontSize: '1.875rem', lineHeight: '1.3', fontWeight: 600, letterSpacing: '-0.01em' },
      3: { fontSize: '1.5rem', lineHeight: '1.4', fontWeight: 600, letterSpacing: '-0.005em' },
      4: { fontSize: '1.25rem', lineHeight: '1.5', fontWeight: 600 },
      5: { fontSize: '1.125rem', lineHeight: '1.5', fontWeight: 600 },
      6: { fontSize: '1rem', lineHeight: '1.5', fontWeight: 600 },
    },
    // Body sizes
    body: {
      xs: { fontSize: '0.75rem', lineHeight: '1.4', fontWeight: 400 },
      sm: { fontSize: '0.875rem', lineHeight: '1.5', fontWeight: 400 },
      base: { fontSize: '1rem', lineHeight: '1.6', fontWeight: 400 },
      lg: { fontSize: '1.125rem', lineHeight: '1.7', fontWeight: 400 },
      xl: { fontSize: '1.25rem', lineHeight: '1.75', fontWeight: 400 },
      '2xl': { fontSize: '1.5rem', lineHeight: '1.8', fontWeight: 400 },
    },
    // Caption/label sizes
    caption: {
      xs: { fontSize: '0.625rem', lineHeight: '1.3', fontWeight: 500 },
      sm: { fontSize: '0.75rem', lineHeight: '1.4', fontWeight: 500 },
      base: { fontSize: '0.875rem', lineHeight: '1.4', fontWeight: 500 },
    },
  },
  fontWeight: {
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
    extrabold: 800,
  },
  letterSpacing: {
    tighter: '-0.05em',
    tight: '-0.025em',
    normal: '0em',
    wide: '0.025em',
    wider: '0.05em',
    widest: '0.1em',
  },
  lineHeight: {
    none: 1,
    tight: 1.25,
    snug: 1.375,
    normal: 1.5,
    relaxed: 1.625,
    loose: 2,
  },
};

// Export typography utilities
export const getFontSize = (category: string, size: string) => {
  return typography.fontSize[category]?.[size] || typography.fontSize.body.base;
};
```

**Success Criteria**:
- Typography scale follows major third (1.2) or major second (1.5) ratios
- Font family includes web-safe fallbacks
- Letter spacing adjusts based on font size
- Line height optimizes for readability (1.5-1.75 for body text)
- Font weights are consistent across the system
- CSS utility classes added for common typography needs

---

### Task 4: Icon System Refinement
**ID**: UX-004
**Priority**: 2 (Important)
**Est LoC**: 150
**Package**: website/components

**Description**: Create a consistent icon system with standardized sizes and stroke widths.

**Closure Details**:
```typescript
// Update website/components/icons.tsx:
import { LucideIcon, LucideProps } from 'lucide-react';

export const iconSizes = {
  xs: 14,
  sm: 16,
  base: 20,
  md: 24,
  lg: 28,
  xl: 32,
  '2xl': 36,
  '3xl': 48,
} as const;

export const iconWeights = {
  light: 1,
  regular: 1.5,
  medium: 2,
  bold: 2.5,
} as const;

export type IconSize = keyof typeof iconSizes;
export type IconWeight = keyof typeof iconWeights;

// Create wrapper component for consistent icon styling
interface IconProps extends Omit<LucideProps, 'size'> {
  size?: IconSize;
  weight?: IconWeight;
}

export const Icon = ({ size = 'md', weight = 'regular', ...props }: IconProps & { children: LucideIcon }) => {
  const IconComponent = props.children;
  return (
    <IconComponent
      size={iconSizes[size]}
      strokeWidth={iconWeights[weight]}
      {...props}
    />
  );
};

// Update all icon exports to use consistent sizes
// Example:
export const DatabaseIcon = (props: Omit<LucideProps, 'size'> & { size?: IconSize }) => (
  <Database size={iconSizes[props.size || 'md']} strokeWidth={iconWeights.regular} {...props} />
);
// ... repeat for all icons
```

**Success Criteria**:
- All icons use consistent size scale (14px to 48px)
- Stroke weight is standardized (1-2.5)
- Icon wrapper component enforces consistency
- All existing icon usages updated to use new system
- Icon exports are type-safe

---

### Task 5: Button System Standardization
**ID**: UX-005
**Priority**: 1 (Critical)
**Est LoC**: 250
**Package**: website/components/ui

**Description**: Standardize button variants, sizes, and states with proper accessibility.

**Closure Details**:
```typescript
// Refactor website/components/ui/button.tsx:
import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-md text-sm font-medium transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        // Primary - Main action buttons
        primary: 'bg-primary text-primary-foreground hover:bg-primary/90 hover:scale-[1.02] active:scale-[0.98] shadow-sm',
        // Secondary - Supporting actions
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80 hover:-translate-y-px active:translate-y-px',
        // Outline - Ghost with border
        outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
        // Ghost - No background
        ghost: 'hover:bg-accent hover:text-accent-foreground',
        // Destructive - Delete/cancel actions
        destructive: 'bg-error text-error-foreground hover:bg-error/90 hover:scale-[1.02] active:scale-[0.98]',
        // Success - Confirm actions
        success: 'bg-success text-success-foreground hover:bg-success/90 hover:scale-[1.02] active:scale-[0.98]',
        // Warning - Caution actions
        warning: 'bg-warning text-warning-foreground hover:bg-warning/90 hover:scale-[1.02] active:scale-[0.98]',
        // Link - Text-only button
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        xs: 'h-7 px-2 text-xs',
        sm: 'h-8 px-3 text-xs',
        md: 'h-9 px-4 text-sm',
        lg: 'h-10 px-5 text-base',
        xl: 'h-12 px-6 text-lg',
        icon: 'h-9 w-9',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  }
);

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {
  asChild?: boolean;
  loading?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, loading, leftIcon, rightIcon, disabled, children, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || loading}
        aria-disabled={disabled || loading}
        {...props}
      >
        {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        {!loading && leftIcon && <span className="mr-2">{leftIcon}</span>}
        {children}
        {!loading && rightIcon && <span className="ml-2">{rightIcon}</span>}
      </button>
    );
  }
);
Button.displayName = 'Button';

export { Button, buttonVariants };
```

**Success Criteria**:
- 7 button variants (primary, secondary, outline, ghost, destructive, success, warning)
- 6 sizes (xs, sm, md, lg, xl, icon)
- Loading state with spinner
- Left and right icon support
- Proper hover/active animations
- Accessibility attributes (aria-disabled, focus-visible)
- All existing button usages updated

---

### Task 6: Form Elements Design
**ID**: UX-006
**Priority**: 2 (Important)
**Est LoC**: 300
**Package**: website/components/ui

**Description**: Standardize form inputs, selects, and checkboxes with consistent styling and states.

**Closure Details**:
```typescript
// Refactor website/components/ui/input.tsx:
import * as React from 'react';
import { cva } from 'class-variance-authority';

const inputVariants = cva(
  'flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm transition-all duration-150 file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
  {
    variants: {
      size: {
        sm: 'h-8 px-2 text-xs',
        md: 'h-9 px-3 text-sm',
        lg: 'h-10 px-4 text-base',
        xl: 'h-12 px-5 text-lg',
      },
      state: {
        default: 'border-border focus-visible:border-primary',
        error: 'border-error focus-visible:border-error focus-visible:ring-error/20',
        success: 'border-success focus-visible:border-success focus-visible:ring-success/20',
      },
    },
    defaultVariants: {
      size: 'md',
      state: 'default',
    },
  }
);

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement>, VariantProps<typeof inputVariants> {
  label?: string;
  error?: string;
  helperText?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, size, state, label, error, helperText, id, ...props }, ref) => {
    const inputId = id || `input-${React.useId()}`;
    const errorMessage = state === 'error' ? error : undefined;

    return (
      <div className="space-y-1">
        {label && (
          <label htmlFor={inputId} className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
            {label}
          </label>
        )}
        <input
          id={inputId}
          className={cn(inputVariants({ size, state, className }))}
          aria-invalid={!!errorMessage}
          aria-describedby={errorMessage ? `${inputId}-error` : helperText ? `${inputId}-helper` : undefined}
          ref={ref}
          {...props}
        />
        {errorMessage && (
          <p id={`${inputId}-error`} className="text-xs text-error">
            {errorMessage}
          </p>
        )}
        {helperText && !errorMessage && (
          <p id={`${inputId}-helper`} className="text-xs text-muted-foreground">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);
Input.displayName = 'Input';

// Similar updates for select.tsx, checkbox.tsx, textarea.tsx
```

**Success Criteria**:
- Consistent styling across all form elements
- 4 sizes (sm, md, lg, xl)
- State variants (default, error, success)
- Label and helper text support
- Error message display
- Accessibility attributes (aria-invalid, aria-describedby)
- Focus states with ring
- All existing form elements updated

---

### Task 7: Card Component Enhancement
**ID**: UX-007
**Priority**: 2 (Important)
**Est LoC**: 200
**Package**: website/components/ui

**Description**: Enhance card component with variants, elevation, and interactive states.

**Closure Details**:
```typescript
// Refactor website/components/ui/card.tsx:
import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

const cardVariants = cva(
  'rounded-xl border bg-card text-card-foreground shadow-sm transition-all duration-150',
  {
    variants: {
      elevation: {
        none: 'shadow-none',
        sm: 'shadow-sm hover:shadow-md',
        md: 'shadow-md hover:shadow-lg',
        lg: 'shadow-lg hover:shadow-xl',
        xl: 'shadow-xl hover:shadow-2xl',
      },
      variant: {
        default: 'border-border',
        elevated: 'border-border/50',
        outlined: 'border-2',
        ghost: 'border-transparent hover:bg-accent/50',
      },
      interactive: {
        true: 'cursor-pointer hover:border-primary/50 hover:shadow-md hover:-translate-y-px active:translate-y-px',
        false: '',
      },
    },
    defaultVariants: {
      elevation: 'sm',
      variant: 'default',
      interactive: false,
    },
  }
);

export interface CardProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof cardVariants> {}

const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, elevation, variant, interactive, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(cardVariants({ elevation, variant, interactive }), className)}
      {...props}
    />
  )
);
Card.displayName = 'Card';

// Update CardHeader, CardTitle, CardDescription, CardContent, CardFooter with consistent typography
```

**Success Criteria**:
- 5 elevation levels (none, sm, md, lg, xl)
- 4 variants (default, elevated, outlined, ghost)
- Interactive mode with hover/active states
- Consistent typography for child components
- All existing card usages updated

---

### Task 8: Navigation Refactoring
**ID**: UX-008
**Priority**: 1 (Critical)
**Est LoC**: 400
**Package**: website/components

**Description**: Simplify navigation by consolidating tabs and adding breadcrumb navigation.

**Closure Details**:
```typescript
// Create website/components/navigation/sidebar.tsx:
import { NavGroup, NavItem } from './types';

// Navigation structure with logical grouping
const navigation: NavGroup[] = [
  {
    title: 'Database',
    items: [
      { id: 'cve', label: 'CVE', icon: Database, badge: 'count' },
      { id: 'cwe', label: 'CWE', icon: Shield },
      { id: 'capec', label: 'CAPEC', icon: AlertTriangle },
      { id: 'attack', label: 'ATT&CK', icon: Zap },
    ],
  },
  {
    title: 'Learning',
    items: [
      { id: 'notes-dashboard', label: 'Dashboard', icon: LayoutDashboard },
      { id: 'study-cards', label: 'Study Cards', icon: Brain },
      { id: 'learning-cve', label: 'Learn CVE', icon: BookOpen },
      { id: 'learning-cwe', label: 'Learn CWE', icon: Book },
    ],
  },
  {
    title: 'Analysis',
    items: [
      { id: 'graph', label: 'Graph Analysis', icon: Network },
      { id: 'cweviews', label: 'CWE Views', icon: GitBranch },
      { id: 'ssg', label: 'SSG Guides', icon: FileCode },
      { id: 'asvs', label: 'ASVS', icon: CheckCircle },
    ],
  },
  {
    title: 'System',
    items: [
      { id: 'bookmarks', label: 'Bookmarks', icon: Bookmark },
      { id: 'sysmon', label: 'System Monitor', icon: Activity },
      { id: 'etl', label: 'ETL Status', icon: RefreshCw },
    ],
  },
];

// Refactor app/page.tsx to use grouped navigation instead of flat tabs
// Add breadcrumb navigation component
```

**Success Criteria**:
- Navigation grouped into 4 logical categories
- Reduced from 11 to 8 primary navigation items
- Collapsible sidebar for mobile
- Breadcrumb component added for context
- Search functionality integrated
- Active state clearly indicated
- Keyboard navigation support

---

### Task 9: Empty States and Loading States
**ID**: UX-009
**Priority**: 2 (Important)
**Est LoC**: 250
**Package**: website/components

**Description**: Create reusable empty state and loading state components.

**Closure Details**:
```typescript
// Create website/components/empty-state.tsx:
interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export const EmptyState = ({ icon, title, description, action }: EmptyStateProps) => (
  <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
    {icon && <div className="mb-4 text-muted-foreground/50">{icon}</div>}
    <h3 className="text-lg font-semibold mb-2">{title}</h3>
    {description && <p className="text-sm text-muted-foreground mb-4 max-w-md">{description}</p>}
    {action && (
      <Button onClick={action.onClick} variant="primary">
        {action.label}
      </Button>
    )}
  </div>
);

// Create website/components/loading-state.tsx:
interface LoadingStateProps {
  message?: string;
  size?: 'sm' | 'md' | 'lg';
}

export const LoadingState = ({ message, size = 'md' }: LoadingStateProps) => (
  <div className="flex flex-col items-center justify-center py-12">
    <Loader2 className={`animate-spin text-primary mb-4 ${size === 'sm' ? 'h-6 w-6' : size === 'lg' ? 'h-12 w-12' : 'h-8 w-8'}`} />
    {message && <p className="text-sm text-muted-foreground">{message}</p>}
  </div>
);

// Add skeleton variations for different content types
```

**Success Criteria**:
- EmptyState component with icon, title, description, and action
- LoadingState component with message and size variants
- Skeleton loaders for cards, tables, and lists
- Consistent visual style across all empty/loading states
- All tables and data displays use these components

---

### Task 10: Error Handling and Toasts
**ID**: UX-010
**Priority**: 1 (Critical)
**Est LoC**: 200
**Package**: website/lib

**Description**: Implement error boundary and consistent toast notifications using Sonner.

**Closure Details**:
```typescript
// Create website/lib/error-handler.tsx:
import { toast } from 'sonner';

export const errorTypes = {
  NETWORK_ERROR: 'network_error',
  VALIDATION_ERROR: 'validation_error',
  AUTH_ERROR: 'auth_error',
  UNKNOWN_ERROR: 'unknown_error',
};

export const showError = (error: Error | string, context?: string) => {
  const errorMessage = error instanceof Error ? error.message : error;
  console.error(context ? `[${context}] ${errorMessage}` : errorMessage);

  toast.error(errorMessage, {
    description: context,
    action: {
      label: 'Dismiss',
      onClick: () => {},
    },
  });
};

export const showSuccess = (message: string) => {
  toast.success(message);
};

export const showWarning = (message: string) => {
  toast.warning(message);
};

export const showInfo = (message: string) => {
  toast.info(message);
};

// Update website/components/error-boundary.tsx to use these handlers
```

**Success Criteria**:
- Consistent error messages using Sonner
- Error boundary catches React errors
- Toast notifications for all user-facing errors
- Actionable error messages with context
- Success/info/warning toast variants
- All API calls wrapped with error handling

---

### Task 11: Accessibility Improvements
**ID**: UX-011
**Priority**: 1 (Critical)
**Est LoC**: 400
**Package**: website/components

**Description**: Add ARIA labels, roles, and keyboard navigation to all interactive components.

**Closure Details**:
```typescript
// For each component, add:
// 1. ARIA labels for buttons without text
<Button aria-label="Close dialog" />

// 2. Role attributes for custom interactive elements
<div role="button" tabIndex={0} onKeyPress={...} />

// 3. ARIA live regions for dynamic content
<div aria-live="polite" aria-atomic="true">
  {loadingMessage}
</div>

// 4. Keyboard navigation
const handleKeyDown = (e: React.KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ' ') {
    handleClick();
  }
};

<div
  role="button"
  tabIndex={0}
  onKeyDown={handleKeyDown}
  onClick={handleClick}
>

// 5. Focus management in modals
// Trap focus within modal
// Return focus to trigger after close

// 6. Skip links for keyboard users
<a href="#main-content" className="sr-only focus:not-sr-only">
  Skip to main content
</a>

// Target components:
// - All buttons (55 total, currently 19 with aria-label)
// - All interactive divs (currently 12 with role)
// - Modals and dialogs
// - Tabs navigation
// - Dropdown menus
// - Form inputs
```

**Success Criteria**:
- All buttons have aria-label or visible text
- All custom interactive elements have role attributes
- Keyboard navigation works for all components
- Focus management in modals
- ARIA live regions for dynamic content
- Skip link for keyboard users
- Screen reader testing passes
- WCAG 2.1 AA compliance verified

---

### Task 12: Dark Mode Enhancement
**ID**: UX-012
**Priority**: 2 (Important)
**Est LoC**: 150
**Package**: website/app

**Description**: Improve dark mode colors, contrast, and transitions.

**Closure Details**:
```typescript
// Update website/app/globals.css dark mode colors:
.dark {
  --background: oklch(0.12 0.005 264);  // Darker background
  --foreground: oklch(0.95 0.005 264);  // Lighter text
  --card: oklch(0.15 0.008 264);
  --border: oklch(1 0 0 / 15%);  // More visible borders
  --input: oklch(1 0 0 / 20%);
  --primary: oklch(0.65 0.18 264);  // Lighter primary
  --primary-foreground: oklch(0.12 0.005 264);
  --secondary: oklch(0.20 0.01 264);
  --secondary-foreground: oklch(0.95 0.005 264);
  --muted: oklch(0.20 0.012 264);
  --muted-foreground: oklch(0.65 0.01 264);
  --accent: oklch(0.65 0.18 264);
  --accent-foreground: oklch(0.12 0.005 264);

  // Ensure WCAG AA contrast ratios
  // Test with axe DevTools or similar
}

// Add smooth theme transitions
* {
  transition-property: background-color, border-color, color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

// Improve learn mode in dark
[data-learn-mode="true"].dark {
  --learn-focus-bg: oklch(0.12 0.008 264);
  --learn-focus-text: oklch(0.98 0.005 264);
  --learn-focus-border: oklch(1 0 0 / 20%);
}
```

**Success Criteria**:
- Dark mode colors have adequate contrast (4.5:1 for text, 3:1 for large text)
- Smooth transitions between light/dark modes
- All colors tested in both modes
- Learn mode works well in dark mode
- No jarring color changes

---

### Task 13: Onboarding and Help System
**ID**: UX-013
**Priority**: 3 (Nice to Have)
**Est LoC**: 350
**Package**: website/components

**Description**: Create onboarding tour, tooltips, and help documentation.

**Closure Details**:
```typescript
// Create website/components/onboarding/tour.tsx:
import { useTour } from '@reactour/tour';

const tourSteps = [
  {
    selector: '.sidebar-nav',
    content: 'Navigate between different data sources using the sidebar.',
  },
  {
    selector: '.view-learn-toggle',
    content: 'Switch between View mode for exploration and Learn mode for studying.',
  },
  {
    selector: '.search-bar',
    content: 'Search across CVEs, CWEs, CAPECs, and ATT&CK techniques.',
  },
  {
    selector: '.cve-table',
    content: 'Browse CVE records with filtering and pagination.',
  },
  {
    selector: '.session-control',
    content: 'Manage ETL data import sessions.',
  },
];

export function OnboardingTour() {
  const { setIsOpen, setSteps, currentStep, setIsOpen } = useTour();

  React.useEffect(() => {
    setSteps(tourSteps);
    const hasSeenTour = localStorage.getItem('hasSeenTour');
    if (!hasSeenTour) {
      setIsOpen(true);
    }
  }, []);

  const handleTourEnd = () => {
    localStorage.setItem('hasSeenTour', 'true');
    setIsOpen(false);
  };

  return <Tour steps={tourSteps} afterClose={handleTourEnd} />;
}

// Create website/components/help/tooltip.tsx:
// Add contextual help tooltips to key features
// Add keyboard shortcut hints (Cmd/Ctrl + K for search, etc.)
```

**Success Criteria**:
- Onboarding tour shows key features to new users
- Tour can be dismissed and replayed
- Contextual tooltips explain complex features
- Keyboard shortcuts documented
- Help documentation accessible
- Tour persists user preference (localStorage)

---

### Task 14: Responsive Design Improvements
**ID**: UX-014
**Priority**: 2 (Important)
**Est LoC**: 300
**Package**: website/app

**Description**: Improve mobile and tablet layouts with proper breakpoints.

**Closure Details**:
```typescript
// Update website/app/layout.tsx:
// Add mobile-specific sidebar
// Ensure proper breakpoints: sm (640px), md (768px), lg (1024px), xl (1280px)

// Create mobile drawer for navigation
// Example: website/components/navigation/mobile-drawer.tsx

// Update page.tsx:
<main className="h-full flex flex-col px-10 py-8">
  {/* Desktop: show both sidebar and main */}
  <div className="hidden md:flex h-full gap-6">
    <aside className="w-80 shrink-0">...</aside>
    <div className="flex-1">...</div>
  </div>

  {/* Mobile: show drawer + main */}
  <div className="flex md:hidden h-full">
    <MobileDrawer />
    <div className="flex-1 overflow-auto">...</div>
  </div>
</main>

// Update breakpoints in globals.css:
// sm: '640px', md: '768px', lg: '1024px', xl: '1280px', '2xl': '1536px'

// Test on mobile: iPhone SE (375px), iPad (768px)
// Test on tablet: iPad Pro (1024px)
```

**Success Criteria**:
- Mobile layout works on phones (375px - 428px)
- Tablet layout works on iPads (768px - 1024px)
- Desktop layout optimized for 1280px+
- Sidebar collapses to drawer on mobile
- Tables scroll horizontally on mobile
- Touch targets are at least 44x44px
- No horizontal scroll on mobile

---

### Task 15: Performance Optimization
**ID**: UX-015
**Priority**: 2 (Important)
**Est LoC**: 400
**Package**: website/components

**Description**: Optimize component rendering with memo, lazy loading, and virtualization.

**Closure Details**:
```typescript
// 1. Add React.memo to large components
export const CVETable = React.memo(function CVETable({ cves, ... }) {
  // ... existing code
}, (prevProps, nextProps) => {
  return prevProps.cves === nextProps.cves &&
         prevProps.isLoading === nextProps.isLoading;
});

// 2. Add useCallback for event handlers
const handleRowClick = useCallback((cve: CVE) => {
  setSelectedCVE(cve);
  setShowDetail(true);
}, []);

// 3. Use useMemo for expensive computations
const filteredCVEs = useMemo(() => {
  return cves.filter(cve =>
    cve.id.toLowerCase().includes(searchQuery.toLowerCase())
  );
}, [cves, searchQuery]);

// 4. Add virtualization for large tables
// Update website/components/cve-table.tsx to use @tanstack/react-virtual
import { useVirtualizer } from '@tanstack/react-virtual';

const virtualizer = useVirtualizer({
  count: cves.length,
  getScrollElement: () => parentRef.current,
  estimateSize: () => 48,
  overscan: 10,
});

// 5. Optimize dynamic imports
// Combine duplicate loading skeletons
const CVETable = dynamic(() => import('@/components/cve-table'), {
  ssr: false,
  loading: () => <TableSkeleton />,
});

// 6. Add image optimization
// Use next/image for static images
```

**Success Criteria**:
- Large components use React.memo with proper comparison
- Event handlers use useCallback
- Expensive computations use useMemo
- Tables use virtualization for >100 items
- Dynamic imports have consistent loading states
- Lighthouse performance score >90
- Bundle size reduced by at least 20%

---

### Task 16: Component Documentation
**ID**: UX-016
**Priority**: 3 (Nice to Have)
**Est LoC**: 500
**Package**: website/components

**Description**: Add JSDoc comments and create component documentation.

**Closure Details**:
```typescript
// Add JSDoc to all components:
/**
 * CVETable Component
 *
 * Displays a paginated table of CVE records with search and filtering capabilities.
 *
 * @component
 * @example
 * ```tsx
 * <CVETable
 *   cves={cveList}
 *   total={1000}
 *   page={0}
 *   pageSize={10}
 *   onPageChange={setPage}
 *   onPageSizeChange={setPageSize}
 * />
 * ```
 *
 * @param {CVE[]} cves - Array of CVE records to display
 * @param {number} total - Total number of CVE records
 * @param {number} page - Current page index (0-based)
 * @param {number} pageSize - Number of items per page
 * @param {(page: number) => void} onPageChange - Callback when page changes
 * @param {(size: number) => void} onPageSizeChange - Callback when page size changes
 * @param {boolean} isLoading - Loading state indicator
 * @returns {JSX.Element} Rendered table component
 */
export const CVETable = ({ cves, total, page, pageSize, onPageChange, onPageSizeChange, isLoading }: CVETableProps) => {
  // ...
};

// Create website/components/README.md with:
# v2e Components

## UI Components

### Button
Primary action button with variants and sizes.

### Card
Container component with elevation and interactive states.

## Feature Components

### CVETable
Paginated table for CVE records.

### SessionControl
Job session management controls.

# Usage Examples
# ...

// Add Storybook-style examples in docs/ directory
```

**Success Criteria**:
- All 55 components have JSDoc comments
- Component props are fully typed
- Examples provided for each component
- Component README created
- Documentation is searchable
- Examples are copy-paste ready

---

## Task Priorities

### Priority 1 (Critical) - Block UX improvements
- UX-001: Design System Foundation
- UX-002: Modern Color Palette
- UX-003: Modern Typography System
- UX-005: Button System Standardization
- UX-008: Navigation Refactoring
- UX-010: Error Handling and Toasts
- UX-011: Accessibility Improvements

### Priority 2 (Important) - Enhance UX
- UX-004: Icon System Refinement
- UX-006: Form Elements Design
- UX-007: Card Component Enhancement
- UX-009: Empty States and Loading States
- UX-012: Dark Mode Enhancement
- UX-014: Responsive Design Improvements
- UX-015: Performance Optimization

### Priority 3 (Nice to Have) - Polish
- UX-013: Onboarding and Help System
- UX-016: Component Documentation

## Execution Notes

1. **Tasks are independent** - Each task can be completed by a separate agent
2. **Follow existing patterns** - Use shadcn/ui, Tailwind CSS, and Next.js conventions
3. **Maintain compatibility** - Don't break existing functionality
4. **Test thoroughly** - Verify changes work in both light and dark modes
5. **Commit frequently** - Make small, incremental commits
6. **Document changes** - Update relevant documentation after each task
7. **Use build.sh** - Run tests with `./build.sh -t` after changes

## Success Metrics

- Design system consistency achieved
- Accessibility compliance (WCAG 2.1 AA)
- Performance score >90 (Lighthouse)
- User testing shows improved satisfaction
- Reduced navigation complexity (11 → 8 primary items)
- Component reusability increased
- Dark mode contrast verified
