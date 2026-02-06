"use client";

import React from "react";
import { useSysMetrics } from "@/lib/hooks";
import { usePollingMetrics } from "@/lib/usePollingMetrics";
import SysmonCharts from "@/components/sysmon-charts";
import MessageStatsCharts from "@/components/message-stats-charts";
import Accordion, { AccordionItem, AccordionTrigger, AccordionContent } from "./ui/accordion";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";


function formatNumber(n?: number, digits = 2) {
  return typeof n === "number" ? n.toFixed(digits) : "—";
}

function formatBytes(bytes?: number) {
  if (typeof bytes !== "number") return "—";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let i = 0;
  let v = bytes;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(2)} ${units[i]}`;
}

function formatUptime(sec?: number) {
  if (typeof sec !== "number") return "—";
  const days = Math.floor(sec / 86400);
  sec = sec % 86400;
  const hours = Math.floor(sec / 3600);
  sec = sec % 3600;
  const mins = Math.floor(sec / 60);
  return `${days}d ${hours}h ${mins}m`;
}

export function SysMonitor() {
  const { data: metrics, isLoading } = useSysMetrics();
  const polling = usePollingMetrics(3000, 120, true);
  const [intervalMs, setIntervalMs] = React.useState<number>(3000);
  const [selectedMetrics, setSelectedMetrics] = React.useState<string[]>(["cpuUsage", "memoryUsage", "loadAvg"]);
  // single shared open panel across SysMonitor and MessageStatsCharts:
  const [openPanel, setOpenPanel] = React.useState<"system" | "global" | "proc" | null>("system");

  const toggleMetric = (key: string) => {
    setSelectedMetrics((s) => (s.includes(key) ? s.filter((x) => x !== key) : [...s, key]));
  };

  // Inline spinner used for per-cell loading indicators
  function Spinner({ size = 16 }: { size?: number }) {
    return (
      <div
        role="status"
        aria-hidden={false}
        className="inline-block animate-spin"
        style={{
          width: size,
          height: size,
          borderWidth: 2,
          borderStyle: "solid",
          borderColor: "rgba(148,163,184,1)",
          borderTopColor: "transparent",
          borderRadius: "9999px",
        }}
      />
    );
  }

  const diskObj: any = metrics?.disk;
  const networkObj: any = metrics?.network;
  const loadAvg: any = metrics?.loadAvg;

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle>System Monitor</CardTitle>
        <CardDescription>View system performance metrics</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 flex flex-col">
        <div className="mb-4">
          <div className="flex items-center gap-4 mb-2">
          <label>Poll interval (ms):</label>
          <input
            type="number"
            value={intervalMs}
            onChange={(e) => {
              const v = Number(e.target.value) || 1000;
              setIntervalMs(v);
              polling.setIntervalMs(v);
            }}
            className="border rounded px-2 py-1 w-32"
          />

          <button
            className="px-3 py-1 rounded bg-slate-700 text-white"
            onClick={() => (polling.isRunning ? polling.stop() : polling.start())}
          >
            {polling.isRunning ? "Stop" : "Start"}
          </button>

          <div className="ml-4 flex items-center gap-2">
            <label className="font-medium">Metrics:</label>
            <label><input type="checkbox" checked={selectedMetrics.includes("cpuUsage")} onChange={() => toggleMetric("cpuUsage")} /> CPU</label>
            <label><input type="checkbox" checked={selectedMetrics.includes("memoryUsage")} onChange={() => toggleMetric("memoryUsage")} /> MEM</label>
            <label><input type="checkbox" checked={selectedMetrics.includes("loadAvg")} onChange={() => toggleMetric("loadAvg")} /> Load</label>
          </div>
        </div>
        <Accordion>
          <AccordionItem>
            <AccordionTrigger onClick={() => setOpenPanel(openPanel === "system" ? null : "system")} open={openPanel === "system"}>
              <div className="font-medium">System Stats</div>
            </AccordionTrigger>
            <AccordionContent open={openPanel === "system"}>
              <SysmonCharts points={polling.points} metricKeys={selectedMetrics} />
            </AccordionContent>
          </AccordionItem>
        </Accordion>

        <MessageStatsCharts points={polling.points} openPanel={openPanel} setOpenPanel={setOpenPanel} />
        </div>

        <div className="flex-1 min-h-0 overflow-auto">
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
          <TableCell>{isLoading ? <Spinner /> : `${formatNumber(metrics?.cpuUsage)}%`}</TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Memory Usage</TableCell>
          <TableCell>{isLoading ? <Spinner /> : `${formatNumber(metrics?.memoryUsage)}%`}</TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Load Average</TableCell>
          <TableCell>
            {isLoading ? (
              <Spinner />
            ) : Array.isArray(loadAvg) ? (
              loadAvg.map((v: number) => v.toFixed(2)).join(", ")
            ) : typeof loadAvg === "number" ? (
              loadAvg.toFixed(2)
            ) : (
              "—"
            )}
          </TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Uptime</TableCell>
          <TableCell>{isLoading ? <Spinner /> : formatUptime(metrics?.uptime)}</TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Disk Usage (total)</TableCell>
          <TableCell>{isLoading ? <Spinner /> : metrics?.diskTotal ? formatBytes(metrics.diskTotal) : "—"}</TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Disk Used</TableCell>
          <TableCell>{isLoading ? <Spinner /> : metrics?.diskUsage ? formatBytes(metrics.diskUsage) : "—"}</TableCell>
        </TableRow>

        {diskObj && typeof diskObj === "object" && Object.keys(diskObj).length > 0 && (
          <>
            {Object.entries(diskObj).map(([path, info]: any) => (
              <TableRow key={path}>
                <TableCell>Disk {path}</TableCell>
                <TableCell>
                  {isLoading ? (
                    <Spinner />
                  ) : (
                    <>
                      {info?.used ? formatBytes(info.used) : "—"} / {info?.total ? formatBytes(info.total) : "—"}
                    </>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </>
        )}

        <TableRow>
          <TableCell>Swap Usage</TableCell>
          <TableCell>{isLoading ? <Spinner /> : metrics?.swapUsage ? formatNumber(metrics.swapUsage) + '%' : '—'}</TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Network RX</TableCell>
          <TableCell>{isLoading ? <Spinner /> : metrics?.netRx ? formatBytes(metrics.netRx) : '—'}</TableCell>
        </TableRow>

        <TableRow>
          <TableCell>Network TX</TableCell>
          <TableCell>{isLoading ? <Spinner /> : metrics?.netTx ? formatBytes(metrics.netTx) : '—'}</TableCell>
        </TableRow>

        {networkObj && typeof networkObj === "object" && Object.keys(networkObj).length > 0 && (
          <>
            {Object.entries(networkObj).map(([ifName, info]: any) => (
              <TableRow key={ifName}>
                <TableCell>Net {ifName}</TableCell>
                <TableCell>{isLoading ? <Spinner /> : <>RX: {info?.rx ? formatBytes(info.rx) : '—'} / TX: {info?.tx ? formatBytes(info.tx) : '—'}</>}</TableCell>
              </TableRow>
            ))}
          </>
        )}
      </TableBody>
      </Table>
      </div>
      </CardContent>
    </Card>
  );
}