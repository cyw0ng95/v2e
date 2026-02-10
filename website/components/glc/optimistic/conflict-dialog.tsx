'use client';

/**
 * GLC Conflict Resolution Dialog
 *
 * Displays when a version conflict is detected between client and server.
 * Shows a diff of changes and provides options to:
 * - Overwrite server with client changes
 * - Discard client changes and use server version
 * - Cancel and keep editing (manual merge)
 */

import { useState, useMemo, useCallback } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  AlertTriangle,
  ArrowRightLeft,
  CheckCircle2,
  Plus,
  Minus,
  Edit,
  XCircle,
  RefreshCw,
} from 'lucide-react';
import type {
  VersionConflictResult,
  CADNode,
  CADEdge,
} from '@/lib/glc/optimistic';
import { diffNodes, diffEdges, parseGLCGraphData } from '@/lib/glc/optimistic';

// ============================================================================
// Types
// ============================================================================

export type ConflictResolution = 'client' | 'server' | 'cancel';

export interface ConflictDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback when open state changes */
  onOpenChange: (open: boolean) => void;
  /** The conflict information */
  conflict: VersionConflictResult | null;
  /** Current client graph nodes */
  clientNodes: CADNode[];
  /** Current client graph edges */
  clientEdges: CADEdge[];
  /** Callback when user resolves the conflict */
  onResolve: (resolution: ConflictResolution) => void;
  /** Whether a resolution is in progress */
  isResolving?: boolean;
}

// ============================================================================
// Diff Display Components
// ============================================================================

interface DiffItemProps {
  type: 'added' | 'removed' | 'modified';
  label: string;
  description?: string;
}

function DiffItem({ type, label, description }: DiffItemProps) {
  const icons = {
    added: <Plus className="h-4 w-4 text-green-500" />,
    removed: <Minus className="h-4 w-4 text-red-500" />,
    modified: <Edit className="h-4 w-4 text-amber-500" />,
  };

  const colors = {
    added: 'border-green-500/50 bg-green-500/10',
    removed: 'border-red-500/50 bg-red-500/10',
    modified: 'border-amber-500/50 bg-amber-500/10',
  };

  return (
    <div
      className={`flex items-center gap-2 px-3 py-2 rounded border ${colors[type]}`}
    >
      {icons[type]}
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium truncate">{label}</div>
        {description && (
          <div className="text-xs text-muted-foreground truncate">
            {description}
          </div>
        )}
      </div>
    </div>
  );
}

interface DiffSummaryProps {
  title: string;
  added: number;
  removed: number;
  modified: number;
  items: Array<{ id: string; label: string; type: 'added' | 'removed' | 'modified' }>;
}

function DiffSummary({ title, added, removed, modified, items }: DiffSummaryProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <h4 className="text-sm font-medium">{title}</h4>
        <div className="flex gap-1">
          {added > 0 && (
            <Badge variant="outline" className="text-green-500 border-green-500/50">
              +{added}
            </Badge>
          )}
          {removed > 0 && (
            <Badge variant="outline" className="text-red-500 border-red-500/50">
              -{removed}
            </Badge>
          )}
          {modified > 0 && (
            <Badge variant="outline" className="text-amber-500 border-amber-500/50">
              ~{modified}
            </Badge>
          )}
          {added === 0 && removed === 0 && modified === 0 && (
            <Badge variant="outline" className="text-muted-foreground">
              No changes
            </Badge>
          )}
        </div>
      </div>
      <div className="space-y-1 max-h-40 overflow-y-auto">
        {items.length === 0 ? (
          <p className="text-sm text-muted-foreground italic">No differences</p>
        ) : (
          items.map((item) => (
            <DiffItem
              key={`${item.type}-${item.id}`}
              type={item.type}
              label={item.label}
            />
          ))
        )}
      </div>
    </div>
  );
}

// ============================================================================
// Main Component
// ============================================================================

