'use client';

import React, { memo } from 'react';
import { BaseEdge, EdgeProps, getBezierPath, EdgeLabelRenderer } from '@xyflow/react';
import { useGLCStore } from '@/lib/glc/store';
import { getEdgeStyle } from '@/lib/glc/canvas/canvas-config';

export const DynamicEdge = memo(({ id, source, target, data, selected }: EdgeProps) => {
  const { currentPreset } = useGLCStore() as any;
  
  if (!currentPreset || !data) {
    return null;
  }

  const edgeType = currentPreset.relationshipTypes.find((rt: any) => rt.id === data.type);
  const edgeStyle = edgeType ? getEdgeStyle(currentPreset, (data as any).type) : {};

  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX: (source as any).x,
    sourceY: (source as any).y,
    sourcePosition: (source as any).position,
    targetX: (target as any).x,
    targetY: (target as any).y,
    targetPosition: (target as any).position,
  });

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        style={edgeStyle}
        className={`transition-all duration-200 ${
          selected ? 'stroke-blue-500 stroke-[3px]' : ''
        }`}
      />
      {edgeType && data.label && (
        <EdgeLabelRenderer>
          <div
            className="px-2 py-1 bg-slate-800 text-white text-xs rounded shadow-md"
            style={{
              transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
              position: 'absolute',
              pointerEvents: 'none',
            }}
          >
            {(data as any).label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
});

DynamicEdge.displayName = 'DynamicEdge';

export default DynamicEdge;
