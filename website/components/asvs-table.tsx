'use client';

import React, { useState } from 'react';
import { useASVSList } from '@/lib/hooks';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useViewLearnMode } from '@/contexts/ViewLearnContext';
import { EntityDetailModal, type EntityType } from '@/components/entity-detail-modal';
import { generateURN } from '@/lib/utils';
import type { ASVSItem } from '@/lib/types';

export function ASVSTable() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [search, setSearch] = useState('');
  const [selectedASVS, setSelectedASVS] = useState<ASVSItem | null>(null);
  const { mode } = useViewLearnMode();
  const isLearnMode = mode === 'learn';

  const { data, isLoading } = useASVSList({
    offset: 0, // Fetch all for client-side filtering
    limit: 1000
  });

  // Client-side search filtering
  const allRequirements = data?.requirements || [];
  const filteredRequirements = React.useMemo(() => {
    if (!search.trim()) return allRequirements;
    const q = search.toLowerCase();
    return allRequirements.filter((req: ASVSItem) =>
      req.requirementID.toLowerCase().includes(q) ||
      req.description.toLowerCase().includes(q) ||
      req.chapter.toLowerCase().includes(q) ||
      req.section.toLowerCase().includes(q)
    );
  }, [allRequirements, search]);

  // Pagination after filtering
  const total = filteredRequirements.length;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const startIndex = page * pageSize;
  const endIndex = Math.min(startIndex + pageSize, total);
  const requirements = filteredRequirements.slice(startIndex, endIndex);

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle>ASVS Requirements</CardTitle>
        <CardDescription>
          Application Security Verification Standard
        </CardDescription>
        <div className="mt-3">
          <Input
            className="w-full"
            placeholder="Search requirement ID or description"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(0);
            }}
          />
        </div>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 overflow-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="border-b">
              <th className="text-left p-2">ID</th>
              <th className="text-left p-2">Chapter</th>
              <th className="text-left p-2">Section</th>
              <th className="text-left p-2">Description</th>
              <th className="text-left p-2">Levels</th>
              <th className="text-left p-2">Action</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={6} className="p-4 text-center text-muted-foreground">
                  Loading...
                </td>
              </tr>
            ) : requirements.length === 0 ? (
              <tr>
                <td colSpan={6} className="p-4 text-center text-muted-foreground">
                  No ASVS requirements found
                </td>
              </tr>
            ) : (
              requirements.map((req: ASVSItem) => (
                <tr
                  key={req.requirementID}
                  className={`border-b hover:bg-muted ${isLearnMode ? 'learn-enhanced' : ''}`}
                >
                  <td className="p-2 font-mono">{req.requirementID}</td>
                  <td className="p-2">{req.chapter}</td>
                  <td className="p-2">{req.section}</td>
                  <td className="p-2 max-w-md truncate" title={req.description}>
                    {req.description}
                  </td>
                  <td className="p-2">
                    <div className="flex gap-1">
                      {req.level1 && <Badge variant="default">L1</Badge>}
                      {req.level2 && <Badge variant="secondary">L2</Badge>}
                      {req.level3 && <Badge variant="outline">L3</Badge>}
                    </div>
                  </td>
                  <td className="p-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setSelectedASVS(req)}
                    >
                      View Detail
                    </Button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>

        {/* Pagination */}
        <div className="flex items-center justify-between mt-4">
          <div className="text-sm text-muted-foreground">
            Showing {(page * pageSize) + 1}-{Math.min((page + 1) * pageSize, total)} of {total} requirements
          </div>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
            >
              Previous
            </Button>
            <div className="text-sm mx-2">
              Page {page + 1} of {totalPages}
            </div>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
            >
              Next
            </Button>
          </div>
        </div>
      </CardContent>

      {/* Detail Modal */}
      <EntityDetailModal
        open={selectedASVS !== null}
        onOpenChange={(open) => !open && setSelectedASVS(null)}
        entityType="ASVS"
        entityId={selectedASVS?.requirementID || ''}
        title={selectedASVS ? `ASVS ${selectedASVS.requirementID}` : ''}
        description={selectedASVS?.description || ''}
        metadata={selectedASVS || undefined}
        urn={selectedASVS ? generateURN('ASVS', selectedASVS.requirementID) : undefined}
      >
        <ASVSDetailContent asvs={selectedASVS} />
      </EntityDetailModal>
    </Card>
  );
}

function ASVSDetailContent({ asvs }: { asvs: ASVSItem | null }) {
  if (!asvs) return null;
  return (
    <div className="space-y-4">
      <div>
        <h3 className="font-semibold mb-2">Requirement</h3>
        <p className="text-sm whitespace-pre-wrap">{asvs.description}</p>
      </div>
      <div>
        <h3 className="font-semibold mb-2">Verification Levels</h3>
        <div className="flex gap-2">
          {asvs.level1 && <Badge variant="default">Level 1</Badge>}
          {asvs.level2 && <Badge variant="secondary">Level 2</Badge>}
          {asvs.level3 && <Badge variant="outline">Level 3</Badge>}
        </div>
      </div>
      {asvs.cwe && (
        <div>
          <h3 className="font-semibold mb-2">Related CWE</h3>
          <Badge variant="outline">{asvs.cwe}</Badge>
        </div>
      )}
    </div>
  );
}
