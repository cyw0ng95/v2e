'use client';

import { useCVEList } from "@/lib/hooks";
import { useCVECount } from "@/lib/hooks";
import { useSessionStatus } from "@/lib/hooks";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { DatabaseIcon as Database, ActivityIcon as Activity, AlertCircleIcon as AlertCircle } from '@/components/icons';
import { useState, Suspense, useMemo, memo, Fragment } from "react";
import dynamic from 'next/dynamic';
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import NotesFramework from '@/components/notes-framework';
import BookmarkTable from '@/components/bookmark-table';
import NotesDashboard from '@/components/notes-dashboard';
import MemoryCardStudy from '@/components/memory-card-study';
import { ViewLearnToggle } from '@/components/view-learn-toggle';

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

const CAPECTable = dynamic(() => import('@/components/capec-table').then(mod => mod.CAPECTable), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

const CWEViews = dynamic(() => import('@/components/cwe-views').then(mod => mod.CWEViews), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

const SysMonitor = dynamic(() => import('@/components/sysmon-table').then(mod => mod.SysMonitor), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

const AttackTable = dynamic(() => import('@/components/attack-table').then(mod => mod.AttackTable), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

const AttackViews = dynamic(() => import('@/components/attack-views').then(mod => mod.AttackViews), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

const SSGTable = dynamic(() => import('@/components/ssg-table').then(mod => mod.SSGTable), {
  ssr: false,
  loading: () => (
    <div className="p-4">
      <Skeleton className="h-32 w-full" />
    </div>
  ),
});

// Memoized Right Column Component for tab content
const RightColumn = memo(function RightColumn({ 
  viewMode, 
  tab, 
  setTab, 
  page, 
  pageSize, 
  searchQuery, 
  setPage, 
  setPageSize, 
  setSearchQuery, 
  cveList, 
  isLoadingList 
}: { 
  viewMode: 'view' | 'learn'; 
  tab: string;
  setTab: (tab: string) => void;
  page: number;
  pageSize: number;
  searchQuery: string;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setSearchQuery: (query: string) => void;
  cveList?: any;
  isLoadingList: boolean;
}) {
  // Dynamic class for tab positioning - more left in Learn mode
  const tabListClass = useMemo(() => 
    viewMode === 'learn' 
      ? "mb-4 justify-start"  // More left-aligned in Learn mode
      : "mb-4"               // Default center alignment in View mode
  , [viewMode]);

  return (
    <main className="w-full md:flex-1 h-screen flex flex-col px-10 py-8">
      <div className="flex-1 flex flex-col h-full">
        <div className="space-y-8 flex-1 flex flex-col h-full page-transition">
          <Tabs value={tab} onValueChange={setTab} className="w-full h-full flex flex-col">
            <TabsList className={tabListClass}>
              {viewMode === 'view' ? (
                // Operational View Tabs
                <Fragment key="view-tabs">
                  <TabsTrigger value="cwe">CWE Database</TabsTrigger>
                  <TabsTrigger value="capec">CAPEC</TabsTrigger>
                  <TabsTrigger value="attack">ATT&CK</TabsTrigger>
                  <TabsTrigger value="ssg">SSG</TabsTrigger>
                  <TabsTrigger value="cweviews">CWE Views</TabsTrigger>
                  <TabsTrigger value="cve">CVE Database</TabsTrigger>
                  <TabsTrigger value="bookmarks">Bookmarks</TabsTrigger>
                  <TabsTrigger value="sysmon">SysMonitor</TabsTrigger>
                </Fragment>
              ) : (
                // Learning View Tabs - positioned more to the left
                <Fragment key="learn-tabs">
                  <TabsTrigger value="notes-dashboard">Notes Dashboard</TabsTrigger>
                  <TabsTrigger value="study-cards">Study Cards</TabsTrigger>
                  {/* Also include all operational view tabs in learn mode */}
                  <TabsTrigger value="cwe">CWE Database</TabsTrigger>
                  <TabsTrigger value="capec">CAPEC</TabsTrigger>
                  <TabsTrigger value="attack">ATT&CK</TabsTrigger>
                  <TabsTrigger value="ssg">SSG</TabsTrigger>
                  <TabsTrigger value="cweviews">CWE Views</TabsTrigger>
                  <TabsTrigger value="cve">CVE Database</TabsTrigger>
                  <TabsTrigger value="bookmarks">Bookmarks</TabsTrigger>
                  <TabsTrigger value="sysmon">SysMonitor</TabsTrigger>
                </Fragment>
              )}
            </TabsList>

            <TabsContent value="cwe" className="h-full">
              <div className="h-full flex flex-col">
                <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                  <CWETable viewMode={viewMode} />
                </Suspense>
              </div>
            </TabsContent>

            <TabsContent value="capec" className="h-full">
              <div className="h-full flex flex-col">
                <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                  <CAPECTable />
                </Suspense>
              </div>
            </TabsContent>

            <TabsContent value="attack" className="h-full">
              <div className="h-full flex flex-col">
                <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                  <AttackViews />
                </Suspense>
              </div>
            </TabsContent>

            <TabsContent value="ssg" className="h-full">
              <div className="h-full flex flex-col">
                <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                  <SSGTable />
                </Suspense>
              </div>
            </TabsContent>

            <TabsContent value="cweviews" className="h-full">
              <div className="h-full flex flex-col">
                <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                  <CWEViews />
                </Suspense>
              </div>
            </TabsContent>

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
                <CardContent className="flex-1 min-h-0">
                  <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
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

            <TabsContent value="bookmarks" className="h-full">
              <Card className="h-full flex flex-col">
                <CardHeader>
                  <CardTitle>Bookmarks & Notes</CardTitle>
                  <CardDescription>Manage your bookmarks and personal notes</CardDescription>
                </CardHeader>
                <CardContent className="flex-1 min-h-0">
                  <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                    <BookmarkTable />
                  </Suspense>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="notes-dashboard" className="h-full">
              <Card className="h-full flex flex-col">
                <CardHeader>
                  <CardTitle>Notes Dashboard</CardTitle>
                  <CardDescription>Your learning progress and activity</CardDescription>
                </CardHeader>
                <CardContent className="flex-1 min-h-0">
                  <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                    <NotesDashboard />
                  </Suspense>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="study-cards" className="h-full">
              <Card className="h-full flex flex-col">
                <CardHeader>
                  <CardTitle>Study Memory Cards</CardTitle>
                  <CardDescription>Review and rate your memory cards</CardDescription>
                </CardHeader>
                <CardContent className="flex-1 min-h-0">
                  <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                    <MemoryCardStudy filterState="to_review" />
                  </Suspense>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="sysmon" className="h-full">
              <Card className="h-full flex flex-col">
                <CardHeader>
                  <CardTitle>System Monitor</CardTitle>
                  <CardDescription>View system performance metrics</CardDescription>
                </CardHeader>
                <CardContent className="flex-1 min-h-0">
                  <Suspense fallback={<div className="p-4"><Skeleton className="h-32 w-full" /></div>}>
                    <SysMonitor />
                  </Suspense>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </main>
  );
});

// Memoized Left Sidebar Component to prevent re-renders on view mode changes
const LeftSidebar = memo(function LeftSidebar({
  viewMode,
  setViewMode,
  setTab,
  cveCount,
  sessionStatus
}: {
  viewMode: 'view' | 'learn';
  setViewMode: (mode: 'view' | 'learn') => void;
  setTab: (tab: any) => void;
  cveCount?: number;
  sessionStatus?: any;
}) {
  return (
    <aside className="w-full md:w-80 shrink-0 h-full flex flex-col">
      <div className="sticky top-0 left-0 bottom-0 p-6 space-y-6 overflow-auto w-full">
        {/* View/Learn Toggle */}
        <ViewLearnToggle
          value={viewMode}
          onValueChange={(newMode) => {
            setViewMode(newMode);
            // Reset to appropriate default tab when switching modes
            if (newMode === 'view') {
              setTab('cwe');
            } else {
              setTab('notes-dashboard');
            }
          }}
        />

        {/* Stats stacked - enhanced styling */}
        <div className="space-y-4">
          <Card className="w-full hover:shadow-md transition-all duration-150">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <CardTitle className="text-sm font-medium">Total CVEs</CardTitle>
              <Database className="h-4 w-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tracking-tight">{cveCount?.toLocaleString() || '0'}</div>
              <p className="text-xs text-muted-foreground mt-1">Local Database</p>
            </CardContent>
          </Card>

          <Card className="w-full hover:shadow-md transition-all duration-150">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <CardTitle className="text-sm font-medium">Session</CardTitle>
              <Activity className="h-4 w-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-lg font-semibold">
                {sessionStatus?.hasSession ? (
                  <Badge variant={sessionStatus.state === 'running' ? 'default' : 'secondary'} className="badge-info">
                    {sessionStatus.state}
                  </Badge>
                ) : (
                  <Badge variant="outline">Idle</Badge>
                )}
              </div>
              <p className="text-xs text-muted-foreground mt-2">{sessionStatus?.hasSession ? sessionStatus.sessionId : 'No active session'}</p>
            </CardContent>
          </Card>

          <Card className="w-full hover:shadow-md transition-all duration-150">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
              <CardTitle className="text-sm font-medium">Progress</CardTitle>
              <AlertCircle className="h-4 w-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tracking-tight">{sessionStatus?.hasSession ? sessionStatus.fetchedCount || 0 : 0}</div>
              <p className="text-xs text-muted-foreground mt-1">{sessionStatus?.hasSession ? `${sessionStatus.storedCount || 0} stored, ${sessionStatus.errorCount || 0} errors` : 'No activity'}</p>
            </CardContent>
          </Card>
        </div>

        <div className="w-full">
          <Suspense fallback={<div className="p-4"><Skeleton className="h-20 w-full" /></div>}>
            <SessionControl />
          </Suspense>
        </div>
      </div>
    </aside>
  );
});

export default function Home() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  const [searchQuery, setSearchQuery] = useState('');
  const [tab, setTab] = useState<string>('cwe'); // Default to CWE
  const [viewMode, setViewMode] = useState<'view' | 'learn'>('view'); // Default to View mode
  
  // memoize offset to avoid unnecessary recalculation on unrelated state updates
  const offset = useMemo(() => page * pageSize, [page, pageSize]);

  const { data: cveList, isLoading: isLoadingList } = useCVEList(offset, pageSize);
  const { data: cveCount } = useCVECount();
  const { data: sessionStatus } = useSessionStatus();

  return (
    <div className="h-[calc(100vh-var(--app-header-height))] w-screen bg-background overflow-hidden">
      <div className="h-full flex flex-col md:flex-row page-transition">
        {/* Left Sidebar - Memoized to prevent re-renders */}
        <LeftSidebar
          viewMode={viewMode}
          setViewMode={setViewMode}
          setTab={setTab}
          cveCount={cveCount ?? undefined}
          sessionStatus={sessionStatus ?? undefined}
        />

        {/* Right Main Area - Memoized to optimize performance */}
        <div className="hidden md:block w-px h-full bg-border/50" />
        <RightColumn
          viewMode={viewMode}
          tab={tab}
          setTab={setTab}
          page={page}
          pageSize={pageSize}
          searchQuery={searchQuery}
          setPage={setPage}
          setPageSize={setPageSize}
          setSearchQuery={setSearchQuery}
          cveList={cveList}
          isLoadingList={isLoadingList}
        />
      </div>
    </div>
  );
}
