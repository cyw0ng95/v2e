'use client';

import { useState, useEffect } from 'react';
import { useGLCStore } from '@/lib/glc/store';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { X, Save, Trash2 } from 'lucide-react';
import { CADEdge } from "@/lib/glc/types"
import { showError } from '../../lib/glc/errors/error-handler';

interface EdgeDetailsSheetProps {
  edgeId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EdgeDetailsSheet({ edgeId, open, onOpenChange }: EdgeDetailsSheetProps) {
  const { currentPreset, edges, updateEdge, deleteEdge, nodes } = useGLCStore() as any;
  const [formData, setFormData] = useState<Record<string, any>>({});

  const edge = edgeId ? edges.find((e: any) => e.id === edgeId) : null;
  const edgeType = edge && currentPreset ? currentPreset.relationshipTypes.find((rt: any) => rt.id === edge.type) : null;
  const sourceNode = edge ? nodes.find((n: any) => n.id === edge.source) : null;
  const targetNode = edge ? nodes.find((n: any) => n.id === edge.target) : null;

  useEffect(() => {
    if (edge) {
      setFormData(edge.data);
    }
  }, [edge, open]);

  const handleSave = () => {
    if (!edge) return;

    try {
      updateEdge(edge.id, {
        data: formData,
      });
      onOpenChange(false);
    } catch (error) {
      showError('Failed to update edge', { edgeId, error });
    }
  };

  const handleDelete = () => {
    if (!edge) return;

    try {
      deleteEdge(edge.id);
      onOpenChange(false);
    } catch (error) {
      showError('Failed to delete edge', { edgeId, error });
    }
  };

  const getValidRelationshipTypes = (sourceNodeId: string, targetNodeId: string) => {
    if (!currentPreset) return [];

    const sourceNode = nodes.find((n: any) => n.id === sourceNodeId);
    const targetNode = nodes.find((n: any) => n.id === targetNodeId);

    if (!sourceNode || !targetNode) return [];

    return currentPreset.relationshipTypes.filter((rel: any) =>
      (rel.sourceNodeTypes.includes('*') || rel.sourceNodeTypes.includes(sourceNode.type)) &&
      (rel.targetNodeTypes.includes('*') || rel.targetNodeTypes.includes(targetNode.type))
    );
  };

  const validRelationshipTypes = edge ? getValidRelationshipTypes(edge.source, edge.target) : [];

  if (!edge || !edgeType) {
    return null;
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[400px] overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center justify-between">
            <span>Edge Details</span>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onOpenChange(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </SheetTitle>
        </SheetHeader>

        <div className="space-y-6 py-4">
          <div>
            <Label htmlFor="edge-id">Edge ID</Label>
            <Input
              id="edge-id"
              value={edge.id}
              disabled
              className="bg-muted"
            />
          </div>

          <div>
            <Label htmlFor="edge-label">Label</Label>
            <Input
              id="edge-label"
              value={formData.label || ''}
              onChange={(e) => setFormData({ ...formData, label: e.target.value })}
              placeholder="Enter edge label"
            />
          </div>

          <div>
            <Label htmlFor="edge-type">Relationship Type</Label>
            <Select value={edge.type} onValueChange={(value) => setFormData({ ...formData, type: value })}>
              <SelectTrigger>
                <SelectValue placeholder="Select relationship type" />
              </SelectTrigger>
              <SelectContent>
                {validRelationshipTypes.map((rel: any) => (
                  <SelectItem key={rel.id} value={rel.id}>
                    {rel.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="source-node">Source</Label>
              <Input
                id="source-node"
                value={sourceNode?.data.name || sourceNode?.type || ''}
                disabled
                className="bg-muted"
              />
            </div>
            <div>
              <Label htmlFor="target-node">Target</Label>
              <Input
                id="target-node"
                value={targetNode?.data.name || targetNode?.type || ''}
                disabled
                className="bg-muted"
              />
            </div>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Relationship Metadata</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div>
                <div className="text-sm font-medium mb-1">{edgeType.name}</div>
                <div className="text-xs text-muted-foreground">
                  {edgeType.description}
                </div>
              </div>

              <div className="flex items-center gap-2">
                <span className="px-2 py-1 bg-muted text-xs rounded-full">
                  {edgeType.category}
                </span>
                <span className="text-xs text-muted-foreground">
                  {edgeType.directionality}
                </span>
                <span className="text-xs text-muted-foreground">
                  {edgeType.multiplicity}
                </span>
              </div>

              {edgeType.properties && edgeType.properties.length > 0 && (
                <div>
                  <Label>Properties</Label>
                  <div className="text-sm text-muted-foreground">
                    {edgeType.properties.length} editable properties available
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          <div className="flex gap-2 pt-4 border-t">
            <Button
              onClick={handleSave}
              className="flex-1 bg-blue-600 hover:bg-blue-700"
            >
              <Save className="mr-2 h-4 w-4" />
              Save
            </Button>
            <Button
              onClick={handleDelete}
              variant="destructive"
              className="flex-1"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}

export default EdgeDetailsSheet;
