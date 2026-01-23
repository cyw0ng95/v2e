"use client"

import React from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog'
import { Button } from './ui/button'
import { useCAPEC } from '@/lib/hooks'
import type { CAPECItem } from '@/lib/types'

export interface CAPECDetailDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  capecId?: string | null
  initial?: CAPECItem | null
}

export function CAPECDetailDialog({ open, onOpenChange, capecId, initial }: CAPECDetailDialogProps) {
  const { data, isLoading, error } = useCAPEC(capecId || undefined)
  const c: CAPECItem | null = (data as any) ?? initial ?? null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{c?.name ?? (capecId ?? 'CAPEC Detail')}</DialogTitle>
          <DialogDescription>{c?.summary}</DialogDescription>
        </DialogHeader>

        <div className="mt-2 max-h-[60vh] overflow-auto space-y-4 text-sm">
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
                  <div className="whitespace-pre-wrap">{c.description}</div>
                </div>
              )}

              {c.summary && (
                <div>
                  <div className="text-xs text-muted-foreground">Summary</div>
                  <div>{c.summary}</div>
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
