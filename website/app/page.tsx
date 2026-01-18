'use client';

import { useCVEList, useCVECount, useSessionStatus } from "@/lib/hooks";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { DatabaseIcon as Database, ActivityIcon as Activity, AlertCircleIcon as AlertCircle } from '@/components/icons';
import { useState, Suspense, useMemo } from "react";
import dynamic from 'next/dynamic';
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

// Lazy-load heavier client components to reduce initial bundle size
const CVETable = dynamic(() => import('@/components/cve-table').then(mod => mod.CVETable), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

const SessionControl = dynamic(() => import('@/components/session-control').then(mod => mod.SessionControl), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-20 w-full" />
    </div>
  ),
});

const CWETable = dynamic(() => import('@/components/cwe-table').then(mod => mod.CWETable), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

export default function Home() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  const [searchQuery, setSearchQuery] = useState('');
  const [tab, setTab] = useState<'cwe' | 'cve'>('cwe'); // Default to CWE
  // memoize offset to avoid unnecessary recalculation on unrelated state updates
  const offset = useMemo(() => page * pageSize, [page, pageSize]);

  const { data: cveList, isLoading: isLoadingList } = useCVEList(offset, pageSize);
  const { data: cveCount } = useCVECount();
  const { data: sessionStatus } = useSessionStatus();

  return (
    <div className="h-screen w-screen bg-background overflow-hidden">
      <div className="h-full p-6">
        <div className="h-full flex flex-col md:flex-row gap-6">
          {/* Left Sidebar */}
          <aside className="w-full md:w-72 shrink-0">
            <div className="space-y-4 sticky top-6">
              <div className="space-y-2">
                <h1 className="text-2xl md:text-3xl font-bold tracking-tight">v2e</h1>
                <p className="text-sm text-muted-foreground">CVE Management</p>
              </div>

              {/* Stats stacked */}
              <div className="space-y-4">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total CVEs</CardTitle>
                    <Database className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-xl font-bold">
                      {cveCount?.count?.toLocaleString() || '0'}
                    </div>
                    <p className="text-xs text-muted-foreground">Local DB</p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Session</CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-lg font-medium">
                      {sessionStatus?.hasSession ? (
                        <Badge variant={sessionStatus.state === 'running' ? 'default' : 'secondary'}>
                          {sessionStatus.state}
                        </Badge>
                      ) : (
                        <Badge variant="outline">Idle</Badge>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">{sessionStatus?.hasSession ? sessionStatus.sessionId : 'No active session'}</p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Progress</CardTitle>
                    <AlertCircle className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-lg font-medium">{sessionStatus?.hasSession ? sessionStatus.fetchedCount || 0 : 0}</div>
                    <p className="text-xs text-muted-foreground mt-1">{sessionStatus?.hasSession ? `${sessionStatus.storedCount || 0} stored, ${sessionStatus.errorCount || 0} errors` : 'No activity'}</p>
                  </CardContent>
                </Card>
              </div>
              {/* Session control in left sidebar (lazy-loaded) */}
              <div className="mt-4 w-full max-w-70">
                <Suspense>
                  <SessionControl />
                </Suspense>
              </div>
            </div>
          </aside>

          {/* Right Main Area */}
          <main className="flex-1 overflow-auto h-full">
            <div className="space-y-6 h-full">
              <Tabs value={tab} onValueChange={setTab} className="w-full h-full flex flex-col">
                <TabsList className="mb-4">
                  <TabsTrigger value="cwe">CWE Database</TabsTrigger>
                  <TabsTrigger value="cve">CVE Database</TabsTrigger>
                </TabsList>
                <TabsContent value="cwe" className="h-full"><CWETable /></TabsContent>
                <TabsContent value="cve" className="h-full">
                  <Card className="h-full flex flex-col">
                    <CardHeader>
                      <CardTitle>CVE Database</CardTitle>
                      <CardDescription>Browse and manage CVE records in the local database</CardDescription>
                      <div className="mt-3">
                        <Input
                          placeholder="Search CVE ID or description (local filter)"
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground mt-1">Note: server-side search not implemented yet; this filters currently loaded results.</p>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <Suspense>
                        <CVETable
                          cves={cveList?.cves || []}
                          total={cveList?.total || 0}
                          page={page}
                          pageSize={pageSize}
                          isLoading={isLoadingList}
                          onPageChange={setPage}
                          onPageSizeChange={setPageSize}
                          searchQuery={searchQuery}
                        />
                      </Suspense>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
