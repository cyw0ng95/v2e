'use client';

import { useState } from 'react';
import Link from 'next/link';
import {
  Undo2,
  Redo2,
  ZoomIn,
  ZoomOut,
  Maximize,
  Grid3X3,
  Map,
  Save,
  Download,
  Share2,
  HelpCircle,
  Home,
} from 'lucide-react';
import { useGLCStore } from '@/lib/glc/store';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import type { CanvasPreset } from '@/lib/glc/types';

interface CanvasToolbarProps {
  preset: CanvasPreset;
  graphName?: string;
  onShowShortcuts?: () => void;
  onShowExport?: () => void;
  onShowShare?: () => void;
}

export function CanvasToolbar({ preset, graphName, onShowShortcuts, onShowExport, onShowShare }: CanvasToolbarProps) {
  const { canUndo, canRedo, undo, redo, graph } = useGLCStore();
  const [showGrid, setShowGrid] = useState(true);
  const [showMinimap, setShowMinimap] = useState(true);

  const theme = preset.theme;

  const handleSave = () => {
    if (!graph) return;
    const json = JSON.stringify(graph, null, 2);
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${graph.metadata.name || 'graph'}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div
      className="flex items-center gap-1 px-4 py-2 rounded-lg shadow-lg border"
      style={{
        backgroundColor: theme.surface,
        borderColor: theme.border,
      }}
    >
      {/* Home */}
      <Button variant="ghost" size="icon" asChild className="h-8 w-8">
        <Link href="/glc">
          <Home className="w-4 h-4" style={{ color: theme.textMuted }} />
        </Link>
      </Button>

      <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

      {/* Graph Name */}
      <span
        className="text-sm font-medium px-2 min-w-[100px] max-w-[200px] truncate"
        style={{ color: theme.text }}
      >
        {graphName || 'Untitled Graph'}
      </span>

      <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

      {/* Undo/Redo */}
      <Button
        variant="ghost"
        size="icon"
        className="h-8 w-8"
        disabled={!canUndo}
        onClick={() => undo()}
      >
        <Undo2 className="w-4 h-4" style={{ color: canUndo ? theme.text : theme.textMuted }} />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className="h-8 w-8"
        disabled={!canRedo}
        onClick={() => redo()}
      >
        <Redo2 className="w-4 h-4" style={{ color: canRedo ? theme.text : theme.textMuted }} />
      </Button>

      <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

      {/* Zoom */}
      <Button variant="ghost" size="icon" className="h-8 w-8">
        <ZoomIn className="w-4 h-4" style={{ color: theme.text }} />
      </Button>
      <Button variant="ghost" size="icon" className="h-8 w-8">
        <ZoomOut className="w-4 h-4" style={{ color: theme.text }} />
      </Button>
      <Button variant="ghost" size="icon" className="h-8 w-8">
        <Maximize className="w-4 h-4" style={{ color: theme.text }} />
      </Button>

      <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

      {/* View toggles */}
      <Button
        variant={showGrid ? 'secondary' : 'ghost'}
        size="icon"
        className="h-8 w-8"
        onClick={() => setShowGrid(!showGrid)}
      >
        <Grid3X3 className="w-4 h-4" style={{ color: theme.text }} />
      </Button>
      <Button
        variant={showMinimap ? 'secondary' : 'ghost'}
        size="icon"
        className="h-8 w-8"
        onClick={() => setShowMinimap(!showMinimap)}
      >
        <Map className="w-4 h-4" style={{ color: theme.text }} />
      </Button>

      <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

      {/* Save */}
      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleSave}>
        <Save className="w-4 h-4" style={{ color: theme.text }} />
      </Button>

      {/* Export */}
      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onShowExport}>
        <Download className="w-4 h-4" style={{ color: theme.text }} />
      </Button>

      {/* Share */}
      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onShowShare}>
        <Share2 className="w-4 h-4" style={{ color: theme.text }} />
      </Button>

      <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

      {/* Help */}
      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onShowShortcuts}>
        <HelpCircle className="w-4 h-4" style={{ color: theme.textMuted }} />
      </Button>

      {/* Preset indicator */}
      <div
        className="ml-2 px-2 py-0.5 rounded text-xs font-medium"
        style={{
          backgroundColor: theme.primary + '20',
          color: theme.primary,
        }}
      >
        {preset.meta.name}
      </div>
    </div>
  );
}
