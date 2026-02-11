'use client';

/**
 * GLC Restore Dialog
 *
 * Confirmation dialog for restoring graph versions with diff preview.
 */

import React, { useState, useEffect, useCallback } from 'react';
import {
  RotateCcw,
  AlertTriangle,
  X,
  RefreshCw,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
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
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { rpcClient } from '@/lib/rpc-client';
import type { GLCGraphVersion } from '@/lib/types';
import {
  diffGraphs,
  formatDiffSummary,
  getChangeTypeColor,
  getChangeTypeBgColor,
  type GraphDiff,
  type ChangeType,
} from '@/lib/glc/versioning';
import type { Graph } from '@/lib/glc/types';

interface RestoreDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  graphId: string;
  targetVersion: number | null;
  currentVersion: number;
  onRestore: (success: boolean, graph?: Graph) => void;
}

/**
 * Format timestamp for display
 */
function formatTimestamp(timestamp: string): string {
  return new Date(timestamp).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * Get icon for change type
 */
function ChangeTypeIcon({ type }: { type: ChangeType }) {
  const iconClass = `h-4 w-4 ${getChangeTypeColor(type)}`;

  switch (type) {
    case 'added':
      return <div className={`${iconClass} font-bold`}>+</div>;
    case 'removed':
      return <div className={`${iconClass} font-bold`}>-</div>;
    case 'modified':
      return <div className={`${iconClass}`}>~</div>;
    default:
      return null;
  }
}

/**
 * Diff preview section
 */
function DiffPreview({
  diff,
  isLoading,
}: {
  diff: GraphDiff | null;
  isLoading: boolean;
}) {
  const [showNodes, setShowNodes] = useState(true);
  const [showEdges, setShowEdges] = useState(true);

  if (isLoading) {
    return (
      <div className="space-y-3 py-4">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-20 w-full" />
      </div>
    );
  }

  if (!diff) {
    return (
      <div className="py-4 text-center text-sm text-gray-500 dark:text-gray-400">
        No changes detected
      </div>
    );
  }

  return (
    <div className="space-y-4 py-2">
      {/* Summary */}
      <div className="flex items-center gap-2 text-sm">
        <span className="font-medium">Changes:</span>
        <Badge variant="outline">{formatDiffSummary(diff)}</Badge>
      </div>

      <Separator />

      {/* Node changes */}
      {diff.nodes.length > 0 && (
        <div className="space-y-2">
          <button
            onClick={() => setShowNodes(!showNodes)}
            className="flex items-center gap-1 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100"
          >
            {showNodes ? (
              <ChevronUp className="h-4 w-4" />
            ) : (
              <ChevronDown className="h-4 w-4" />
            )}
            Nodes ({diff.summary.nodesAdded + diff.summary.nodesRemoved + diff.summary.nodesModified})
          </button>

          {showNodes && (
            <div className="overflow-hidden transition-all duration-200">
              <div className="h-32 overflow-y-auto rounded border border-gray-200 dark:border-gray-700">
                <div className="p-2 space-y-1">
                  {diff.nodes.map((change, idx) => (
                    <div
                      key={`${change.nodeId}-${idx}`}
                      className={`flex items-center gap-2 px-2 py-1.5 rounded text-sm ${getChangeTypeBgColor(
                        change.type
                      )}`}
                    >
                      <ChangeTypeIcon type={change.type} />
                      <span className="font-medium">{change.label}</span>
                      <span className="text-gray-500 dark:text-gray-400 text-xs">
                        ({change.nodeType})
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Edge changes */}
      {diff.edges.length > 0 && (
        <div className="space-y-2">
          <button
            onClick={() => setShowEdges(!showEdges)}
            className="flex items-center gap-1 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100"
          >
            {showEdges ? (
              <ChevronUp className="h-4 w-4" />
            ) : (
              <ChevronDown className="h-4 w-4" />
            )}
            Edges ({diff.summary.edgesAdded + diff.summary.edgesRemoved + diff.summary.edgesModified})
          </button>

          {showEdges && (
            <div className="overflow-hidden transition-all duration-200">
              <div className="h-32 overflow-y-auto rounded border border-gray-200 dark:border-gray-700">
                <div className="p-2 space-y-1">
                  {diff.edges.map((change, idx) => (
                    <div
                      key={`${change.edgeId}-${idx}`}
                      className={`flex items-center gap-2 px-2 py-1.5 rounded text-sm ${getChangeTypeBgColor(
                        change.type
                      )}`}
                    >
                      <ChangeTypeIcon type={change.type} />
                      <span className="font-mono text-xs">
                        {change.source}
                      </span>
                      <span className="text-gray-400">to</span>
                      <span className="font-mono text-xs">
                        {change.target}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

/**
 * Restore Dialog Component
 */
export function RestoreDialog({
  open,
  onOpenChange,
  graphId,
  targetVersion,
  currentVersion,
  onRestore,
}: RestoreDialogProps) {
  const [targetVersionData, setTargetVersionData] =
    useState<GLCGraphVersion | null>(null);
  const [diff, setDiff] = useState<GraphDiff | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch version data
  const fetchVersionData = useCallback(async () => {
    if (!targetVersion || !open) return;

    setIsLoading(true);
    setError(null);

    try {
      // Fetch target version
      const targetResponse = await rpcClient.getGLCVersion({
        graph_id: graphId,
        version: targetVersion,
      });

      if (targetResponse.retcode !== 0 || !targetResponse.payload) {
        throw new Error('Failed to fetch target version');
      }

      setTargetVersionData(targetResponse.payload.version);

      // Fetch current version for comparison
      const currentResponse = await rpcClient.getGLCVersion({
        graph_id: graphId,
        version: currentVersion,
      });

      if (currentResponse.retcode === 0 && currentResponse.payload) {
        // Calculate diff
        const computedDiff = diffGraphs(
          currentResponse.payload.version,
          targetResponse.payload.version
        );
        setDiff(computedDiff);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load version data');
    } finally {
      setIsLoading(false);
    }
  }, [graphId, targetVersion, currentVersion, open]);

  useEffect(() => {
    if (open) {
      fetchVersionData();
    } else {
      // Reset state when dialog closes
      setTargetVersionData(null);
      setDiff(null);
      setError(null);
    }
  }, [open, fetchVersionData]);

  // Handle restore
  const handleRestore = async () => {
    if (!targetVersion) return;

    setIsRestoring(true);
    setError(null);

    try {
      const response = await rpcClient.restoreGLCVersion({
        graph_id: graphId,
        version: targetVersion,
      });

      if (response.retcode !== 0 || !response.payload?.success) {
        throw new Error(response.message || 'Failed to restore version');
      }

      // Parse restored graph
      const restoredGraph = response.payload.graph;
      if (restoredGraph) {
        const graph: Graph = {
          metadata: {
            id: restoredGraph.graph_id,
            name: restoredGraph.name,
            description: restoredGraph.description,
            presetId: restoredGraph.preset_id,
            tags: restoredGraph.tags ? restoredGraph.tags.split(',') : [],
            createdAt: restoredGraph.created_at,
            updatedAt: restoredGraph.updated_at,
            version: restoredGraph.version,
          },
          nodes: JSON.parse(restoredGraph.nodes || '[]'),
          edges: JSON.parse(restoredGraph.edges || '[]'),
          viewport: restoredGraph.viewport
            ? JSON.parse(restoredGraph.viewport)
            : undefined,
        };

        onRestore(true, graph);
      } else {
        onRestore(true);
      }

      onOpenChange(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to restore version');
      onRestore(false);
    } finally {
      setIsRestoring(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <RotateCcw className="h-5 w-5 text-blue-600" />
            Restore Version
          </DialogTitle>
          <DialogDescription>
            Restore your graph to a previous version. This will create a new version
            with the restored content.
          </DialogDescription>
        </DialogHeader>

        {/* Version info */}
        <div className="space-y-4 py-4">
          {/* Target version info */}
          <div className="rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50/50 dark:bg-blue-900/20 p-3">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium text-blue-700 dark:text-blue-300">
                  Target: Version {targetVersion}
                </div>
                {targetVersionData && (
                  <div className="text-xs text-blue-600 dark:text-blue-400 mt-1">
                    {formatTimestamp(targetVersionData.created_at)}
                  </div>
                )}
              </div>
              <Badge variant="outline" className="bg-blue-100 dark:bg-blue-900/50">
                {targetVersionData
                  ? `${JSON.parse(targetVersionData.nodes || '[]').length} nodes`
                  : '...'}
              </Badge>
            </div>
          </div>

          {/* Warning */}
          <div className="flex items-start gap-3 rounded-lg border border-amber-200 dark:border-amber-800 bg-amber-50/50 dark:bg-amber-900/20 p-3">
            <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
            <div className="text-sm text-amber-700 dark:text-amber-300">
              <p className="font-medium">Current version will be preserved</p>
              <p className="mt-1 text-amber-600 dark:text-amber-400">
                Your current version ({currentVersion}) will remain in the history.
                Restoring creates a new version with the target content.
              </p>
            </div>
          </div>

          {/* Diff preview */}
          <div>
            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Changes Preview
            </h4>
            <DiffPreview diff={diff} isLoading={isLoading} />
          </div>

          {/* Error display */}
          {error && (
            <div className="text-sm text-red-600 dark:text-red-400 flex items-center gap-2">
              <X className="h-4 w-4" />
              {error}
            </div>
          )}
        </div>

        <DialogFooter className="gap-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isRestoring}
          >
            Cancel
          </Button>
          <Button
            onClick={handleRestore}
            disabled={isRestoring || isLoading}
            className="bg-blue-600 hover:bg-blue-700"
          >
            {isRestoring ? (
              <>
                <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                Restoring...
              </>
            ) : (
              <>
                <RotateCcw className="h-4 w-4 mr-2" />
                Restore Version {targetVersion}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default RestoreDialog;
