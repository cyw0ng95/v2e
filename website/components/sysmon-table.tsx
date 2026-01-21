"use client";

import { useSysMetrics } from "@/lib/hooks";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

export function SysMonitor() {
  const { data: metrics, isLoading } = useSysMetrics();

  if (isLoading) {
    return <Skeleton className="h-32 w-full" />;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Metric</TableHead>
          <TableHead>Value</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>CPU Usage</TableCell>
          <TableCell>{metrics?.cpuUsage?.toFixed(2)}%</TableCell>
        </TableRow>
        <TableRow>
          <TableCell>Memory Usage</TableCell>
          <TableCell>{metrics?.memoryUsage?.toFixed(2)}%</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  );
}