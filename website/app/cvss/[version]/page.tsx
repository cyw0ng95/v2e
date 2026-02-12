'use client';

/**
 * CVSS Calculator Page - Dynamic Route
 * /cvss/[version] where version is 3.0, 3.1, or 4.0
 * DEPRECATED: Redirects to /desktop?app=cvss&version=X
 */

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getAppById } from '@/lib/desktop/app-registry';

interface PageProps {
  params: Promise<{ version: string }>;
}

export default async function CVSSVersionPage({ params }: PageProps) {
  const { version } = await params;
  const router = useRouter();

  useEffect(() => {
    // Redirect to desktop and open CVSS calculator with version
    router.replace('/desktop');

    // Open app window after redirect completes
    setTimeout(() => {
      const { openWindow } = require('@/lib/desktop/store').useDesktopStore.getState();
      const app = getAppById('cvss');
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
          // Pass version as extra data for the component to use
          initialVersion: version,
        });
      }
    }, 100);
  }, [router, version]);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <p className="text-lg text-gray-600">Redirecting to desktop...</p>
      </div>
    </div>
  );
}
