'use client';

import { useState } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useGLCStore } from '@/lib/glc/store';

interface EdgeDetailsSheetProps {
  edgeId: string;
  onClose: () => void;
}

function EdgeDetailsSheetContent({ edgeId, onClose }: EdgeDetailsSheetProps) {
  const { graph, currentPreset, updateEdge } = useGLCStore();

  // Find the edge
  const edge = graph?.edges.find((e) => e.id === edgeId);

  // Initialize form state with edge data (component is remounted when edgeId changes)
  const [relationshipId, setRelationshipId] = useState(edge?.data?.relationshipId || '');
  const [label, setLabel] = useState(edge?.data?.label || '');
  const [notes, setNotes] = useState(edge?.data?.notes || '');

  // Get available relationships based on source/target node types
  const sourceNode = graph?.nodes.find((n) => n.id === edge?.source);
  const targetNode = graph?.nodes.find((n) => n.id === edge?.target);

  const availableRelationships = currentPreset?.relations.filter(
    (r) =>
      (r.sourceTypes.includes(sourceNode?.data.typeId || '') ||
        r.sourceTypes.length === 0) &&
      (r.targetTypes.includes(targetNode?.data.typeId || '') ||
        r.targetTypes.length === 0)
  ) || [];

  // Save changes
  const handleSave = () => {
    if (!edge) return;
    updateEdge(edgeId, {
      relationshipId,
      label,
      notes,
    });
    onClose();
  };

  if (!edge || !currentPreset) return null;

  const theme = currentPreset.theme;
  const selectedRelationship = currentPreset.relations.find(
    (r) => r.id === relationshipId
  );

  return (
    <div
      className="absolute right-0 top-0 h-full w-80 border-l shadow-xl z-50 flex flex-col"
      style={{
        backgroundColor: theme.surface,
        borderColor: theme.border,
      }}
    >
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-3 border-b"
        style={{ borderColor: theme.border }}
      >
        <h3 className="font-semibold" style={{ color: theme.text }}>
          Edge Details
        </h3>
        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onClose}>
          <X className="w-4 h-4" style={{ color: theme.textMuted }} />
        </Button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {/* Source → Target */}
        <div className="mb-4 p-3 rounded-lg" style={{ backgroundColor: theme.background }}>
          <div className="flex items-center gap-2 text-sm">
            <span style={{ color: theme.textMuted }}>From:</span>
            <span style={{ color: theme.text }}>{sourceNode?.data.label}</span>
          </div>
          <div className="flex items-center gap-2 text-sm mt-1">
            <span style={{ color: theme.textMuted }}>To:</span>
            <span style={{ color: theme.text }}>{targetNode?.data.label}</span>
          </div>
        </div>

        {/* Relationship Type */}
        <div className="mb-4">
          <Label style={{ color: theme.text }}>Relationship Type</Label>
          <select
            value={relationshipId}
            onChange={(e) => setRelationshipId(e.target.value)}
            className="w-full h-10 px-3 mt-1 rounded border text-sm"
            style={{
              backgroundColor: theme.background,
              borderColor: theme.border,
              color: theme.text,
            }}
          >
            <option value="">Select relationship...</option>
            {availableRelationships.map((rel) => (
              <option key={rel.id} value={rel.id}>
                {rel.label}
              </option>
            ))}
          </select>
          {selectedRelationship?.description && (
            <p className="text-xs mt-1" style={{ color: theme.textMuted }}>
              {selectedRelationship.description}
            </p>
          )}
        </div>

        {/* Custom Label */}
        <div className="mb-4">
          <Label style={{ color: theme.text }}>Custom Label (optional)</Label>
          <Input
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="Override default label..."
            className="mt-1"
            style={{
              backgroundColor: theme.background,
              borderColor: theme.border,
              color: theme.text,
            }}
          />
        </div>

        {/* Notes */}
        <div className="mb-4">
          <Label style={{ color: theme.text }}>Notes</Label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Add notes..."
            className="w-full h-32 p-3 mt-1 rounded-lg border resize-none text-sm"
            style={{
              backgroundColor: theme.background,
              borderColor: theme.border,
              color: theme.text,
            }}
          />
        </div>
      </div>

      {/* Footer */}
      <div
        className="flex gap-2 px-4 py-3 border-t"
        style={{ borderColor: theme.border }}
      >
        <Button variant="outline" className="flex-1" onClick={onClose}>
          Cancel
        </Button>
        <Button className="flex-1" onClick={handleSave}>
          Save
        </Button>
      </div>
    </div>
  );
}

// Wrapper that remounts content when edgeId changes to reset form state
export function EdgeDetailsSheet(props: EdgeDetailsSheetProps) {
  // Use key to force remount when edgeId changes
  return <EdgeDetailsSheetContent key={props.edgeId} {...props} />;
}
