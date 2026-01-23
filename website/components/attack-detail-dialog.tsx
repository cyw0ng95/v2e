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
  useAttackGroups 
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
  const [relatedItems, setRelatedItems] = useState<RelatedItem[]>([]);
  
  // For demonstration purposes, we'll create mock related items
  // In a real implementation, these would be fetched from the backend
  useEffect(() => {
    if (item && open) {
      // Generate mock related items based on the current item
      const mockRelatedItems: RelatedItem[] = [];
      
      // Add some related items based on the current type
      switch(type) {
        case 'techniques':
          // Add some related tactics, mitigations, etc.
          mockRelatedItems.push(
            { id: 'TA0001', name: 'Initial Access', type: 'tactics' },
            { id: 'M1036', name: 'Credential Access Prevention', type: 'mitigations' },
            { id: 'S0001', name: 'Some Malware', type: 'software' }
          );
          break;
        case 'tactics':
          // Add some related techniques
          mockRelatedItems.push(
            { id: 'T1001', name: 'Some Technique', type: 'techniques' },
            { id: 'T1002', name: 'Another Technique', type: 'techniques' }
          );
          break;
        case 'mitigations':
          // Add some related techniques
          mockRelatedItems.push(
            { id: 'T1003', name: 'Some Technique', type: 'techniques' },
            { id: 'T1004', name: 'Another Technique', type: 'techniques' }
          );
          break;
        case 'software':
          // Add some related techniques
          mockRelatedItems.push(
            { id: 'T1005', name: 'Some Technique', type: 'techniques' },
            { id: 'G0001', name: 'Some Group', type: 'groups' }
          );
          break;
        case 'groups':
          // Add some related software and techniques
          mockRelatedItems.push(
            { id: 'S0002', name: 'Some Software', type: 'software' },
            { id: 'T1006', name: 'Some Technique', type: 'techniques' }
          );
          break;
      }
      
      setRelatedItems(mockRelatedItems);
    }
  }, [item, open, type]);

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

  const getRelatedItemsByType = (itemType: string) => {
    return relatedItems.filter(ri => ri.type === itemType);
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            <div className="flex items-center gap-2">
              <Badge variant={getTypeColor(type)}>
                {getReadableType(type)}
              </Badge>
              <span>{item?.name || item?.id || 'Unknown Item'}</span>
              {item?.id && (
                <Badge variant="outline" className="font-mono">
                  {item.id}
                </Badge>
              )}
            </div>
          </DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4">
          {/* Basic Information */}
          <div className="space-y-2">
            <h3 className="text-lg font-semibold">Information</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {item?.id && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">ID</h4>
                  <p className="font-mono">{item.id}</p>
                </div>
              )}
              {item?.name && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Name</h4>
                  <p>{item.name}</p>
                </div>
              )}
              {item?.domain && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Domain</h4>
                  <p>{item.domain}</p>
                </div>
              )}
              {item?.platform && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Platform</h4>
                  <p>{item.platform}</p>
                </div>
              )}
              {item?.type && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Type</h4>
                  <p>{item.type}</p>
                </div>
              )}
              {item?.created && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Created</h4>
                  <p>{item.created}</p>
                </div>
              )}
              {item?.modified && (
                <div>
                  <h4 className="text-sm font-medium text-muted-foreground">Modified</h4>
                  <p>{item.modified}</p>
                </div>
              )}
            </div>
          </div>
          
          {/* Description */}
          {item?.description && (
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Description</h3>
              <p className="whitespace-pre-line">{item.description}</p>
            </div>
          )}
          
          {/* Related Items */}
          <div className="space-y-2">
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
                              <Button variant="outline" size="sm">
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