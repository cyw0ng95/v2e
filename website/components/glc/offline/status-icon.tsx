'use client';

/**
 * GLC Offline Status Icon
 *
 * A status indicator that shows online/offline state and pending operations.
 */

import { useCallback } from 'react';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Badge } from '@/components/ui/badge';
import {
  Wifi,
  WifiOff,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  useOnlineStatus,
  useOperationQueue,
  useSyncStatus,
} from '@/lib/glc/offline/hooks';

// ============================================================================
// Types
// ============================================================================

export interface OfflineStatusIconProps {
  /** Show as a button that opens a dropdown menu */
  showDropdown?: boolean;
  /** Show pending count badge */
  showBadge?: boolean;
  /** Custom class name */
  className?: string;
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Callback when sync is triggered */
  onSync?: () => void;
}

// ============================================================================
// Helper Components
// ============================================================================

interface StatusIndicatorProps {
  isOffline: boolean;
  isSyncing: boolean;
  syncStatus: string;
  size: 'sm' | 'md' | 'lg';
}

function StatusIndicator({
  isOffline,
  isSyncing,
  syncStatus,
  size,
}: StatusIndicatorProps) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-5 w-5',
    lg: 'h-6 w-6',
  };

  const baseClass = sizeClasses[size];

  if (isOffline) {
    return <WifiOff className={cn(baseClass, 'text-amber-500')} />;
  }

  if (isSyncing) {
    return <RefreshCw className={cn(baseClass, 'text-blue-500 animate-spin')} />;
  }

  if (syncStatus === 'error') {
    return <AlertCircle className={cn(baseClass, 'text-red-500')} />;
  }

  if (syncStatus === 'conflict') {
    return <AlertCircle className={cn(baseClass, 'text-amber-500')} />;
  }

  return <Wifi className={cn(baseClass, 'text-green-500')} />;
}

function getStatusLabel(
  isOnline: boolean,
  isOffline: boolean,
  isSyncing: boolean,
  syncStatus: string,
  pendingCount: number
): string {
  if (isOffline) {
    return `Offline${pendingCount > 0 ? ` - ${pendingCount} pending` : ''}`;
  }

  if (isSyncing) {
    return 'Syncing...';
  }

  if (syncStatus === 'error') {
    return 'Sync failed';
  }

  if (syncStatus === 'conflict') {
    return 'Sync conflict';
  }

  if (pendingCount > 0) {
    return `${pendingCount} changes pending`;
  }

  return 'Online - All synced';
}

// ============================================================================
// Main Component
// ============================================================================

export function OfflineStatusIcon({
  showDropdown = true,
  showBadge = true,
  className,
  size = 'md',
  onSync,
}: OfflineStatusIconProps) {
  const { isOnline, isOffline } = useOnlineStatus();
  const { pendingCount, hasPending } = useOperationQueue();
  const { state: syncState, isSyncing, sync } = useSyncStatus();

  const handleSync = useCallback(async () => {
    await sync();
    onSync?.();
  }, [sync, onSync]);

  const icon = (
    <StatusIndicator
      isOffline={isOffline}
      isSyncing={isSyncing}
      syncStatus={syncState.status}
      size={size}
    />
  );

  const label = getStatusLabel(
    isOnline,
    isOffline,
    isSyncing,
    syncState.status,
    pendingCount
  );

  // Simple tooltip version
  if (!showDropdown) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className={cn('relative', className)}>
              {icon}
              {showBadge && hasPending && (
                <Badge
                  variant="destructive"
                  className="absolute -top-1 -right-1 h-4 min-w-4 px-1 text-[10px] flex items-center justify-center"
                >
                  {pendingCount > 99 ? '99+' : pendingCount}
                </Badge>
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <p>{label}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  // Dropdown version with actions
  return (
    <TooltipProvider>
      <DropdownMenu>
        <Tooltip>
          <TooltipTrigger asChild>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className={cn('relative', className)}
              >
                {icon}
                {showBadge && hasPending && (
                  <Badge
                    variant="destructive"
                    className="absolute -top-1 -right-1 h-4 min-w-4 px-1 text-[10px] flex items-center justify-center"
                  >
                    {pendingCount > 99 ? '99+' : pendingCount}
                  </Badge>
                )}
              </Button>
            </DropdownMenuTrigger>
          </TooltipTrigger>
          <TooltipContent>
            <p>{label}</p>
          </TooltipContent>
        </Tooltip>

        <DropdownMenuContent align="end" className="w-56">
          <DropdownMenuLabel className="flex items-center gap-2">
            {isOffline ? (
              <>
                <WifiOff className="h-4 w-4 text-amber-500" />
                <span>Offline</span>
              </>
            ) : (
              <>
                <Wifi className="h-4 w-4 text-green-500" />
                <span>Online</span>
              </>
            )}
          </DropdownMenuLabel>

          <DropdownMenuSeparator />

          {hasPending && (
            <>
              <DropdownMenuLabel className="text-xs text-muted-foreground font-normal">
                {pendingCount} {pendingCount === 1 ? 'change' : 'changes'} waiting to sync
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
            </>
          )}

          {isOnline && hasPending && (
            <DropdownMenuItem onClick={handleSync} disabled={isSyncing}>
              {isSyncing ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Syncing...
                </>
              ) : (
                <>
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Sync Now
                </>
              )}
            </DropdownMenuItem>
          )}

          {syncState.status === 'error' && (
            <DropdownMenuItem onClick={handleSync} disabled={isSyncing}>
              <AlertCircle className="h-4 w-4 mr-2 text-red-500" />
              Retry Sync
            </DropdownMenuItem>
          )}

          {syncState.status === 'conflict' && (
            <DropdownMenuItem>
              <AlertCircle className="h-4 w-4 mr-2 text-amber-500" />
              Resolve Conflict
            </DropdownMenuItem>
          )}

          {!hasPending && isOnline && (
            <div className="px-2 py-1.5 text-xs text-muted-foreground flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-green-500" />
              All changes synced
            </div>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </TooltipProvider>
  );
}

// ============================================================================
// Compact Status Icon
// ============================================================================

export interface CompactStatusIconProps {
  className?: string;
}

/**
 * A minimal status icon without dropdown, just shows the icon and badge
 */
export function CompactStatusIcon({ className }: CompactStatusIconProps) {
  return (
    <OfflineStatusIcon
      showDropdown={false}
      showBadge={true}
      size="sm"
      className={className}
    />
  );
}

// ============================================================================
// Export
// ============================================================================

export default OfflineStatusIcon;
