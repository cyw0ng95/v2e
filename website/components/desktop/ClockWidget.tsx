/**
 * v2e Portal - Clock Widget Component
 *
 * Simple clock widget for desktop
 * Phase 4: Desktop Widgets
 * Backend Independence: Works completely offline
 */

'use client';

import React, { useState, useEffect } from 'react';
import type { WidgetConfig } from '@/types/desktop';

/**
 * Clock widget component
 */
export function ClockWidget({ widget }: { widget: WidgetConfig }) {
  const [time, setTime] = useState(new Date());

  useEffect(() => {
    // Update time every second
    const interval = setInterval(() => {
      setTime(new Date());
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const formatTime = (date: Date): string => {
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    hour12: false,
    });
  };

  return (
    <div
      className="absolute bg-white/10 backdrop-blur-sm rounded-lg p-4 shadow-lg border border-white/20 pointer-events-auto"
      style={{
        left: `${widget.position.x}px`,
        top: `${widget.position.y}px`,
      }}
      role="region"
      aria-label="Clock widget"
    >
      <div className="text-center">
        <div className="text-4xl font-light text-gray-800">
          {formatTime(time)}
        </div>
        <div className="text-sm text-gray-500 mt-1">
          {time.toLocaleDateString('en-US', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric',
          })}
        </div>
      </div>
    </div>
  );
}
