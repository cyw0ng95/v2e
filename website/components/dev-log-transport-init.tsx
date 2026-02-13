'use client';

import { useEffect } from 'react';
import { initDevLogTransport, destroyDevLogTransport } from '@/lib/dev-log-transport';

export function DevLogTransportInit() {
  useEffect(() => {
    initDevLogTransport();

    return () => {
      destroyDevLogTransport();
    };
  }, []);

  return null;
}
