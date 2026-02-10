/**
 * GLC Offline Detector
 *
 * Provides online/offline detection using:
 * - navigator.onLine property
 * - 'online' and 'offline' browser events
 * - Optional periodic connectivity checks
 */

// ============================================================================
// Types
// ============================================================================

export type OnlineStatus = 'online' | 'offline' | 'checking';

export interface OnlineStatusListener {
  (status: OnlineStatus): void;
}

export interface ConnectivityCheckOptions {
  /** URL to ping for connectivity check */
  url?: string;
  /** Interval in ms between checks when offline (default: 30000) */
  interval?: number;
  /** Timeout in ms for connectivity check (default: 5000) */
  timeout?: number;
}

// ============================================================================
// Constants
// ============================================================================

const DEFAULT_CHECK_INTERVAL = 30000; // 30 seconds
const DEFAULT_CHECK_TIMEOUT = 5000; // 5 seconds
const DEFAULT_CHECK_URL = '/restful/health';

// ============================================================================
// Online Status Detector Class
// ============================================================================

/**
 * Detector for online/offline status
 *
 * Usage:
 * ```ts
 * const detector = new OnlineStatusDetector();
 *
 * // Subscribe to status changes
 * const unsubscribe = detector.subscribe((status) => {
 *   console.log('Status changed:', status);
 * });
 *
 * // Get current status
 * console.log(detector.status);
 *
 * // Cleanup
 * unsubscribe();
 * detector.destroy();
 * ```
 */
export class OnlineStatusDetector {
  private listeners: Set<OnlineStatusListener> = new Set();
  private _status: OnlineStatus;
  private checkInterval: ReturnType<typeof setInterval> | null = null;
  private options: Required<ConnectivityCheckOptions>;
  private isChecking = false;

  constructor(options?: ConnectivityCheckOptions) {
    this.options = {
      url: options?.url ?? DEFAULT_CHECK_URL,
      interval: options?.interval ?? DEFAULT_CHECK_INTERVAL,
      timeout: options?.timeout ?? DEFAULT_CHECK_TIMEOUT,
    };

    // Initialize status from navigator.onLine
    this._status = typeof navigator !== 'undefined' && navigator.onLine ? 'online' : 'offline';

    // Bind event handlers
    this.handleOnline = this.handleOnline.bind(this);
    this.handleOffline = this.handleOffline.bind(this);

    // Attach event listeners if in browser
    if (typeof window !== 'undefined') {
      window.addEventListener('online', this.handleOnline);
      window.addEventListener('offline', this.handleOffline);

      // Start periodic connectivity check when offline
      if (!navigator.onLine) {
        this.startConnectivityCheck();
      }
    }
  }

  /**
   * Get current online status
   */
  get status(): OnlineStatus {
    return this._status;
  }

  /**
   * Check if currently online
   */
  get isOnline(): boolean {
    return this._status === 'online';
  }

  /**
   * Check if currently offline
   */
  get isOffline(): boolean {
    return this._status === 'offline';
  }

  /**
   * Subscribe to status changes
   * @returns Unsubscribe function
   */
  subscribe(listener: OnlineStatusListener): () => void {
    this.listeners.add(listener);

    // Immediately notify of current status
    listener(this._status);

    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Manually trigger a connectivity check
   */
  async checkConnectivity(): Promise<boolean> {
    if (this.isChecking) {
      return this._status === 'online';
    }

    this.isChecking = true;
    const previousStatus = this._status;

    // Set status to checking
    this._status = 'checking';
    this.notifyListeners();

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.options.timeout);

      await fetch(this.options.url, {
        method: 'HEAD',
        mode: 'no-cors', // Allow cross-origin requests
        signal: controller.signal,
        cache: 'no-store',
      });

      clearTimeout(timeoutId);

      // For no-cors requests, response.ok is always false
      // We consider the request successful if it didn't throw
      const isOnline = true;
      this._status = isOnline ? 'online' : 'offline';

      if (isOnline && previousStatus === 'offline') {
        // Just came back online
        this.stopConnectivityCheck();
      } else if (!isOnline && previousStatus === 'online') {
        // Just went offline
        this.startConnectivityCheck();
      }

      this.notifyListeners();
      return isOnline;
    } catch {
      // Network error - we're offline
      this._status = 'offline';

      if (previousStatus !== 'offline') {
        this.startConnectivityCheck();
      }

      this.notifyListeners();
      return false;
    } finally {
      this.isChecking = false;
    }
  }

  /**
   * Destroy the detector and cleanup resources
   */
  destroy(): void {
    if (typeof window !== 'undefined') {
      window.removeEventListener('online', this.handleOnline);
      window.removeEventListener('offline', this.handleOffline);
    }
    this.stopConnectivityCheck();
    this.listeners.clear();
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private handleOnline(): void {
    const previousStatus = this._status;
    this._status = 'online';
    this.stopConnectivityCheck();

    if (previousStatus !== 'online') {
      this.notifyListeners();
    }
  }

  private handleOffline(): void {
    const previousStatus = this._status;
    this._status = 'offline';
    this.startConnectivityCheck();

    if (previousStatus !== 'offline') {
      this.notifyListeners();
    }
  }

  private notifyListeners(): void {
    this.listeners.forEach((listener) => {
      try {
        listener(this._status);
      } catch (error) {
        console.error('Error in online status listener:', error);
      }
    });
  }

  private startConnectivityCheck(): void {
    if (this.checkInterval) {
      return;
    }

    this.checkInterval = setInterval(() => {
      this.checkConnectivity();
    }, this.options.interval);
  }

  private stopConnectivityCheck(): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
      this.checkInterval = null;
    }
  }
}

// ============================================================================
// Singleton Instance
// ============================================================================

let detectorInstance: OnlineStatusDetector | null = null;

/**
 * Get the singleton online status detector
 */
export function getOnlineStatusDetector(): OnlineStatusDetector {
  if (!detectorInstance) {
    detectorInstance = new OnlineStatusDetector();
  }
  return detectorInstance;
}

/**
 * Destroy the singleton detector (useful for testing)
 */
export function destroyOnlineStatusDetector(): void {
  if (detectorInstance) {
    detectorInstance.destroy();
    detectorInstance = null;
  }
}

// ============================================================================
// React Hook
// ============================================================================

/**
 * Hook to get the current online status
 *
 * Usage:
 * ```tsx
 * function MyComponent() {
 *   const { isOnline, isOffline, status } = useOnlineStatus();
 *
 *   return (
 *     <div>
 *       Status: {status}
 *       {isOffline && <p>You are offline</p>}
 *     </div>
 *   );
 * }
 * ```
 */
export function useOnlineStatus(): {
  status: OnlineStatus;
  isOnline: boolean;
  isOffline: boolean;
  checkConnectivity: () => Promise<boolean>;
} {
  // This is a placeholder - the actual hook is in ./hooks.ts
  // This file provides the core logic without React dependency
  throw new Error('useOnlineStatus hook is defined in ./hooks.ts');
}
