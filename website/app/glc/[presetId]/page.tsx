'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useGLCStore } from '@/lib/glc/store';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Loader2, Plus } from 'lucide-react';
import CanvasWrapper from '@/components/canvas/canvas-wrapper';
import { nodeTypes, createFlowNodes } from '@/components/canvas/node-factory';
import { edgeTypes, createFlowEdges } from '@/components/canvas/edge-factory';
import { ReactFlowProvider } from '@xyflow/react';

export default function CanvasPage() {
  const params = useParams();
  const router = useRouter();
  const { currentPreset, setCurrentPreset, getPresetById, nodes, edges } = useGLCStore();

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

            <Button
              onClick={handleAddNode}
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Node
            </Button>
          </div>
        </div>

        <div className="flex-1 relative">
          <CanvasWrapper>
            <ReactFlow
              nodes={createFlowNodes(nodes)}
              edges={createFlowEdges(edges)}
              nodeTypes={nodeTypes}
              edgeTypes={edgeTypes}
              fitView
              attributionPosition="bottom-left"
            />
          </CanvasWrapper>
        </div>
      </div>
    </ReactFlowProvider>
  );
}
