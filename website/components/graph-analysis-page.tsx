'use client';

import React, { useState, useCallback, useEffect } from 'react';
import GraphViewer from '@/components/graph-viewer';
import GraphControlPanel from '@/components/graph-controls';
import NodeDetailDialog from '@/components/node-detail-dialog';
import { useNodesByType, useFindPath } from '@/lib/hooks';
import { Node, Edge, applyNodeChanges, applyEdgeChanges, NodeChange, EdgeChange } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { AlertCircle, Route, X, MapPin } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

const nodeTypeColors: Record<string, string> = {
  cve: '#EF4444',
  cwe: '#F97316',
  capec: '#EAB308',
  attack: '#3B82F6',
  ssg: '#22C55E',
};

const getNodeLabel = (id: string): string => {
  const parts = id.split('::');
  return parts[parts.length - 1] || id;
};

const generatePosition = (index: number, total: number) => {
  const centerX = 500;
  const centerY = 300;
  const radius = Math.min(400, total * 20);
  const angle = (index / total) * 2 * Math.PI;

  return {
    x: centerX + radius * Math.cos(angle),
    y: centerY + radius * Math.sin(angle),
  };
};

function GraphAnalysisPage() {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  // const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);
  const [highlightedPath, setHighlightedPath] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [pathSourceNode, setPathSourceNode] = useState<Node | null>(null);
  const [isSelectingPath, setIsSelectingPath] = useState(false);

  const { nodes: nodesByTypeData, fetchNodes } = useNodesByType();
  const { path: pathData, isLoading: pathLoading, error: pathError, findPath } = useFindPath();

  const findPathBetweenNodes = useCallback(async (from: string, to: string) => {
    await findPath(from, to);
    if (pathData?.path) {
      setHighlightedPath(pathData.path);
    }
  }, [findPath, pathData]);

  const handleClearPath = useCallback(() => {
    setHighlightedPath([]);
    setPathSourceNode(null);
    setIsSelectingPath(false);
  }, []);

  const handleStartPathSelection = useCallback(() => {
    setPathSourceNode(null);
    setHighlightedPath([]);
    setIsSelectingPath(true);
    setSelectedNode(null);
  }, []);

  const handleNodesChange = useCallback(
    (changes: NodeChange[]) => {
      setNodes((nds) => applyNodeChanges(changes, nds));
    },
    []
  );

  const handleEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      setEdges((eds) => applyEdgeChanges(changes, eds));
    },
    []
  );

  // const handleConnect = useCallback(
  //   (connection: Connection) => {
  //     setEdges((eds) => addEdge({ ...connection, type: 'default' }, eds));
  //   },
  //   []
  // );

  const handleNodeClick = useCallback((node: Node) => {
    if (isSelectingPath) {
      if (pathSourceNode === null) {
        setPathSourceNode(node);
      } else if (pathSourceNode.id !== node.id) {
        findPathBetweenNodes(pathSourceNode.id, node.id);
        setPathSourceNode(null);
        setIsSelectingPath(false);
      }
    } else {
      setSelectedNode(node);
      setHighlightedPath([]);
      setError(null);
    }
  }, [isSelectingPath, pathSourceNode, findPathBetweenNodes]);

  // const handleEdgeClick = useCallback((edge: Edge) => {
  //   setSelectedEdge(edge);
  //   setSelectedNode(null);
  //   setHighlightedPath([]);
  //   setError(null);
  // }, []);

  const handleFilterChange = useCallback(async (types: string[]) => {
    try {
      setError(null);
      const allNodes: Node[] = [];
      const allEdges: Edge[] = [];

      for (const type of types) {
        await fetchNodes(type);
        const responseNodes = nodesByTypeData;
        if (responseNodes) {
          const newNodes: Node[] = responseNodes.map((item: { urn: string; properties: Record<string, unknown> }) => ({
            id: item.urn,
            type: 'custom',
            position: generatePosition(allNodes.length, responseNodes.length),
            data: { ...item, id: item.urn },
          }));
          allNodes.push(...newNodes);
        }
      }

      setNodes(allNodes);
      setEdges(allEdges);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to filter nodes');
    }
  }, [fetchNodes, nodesByTypeData]);

  const handleSearchChange = useCallback((query: string) => {
    if (!query) {
      setHighlightedPath([]);
      return;
    }

    const matchedNodeIds = nodes
      .filter((node) => node.id.toLowerCase().includes(query.toLowerCase()))
      .map((node) => node.id);

    setHighlightedPath(matchedNodeIds);
  }, [nodes]);

  const handlePathHighlight = useCallback((path: string[]) => {
    setHighlightedPath(path);
  }, []);

  const handleExpandNeighbors = useCallback((event: CustomEvent) => {
    const { nodeId, neighbors } = event.detail;

    const newNodes: Node[] = neighbors
      .filter((neighborUrn: string) => !nodes.find((n) => n.id === neighborUrn))
      .map((neighborUrn: string) => ({
        id: neighborUrn,
        type: 'custom',
        position: generatePosition(nodes.length, 5),
        data: { id: neighborUrn, properties: {} },
      }));

    const newEdges: Edge[] = neighbors.map((neighborUrn: string) => ({
      id: `${nodeId}-${neighborUrn}`,
      source: nodeId,
      target: neighborUrn,
      type: 'default',
      label: 'references',
    }));

    setNodes((nds) => [...nds, ...newNodes]);
    setEdges((eds) => [...eds, ...newEdges]);
  }, [nodes]);

  useEffect(() => {
    window.addEventListener('expandNeighbors', handleExpandNeighbors as EventListener);
    return () => {
      window.removeEventListener('expandNeighbors', handleExpandNeighbors as EventListener);
    };
  }, [handleExpandNeighbors]);

  const styledNodes = nodes.map((node) => {
    const isPathNode = highlightedPath.includes(node.id);
    const isSourceNode = pathSourceNode?.id === node.id;

    return {
      ...node,
      style: highlightedPath.length > 0
        ? {
            opacity: isPathNode ? 1 : 0.2,
          }
        : isSourceNode
        ? {
            opacity: 1,
          }
        : undefined,
      className: isPathNode ? 'path-node' : isSourceNode ? 'source-node' : '',
    };
  });

  const styledEdges = edges.map((edge) => {
    const isPathEdge = highlightedPath.includes(edge.source) && highlightedPath.includes(edge.target);
    const pathIndex = highlightedPath.indexOf(edge.source);

    return {
      ...edge,
      style: highlightedPath.length > 0
        ? {
            opacity: isPathEdge ? 1 : 0.2,
            strokeWidth: isPathEdge ? 3 : 1,
          }
        : undefined,
      className: isPathEdge ? 'path-edge' : '',
      animated: isPathEdge && pathIndex !== -1,
    };
  });

  return (
    <div className="h-full flex flex-col md:flex-row gap-4 p-4">
      {/* Left Sidebar - Controls */}
      <div className="w-full md:w-80 shrink-0 overflow-auto flex flex-col gap-4">
        <GraphControlPanel
          onFilterChange={handleFilterChange}
          onSearchChange={handleSearchChange}
        />

        {/* Path Finding Panel */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-sm">
              <Route className="w-4 h-4" />
              Path Finding
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {!isSelectingPath && highlightedPath.length === 0 ? (
              <Button
                onClick={handleStartPathSelection}
                variant="outline"
                className="w-full"
                size="sm"
              >
                <MapPin className="w-4 h-4 mr-2" />
                Find Path
              </Button>
            ) : isSelectingPath && pathSourceNode === null ? (
              <div className="text-sm text-muted-foreground">
                Select source node...
              </div>
            ) : isSelectingPath && pathSourceNode !== null ? (
              <div className="text-sm">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="outline">Source</Badge>
                  <span className="truncate font-mono">
                    {getNodeLabel(pathSourceNode.id)}
                  </span>
                </div>
                <div className="text-muted-foreground">
                  Select target node...
                </div>
              </div>
            ) : null}

            {pathLoading && (
              <div className="text-sm text-muted-foreground animate-pulse">
                Finding shortest path...
              </div>
            )}

            {pathError && (
              <Alert variant="destructive" className="py-2">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription className="text-xs">{pathError}</AlertDescription>
              </Alert>
            )}

            {pathData && pathData.path.length > 0 && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="text-sm font-medium">
                    Path Found ({pathData.length} hops)
                  </div>
                  <Button
                    onClick={handleClearPath}
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2"
                  >
                    <X className="w-3 h-3 mr-1" />
                    Clear
                  </Button>
                </div>
                <div className="space-y-1 max-h-48 overflow-y-auto">
                  {pathData.path.map((urn: string) => {
                    const typeKey = Object.keys(nodeTypeColors).find((t: string) => 
                      urn.includes(t === 'cve' ? 'nvd' : t === 'ssg' ? 'ssg' : t)
                    ) as keyof typeof nodeTypeColors;
                    const type = typeKey || 'cve';
                    const color = nodeTypeColors[type];
                    return (
                      <div
                        key={urn}
                        className="flex items-center gap-2 p-2 rounded hover:bg-accent transition-colors"
                      >
                        <span className="text-muted-foreground text-xs w-4">
                          {pathData.path.indexOf(urn) + 1}.
                        </span>
                        <Badge
                          className="text-xs"
                          style={{
                            backgroundColor: color,
                            border: 'none',
                          }}
                        >
                          {type.toUpperCase()}
                        </Badge>
                        <span className="text-xs font-mono truncate flex-1">
                          {getNodeLabel(urn)}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Right Side - Graph Viewer */}
      <div className="flex-1 min-h-0">
        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="h-full border rounded-lg overflow-hidden bg-background">
          <GraphViewer
            nodes={styledNodes}
            edges={styledEdges}
            onNodesChange={handleNodesChange}
            onEdgesChange={handleEdgesChange}
            onNodeClick={handleNodeClick}
          />
        </div>
      </div>

      {/* Node Detail Dialog */}
      <NodeDetailDialog
        open={selectedNode !== null}
        onOpenChange={(open) => !open && setSelectedNode(null)}
        node={selectedNode}
        onPathHighlight={handlePathHighlight}
      />
    </div>
  );
}

export default GraphAnalysisPage;
