'use client';

/**
 * ETL Engine Page - Legacy Route Redirect
 * /etl route - redirects to SPA root with app parameter
 */

import { useEffect } from 'react';

export default function ETLEnginePage() {
  useEffect(() => {
    // Immediate client-side redirect to SPA root
    window.location.href = '/?app=etl';
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
