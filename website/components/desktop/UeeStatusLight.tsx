'use client';

import React, { useCallback } from 'react';
import { Activity, AlertTriangle, XCircle, HelpCircle } from 'lucide-react';
import { useUeeStatus } from '@/lib/hooks';
import { useDesktopStore } from '@/lib/desktop/store';
import { getAppById } from '@/lib/desktop/app-registry';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { WindowState } from '@/types/desktop';

type UeeStatusValue = 'healthy' | 'degraded' | 'error' | 'unknown';

interface UeeStatusLightProps {
  className?: string;
}

const statusConfig: Record<UeeStatusValue, { color: string; bgColor: string; label: string; Icon: React.ComponentType<{ className?: string }> }> = {
  healthy: {
    color: 'bg-green-500',
    bgColor: 'bg-green-500/20',
    label: 'UEE: All systems operational',
    Icon: Activity,
  },
  degraded: {
    color: 'bg-yellow-500',
    bgColor: 'bg-yellow-500/20',
    label: 'UEE: Some providers degraded',
    Icon: AlertTriangle,
  },
  error: {
    color: 'bg-red-500',
    bgColor: 'bg-red-500/20',
    label: 'UEE: Provider error detected',
    Icon: XCircle,
  },
  unknown: {
    color: 'bg-gray-400',
    bgColor: 'bg-gray-400/20',
    label: 'UEE: Status unknown',
    Icon: HelpCircle,
  },
};

export function UeeStatusLight({ className }: UeeStatusLightProps) {
  const { status, isLoading } = useUeeStatus(5000);
  const { isOnline } = useDesktopStore();
  const { openWindow } = useDesktopStore();
  
  const config = statusConfig[status];
  const Icon = config.Icon;

  const handleClick = useCallback(() => {
    const app = getAppById('etl');
    if (app) {
      openWindow({
        appId: 'etl',
        title: app.name,
        position: { x: 100, y: 50 },
        size: { width: app.defaultWidth, height: app.defaultHeight },
        minWidth: app.minWidth,
        minHeight: app.minHeight,
        maxWidth: app.maxWidth,
        maxHeight: app.maxHeight,
        isFocused: true,
        isMinimized: false,
        isMaximized: false,
        state: WindowState.Open,
      });
    }
  }, [openWindow]);

  const displayStatus = isOnline ? status : 'unknown';
  const displayConfig = statusConfig[displayStatus];

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            onClick={handleClick}
            className={cn(
              'relative flex items-center justify-center w-5 h-5 rounded-full cursor-pointer hover:scale-110 transition-transform',
              displayConfig.color,
              className
            )}
            aria-label={displayConfig.label}
          >
            <Icon className="w-3 h-3 text-white" />
          </button>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <p>{displayConfig.label}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export default UeeStatusLight;
