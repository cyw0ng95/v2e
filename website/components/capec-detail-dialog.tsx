"use client"

import React from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog'
import { Button } from './ui/button'
import { useCAPEC } from '@/lib/hooks'
import { useViewLearnMode } from '@/contexts/ViewLearnContext'
import BookmarkStar from '@/components/bookmark-star'
import { generateURN } from '@/lib/utils'
import type { CAPECItem } from '@/lib/types'

export interface CAPECDetailDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  capecId?: string | null
  initial?: CAPECItem | null
}

export function CAPECDetailDialog({ open, onOpenChange, capecId, initial }: CAPECDetailDialogProps) {
  const { data, isLoading, error } = useCAPEC(capecId || undefined)
  const { mode } = useViewLearnMode()
  const isLearnMode = mode === 'learn'
  const c: CAPECItem | null = (data as any) ?? initial ?? null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={isLearnMode ? 'learn-focus-area' : ''}>
        <DialogHeader className="flex items-start justify-between gap-4 pb-4">
          <div className="flex-1 min-w-0">
            <DialogTitle className={`text-lg font-semibold pr-2 break-words ${isLearnMode ? 'learn-enhanced' : ''}`}>
              {c?.name ?? (capecId ?? 'CAPEC Detail')}
            </DialogTitle>
            <DialogDescription className={`break-words ${isLearnMode ? 'learn-enhanced' : ''}`}>
              {c?.summary}
            </DialogDescription>
            {c && (
              <div className="mt-2">
                <div className="text-xs text-muted-foreground font-mono bg-muted px-2 py-1 rounded inline-block">
                  URN: {generateURN('CAPEC', c.id)}
                </div>
              </div>
            )}
          </div>
          <BookmarkStar
            itemId={c?.id || capecId || ''}
            itemType="CAPEC"
            itemTitle={c?.name || 'CAPEC'}
            itemDescription={c?.summary || ''}
            viewMode={mode}
          />
        </DialogHeader>

        <div className={`mt-2 max-h-[60vh] overflow-auto space-y-4 text-sm ${isLearnMode ? 'learn-enhanced' : ''}`}>
          {isLoading && (
            <div className="text-muted-foreground">Loading...</div>
          )}

          {error && (
            <div className="text-destructive">Error loading CAPEC details</div>
          )}

          {!isLoading && !error && c && (
            <div className="space-y-3">
              <div>
                <div className="text-xs text-muted-foreground">ID</div>
                <div className="font-mono">{c.id}</div>
              </div>

              <div>
                <div className="text-xs text-muted-foreground">Name</div>
                <div>{c.name}</div>
              </div>

              {c.description && (
                <div>
                  <div className="text-xs text-muted-foreground">Description</div>
                  <div className={`whitespace-pre-wrap ${isLearnMode ? 'learn-text-lg' : ''}`}>{c.description}</div>
                </div>
              )}

              {c.summary && (
                <div>
                  <div className="text-xs text-muted-foreground">Summary</div>
                  <div className={isLearnMode ? 'learn-text-lg' : ''}>{c.summary}</div>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                {c.status && (
                  <div>
                    <div className="text-xs text-muted-foreground">Status</div>
                    <div>{c.status}</div>
                  </div>
                )}

                {c.likelihood && (
                  <div>
                    <div className="text-xs text-muted-foreground">Likelihood</div>
                    <div>{c.likelihood}</div>
                  </div>
                )}

                {c.typicalSeverity && (
                  <div>
                    <div className="text-xs text-muted-foreground">Typical Severity</div>
                    <div>{c.typicalSeverity}</div>
                  </div>
                )}
              </div>

              {c.relatedWeaknesses && c.relatedWeaknesses.length > 0 && (
                <div>
                  <div className="text-xs text-muted-foreground">Related Weaknesses</div>
                  <ul className="list-disc ml-5">
                    {c.relatedWeaknesses.map((rw, i) => (
                      <li key={i}>{rw.cweId ?? 'N/A'}</li>
                    ))}
                  </ul>
                </div>
              )}

              {c.references && c.references.length > 0 && (
                <div>
                  <div className="text-xs text-muted-foreground">References</div>
                  <ul className="list-disc ml-5">
                    {c.references.map((r, i) => (
                      <li key={i}><a className="text-primary underline" href={r.url} target="_blank" rel="noreferrer">{r.source ?? r.url}</a></li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>

        <div className="mt-6 flex justify-end">
          <Button variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export default CAPECDetailDialog
