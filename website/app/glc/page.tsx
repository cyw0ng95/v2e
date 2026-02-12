'use client';

/**
 * GLC Landing Page - Legacy Route Redirect
 * /glc route - redirects to SPA root with app parameter
 */

import { useEffect } from 'react';

export default function GLCLandingPage() {
  useEffect(() => {
    // Immediate client-side redirect to SPA root
    window.location.href = '/?app=glc';
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