export function ConflictDialog({
  open,
  onOpenChange,
  conflict,
  clientNodes,
  clientEdges,
  onResolve,
  isResolving = false,
}: ConflictDialogProps) {
  // Reset selection when conflict changes (using key pattern on parent)
  const [selectedResolution, setSelectedResolution] = useState<ConflictResolution | null>(null);

  // Parse server data
  const serverData = useMemo(() => {
    if (!conflict?.serverGraph) {
      return { nodes: [] as CADNode[], edges: [] as CADEdge[] };
    }
    return parseGLCGraphData(conflict.serverGraph);
  }, [conflict]);

  // Compute diffs
  const nodeDiff = useMemo(() => {
    return diffNodes(clientNodes, serverData.nodes);
  }, [clientNodes, serverData.nodes]);

  const edgeDiff = useMemo(() => {
    return diffEdges(clientEdges, serverData.edges);
  }, [clientEdges, serverData.edges]);

  // Prepare diff items for display
  const nodeDiffItems = useMemo(() => {
    const items: Array<{ id: string; label: string; type: 'added' | 'removed' | 'modified' }> = [];

    nodeDiff.added.forEach((node) => {
      items.push({ id: node.id, label: node.data.label || node.id, type: 'added' });
    });
    nodeDiff.removed.forEach((node) => {
      items.push({ id: node.id, label: node.data.label || node.id, type: 'removed' });
    });
    nodeDiff.modified.forEach((node) => {
      items.push({ id: node.id, label: node.data.label || node.id, type: 'modified' });
    });

    return items;
  }, [nodeDiff]);

  const edgeDiffItems = useMemo(() => {
    const items: Array<{ id: string; label: string; type: 'added' | 'removed' | 'modified' }> = [];

    edgeDiff.added.forEach((edge) => {
      const sourceNode = clientNodes.find((n) => n.id === edge.source);
      const targetNode = clientNodes.find((n) => n.id === edge.target);
      items.push({
        id: edge.id,
        label: `${sourceNode?.data.label || edge.source} \u2192 ${targetNode?.data.label || edge.target}`,
        type: 'added',
      });
    });
    edgeDiff.removed.forEach((edge) => {
      const sourceNode = serverData.nodes.find((n) => n.id === edge.source);
      const targetNode = serverData.nodes.find((n) => n.id === edge.target);
      items.push({
        id: edge.id,
        label: `${sourceNode?.data.label || edge.source} \u2192 ${targetNode?.data.label || edge.target}`,
        type: 'removed',
      });
    });
    edgeDiff.modified.forEach((edge) => {
      const sourceNode = clientNodes.find((n) => n.id === edge.source);
      const targetNode = clientNodes.find((n) => n.id === edge.target);
      items.push({
        id: edge.id,
        label: `${sourceNode?.data.label || edge.source} \u2192 ${targetNode?.data.label || edge.target}`,
        type: 'modified',
      });
    });

    return items;
  }, [edgeDiff, clientNodes, serverData.nodes]);

  // Handle resolution
  const handleResolve = (resolution: ConflictResolution) => {
    setSelectedResolution(resolution);
    onResolve(resolution);
  };

  // Check if there are any differences (unused but kept for potential future use)
  const _hasChanges = nodeDiffItems.length > 0 || edgeDiffItems.length > 0;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            <DialogTitle>Version Conflict Detected</DialogTitle>
          </div>
          <DialogDescription>
            The graph has been modified on the server since you last loaded it.
            Your version (v{conflict?.clientVersion ?? '?'}) is behind the server version (v{conflict?.serverVersion ?? '?'}).
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Summary */}
          <div className="flex items-center justify-center gap-4 p-4 rounded-lg bg-muted/50">
            <div className="text-center">
              <div className="text-xs text-muted-foreground mb-1">Your Version</div>
              <Badge variant="secondary" className="text-base">
                v{conflict?.clientVersion ?? '?'}
              </Badge>
            </div>
            <ArrowRightLeft className="h-5 w-5 text-muted-foreground" />
            <div className="text-center">
              <div className="text-xs text-muted-foreground mb-1">Server Version</div>
              <Badge variant="default" className="text-base">
                v{conflict?.serverVersion ?? '?'}
              </Badge>
            </div>
          </div>

          {/* Diff tabs */}
          <Tabs defaultValue="nodes" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="nodes">
                Nodes ({nodeDiffItems.length})
              </TabsTrigger>
              <TabsTrigger value="edges">
                Connections ({edgeDiffItems.length})
              </TabsTrigger>
            </TabsList>
            <TabsContent value="nodes" className="mt-4">
              <DiffSummary
                title="Node Changes"
                added={nodeDiff.added.length}
                removed={nodeDiff.removed.length}
                modified={nodeDiff.modified.length}
                items={nodeDiffItems}
              />
            </TabsContent>
            <TabsContent value="edges" className="mt-4">
              <DiffSummary
                title="Connection Changes"
                added={edgeDiff.added.length}
                removed={edgeDiff.removed.length}
                modified={edgeDiff.modified.length}
                items={edgeDiffItems}
              />
            </TabsContent>
          </Tabs>

          {/* Resolution options */}
          <div className="space-y-2 pt-2">
            <h4 className="text-sm font-medium">How would you like to resolve this?</h4>
            <div className="grid gap-2">
              <Button
                variant="outline"
                className="justify-start h-auto py-3"
                onClick={() => handleResolve('client')}
                disabled={isResolving}
              >
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-green-500 mt-0.5" />
                  <div className="text-left">
                    <div className="font-medium">Keep my changes</div>
                    <div className="text-xs text-muted-foreground">
                      Overwrite the server with your local changes. The other person&apos;s changes will be lost.
                    </div>
                  </div>
                </div>
              </Button>

              <Button
                variant="outline"
                className="justify-start h-auto py-3"
                onClick={() => handleResolve('server')}
                disabled={isResolving}
              >
                <div className="flex items-start gap-3">
                  <RefreshCw className="h-5 w-5 text-blue-500 mt-0.5" />
                  <div className="text-left">
                    <div className="font-medium">Use server version</div>
                    <div className="text-xs text-muted-foreground">
                      Discard your local changes and load the server version. Your changes will be lost.
                    </div>
                  </div>
                </div>
              </Button>

              <Button
                variant="outline"
                className="justify-start h-auto py-3"
                onClick={() => handleResolve('cancel')}
                disabled={isResolving}
              >
                <div className="flex items-start gap-3">
                  <XCircle className="h-5 w-5 text-muted-foreground mt-0.5" />
                  <div className="text-left">
                    <div className="font-medium">Cancel</div>
                    <div className="text-xs text-muted-foreground">
                      Close this dialog and continue editing. You can manually merge changes later.
                    </div>
                  </div>
                </div>
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          {isResolving && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <RefreshCw className="h-4 w-4 animate-spin" />
              Resolving...
            </div>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ============================================================================
// Hook for Conflict Dialog
// ============================================================================

export interface UseConflictDialogResult {
  /** Whether the dialog is open */
  isOpen: boolean;
  /** Current conflict data */
  conflict: VersionConflictResult | null;
  /** Whether resolution is in progress */
  isResolving: boolean;
  /** Show the dialog with conflict data */
  showConflict: (conflict: VersionConflictResult) => void;
  /** Close the dialog */
  closeDialog: () => void;
  /** Handle resolution selection */
  handleResolve: (resolution: ConflictResolution, onResolve: (resolution: ConflictResolution) => Promise<void>) => Promise<void>;
}

/**
 * Hook to manage conflict dialog state
 */
export function useConflictDialog(): UseConflictDialogResult {
  const [isOpen, setIsOpen] = useState(false);
  const [conflict, setConflict] = useState<VersionConflictResult | null>(null);
  const [isResolving, setIsResolving] = useState(false);

  const showConflict = useCallback((newConflict: VersionConflictResult) => {
    setConflict(newConflict);
    setIsOpen(true);
  }, []);

  const closeDialog = useCallback(() => {
    setIsOpen(false);
    // Don't clear conflict immediately to allow animation to complete
    setTimeout(() => setConflict(null), 200);
  }, []);

  const handleResolve = useCallback(
    async (resolution: ConflictResolution, onResolve: (resolution: ConflictResolution) => Promise<void>) => {
      setIsResolving(true);
      try {
        await onResolve(resolution);
        closeDialog();
      } finally {
        setIsResolving(false);
      }
    },
    [closeDialog]
  );

  return {
    isOpen,
    conflict,
    isResolving,
    showConflict,
    closeDialog,
    handleResolve,
  };
}
