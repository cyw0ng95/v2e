'use client';

import React, { memo } from 'react';
import { BaseEdge, EdgeProps, getBezierPath, EdgeLabelRenderer } from '@xyflow/react';
import { useGLCStore } from '../../store';
import { getEdgeStyle } from '../../lib/glc/canvas/canvas-config';

export const DynamicEdge = memo(({ id, source, target, data, selected }: EdgeProps) => {
  const { currentPreset } = useGLCStore();
  
  if (!currentPreset) {
    return null;
  }

  const edgeType = currentPreset.relationshipTypes.find(rt => rt.id === data.type);
  const edgeStyle = edgeType ? getEdgeStyle(currentPreset, data.type) : {};

  const [edgePath, labelX, labelY] = getBezierPath({
    source: source as any,
    target: target as any,
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
            {data.label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
});

DynamicEdge.displayName = 'DynamicEdge';

export default DynamicEdge;
