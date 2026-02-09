'use client';

import { useEffect, useMemo, useCallback, useState } from 'react';
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
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { useGLCStore } from '@/lib/glc/store';
import { NodePalette } from '@/components/glc/palette/node-palette';
import { CanvasToolbar } from '@/components/glc/toolbar/canvas-toolbar';
import { DynamicNode } from '@/components/glc/canvas/dynamic-node';
import { DynamicEdge } from '@/components/glc/canvas/dynamic-edge';

// Register custom node and edge types
const nodeTypes = { glc: DynamicNode as never };
const edgeTypes = { glc: DynamicEdge as never };

export default function GLCCanvasPage() {
  const params = useParams();
  const router = useRouter();
  const presetId = params.presetId as string;

  const {
    currentPreset,
    builtInPresets,
    userPresets,
    setCurrentPreset,
    graph,
    setGraph,
  } = useGLCStore();

  // Local React Flow state
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);

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
  }, []);

  // Handle edge changes
  const onEdgesChange = useCallback((changes: EdgeChange[]) => {
    setEdges((eds) => applyEdgeChanges(changes, eds));
  }, []);

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

  // Handle dropping nodes from palette
  const handleDrop = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    const typeData = event.dataTransfer.getData('application/glc-node');
    if (!typeData || !currentPreset) return;

    const nodeType = JSON.parse(typeData);
    const reactFlowBounds = event.currentTarget.getBoundingClientRect();
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
    <div className="h-screen w-full flex" onDrop={handleDrop} onDragOver={handleDragOver}>
      {/* Node Palette */}
      <NodePalette preset={currentPreset} />

      {/* Canvas */}
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
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
            <CanvasToolbar preset={currentPreset} graphName={graph?.metadata.name} />
          </Panel>
        </ReactFlow>
      </div>
    </div>
  );
}
