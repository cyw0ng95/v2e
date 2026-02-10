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
  MoreHorizontal,
} from 'lucide-react';
import { useGLCStore } from '@/lib/glc/store';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { useResponsive, TOUCH_TARGET_SIZE } from '@/lib/glc/responsive';
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
  const [showMore, setShowMore] = useState(false);
  const { isMobile, isTablet } = useResponsive();

  const theme = preset.theme;

  // Responsive button size - 44px for mobile, 32px for desktop
  const buttonClass = isMobile ? 'h-11 w-11' : 'h-8 w-8';
  const iconSize = isMobile ? 'w-5 h-5' : 'w-4 h-4';

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

  // Mobile toolbar - minimal buttons with overflow menu
  if (isMobile) {
    return (
      <div
        className="flex items-center gap-2 px-2 py-2 rounded-lg shadow-lg border"
        style={{
          backgroundColor: theme.surface,
          borderColor: theme.border,
        }}
      >
        {/* Home */}
        <Button variant="ghost" asChild className={buttonClass} style={{ minWidth: TOUCH_TARGET_SIZE, minHeight: TOUCH_TARGET_SIZE }}>
          <Link href="/glc">
            <Home className={iconSize} style={{ color: theme.textMuted }} />
          </Link>
        </Button>

        {/* Graph Name - truncated on mobile */}
        <span
          className="text-sm font-medium px-2 flex-1 min-w-0 truncate text-center"
          style={{ color: theme.text }}
        >
          {graphName || 'Untitled'}
        </span>

        {/* Undo/Redo */}
        <Button
          variant="ghost"
          className={buttonClass}
          style={{ minWidth: TOUCH_TARGET_SIZE, minHeight: TOUCH_TARGET_SIZE }}
          disabled={!canUndo}
          onClick={() => undo()}
        >
          <Undo2 className={iconSize} style={{ color: canUndo ? theme.text : theme.textMuted }} />
        </Button>
        <Button
          variant="ghost"
          className={buttonClass}
          style={{ minWidth: TOUCH_TARGET_SIZE, minHeight: TOUCH_TARGET_SIZE }}
          disabled={!canRedo}
          onClick={() => redo()}
        >
          <Redo2 className={iconSize} style={{ color: canRedo ? theme.text : theme.textMuted }} />
        </Button>

        {/* More options */}
        <div className="relative">
          <Button
            variant="ghost"
            className={buttonClass}
            style={{ minWidth: TOUCH_TARGET_SIZE, minHeight: TOUCH_TARGET_SIZE }}
            onClick={() => setShowMore(!showMore)}
          >
            <MoreHorizontal className={iconSize} style={{ color: theme.text }} />
          </Button>

          {/* Overflow menu dropdown */}
          {showMore && (
            <>
              <div
                className="fixed inset-0 z-40"
                onClick={() => setShowMore(false)}
              />
              <div
                className="absolute right-0 top-full mt-1 z-50 flex flex-col gap-1 p-2 rounded-lg shadow-lg border min-w-[160px]"
                style={{
                  backgroundColor: theme.surface,
                  borderColor: theme.border,
                }}
              >
                <Button
                  variant="ghost"
                  className="justify-start h-11 px-3"
                  onClick={() => { setShowGrid(!showGrid); setShowMore(false); }}
                >
                  <Grid3X3 className="w-5 h-5 mr-3" style={{ color: theme.text }} />
                  <span style={{ color: theme.text }}>{showGrid ? 'Hide' : 'Show'} Grid</span>
                </Button>
                <Button
                  variant="ghost"
                  className="justify-start h-11 px-3"
                  onClick={() => { setShowMinimap(!showMinimap); setShowMore(false); }}
                >
                  <Map className="w-5 h-5 mr-3" style={{ color: theme.text }} />
                  <span style={{ color: theme.text }}>{showMinimap ? 'Hide' : 'Show'} Minimap</span>
                </Button>
                <Separator orientation="horizontal" className="my-1" style={{ backgroundColor: theme.border }} />
                <Button
                  variant="ghost"
                  className="justify-start h-11 px-3"
                  onClick={() => { handleSave(); setShowMore(false); }}
                >
                  <Save className="w-5 h-5 mr-3" style={{ color: theme.text }} />
                  <span style={{ color: theme.text }}>Save</span>
                </Button>
                <Button
                  variant="ghost"
                  className="justify-start h-11 px-3"
                  onClick={() => { onShowExport?.(); setShowMore(false); }}
                >
                  <Download className="w-5 h-5 mr-3" style={{ color: theme.text }} />
                  <span style={{ color: theme.text }}>Export</span>
                </Button>
                <Button
                  variant="ghost"
                  className="justify-start h-11 px-3"
                  onClick={() => { onShowShare?.(); setShowMore(false); }}
                >
                  <Share2 className="w-5 h-5 mr-3" style={{ color: theme.text }} />
                  <span style={{ color: theme.text }}>Share</span>
                </Button>
                <Separator orientation="horizontal" className="my-1" style={{ backgroundColor: theme.border }} />
                <Button
                  variant="ghost"
                  className="justify-start h-11 px-3"
                  onClick={() => { onShowShortcuts?.(); setShowMore(false); }}
                >
                  <HelpCircle className="w-5 h-5 mr-3" style={{ color: theme.textMuted }} />
                  <span style={{ color: theme.text }}>Help</span>
                </Button>
              </div>
            </>
          )}
        </div>
      </div>
    );
  }

  // Tablet toolbar - condensed layout
  if (isTablet) {
    return (
      <div
        className="flex items-center gap-1 px-3 py-2 rounded-lg shadow-lg border"
        style={{
          backgroundColor: theme.surface,
          borderColor: theme.border,
        }}
      >
        {/* Home */}
        <Button variant="ghost" size="icon" asChild className="h-9 w-9">
          <Link href="/glc">
            <Home className="w-4 h-4" style={{ color: theme.textMuted }} />
          </Link>
        </Button>

        <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

        {/* Graph Name */}
        <span
          className="text-sm font-medium px-2 min-w-[80px] max-w-[150px] truncate"
          style={{ color: theme.text }}
        >
          {graphName || 'Untitled Graph'}
        </span>

        <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

        {/* Undo/Redo */}
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9"
          disabled={!canUndo}
          onClick={() => undo()}
        >
          <Undo2 className="w-4 h-4" style={{ color: canUndo ? theme.text : theme.textMuted }} />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9"
          disabled={!canRedo}
          onClick={() => redo()}
        >
          <Redo2 className="w-4 h-4" style={{ color: canRedo ? theme.text : theme.textMuted }} />
        </Button>

        <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

        {/* Zoom */}
        <Button variant="ghost" size="icon" className="h-9 w-9">
          <ZoomIn className="w-4 h-4" style={{ color: theme.text }} />
        </Button>
        <Button variant="ghost" size="icon" className="h-9 w-9">
          <ZoomOut className="w-4 h-4" style={{ color: theme.text }} />
        </Button>

        <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

        {/* View toggles */}
        <Button
          variant={showGrid ? 'secondary' : 'ghost'}
          size="icon"
          className="h-9 w-9"
          onClick={() => setShowGrid(!showGrid)}
        >
          <Grid3X3 className="w-4 h-4" style={{ color: theme.text }} />
        </Button>

        <Separator orientation="vertical" className="h-6 mx-1" style={{ backgroundColor: theme.border }} />

        {/* Save */}
        <Button variant="ghost" size="icon" className="h-9 w-9" onClick={handleSave}>
          <Save className="w-4 h-4" style={{ color: theme.text }} />
        </Button>

        {/* Export */}
        <Button variant="ghost" size="icon" className="h-9 w-9" onClick={onShowExport}>
          <Download className="w-4 h-4" style={{ color: theme.text }} />
        </Button>

        {/* Help */}
        <Button variant="ghost" size="icon" className="h-9 w-9" onClick={onShowShortcuts}>
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

  // Desktop toolbar - full layout (original)
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
