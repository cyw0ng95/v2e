'use client';

/**
 * Mcards Page
 * /mcards route
 * DEPRECATED: Redirects to /desktop?app=mcards
 */

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getAppById } from '@/lib/desktop/app-registry';

export default function McardsPage() {
  const router = useRouter();

  useEffect(() => {
    // Redirect to desktop and open Mcards app
    router.replace('/desktop');

    // Open app window after redirect completes
    setTimeout(() => {
      const { openWindow } = require('@/lib/desktop/store').useDesktopStore.getState();
      const app = getAppById('mcards');
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
        });
      }
    }, 100);
  }, [router]);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <p className="text-lg text-gray-600">Redirecting to desktop...</p>
      </div>
    </div>
  );
}

export const dynamic = 'force-static';
