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

interface GraphListProps {
  graphs: GLCGraph[];
  onDelete: (graphId: string) => void;
  onDuplicate: (graph: GLCGraph) => void;
  onExport: (graph: GLCGraph) => void;
  onShare: (graph: GLCGraph) => void;
}

export function GraphList({ graphs, onDelete, onDuplicate, onExport, onShare }: GraphListProps) {
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const handleDelete = async (graphId: string) => {
    if (deletingId) return;
    setDeletingId(graphId);
    try {
      await onDelete(graphId);
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="divide-y divide-border">
      {graphs.map((graph) => {
        const tags = graph.tags ? graph.tags.split(',').filter(Boolean) : [];
        const formattedDate = new Date(graph.updated_at).toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
          year: 'numeric',
        });

        return (
          <div
            key={graph.graph_id}
            className="flex items-center gap-4 p-4 hover:bg-surface/50 transition-colors group"
          >
            {/* Thumbnail */}
            <div className="w-24 h-16 flex-shrink-0 rounded border border-border/50 overflow-hidden bg-surface/50">
              {graph.thumbnail ? (
                <img
                  src={graph.thumbnail}
                  alt={graph.name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-indigo-500/10 to-purple-500/10">
                  <FileJson className="w-6 h-6 text-textMuted/50" />
                </div>
              )}
            </div>

            {/* Info */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <h3 className="font-semibold text-text truncate">{graph.name}</h3>
                <Badge variant="secondary" className="text-xs flex-shrink-0">
                  {graph.preset_id}
                </Badge>
              </div>
              {graph.description && (
                <p className="text-sm text-textMuted truncate mb-1">{graph.description}</p>
              )}
              <div className="flex items-center gap-4 text-xs text-textMuted">
                <span className="flex items-center">
                  <Calendar className="w-3 h-3 mr-1" />
                  {formattedDate}
                </span>
                {tags.length > 0 && (
                  <span className="flex items-center">
                    <Tag className="w-3 h-3 mr-1" />
                    {tags.slice(0, 2).join(', ')}
                    {tags.length > 2 && ` +${tags.length - 2}`}
                  </span>
                )}
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-2">
              <Button asChild size="sm" variant="outline" className="hidden group-hover:flex">
                <Link href={`/glc/${graph.preset_id}?graphId=${graph.graph_id}`}>
                  <ExternalLink className="w-4 h-4 mr-1" />
                  Open
                </Link>
              </Button>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
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
                    onClick={() => handleDelete(graph.graph_id)}
                    disabled={deletingId === graph.graph_id}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    {deletingId === graph.graph_id ? 'Deleting...' : 'Delete'}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        );
      })}
    </div>
  );
}
