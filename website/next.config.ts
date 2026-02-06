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
};

export default nextConfig;
