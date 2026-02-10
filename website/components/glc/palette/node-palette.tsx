'use client';

import { useState } from 'react';
import * as Icons from 'lucide-react';
import {
  ChevronDown,
  ChevronRight,
  Search,
  X,
  PanelLeftClose,
  PanelLeft,
  GripVertical,
} from 'lucide-react';
import { useGLCStore } from '@/lib/glc/store';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useResponsive, TOUCH_TARGET_SIZE } from '@/lib/glc/responsive';
import type { CanvasPreset, NodeTypeDefinition } from '@/lib/glc/types';

// Dynamic icon component
function DynamicIcon({ name, className, style }: { name?: string; className?: string; style?: React.CSSProperties }) {
  if (!name) return null;
  const IconComponent = (Icons as unknown as Record<string, { displayName?: string } & React.ComponentType<{ className?: string; style?: React.CSSProperties }>>)[name];
  return IconComponent ? <IconComponent className={className} style={style} /> : null;
}

interface NodePaletteProps {
  preset: CanvasPreset;
}

export function NodePalette({ preset }: NodePaletteProps) {
  const [search, setSearch] = useState('');
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(
    new Set(preset.nodeTypes.map((n) => n.category))
  );
  const { nodePaletteOpen, toggleNodePalette } = useGLCStore();
  const { isMobile, isTablet } = useResponsive();

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

  // Responsive sizing
  const paletteWidth = isMobile ? 'w-56' : isTablet ? 'w-60' : 'w-64';
  const collapsedWidth = isMobile ? 'w-14' : 'w-12';
  const itemPadding = isMobile ? 'px-3 py-3' : 'px-3 py-2';
  const touchMinHeight = isMobile ? TOUCH_TARGET_SIZE : undefined;
  const inputHeight = isMobile ? 'h-11' : 'h-9';
  const iconSize = isMobile ? 'w-5 h-5' : 'w-4 h-4';
  const smallIconSize = isMobile ? 'w-4 h-4' : 'w-3.5 h-3.5';
  const nodeIconBoxSize = isMobile ? 'w-10 h-10' : 'w-8 h-8';

  if (!nodePaletteOpen) {
    return (
      <div
        className={`${collapsedWidth} border-r flex flex-col items-center py-4`}
        style={{
          backgroundColor: theme.surface,
          borderColor: theme.border,
        }}
      >
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleNodePalette}
          className="text-textMuted hover:text-text"
          style={{ minWidth: isMobile ? TOUCH_TARGET_SIZE : undefined, minHeight: isMobile ? TOUCH_TARGET_SIZE : undefined }}
        >
          <PanelLeft className={iconSize} />
        </Button>
      </div>
    );
  }

  return (
    <div
      className={`${paletteWidth} border-r flex flex-col`}
      style={{
        backgroundColor: theme.surface,
        borderColor: theme.border,
      }}
    >
      {/* Header */}
      <div
        className="px-4 py-3 border-b flex items-center justify-between"
        style={{ borderColor: theme.border }}
      >
        <h2 className="font-semibold" style={{ color: theme.text }}>
          Node Types
        </h2>
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleNodePalette}
          className="h-8 w-8 text-textMuted hover:text-text"
        >
          <PanelLeftClose className={smallIconSize} />
        </Button>
      </div>

      {/* Search */}
      <div className="px-3 py-2">
        <div className="relative">
          <Search className={`absolute left-2.5 top-1/2 transform -translate-y-1/2 ${smallIconSize} text-textMuted`} />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search nodes..."
            className={`pl-8 pr-8 ${inputHeight}`}
            style={{
              backgroundColor: theme.background,
              borderColor: theme.border,
              color: theme.text,
              fontSize: isMobile ? '16px' : undefined, // Prevent zoom on iOS
            }}
          />
          {search && (
            <button
              onClick={() => setSearch('')}
              className="absolute right-2.5 top-1/2 transform -translate-y-1/2"
              style={{ minWidth: isMobile ? TOUCH_TARGET_SIZE : undefined, minHeight: isMobile ? TOUCH_TARGET_SIZE : undefined }}
            >
              <X className={`${smallIconSize} text-textMuted hover:text-text`} />
            </button>
          )}
        </div>
      </div>

      {/* Node Categories */}
      <div className="flex-1 overflow-y-auto">
        {Object.entries(filteredCategories).map(([category, types]) => (
          <div key={category}>
            {/* Category Header */}
            <button
              onClick={() => toggleCategory(category)}
              className={`w-full px-4 py-2 flex items-center gap-2 hover:bg-background/50 transition-colors`}
              style={{ color: theme.textMuted, minHeight: touchMinHeight }}
            >
              {expandedCategories.has(category) ? (
                <ChevronDown className={iconSize} />
              ) : (
                <ChevronRight className={iconSize} />
              )}
              <span className={`font-medium ${isMobile ? 'text-base' : 'text-sm'}`}>{category}</span>
              <span className={`ml-auto ${isMobile ? 'text-sm' : 'text-xs'}`}>({types.length})</span>
            </button>

            {/* Node Types */}
            {expandedCategories.has(category) && (
              <div className="px-2 pb-2">
                {types.map((nodeType) => (
                  <div
                    key={nodeType.id}
                    draggable
                    onDragStart={(e) => handleDragStart(e, nodeType)}
                    className={`flex items-center gap-3 ${itemPadding} rounded-lg cursor-grab active:cursor-grabbing hover:bg-background/50 transition-colors`}
                    style={{
                      borderLeft: `${isMobile ? '4px' : '3px'} solid ${nodeType.color}`,
                      minHeight: touchMinHeight,
                    }}
                  >
                    {/* Drag handle for mobile */}
                    {isMobile && (
                      <GripVertical className="w-4 h-4 shrink-0" style={{ color: theme.textMuted }} />
                    )}
                    <div
                      className={`${nodeIconBoxSize} rounded-lg flex items-center justify-center`}
                      style={{ backgroundColor: nodeType.color + '20' }}
                    >
                      <DynamicIcon
                        name={nodeType.icon}
                        className={iconSize}
                        style={{ color: nodeType.color } as React.CSSProperties}
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div
                        className={`font-medium truncate ${isMobile ? 'text-sm' : 'text-sm'}`}
                        style={{ color: theme.text }}
                      >
                        {nodeType.label}
                      </div>
                      {nodeType.description && (
                        <div
                          className={`truncate ${isMobile ? 'text-sm' : 'text-xs'}`}
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
  );
}
