'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle 
} from '@/components/ui/dialog';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';
import { 
  useAttackTechniques, 
  useAttackTactics, 
  useAttackMitigations, 
  useAttackSoftware, 
  useAttackGroups,
  useAttackTechnique,
  useAttackTactic,
  useAttackMitigation,
  useAttackSoftwareById,
  useAttackGroupById
} from '@/lib/hooks';

interface AttackDetailDialogProps {
  open: boolean;
  onClose: () => void;
  item: any;
  type: 'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups';
}

interface RelatedItem {
  id: string;
  name: string;
  type: string;
}

export function AttackDetailContent({ item, type, onBack }: { item: any; type: string; onBack?: () => void }) {
  const [currentItem, setCurrentItem] = useState<any>(item || {});
  const [navigationHistory, setNavigationHistory] = useState<Array<any>>([]);

  useEffect(() => {
    setCurrentItem(item || {});
    setNavigationHistory([]);
  }, [item]);

  const getReadableType = (t: string) => {
    switch(t) {
      case 'techniques': return 'Technique';
      case 'tactics': return 'Tactic';
      case 'mitigations': return 'Mitigation';
      case 'software': return 'Software';
      case 'groups': return 'Group';
      default: return t;
    }
  };

  const getRelatedItemsByType = (current: any, itemType: string) => {
    if (!current) return [];
    const ct = (current.type || 'techniques');
    switch(ct) {
      case 'techniques':
        switch(itemType) {
          case 'tactics':
            return current.tactic ? [{ id: current.tactic, name: current.tacticName || current.tactic, type: 'tactics' }] : [];
          case 'mitigations':
            return [
              { id: 'M1036', name: 'Credential Access Prevention', type: 'mitigations' },
              { id: 'M1053', name: 'Run Command', type: 'mitigations' },
            ].slice(0,2);
          case 'software':
            return [
              { id: 'S0001', name: 'Compiled HTML File', type: 'software' },
              { id: 'S0002', name: 'RegSvr32', type: 'software' },
            ].slice(0,2);
          case 'groups':
            return [
              { id: 'G0001', name: 'APT1', type: 'groups' },
              { id: 'G0006', name: 'APT28', type: 'groups' },
            ].slice(0,2);
          case 'techniques':
            return [
              { id: 'T1001', name: 'Data Obfuscation', type: 'techniques' },
              { id: 'T1071', name: 'Application Layer Protocol', type: 'techniques' },
            ].slice(0,2);
          default:
            return [];
        }
      case 'tactics':
        if (itemType === 'techniques') {
          return [
            { id: 'T1003', name: 'OS Credential Dumping', type: 'techniques' },
            { id: 'T1005', name: 'Data from Local System', type: 'techniques' },
          ];
        }
        return [];
      case 'mitigations':
        if (itemType === 'techniques') {
          return [
            { id: 'T1003', name: 'OS Credential Dumping', type: 'techniques' },
            { id: 'T1053', name: 'Create or Modify System Process', type: 'techniques' },
          ];
        }
        return [];
      case 'software':
        if (itemType === 'techniques') {
          return [
            { id: 'T1003', name: 'OS Credential Dumping', type: 'techniques' },
            { id: 'T1071', name: 'Application Layer Protocol', type: 'techniques' },
          ];
        } else if (itemType === 'groups') {
          return [
            { id: 'G0001', name: 'APT1', type: 'groups' },
            { id: 'G0006', name: 'APT28', type: 'groups' },
          ];
        }
        return [];
      case 'groups':
        if (itemType === 'software') {
          return [
            { id: 'S0001', name: 'Compiled HTML File', type: 'software' },
            { id: 'S0002', name: 'RegSvr32', type: 'software' },
          ];
        } else if (itemType === 'techniques') {
          return [
            { id: 'T1003', name: 'OS Credential Dumping', type: 'techniques' },
            { id: 'T1071', name: 'Application Layer Protocol', type: 'techniques' },
          ];
        }
        return [];
      default:
        return [];
    }
  };

  const handleViewItem = (related: RelatedItem) => {
    setNavigationHistory([...navigationHistory, currentItem]);
    setCurrentItem({ id: related.id, name: related.name, type: related.type });
  };

  const handleBack = () => {
    const newHistory = [...navigationHistory];
    const last = newHistory.pop();
    if (last) {
      setCurrentItem(last);
      setNavigationHistory(newHistory);
    } else if (onBack) {
      onBack();
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-3">
          <Badge variant="outline">{currentItem.id}</Badge>
          <h2 className="text-lg font-semibold">{currentItem.name || 'Detail'}</h2>
        </div>
        <div className="mt-3">
          <h3 className="text-lg font-semibold">Information</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
            {currentItem.id && (
              <div>
                <h4 className="text-sm font-medium text-muted-foreground">ID</h4>
                <p className="font-mono">{currentItem.id}</p>
              </div>
            )}
            {currentItem.name && (
              <div>
                <h4 className="text-sm font-medium text-muted-foreground">Name</h4>
                <p>{currentItem.name}</p>
              </div>
            )}
            {currentItem.domain && (
              <div>
                <h4 className="text-sm font-medium text-muted-foreground">Domain</h4>
                <p>{currentItem.domain}</p>
              </div>
            )}
            {currentItem.platform && (
              <div>
                <h4 className="text-sm font-medium text-muted-foreground">Platform</h4>
                <p>{currentItem.platform}</p>
              </div>
            )}
            {currentItem.type && (
              <div>
                <h4 className="text-sm font-medium text-muted-foreground">Type</h4>
                <p>{currentItem.type}</p>
              </div>
            )}
            {currentItem.created && (
              <div>
                <h4 className="text-sm font-medium text-muted-foreground">Created</h4>
                <p>{currentItem.created}</p>
              </div>
            )}
            {currentItem.modified && (
              <div>
                <h4 className="text-sm font-medium text-muted-foreground">Modified</h4>
                <p>{currentItem.modified}</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {currentItem.description && (
        <div>
          <h3 className="text-lg font-semibold">Description</h3>
          <p className="whitespace-pre-line">{currentItem.description}</p>
        </div>
      )}

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Related Items</h3>
        {['techniques', 'tactics', 'mitigations', 'software', 'groups'].map((itemType) => {
          const items = getRelatedItemsByType(currentItem, itemType);
          if (items.length === 0) return null;

          return (
            <div key={itemType}>
              <h4 className="text-md font-medium capitalize mb-2">{getReadableType(itemType)}</h4>
              <div className="space-y-2">
                {items.map((ri, idx) => (
                  <div key={idx} className="flex items-center justify-between border rounded p-2">
                    <div>
                      <div className="font-mono">{ri.id}</div>
                      <div>{ri.name}</div>
                    </div>
                    <div>
                      <Button variant="outline" size="sm" onClick={() => handleViewItem(ri)}>View</Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export function AttackDetailDialog({ open, onClose, item, type }: AttackDetailDialogProps) {
  // Local state for dialog navigation (kept minimal for dialog usage)
  const [currentItem, setCurrentItem] = useState<any>(item);

  useEffect(() => {
    if (open) setCurrentItem(item);
  }, [open, item]);

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="w-[95vw] max-w-none max-h-[90vh] p-6 overflow-hidden">
        <DialogHeader>
          <DialogTitle />
        </DialogHeader>
        <AttackDetailContent item={currentItem} type={type} />
      </DialogContent>
    </Dialog>
  );
}