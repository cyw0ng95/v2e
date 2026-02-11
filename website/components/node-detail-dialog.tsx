'use client';

import React, { useState, useCallback } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useNeighbors, useFindPath } from '@/lib/hooks';
import { Network, Search, ArrowRight } from 'lucide-react';
import { Node } from '@xyflow/react';
import { Skeleton } from '@/components/ui/skeleton';

interface NodeDetailDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  node: Node | null;
  onPathHighlight?: (path: string[]) => void;
}

const nodeTypeColors: Record<string, string> = {
  cve: '#EF4444',
  cwe: '#F97316',
  capec: '#EAB308',
  attack: '#3B82F6',
  ssg: '#22C55E',
};

const getNodeType = (id: string): string => {
  if (id.startsWith('v2e::nvd::cve::')) return 'cve';
  if (id.startsWith('v2e::mitre::cwe::')) return 'cwe';
  if (id.startsWith('v2e::mitre::capec::')) return 'capec';
  if (id.startsWith('v2e::mitre::attack::')) return 'attack';
  if (id.startsWith('v2e::ssg::')) return 'ssg';
  return 'cve';
};

const getNodeLabel = (id: string): string => {
  const parts = id.split('::');
  return parts[parts.length - 1] || id;
};

export default function NodeDetailDialog({
  open,
  onOpenChange,
  node,
  onPathHighlight,
}: NodeDetailDialogProps) {
  const [pathTargetUrn, setPathTargetUrn] = useState<string>('');

  const { neighbors: neighborsData, isLoading: neighborsLoading } = useNeighbors(node?.id || '');
  const { path: pathData, isLoading: pathLoading, findPath } = useFindPath();

  const neighbors = neighborsData || [];

  const handleFindPath = useCallback(async () => {
    if (!node || !pathTargetUrn) return;

    await findPath(node.id, pathTargetUrn);

    if (pathData?.path) {
      onPathHighlight?.(pathData.path);
    }
  }, [node, pathTargetUrn, findPath, pathData, onPathHighlight]);

  const handleExpandNeighbors = useCallback(() => {
    if (!node) return;

    const event = new CustomEvent('expandNeighbors', {
      detail: { nodeId: node.id, neighbors },
    });
    window.dispatchEvent(event);
  }, [node, neighbors]);

  if (!node) return null;

  const nodeType = getNodeType(node.id);
  const nodeColor = nodeTypeColors[nodeType] || nodeTypeColors.cve;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <div
              className="w-3 h-3 rounded-full"
              style={{ backgroundColor: nodeColor }}
            />
            {getNodeLabel(node.id)}
          </DialogTitle>
          <DialogDescription>
            URN: {node.id}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Node Type Badge */}
          <div>
            <Badge
              style={{
                backgroundColor: nodeColor,
                color: '#FFFFFF',
                border: 'none',
                padding: '6px 12px',
                fontSize: '14px',
              }}
            >
              {nodeType.toUpperCase()}
            </Badge>
          </div>

          {/* Node Properties */}
          <div className="space-y-3">
            <h3 className="font-semibold flex items-center gap-2">
              <Network className="w-4 h-4" />
              Properties
            </h3>
            {node.data?.properties ? (
              <div className="space-y-2">
                {Object.entries(node.data.properties as Record<string, unknown>).map(([key, value]) => (
                  <div key={key} className="text-sm">
                    <span className="font-medium text-muted-foreground">{key}:</span>{' '}
                    <span className="ml-1">
                      {typeof value === 'string' ? value : JSON.stringify(value)}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-sm text-muted-foreground">No properties available</div>
            )}
          </div>

          {/* Neighbors */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold flex items-center gap-2">
                <Network className="w-4 h-4" />
                Neighbors
                {neighborsLoading && <Skeleton className="w-4 h-4" />}
              </h3>
              <Button onClick={handleExpandNeighbors} variant="outline" size="sm">
                Expand All
              </Button>
            </div>
            {neighborsLoading ? (
              <div className="space-y-2">
                <Skeleton className="h-8 w-full" />
                <Skeleton className="h-8 w-full" />
                <Skeleton className="h-8 w-full" />
              </div>
            ) : neighbors.length > 0 ? (
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {neighbors.map((neighborUrn: string) => {
                  const neighborType = getNodeType(neighborUrn);
                  const neighborColor = nodeTypeColors[neighborType] || nodeTypeColors.cve;
                  return (
                    <div
                      key={neighborUrn}
                      className="flex items-center gap-2 p-2 rounded hover:bg-accent cursor-pointer transition-colors"
                    >
                      <div
                        className="w-2 h-2 rounded-full flex-shrink-0"
                        style={{ backgroundColor: neighborColor }}
                      />
                      <Badge variant="outline" className="mr-2">
                        {neighborType.toUpperCase()}
                      </Badge>
                      <span className="text-sm font-mono flex-1 truncate">
                        {getNodeLabel(neighborUrn)}
                      </span>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="text-sm text-muted-foreground">No neighbors found</div>
            )}
          </div>

          {/* Find Path */}
          <div className="space-y-3">
            <h3 className="font-semibold flex items-center gap-2">
              <Search className="w-4 h-4" />
              Find Path
            </h3>
            <div className="flex gap-2">
              <Input
                placeholder="Target URN..."
                value={pathTargetUrn}
                onChange={(e) => setPathTargetUrn(e.target.value)}
              />
              <Button
                onClick={handleFindPath}
                disabled={pathLoading || !pathTargetUrn}
              >
                {pathLoading ? (
                  'Searching...'
                ) : (
                  <>
                    <ArrowRight className="w-4 h-4 mr-2" />
                    Find
                  </>
                )}
              </Button>
            </div>
            {pathData?.path && (
              <div className="space-y-2 p-3 bg-accent rounded">
                <div className="text-sm font-medium">Path Found ({pathData.length} nodes)</div>
                <div className="text-xs space-y-1 max-h-40 overflow-y-auto">
                  {pathData.path.map((urn: string, index: number) => {
                    const type = getNodeType(urn);
                    const color = nodeTypeColors[type] || nodeTypeColors.cve;
                    return (
                      <div key={urn} className="flex items-center gap-2">
                        <span className="text-muted-foreground">{index + 1}.</span>
                        <Badge
                          style={{
                            backgroundColor: color,
                            color: '#FFFFFF',
                            border: 'none',
                            fontSize: '10px',
                          }}
                        >
                          {type}
                        </Badge>
                        <span className="font-mono truncate">
                          {getNodeLabel(urn)}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
