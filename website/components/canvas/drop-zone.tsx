'use client';

import { useRef, useState } from 'react';
import { useGLCStore } from '../../store';
import { onDragOver, onDrop, isValidDropTarget, calculateDropPosition } from '../../lib/glc/canvas/drag-drop';
import { NodeTypeDefinition } from '../../types';

interface DropZoneProps {
  children: React.ReactNode;
}

export function DropZone({ children }: DropZoneProps) {
  const { addNode, nodes, currentPreset } = useGLCStore();
  const canvasRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [dropPosition, setDropPosition] = useState<{ x: number; y: number } | null>(null);

  const handleDragEnter = (event: React.DragEvent) => {
    if (isValidDropTarget(event)) {
      setIsDragging(true);
    }
  };

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    event.stopPropagation();
    
    if (isDragging && canvasRef.current) {
      const bounds = canvasRef.current.getBoundingClientRect();
      const pos = calculateDropPosition(event, bounds);
      setDropPosition(pos);
    }
    
    onDragOver(event);
  };

  const handleDragLeave = (event: React.DragEvent) => {
    const rect = canvasRef.current?.getBoundingClientRect();
    if (!rect) return;

    const x = event.clientX;
    const y = event.clientY;

    if (
      x < rect.left ||
      x >= rect.right ||
      y < rect.top ||
      y >= rect.bottom
    ) {
      setIsDragging(false);
      setDropPosition(null);
    }
  };

  const handleDrop = (event: React.DragEvent) => {
    event.preventDefault();
    event.stopPropagation();

    if (!canvasRef.current || !currentPreset) {
      setIsDragging(false);
      setDropPosition(null);
      return;
    }

    const canvasBounds = canvasRef.current.getBoundingClientRect();

    const handleCreateNode = (nodeType: NodeTypeDefinition, position: { x: number; y: number }) => {
      const newNode = {
        id: `node-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        type: nodeType.id,
        position,
        data: {
          name: `${nodeType.name} ${nodes.length + 1}`,
        },
      };

      addNode(newNode);
    };

    onDrop(event, canvasBounds, handleCreateNode);

    setIsDragging(false);
    setDropPosition(null);
  };

  return (
    <div
      ref={canvasRef}
      onDragEnter={handleDragEnter}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className="relative flex-1 overflow-hidden"
    >
      {children}

      {isDragging && dropPosition && (
        <div
          className="pointer-events-none absolute rounded-lg border-2 border-dashed border-blue-500 bg-blue-500/10 flex items-center justify-center"
          style={{
            left: dropPosition.x - 50,
            top: dropPosition.y - 25,
            width: 100,
            height: 50,
            zIndex: 1000,
          }}
        >
          <div className="text-blue-600 dark:text-blue-400 font-medium text-sm">
            Drop here
          </div>
        </div>
      )}

      {isDragging && !dropPosition && (
        <div
          className="pointer-events-none absolute inset-0 bg-blue-500/5 border-2 border-dashed border-blue-500 rounded-lg"
          style={{
            zIndex: 999,
          }}
        >
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="text-blue-600 dark:text-blue-400 font-medium">
              Drop node on canvas
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default DropZone;
