'use client';

import { useState, useEffect } from 'react';
import { X, Plus, Trash2, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useGLCStore } from '@/lib/glc/store';
import type { Property, Reference, NodeTypeDefinition } from '@/lib/glc/types';

interface NodeDetailsSheetProps {
  nodeId: string;
  onClose: () => void;
}

export function NodeDetailsSheet({ nodeId, onClose }: NodeDetailsSheetProps) {
  const { graph, currentPreset, updateNode } = useGLCStore();
  const [label, setLabel] = useState('');
  const [notes, setNotes] = useState('');
  const [properties, setProperties] = useState<Property[]>([]);
  const [references, setReferences] = useState<Reference[]>([]);

  // Find the node
  const node = graph?.nodes.find((n) => n.id === nodeId);
  const nodeType = currentPreset?.nodeTypes.find((t) => t.id === node?.data.typeId);

  // Initialize form state when node changes
  useEffect(() => {
    if (node) {
      setLabel(node.data.label || '');
      setNotes(node.data.notes || '');
      setProperties(node.data.properties || []);
      setReferences(node.data.references || []);
    }
  }, [node]);

  // Save changes
  const handleSave = () => {
    if (!node) return;
    updateNode(nodeId, {
      label,
      notes,
      properties,
      references,
    });
    onClose();
  };

  // Add property
  const addProperty = () => {
    setProperties([...properties, { key: '', value: '', type: 'string' }]);
  };

  // Update property
  const updateProperty = (index: number, updates: Partial<Property>) => {
    const newProperties = [...properties];
    newProperties[index] = { ...newProperties[index], ...updates };
    setProperties(newProperties);
  };

  // Remove property
  const removeProperty = (index: number) => {
    setProperties(properties.filter((_, i) => i !== index));
  };

  // Add reference
  const addReference = () => {
    setReferences([...references, { type: 'url', id: '' }]);
  };

  // Update reference
  const updateReference = (index: number, updates: Partial<Reference>) => {
    const newReferences = [...references];
    newReferences[index] = { ...newReferences[index], ...updates };
    setReferences(newReferences);
  };

  // Remove reference
  const removeReference = (index: number) => {
    setReferences(references.filter((_, i) => i !== index));
  };

  if (!node || !currentPreset) return null;

  const theme = currentPreset.theme;

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
          Node Details
        </h3>
        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onClose}>
          <X className="w-4 h-4" style={{ color: theme.textMuted }} />
        </Button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {/* Node Type Badge */}
        <div className="mb-4">
          <div
            className="inline-flex items-center px-2 py-1 rounded text-xs font-medium"
            style={{
              backgroundColor: nodeType?.color + '20',
              color: nodeType?.color,
            }}
          >
            {nodeType?.label || 'Unknown'}
          </div>
        </div>

        {/* Label */}
        <div className="mb-4">
          <Label style={{ color: theme.text }}>Label</Label>
          <Input
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            className="mt-1"
            style={{
              backgroundColor: theme.background,
              borderColor: theme.border,
              color: theme.text,
            }}
          />
        </div>

        {/* Tabs for Properties, References, Notes */}
        <Tabs defaultValue="properties" className="w-full">
          <TabsList className="w-full">
            <TabsTrigger value="properties" className="flex-1">Properties</TabsTrigger>
            <TabsTrigger value="references" className="flex-1">References</TabsTrigger>
            <TabsTrigger value="notes" className="flex-1">Notes</TabsTrigger>
          </TabsList>

          {/* Properties Tab */}
          <TabsContent value="properties" className="mt-4">
            <div className="space-y-2">
              {properties.map((prop, index) => (
                <div key={index} className="flex gap-2 items-start">
                  <Input
                    value={prop.key}
                    onChange={(e) => updateProperty(index, { key: e.target.value })}
                    placeholder="Key"
                    className="flex-1 text-sm"
                    style={{
                      backgroundColor: theme.background,
                      borderColor: theme.border,
                      color: theme.text,
                    }}
                  />
                  <Input
                    value={prop.value}
                    onChange={(e) => updateProperty(index, { value: e.target.value })}
                    placeholder="Value"
                    className="flex-1 text-sm"
                    style={{
                      backgroundColor: theme.background,
                      borderColor: theme.border,
                      color: theme.text,
                    }}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9"
                    onClick={() => removeProperty(index)}
                  >
                    <Trash2 className="w-4 h-4" style={{ color: theme.error }} />
                  </Button>
                </div>
              ))}
              <Button
                variant="outline"
                size="sm"
                className="w-full mt-2"
                onClick={addProperty}
              >
                <Plus className="w-4 h-4 mr-2" />
                Add Property
              </Button>
            </div>
          </TabsContent>

          {/* References Tab */}
          <TabsContent value="references" className="mt-4">
            <div className="space-y-2">
              {references.map((ref, index) => (
                <div key={index} className="flex gap-2 items-start">
                  <select
                    value={ref.type}
                    onChange={(e) => updateReference(index, { type: e.target.value as Reference['type'] })}
                    className="h-9 px-2 rounded border text-sm"
                    style={{
                      backgroundColor: theme.background,
                      borderColor: theme.border,
                      color: theme.text,
                    }}
                  >
                    <option value="cve">CVE</option>
                    <option value="cwe">CWE</option>
                    <option value="capec">CAPEC</option>
                    <option value="attack">ATT&CK</option>
                    <option value="d3fend">D3FEND</option>
                    <option value="url">URL</option>
                    <option value="stix">STIX</option>
                  </select>
                  <Input
                    value={ref.id}
                    onChange={(e) => updateReference(index, { id: e.target.value })}
                    placeholder="ID / URL"
                    className="flex-1 text-sm"
                    style={{
                      backgroundColor: theme.background,
                      borderColor: theme.border,
                      color: theme.text,
                    }}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9"
                    onClick={() => removeReference(index)}
                  >
                    <Trash2 className="w-4 h-4" style={{ color: theme.error }} />
                  </Button>
                </div>
              ))}
              <Button
                variant="outline"
                size="sm"
                className="w-full mt-2"
                onClick={addReference}
              >
                <Plus className="w-4 h-4 mr-2" />
                Add Reference
              </Button>
            </div>
          </TabsContent>

          {/* Notes Tab */}
          <TabsContent value="notes" className="mt-4">
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Add notes..."
              className="w-full h-48 p-3 rounded-lg border resize-none text-sm"
              style={{
                backgroundColor: theme.background,
                borderColor: theme.border,
                color: theme.text,
              }}
            />
          </TabsContent>
        </Tabs>
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
