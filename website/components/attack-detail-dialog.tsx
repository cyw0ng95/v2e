'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
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
import { useViewLearnMode } from '@/contexts/ViewLearnContext';
import BookmarkStar from '@/components/bookmark-star';
import { generateURN } from '@/lib/utils';

interface RelatedItem {
  id: string;
  name: string;
  type: string;
}

export function AttackDetailContent({ item, type, onBack }: { item: any; type: string; onBack?: () => void }) {
  const { mode } = useViewLearnMode();
  const isLearnMode = mode === 'learn';
  const [currentItem, setCurrentItem] = useState<any>(item || {});
  const [navigationHistory, setNavigationHistory] = useState<Array<any>>([]);

  const tech = useAttackTechnique(currentItem?.id || '');
  const tac = useAttackTactic(currentItem?.id || '');
  const mit = useAttackMitigation(currentItem?.id || '');
  const soft = useAttackSoftwareById(currentItem?.id || '');
  const grp = useAttackGroupById(currentItem?.id || '');

  useEffect(() => {
    setCurrentItem(item || {});
    setNavigationHistory([]);
  }, [item]);

  const detailed = tech.data || tac.data || mit.data || soft.data || grp.data || currentItem;

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
    const ct = (current.type || type || 'techniques');
    switch(ct) {
      case 'techniques':
        switch(itemType) {
          case 'tactics':
            return current.tactic ? [{ id: current.tactic, name: current.tacticName || current.tactic, type: 'tactics' }] : [];
          case 'mitigations':
            return [
              { id: 'M1036', name: 'Credential Access Prevention', type: 'mitigations' },
              { id: 'M1053', name: 'Run Command', type: 'mitigations' },
            ].slice(0, 2);
          case 'software':
            return [
              { id: 'S0001', name: 'Compiled HTML File', type: 'software' },
              { id: 'S0002', name: 'RegSvr32', type: 'software' },
            ].slice(0, 2);
          case 'groups':
            return [
              { id: 'G0001', name: 'APT1', type: 'groups' },
              { id: 'G0006', name: 'APT28', type: 'groups' },
            ].slice(0, 2);
          case 'techniques':
            return [
              { id: 'T1001', name: 'Data Obfuscation', type: 'techniques' },
              { id: 'T1071', name: 'Application Layer Protocol', type: 'techniques' },
            ].slice(0, 2);
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
        }
        if (itemType === 'groups') {
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
        }
        if (itemType === 'techniques') {
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
    <div className={`max-h-[80vh] overflow-y-auto pr-2 space-y-6 p-2 sm:p-4 ${isLearnMode ? 'learn-focus-area' : ''}`}>
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 flex-wrap mb-3">
            <Badge variant="outline">{detailed?.id || currentItem.id}</Badge>
            <h2 className={`text-lg font-semibold ${isLearnMode ? 'learn-enhanced' : ''}`}>{detailed?.name || currentItem.name || 'Detail'}</h2>
          </div>
          {detailed?.id && (
            <div className="text-xs text-muted-foreground font-mono bg-muted px-2 py-1 rounded inline-block">
              URN: {generateURN('ATTACK', detailed.id)}
            </div>
          )}
        </div>
        <BookmarkStar
          itemId={detailed?.id || currentItem.id || ''}
          itemType="ATTACK"
          itemTitle={detailed?.name || currentItem.name || 'ATT&CK'}
          itemDescription={detailed?.description || ''}
          viewMode={mode}
        />
      </div>
      <div className="mt-3">
        <h3 className="text-lg font-semibold">Information</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
          {detailed?.id && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground">ID</h4>
              <p className="font-mono">{detailed.id}</p>
            </div>
          )}
          {detailed?.name && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground">Name</h4>
              <p>{detailed.name}</p>
            </div>
          )}
          {detailed?.domain && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground">Domain</h4>
              <p>{detailed.domain}</p>
            </div>
          )}
          {detailed?.platform && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground">Platform</h4>
              <p>{Array.isArray(detailed.platform) ? detailed.platform.join(', ') : detailed.platform}</p>
            </div>
          )}
          {detailed?.type && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground">Type</h4>
              <p>{detailed.type}</p>
            </div>
          )}
          {detailed?.created && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground">Created</h4>
              <p>{detailed.created}</p>
            </div>
          )}
          {detailed?.modified && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground">Modified</h4>
              <p>{detailed.modified}</p>
            </div>
          )}
          {detailed?.objective && (
            <div className="md:col-span-2">
              <h4 className="text-sm font-medium text-muted-foreground">Objective</h4>
              <p className="text-sm text-muted-foreground">{detailed.objective}</p>
            </div>
          )}
        </div>
      </div>

      {detailed?.description && (
        <div>
          <h3 className="text-lg font-semibold">Description</h3>
          <p className={`whitespace-pre-line ${isLearnMode ? 'learn-text-lg' : ''}`}>{detailed.description}</p>
        </div>
      )}

      {detailed?.references && Array.isArray(detailed.references) && (
        <div>
          <h3 className="text-lg font-semibold">References</h3>
          <ul className="list-disc pl-5">
            {detailed.references.map((r: any, i: number) => (
              <li key={i} className="text-sm">{r}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Related Items</h3>
        {['techniques', 'tactics', 'mitigations', 'software', 'groups'].map((itemType) => {
          const items = getRelatedItemsByType(detailed || currentItem, itemType);
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
