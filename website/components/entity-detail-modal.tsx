'use client';

import React from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { useViewLearnMode } from '@/contexts/ViewLearnContext';
import BookmarkStar from '@/components/bookmark-star';

export type EntityType = 'CVE' | 'CWE' | 'CAPEC' | 'ATTACK' | 'ASVS' | 'SSG';

interface EntityDetailModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  entityType: EntityType;
  entityId: string;
  title: string;
  description: string;
  metadata?: Record<string, any>;
  children?: React.ReactNode;
  urn?: string;
}

export function EntityDetailModal({
  open,
  onOpenChange,
  entityType,
  entityId,
  title,
  description,
  metadata,
  children,
  urn
}: EntityDetailModalProps) {
  const { mode } = useViewLearnMode();
  const isLearnMode = mode === 'learn';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={`max-w-4xl max-h-[85vh] overflow-auto ${
          isLearnMode ? 'learn-focus-area' : ''
        }`}
      >
        <DialogHeader className="flex items-start justify-between gap-4 pb-4">
          <div className="flex-1 min-w-0">
            <DialogTitle
              className={`text-xl font-semibold pr-2 break-words ${
                isLearnMode ? 'learn-enhanced' : ''
              }`}
            >
              {title}
            </DialogTitle>
            <p
              className={`text-sm text-muted-foreground mt-2 break-words ${
                isLearnMode ? 'learn-enhanced' : ''
              }`}
            >
              {description}
            </p>
            {urn && (
              <div className="text-xs text-muted-foreground font-mono bg-muted px-2 py-1 rounded mt-2 inline-block">
                URN: {urn}
              </div>
            )}
          </div>
          <div className="flex flex-col items-end gap-2 shrink-0">
            <Badge variant="outline">{entityType}</Badge>
            <BookmarkStar
              itemId={entityId}
              itemType={entityType}
              itemTitle={title}
              itemDescription={description}
              viewMode={mode}
            />
          </div>
        </DialogHeader>
        <div className={`mt-4 ${isLearnMode ? 'learn-enhanced' : ''}`}>
          {children}
        </div>
      </DialogContent>
    </Dialog>
  );
}
