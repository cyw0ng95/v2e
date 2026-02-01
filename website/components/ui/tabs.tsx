import * as React from "react";

// Create a context to share tab state
interface TabsContextValue {
  value: string;
  onValueChange: (value: string) => void;
}

const TabsContext = React.createContext<TabsContextValue | null>(null);

// Hook to access tabs context
function useTabsContext() {
  const context = React.useContext(TabsContext);
  if (!context) {
    throw new Error('Tabs components must be used within a Tabs provider');
  }
  return context;
}

interface TabsProps {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  children: React.ReactNode;
  className?: string;
}

export function Tabs({ value, defaultValue, onValueChange, children, className }: TabsProps) {
  // Use controlled or uncontrolled state
  const [internalValue, setInternalValue] = React.useState(defaultValue || '');
  
  // Determine the current value (controlled vs uncontrolled)
  const currentValue = value !== undefined ? value : internalValue;
  
  // Handle value change
  const handleValueChange = React.useCallback((newValue: string) => {
    // Update internal state for uncontrolled mode
    if (value === undefined) {
      setInternalValue(newValue);
    }
    // Call external handler
    if (onValueChange) {
      onValueChange(newValue);
    }
  }, [value, onValueChange]);
  
  // Create context value
  const contextValue = React.useMemo(() => ({
    value: currentValue,
    onValueChange: handleValueChange,
  }), [currentValue, handleValueChange]);
  
  return (
    <TabsContext.Provider value={contextValue}>
      <div className={className}>
        {children}
      </div>
    </TabsContext.Provider>
  );
}
Tabs.displayName = 'Tabs';

interface TabsListProps {
  children: React.ReactNode;
  className?: string;
}

export function TabsList({ children, className }: TabsListProps) {
  return (
    <div className={"inline-flex gap-1 p-1 bg-muted rounded-lg " + (className || "")}>
      {children}
    </div>
  );
}
TabsList.displayName = 'TabsList';

interface TabsTriggerProps {
  value: string;
  children: React.ReactNode;
  disabled?: boolean;
  className?: string;
}

export function TabsTrigger({ value, children, disabled, className }: TabsTriggerProps) {
  const { value: selectedValue, onValueChange } = useTabsContext();
  const isSelected = selectedValue === value;
  
  const handleClick = React.useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    if (!disabled) {
      onValueChange(value);
    }
  }, [disabled, onValueChange, value]);
  
  const buttonClassName = React.useMemo(() => {
    const baseClass = "px-4 py-2 rounded-md transition-colors cursor-pointer";
    const stateClass = isSelected 
      ? "bg-background text-foreground shadow-sm" 
      : "text-muted-foreground hover:text-foreground hover:bg-muted/50";
    const disabledClass = disabled ? "opacity-50 cursor-not-allowed" : "";
    return `${baseClass} ${stateClass} ${disabledClass} ${className || ""}`.trim();
  }, [isSelected, disabled, className]);
  
  return (
    <button
      type="button"
      role="tab"
      aria-selected={isSelected}
      disabled={disabled}
      className={buttonClassName}
      onClick={handleClick}
    >
      {children}
    </button>
  );
}
TabsTrigger.displayName = 'TabsTrigger';

interface TabsContentProps {
  value: string;
  children: React.ReactNode;
  className?: string;
}

export function TabsContent({ value, children, className }: TabsContentProps) {
  const { value: selectedValue } = useTabsContext();
  const isSelected = selectedValue === value;
  
  if (!isSelected) {
    return null;
  }
  
  return (
    <div 
      role="tabpanel"
      className={`flex-1 min-h-0 overflow-auto ${className || ""}`}
    >
      {children}
    </div>
  );
}
TabsContent.displayName = 'TabsContent';
