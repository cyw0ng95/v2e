"use client"

import React from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog'
import { Badge } from './ui/badge'
import { useSSGProfile, useSSGRule } from '@/lib/hooks'

export interface SSGDetailDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  itemId?: string | null
  itemType?: 'profile' | 'rule'
  initial?: unknown
}

export default function SSGDetailDialog({ open, onOpenChange, itemId, itemType = 'profile', initial }: SSGDetailDialogProps) {
  const { data: profileData, isLoading: profileLoading, error: profileError } = useSSGProfile(
    itemType === 'profile' ? (itemId || undefined) : undefined
  )
  const { data: ruleData, isLoading: ruleLoading, error: ruleError } = useSSGRule(
    itemType === 'rule' ? (itemId || undefined) : undefined
  )

  const isLoading = itemType === 'profile' ? profileLoading : ruleLoading
  const error = itemType === 'profile' ? profileError : ruleError
  const item = itemType === 'profile' ? (profileData?.profile ?? initial) : (ruleData?.rule ?? initial)

  const getSeverityVariant = (severity: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (severity?.toLowerCase()) {
      case 'critical':
        return 'destructive';
      case 'high':
        return 'default';
      case 'medium':
        return 'secondary';
      case 'low':
        return 'outline';
      default:
        return 'outline';
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>{item?.title ?? (itemId ?? 'SSG Detail')}</DialogTitle>
          <DialogDescription>
            {itemType === 'profile' ? 'Security Profile' : 'Security Rule'}
          </DialogDescription>
        </DialogHeader>

        <div className="mt-2 max-h-[60vh] overflow-auto space-y-4 text-sm">
          {isLoading && (
            <div className="text-muted-foreground">Loading...</div>
          )}

          {error && (
            <div className="text-destructive">Error loading SSG details</div>
          )}

          {!isLoading && !error && item && itemType === 'profile' && (
            <div className="space-y-3">
              <div>
                <div className="text-xs text-muted-foreground">Profile ID</div>
                <div className="font-mono text-xs break-all">{item.id}</div>
              </div>

              <div>
                <div className="text-xs text-muted-foreground">Title</div>
                <div className="font-medium">{item.title}</div>
              </div>

              {item.description && (
                <div>
                  <div className="text-xs text-muted-foreground">Description</div>
                  <div className="text-sm">{item.description}</div>
                </div>
              )}

              {item.ruleCount !== undefined && (
                <div>
                  <div className="text-xs text-muted-foreground">Number of Rules</div>
                  <Badge variant="outline">{item.ruleCount} rules</Badge>
                </div>
              )}
            </div>
          )}

          {!isLoading && !error && item && itemType === 'rule' && (
            <div className="space-y-3">
              <div>
                <div className="text-xs text-muted-foreground">Rule ID</div>
                <div className="font-mono text-xs break-all">{item.id}</div>
              </div>

              <div>
                <div className="text-xs text-muted-foreground">Title</div>
                <div className="font-medium">{item.title}</div>
              </div>

              {item.severity && (
                <div>
                  <div className="text-xs text-muted-foreground">Severity</div>
                  <Badge variant={getSeverityVariant(item.severity)}>
                    {item.severity}
                  </Badge>
                </div>
              )}

              {item.description && (
                <div>
                  <div className="text-xs text-muted-foreground">Description</div>
                  <div className="text-sm whitespace-pre-wrap">{item.description}</div>
                </div>
              )}

              {item.rationale && (
                <div>
                  <div className="text-xs text-muted-foreground">Rationale</div>
                  <div className="text-sm whitespace-pre-wrap">{item.rationale}</div>
                </div>
              )}

              {item.warning && (
                <div>
                  <div className="text-xs text-muted-foreground">Warning</div>
                  <div className="text-sm text-amber-600 whitespace-pre-wrap">{item.warning}</div>
                </div>
              )}
            </div>
          )}

          {!isLoading && !error && !item && (
            <div className="text-muted-foreground">No data available</div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
