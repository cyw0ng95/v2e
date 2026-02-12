'use client';

/**
 * CVSS Calculator - Legacy Route Redirect
 * /cvss route - redirects to SPA root with app parameter
 */

import { useEffect } from 'react';

export default function CVSSPage() {
  useEffect(() => {
    // Immediate client-side redirect to SPA root
    window.location.href = '/?app=cvss';
  }, []);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <p className="text-lg text-gray-600">Redirecting to desktop...</p>
      </div>
    </div>
  );
}

export const dynamic = 'force-static';
