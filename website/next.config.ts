import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* Static Site Generation (SSG) configuration */
  output: 'export',
  
  /* Disable image optimization for static export to reduce build size */
  images: {
    unoptimized: false,
  },
  
  /* Use relative paths for assets */
  basePath: '',
  
  /* Trailing slash for static hosting */
  trailingSlash: true,
  
  /* Optimize production builds - disable source maps to reduce bundle size */
  productionBrowserSourceMaps: false,
  
  /* Compress output */
  compress: true,
  
  /* Allow remote development access */
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
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
