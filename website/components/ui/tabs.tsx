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
    <div className={"inline-flex gap-2 " + (className || "")}>{
      React.Children.map(children, child =>
        React.cloneElement(child, { selected: value === child.props.value, onClick: () => onValueChange(child.props.value) })
      )
    }</div>
  );
}
TabsList.displayName = 'TabsList';

export function TabsTrigger({ value, selected, onClick, children }: any) {
  return (
    <button
      type="button"
      className={
        "px-4 py-1 rounded border-b-2 " +
        (selected ? "border-primary font-bold" : "border-transparent text-muted-foreground hover:border-muted-foreground")
      }
      onClick={onClick}
    >
      {children}
    </button>
  );
}
TabsTrigger.displayName = 'TabsTrigger';

export function TabsContent({ children, className }: any) {
  return <div className={className}>{children}</div>;
}
TabsContent.displayName = 'TabsContent';
