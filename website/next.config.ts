import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* Static Site Generation (SSG) configuration */
  output: 'export',
  
  /* Disable image optimization for static export */
  images: {
    unoptimized: true,
  },
  
  /* Use relative paths for assets */
  basePath: '',
  
  /* Trailing slash for static hosting */
  trailingSlash: true,
};

export default nextConfig;
