'use client';

import { useEffect, useMemo, useCallback, useState, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  Panel,
  addEdge,
  applyNodeChanges,
  applyEdgeChanges,
  type Node,
  type Edge,
  type OnConnect,
  type NodeChange,
  type EdgeChange,
  useReactFlow,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { useGLCStore } from '@/lib/glc/store';
import { NodePalette } from '@/components/glc/palette/node-palette';
import { CanvasToolbar } from '@/components/glc/toolbar/canvas-toolbar';
import { DynamicNode } from '@/components/glc/canvas/dynamic-node';
import { DynamicEdge } from '@/components/glc/canvas/dynamic-edge';
import { NodeDetailsSheet } from '@/components/glc/canvas/node-details-sheet';
import { EdgeDetailsSheet } from '@/components/glc/canvas/edge-details-sheet';
import { CanvasContextMenu } from '@/components/glc/context-menu/canvas-context-menu';
import { useShortcuts, ShortcutsDialog } from '@/lib/glc/shortcuts';

// Register custom node and edge types
const nodeTypes = { glc: DynamicNode as never };
const edgeTypes = { glc: DynamicEdge as never };

export default function GLCCanvasPage() {
  const params = useParams();
  const router = useRouter();
  const presetId = params.presetId as string;
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const { fitView, zoomIn, zoomOut } = useReactFlow();

  const {
    currentPreset,
    builtInPresets,
    userPresets,
    setCurrentPreset,
    graph,
    setGraph,
    removeNode,
    removeEdge,
  } = useGLCStore();

  // Local React Flow state
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);

  // Detail sheet state
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [selectedEdge, setSelectedEdge] = useState<string | null>(null);

  // Dialog state
  const [showShortcuts, setShowShortcuts] = useState(false);

  // Find and set the preset
  useEffect(() => {
    const allPresets = [...builtInPresets, ...userPresets];
    const preset = allPresets.find((p) => p.meta.id === presetId);

    if (preset) {
      setCurrentPreset(preset);
    } else {
      router.push('/glc');
    }
  }, [presetId, builtInPresets, userPresets, setCurrentPreset, router]);

  // Initialize graph if needed
  useEffect(() => {
    if (currentPreset && !graph) {
      const newGraph = {
        metadata: {
          id: crypto.randomUUID(),
          name: 'Untitled Graph',
          presetId: currentPreset.meta.id,
          tags: [],
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          version: 1,
        },
        nodes: [],
        edges: [],
      };
      setGraph(newGraph);
    }
  }, [currentPreset, graph, setGraph]);

  // Handle node changes
  const onNodesChange = useCallback((changes: NodeChange[]) => {
    setNodes((nds) => applyNodeChanges(changes, nds));

    // Handle deletions
    changes.forEach((change) => {
      if (change.type === 'remove') {
        removeNode(change.id);
      }
    });
  }, [removeNode]);

  // Handle edge changes
  const onEdgesChange = useCallback((changes: EdgeChange[]) => {
    setEdges((eds) => applyEdgeChanges(changes, eds));

    // Handle deletions
    changes.forEach((change) => {
      if (change.type === 'remove') {
        removeEdge(change.id);
      }
    });
  }, [removeEdge]);

  // Handle new connections
  const onConnect: OnConnect = useCallback((connection) => {
    setEdges((eds) =>
      addEdge(
        {
          ...connection,
          type: 'glc',
          data: { relationshipId: 'connects' },
        },
        eds
      )
    );
  }, []);

  // Handle node click - open details sheet
  const onNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
    setSelectedNode(node.id);
    setSelectedEdge(null);
  }, []);

  // Handle edge click - open details sheet
  const onEdgeClick = useCallback((event: React.MouseEvent, edge: Edge) => {
    setSelectedEdge(edge.id);
    setSelectedNode(null);
  }, []);

  // Handle canvas click - close details
  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
    setSelectedEdge(null);
  }, []);

  // Handle dropping nodes from palette
  const handleDrop = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    const typeData = event.dataTransfer.getData('application/glc-node');
    if (!typeData || !currentPreset || !reactFlowWrapper.current) return;

    const nodeType = JSON.parse(typeData);
    const reactFlowBounds = reactFlowWrapper.current.getBoundingClientRect();
    const position = {
      x: event.clientX - reactFlowBounds.left,
      y: event.clientY - reactFlowBounds.top,
    };

    const newNode: Node = {
      id: crypto.randomUUID(),
      type: 'glc',
      position,
      data: {
        label: nodeType.label,
        typeId: nodeType.id,
        properties: [],
        references: [],
        color: nodeType.color,
        icon: nodeType.icon,
        d3fendClass: nodeType.d3fendClass,
      },
    };

    setNodes((nds) => [...nds, newNode]);
  }, [currentPreset]);

  const handleDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  // Keyboard shortcuts
  useShortcuts({
    onDelete: () => {
      // React Flow handles this with deleteKeyCode
    },
    onFitView: () => {
      fitView({ padding: 0.2 });
    },
    onZoomIn: () => {
      zoomIn();
    },
    onZoomOut: () => {
      zoomOut();
    },
    onToggleHelp: () => {
      setShowShortcuts(true);
    },
  });

  // Theme colors for React Flow
  const flowStyle = useMemo(() => {
    if (!currentPreset) return {};
    return {
      '--rf-background-color': currentPreset.theme.background,
      '--rf-node-background': currentPreset.theme.surface,
      '--rf-node-border-color': currentPreset.theme.border,
      '--rf-node-text': currentPreset.theme.text,
      '--rf-edge-color': currentPreset.theme.border,
    } as React.CSSProperties;
  }, [currentPreset]);

  if (!currentPreset) {
    return (
      <div className="flex items-center justify-center h-screen bg-background">
        <div className="text-textMuted">Loading canvas...</div>
      </div>
    );
  }

  return (
    <div className="h-screen w-full flex">
      {/* Node Palette */}
      <NodePalette preset={currentPreset} />

      {/* Canvas */}
      <div
        className="flex-1 relative"
        ref={reactFlowWrapper}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
      >
        <CanvasContextMenu
          onFitView={() => fitView({ padding: 0.2 })}
        >
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onEdgeClick={onEdgeClick}
            onPaneClick={onPaneClick}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            style={flowStyle}
            fitView
            snapToGrid={currentPreset.behavior.snapToGrid}
            snapGrid={[currentPreset.behavior.gridSize, currentPreset.behavior.gridSize]}
            deleteKeyCode="Delete"
            className="bg-background"
          >
            <Background color={currentPreset.theme.border} gap={currentPreset.behavior.gridSize} />
            <Controls className="!bg-surface !border-border !text-text" />
            <MiniMap className="!bg-surface !border-border" />
            <Panel position="top-center">
              <CanvasToolbar
                preset={currentPreset}
                graphName={graph?.metadata.name}
                onShowShortcuts={() => setShowShortcuts(true)}
              />
            </Panel>
          </ReactFlow>
        </CanvasContextMenu>
      </div>

      {/* Node Details Sheet */}
      {selectedNode && (
        <NodeDetailsSheet
          nodeId={selectedNode}
          onClose={() => setSelectedNode(null)}
        />
      )}

      {/* Edge Details Sheet */}
      {selectedEdge && (
        <EdgeDetailsSheet
          edgeId={selectedEdge}
          onClose={() => setSelectedEdge(null)}
        />
      )}

      {/* Shortcuts Dialog */}
      <ShortcutsDialog
        open={showShortcuts}
        onOpenChange={setShowShortcuts}
      />
    </div>
  );
}
