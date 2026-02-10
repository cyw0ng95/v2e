'use client';

import { useState } from 'react';
import Link from 'next/link';
import {
  MoreVertical,
  ExternalLink,
  Copy,
  Trash2,
  Share2,
  Download,
  Calendar,
  Tag,
  FileJson,
} from 'lucide-react';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { GLCGraph } from '@/lib/types';

interface GraphCardProps {
  graph: GLCGraph;
  onDelete: (graphId: string) => void;
  onDuplicate: (graph: GLCGraph) => void;
  onExport: (graph: GLCGraph) => void;
  onShare: (graph: GLCGraph) => void;
}

export function GraphCard({ graph, onDelete, onDuplicate, onExport, onShare }: GraphCardProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [imageError, setImageError] = useState(false);

  const formattedDate = new Date(graph.updated_at).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

  const tags = graph.tags ? graph.tags.split(',').filter(Boolean) : [];

  const handleDelete = async () => {
    if (isDeleting) return;
    setIsDeleting(true);
    try {
      await onDelete(graph.graph_id);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Card className="group hover:border-accent/50 transition-all hover:shadow-md">
      {/* Thumbnail */}
      <div className="relative aspect-video bg-surface/50 overflow-hidden">
        {graph.thumbnail && !imageError ? (
          <img
            src={graph.thumbnail}
            alt={graph.name}
            className="w-full h-full object-cover"
            onError={() => setImageError(true)}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-indigo-500/10 to-purple-500/10">
            <FileJson className="w-12 h-12 text-textMuted/50" />
          </div>
        )}
        {/* Quick Actions Overlay */}
        <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
          <Button asChild size="sm" variant="secondary">
            <Link href={`/glc/${graph.preset_id}?graphId=${graph.graph_id}`}>
              <ExternalLink className="w-4 h-4 mr-1" />
              Open
            </Link>
          </Button>
        </div>
        {/* Dropdown Menu */}
        <div className="absolute top-2 right-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="secondary"
                size="icon"
                className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem asChild>
                <Link href={`/glc/${graph.preset_id}?graphId=${graph.graph_id}`}>
                  <ExternalLink className="mr-2 h-4 w-4" />
                  Open
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDuplicate(graph)}>
                <Copy className="mr-2 h-4 w-4" />
                Duplicate
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onShare(graph)}>
                <Share2 className="mr-2 h-4 w-4" />
                Share
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onExport(graph)}>
                <Download className="mr-2 h-4 w-4" />
                Export
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={handleDelete}
                disabled={isDeleting}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                {isDeleting ? 'Deleting...' : 'Delete'}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        {/* Preset Badge */}
        <div className="absolute top-2 left-2">
          <Badge variant="secondary" className="text-xs">
            {graph.preset_id}
          </Badge>
        </div>
      </div>

      {/* Content */}
      <CardContent className="p-4">
        <h3 className="font-semibold text-text truncate mb-1">{graph.name}</h3>
        {graph.description && (
          <p className="text-sm text-textMuted line-clamp-2 mb-2">{graph.description}</p>
        )}
        <div className="flex items-center text-xs text-textMuted">
          <Calendar className="w-3 h-3 mr-1" />
          {formattedDate}
        </div>
      </CardContent>

      {/* Tags Footer */}
      {tags.length > 0 && (
        <CardFooter className="px-4 pb-4 pt-0">
          <div className="flex flex-wrap gap-1">
            {tags.slice(0, 3).map((tag, index) => (
              <Badge key={index} variant="outline" className="text-xs">
                <Tag className="w-3 h-3 mr-1" />
                {tag.trim()}
              </Badge>
            ))}
            {tags.length > 3 && (
              <Badge variant="outline" className="text-xs">
                +{tags.length - 3}
              </Badge>
            )}
          </div>
        </CardFooter>
      )}
    </Card>
  );
}
