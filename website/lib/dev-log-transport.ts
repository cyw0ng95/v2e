/**
 * Development-only log transport from browser to server
 * Only active in development mode, only intercepts console.error
 */

const DEV_LOG_ENDPOINT = '/api/dev-logs';
const BATCH_SIZE = 5;
const FLUSH_INTERVAL = 2000;

interface LogEntry {
  message: string;
  stack?: string;
  url: string;
  timestamp: string;
}

class DevLogTransport {
  private logQueue: LogEntry[] = [];
  private flushTimer: ReturnType<typeof setInterval> | null = null;
  private originalError: typeof console.error;
  private isEnabled: boolean;

  constructor() {
    // Only enable in development mode
    this.isEnabled =
      typeof window !== 'undefined' &&
      (process.env.NODE_ENV === 'development') &&
      (process.env.NEXT_PUBLIC_API_BASE_URL !== undefined);

    if (this.isEnabled) {
      this.originalError = console.error;
      this.setupInterceptor();
      this.startFlushTimer();
    }
  }

  private setupInterceptor(): void {
    const self = this;
    console.error = function (...args: unknown[]) {
      // Call original console.error first
      self.originalError.apply(console, args);

      // Extract error details
      let message = '';
      let stack: string | undefined;

      for (const arg of args) {
        if (arg instanceof Error) {
          message = arg.message;
          stack = arg.stack;
        } else {
          message += String(arg) + ' ';
        }
      }

      self.addLog({
        message: message.trim(),
        stack,
        url: window.location.href,
        timestamp: new Date().toISOString(),
      });
    };
  }

  private addLog(entry: LogEntry): void {
    this.logQueue.push(entry);

    if (this.logQueue.length >= BATCH_SIZE) {
      this.flush();
    }
  }

  private flush(): void {
    if (this.logQueue.length === 0) return;

    const logsToSend = [...this.logQueue];
    this.logQueue = [];

    if (typeof window === 'undefined') return;

    const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || '';
    const endpoint = `${baseUrl}${DEV_LOG_ENDPOINT}`;

    // Use sendBeacon for reliable delivery, fallback to fetch
    const data = JSON.stringify({ logs: logsToSend });

    if (navigator.sendBeacon) {
      navigator.sendBeacon(endpoint, data);
    } else {
      fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: data,
        keepalive: true,
      }).catch(() => {
        // Silently fail - this is dev-only logging
      });
    }
  }

  private startFlushTimer(): void {
    this.flushTimer = setInterval(() => {
      this.flush();
    }, FLUSH_INTERVAL);
  }

  destroy(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }

    // Restore original console.error
    if (this.isEnabled) {
      console.error = this.originalError;
    }

    // Flush any remaining logs
    this.flush();
  }
}

let transportInstance: DevLogTransport | null = null;

export function initDevLogTransport(): void {
  if (typeof window !== 'undefined' && !transportInstance) {
    transportInstance = new DevLogTransport();
  }
}

export function destroyDevLogTransport(): void {
  if (transportInstance) {
    transportInstance.destroy();
    transportInstance = null;
  }
}
