import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* SPA Configuration with Static Export */
  output: 'export',

  /* Disable image optimization for static export to reduce build size */
  images: {
    unoptimized: true,
  },

  /* Use relative paths for assets */
  basePath: '',

  /* Trailing slash for static hosting */
  trailingSlash: true,

  /* Optimize production builds - disable source maps to reduce bundle size */
  productionBrowserSourceMaps: false,

  /* Compress output */
  compress: true,

  /* Security headers and CORS configuration */
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          // Content Security Policy to prevent XSS attacks
          {
            key: 'Content-Security-Policy',
            value: [
              "default-src 'self'",
              "script-src 'self' 'unsafe-inline' 'unsafe-eval'", // unsafe-inline/eval needed for Next.js development
              "style-src 'self' 'unsafe-inline'", // unsafe-inline needed for styled-components/inline styles
              "img-src 'self' data: https:",
              "font-src 'self' data:",
              "connect-src 'self' http://localhost:* https: http://localhost:* https://*.monkeycode-ai.online",
              "frame-ancestors 'none'",
              "base-uri 'self'",
              "form-action 'self'",
              "object-src 'none'", // Prevent plugin execution
            ].join('; '),
          },
          // Prevent clickjacking
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          // Prevent MIME type sniffing
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          // Enable XSS filtering
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block',
          },
          // Referrer policy
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
          // CORS headers for API access
          {
            key: 'Access-Control-Allow-Origin',
            value: '*',
          },
          {
            key: 'Access-Control-Allow-Methods',
            value: 'GET, POST, PUT, DELETE, OPTIONS',
          },
          {
            key: 'Access-Control-Allow-Headers',
            value: 'Content-Type, Authorization',
          },
        ],
      },
    ];
  },
};

export default nextConfig;
