'use client';

import { useCallback, useState, useEffect } from 'react';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import {
  Copy,
  Scissors,
  Clipboard,
  Trash2,
  Edit3,
  ArrowRightLeft,
  RotateCcw,
  RotateCw,
} from 'lucide-react';
import { useGLCStore } from '@/lib/glc/store';

interface NodeContextMenuProps {
  children: React.ReactNode;
  nodeId: string;
  onEdit?: () => void;
  onDuplicate?: () => void;
}

export function NodeContextMenu({
  children,
  nodeId,
  onEdit,
  onDuplicate,
}: NodeContextMenuProps) {
  const { graph, removeNode, canUndo, canRedo, undo, redo } = useGLCStore();

  const handleDelete = useCallback(() => {
    removeNode(nodeId);
  }, [nodeId, removeNode]);

  const handleDuplicate = useCallback(() => {
    onDuplicate?.();
  }, [onDuplicate]);

  const handleEdit = useCallback(() => {
    onEdit?.();
  }, [onEdit]);

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent className="w-48">
        <ContextMenuItem onClick={handleEdit}>
          <Edit3 className="w-4 h-4 mr-2" />
          Edit
        </ContextMenuItem>
        <ContextMenuItem onClick={handleDuplicate}>
          <Copy className="w-4 h-4 mr-2" />
          Duplicate
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem onClick={() => undo()} disabled={!canUndo}>
          <RotateCcw className="w-4 h-4 mr-2" />
          Undo
        </ContextMenuItem>
        <ContextMenuItem onClick={() => redo()} disabled={!canRedo}>
          <RotateCw className="w-4 h-4 mr-2" />
          Redo
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem onClick={handleDelete} className="text-destructive">
          <Trash2 className="w-4 h-4 mr-2" />
          Delete
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}

interface EdgeContextMenuProps {
  children: React.ReactNode;
  edgeId: string;
  onEdit?: () => void;
  onReverse?: () => void;
}

export function EdgeContextMenu({
  children,
  edgeId,
  onEdit,
  onReverse,
}: EdgeContextMenuProps) {
  const { removeEdge } = useGLCStore();

  const handleDelete = useCallback(() => {
    removeEdge(edgeId);
  }, [edgeId, removeEdge]);

  const handleReverse = useCallback(() => {
    onReverse?.();
  }, [onReverse]);

  const handleEdit = useCallback(() => {
    onEdit?.();
  }, [onEdit]);

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent className="w-48">
        <ContextMenuItem onClick={handleEdit}>
          <Edit3 className="w-4 h-4 mr-2" />
          Edit
        </ContextMenuItem>
        <ContextMenuItem onClick={handleReverse}>
          <ArrowRightLeft className="w-4 h-4 mr-2" />
          Reverse
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem onClick={handleDelete} className="text-destructive">
          <Trash2 className="w-4 h-4 mr-2" />
          Delete
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}

interface CanvasContextMenuProps {
  children: React.ReactNode;
  onPaste?: () => void;
  onFitView?: () => void;
}

export function CanvasContextMenu({
  children,
  onPaste,
  onFitView,
}: CanvasContextMenuProps) {
  const { canUndo, canRedo, undo, redo } = useGLCStore();

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent className="w-48">
        <ContextMenuItem onClick={onPaste}>
          <Clipboard className="w-4 h-4 mr-2" />
          Paste
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem onClick={() => undo()} disabled={!canUndo}>
          <RotateCcw className="w-4 h-4 mr-2" />
          Undo
        </ContextMenuItem>
        <ContextMenuItem onClick={() => redo()} disabled={!canRedo}>
          <RotateCw className="w-4 h-4 mr-2" />
          Redo
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem onClick={onFitView}>
          Fit View
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}
