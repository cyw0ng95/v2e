import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === 'development';

const nextConfig: NextConfig = {
  output: isDev ? undefined : 'export',

  images: {
    unoptimized: !isDev,
  },

  basePath: '',

  trailingSlash: !isDev,

  allowedDevOrigins: ['e1410ca11ca5.monkeycode-ai.online', 'localhost'],

  async rewrites() {
    if (!isDev) return [];
    
    const v2accessUrl = process.env.V2ACCESS_URL || 'http://localhost:8080';
    return [
      {
        source: '/restful/:path*',
        destination: `${v2accessUrl}/restful/:path*`,
      },
    ];
  },

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
