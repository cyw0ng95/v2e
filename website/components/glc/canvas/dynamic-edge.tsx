'use client';

import { memo } from 'react';
import {
  BaseEdge,
  EdgeLabelRenderer,
  getBezierPath,
  Position,
} from '@xyflow/react';
import { useGLCStore } from '@/lib/glc/store';

interface EdgeData {
  relationshipId?: string;
  label?: string;
  [key: string]: unknown;
}

interface DynamicEdgeProps {
  id: string;
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: Position;
  targetPosition: Position;
  data?: EdgeData;
  selected?: boolean;
}

export const DynamicEdge = memo(function DynamicEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  selected,
}: DynamicEdgeProps) {
  const currentPreset = useGLCStore((state) => state.currentPreset);

  if (!currentPreset || !data) return null;

  const relationship = currentPreset.relations.find((r) => r.id === data.relationshipId);
  const theme = currentPreset.theme;

  const style = relationship?.style || {};
  const strokeColor = selected ? theme.accent : (style.strokeColor || theme.border);
  const strokeWidth = style.strokeWidth || 2;
  const strokeStyle = style.strokeStyle || 'solid';

  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const strokeDasharray = strokeStyle === 'dashed' ? '5,5' : strokeStyle === 'dotted' ? '2,2' : undefined;

  const labelText = data.label || relationship?.label;

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          stroke: strokeColor,
          strokeWidth,
          strokeDasharray,
        }}
      />
      {labelText && (
        <EdgeLabelRenderer>
          <div
            className="absolute px-2 py-0.5 rounded text-xs font-medium pointer-events-auto cursor-context-menu transform -translate-x-1/2 -translate-y-1/2"
            style={{
              left: labelX,
              top: labelY,
              backgroundColor: currentPreset.theme.surface,
              color: currentPreset.theme.textMuted,
              border: `1px solid ${currentPreset.theme.border}`,
            }}
          >
            {labelText}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
});
