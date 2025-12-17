/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  experimental: {
    serverActions: {
      bodySizeLimit: '2mb',
    },
  },
  // Turbopack configuration (Next.js 16+)
  turbopack: {
    resolveAlias: {
      'maplibre-gl': 'maplibre-gl/dist/maplibre-gl.js',
    },
  },
  // Webpack configuration (fallback for --webpack mode)
  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      'maplibre-gl': 'maplibre-gl/dist/maplibre-gl.js',
    };
    return config;
  },
};

module.exports = nextConfig;
