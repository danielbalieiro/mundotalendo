/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  experimental: {
    serverActions: {
      bodySizeLimit: '2mb',
    },
  },
  // Allow dev server access from local network
  allowedDevOrigins: [
    '192.168.1.35', // Your local network IP
    'localhost',
    '127.0.0.1',
  ],
  // Security headers
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
        ],
      },
    ]
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
