'use client';

import { useCallback, useState } from 'react';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
  ContextMenuSub,
  ContextMenuSubTrigger,
  ContextMenuSubContent,
} from '@/components/ui/context-menu';
import {
  Copy,
  Clipboard,
  Trash2,
  Edit3,
  ArrowRightLeft,
  RotateCcw,
  RotateCw,
  Palette,
  Lightbulb,
  Check,
} from 'lucide-react';
import { useGLCStore, createNode } from '@/lib/glc/store';
import type { CADNode, CADEdge } from '@/lib/glc/types';
import type { Node, Edge } from '@xyflow/react';

// Predefined color palette
const COLOR_PALETTE = [
  { name: 'Default', value: null },
  { name: 'Red', value: '#ef4444' },
  { name: 'Orange', value: '#f97316' },
  { name: 'Amber', value: '#f59e0b' },
  { name: 'Yellow', value: '#eab308' },
  { name: 'Lime', value: '#84cc16' },
  { name: 'Green', value: '#22c55e' },
  { name: 'Emerald', value: '#10b981' },
  { name: 'Teal', value: '#14b8a6' },
  { name: 'Cyan', value: '#06b6d4' },
  { name: 'Sky', value: '#0ea5e9' },
  { name: 'Blue', value: '#3b82f6' },
  { name: 'Indigo', value: '#6366f1' },
  { name: 'Violet', value: '#8b5cf6' },
  { name: 'Purple', value: '#a855f7' },
  { name: 'Fuchsia', value: '#d946ef' },
  { name: 'Pink', value: '#ec4899' },
  { name: 'Rose', value: '#f43f5e' },
];

// Type for D3FEND inference
interface D3FENDInference {
  targetTypes: string[];
  relationshipId: string;
  label: string;
}

// D3FEND inference rules - suggest relationships based on node types
const D3FEND_INFERENCES: Record<string, D3FENDInference[]> = {
  'attack-technique': [
    { targetTypes: ['technique'], relationshipId: 'mitigates', label: 'Can be mitigated by' },
    { targetTypes: ['technique'], relationshipId: 'detects', label: 'Can be detected by' },
    { targetTypes: ['vulnerability', 'weakness'], relationshipId: 'exploits', label: 'Exploits' },
    { targetTypes: ['asset', 'software'], relationshipId: 'targets', label: 'Targets' },
  ],
  'technique': [
    { targetTypes: ['attack-technique'], relationshipId: 'mitigates', label: 'Mitigates' },
    { targetTypes: ['attack-technique'], relationshipId: 'detects', label: 'Detects' },
    { targetTypes: ['vulnerability', 'weakness'], relationshipId: 'mitigates', label: 'Mitigates' },
  ],
  'vulnerability': [
    { targetTypes: ['weakness'], relationshipId: 'caused-by', label: 'Caused by' },
    { targetTypes: ['technique'], relationshipId: 'mitigates', label: 'Can be mitigated by' },
  ],
  'weakness': [
    { targetTypes: ['vulnerability'], relationshipId: 'caused-by', label: 'Causes' },
    { targetTypes: ['technique'], relationshipId: 'mitigates', label: 'Can be mitigated by' },
  ],
  'group': [
    { targetTypes: ['attack-technique'], relationshipId: 'uses-malware', label: 'Uses' },
    { targetTypes: ['asset', 'software'], relationshipId: 'targets', label: 'Targets' },
  ],
  'asset': [
    { targetTypes: ['vulnerability'], relationshipId: 'has-vulnerability', label: 'Has vulnerability' },
    { targetTypes: ['asset'], relationshipId: 'connected-to', label: 'Connected to' },
  ],
  'software': [
    { targetTypes: ['asset'], relationshipId: 'runs-on', label: 'Runs on' },
    { targetTypes: ['vulnerability'], relationshipId: 'has-vulnerability', label: 'Has vulnerability' },
    { targetTypes: ['weakness'], relationshipId: 'has-weakness', label: 'Has weakness' },
  ],
};

interface NodeContextMenuProps {
  children: React.ReactNode;
  nodeId: string;
  node?: Node;
  nodes?: Node[];
  edges?: Edge[];
  onEdit?: () => void;
  onDuplicate?: (newNode: Node) => void;
  onCreateEdge?: (sourceId: string, targetId: string, relationshipId: string) => void;
}

