'use client';

/**
 * Desktop Page - Legacy Route Redirect
 * /desktop route - redirects to SPA root
 */

import { useEffect } from 'react';

export default function DesktopPage() {
  useEffect(() => {
    // Immediate client-side redirect to SPA root
    window.location.href = '/';
  }, []);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <p className="text-lg text-gray-600">Redirecting to home...</p>
      </div>
    </div>
  );
}

export const dynamic = 'force-static';
