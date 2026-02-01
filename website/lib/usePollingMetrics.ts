import { useEffect, useRef, useState, useCallback } from "react";
import { rpcClient } from "./rpc-client";
import type { SysMetrics } from "./types";
import { createLogger } from "./logger";

const logger = createLogger("usePollingMetrics");

export interface SysMetricsPoint {
  ts: number;
  metrics: SysMetrics;
}

// keep up to `maxPoints` samples; default 60
export function usePollingMetrics(intervalMs = 3000, maxPoints = 60, autoStart = true) {
  const [points, setPoints] = useState<SysMetricsPoint[]>([]);
  const [isRunning, setIsRunning] = useState<boolean>(false);
  const timerRef = useRef<number | null>(null);
  const intervalRef = useRef(intervalMs);

  const fetchOnce = useCallback(async () => {
    try {
      const resp = await rpcClient.getSysMetrics();
      if (resp.retcode === 0 && resp.payload) {
        const point: SysMetricsPoint = { ts: Date.now(), metrics: resp.payload as SysMetrics };
        setPoints((p) => {
          const next = p.concat(point);
          if (next.length > maxPoints) next.shift();
          return next;
        });
      }
    } catch (err) {
      // swallow - consumer can surface errors if needed
      logger.error("Metrics fetch error", err, { intervalMs, maxPoints });
    }
  }, [maxPoints]);

  const start = useCallback(() => {
    if (isRunning) return;
    setIsRunning(true);
    fetchOnce();
    timerRef.current = window.setInterval(() => fetchOnce(), intervalRef.current);
  }, [fetchOnce]);

  const stop = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
    setIsRunning(false);
  }, []);

  const setIntervalMs = useCallback((ms: number) => {
    intervalRef.current = ms;
    if (isRunning) {
      stop();
      // restart with new interval
      timerRef.current = window.setInterval(() => fetchOnce(), intervalRef.current);
      setIsRunning(true);
    }
  }, [fetchOnce, stop]);

  useEffect(() => {
    // auto-start if requested
    if (autoStart) start();
    return () => stop();
  }, [start, stop]);

  return {
    points,
    start,
    stop,
    setIntervalMs,
    isRunning,
  };
}
