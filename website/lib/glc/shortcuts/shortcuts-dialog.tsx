'use client';

import { Keyboard, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useGLCStore } from '@/lib/glc/store';
import { SHORTCUTS } from './use-shortcuts';

interface ShortcutsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ShortcutsDialog({ open, onOpenChange }: ShortcutsDialogProps) {
  const currentPreset = useGLCStore((state) => state.currentPreset);

  if (!currentPreset) return null;

  const theme = currentPreset.theme;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-md"
        style={{
          backgroundColor: theme.surface,
          borderColor: theme.border,
        }}
      >
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2" style={{ color: theme.text }}>
            <Keyboard className="w-5 h-5" />
            Keyboard Shortcuts
          </DialogTitle>
        </DialogHeader>

        <div className="grid gap-2 py-4">
          {SHORTCUTS.map((shortcut) => (
            <div
              key={shortcut.key}
              className="flex items-center justify-between py-1"
            >
              <span className="text-sm" style={{ color: theme.textMuted }}>
                {shortcut.description}
              </span>
              <kbd
                className="px-2 py-1 rounded text-xs font-mono"
                style={{
                  backgroundColor: theme.background,
                  border: `1px solid ${theme.border}`,
                  color: theme.text,
                }}
              >
                {shortcut.key}
              </kbd>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