export function NodeContextMenu({
  children,
  nodeId,
  node,
  nodes = [],
  edges = [],
  onEdit,
  onDuplicate,
  onCreateEdge,
}: NodeContextMenuProps) {
  const { graph, currentPreset, removeNode, updateNode, canUndo, canRedo, undo, redo } = useGLCStore();

  // Get current node color
  const currentColor = node?.data?.color || null;

  // Check if D3FEND preset with inference enabled
  const isInferenceEnabled = currentPreset?.meta.id === 'd3fend' && currentPreset?.behavior.enableInference;

  // Get inferred relationships for this node type
  const nodeTypeId = node?.data?.typeId as string | undefined;
  const inferences: D3FENDInference[] = nodeTypeId ? (D3FEND_INFERENCES[nodeTypeId] || []) : [];

  const handleDelete = useCallback(() => {
    removeNode(nodeId);
  }, [nodeId, removeNode]);

  const handleDuplicate = useCallback(() => {
    if (!node) return;

    // Create a duplicate node with offset position
    const newNode: Node = {
      ...node,
      id: crypto.randomUUID(),
      position: {
        x: node.position.x + 40,
        y: node.position.y + 40,
      },
      data: {
        ...node.data,
        label: `${node.data.label} (copy)`,
      },
      selected: false,
    };

    onDuplicate?.(newNode);
  }, [node, onDuplicate]);

  const handleEdit = useCallback(() => {
    onEdit?.();
  }, [onEdit]);

  const handleColorChange = useCallback((color: string | null) => {
    updateNode(nodeId, { color: color || undefined });
  }, [nodeId, updateNode]);

  const handleInference = useCallback((targetTypeId: string, relationshipId: string) => {
    // Find nodes of the target type that don't already have this relationship
    const existingTargets = new Set(
      edges
        .filter(e => e.source === nodeId && e.data?.relationshipId === relationshipId)
        .map(e => e.target)
    );

    const availableTargets = nodes.filter(
      n => n.id !== nodeId &&
      targetTypeId === n.data?.typeId &&
      !existingTargets.has(n.id)
    );

    // Create edges to available targets (limit to first 3 to avoid spam)
    availableTargets.slice(0, 3).forEach(target => {
      onCreateEdge?.(nodeId, target.id, relationshipId);
    });
  }, [nodeId, edges, nodes, onCreateEdge]);

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent className="w-56">
        <ContextMenuItem onClick={handleEdit}>
          <Edit3 className="w-4 h-4 mr-2" />
          Edit
        </ContextMenuItem>
        <ContextMenuItem onClick={handleDuplicate}>
          <Copy className="w-4 h-4 mr-2" />
          Duplicate
        </ContextMenuItem>

        {/* Color Picker Submenu */}
        <ContextMenuSub>
          <ContextMenuSubTrigger>
            <Palette className="w-4 h-4 mr-2" />
            Color
          </ContextMenuSubTrigger>
          <ContextMenuSubContent className="w-48">
            {COLOR_PALETTE.map((color) => (
              <ContextMenuItem
                key={color.name}
                onClick={() => handleColorChange(color.value)}
                className="flex items-center justify-between"
              >
                <div className="flex items-center gap-2">
                  {color.value && (
                    <div
                      className="w-4 h-4 rounded border"
                      style={{ backgroundColor: color.value }}
                    />
                  )}
                  <span>{color.name}</span>
                </div>
                {currentColor === color.value && (
                  <Check className="w-4 h-4 text-primary" />
                )}
              </ContextMenuItem>
            ))}
          </ContextMenuSubContent>
        </ContextMenuSub>

        {/* D3FEND Inferences */}
        {isInferenceEnabled && inferences.length > 0 && (
          <>
            <ContextMenuSeparator />
            <ContextMenuSub>
              <ContextMenuSubTrigger>
                <Lightbulb className="w-4 h-4 mr-2" />
                D3FEND Inferences
              </ContextMenuSubTrigger>
              <ContextMenuSubContent className="w-56">
                {inferences.map((inf: D3FENDInference) => {
                  const targetCount = nodes.filter(
                    n => n.id !== nodeId && n.data?.typeId && inf.targetTypes.includes(n.data.typeId as string)
                  ).length;
                  return (
                    <ContextMenuItem
                      key={inf.relationshipId}
                      onClick={() => handleInference(inf.targetTypes[0], inf.relationshipId)}
                      disabled={targetCount === 0}
                    >
                      <span>{inf.label}</span>
                      {targetCount > 0 && (
                        <span className="ml-auto text-xs text-muted-foreground">
                          {targetCount} target{targetCount !== 1 ? 's' : ''}
                        </span>
                      )}
                    </ContextMenuItem>
                  );
                })}
              </ContextMenuSubContent>
            </ContextMenuSub>
          </>
        )}

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
  edge?: Edge;
  onEdit?: () => void;
  onReverse?: (reversedEdge: Edge) => void;
}

export function EdgeContextMenu({
  children,
  edgeId,
  edge,
  onEdit,
  onReverse,
}: EdgeContextMenuProps) {
  const { removeEdge, canUndo, canRedo, undo, redo } = useGLCStore();

  const handleDelete = useCallback(() => {
    removeEdge(edgeId);
  }, [edgeId, removeEdge]);

  const handleReverse = useCallback(() => {
    if (!edge) return;

    // Create a reversed edge with swapped source/target
    const reversedEdge: Edge = {
      ...edge,
      id: `edge-${edge.target}-${edge.source}`,
      source: edge.target,
      target: edge.source,
      selected: false,
    };

    onReverse?.(reversedEdge);
  }, [edge, onReverse]);

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
