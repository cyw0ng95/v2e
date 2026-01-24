import * as React from "react";

export function Tabs({ value, onValueChange, children, className }: any) {
  const [tab, setTab] = React.useState(value);
  React.useEffect(() => { setTab(value); }, [value]);
  return (
    <div className={className}>{
      React.Children.map(children, child => {
        if (child.type.displayName === 'TabsList') {
          return React.cloneElement(child, { value: tab, onValueChange: onValueChange || setTab });
        }
        if (child.type.displayName === 'TabsContent') {
          return tab === child.props.value ? child : null;
        }
        return child;
      })
    }</div>
  );
}
Tabs.displayName = 'Tabs';

export function TabsList({ children, value, onValueChange, className }: any) {
  return (
    <div className={"inline-flex gap-1 p-1 bg-muted rounded-lg " + (className || "")}>
      {React.Children.map(children, child =>
        React.cloneElement(child, { selected: value === child.props.value, onClick: () => onValueChange(child.props.value) })
      )}
    </div>
  );
}
TabsList.displayName = 'TabsList';

export function TabsTrigger({ value, selected, onClick, children }: any) {
  return (
    <button
      type="button"
      className={`px-4 py-2 rounded-md transition-colors ${selected 
        ? "bg-background text-foreground shadow-sm" 
        : "text-muted-foreground hover:text-foreground hover:bg-muted/50"}`}
      onClick={onClick}
    >
      {children}
    </button>
  );
}
TabsTrigger.displayName = 'TabsTrigger';

export function TabsContent({ children, className }: any) {
  return <div className={`flex-1 min-h-0 overflow-auto ${className}`}>{children}</div>;
}
TabsContent.displayName = 'TabsContent';
