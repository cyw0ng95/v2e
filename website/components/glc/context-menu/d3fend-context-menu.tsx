/**
 * D3FEND Context Menu
 *
 * Shows inferences and recommendations for D3FEND nodes.
 * Allows adding suggested relationships with one click.
 */

import React, { useCallback, useMemo } from 'react';
import { Position, useReactFlow } from '@xyflow/react';
import type { Node } from '@xyflow/react';
import * as LucideIcons from 'lucide-react';
import { toast } from 'sonner';
import {
  getNodeInferences,
  type InferenceResult,
  type InferenceType,
  type Severity,
} from '@/lib/glc/d3fend';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

// ============================================================================
// Props
// ============================================================================

interface D3FENDContextMenuProps {
  nodeId: string;
  nodes: Node[];
  position: { x: number; y: number };
  onClose: () => void;
}

// ============================================================================
// Helper Components
// ============================================================================

const SeverityBadge: React.FC<{ severity: Severity }> = ({ severity }) => {
  const variants: Record<Severity, { color: string; icon: React.ReactNode }> = {
    critical: { color: 'bg-red-500', icon: <LucideIcons.AlertCircle className="h-3 w-3" /> },
    high: { color: 'bg-orange-500', icon: <LucideIcons.AlertTriangle className="h-3 w-3" /> },
    medium: { color: 'bg-yellow-500', icon: <LucideIcons.Info className="h-3 w-3" /> },
    low: { color: 'bg-blue-500', icon: <LucideIcons.Info className="h-3 w-3" /> },
    info: { color: 'bg-gray-500', icon: <LucideIcons.Info className="h-3 w-3" /> },
  };

  const { color, icon } = variants[severity];

  return (
    <Badge className={`${color} text-white flex items-center gap-1`}>
      {icon}
      <span className="capitalize">{severity}</span>
    </Badge>
  );
};

const InferenceTypeIcon: React.FC<{ type: InferenceType }> = ({ type }) => {
  const icons: Record<InferenceType, React.ReactNode> = {
    sensor: <LucideIcons.Radar className="h-4 w-4" />,
    mitigation: <LucideIcons.Shield className="h-4 w-4" />,
    detection: <LucideIcons.Eye className="h-4 w-4" />,
    weakness: <LucideIcons.Bug className="h-4 w-4" />,
  };

  return icons[type];
};

// ============================================================================
// D3FEND Context Menu Component
// ============================================================================

