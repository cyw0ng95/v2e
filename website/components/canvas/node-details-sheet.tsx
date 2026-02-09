'use client';

import { useState, useEffect } from 'react';
import { useGLCStore } from '@/lib/glc/store';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { X, Save, Trash2 } from 'lucide-react';
import { CADNode } from "@/lib/glc/types"
import { showError } from '../../lib/glc/errors/error-handler';
import { validateNodePosition } from '../../lib/glc/utils';

interface NodeDetailsSheetProps {
  nodeId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function NodeDetailsSheet({ nodeId, open, onOpenChange }: NodeDetailsSheetProps) {
  const { currentPreset, nodes, updateNode, deleteNode } = useGLCStore() as any;
  const [formData, setFormData] = useState<Record<string, any>>({});
  const [position, setPosition] = useState({ x: 0, y: 0 });

  const node = nodeId ? nodes.find((n: any) => n.id === nodeId) : null;
  const nodeType = node && currentPreset ? currentPreset.nodeTypes.find((nt: any) => nt.id === node.type) : null;

  useEffect(() => {
    if (node) {
      setFormData(node.data);
      setPosition(node.position);
    }
  }, [node, open]);

  const handleSave = () => {
    if (!node) return;

    try {
      updateNode(node.id, {
        data: formData,
        position,
      });
      onOpenChange(false);
    } catch (error) {
      showError('Failed to update node', { nodeId, error });
    }
  };

  const handleDelete = () => {
    if (!node) return;
    
    try {
      deleteNode(node.id);
      onOpenChange(false);
    } catch (error) {
      showError('Failed to delete node', { nodeId, error });
    }
  };

  const handlePositionChange = (field: 'x' | 'y', value: string) => {
    const numValue = parseFloat(value);
    if (isNaN(numValue)) return;

    const newPosition = { ...position, [field]: numValue };
    
    if (!validateNodePosition(nodes.filter((n: any) => n.id !== nodeId), newPosition)) {
      return;
    }

    setPosition(newPosition);
  };

  if (!node || !nodeType) {
    return null;
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[400px] overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center justify-between">
            <span>Node Details</span>
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
            <Label htmlFor="node-name">Name</Label>
            <Input
              id="node-name"
              value={formData.name || ''}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Node name"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="node-x">X Position</Label>
              <Input
                id="node-x"
                type="number"
                value={position.x}
                onChange={(e) => handlePositionChange('x', e.target.value)}
                step={currentPreset?.behavior.snapToGrid ? currentPreset.behavior.gridSize : 1}
              />
            </div>
            <div>
              <Label htmlFor="node-y">Y Position</Label>
              <Input
                id="node-y"
                type="number"
                value={position.y}
                onChange={(e) => handlePositionChange('y', e.target.value)}
                step={currentPreset?.behavior.snapToGrid ? currentPreset.behavior.gridSize : 1}
              />
            </div>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Properties</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {nodeType.properties.map((prop: any) => (
                <div key={prop.id}>
                  <Label htmlFor={`prop-${prop.id}`}>
                    {prop.name}
                    {prop.required && <span className="text-red-500 ml-1">*</span>}
                  </Label>
                  {prop.type === 'text' && (
                    <Input
                      id={`prop-${prop.id}`}
                      value={formData[prop.id] || ''}
                      onChange={(e) => setFormData({ ...formData, [prop.id]: e.target.value })}
                      placeholder={`Enter ${prop.name.toLowerCase()}`}
                      required={prop.required}
                    />
                  )}
                  {prop.type === 'number' && (
                    <Input
                      id={`prop-${prop.id}`}
                      type="number"
                      value={formData[prop.id] || ''}
                      onChange={(e) => setFormData({ ...formData, [prop.id]: parseFloat(e.target.value) })}
                      required={prop.required}
                    />
                  )}
                  {prop.type === 'boolean' && (
                    <select
                      id={`prop-${prop.id}`}
                      value={formData[prop.id] ? 'true' : 'false'}
                      onChange={(e) => setFormData({ ...formData, [prop.id]: e.target.value === 'true' })}
                      className="w-full px-3 py-2 border rounded-md bg-background"
                    >
                      <option value="true">Yes</option>
                      <option value="false">No</option>
                    </select>
                  )}
                  {prop.type === 'enum' && prop.options && (
                    <select
                      id={`prop-${prop.id}`}
                      value={formData[prop.id] || ''}
                      onChange={(e) => setFormData({ ...formData, [prop.id]: e.target.value })}
                      className="w-full px-3 py-2 border rounded-md bg-background"
                      required={prop.required}
                    >
                      <option value="">Select...</option>
                      {prop.options.map((option: any) => (
                        <option key={option} value={option}>{option}</option>
                      ))}
                    </select>
                  )}
                </div>
              ))}
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

export default NodeDetailsSheet;
