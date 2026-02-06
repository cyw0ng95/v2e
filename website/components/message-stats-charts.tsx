"use client";

import React, { useMemo, useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import type { SysMetricsPoint } from "@/lib/usePollingMetrics";
import Accordion, { AccordionContent, AccordionItem, AccordionTrigger } from "./ui/accordion";

function tsLabel(ts: number) {
  const d = new Date(ts);
  return d.toLocaleTimeString();
}

export default function MessageStatsCharts({ points, openPanel, setOpenPanel }: { points: SysMetricsPoint[]; openPanel: "system" | "global" | "proc" | null; setOpenPanel: (p: "system" | "global" | "proc" | null) => void; }) {
  // Build series for totals and per-process
  const data = useMemo(() => {
    const toSnake = (s: string) => s.replace(/[A-Z]/g, (ch) => `_${ch.toLowerCase()}`);
    const normalizeKey = (k: string) => (k.includes("_") ? k : toSnake(k));

    return points.map((p, idx) => {
      const out: any = { ts: p.ts };
      const ms: any = (p.metrics as any)?.messageStats || (p.metrics as any)?.message_stats;
      if (ms && ms.total) {
        Object.entries(ms.total).forEach(([k, v]) => {
          if (typeof v === "number") {
            const snake = normalizeKey(k);
            out[`total_${snake}`] = v;
          }
        });
      }
      // compute total accumulated and speed
      const prev = idx > 0 ? points[idx - 1] : null;
      const prevMs: any = prev ? (prev.metrics as any)?.messageStats || (prev.metrics as any)?.message_stats : null;
      const totalSent = ms?.total?.total_sent ?? ms?.total?.totalSent ?? ms?.total_sent ?? 0;
      const totalReceived = ms?.total?.total_received ?? ms?.total?.totalReceived ?? ms?.total_received ?? 0;
      const totalAccum = (Number(totalSent) || 0) + (Number(totalReceived) || 0);
      out[`total_accumulated`] = totalAccum;
      if (prevMs && prevMs.total) {
        const prevSent = prevMs.total.total_sent ?? prevMs.total.totalSent ?? prevMs.total_sent ?? 0;
        const prevReceived = prevMs.total.total_received ?? prevMs.total.totalReceived ?? prevMs.total_received ?? 0;
        const prevAccum = (Number(prevSent) || 0) + (Number(prevReceived) || 0);
        const dtSec = prev ? (p.ts - prev.ts) / 1000 : 0;
        out[`total_speed`] = dtSec > 0 ? (totalAccum - prevAccum) / dtSec : 0;
      } else {
        out[`total_speed`] = 0;
      }

      const perProc = ms && (ms.perProcess || ms.per_process);
      const prevPerProc = prevMs && (prevMs.perProcess || prevMs.per_process);

      if (perProc) {
        Object.entries(perProc).forEach(([pid, stats]: any) => {
          if (!stats) return;
          const sent = (stats.total_sent ?? stats.totalSent ?? 0) as number;
          const received = (stats.total_received ?? stats.totalReceived ?? 0) as number;
          const accum = sent + received;
          out[`p_${pid}_accumulated`] = accum;

          // speed: delta accumulated / delta seconds
          if (prevPerProc && prevPerProc[pid]) {
            const prevStats: any = prevPerProc[pid];
            const prevSent = (prevStats.total_sent ?? prevStats.totalSent ?? 0) as number;
            const prevReceived = (prevStats.total_received ?? prevStats.totalReceived ?? 0) as number;
            const prevAccum = prevSent + prevReceived;
            const dtSec = prev ? (p.ts - prev.ts) / 1000 : 0;
            out[`p_${pid}_speed`] = dtSec > 0 ? (accum - prevAccum) / dtSec : 0;
          } else {
            out[`p_${pid}_speed`] = 0;
          }

          // keep explicit sent/received keys too
          out[`p_${pid}_total_sent`] = sent;
          out[`p_${pid}_total_received`] = received;
        });
      }

      return out;
    });
  }, [points]);

  const procList = useMemo(() => {
    const set = new Set<string>();
    for (const p of points) {
      const ms: any = (p.metrics as any)?.messageStats || (p.metrics as any)?.message_stats;
      const perProc = ms && (ms.perProcess || ms.per_process);
      if (perProc) {
        Object.keys(perProc).forEach((k) => set.add(k));
      }
    }
    return Array.from(set);
  }, [points]);
  const [selectedProcs, setSelectedProcs] = useState<string[]>([]);
  type Metric = "accumulated" | "speed";
  const [procMetric, setProcMetric] = useState<Metric>("accumulated");
  const [totalMetric, setTotalMetric] = useState<Metric>("accumulated");
  // openPanel and setOpenPanel are provided by parent to ensure only one panel
  // across the page is open at a time.

  React.useEffect(() => {
    if (selectedProcs.length === 0 && procList.length > 0) setSelectedProcs(procList.slice(0, 3));
  }, [procList, selectedProcs.length]);

  // colors
  const colors = ["#8884d8", "#82ca9d", "#ff7300", "#387908", "#ff0000", "#00aaff", "#aa00ff"];

  return (
    <div className="mb-4">
      <Accordion>
        <AccordionItem>
          <AccordionTrigger onClick={() => setOpenPanel(openPanel === "global" ? null : "global")} open={openPanel === "global"}>
            <div className="flex items-center gap-3">
              <div className="font-medium">Global RPC Stats</div>
            </div>
          </AccordionTrigger>
          <AccordionContent open={openPanel === "global"}>
            <div className="mb-2 flex items-center gap-2">
              <label className="text-sm">Metric:</label>
              <select value={totalMetric} onChange={(e) => setTotalMetric(e.target.value as any)} className="border rounded px-2 py-1 text-sm">
                <option value="accumulated">accumulated (in+out)</option>
                <option value="speed">speed (per sec)</option>
              </select>
            </div>
            <div style={{ width: "100%", height: 200 }} className="mb-4">
              {openPanel === "global" && (
                <ResponsiveContainer>
                  <LineChart data={data} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="ts" tickFormatter={tsLabel} />
                    <YAxis />
                    <Tooltip labelFormatter={(v) => tsLabel(Number(v))} />
                    <Legend />
                    <Line type="monotone" dataKey={`total_${totalMetric}`} stroke={colors[0]} dot={false} name={totalMetric} />
                    <Line type="monotone" dataKey="total_request_count" stroke={colors[2]} dot={false} name="request_count" />
                    <Line type="monotone" dataKey="total_response_count" stroke={colors[3]} dot={false} name="response_count" />
                  </LineChart>
                </ResponsiveContainer>
              )}
            </div>
          </AccordionContent>
        </AccordionItem>

        <AccordionItem>
          <AccordionTrigger onClick={() => setOpenPanel(openPanel === "proc" ? null : "proc")} open={openPanel === "proc"}>
            <div className="font-medium">Process RPC Stats</div>
          </AccordionTrigger>
          <AccordionContent open={openPanel === "proc"}>
            <div className="flex items-center gap-3 mb-2">
              <div className="ml-auto flex items-center gap-2">
                <label className="text-sm">Metric:</label>
                <select value={procMetric} onChange={(e) => setProcMetric(e.target.value as any)} className="border rounded px-2 py-1 text-sm">
                  <option value="accumulated">accumulated (in+out)</option>
                  <option value="speed">speed (per sec)</option>
                </select>
              </div>
            </div>
            <div className="flex gap-2 items-center mb-2 flex-wrap">
              {procList.map((p) => (
                <label key={p} className="text-sm">
                  <input
                    type="checkbox"
                    checked={selectedProcs.includes(p)}
                    onChange={() => setSelectedProcs((s) => (s.includes(p) ? s.filter((x) => x !== p) : [...s, p]))}
                    className="mr-1"
                  />
                  {p}
                </label>
              ))}
            </div>
            <div style={{ width: "100%", height: 260 }}>
              {openPanel === "proc" && (
                <ResponsiveContainer>
                  <LineChart data={data} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="ts" tickFormatter={tsLabel} />
                    <YAxis />
                    <Tooltip labelFormatter={(v) => tsLabel(Number(v))} />
                    <Legend />
                    {selectedProcs.map((p, idx) => {
                      const key = `p_${p}_` + procMetric;
                      return <Line key={p} type="monotone" dataKey={key} stroke={colors[idx % colors.length]} dot={false} name={`${p}:${procMetric.replace("total_", "")}`} />;
                    })}
                  </LineChart>
                </ResponsiveContainer>
              )}
            </div>
          </AccordionContent>
        </AccordionItem>
      </Accordion>

      
    </div>
  );
}
