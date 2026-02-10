'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Plus, FolderOpen, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { rpcClient } from '@/lib/rpc-client';
import { GraphCard, GraphList, GraphFilters, type DateRange } from '@/components/glc/graph-browser';
import type { GLCGraph } from '@/lib/types';

const PRESETS = [
  { id: 'topo', name: 'Topo' },
  { id: 'd3fend', name: 'D3FEND' },
];

export default function MyGraphsPage() {
  const router = useRouter();
  const [graphs, setGraphs] = useState<GLCGraph[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [search, setSearch] = useState('');
  const [presetFilter, setPresetFilter] = useState('all');
  const [dateRange, setDateRange] = useState<DateRange | undefined>();
  const [pageSize, setPageSize] = useState(12);
  const [page, setPage] = useState(0);

  const hasFilters = search !== '' || presetFilter !== 'all' || !!dateRange?.from;

  const clearFilters = useCallback(() => {
    setSearch('');
    setPresetFilter('all');
    setDateRange(undefined);
    setPage(0);
  }, []);

  // Filter graphs on the client side since the API may not support all filters
  const filteredGraphs = useMemo(() => {
    let result = graphs;

    // Search filter
    if (search) {
      const searchLower = search.toLowerCase();
      result = result.filter(
        (g) =>
          g.name.toLowerCase().includes(searchLower) ||
          (g.description && g.description.toLowerCase().includes(searchLower)) ||
          (g.tags && g.tags.toLowerCase().includes(searchLower))
      );
    }

    // Preset filter
    if (presetFilter !== 'all') {
      result = result.filter((g) => g.preset_id === presetFilter);
    }

    // Date range filter
    if (dateRange?.from) {
      const fromTime = dateRange.from.getTime();
      const toTime = dateRange.to ? dateRange.to.getTime() : Date.now();
      result = result.filter((g) => {
        const updatedAt = new Date(g.updated_at).getTime();
        return updatedAt >= fromTime && updatedAt <= toTime;
      });
    }

    return result;
  }, [graphs, search, presetFilter, dateRange]);

  const totalPages = Math.ceil(filteredGraphs.length / pageSize);
  const paginatedGraphs = filteredGraphs.slice(page * pageSize, (page + 1) * pageSize);

  // Fetch graphs
  const fetchGraphs = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await rpcClient.listGLCGraphs({ limit: 1000 });
      if (response.retcode === 0 && response.payload) {
        setGraphs(response.payload.graphs);
        setTotal(response.payload.total);
      } else {
        setError(response.message || 'Failed to load graphs');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchGraphs();
  }, [fetchGraphs]);

  // Handlers
  const handleDelete = useCallback(
    async (graphId: string) => {
      try {
        const response = await rpcClient.deleteGLCGraph({ graph_id: graphId });
        if (response.retcode === 0 && response.payload?.success) {
          setGraphs((prev) => prev.filter((g) => g.graph_id !== graphId));
          toast.success('Graph deleted');
        } else {
          toast.error(response.message || 'Failed to delete graph');
        }
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete graph');
      }
    },
    []
  );

  const handleDuplicate = useCallback(
    async (graph: GLCGraph) => {
      try {
        const response = await rpcClient.createGLCGraph({
          name: `${graph.name} (Copy)`,
          description: graph.description,
          preset_id: graph.preset_id,
          tags: graph.tags,
        });
        if (response.retcode === 0 && response.payload?.graph) {
          setGraphs((prev) => [...prev, response.payload!.graph]);
          toast.success('Graph duplicated');
        } else {
          toast.error(response.message || 'Failed to duplicate graph');
        }
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to duplicate graph');
      }
    },
    []
  );

  const handleExport = useCallback((graph: GLCGraph) => {
    // Export graph as JSON file
    const exportData = {
      name: graph.name,
      description: graph.description,
      presetId: graph.preset_id,
      tags: graph.tags,
      nodes: graph.nodes,
      edges: graph.edges,
      viewport: graph.viewport,
      exportedAt: new Date().toISOString(),
    };
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${graph.name.replace(/[^a-z0-9]/gi, '_')}.json`;
    a.click();
    URL.revokeObjectURL(url);
    toast.success('Graph exported');
  }, []);

  const handleShare = useCallback(
    (graph: GLCGraph) => {
      // Navigate to graph editor with share dialog open
      router.push(`/glc/${graph.preset_id}?graphId=${graph.graph_id}&share=1`);
    },
    [router]
  );

  // Reset page when filters change
  useEffect(() => {
    setPage(0);
  }, [search, presetFilter, dateRange, pageSize]);

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-surface/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-text">My Graphs</h1>
              <p className="text-sm text-textMuted mt-1">
                {total} graph{total !== 1 ? 's' : ''} total
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" asChild>
                <Link href="/glc">Back to GLC</Link>
              </Button>
              <Button asChild>
                <Link href="/glc/topo">
                  <Plus className="w-4 h-4 mr-2" />
                  New Graph
                </Link>
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Filters */}
        <div className="mb-6">
          <GraphFilters
            viewMode={viewMode}
            onViewModeChange={setViewMode}
            search={search}
            onSearchChange={setSearch}
            presetFilter={presetFilter}
            onPresetFilterChange={setPresetFilter}
            dateRange={dateRange}
            onDateRangeChange={setDateRange}
            pageSize={pageSize}
            onPageSizeChange={setPageSize}
            presets={PRESETS}
            hasFilters={hasFilters}
            onClearFilters={clearFilters}
          />
        </div>

        {/* Error State */}
        {error && (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <AlertCircle className="w-12 h-12 text-error mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-text mb-2">Failed to load graphs</h3>
              <p className="text-textMuted mb-4">{error}</p>
              <Button onClick={fetchGraphs}>Try Again</Button>
            </div>
          </div>
        )}

        {/* Loading State */}
        {isLoading && !error && (
          <div
            className={viewMode === 'grid' ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4' : 'space-y-2'}
          >
            {Array.from({ length: pageSize }).map((_, i) =>
              viewMode === 'grid' ? (
                <div key={i} className="rounded-lg border overflow-hidden">
                  <Skeleton className="aspect-video" />
                  <div className="p-4 space-y-2">
                    <Skeleton className="h-5 w-3/4" />
                    <Skeleton className="h-4 w-1/2" />
                  </div>
                </div>
              ) : (
                <Skeleton key={i} className="h-20 w-full" />
              )
            )}
          </div>
        )}

        {/* Empty State */}
        {!isLoading && !error && filteredGraphs.length === 0 && (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <FolderOpen className="w-12 h-12 text-textMuted mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-text mb-2">
                {hasFilters ? 'No graphs match your filters' : 'No graphs yet'}
              </h3>
              <p className="text-textMuted mb-4">
                {hasFilters
                  ? 'Try adjusting your search or filters'
                  : 'Create your first graph to get started'}
              </p>
              {hasFilters ? (
                <Button variant="outline" onClick={clearFilters}>
                  Clear Filters
                </Button>
              ) : (
                <Button asChild>
                  <Link href="/glc/topo">
                    <Plus className="w-4 h-4 mr-2" />
                    Create Graph
                  </Link>
                </Button>
              )}
            </div>
          </div>
        )}

        {/* Graph Grid/List */}
        {!isLoading && !error && filteredGraphs.length > 0 && (
          <>
            {viewMode === 'grid' ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {paginatedGraphs.map((graph) => (
                  <GraphCard
                    key={graph.graph_id}
                    graph={graph}
                    onDelete={handleDelete}
                    onDuplicate={handleDuplicate}
                    onExport={handleExport}
                    onShare={handleShare}
                  />
                ))}
              </div>
            ) : (
              <div className="rounded-lg border">
                <GraphList
                  graphs={paginatedGraphs}
                  onDelete={handleDelete}
                  onDuplicate={handleDuplicate}
                  onExport={handleExport}
                  onShare={handleShare}
                />
              </div>
            )}

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-2 mt-6">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === 0}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Previous
                </Button>
                <div className="flex items-center gap-1">
                  {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                    const pageNum = Math.max(0, Math.min(page - 2 + i, totalPages - 5));
                    const actualPage = totalPages <= 5 ? i : pageNum + i;
                    if (actualPage < 0 || actualPage >= totalPages) return null;
                    return (
                      <Button
                        key={actualPage}
                        variant={actualPage === page ? 'default' : 'outline'}
                        size="sm"
                        className="w-9"
                        onClick={() => setPage(actualPage)}
                      >
                        {actualPage + 1}
                      </Button>
                    );
                  })}
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages - 1}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next
                </Button>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}
