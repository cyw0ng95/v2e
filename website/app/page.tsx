'use client';

import { useCVEList, useCVECount, useSessionStatus } from "@/lib/hooks";
import { CVETable } from "@/components/cve-table";
import { SessionControl } from "@/components/session-control";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Database, Activity, AlertCircle } from "lucide-react";
import { useState } from "react";

export default function Home() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  const offset = page * pageSize;

  const { data: cveList, isLoading: isLoadingList } = useCVEList(offset, pageSize);
  const { data: cveCount } = useCVECount();
  const { data: sessionStatus } = useSessionStatus();

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <h1 className="text-4xl font-bold tracking-tight">v2e Dashboard</h1>
          <p className="text-muted-foreground">
            CVE (Common Vulnerabilities and Exposures) Management System
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total CVEs</CardTitle>
              <Database className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {cveCount?.count?.toLocaleString() || '0'}
              </div>
              <p className="text-xs text-muted-foreground">
                In local database
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Session Status</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {sessionStatus?.hasSession ? (
                  <Badge variant={sessionStatus.state === 'running' ? 'default' : 'secondary'}>
                    {sessionStatus.state}
                  </Badge>
                ) : (
                  <Badge variant="outline">Idle</Badge>
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {sessionStatus?.hasSession ? sessionStatus.sessionId : 'No active session'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Session Progress</CardTitle>
              <AlertCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {sessionStatus?.hasSession ? sessionStatus.fetchedCount || 0 : 0}
              </div>
              <p className="text-xs text-muted-foreground">
                CVEs fetched (
                {sessionStatus?.hasSession ? sessionStatus.storedCount || 0 : 0} stored,{' '}
                {sessionStatus?.hasSession ? sessionStatus.errorCount || 0 : 0} errors)
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Session Control */}
        <SessionControl />

        {/* CVE Table */}
        <Card>
          <CardHeader>
            <CardTitle>CVE Database</CardTitle>
            <CardDescription>
              Browse and manage CVE records in the local database
            </CardDescription>
          </CardHeader>
          <CardContent>
            <CVETable
              cves={cveList?.cves || []}
              total={cveList?.total || 0}
              page={page}
              pageSize={pageSize}
              isLoading={isLoadingList}
              onPageChange={setPage}
              onPageSizeChange={setPageSize}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
