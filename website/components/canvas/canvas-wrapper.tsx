'use client';

import React, { useCallback, useEffect } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useGLCStore } from '@/lib/glc/store';
import { getCanvasConfig, getCanvasBackground, applyPresetTheme, removePresetTheme } from '@/lib/glc/canvas/canvas-config';
import { GraphErrorBoundary } from '@/lib/glc/errors/error-boundaries';

interface CanvasWrapperProps {
  children?: React.ReactNode;
}

export function CanvasWrapper({ children }: CanvasWrapperProps) {
  const { currentPreset, nodes, edges, setViewport, addNode, addEdge: addStoreEdge } = useGLCStore(
    (state: any) => ({
      currentPreset: state.currentPreset,
      nodes: state.nodes,
      edges: state.edges,
      setViewport: state.setViewport,
      addNode: state.addNode,
      addEdge: state.addEdge,
    })
  );

  const [flowNodes, setFlowNodes, onNodesChange] = useNodesState(nodes);
  const [flowEdges, setFlowEdges, onEdgesChange] = useEdgesState(edges);

  useEffect(() => {
    if (currentPreset) {
      applyPresetTheme(currentPreset);
    }

    return () => {
      removePresetTheme();
    };
  }, [currentPreset]);

  useEffect(() => {
    setFlowNodes(nodes);
  }, [nodes, setFlowNodes]);

  useEffect(() => {
    setFlowEdges(edges);
  }, [edges, setFlowEdges]);

  const onConnect = useCallback(
    (connection: Connection) => {
      if (!connection.source || !connection.target) {
        return;
      }

      const edge = {
        ...connection,
        id: `edge-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        type: 'default',
        data: {},
      } as Edge;

      setFlowEdges((eds) => addEdge(edge, eds));
      addStoreEdge(edge);
    },
    [addStoreEdge]
  );

  const onMoveEnd = useCallback(
    (_: React.MouseEvent, node: Node) => {
      (useGLCStore.getState() as any).updateNode(node.id, {
        position: node.position,
      });
    },
    []
  );

  if (!currentPreset) {
    return (
      <div className="flex items-center justify-center h-screen bg-slate-900 text-white">
        <div className="text-center">
          <div className="text-4xl mb-4">No preset selected</div>
          <div className="text-slate-400">Please select a preset to continue</div>
        </div>
      </div>
    );
  }

  const config = getCanvasConfig(currentPreset);
  const background = getCanvasBackground(currentPreset);

  return (
    <GraphErrorBoundary fallback={<ErrorFallback />}>
      <div className="w-full h-full">
        <ReactFlow
          nodes={flowNodes}
          edges={flowEdges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeDragStop={onMoveEnd}
          defaultViewport={config.defaultViewport}
          minZoom={config.minZoom}
          maxZoom={config.maxZoom}
          snapToGrid={config.snapToGrid}
          snapGrid={config.snapGrid}
          nodesDraggable={config.nodesDraggable}
          nodesConnectable={config.nodesConnectable}
          elementsSelectable={config.elementsSelectable}
          panOnDrag={config.panOnDrag}
          panOnScroll={config.panOnScroll}
          zoomOnScroll={config.zoomOnScroll}
          zoomOnPinch={config.zoomOnPinch}
          zoomOnDoubleClick={config.zoomOnDoubleClick}
          preventScrolling={config.preventScrolling}
          fitView={config.fitViewOnInit}
          deleteKeyCode={config.deleteKeyCode}
          selectionKeyCode={config.selectionKeyCode}
          multiSelectionKeyCode={config.multiSelectionKeyCode}
          className="bg-slate-900"
          style={{ background: currentPreset.styling.backgroundColor }}
        >
          <Background color={currentPreset.styling.gridColor} gap={currentPreset.behavior.gridSize} />
          <Controls />
          <MiniMap 
            nodeColor={currentPreset.styling.primaryColor}
            maskColor="rgba(0, 0, 0, 0.1)"
          />
          {children}
        </ReactFlow>
      </div>
    </GraphErrorBoundary>
  );
}

function ErrorFallback() {
  return (
    <div className="flex items-center justify-center h-screen bg-slate-900 text-white">
      <div className="text-center">
        <div className="text-4xl mb-4">Canvas Error</div>
        <div className="text-slate-400">An error occurred while rendering the canvas</div>
        <button
          onClick={() => window.location.reload()}
          className="mt-4 px-4 py-2 bg-blue-600 rounded hover:bg-blue-700"
        >
          Reload Page
        </button>
      </div>
    </div>
  );
}

export default CanvasWrapper;
