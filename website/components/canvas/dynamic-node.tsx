'use client';

import React, { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { useGLCStore } from '@/lib/glc/store';
import { getNodeStyle } from '@/lib/glc/canvas/canvas-config';
import { Shield, Box, File, User, Bug, AlertTriangle, Activity, Terminal, StickyNote } from 'lucide-react';

const iconMap: Record<string, any> = {
  Shield,
  Box,
  File,
  User,
  Bug,
  AlertTriangle,
  Activity,
  Terminal,
  StickyNote,
};

export { iconMap };

export const DynamicNode = memo(({ data, selected }: NodeProps) => {
  const { currentPreset } = useGLCStore() as any;
  
  if (!currentPreset) {
    return null;
  }

  const nodeStyle = getNodeStyle(currentPreset, (data as any).type);
  const nodeType = currentPreset.nodeTypes.find((nt: any) => nt.id === (data as any).type);
  
  if (!nodeType) {
    return null;
  }

  const Icon = iconMap[nodeType.style.icon] || Box;

  return (
    <div
      className={`px-4 py-3 shadow-md rounded-md border-2 min-w-[150px] ${
        selected ? 'ring-2 ring-blue-500 ring-offset-2 ring-offset-slate-900' : ''
      }`}
      style={nodeStyle}
    >
      <Handle type="target" position={Position.Top} className="w-3 h-3" />
      
      <div className="flex items-center gap-2 mb-2">
        <div className="w-8 h-8 rounded-full flex items-center justify-center bg-black/10">
          <Icon className="w-4 h-4" />
        </div>
        <span className="font-medium text-sm truncate" style={{ color: nodeStyle.color }}>
          {(data as any).name || nodeType.name}
        </span>
      </div>

      {nodeType.properties.length > 0 && (
        <div className="space-y-1">
          {nodeType.properties.slice(0, 3).map((prop: any) => {
            const value = data[prop.id];
            if (value === undefined || value === null || value === '') {
              return null;
            }

            return (
              <div key={prop.id} className="flex items-center gap-2 text-xs" style={{ color: nodeStyle.color }}>
                <span className="opacity-60 truncate">{prop.name}:</span>
                <span className="font-medium truncate">{String(value)}</span>
              </div>
            );
          })}
        </div>
      )}

      {(data as any)._ontologyClass && (
        <div className="mt-2 pt-2 border-t border-black/10 text-xs" style={{ color: nodeStyle.color }}>
          <span className="opacity-60">D3FEND:</span>
          <span className="ml-1 font-mono">{(data as any)._ontologyClass}</span>
        </div>
      )}

      <Handle type="source" position={Position.Bottom} className="w-3 h-3" />
    </div>
  );
});

DynamicNode.displayName = 'DynamicNode';

export default DynamicNode;
