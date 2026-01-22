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

export interface SysmonChartsProps {
  points: SysMetricsPoint[];
  metricKeys?: string[]; // e.g., ["cpuUsage", "memoryUsage"]
}

function tsLabel(ts: number) {
  const d = new Date(ts);
  return d.toLocaleTimeString();
}

export default function SysmonCharts({ points, metricKeys = ["cpuUsage", "memoryUsage"] }: SysmonChartsProps) {
  const data = useMemo(() => {
    return points.map((p) => {
      const base: any = { ts: p.ts };
      metricKeys.forEach((k) => {
        // support nested like loadAvg (array) -> use first element
        const v: any = (p.metrics as any)[k];
        if (Array.isArray(v)) base[k] = v[0];
        else base[k] = v ?? null;
      });
      return base;
    });
  }, [points, metricKeys]);

  const colors = ["#8884d8", "#82ca9d", "#ff7300", "#387908"];

  return (
    <div style={{ width: "100%", height: 240 }}>
      <ResponsiveContainer>
        <LineChart data={data} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="ts" tickFormatter={tsLabel} />
          <YAxis />
          <Tooltip labelFormatter={(v) => tsLabel(Number(v))} />
          <Legend />
          {metricKeys.map((k, idx) => (
            <Line key={k} type="monotone" dataKey={k} stroke={colors[idx % colors.length]} dot={false} />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
