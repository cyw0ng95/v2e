import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* Development mode configuration */
  output: undefined, // Not static export in dev mode

  /* Disable image optimization */
  images: {
    unoptimized: false,
  },

  /* Use relative paths for assets */
  basePath: '',

  /* Configure RPC proxy for development mode */
  async rewrites() {
    const v2accessUrl = process.env.V2ACCESS_URL || 'http://localhost:8080';
    return [
      {
        source: '/restful/:path*',
        destination: `${v2accessUrl}/restful/:path*`,
      },
    ];
  },

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
