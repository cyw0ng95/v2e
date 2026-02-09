'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import ExampleGallery from './example-gallery';
import { ExampleGraph } from '@/lib/glc/lib/examples/example-types';
import { Node, Edge, useReactFlow } from '@xyflow/react';

interface ExampleDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  preset?: string;
}

export default function ExampleDialog({ open, onOpenChange, preset }: ExampleDialogProps) {
  const { setNodes, setEdges } = useReactFlow();

  function handleOpenExample(example: ExampleGraph) {
    const nodes: Node[] = example.nodes.map((node) => ({
      id: node.id,
      type: node.type,
      position: node.position,
      data: node.data,
    }));

    const edges: Edge[] = example.edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: edge.type,
      data: edge.data,
    }));

    setNodes(nodes);
    setEdges(edges);

    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Load Example Graph</DialogTitle>
        </DialogHeader>
        <ExampleGallery
          preset={preset}
          onOpenExample={handleOpenExample}
          onClose={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
