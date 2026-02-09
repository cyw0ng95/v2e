'use client';

import { useState } from 'react';
import { useGLCStore } from '@/lib/glc/store';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { X, Link } from 'lucide-react';
import { RelationshipDefinition } from "@/lib/glc/types"

interface RelationshipPickerProps {
  sourceNodeId: string;
  targetNodeId: string;
  onRelationshipSelect: (relationshipId: string) => void;
  children: React.ReactNode;
}

export function RelationshipPicker({
  sourceNodeId,
  targetNodeId,
  onRelationshipSelect,
  children,
}: RelationshipPickerProps) {
  const { currentPreset, nodes } = useGLCStore() as any;
  const [open, setOpen] = useState(false);
  const [selectedRelationship, setSelectedRelationship] = useState<string | null>(null);

  if (!currentPreset) {
    return null;
  }

  const sourceNode = nodes.find((n: any) => n.id === sourceNodeId);
  const targetNode = nodes.find((n: any) => n.id === targetNodeId);

  if (!sourceNode || !targetNode) {
    return null;
  }

  const validRelationships = currentPreset.relationshipTypes.filter(rel: any =>
    (rel.sourceNodeTypes.includes('*') || rel.sourceNodeTypes.includes(sourceNode.type)) &&
    (rel.targetNodeTypes.includes('*') || rel.targetNodeTypes.includes(targetNode.type))
  );

  const handleConfirm = () => {
    if (selectedRelationship) {
      onRelationshipSelect(selectedRelationship);
      setOpen(false);
      setSelectedRelationship(null);
    }
  };

  const handleCancel = () => {
    setOpen(false);
    setSelectedRelationship(null);
  };

  if (validRelationships.length === 0) {
    return null;
  }

  if (validRelationships.length === 1) {
    const handleSingleRelationship = () => {
      onRelationshipSelect(validRelationships[0].id);
    };

    return (
      <DialogTrigger asChild onClick={handleSingleRelationship}>
        {children}
      </DialogTrigger>
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Link className="h-5 w-5" />
              Select Relationship Type
            </span>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setOpen(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="text-sm text-muted-foreground">
            Select the type of relationship to create between:
          </div>

          <div className="flex items-center gap-2 text-sm">
            <div className="font-medium">{sourceNode.data.name || sourceNode.type}</div>
            <div className="text-muted-foreground">→</div>
            <div className="font-medium">{targetNode.data.name || targetNode.type}</div>
          </div>

          <div className="grid gap-3">
            {validRelationships.map(rel: any => (
              <button
                key={rel.id}
                onClick={() => setSelectedRelationship(rel.id)}
                className={`flex items-start gap-3 p-4 rounded-lg border-2 text-left transition-all ${
                  selectedRelationship === rel.id
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-950'
                    : 'border-border hover:border-blue-300 dark:hover:border-blue-700'
                }`}
              >
                <div className="flex-1">
                  <div className="font-medium mb-1">{rel.name}</div>
                  <div className="text-sm text-muted-foreground">
                    {rel.description}
                  </div>
                  <div className="mt-2 flex items-center gap-2">
                    <span className="px-2 py-0.5 bg-muted text-xs rounded-full">
                      {rel.category}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {rel.directionality}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {rel.multiplicity}
                    </span>
                  </div>
                </div>
                <div className="mt-2">
                  <div className="w-3 h-3 rounded-full" style={{ backgroundColor: rel.style.strokeColor }} />
                </div>
              </button>
            ))}
          </div>

          <div className="flex gap-2 justify-end pt-4 border-t">
            <Button variant="outline" onClick={handleCancel}>
              Cancel
            </Button>
            <Button
              onClick={handleConfirm}
              disabled={!selectedRelationship}
            >
              Create Relationship
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export default RelationshipPicker;
