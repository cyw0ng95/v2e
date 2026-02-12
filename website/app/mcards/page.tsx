'use client';

/**
 * Mcards Page - Legacy Route Redirect
 * /mcards route - redirects to SPA root with app parameter
 */

import { useEffect } from 'react';

export default function McardsPage() {
  useEffect(() => {
    // Immediate client-side redirect to SPA root
    window.location.href = '/?app=mcards';
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
