'use client';

import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import type { VirtualItem } from '@tanstack/virtual-core';
import {
  ChevronRight,
  ChevronDown,
  Search,
  Loader2,
  Shield,
  AlertCircle,
} from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  loadD3FENDOntology,
  buildTreeNodes,
  getOntologyLoadState,
} from '@/lib/glc/d3fend/loader';
import type {
  D3FENDClassTreeNode,
  D3FENDOntologyClass,
  ClassBrowserOptions,
  OntologyLoadState,
} from '@/lib/glc/d3fend/types';

interface ClassBrowserProps extends ClassBrowserOptions {
  /** Additional CSS class name */
  className?: string;
  /** Height of the browser (default: 400px) */
  height?: number | string;
}

export function ClassBrowser({
  className = '',
  height = 400,
  searchQuery: externalSearchQuery,
  initiallyExpanded = [],
  maxDepth = -1,
  onSelect,
}: ClassBrowserProps) {
  const [loadState, setLoadState] = useState<OntologyLoadState>('idle');
  const [searchQuery, setSearchQuery] = useState(externalSearchQuery || '');
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set(initiallyExpanded));
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const parentRef = useRef<HTMLDivElement>(null);

  // Load ontology on mount
  useEffect(() => {
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
  }, []);

  // Build tree nodes from loaded data
  const treeNodes = useMemo(() => {
    if (loadState !== 'loaded') return [];
    return buildTreeNodes(undefined, expandedIds, searchQuery, maxDepth);
  }, [loadState, expandedIds, searchQuery, maxDepth]);

  // Virtualizer for efficient rendering
  const virtualizer = useVirtualizer({
    count: treeNodes.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 36, // Approximate row height
    overscan: 10,
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

  // Handle node selection
  const handleSelect = useCallback(
    (node: D3FENDClassTreeNode) => {
      setSelectedId(node.id);
      if (onSelect) {
        // Get full class data
        const cls: D3FENDOntologyClass = {
          id: node.id,
          label: node.label,
          description: node.description,
          parent: node.parentId,
          children: node.childIds,
        };
        onSelect(node.id, cls);
      }
    },
    [onSelect]
  );

  // Handle search input
  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  }, []);

  // Render loading state
  if (loadState === 'idle' || loadState === 'loading') {
    return (
      <div className={`flex flex-col ${className}`} style={{ height }}>
        <div className="p-2 border-b">
          <Skeleton className="h-9 w-full" />
        </div>
        <div className="flex-1 flex items-center justify-center">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Loader2 className="w-4 h-4 animate-spin" />
            <span>Loading D3FEND ontology...</span>
          </div>
        </div>
      </div>
    );
  }

  // Render error state
  if (loadState === 'error') {
    return (
      <div className={`flex flex-col ${className}`} style={{ height }}>
        <div className="p-2 border-b">
          <Input
            placeholder="Search classes..."
            disabled
            className="bg-muted"
          />
        </div>
        <div className="flex-1 flex items-center justify-center">
          <div className="flex items-center gap-2 text-destructive">
            <AlertCircle className="w-4 h-4" />
            <span>Failed to load ontology</span>
          </div>
        </div>
      </div>
    );
  }

  // Render virtualized tree
  return (
    <div className={`flex flex-col ${className}`} style={{ height }}>
      {/* Search bar */}
      <div className="p-2 border-b shrink-0">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search D3FEND classes..."
            value={searchQuery}
            onChange={handleSearchChange}
            className="pl-8"
          />
        </div>
      </div>

      {/* Virtualized tree */}
      <div ref={parentRef} className="flex-1 overflow-auto">
        {treeNodes.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-muted-foreground">
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

              const isSelected = selectedId === node.id;
              const indentStyle = { paddingLeft: `${node.depth * 20 + 8}px` };

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
                    className={`flex items-center gap-1 py-1.5 px-2 cursor-pointer hover:bg-accent/50 transition-colors ${
                      isSelected ? 'bg-accent text-accent-foreground' : ''
                    }`}
                    style={indentStyle}
                    onClick={() => handleSelect(node)}
                    role="treeitem"
                    aria-selected={isSelected}
                    aria-expanded={node.isExpanded}
                  >
                    {/* Expand/collapse button */}
                    {node.hasChildren ? (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-5 w-5 shrink-0"
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
                      <span className="w-5" /> // Spacer for alignment
                    )}

                    {/* Icon */}
                    <Shield className="w-4 h-4 shrink-0 text-primary/70" />

                    {/* Label */}
                    <span className="truncate text-sm">{node.label}</span>

                    {/* ID badge */}
                    <span className="text-xs text-muted-foreground ml-auto shrink-0">
                      {node.id.replace('d3f:', '')}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

export default ClassBrowser;
