'use client';

/**
 * GLC Canvas Page - Dynamic Route
 * /glc/[presetId] where presetId is d3fend, topo, etc.
 * DEPRECATED: Redirects to /desktop?app=glc&preset=X
 */

import { useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { getAppById } from '@/lib/desktop/app-registry';

export default function GLCCanvasPage() {
  const params = useParams();
  const router = useRouter();
  const presetId = params.presetId as string;

  useEffect(() => {
    // Redirect to desktop and open GLC app with preset
    router.replace('/desktop');

    // Open app window after redirect completes
    setTimeout(() => {
      const { openWindow } = require('@/lib/desktop/store').useDesktopStore.getState();
      const app = getAppById('glc');
      if (app) {
        openWindow({
          appId: app.id,
          title: app.name,
          position: {
            x: Math.max(0, (window.innerWidth - app.defaultWidth) / 2),
            y: Math.max(28, (window.innerHeight - app.defaultHeight) / 2),
          },
          size: {
            width: app.defaultWidth,
            height: app.defaultHeight,
          },
          minWidth: app.minWidth,
          minHeight: app.minHeight,
          maxWidth: app.maxWidth,
          maxHeight: app.maxHeight,
          isFocused: true,
          isMinimized: false,
          isMaximized: false,
          state: require('@/types/desktop').WindowState.Open,
          // Pass preset ID as extra data for the component to use
          initialPreset: presetId,
        });
      }
    }, 100);
  }, [router, presetId]);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <p className="text-lg text-gray-600">Redirecting to desktop...</p>
      </div>
    </div>
  );
}
