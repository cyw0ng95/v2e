/**
 * Network status detection hook
 * Monitors online/offline status and updates desktop store
 */

'use client';

import { useEffect } from 'react';
import { useDesktopStore } from '@/lib/desktop/store';

/**
 * Hook to detect and sync network status with desktop store
 * Listens to browser online/offline events
 */
export function useNetworkStatus() {
  const { isOnline, setOnlineStatus } = useDesktopStore();

  useEffect(() => {
    // Update initial status
    setOnlineStatus(navigator.onLine);

    // Handle online event
    const handleOnline = () => {
      setOnlineStatus(true);
    };

    // Handle offline event
    const handleOffline = () => {
      setOnlineStatus(false);
    };

    // Add event listeners
    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Cleanup listeners on unmount
    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [setOnlineStatus]);

  return { isOnline };
}
