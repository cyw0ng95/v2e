'use client';

import React, { useState, useCallback } from 'react';
import { Upload, FileJson, AlertCircle, CheckCircle, X } from 'lucide-react';
import { toast } from 'sonner';
import {
  importSTIX,
  validateSTIX,
  type STIXImportOptions,
  type STIXImportResult,
} from '@/lib/glc/stix';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useReactFlow } from '@xyflow/react';

// ============================================================================
// Props
// ============================================================================

interface STIXImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// ============================================================================
// STIX Import Dialog Component
// ============================================================================

export const STIXImportDialog: React.FC<STIXImportDialogProps> = ({
  open,
  onOpenChange,
}) => {
  const { setNodes, setEdges } = useReactFlow();

  const [jsonContent, setJsonContent] = useState('');
  const [isDragging, setIsDragging] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [importResult, setImportResult] = useState<STIXImportResult | null>(null);

  // Import options
  const [mapToGLCTypes, setMapToGLCTypes] = useState(true);
  const [mapToD3FEND, setMapToD3FEND] = useState(false);
  const [includeRelationships, setIncludeRelationships] = useState(true);

  const options: STIXImportOptions = {
    mapToGLCTypes,
    mapToD3FEND,
    includeRelationships,
  };

  const handleFileUpload = useCallback(
    async (file: File) => {
      if (!file.name.endsWith('.json')) {
        toast.error('Please upload a JSON file');
        return;
      }

      try {
        const content = await file.text();
        setJsonContent(content);

        // Validate immediately
        const validation = await validateSTIX(content);
        if (!validation.valid && validation.errors.length > 0) {
          toast.error(`Validation failed: ${validation.errors[0].message}`);
        } else {
          toast.success('STIX file loaded successfully');
        }
      } catch (error) {
        toast.error('Failed to read file');
        console.error(error);
      }
    },
    []
  );

  const handleDrop = useCallback(
    async (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragging(false);

      const file = e.dataTransfer.files[0];
      if (file) {
        await handleFileUpload(file);
      }
    },
    [handleFileUpload]
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback(() => {
    setIsDragging(false);
  }, []);

  const handleImport = useCallback(async () => {
    if (!jsonContent.trim()) {
      toast.error('Please provide STIX JSON content');
      return;
    }

    setIsLoading(true);

    try {
      const result = await importSTIX(jsonContent, options);

      if (result.nodes.length === 0 && result.edges.length === 0) {
        toast.warning('No objects were imported. Check filters and try again.');
        setImportResult(result);
        setIsLoading(false);
        return;
      }

      // Add nodes to graph
      setNodes((nodes) => [...nodes, ...result.nodes]);
      setEdges((edges) => [...edges, ...result.edges]);

      setImportResult(result);

      toast.success(
        `Imported ${result.nodes.length} nodes and ${result.edges.length} edges`
      );

      // Close dialog on success
      setTimeout(() => {
        onOpenChange(false);
        setJsonContent('');
        setImportResult(null);
      }, 2000);
    } catch (error) {
      toast.error('Import failed');
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, [jsonContent, options, setNodes, setEdges, onOpenChange]);

  const handleClose = useCallback(() => {
    onOpenChange(false);
    setJsonContent('');
    setImportResult(null);
  }, [onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileJson className="h-5 w-5" />
            Import STIX 2.1
          </DialogTitle>
          <DialogDescription>
            Import STIX 2.1 JSON files and convert to GLC graph
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="upload" className="mt-4">
          <TabsList>
            <TabsTrigger value="upload">Upload</TabsTrigger>
            <TabsTrigger value="paste">Paste JSON</TabsTrigger>
            <TabsTrigger value="results">Results</TabsTrigger>
          </TabsList>

          <TabsContent value="upload" className="space-y-4">
            <div
              className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                isDragging
                  ? 'border-primary bg-primary/5'
                  : 'border-border hover:border-primary/50'
              }`}
              onDrop={handleDrop}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
            >
              <Upload className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <div className="text-lg font-medium mb-2">
                Drag and drop STIX JSON file
              </div>
              <div className="text-sm text-muted-foreground mb-4">
                or click to browse
              </div>
              <input
                type="file"
                accept=".json"
                className="hidden"
                id="stix-file-input"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFileUpload(file);
                }}
              />
              <Button
                variant="outline"
                onClick={() =>
                  document.getElementById('stix-file-input')?.click()
                }
              >
                Browse Files
              </Button>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label htmlFor="map-to-glc">
                  Map to GLC node types
                </Label>
                <Switch
                  id="map-to-glc"
                  checked={mapToGLCTypes}
                  onCheckedChange={setMapToGLCTypes}
                />
              </div>

              <div className="flex items-center justify-between">
                <Label htmlFor="map-to-d3fend">
                  Map to D3FEND ontology
                </Label>
                <Switch
                  id="map-to-d3fend"
                  checked={mapToD3FEND}
                  onCheckedChange={setMapToD3FEND}
                />
              </div>

              <div className="flex items-center justify-between">
                <Label htmlFor="include-relationships">
                  Include relationships
                </Label>
                <Switch
                  id="include-relationships"
                  checked={includeRelationships}
                  onCheckedChange={setIncludeRelationships}
                />
              </div>
            </div>
          </TabsContent>

          <TabsContent value="paste" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="stix-json-input">STIX JSON</Label>
              <textarea
                id="stix-json-input"
                className="w-full h-64 px-3 py-2 rounded-md border bg-background text-sm font-mono resize-none"
                placeholder='{
  "type": "bundle",
  "id": "bundle--..."
  "objects": [...]
}'
                value={jsonContent}
                onChange={(e) => setJsonContent(e.target.value)}
              />
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label htmlFor="map-to-glc-paste">
                  Map to GLC node types
                </Label>
                <Switch
                  id="map-to-glc-paste"
                  checked={mapToGLCTypes}
                  onCheckedChange={setMapToGLCTypes}
                />
              </div>

              <div className="flex items-center justify-between">
                <Label htmlFor="map-to-d3fend-paste">
                  Map to D3FEND ontology
                </Label>
                <Switch
                  id="map-to-d3fend-paste"
                  checked={mapToD3FEND}
                  onCheckedChange={setMapToD3FEND}
                />
              </div>
            </div>
          </TabsContent>

          <TabsContent value="results" className="space-y-4">
            {importResult ? (
              <>
                {/* Statistics */}
                <Card>
                  <CardHeader>
                    <CardTitle>Import Statistics</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <div className="text-2xl font-bold">
                          {importResult.stats.importedObjects}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          Objects Imported
                        </div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold">
                          {importResult.stats.relationshipCount}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          Relationships
                        </div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-green-600">
                          {importResult.nodes.length}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          Nodes Created
                        </div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-blue-600">
                          {importResult.edges.length}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          Edges Created
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Errors */}
                {importResult.errors.length > 0 && (
                  <Card className="border-red-200">
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2 text-red-600">
                        <AlertCircle className="h-5 w-5" />
                        Validation Errors ({importResult.errors.length})
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2 max-h-48 overflow-y-auto">
                        {importResult.errors.map((error, index) => (
                          <div
                            key={index}
                            className="flex items-start gap-2 p-2 rounded bg-red-50"
                          >
                            <Badge variant="destructive" className="shrink-0">
                              {error.type}
                            </Badge>
                            <span className="text-sm">{error.message}</span>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Success */}
                {importResult.errors.length === 0 && (
                  <Card className="border-green-200">
                    <CardContent className="pt-6">
                      <div className="flex items-center gap-3 text-green-600">
                        <CheckCircle className="h-8 w-8" />
                        <div>
                          <div className="font-medium">Import Successful!</div>
                          <div className="text-sm text-muted-foreground">
                            All objects validated and imported
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                )}
              </>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                Import a STIX file to see results
              </div>
            )}
          </TabsContent>
        </Tabs>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            Cancel
          </Button>
          <Button onClick={handleImport} disabled={isLoading || !jsonContent.trim()}>
            {isLoading ? 'Importing...' : 'Import STIX'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
