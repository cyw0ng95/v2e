'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import type { VirtualItem } from '@tanstack/virtual-core';
import {
  ChevronRight,
  ChevronDown,
  Search,
  Loader2,
  Shield,
  Check,
  X,
} from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import {
  loadD3FENDOntology,
  buildTreeNodes,
  getOntologyLoadState,
  getClassById,
} from '@/lib/glc/d3fend/loader';
import type {
  D3FENDClassTreeNode,
  D3FENDOntologyClass,
  ClassPickerOptions,
  OntologyLoadState,
} from '@/lib/glc/d3fend/types';

interface ClassPickerProps extends ClassPickerOptions {
  /** Additional CSS class name */
  className?: string;
  /** Height of the dropdown (default: 300px) */
  dropdownHeight?: number;
  /** Width of the dropdown (default: 320px) */
  dropdownWidth?: number;
}

export function ClassPicker({
  className = '',
  dropdownHeight = 300,
  dropdownWidth = 320,
  value,
  placeholder = 'Select D3FEND class...',
  disabled = false,
  rootClassId,
  onChange,
}: ClassPickerProps) {
  const [open, setOpen] = useState(false);
  const [loadState, setLoadState] = useState<OntologyLoadState>('idle');
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [internalValue, setInternalValue] = useState<string | null>(value || null);

  const parentRef = useRef<HTMLDivElement>(null);

  // Sync internal value with external value
  useEffect(() => {
    setInternalValue(value || null);
  }, [value]);

  // Load ontology on first open
  useEffect(() => {
    if (!open || loadState !== 'idle') return;

    const loadOntology = async () => {
      setLoadState(getOntologyLoadState());
      try {
        await loadD3FENDOntology();
        setLoadState('loaded');
      } catch (error) {
        console.error('Failed to load D3FEND ontology:', error);
        setLoadState('error');
      }
    };

    loadOntology();
  }, [open, loadState]);

  // Get selected class data
  const selectedClass = internalValue ? getClassById(internalValue) : null;

  // Build tree nodes
  const rootIds = rootClassId ? [rootClassId] : undefined;
  const treeNodes =
    loadState === 'loaded' ? buildTreeNodes(rootIds, expandedIds, searchQuery, -1) : [];

  // Virtualizer
  const virtualizer = useVirtualizer({
    count: treeNodes.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 32,
    overscan: 8,
  });

  // Toggle node expansion
  const toggleNode = useCallback((nodeId: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  }, []);

  // Handle selection
  const handleSelect = useCallback(
    (node: D3FENDClassTreeNode) => {
      setInternalValue(node.id);
      setOpen(false);
      setSearchQuery('');
      setExpandedIds(new Set());

      if (onChange) {
        const cls: D3FENDOntologyClass = {
          id: node.id,
          label: node.label,
          description: node.description,
          parent: node.parentId,
          children: node.childIds,
        };
        onChange(node.id, cls);
      }
    },
    [onChange]
  );

  // Clear selection
  const handleClear = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      setInternalValue(null);
      if (onChange) {
        onChange(null, null);
      }
    },
    [onChange]
  );

  // Handle search
  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  }, []);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={`w-full justify-between font-normal ${className}`}
          style={{ minWidth: dropdownWidth }}
        >
          {selectedClass ? (
            <div className="flex items-center gap-2 truncate">
              <Shield className="w-4 h-4 shrink-0 text-primary/70" />
              <span className="truncate">{selectedClass.label}</span>
              <Badge variant="secondary" className="text-xs shrink-0">
                {selectedClass.id.replace('d3f:', '')}
              </Badge>
            </div>
          ) : (
            <span className="text-muted-foreground">{placeholder}</span>
          )}
          {internalValue && !disabled ? (
            <X
              className="ml-2 h-4 w-4 shrink-0 opacity-50 hover:opacity-100"
              onClick={handleClear}
            />
          ) : null}
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="p-0 w-auto"
        style={{ width: dropdownWidth }}
        align="start"
      >
        {/* Search bar */}
        <div className="p-2 border-b">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            {loadState === 'loading' ? (
              <Loader2 className="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 animate-spin text-muted-foreground" />
            ) : null}
            <Input
              placeholder="Search classes..."
              value={searchQuery}
              onChange={handleSearchChange}
              className="pl-8"
              autoFocus
            />
          </div>
        </div>

        {/* Virtualized tree */}
        <div
          ref={parentRef}
          className="overflow-auto"
          style={{ height: dropdownHeight }}
        >
          {loadState === 'loading' ? (
            <div className="p-4 space-y-2">
              <Skeleton className="h-6 w-3/4" />
              <Skeleton className="h-6 w-2/3 ml-4" />
              <Skeleton className="h-6 w-1/2 ml-8" />
            </div>
          ) : treeNodes.length === 0 ? (
            <div className="flex items-center justify-center h-24 text-muted-foreground text-sm">
              {searchQuery ? 'No classes found' : 'No classes available'}
            </div>
          ) : (
            <div
              style={{
                height: `${virtualizer.getTotalSize()}px`,
                width: '100%',
                position: 'relative',
              }}
            >
              {virtualizer.getVirtualItems().map((virtualItem: VirtualItem) => {
                const node = treeNodes[virtualItem.index];
                if (!node) return null;

                const isSelected = internalValue === node.id;
                const indentStyle = { paddingLeft: `${node.depth * 16 + 8}px` };

                return (
                  <div
                    key={virtualItem.key}
                    data-index={virtualItem.index}
                    ref={virtualizer.measureElement}
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      width: '100%',
                      transform: `translateY(${virtualItem.start}px)`,
                    }}
                  >
                    <div
                      className={`flex items-center gap-1 py-1 px-2 cursor-pointer hover:bg-accent/50 transition-colors ${
                        isSelected ? 'bg-accent/80' : ''
                      }`}
                      style={indentStyle}
                      onClick={() => handleSelect(node)}
                    >
                      {/* Expand/collapse */}
                      {node.hasChildren ? (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-4 w-4 shrink-0"
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleNode(node.id);
                          }}
                        >
                          {node.isExpanded ? (
                            <ChevronDown className="w-3 h-3" />
                          ) : (
                            <ChevronRight className="w-3 h-3" />
                          )}
                        </Button>
                      ) : (
                        <span className="w-4" />
                      )}

                      {/* Icon */}
                      <Shield className="w-3.5 h-3.5 shrink-0 text-primary/70" />

                      {/* Label */}
                      <span className="truncate text-sm flex-1">{node.label}</span>

                      {/* Selected indicator */}
                      {isSelected && <Check className="w-3.5 h-3.5 text-primary" />}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}

export default ClassPicker;