export const D3FENDContextMenu: React.FC<D3FENDContextMenuProps> = ({
  nodeId,
  nodes,
  position,
  onClose,
}) => {
  const { setNodes, setEdges, getNodes, getEdges, addNodes, addEdges, screenToFlowPosition } = useReactFlow();

  const inferences = useMemo(() => {
    return getNodeInferences(nodes, getEdges(), nodeId);
  }, [nodeId, nodes, getEdges]);

  const getNodeById = useCallback(
    (id: string) => nodes.find(n => n.id === id),
    [nodes]
  );

  const handleAddNode = useCallback(
    async (inference: InferenceResult) => {
      const sourceNode = getNodeById(nodeId);
      if (!sourceNode) return;

      try {
        // Calculate position for new node (offset from source)
        const newNodePosition = screenToFlowPosition({
          x: position.x + 200,
          y: position.y + Math.random() * 100,
        });

        // Create new node based on inference type
        const newNodeId = `node-${Date.now()}`;
        const newNode: Node = {
          id: newNodeId,
          type: 'd3fend-node',
          position: newNodePosition,
          data: {
            label: inference.title,
            d3fendClass: inference.metadata?.d3fendClass,
            properties: inference.metadata?.cweIds
              ? [{ key: 'CWE IDs', value: inference.metadata.cweIds.join(', '), type: 'string' }]
              : [],
          },
        };

        // Create edge between nodes
        const newEdge = {
          id: `edge-${Date.now()}`,
          source: nodeId,
          target: newNodeId,
          type: 'd3fend-edge',
          label: inference.recommendedEdgeType,
          data: {
            relationshipType: inference.recommendedEdgeType,
          },
        };

        addNodes([newNode]);
        addEdges([newEdge]);

        toast.success(`Added ${inference.title} to graph`);
        onClose();
      } catch (error) {
        console.error('Failed to add node:', error);
        toast.error('Failed to add node to graph');
      }
    },
    [nodeId, position, getNodeById, onClose]
  );

  const groupInferencesByType = useCallback(() => {
    const groups: Record<InferenceType, InferenceResult[]> = {
      sensor: [],
      mitigation: [],
      detection: [],
      weakness: [],
    };

    inferences.forEach(inf => {
      groups[inf.type].push(inf);
    });

    return groups;
  }, [inferences]);

  const grouped = groupInferencesByType();

  if (inferences.length === 0) {
    return (
      <Card className="w-80 shadow-lg">
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <LucideIcons.Lightbulb className="h-5 w-5" />
            No Inferences Found
          </CardTitle>
        </CardHeader>
        <CardContent>
          <CardDescription>
            This node has no D3FEND inferences available. Try adding more nodes to the graph.
          </CardDescription>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-96 max-h-[80vh] overflow-y-auto shadow-lg">
      <CardHeader>
        <CardTitle className="text-lg flex items-center justify-between">
          <div className="flex items-center gap-2">
            <LucideIcons.Lightbulb className="h-5 w-5" />
            D3FEND Inferences
          </div>
          <Badge variant="outline">{inferences.length}</Badge>
        </CardTitle>
        <CardDescription>
          Click an inference to add it to the graph
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Mitigations */}
        {grouped.mitigation.length > 0 && (
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm font-medium text-foreground">
              <InferenceTypeIcon type="mitigation" />
              <span>Mitigations ({grouped.mitigation.length})</span>
            </div>
            <div className="space-y-2">
              {grouped.mitigation.map(inference => (
                <div
                  key={inference.id}
                  className="flex items-start gap-2 p-3 rounded-lg border hover:bg-accent/50 transition-colors cursor-pointer"
                  onClick={() => handleAddNode(inference)}
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <InferenceTypeIcon type={inference.type} />
                      <span className="font-medium text-sm">{inference.title}</span>
                      <SeverityBadge severity={inference.severity} />
                    </div>
                    <p className="text-xs text-muted-foreground line-clamp-2">
                      {inference.description}
                    </p>
                    {inference.confidence > 0 && (
                      <div className="flex items-center gap-1 mt-1">
                        <div className="h-1.5 w-16 bg-muted rounded-full overflow-hidden">
                          <div
                            className="h-full bg-primary transition-all"
                            style={{ width: `${inference.confidence}%` }}
                          />
                        </div>
                        <span className="text-xs text-muted-foreground">
                          {inference.confidence}%
                        </span>
                      </div>
                    )}
                  </div>
                  <LucideIcons.Plus className="h-4 w-4 text-muted-foreground mt-1" />
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Weaknesses */}
        {grouped.weakness.length > 0 && (
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm font-medium text-foreground">
              <InferenceTypeIcon type="weakness" />
              <span>Weaknesses ({grouped.weakness.length})</span>
            </div>
            <div className="space-y-2">
              {grouped.weakness.map(inference => (
                <div
                  key={inference.id}
                  className="flex items-start gap-2 p-3 rounded-lg border hover:bg-accent/50 transition-colors cursor-pointer"
                  onClick={() => handleAddNode(inference)}
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <InferenceTypeIcon type={inference.type} />
                      <span className="font-medium text-sm">{inference.title}</span>
                      <SeverityBadge severity={inference.severity} />
                    </div>
                    <p className="text-xs text-muted-foreground line-clamp-2">
                      {inference.description}
                    </p>
                  </div>
                  <LucideIcons.Plus className="h-4 w-4 text-muted-foreground mt-1" />
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Sensors */}
        {grouped.sensor.length > 0 && (
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm font-medium text-foreground">
              <InferenceTypeIcon type="sensor" />
              <span>Active Sensors ({grouped.sensor.length})</span>
            </div>
            <div className="space-y-2">
              {grouped.sensor.map(inference => (
                <div key={inference.id} className="p-3 rounded-lg border bg-muted/30">
                  <div className="flex items-center gap-2 mb-1">
                    <InferenceTypeIcon type={inference.type} />
                    <span className="font-medium text-sm">{inference.title}</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {inference.description}
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Detections */}
        {grouped.detection.length > 0 && (
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm font-medium text-foreground">
              <InferenceTypeIcon type="detection" />
              <span>Detections ({grouped.detection.length})</span>
            </div>
            <div className="space-y-2">
              {grouped.detection.map(inference => (
                <div key={inference.id} className="p-3 rounded-lg border bg-muted/30">
                  <div className="flex items-center gap-2 mb-1">
                    <InferenceTypeIcon type={inference.type} />
                    <span className="font-medium text-sm">{inference.title}</span>
                    <SeverityBadge severity={inference.severity} />
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {inference.description}
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>

      <CardHeader className="pt-0">
        <Button variant="outline" className="w-full" onClick={onClose}>
          Close
        </Button>
      </CardHeader>
    </Card>
  );
};
