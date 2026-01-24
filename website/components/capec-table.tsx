'use client';

import React from 'react';
import { useCAPECList } from '@/lib/hooks';
import { Button } from './ui/button';
import CAPECDetailDialog from './capec-detail-dialog';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

export function CAPECTable() {
  const [page, setPage] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(20);
  const offset = page * pageSize;

  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [selectedCapecId, setSelectedCapecId] = React.useState<string | null>(null);
  const [selectedRow, setSelectedRow] = React.useState<any | null>(null);

  const { data, isLoading } = useCAPECList(offset, pageSize);
  const capecs = data?.capecs || [];
  const total: number = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const visiblePages = 7;
  const getPageRange = () => {
    const half = Math.floor(visiblePages / 2);
    let start = Math.max(0, page - half);
    let end = Math.min(pageCount - 1, start + visiblePages - 1);
    if (end - start + 1 < visiblePages) start = Math.max(0, end - visiblePages + 1);
    return { start, end };
  };
  const { start, end } = getPageRange();
  const pages: number[] = [];
  for (let p = start; p <= end; p++) pages.push(p);

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle>CAPEC Database</CardTitle>
        <CardDescription>Browse and manage CAPEC records in the local database</CardDescription>
        <div className="mt-3">
          <Input
            className="w-full"
            placeholder="Search CAPEC ID or name"
            // Add search functionality here if available
          />
        </div>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 flex flex-col">
        <div className="flex-1 min-h-0 overflow-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left p-2">ID</th>
                <th className="text-left p-2">Name</th>
                <th className="text-left p-2">Summary</th>
                <th className="text-left p-2">Actions</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={4} className="p-4 text-sm text-muted-foreground">Loading...</td>
                </tr>
              ) : capecs.length === 0 ? (
                <tr>
                  <td colSpan={4} className="p-4 text-sm text-muted-foreground">No CAPEC entries</td>
                </tr>
              ) : (
                capecs.map((c: any) => (
                  <tr key={c.id || c.CAPECID} className="hover:bg-muted">
                    <td className="p-2 font-mono">{c.id}</td>
                    <td className="p-2">{c.name}</td>
                    <td className="p-2 max-w-xs truncate" title={c.summary || c.description}>{c.summary || c.description}</td>
                    <td className="p-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          setSelectedRow(c);
                          setSelectedCapecId(c.id || c.CAPECID);
                          setDialogOpen(true);
                        }}
                      >
                        View Detail
                      </Button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        
        <div className="pt-2 flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            {isLoading ? (
              'Loading...'
            ) : (
              `Showing ${Math.min(offset + 1, total || 0)}-${Math.min(offset + capecs.length, total || offset + capecs.length)} of ${total}`
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button size="sm" variant="outline" onClick={() => setPage(0)} disabled={page <= 0}>First</Button>
            <Button size="sm" variant="outline" onClick={() => setPage((p) => Math.max(0, p - 1))} disabled={page <= 0}>Prev</Button>

            <div className="hidden sm:flex items-center gap-1">
              {pages.map((p) => (
                <Button
                  key={p}
                  size="sm"
                  variant={p === page ? 'secondary' : 'outline'}
                  onClick={() => setPage(p)}
                >
                  {p + 1}
                </Button>
              ))}
            </div>

            <Button size="sm" variant="outline" onClick={() => setPage((p) => Math.min(pageCount - 1, p + 1))} disabled={page >= pageCount - 1}>Next</Button>
            <Button size="sm" variant="outline" onClick={() => setPage(pageCount - 1)} disabled={page >= pageCount - 1}>Last</Button>

            <div className="text-sm">{page + 1} / {pageCount}</div>

            <label className="text-sm">Page size:</label>
            <select
              value={pageSize}
              onChange={(e) => { setPageSize(Number(e.target.value)); setPage(0); }}
              className="border rounded px-2 py-1 text-sm"
            >
              {[10,20,25,50,100].map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          </div>
        </div>
      </CardContent>
      <CAPECDetailDialog
        open={dialogOpen}
        onOpenChange={(v) => {
          if (!v) {
            setDialogOpen(false);
            setSelectedCapecId(null);
            setSelectedRow(null);
          } else {
            setDialogOpen(true);
          }
        }}
        capecId={selectedCapecId}
        initial={selectedRow}
      />
    </Card>
  );
}