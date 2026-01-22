'use client';

import React from 'react';
import { useCAPECList } from '@/lib/hooks';

export function CAPECTable() {
  const [page, setPage] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(50);
  const offset = page * pageSize;

  const { data, isLoading } = useCAPECList(offset, pageSize);
  const capecs = data?.capecs || [];
  const total: number = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="h-full flex flex-col">
      <div className="flex-1 min-h-0 overflow-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="border-b">
              <th className="text-left p-2">ID</th>
              <th className="text-left p-2">Name</th>
              <th className="text-left p-2">Summary</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={3} className="p-4 text-sm text-muted-foreground">Loading...</td>
              </tr>
            ) : capecs.length === 0 ? (
              <tr>
                <td colSpan={3} className="p-4 text-sm text-muted-foreground">No CAPEC entries</td>
              </tr>
            ) : (
              capecs.map((c: any) => (
                <tr key={c.id || c.CAPECID} className="hover:bg-muted">
                  <td className="p-2 font-mono">{c.id}</td>
                  <td className="p-2">{c.name}</td>
                  <td className="p-2 max-w-xs truncate" title={c.summary || c.description}>{c.summary || c.description}</td>
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
          <label className="text-sm">Page size:</label>
          <select
            value={pageSize}
            onChange={(e) => { setPageSize(Number(e.target.value)); setPage(0); }}
            className="border rounded px-2 py-1 text-sm"
          >
            {[10,25,50,100].map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>

          <button
            className="px-2 py-1 border rounded text-sm"
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page <= 0}
          >
            Prev
          </button>
          <div className="text-sm">{page + 1} / {pageCount}</div>
          <button
            className="px-2 py-1 border rounded text-sm"
            onClick={() => setPage((p) => Math.min(pageCount - 1, p + 1))}
            disabled={page >= pageCount - 1}
          >
            Next
          </button>
        </div>
      </div>
    </div>
  );
}
