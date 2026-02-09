'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useGLCStore } from '@/lib/glc/store';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Loader2, Plus, LayoutTemplate } from 'lucide-react';
import CanvasWrapper from '@/components/canvas/canvas-wrapper';
import { nodeTypes, createFlowNodes } from '@/components/canvas/node-factory';
import { edgeTypes, createFlowEdges } from '@/components/canvas/edge-factory';
import { NodeDetailsSheet } from '@/components/canvas/node-details-sheet';
import { EdgeDetailsSheet } from '@/components/canvas/edge-details-sheet';
import { RelationshipPicker } from '@/components/canvas/relationship-picker';
import { NodePalette } from '@/components/canvas/node-palette';
import { DropZone } from '@/components/canvas/drop-zone';
import { ReactFlowProvider, useReactFlow } from '@xyflow/react';

export default function CanvasPage() {
  const params = useParams();
  const router = useRouter();
  const { currentPreset, setCurrentPreset, getPresetById, nodes, edges, setSelectedNodeId, setSelectedEdgeId, addEdge, updateNode, deleteNode } = useGLCStore();
  const [nodeDetailsOpen, setNodeDetailsOpen] = useState(false);
  const [edgeDetailsOpen, setEdgeDetailsOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(true);
  const { setNodes, setEdges } = useReactFlow();

  useEffect(() => {
    const presetId = params.presetId as string;
    const preset = getPresetById(presetId);

    if (!preset) {
      router.push('/glc');
      return;
    }

    if (!currentPreset || currentPreset.id !== presetId) {
      setCurrentPreset(preset);
    }
  }, [params.presetId, currentPreset, getPresetById, setCurrentPreset, router]);

  const handleAddNode = () => {
    if (!currentPreset) return;

    const nodeType = currentPreset.nodeTypes[0];
    const newNode = {
      id: `node-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      type: nodeType.id,
      position: {
        x: Math.random() * 400 + 200,
        y: Math.random() * 300 + 150,
      },
      data: {
        name: `${nodeType.name} ${nodes.length + 1}`,
      },
    };

    useGLCStore.getState().addNode(newNode);
  };

  const onNodesClick = (_: React.MouseEvent, node: any) => {
    setSelectedNodeId(node.id);
    setSelectedEdgeId(null);
    setEdgeDetailsOpen(false);
    setNodeDetailsOpen(true);
  };

  const onEdgesClick = (_: React.MouseEvent, edge: any) => {
    setSelectedNodeId(null);
    setSelectedEdgeId(edge.id);
    setNodeDetailsOpen(false);
    setEdgeDetailsOpen(true);
  };

  const onConnect = (params: any) => {
    if (!currentPreset) return;

    const existingEdge = edges.find(
      e => e.source === params.source && e.target === params.target
    );

    if (existingEdge) {
      return;
    }

    const validRelationships = currentPreset.relationshipTypes.filter(rel =>
      (rel.sourceNodeTypes.includes('*') || rel.sourceNodeTypes.includes(params.sourceType)) &&
      (rel.targetNodeTypes.includes('*') || rel.targetNodeTypes.includes(params.targetType))
    );

    if (validRelationships.length === 0) {
      return;
    }

    const relationshipType = validRelationships[0].id;

    const newEdge = {
      id: `edge-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      source: params.source,
      target: params.target,
      type: relationshipType,
      data: {},
    };

    addEdge(newEdge);
    setEdges((eds: any[]) => [...eds, {
      ...newEdge,
      id: `flow-${newEdge.id}`,
      type: 'dynamic-edge',
    }]);
  };

  if (!currentPreset) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900">
        <div className="text-center">
          <Loader2 className="w-12 h-12 text-blue-500 animate-spin mx-auto mb-4" />
          <p className="text-white">Loading preset...</p>
        </div>
      </div>
    );
  }

  return (
    <ReactFlowProvider>
      <div className="min-h-screen bg-slate-900 flex flex-col">
        <div className="container mx-auto px-4 py-4 border-b border-slate-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Button
                variant="ghost"
                onClick={() => router.push('/glc')}
                className="text-slate-300 hover:text-white hover:bg-slate-800"
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back
              </Button>

              <div>
                <h1 className="text-xl font-bold text-white">{currentPreset.name}</h1>
                <p className="text-sm text-slate-400">{currentPreset.description}</p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button
                onClick={() => setPaletteOpen(!paletteOpen)}
                variant="outline"
                className="border-slate-600 text-slate-300 hover:bg-slate-800"
              >
                <LayoutTemplate className="mr-2 h-4 w-4" />
                {paletteOpen ? 'Hide' : 'Show'} Palette
              </Button>

              <Button
                onClick={handleAddNode}
                className="bg-blue-600 hover:bg-blue-700 text-white"
              >
                <Plus className="mr-2 h-4 w-4" />
                Add Node
              </Button>
            </div>
          </div>
        </div>

        <div className="flex-1 relative flex">
          <NodePalette isOpen={paletteOpen} onToggle={() => setPaletteOpen(!paletteOpen)} />

          <DropZone>
            <div className="w-full h-full">
              <CanvasWrapper>
                <ReactFlow
                  nodes={createFlowNodes(nodes)}
                  edges={createFlowEdges(edges)}
                  nodeTypes={nodeTypes}
                  edgeTypes={edgeTypes}
                  onNodesClick={onNodesClick}
                  onEdgesClick={onEdgesClick}
                  onConnect={onConnect}
                  fitView
                  attributionPosition="bottom-left"
                />
              </CanvasWrapper>
            </div>
          </DropZone>

          <NodeDetailsSheet
            nodeId={nodes.find(n => n.id === useGLCStore.getState().selectedNodeId)?.id || null}
            open={nodeDetailsOpen}
            onOpenChange={setNodeDetailsOpen}
          />

          <EdgeDetailsSheet
            edgeId={edges.find(e => e.id === useGLCStore.getState().selectedEdgeId)?.id || null}
            open={edgeDetailsOpen}
            onOpenChange={setEdgeDetailsOpen}
          />
        </div>
      </div>
    </ReactFlowProvider>
  );
}
