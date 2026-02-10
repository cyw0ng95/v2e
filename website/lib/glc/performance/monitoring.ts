/**
 * GLC Performance Utilities
 *
 * FPS monitoring, memoization helpers, and optimization utilities
 */

/**
 * FPS Monitor
 */
export class FPSMonitor {
  private frames: number[] = [];
  private lastFrameTime: number = 0;
  private running = false;

  start(): void {
    this.running = true;
    this.lastFrameTime = performance.now();
    this.tick();
  }

  stop(): void {
    this.running = false;
  }

  private tick = (): void => {
    if (!this.running) return;

    const now = performance.now();
    const delta = now - this.lastFrameTime;
    this.lastFrameTime = now;

    this.frames.push(delta);
    if (this.frames.length > 60) {
      this.frames.shift();
    }

    requestAnimationFrame(this.tick);
  };

  getFPS(): number {
    if (this.frames.length === 0) return 0;
    const avgDelta = this.frames.reduce((a, b) => a + b, 0) / this.frames.length;
    return Math.round(1000 / avgDelta);
  }

  isHealthy(): boolean {
    return this.getFPS() >= 55;
  }
}

/**
 * Debounce function
 */
export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  return (...args: Parameters<T>) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    timeoutId = setTimeout(() => {
      fn(...args);
      timeoutId = null;
    }, delay);
  };
}

/**
 * Throttle function
 */
export function throttle<T extends (...args: unknown[]) => unknown>(
  fn: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle = false;

  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      fn(...args);
      inThrottle = true;
      setTimeout(() => {
        inThrottle = false;
      }, limit);
    }
  };
}

/**
 * Batch updates using requestAnimationFrame
 */
export function batchUpdate(callback: () => void): void {
  requestAnimationFrame(() => {
    callback();
  });
}

/**
 * Check if element is visible in viewport
 */
export function isInViewport(element: HTMLElement): boolean {
  const rect = element.getBoundingClientRect();
  return (
    rect.top >= 0 &&
    rect.left >= 0 &&
    rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
    rect.right <= (window.innerWidth || document.documentElement.clientWidth)
  );
}

/**
 * Create intersection observer for lazy loading
 */
export function createVisibilityObserver(
  callback: (entry: IntersectionObserverEntry) => void,
  options?: IntersectionObserverInit
): IntersectionObserver {
  return new IntersectionObserver((entries) => {
    entries.forEach(callback);
  }, options);
}

/**
 * Measure render time
 */
export function measureRender(name: string): () => number {
  const start = performance.now();
  return () => {
    const duration = performance.now() - start;
    if (process.env.NODE_ENV === 'development') {
      console.log(`[GLC] ${name} rendered in ${duration.toFixed(2)}ms`);
    }
    return duration;
  };
}
