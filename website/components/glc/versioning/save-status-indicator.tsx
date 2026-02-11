'use client';

/**
 * GLC Save Status Indicator
 *
 * Shows the current save status with visual feedback.
 */

import React, { useEffect, useState } from 'react';
import {
  Check,
  Loader2,
  AlertCircle,
  Save,
  Clock,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import type { AutoSaverState, SaveStatus } from '@/lib/glc/versioning';

interface SaveStatusIndicatorProps {
  state: AutoSaverState;
  className?: string;
  showLabel?: boolean;
  compact?: boolean;
}

/**
 * Get status configuration
 */
function getStatusConfig(status: SaveStatus) {
  switch (status) {
    case 'idle':
      return {
        icon: Save,
        color: 'text-gray-400',
        bgColor: 'bg-gray-100 dark:bg-gray-800',
        label: 'Ready',
      };
    case 'pending':
      return {
        icon: Clock,
        color: 'text-amber-500',
        bgColor: 'bg-amber-50 dark:bg-amber-900/30',
        label: 'Pending',
      };
    case 'saving':
      return {
        icon: Loader2,
        color: 'text-blue-500',
        bgColor: 'bg-blue-50 dark:bg-blue-900/30',
        label: 'Saving...',
        animate: true,
      };
    case 'saved':
      return {
        icon: Check,
        color: 'text-green-500',
        bgColor: 'bg-green-50 dark:bg-green-900/30',
        label: 'Saved',
      };
    case 'error':
      return {
        icon: AlertCircle,
        color: 'text-red-500',
        bgColor: 'bg-red-50 dark:bg-red-900/30',
        label: 'Error',
      };
    default:
      return {
        icon: Save,
        color: 'text-gray-400',
        bgColor: 'bg-gray-100 dark:bg-gray-800',
        label: 'Unknown',
      };
  }
}

/**
 * Format time ago
 */
function formatTimeAgo(timestamp: string | null): string {
  if (!timestamp) return '';

  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;

  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
  });
}

/**
 * Save Status Indicator Component
 */
export function SaveStatusIndicator({
  state,
  className = '',
  showLabel = true,
  compact = false,
}: SaveStatusIndicatorProps) {
  const config = getStatusConfig(state.status);
  const Icon = config.icon;
  const timeAgo = formatTimeAgo(state.lastSavedAt);

  const tooltipContent = state.error
    ? `Error: ${state.error}`
    : state.lastSavedAt
    ? `Last saved ${timeAgo} (v${state.lastVersion})`
    : 'Not yet saved';

  if (compact) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className={`flex items-center gap-1 ${className}`}>
              <Icon
                className={`h-3.5 w-3.5 ${config.color} ${
                  config.animate ? 'animate-spin' : ''
                }`}
              />
              {state.pendingChanges && state.status !== 'saving' && (
                <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <p>{tooltipContent}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Badge
            variant="outline"
            className={`${config.bgColor} ${config.color} border-transparent cursor-default ${className}`}
          >
            <div className="flex items-center gap-1.5">
              <Icon
                className={`h-3.5 w-3.5 ${config.animate ? 'animate-spin' : ''}`}
              />
              {showLabel && (
                <span className="text-xs font-medium">{config.label}</span>
              )}
              {state.pendingChanges && state.status !== 'saving' && (
                <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
              )}
            </div>
          </Badge>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <div className="space-y-1">
            <p>{tooltipContent}</p>
            {state.pendingChanges && (
              <p className="text-amber-400 text-xs">Unsaved changes</p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Save Status with detailed info
 */
export function SaveStatusDetailed({
  state,
  className = '',
}: {
  state: AutoSaverState;
  className?: string;
}) {
  const config = getStatusConfig(state.status);
  const Icon = config.icon;
  const timeAgo = formatTimeAgo(state.lastSavedAt);

  return (
    <div className={`flex items-center justify-between ${className}`}>
      <div className="flex items-center gap-2">
        <Icon
          className={`h-4 w-4 ${config.color} ${
            config.animate ? 'animate-spin' : ''
          }`}
        />
        <span className={`text-sm ${config.color}`}>{config.label}</span>
      </div>

      <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
        {state.lastVersion !== null && (
          <span>v{state.lastVersion}</span>
        )}
        {state.lastSavedAt && (
          <>
            <span>|</span>
            <span>{timeAgo}</span>
          </>
        )}
      </div>
    </div>
  );
}

/**
 * Hook for using save status in components
 */
export function useAutoSaveStatus(autoSaver: { subscribe: (listener: (state: AutoSaverState) => void) => () => void } | null) {
  const [state, setState] = useState<AutoSaverState>({
    status: 'idle',
    lastSavedAt: null,
    lastVersion: null,
    error: null,
    pendingChanges: false,
  });

  useEffect(() => {
    if (!autoSaver) return;

    const unsubscribe = autoSaver.subscribe(setState);
    return unsubscribe;
  }, [autoSaver]);

  return state;
}

export default SaveStatusIndicator;
