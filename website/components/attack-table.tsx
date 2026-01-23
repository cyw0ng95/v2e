'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { 
  useAttackTechniques, 
  useAttackTactics, 
  useAttackMitigations, 
  useAttackSoftware, 
  useAttackGroups 
} from '@/lib/hooks';
import { AttackTechnique } from '@/lib/types';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';

interface AttackTableProps {
  type?: 'techniques' | 'tactics' | 'mitigations' | 'software' | 'groups'; // Default to techniques
}

export function AttackTable({ type = 'techniques' }: AttackTableProps) {
  const [page, setPage] = useState(0);
  const [search, setSearch] = useState("");
  // Fixed page size: show 20 items per page
  const [pageSize, setPageSize] = useState(20);

  // Use different hooks based on type
  const { data: techniquesData, isLoading: techniquesLoading } = useAttackTechniques({ offset: page * pageSize, limit: pageSize, search });
  const { data: tacticsData, isLoading: tacticsLoading } = useAttackTactics({ offset: page * pageSize, limit: pageSize, search });
  const { data: mitigationsData, isLoading: mitigationsLoading } = useAttackMitigations({ offset: page * pageSize, limit: pageSize, search });
  const { data: softwareData, isLoading: softwareLoading } = useAttackSoftware({ offset: page * pageSize, limit: pageSize, search });
  const { data: groupsData, isLoading: groupsLoading } = useAttackGroups({ offset: page * pageSize, limit: pageSize, search });

  const isLoading = 
    (type === 'techniques' && techniquesLoading) ||
    (type === 'tactics' && tacticsLoading) ||
    (type === 'mitigations' && mitigationsLoading) ||
    (type === 'software' && softwareLoading) ||
    (type === 'groups' && groupsLoading);

  // Map backend ATT&CK items to a plain object for table display based on type
  let attackList = [];
  let total = 0;

  switch(type) {
    case 'tactics':
      attackList = Array.isArray(tacticsData?.tactics)
        ? tacticsData.tactics.map((item: any) => ({
            id: item.id || item.ID || item.attackId || item.AttackId || '',
            name: item.name || item.Name || '',
            description: item.description || item.Description || '',
            domain: item.domain || item.Domain || '',
          }))
        : [];
      total = tacticsData?.total || 0;
      break;
    case 'mitigations':
      attackList = Array.isArray(mitigationsData?.mitigations)
        ? mitigationsData.mitigations.map((item: any) => ({
            id: item.id || item.ID || item.attackId || item.AttackId || '',
            name: item.name || item.Name || '',
            description: item.description || item.Description || '',
            domain: item.domain || item.Domain || '',
          }))
        : [];
      total = mitigationsData?.total || 0;
      break;
    case 'software':
      attackList = Array.isArray(softwareData?.software)
        ? softwareData.software.map((item: any) => ({
            id: item.id || item.ID || item.attackId || item.AttackId || '',
            name: item.name || item.Name || '',
            description: item.description || item.Description || '',
            type: item.type || item.Type || '',
            domain: item.domain || item.Domain || '',
          }))
        : [];
      total = softwareData?.total || 0;
      break;
    case 'groups':
      attackList = Array.isArray(groupsData?.groups)
        ? groupsData.groups.map((item: any) => ({
            id: item.id || item.ID || item.attackId || item.AttackId || '',
            name: item.name || item.Name || '',
            description: item.description || item.Description || '',
            domain: item.domain || item.Domain || '',
          }))
        : [];
      total = groupsData?.total || 0;
      break;
    case 'techniques':
    default:
      attackList = Array.isArray(techniquesData?.techniques)
        ? techniquesData.techniques.map((item: any) => ({
            id: item.id || item.ID || item.attackId || item.AttackId || '',
            name: item.name || item.Name || '',
            description: item.description || item.Description || '',
            domain: item.domain || item.Domain || '',
            platform: item.platform || item.Platform || '',
          }))
        : [];
      total = techniquesData?.total || 0;
      break;
  }

  // Calculate total pages
  const totalPages = Math.ceil(total / pageSize);

  // Determine table headers based on type
  const renderTableHeaders = () => {
    switch(type) {
      case 'software':
        return (
          <TableRow>
            <TableHead className="w-[100px]">ID</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Domain</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Action</TableHead>
          </TableRow>
        );
      case 'groups':
        return (
          <TableRow>
            <TableHead className="w-[100px]">ID</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Domain</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Action</TableHead>
          </TableRow>
        );
      default:
        return (
          <TableRow>
            <TableHead className="w-[100px]">ID</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Domain</TableHead>
            <TableHead>Platform</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Action</TableHead>
          </TableRow>
        );
    }
  };

  // Render table cells based on type
  const renderTableCells = (attack: any) => {
    switch(type) {
      case 'software':
        return (
          <>
            <TableCell className="font-mono">
              <Badge variant="outline">{attack.id}</Badge>
            </TableCell>
            <TableCell className="font-medium">{attack.name}</TableCell>
            <TableCell>{attack.type}</TableCell>
            <TableCell>{attack.domain}</TableCell>
            <TableCell className="max-w-xs truncate">{attack.description}</TableCell>
            <TableCell>
              <Button variant="outline" size="sm">
                View Detail
              </Button>
            </TableCell>
          </>
        );
      case 'groups':
        return (
          <>
            <TableCell className="font-mono">
              <Badge variant="outline">{attack.id}</Badge>
            </TableCell>
            <TableCell className="font-medium">{attack.name}</TableCell>
            <TableCell>{attack.domain}</TableCell>
            <TableCell className="max-w-xs truncate">{attack.description}</TableCell>
            <TableCell>
              <Button variant="outline" size="sm">
                View Detail
              </Button>
            </TableCell>
          </>
        );
      default:
        return (
          <>
            <TableCell className="font-mono">
              <Badge variant="outline">{attack.id}</Badge>
            </TableCell>
            <TableCell className="font-medium">{attack.name}</TableCell>
            <TableCell>{attack.domain}</TableCell>
            <TableCell>{attack.platform}</TableCell>
            <TableCell className="max-w-xs truncate">{attack.description}</TableCell>
            <TableCell>
              <Button variant="outline" size="sm">
                View Detail
              </Button>
            </TableCell>
          </>
        );
    }
  };

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle>{type.charAt(0).toUpperCase() + type.slice(1)} Database</CardTitle>
        <div className="mt-3">
          <Input
            className="w-full"
            placeholder={`Search ${type} ID or name`}
            value={search}
            onChange={e => {
              setSearch(e.target.value);
              setPage(0); // Reset to first page when searching
            }}
          />
        </div>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 overflow-auto">
        {isLoading ? (
          <Skeleton className="h-32 w-full" />
        ) : (
          <>
            <Table className="min-w-full">
              <TableHeader>
                {renderTableHeaders()}
              </TableHeader>
              <TableBody>
                {attackList.length > 0 ? (
                  attackList.map((attack: any, idx: number) => (
                    <TableRow key={attack.id || idx} className="hover:bg-muted/30">
                      {renderTableCells(attack)}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={type === 'software' ? 6 : type === 'groups' ? 5 : 6} className="text-center text-muted-foreground py-8">
                      No {type} found
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>

            {/* Pagination Controls */}
            <div className="flex items-center justify-between mt-4">
              <div className="text-sm text-muted-foreground">
                Showing {(page * pageSize) + 1}-{Math.min((page + 1) * pageSize, total)} of {total} {type}
              </div>
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(Math.max(0, page - 1))}
                  disabled={page === 0}
                >
                  Previous
                </Button>
                <div className="text-sm mx-2">
                  Page {page + 1} of {totalPages}
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
                  disabled={page >= totalPages - 1}
                >
                  Next
                </Button>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}