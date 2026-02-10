'use client';

import { useState, useCallback } from 'react';
import * as Icons from 'lucide-react';
import {
  ChevronDown,
  ChevronRight,
  Search,
  X,
  Menu,
  GripVertical,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { useResponsive, TOUCH_TARGET_SIZE } from '@/lib/glc/responsive';
import type { CanvasPreset, NodeTypeDefinition } from '@/lib/glc/types';

// Dynamic icon component
function DynamicIcon({ name, className, style }: { name?: string; className?: string; style?: React.CSSProperties }) {
  if (!name) return null;
  const IconComponent = (Icons as unknown as Record<string, { displayName?: string } & React.ComponentType<{ className?: string; style?: React.CSSProperties }>>)[name];
  return IconComponent ? <IconComponent className={className} style={style} /> : null;
}

interface DrawerPaletteProps {
  preset: CanvasPreset;
  trigger?: React.ReactNode;
}

export function DrawerPalette({ preset, trigger }: DrawerPaletteProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(
    new Set(preset.nodeTypes.map((n) => n.category))
  );
  const { isMobile } = useResponsive();

  const theme = preset.theme;

  // Group node types by category
  const categories = preset.nodeTypes.reduce(
    (acc, nodeType) => {
      if (!acc[nodeType.category]) {
        acc[nodeType.category] = [];
      }
      acc[nodeType.category].push(nodeType);
      return acc;
    },
    {} as Record<string, NodeTypeDefinition[]>
  );

  // Filter by search
  const filteredCategories = Object.entries(categories).reduce(
    (acc, [category, types]) => {
      const filtered = types.filter(
        (t) =>
          t.label.toLowerCase().includes(search.toLowerCase()) ||
          t.id.toLowerCase().includes(search.toLowerCase())
      );
      if (filtered.length > 0) {
        acc[category] = filtered;
      }
      return acc;
    },
    {} as Record<string, NodeTypeDefinition[]>
  );

  const toggleCategory = (category: string) => {
    setExpandedCategories((prev) => {
      const next = new Set(prev);
      if (next.has(category)) {
        next.delete(category);
      } else {
        next.add(category);
      }
      return next;
    });
  };

  const handleDragStart = (event: React.DragEvent, nodeType: NodeTypeDefinition) => {
    event.dataTransfer.setData('application/glc-node', JSON.stringify(nodeType));
    event.dataTransfer.effectAllowed = 'move';
  };

  const handleNodeSelect = useCallback((nodeType: NodeTypeDefinition) => {
    // For touch devices, we could emit an event or use a different mechanism
    // For now, just close the drawer
    if (isMobile) {
      // Could dispatch a custom event for the canvas to handle
      const event = new CustomEvent('glc:add-node', { detail: nodeType });
      window.dispatchEvent(event);
      setOpen(false);
    }
  }, [isMobile]);

  const defaultTrigger = (
    <Button
      variant="ghost"
      size="icon"
      className="touch-target"
      style={{ minWidth: TOUCH_TARGET_SIZE, minHeight: TOUCH_TARGET_SIZE }}
    >
      <Menu className="w-5 h-5" style={{ color: theme.text }} />
    </Button>
  );

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        {trigger || defaultTrigger}
      </SheetTrigger>
      <SheetContent
        side="left"
        className="w-[85vw] max-w-[320px] p-0"
        style={{
          backgroundColor: theme.surface,
          borderColor: theme.border,
        }}
      >
        <SheetHeader className="px-4 py-3 border-b" style={{ borderColor: theme.border }}>
          <SheetTitle style={{ color: theme.text }}>Node Types</SheetTitle>
        </SheetHeader>

        {/* Search */}
        <div className="px-4 py-3 border-b" style={{ borderColor: theme.border }}>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5" style={{ color: theme.textMuted }} />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search nodes..."
              className="pl-10 pr-10 h-12"
              style={{
                backgroundColor: theme.background,
                borderColor: theme.border,
                color: theme.text,
                fontSize: '16px', // Prevent zoom on iOS
              }}
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 touch-target"
                style={{ minWidth: TOUCH_TARGET_SIZE, minHeight: TOUCH_TARGET_SIZE }}
              >
                <X className="w-5 h-5" style={{ color: theme.textMuted }} />
              </button>
            )}
          </div>
        </div>

        {/* Node Categories */}
        <div className="flex-1 overflow-y-auto h-[calc(100vh-180px)]">
          <div className="py-2">
            {Object.entries(filteredCategories).map(([category, types]) => (
              <div key={category}>
                {/* Category Header */}
                <button
                  onClick={() => toggleCategory(category)}
                  className="w-full px-4 py-3 flex items-center gap-3 hover:bg-background/50 transition-colors"
                  style={{ color: theme.textMuted, minHeight: TOUCH_TARGET_SIZE }}
                >
                  {expandedCategories.has(category) ? (
                    <ChevronDown className="w-5 h-5" />
                  ) : (
                    <ChevronRight className="w-5 h-5" />
                  )}
                  <span className="font-medium">{category}</span>
                  <span className="ml-auto text-sm">({types.length})</span>
                </button>

                {/* Node Types */}
                {expandedCategories.has(category) && (
                  <div className="px-3 pb-2">
                    {types.map((nodeType) => (
                      <div
                        key={nodeType.id}
                        draggable
                        onDragStart={(e) => handleDragStart(e, nodeType)}
                        onClick={() => handleNodeSelect(nodeType)}
                        className="flex items-center gap-3 px-3 py-3 rounded-lg cursor-grab active:cursor-grabbing hover:bg-background/50 transition-colors"
                        style={{
                          borderLeft: `4px solid ${nodeType.color}`,
                          minHeight: TOUCH_TARGET_SIZE,
                        }}
                      >
                        <GripVertical className="w-5 h-5 shrink-0" style={{ color: theme.textMuted }} />
                        <div
                          className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
                          style={{ backgroundColor: nodeType.color + '20' }}
                        >
                          <DynamicIcon
                            name={nodeType.icon}
                            className="w-5 h-5"
                            style={{ color: nodeType.color } as React.CSSProperties}
                          />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div
                            className="font-medium truncate"
                            style={{ color: theme.text }}
                          >
                            {nodeType.label}
                          </div>
                          {nodeType.description && (
                            <div
                              className="text-sm truncate"
                              style={{ color: theme.textMuted }}
                            >
                              {nodeType.description}
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
