'use client';

import { memo, useCallback } from 'react';
import { Handle, Position, useReactFlow } from '@xyflow/react';
import * as Icons from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useGLCStore } from '@/lib/glc/store';
import { NodeContextMenu } from '@/components/glc/context-menu/canvas-context-menu';
import type { Node, Edge } from '@xyflow/react';

interface NodeData {
  label: string;
  typeId: string;
  color?: string;
  icon?: string;
  d3fendClass?: string;
  [key: string]: unknown;
}

interface DynamicNodeProps {
  data: NodeData;
  selected?: boolean;
  id: string;
}

// Dynamic icon component
function DynamicIcon({ name, className, style }: { name?: string; className?: string; style?: React.CSSProperties }) {
  if (!name) return null;
  const IconComponent = (Icons as unknown as Record<string, LucideIcon>)[name];
  return IconComponent ? <IconComponent className={className} style={style} /> : null;
}

export const DynamicNode = memo(function DynamicNode({ data, selected, id }: DynamicNodeProps) {
  const currentPreset = useGLCStore((state) => state.currentPreset);
  const { addNodes, addEdges, getNode, getNodes, getEdges } = useReactFlow();

  if (!currentPreset || !data) return null;

  const nodeType = currentPreset.nodeTypes.find((t) => t.id === data.typeId);
  const theme = currentPreset.theme;

  const backgroundColor = data.color || nodeType?.backgroundColor || theme.surface;
  const borderColor = selected ? theme.accent : (data.color || nodeType?.borderColor || theme.border);
  const textColor = theme.text;

  // Handle duplicate
  const handleDuplicate = useCallback((newNode: Node) => {
    addNodes([newNode]);
  }, [addNodes]);

  // Handle create edge from inference
  const handleCreateEdge = useCallback((sourceId: string, targetId: string, relationshipId: string) => {
    const newEdge: Edge = {
      id: `edge-${sourceId}-${targetId}-${Date.now()}`,
      source: sourceId,
      target: targetId,
      type: 'glc',
      data: { relationshipId },
    };
    addEdges([newEdge]);
  }, [addEdges]);

  // Get current node and related data
  const currentNode = getNode(id);
  const nodes = getNodes();
  const edges = getEdges();

  const nodeContent = (
    <div
      className="px-4 py-2 rounded-lg border-2 shadow-lg min-w-[120px] max-w-[200px] transition-all duration-150"
      style={{
        backgroundColor,
        borderColor,
        color: textColor,
      }}
    >
      {/* Input Handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="!w-3 !h-3 !bg-surface !border-2"
        style={{ borderColor: theme.border }}
      />

      {/* Node Content */}
      <div className="flex items-center gap-2">
        {data.icon && (
          <DynamicIcon name={data.icon} className="w-4 h-4 flex-shrink-0 opacity-80" />
        )}
        <span className="text-sm font-medium truncate">{data.label}</span>
      </div>

      {/* D3FEND class indicator */}
      {data.d3fendClass && (
        <div className="text-xs opacity-60 mt-1 truncate">{data.d3fendClass}</div>
      )}

      {/* Output Handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!w-3 !h-3 !bg-surface !border-2"
        style={{ borderColor: theme.border }}
      />
    </div>
  );

  // Wrap with context menu
  return (
    <NodeContextMenu
      nodeId={id}
      node={currentNode}
      nodes={nodes}
      edges={edges}
      onDuplicate={handleDuplicate}
      onCreateEdge={handleCreateEdge}
    >
      {nodeContent}
    </NodeContextMenu>
  );
});
