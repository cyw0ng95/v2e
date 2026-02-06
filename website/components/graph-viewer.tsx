'use client';

import React, { useCallback, useMemo } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  Node,
  Edge,
  BackgroundVariant,
  ControlButton,
  Panel,
  useReactFlow,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { ZoomIn, ZoomOut, Download, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface GraphViewerProps {
  nodes: Node[];
  edges: Edge[];
  onNodesChange?: (changes: any) => void;
  onEdgesChange?: (changes: any) => void;
  onNodeClick?: (node: Node) => void;
  onEdgeClick?: (edge: Edge) => void;
}

const nodeTypeStyles = {
  cve: {
    backgroundColor: '#EF4444',
    borderColor: '#DC2626',
    textColor: '#FFFFFF',
  },
  cwe: {
    backgroundColor: '#F97316',
    borderColor: '#EA580C',
    textColor: '#FFFFFF',
  },
  capec: {
    backgroundColor: '#EAB308',
    borderColor: '#CA8A04',
    textColor: '#FFFFFF',
  },
  attack: {
    backgroundColor: '#3B82F6',
    borderColor: '#2563EB',
    textColor: '#FFFFFF',
  },
  ssg: {
    backgroundColor: '#22C55E',
    borderColor: '#16A34A',
    textColor: '#FFFFFF',
  },
};

const edgeTypeStyles = {
  references: {
    stroke: '#6B7280',
    strokeWidth: 1,
    strokeDasharray: undefined,
    type: 'default' as const,
  },
  related_to: {
    stroke: '#3B82F6',
    strokeWidth: 1.5,
    strokeDasharray: '5,5',
    type: 'default' as const,
  },
  mitigates: {
    stroke: '#22C55E',
    strokeWidth: 1.5,
    strokeDasharray: '2,2',
    type: 'default' as const,
  },
  exploits: {
    stroke: '#EF4444',
    strokeWidth: 3,
    strokeDasharray: undefined,
    type: 'default' as const,
  },
  contains: {
    stroke: '#A855F7',
    strokeWidth: 1,
    strokeDasharray: undefined,
    type: 'default' as const,
  },
};

const extractNodeType = (id: string): string => {
  if (id.startsWith('v2e::nvd::cve::')) return 'cve';
  if (id.startsWith('v2e::mitre::cwe::')) return 'cwe';
  if (id.startsWith('v2e::mitre::capec::')) return 'capec';
  if (id.startsWith('v2e::mitre::attack::')) return 'attack';
  if (id.startsWith('v2e::ssg::')) return 'ssg';
  return 'cve';
};

const extractEdgeType = (edge: Edge): string => {
  if (edge.label && Object.keys(edgeTypeStyles).includes(edge.label as string)) {
    return edge.label as string;
  }
  return 'references';
};

const getNodeLabel = (id: string): string => {
  const parts = id.split('::');
  return parts[parts.length - 1] || id;
};

function CustomNode({ data, selected }: { data: any; selected: boolean }) {
  const nodeType = extractNodeType(data.id);
  const style = nodeTypeStyles[nodeType as keyof typeof nodeTypeStyles] || nodeTypeStyles.cve;

  return (
    <div
      style={{
        padding: '10px 15px',
        borderRadius: '8px',
        backgroundColor: style.backgroundColor,
        borderColor: style.borderColor,
        borderWidth: selected ? '3px' : '2px',
        borderStyle: 'solid',
        color: style.textColor,
        minWidth: '120px',
        textAlign: 'center',
        fontSize: '12px',
        fontWeight: '500',
        boxShadow: selected ? '0 0 0 3px rgba(0,0,0,0.2)' : '0 2px 4px rgba(0,0,0,0.1)',
        cursor: 'pointer',
        transition: 'all 0.2s',
      }}
    >
      <div style={{ fontSize: '14px', marginBottom: '4px' }}>
        {extractNodeType(data.id).toUpperCase()}
      </div>
      <div style={{ fontSize: '11px', opacity: 0.9 }}>
        {getNodeLabel(data.id)}
      </div>
    </div>
  );
}

const nodeTypes = {
  custom: CustomNode,
};

export default function GraphViewer({
  nodes,
  edges,
  onNodesChange,
  onEdgesChange,
  onNodeClick,
  onEdgeClick,
}: GraphViewerProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow();

  const styledNodes = useMemo(() => {
    return nodes.map(node => ({
      ...node,
      type: 'custom',
      data: {
        ...node.data,
        id: node.id,
      },
    }));
  }, [nodes]);

  const styledEdges = useMemo(() => {
    return edges.map(edge => {
      const edgeType = extractEdgeType(edge);
      const style = edgeTypeStyles[edgeType as keyof typeof edgeTypeStyles] || edgeTypeStyles.references;

      return {
        ...edge,
        type: style.type,
        animated: edgeType === 'exploits',
        style: {
          stroke: style.stroke,
          strokeWidth: style.strokeWidth,
          strokeDasharray: style.strokeDasharray,
        },
        label: edgeType,
        labelStyle: {
          fontSize: '10px',
          fontWeight: 'bold',
          fill: style.stroke,
        },
        labelBgStyle: {
          fill: '#FFFFFF',
          fillOpacity: 0.8,
        },
      };
    });
  }, [edges]);

  const handleZoomIn = useCallback(() => {
    zoomIn({ duration: 200 });
  }, [zoomIn]);

  const handleZoomOut = useCallback(() => {
    zoomOut({ duration: 200 });
  }, [zoomOut]);

  const handleFitView = useCallback(() => {
    fitView({ duration: 200, padding: 0.2 });
  }, [fitView]);

  const handleExportImage = useCallback(async () => {
    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify({ nodes, edges }, null, 2));
    const downloadAnchorNode = document.createElement('a');
    downloadAnchorNode.setAttribute('href', dataStr);
    downloadAnchorNode.setAttribute('download', 'graph.json');
    document.body.appendChild(downloadAnchorNode);
    downloadAnchorNode.click();
    downloadAnchorNode.remove();
  }, [nodes, edges]);

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <ReactFlow
        nodes={styledNodes}
        edges={styledEdges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={(event, node) => onNodeClick?.(node)}
        onEdgeClick={(event, edge) => onEdgeClick?.(edge)}
        nodeTypes={nodeTypes}
        fitView
        minZoom={0.1}
        maxZoom={4}
        defaultViewport={{ x: 0, y: 0, zoom: 1 }}
      >
        <Background variant={BackgroundVariant.Dots} gap={16} size={1} color="#64748B" />

        <Panel position="top-right" className="flex gap-2 p-2 bg-background/95 backdrop-blur rounded-lg border shadow-lg">
          <ControlButton onClick={handleZoomIn} title="Zoom In">
            <ZoomIn className="w-4 h-4" />
          </ControlButton>
          <ControlButton onClick={handleZoomOut} title="Zoom Out">
            <ZoomOut className="w-4 h-4" />
          </ControlButton>
          <ControlButton onClick={handleFitView} title="Fit View">
            <RefreshCw className="w-4 h-4" />
          </ControlButton>
          <ControlButton onClick={handleExportImage} title="Export JSON">
            <Download className="w-4 h-4" />
          </ControlButton>
        </Panel>

        <Controls />

        <MiniMap
          nodeColor={(node) => {
            const nodeType = extractNodeType(node.id);
            const style = nodeTypeStyles[nodeType as keyof typeof nodeTypeStyles] || nodeTypeStyles.cve;
            return style.backgroundColor;
          }}
          nodeBorderRadius={2}
        />
      </ReactFlow>
    </div>
  );
}
