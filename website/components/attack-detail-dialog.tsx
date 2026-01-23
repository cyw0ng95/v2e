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

export function AttackDetailDialog({ open, onClose, item, type }: AttackDetailDialogProps) {
  // Track the current item being displayed (could be the initial item or a related item)
  const [currentItem, setCurrentItem] = useState<any>(item);
  const [currentType, setCurrentType] = useState<'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups'>(type);
  // Track navigation history for back button functionality
  const [navigationHistory, setNavigationHistory] = useState<Array<{item: any, type: string}>>([]);
  
  // For demonstration, we'll use a simple approach to define relationships
  // In a real implementation, these would be stored in the database
  const getRelatedItemsByType = (itemType: string) => {
    if (!currentItem) return [];
    
    // This is a simplified example - in a real implementation, 
    // relationships would be stored in the database
    switch(currentType) {
      case 'techniques':
        switch(itemType) {
          case 'tactics':
            // Return the tactic this technique belongs to (if known)
            return currentItem.tactic ? [{ id: currentItem.tactic, name: currentItem.tacticName || currentItem.tactic, type: 'tactics' }] : [];
          case 'mitigations':
            // Return some related mitigations (simplified)
            return [
              { id: 'M1036', name: 'Credential Access Prevention', type: 'mitigations' },
              { id: 'M1053', name: 'Run Command', type: 'mitigations' },
            ].slice(0, 2);
          case 'software':
            // Return related software (simplified)
            return [
              { id: 'S0001', name: 'Compiled HTML File', type: 'software' },
              { id: 'S0002', name: 'RegSvr32', type: 'software' },
            ].slice(0, 2);
          case 'groups':
            // Return related groups (simplified)
            return [
              { id: 'G0001', name: 'APT1', type: 'groups' },
              { id: 'G0006', name: 'APT28', type: 'groups' },
            ].slice(0, 2);
          case 'techniques':
            // Return related techniques (simplified)
            return [
              { id: 'T1001', name: 'Data Obfuscation', type: 'techniques' },
              { id: 'T1071', name: 'Application Layer Protocol', type: 'techniques' },
            ].slice(0, 2);
          default:
            return [];
        }
      case 'tactics':
        // For tactics, show related techniques
        if (itemType === 'techniques') {
          return [
            { id: 'T1003', name: 'OS Credential Dumping', type: 'techniques' },
            { id: 'T1005', name: 'Data from Local System', type: 'techniques' },
          ];
        }
        return [];
      case 'mitigations':
        // For mitigations, show related techniques
        if (itemType === 'techniques') {
          return [
            { id: 'T1003', name: 'OS Credential Dumping', type: 'techniques' },
            { id: 'T1053', name: 'Create or Modify System Process', type: 'techniques' },
          ];
        }
        return [];
      case 'software':
        // For software, show related techniques and groups
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
        // For groups, show related software and techniques
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

  const getReadableType = (type: string) => {
    switch(type) {
      case 'techniques': return 'Technique';
      case 'tactics': return 'Tactic';
      case 'mitigations': return 'Mitigation';
      case 'software': return 'Software';
      case 'groups': return 'Group';
      default: return type;
    }
  };

  const getTypeColor = (type: string) => {
    switch(type) {
      case 'techniques': return 'default';
      case 'tactics': return 'secondary';
      case 'mitigations': return 'outline';
      case 'software': return 'destructive';
      case 'groups': return 'outline';
      default: return 'default';
    }
  };

  // Navigate back to the previous item in history
  const handleBack = () => {
    if (navigationHistory.length > 0) {
      const previousState = navigationHistory[navigationHistory.length - 1];
      const newHistory = [...navigationHistory.slice(0, -1)];
      
      setCurrentItem(previousState.item);
      setCurrentType(previousState.type as 'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups');
      setNavigationHistory(newHistory);
    }
  };

  // Handle viewing a related item - update current item and type
  const handleViewItem = (relatedItem: RelatedItem) => {
    // Save current state to history before navigating
    setNavigationHistory([...navigationHistory, { item: currentItem, type: currentType }]);
    
    // In a real implementation, we would fetch the actual item details here
    console.log(`Viewing ${relatedItem.type} ${relatedItem.id}`);
    
    // For now, we'll just update the current item and type to simulate navigation
    setCurrentItem({...relatedItem, name: relatedItem.name, id: relatedItem.id});
    setCurrentType(relatedItem.type as 'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups');
  };

  // Reset to the original item when dialog reopens
  useEffect(() => {
    if (open) {
      setCurrentItem(item);
      setCurrentType(type);
      setNavigationHistory([]);
    }
  }, [open, item, type]);

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="w-[90vw] h-[80vh] max-w-none p-6 overflow-y-auto">
        <DialogHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {navigationHistory.length > 0 && (
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={handleBack}
                  className="mr-2"
                >
                  ← Back
                </Button>
              )}
              <DialogTitle>
                <div className="flex items-center gap-2">
                  <Badge variant={getTypeColor(currentType)}>
                    {getReadableType(currentType)}
                  </Badge>
                  <span>{currentItem?.name || currentItem?.id || 'Unknown Item'}</span>
                  {currentItem?.id && (
                    <Badge variant="outline" className="font-mono">
                      {currentItem.id}
                    </Badge>
                  )}
                </div>
              </DialogTitle>
            </div>
          </div>
        </DialogHeader>
        
        <div className="space-y-4 h-[calc(100%-3rem)] overflow-y-auto pr-2">
          {/* Basic Information */}
          <div className="space-y-2">
            <h3 className="text-lg font-semibold">Information</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {currentItem?.id && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">ID</h4>
                  <p className="font-mono">{currentItem.id}</p>
                </div>
              )}
              {currentItem?.name && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Name</h4>
                  <p>{currentItem.name}</p>
                </div>
              )}
              {currentItem?.domain && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Domain</h4>
                  <p>{currentItem.domain}</p>
                </div>
              )}
              {currentItem?.platform && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Platform</h4>
                  <p>{currentItem.platform}</p>
                </div>
              )}
              {currentItem?.type && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Type</h4>
                  <p>{currentItem.type}</p>
                </div>
              )}
              {currentItem?.created && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Created</h4>
                  <p>{currentItem.created}</p>
                </div>
              )}
              {currentItem?.modified && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Modified</h4>
                  <p>{currentItem.modified}</p>
                </div>
              )}
            </div>
          </div>
          
          {/* Description */}
          {currentItem?.description && (
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Description</h3>
              <p className="whitespace-pre-line">{currentItem.description}</p>
            </div>
          )}
          
          {/* Related Items */}
          <div className="space-y-2 flex-grow overflow-y-auto">
            <h3 className="text-lg font-semibold">Related Items</h3>
            <div className="space-y-4">
              {['techniques', 'tactics', 'mitigations', 'software', 'groups'].map((itemType) => {
                const items = getRelatedItemsByType(itemType);
                if (items.length === 0) return null;
                
                return (
                  <div key={itemType}>
                    <h4 className="text-md font-medium capitalize mb-2">{getReadableType(itemType)}</h4>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[100px]">ID</TableHead>
                          <TableHead>Name</TableHead>
                          <TableHead>Action</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {items.map((relatedItem, idx) => (
                          <TableRow key={idx}>
                            <TableCell className="font-mono">
                              <Badge variant="outline">{relatedItem.id}</Badge>
                            </TableCell>
                            <TableCell>{relatedItem.name}</TableCell>
                            <TableCell>
                              <Button 
                                variant="outline" 
                                size="sm"
                                onClick={() => handleViewItem(relatedItem)}
                              >
                                View
                              </Button>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}