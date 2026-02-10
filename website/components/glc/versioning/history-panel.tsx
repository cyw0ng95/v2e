'use client';

/**
 * GLC Version History Panel
 *
 * Displays version history with diff visualization and restore options.
 */

import React, { useState, useEffect, useCallback } from 'react';
import {
  History,
  RotateCcw,
  ChevronRight,
  ChevronDown,
  Clock,
  Plus,
  Minus,
  Pencil,
  Eye,
  AlertCircle,
  RefreshCw,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
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

interface VersionHistoryPanelProps {
  graphId: string;
  currentVersion: number;
  onRestoreVersion?: (version: number) => void;
  onPreviewVersion?: (version: GLCGraphVersion) => void;
  className?: string;
}

interface VersionWithDiff extends GLCGraphVersion {
  diff?: GraphDiff;
  diffSummary?: string;
}

/**
 * Format timestamp for display
 */
function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * Get icon for change type
 */
function ChangeTypeIcon({ type }: { type: ChangeType }) {
  const iconClass = `h-3 w-3 ${getChangeTypeColor(type)}`;

  switch (type) {
    case 'added':
      return <Plus className={iconClass} />;
    case 'removed':
      return <Minus className={iconClass} />;
    case 'modified':
      return <Pencil className={iconClass} />;
    default:
      return null;
  }
}

/**
 * Version item component
 */
function VersionItem({
  version,
  isCurrent,
  isExpanded,
  onToggle,
  onRestore,
  onPreview,
  isLoading,
}: {
  version: VersionWithDiff;
  isCurrent: boolean;
  isExpanded: boolean;
  onToggle: () => void;
  onRestore: () => void;
  onPreview: () => void;
  isLoading: boolean;
}) {
  const hasChanges = version.diff && version.diff.summary.totalChanges > 0;

  return (
    <div
      className={`border-b border-gray-200 dark:border-gray-700 last:border-b-0 ${
        isCurrent ? 'bg-blue-50/50 dark:bg-blue-900/10' : ''
      }`}
    >
      {/* Version header */}
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-2 p-3 text-left hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
      >
        {isExpanded ? (
          <ChevronDown className="h-4 w-4 text-gray-400" />
        ) : (
          <ChevronRight className="h-4 w-4 text-gray-400" />
        )}

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm">
              v{version.version}
            </span>
            {isCurrent && (
              <Badge variant="secondary" className="text-xs">
                Current
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            <Clock className="h-3 w-3" />
            {formatTimestamp(version.created_at)}
          </div>
        </div>

        {/* Change summary badge */}
        {version.diffSummary && (
          <div className="flex items-center gap-1 text-xs">
            {hasChanges && (
              <Badge
                variant="outline"
                className="bg-gray-50 dark:bg-gray-800 text-gray-600 dark:text-gray-300"
              >
                {version.diffSummary}
              </Badge>
            )}
          </div>
        )}
      </button>

      {/* Expanded details */}
      {isExpanded && version.diff && (
        <div className="overflow-hidden transition-all duration-200">
          <div className="px-3 pb-3 pt-1 ml-6 space-y-2">
            {/* Change details */}
            {hasChanges && (
              <div className="space-y-1.5">
                {/* Node changes */}
                {version.diff.nodes.length > 0 && (
                  <div className="text-xs">
                    <div className="font-medium text-gray-600 dark:text-gray-400 mb-1">
                      Nodes
                    </div>
                    <div className="space-y-1 max-h-32 overflow-y-auto">
                      {version.diff.nodes.slice(0, 5).map((change, idx) => (
                        <div
                          key={`${change.nodeId}-${idx}`}
                          className={`flex items-center gap-1.5 px-2 py-1 rounded ${getChangeTypeBgColor(
                            change.type
                          )}`}
                        >
                          <ChangeTypeIcon type={change.type} />
                          <span className="truncate">{change.label || change.nodeId}</span>
                        </div>
                      ))}
                      {version.diff.nodes.length > 5 && (
                        <div className="text-xs text-gray-500 pl-2">
                          +{version.diff.nodes.length - 5} more
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {/* Edge changes */}
                {version.diff.edges.length > 0 && (
                  <div className="text-xs">
                    <div className="font-medium text-gray-600 dark:text-gray-400 mb-1">
                      Edges
                    </div>
                    <div className="space-y-1 max-h-24 overflow-y-auto">
                      {version.diff.edges.slice(0, 3).map((change, idx) => (
                        <div
                          key={`${change.edgeId}-${idx}`}
                          className={`flex items-center gap-1.5 px-2 py-1 rounded ${getChangeTypeBgColor(
                            change.type
                          )}`}
                        >
                          <ChangeTypeIcon type={change.type} />
                          <span className="truncate">
                            {change.source} to {change.target}
                          </span>
                        </div>
                      ))}
                      {version.diff.edges.length > 3 && (
                        <div className="text-xs text-gray-500 pl-2">
                          +{version.diff.edges.length - 3} more
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Actions */}
            {!isCurrent && (
              <div className="flex gap-2 pt-2">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={(e) => {
                          e.stopPropagation();
                          onPreview();
                        }}
                        disabled={isLoading}
                      >
                        <Eye className="h-3 w-3 mr-1" />
                        Preview
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Preview this version</TooltipContent>
                  </Tooltip>
                </TooltipProvider>

                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={(e) => {
                          e.stopPropagation();
                          onRestore();
                        }}
                        disabled={isLoading}
                      >
                        <RotateCcw className="h-3 w-3 mr-1" />
                        Restore
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Restore to this version</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

/**
 * Version History Panel Component
 */
export function VersionHistoryPanel({
  graphId,
  currentVersion,
  onRestoreVersion,
  onPreviewVersion,
  className = '',
}: VersionHistoryPanelProps) {
  const [versions, setVersions] = useState<VersionWithDiff[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedVersion, setExpandedVersion] = useState<number | null>(null);
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false);
  const [selectedVersion, setSelectedVersion] = useState<number | null>(null);
  const [isRestoring, setIsRestoring] = useState(false);

  // Fetch version history
  const fetchVersions = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await rpcClient.listGLCVersions({
        graph_id: graphId,
        limit: 50,
      });

      if (response.retcode !== 0 || !response.payload) {
        throw new Error(response.message || 'Failed to fetch versions');
      }

      const versionList = response.payload.versions as GLCGraphVersion[];

      // Calculate diffs between consecutive versions
      const versionsWithDiffs: VersionWithDiff[] = await Promise.all(
        versionList.map(async (version, index) => {
          let diff: GraphDiff | undefined;
          let diffSummary: string | undefined;

          // Get the previous version for diff comparison
          if (index < versionList.length - 1) {
            const prevVersion = versionList[index + 1];
            diff = diffGraphs(prevVersion, version);
            diffSummary = formatDiffSummary(diff);
          } else {
            // First version - no previous to compare
            diff = diffGraphs(null, version);
            diffSummary = 'Initial version';
          }

          return {
            ...version,
            diff,
            diffSummary,
          };
        })
      );

      setVersions(versionsWithDiffs);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load version history');
    } finally {
      setIsLoading(false);
    }
  }, [graphId]);

  useEffect(() => {
    fetchVersions();
  }, [fetchVersions, currentVersion]);

  // Handle restore
  const handleRestore = async () => {
    if (selectedVersion === null) return;

    setIsRestoring(true);
    try {
      const response = await rpcClient.restoreGLCVersion({
        graph_id: graphId,
        version: selectedVersion,
      });

      if (response.retcode !== 0 || !response.payload?.success) {
        throw new Error(response.message || 'Failed to restore version');
      }

      onRestoreVersion?.(selectedVersion);
      setRestoreDialogOpen(false);
      setSelectedVersion(null);

      // Refresh version list
      await fetchVersions();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to restore version');
    } finally {
      setIsRestoring(false);
    }
  };

  // Handle preview
  const handlePreview = async (version: number) => {
    try {
      const response = await rpcClient.getGLCVersion({
        graph_id: graphId,
        version,
      });

      if (response.retcode !== 0 || !response.payload) {
        throw new Error(response.message || 'Failed to fetch version');
      }

      onPreviewVersion?.(response.payload.version);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to preview version');
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className={`p-4 ${className}`}>
        <div className="flex items-center gap-2 mb-4">
          <History className="h-4 w-4 text-gray-400" />
          <h3 className="font-medium text-sm">Version History</h3>
        </div>
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="space-y-2">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-3 w-32" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={`p-4 ${className}`}>
        <div className="flex items-center gap-2 mb-4">
          <History className="h-4 w-4 text-gray-400" />
          <h3 className="font-medium text-sm">Version History</h3>
        </div>
        <div className="flex items-center gap-2 text-red-600 dark:text-red-400 text-sm">
          <AlertCircle className="h-4 w-4" />
          <span>{error}</span>
        </div>
        <Button
          size="sm"
          variant="outline"
          onClick={fetchVersions}
          className="mt-2"
        >
          <RefreshCw className="h-3 w-3 mr-1" />
          Retry
        </Button>
      </div>
    );
  }

  // Empty state
  if (versions.length === 0) {
    return (
      <div className={`p-4 ${className}`}>
        <div className="flex items-center gap-2 mb-4">
          <History className="h-4 w-4 text-gray-400" />
          <h3 className="font-medium text-sm">Version History</h3>
        </div>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          No version history available.
        </p>
      </div>
    );
  }

  return (
    <div className={className}>
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center gap-2">
          <History className="h-4 w-4 text-gray-400" />
          <h3 className="font-medium text-sm">Version History</h3>
        </div>
        <Badge variant="secondary" className="text-xs">
          {versions.length} versions
        </Badge>
      </div>

      {/* Version list */}
      <div className="h-[400px] overflow-y-auto">
        <div>
          {versions.map((version) => (
            <VersionItem
              key={version.id}
              version={version}
              isCurrent={version.version === currentVersion}
              isExpanded={expandedVersion === version.version}
              onToggle={() =>
                setExpandedVersion(
                  expandedVersion === version.version ? null : version.version
                )
              }
              onRestore={() => {
                setSelectedVersion(version.version);
                setRestoreDialogOpen(true);
              }}
              onPreview={() => handlePreview(version.version)}
              isLoading={isLoading}
            />
          ))}
        </div>
      </div>

      {/* Restore confirmation dialog */}
      <Dialog open={restoreDialogOpen} onOpenChange={setRestoreDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Restore Version</DialogTitle>
            <DialogDescription>
              Are you sure you want to restore to version {selectedVersion}? This will
              create a new version with the contents of version {selectedVersion}.
              Current changes will be saved to the history.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setRestoreDialogOpen(false)}
              disabled={isRestoring}
            >
              Cancel
            </Button>
            <Button
              onClick={handleRestore}
              disabled={isRestoring}
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
                  Restore
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default VersionHistoryPanel;
