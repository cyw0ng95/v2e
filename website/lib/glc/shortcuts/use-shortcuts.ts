'use client';

import { useEffect, useCallback } from 'react';
import { useGLCStore } from '@/lib/glc/store';

interface UseShortcutsOptions {
  onDelete?: () => void;
  onCopy?: () => void;
  onPaste?: () => void;
  onFitView?: () => void;
  onZoomIn?: () => void;
  onZoomOut?: () => void;
  onToggleHelp?: () => void;
}

export function useShortcuts(options: UseShortcutsOptions = {}) {
  const { canUndo, canRedo, undo, redo } = useGLCStore();

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Ignore if typing in input/textarea
      if (
        event.target instanceof HTMLInputElement ||
        event.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      const { key, ctrlKey, metaKey, shiftKey } = event;
      const modKey = ctrlKey || metaKey;

      // Undo: Ctrl/Cmd + Z
      if (modKey && !shiftKey && key === 'z') {
        event.preventDefault();
        if (canUndo) {
          undo();
        }
        return;
      }

      // Redo: Ctrl/Cmd + Shift + Z or Ctrl/Cmd + Y
      if (modKey && (shiftKey ? key === 'z' : key === 'y')) {
        event.preventDefault();
        if (canRedo) {
          redo();
        }
        return;
      }

      // Delete: Delete or Backspace
      if (key === 'Delete' || key === 'Backspace') {
        event.preventDefault();
        options.onDelete?.();
        return;
      }

      // Copy: Ctrl/Cmd + C
      if (modKey && key === 'c') {
        // Let React Flow handle node/edge copy
        return;
      }

      // Paste: Ctrl/Cmd + V
      if (modKey && key === 'v') {
        // Let React Flow handle paste
        return;
      }

      // Fit View: F
      if (key === 'f' && !modKey) {
        event.preventDefault();
        options.onFitView?.();
        return;
      }

      // Zoom In: + or =
      if ((key === '+' || key === '=') && !modKey) {
        event.preventDefault();
        options.onZoomIn?.();
        return;
      }

      // Zoom Out: -
      if (key === '-' && !modKey) {
        event.preventDefault();
        options.onZoomOut?.();
        return;
      }

      // Help: ?
      if (key === '?' || (shiftKey && key === '/')) {
        event.preventDefault();
        options.onToggleHelp?.();
        return;
      }

      // Escape: Close dialogs
      if (key === 'Escape') {
        // Handled by individual components
        return;
      }
    },
    [canUndo, canRedo, undo, redo, options]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);

  return {
    canUndo,
    canRedo,
    undo,
    redo,
  };
}

// Shortcut definitions for help dialog
export const SHORTCUTS = [
  { key: 'Ctrl+Z', description: 'Undo' },
  { key: 'Ctrl+Shift+Z', description: 'Redo' },
  { key: 'Ctrl+Y', description: 'Redo' },
  { key: 'Delete', description: 'Delete selected' },
  { key: 'Ctrl+C', description: 'Copy' },
  { key: 'Ctrl+V', description: 'Paste' },
  { key: 'F', description: 'Fit view' },
  { key: '+', description: 'Zoom in' },
  { key: '-', description: 'Zoom out' },
  { key: '?', description: 'Show shortcuts' },
  { key: 'Esc', description: 'Close dialog' },
] as const;
