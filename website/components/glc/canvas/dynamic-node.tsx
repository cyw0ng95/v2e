'use client';

import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import * as Icons from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useGLCStore } from '@/lib/glc/store';

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
}

// Dynamic icon component
function DynamicIcon({ name, className, style }: { name?: string; className?: string; style?: React.CSSProperties }) {
  if (!name) return null;
  const IconComponent = (Icons as unknown as Record<string, LucideIcon>)[name];
  return IconComponent ? <IconComponent className={className} style={style} /> : null;
}

export const DynamicNode = memo(function DynamicNode({ data, selected }: DynamicNodeProps) {
  const currentPreset = useGLCStore((state) => state.currentPreset);

  if (!currentPreset || !data) return null;

  const nodeType = currentPreset.nodeTypes.find((t) => t.id === data.typeId);
  const theme = currentPreset.theme;

  const backgroundColor = data.color || nodeType?.backgroundColor || theme.surface;
  const borderColor = selected ? theme.accent : (data.color || nodeType?.borderColor || theme.border);
  const textColor = theme.text;

  return (
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
});
