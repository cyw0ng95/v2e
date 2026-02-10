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
import { DrawerPalette } from '@/components/glc/responsive/drawer-palette';
import { CanvasToolbar } from '@/components/glc/toolbar/canvas-toolbar';
import { DynamicNode } from '@/components/glc/canvas/dynamic-node';
import { DynamicEdge } from '@/components/glc/canvas/dynamic-edge';
import { NodeDetailsSheet } from '@/components/glc/canvas/node-details-sheet';
import { EdgeDetailsSheet } from '@/components/glc/canvas/edge-details-sheet';
import { CanvasContextMenu, EdgeContextMenu } from '@/components/glc/context-menu/canvas-context-menu';
import { D3FENDContextMenu } from '@/components/glc/context-menu/d3fend-context-menu';
import { InferencePanel } from '@/components/glc/d3fend';
import { useShortcuts, ShortcutsDialog } from '@/lib/glc/shortcuts';
import { ExportDialog } from '@/components/glc/export';
import { ShareDialog } from '@/components/glc/share';
import { useResponsive, TOUCH_TARGET_SIZE } from '@/lib/glc/responsive';

// Register custom node and edge types
const nodeTypes = { glc: DynamicNode as never };
const edgeTypes = { glc: DynamicEdge as never };

export default function GLCCanvasPage() {
  const params = useParams();
  const router = useRouter();
  const presetId = params.presetId as string;
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const { fitView, zoomIn, zoomOut, getEdge } = useReactFlow();
  const { isMobile, isDesktop } = useResponsive();

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

  // Edge context menu state
  const [edgeContextMenu, setEdgeContextMenu] = useState<{
    edgeId: string;
    x: number;
    y: number;
  } | null>(null);

  // D3FEND context menu state
  const [d3fendContextMenu, setD3fendContextMenu] = useState<{
    nodeId: string;
    x: number;
    y: number;
  } | null>(null);

  // Dialog state
  const [showShortcuts, setShowShortcuts] = useState(false);
  const [showExport, setShowExport] = useState(false);
  const [showShare, setShowShare] = useState(false);
  const [showInferences, setShowInferences] = useState(false);

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
    setEdgeContextMenu(null);
    setD3fendContextMenu(null);
  }, []);

  // Handle edge context menu
  const onEdgeContextMenu = useCallback((event: React.MouseEvent, edge: Edge) => {
    event.preventDefault();
    setEdgeContextMenu({
      edgeId: edge.id,
      x: event.clientX,
      y: event.clientY,
    });
  }, []);

  // Handle node context menu for D3FEND inferences
  const onNodeContextMenu = useCallback((event: React.MouseEvent, node: Node) => {
    // Only show D3FEND context menu for D3FEND preset
    if (currentPreset?.meta.id === 'd3fend' && currentPreset?.behavior.enableInference) {
      event.preventDefault();
      setD3fendContextMenu({
        nodeId: node.id,
        x: event.clientX,
        y: event.clientY,
      });
    }
  }, [currentPreset]);

  // Close edge context menu
  const closeEdgeContextMenu = useCallback(() => {
    setEdgeContextMenu(null);
  }, []);

  // Handle edge reverse
  const handleEdgeReverse = useCallback((reversedEdge: Edge) => {
    setEdges((eds) => {
      // Remove the old edge
      const filtered = eds.filter((e) => e.id !== edgeContextMenu?.edgeId);
      // Add the reversed edge
      return [...filtered, reversedEdge];
    });
    setEdgeContextMenu(null);
  }, [edgeContextMenu]);

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
      {/* Node Palette - Desktop/Tablet only, Mobile uses drawer */}
      {!isMobile && <NodePalette preset={currentPreset} />}

      {/* Mobile Drawer Toggle Button */}
      {isMobile && (
        <div
          className="absolute left-4 top-4 z-30"
          style={{ marginTop: '60px' }} // Below toolbar
        >
          <DrawerPalette preset={currentPreset} />
        </div>
      )}

      {/* Canvas */}
      <div
        className={`flex-1 relative ${isMobile ? 'ml-0' : ''}`}
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
            onEdgeContextMenu={onEdgeContextMenu}
            onNodeContextMenu={onNodeContextMenu}
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
            <Controls
              className="!bg-surface !border-border !text-text"
              style={{
                // Larger touch targets on mobile
                ...(isMobile && {
                  '--rf-controls-button-width': `${TOUCH_TARGET_SIZE}px`,
                  '--rf-controls-button-height': `${TOUCH_TARGET_SIZE}px`,
                } as React.CSSProperties),
              }}
            />
            {/* MiniMap only on desktop */}
            {isDesktop && <MiniMap className="!bg-surface !border-border" />}
            <Panel position="top-center">
              <CanvasToolbar
                preset={currentPreset}
                graphName={graph?.metadata.name}
                onShowShortcuts={() => setShowShortcuts(true)}
                onShowExport={() => setShowExport(true)}
                onShowShare={() => setShowShare(true)}
                onShowInferences={() => setShowInferences(true)}
              />
            </Panel>
          </ReactFlow>
        </CanvasContextMenu>

        {/* Edge Context Menu (positioned absolutely) */}
        {edgeContextMenu && (
          <div
            className="fixed z-50"
            style={{ left: edgeContextMenu.x, top: edgeContextMenu.y }}
          >
            <EdgeContextMenu
              edgeId={edgeContextMenu.edgeId}
              edge={getEdge(edgeContextMenu.edgeId)}
              onEdit={() => {
                setSelectedEdge(edgeContextMenu.edgeId);
                closeEdgeContextMenu();
              }}
              onReverse={handleEdgeReverse}
            >
              <div className="w-0 h-0" />
            </EdgeContextMenu>
          </div>
        )}

        {/* D3FEND Context Menu (positioned absolutely) */}
        {d3fendContextMenu && (
          <div
            className="fixed z-50"
            style={{ left: d3fendContextMenu.x, top: d3fendContextMenu.y }}
          >
            <D3FENDContextMenu
              nodeId={d3fendContextMenu.nodeId}
              nodes={nodes}
              position={{ x: d3fendContextMenu.x, y: d3fendContextMenu.y }}
              onClose={() => setD3fendContextMenu(null)}
            />
          </div>
        )}
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

      {/* Export Dialog */}
      <ExportDialog
        open={showExport}
        onOpenChange={setShowExport}
        canvasRef={reactFlowWrapper}
      />

      {/* Share Dialog */}
      <ShareDialog
        open={showShare}
        onOpenChange={setShowShare}
      />

      {/* Inference Panel */}
      {showInferences && (
        <div className="fixed top-20 right-4 z-40">
          <InferencePanel
            nodes={nodes}
            edges={edges}
            isOpen={showInferences}
            onClose={() => setShowInferences(false)}
          />
        </div>
      )}
    </div>
  );
}
