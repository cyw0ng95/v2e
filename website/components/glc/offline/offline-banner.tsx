'use client';

/**
 * GLC Offline Banner
 *
 * Displays a banner when the user is offline or has pending sync operations.
 * Shows sync status and provides a manual sync button.
 */

import { useCallback, useEffect, useState, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import {
  AlertCircle,
  CloudOff,
  RefreshCw,
  WifiOff,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useOfflineIndicator } from '@/lib/glc/offline/hooks';

// ============================================================================
// Types
// ============================================================================

export interface OfflineBannerProps {
  /** Custom class name */
  className?: string;
  /** Whether to show when online with pending changes */
  showPendingWhenOnline?: boolean;
  /** Callback when user clicks sync */
  onSync?: () => void;
  /** Callback when dismissed */
  onDismiss?: () => void;
  /** Whether the banner can be dismissed when online */
  dismissibleWhenOnline?: boolean;
}

// ============================================================================
// Main Component
// ============================================================================

export function OfflineBanner({
  className,
  showPendingWhenOnline = true,
  onSync,
  onDismiss,
  dismissibleWhenOnline = true,
}: OfflineBannerProps) {
  const {
    isOffline,
    isOnline,
    pendingCount,
    hasPending,
    isSyncing,
    syncProgress,
    syncStatus,
    sync,
  } = useOfflineIndicator();

  const [isDismissed, setIsDismissed] = useState(false);

  // Reset dismissed state when going offline - use a ref to avoid synchronous setState
  const prevIsOffline = useMemo(() => isOffline, [isOffline]);

  useEffect(() => {
    // Only reset when transitioning to offline, not on every render
    if (isOffline && !prevIsOffline) {
      // Use microtask to defer setState
      const timeoutId = setTimeout(() => {
        setIsDismissed(false);
      }, 0);
      return () => clearTimeout(timeoutId);
    }
  }, [isOffline, prevIsOffline]);

  const handleSync = useCallback(async () => {
    await sync();
    onSync?.();
  }, [sync, onSync]);

  const handleDismiss = useCallback(() => {
    setIsDismissed(true);
    onDismiss?.();
  }, [onDismiss]);

  // Compute visibility
  const shouldShow = useMemo(() => {
    // Don't show if online with no pending, or if dismissed while online
    if ((isOnline && !hasPending) || (isOnline && isDismissed)) {
      return false;
    }

    // Don't show if not showing pending when online
    if (isOnline && !showPendingWhenOnline) {
      return false;
    }

    return true;
  }, [isOnline, hasPending, isDismissed, showPendingWhenOnline]);

  if (!shouldShow) {
    return null;
  }

  return (
    <div
      className={cn(
        'fixed top-0 left-0 right-0 z-50',
        'flex items-center justify-center gap-3 px-4 py-2',
        'border-b shadow-sm',
        isOffline
          ? 'bg-amber-50 dark:bg-amber-950 border-amber-200 dark:border-amber-800'
          : 'bg-blue-50 dark:bg-blue-950 border-blue-200 dark:border-blue-800',
        className
      )}
    >
      {/* Icon */}
      <div className="flex-shrink-0">
        {isOffline ? (
          <WifiOff className="h-5 w-5 text-amber-600 dark:text-amber-400" />
        ) : isSyncing ? (
          <RefreshCw className="h-5 w-5 text-blue-600 dark:text-blue-400 animate-spin" />
        ) : syncStatus === 'error' ? (
          <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
        ) : (
          <CloudOff className="h-5 w-5 text-blue-600 dark:text-blue-400" />
        )}
      </div>

      {/* Message */}
      <div className="flex-1 text-sm">
        {isOffline ? (
          <span className="text-amber-800 dark:text-amber-200">
            You are offline. Changes will be saved locally and synced when you reconnect.
          </span>
        ) : isSyncing ? (
          <div className="flex items-center gap-2">
            <span className="text-blue-800 dark:text-blue-200">
              Syncing {pendingCount} changes...
            </span>
            <div className="h-1.5 w-24 bg-blue-200 dark:bg-blue-900 rounded-full overflow-hidden">
              <div
                className="h-full bg-blue-500 dark:bg-blue-400 transition-all duration-300"
                style={{ width: `${syncProgress}%` }}
              />
            </div>
          </div>
        ) : syncStatus === 'error' ? (
          <span className="text-red-800 dark:text-red-200">
            Failed to sync changes. Please try again.
          </span>
        ) : (
          <span className="text-blue-800 dark:text-blue-200">
            {pendingCount} {pendingCount === 1 ? 'change' : 'changes'} pending sync.
          </span>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        {isOnline && hasPending && !isSyncing && (
          <Button
            size="sm"
            variant="outline"
            onClick={handleSync}
            className="h-7 text-xs"
          >
            <RefreshCw className="h-3 w-3 mr-1" />
            Sync Now
          </Button>
        )}

        {isOnline && dismissibleWhenOnline && !isSyncing && (
          <Button
            size="sm"
            variant="ghost"
            onClick={handleDismiss}
            className="h-7 w-7 p-0"
          >
            <span className="sr-only">Dismiss</span>
            <span className="text-lg">&times;</span>
          </Button>
        )}
      </div>
    </div>
  );
}

// ============================================================================
// Compact Offline Banner
// ============================================================================

export interface CompactOfflineBannerProps {
  className?: string;
}

/**
 * A more compact version of the offline banner for tighter spaces
 */
export function CompactOfflineBanner({ className }: CompactOfflineBannerProps) {
  const { isOffline, pendingCount, hasPending, isSyncing, sync } =
    useOfflineIndicator();

  if (!isOffline && !hasPending) {
    return null;
  }

  return (
    <div
      className={cn(
        'flex items-center gap-2 px-3 py-1.5 rounded-md text-sm',
        isOffline
          ? 'bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-200'
          : 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
        className
      )}
    >
      {isOffline ? (
        <>
          <WifiOff className="h-4 w-4" />
          <span>Offline</span>
        </>
      ) : isSyncing ? (
        <>
          <RefreshCw className="h-4 w-4 animate-spin" />
          <span>Syncing...</span>
        </>
      ) : (
        <>
          <CloudOff className="h-4 w-4" />
          <span>{pendingCount} pending</span>
          <Button
            size="sm"
            variant="link"
            onClick={() => sync()}
            className="h-auto p-0 text-xs underline"
          >
            Sync
          </Button>
        </>
      )}
    </div>
  );
}

// ============================================================================
// Export
// ============================================================================

export default OfflineBanner;
