'use client';

import { ExampleGraph } from '@/lib/glc/lib/examples/example-types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { FileText, Layers, Network } from 'lucide-react';

interface ExampleCardProps {
  example: ExampleGraph;
  viewMode: 'grid' | 'list';
  onClick: () => void;
}

export default function ExampleCard({ example, viewMode, onClick }: ExampleCardProps) {
  const getPresetIcon = () => {
    switch (example.preset) {
      case 'd3fend':
        return <Network className="h-6 w-6" />;
      case 'topo':
        return <Layers className="h-6 w-6" />;
      default:
        return <FileText className="h-6 w-6" />;
    }
  };

  if (viewMode === 'list') {
    return (
      <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={onClick}>
        <CardContent className="p-4">
          <div className="flex items-center gap-4">
            <div className="flex-shrink-0">
              {getPresetIcon()}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h3 className="font-semibold">{example.name}</h3>
                <Badge variant="outline" className="text-xs">
                  {example.metadata.complexity}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                {example.description}
              </p>
            </div>
            <div className="flex-shrink-0 text-right text-sm text-muted-foreground">
              <p>{example.metadata.nodeCount} nodes</p>
              <p>{example.metadata.edgeCount} edges</p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer flex flex-col" onClick={onClick}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            {getPresetIcon()}
            <Badge variant="outline" className="text-xs">
              {example.preset.toUpperCase()}
            </Badge>
          </div>
          <Badge variant="secondary" className="text-xs">
            {example.metadata.complexity}
          </Badge>
        </div>
        <CardTitle className="mt-2">{example.name}</CardTitle>
        <CardDescription className="line-clamp-2">
          {example.description}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1">
        <div className="space-y-2">
          <div className="text-sm text-muted-foreground">
            <Badge variant="outline" className="mr-2">
              {example.category}
            </Badge>
          </div>
          <div className="flex justify-between text-sm text-muted-foreground">
            <span>{example.metadata.nodeCount} nodes</span>
            <span>{example.metadata.edgeCount} edges</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
