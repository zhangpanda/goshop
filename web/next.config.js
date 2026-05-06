/** @type {import('next').NextConfig} */
const isProd = process.env.NODE_ENV === 'production'
const nextConfig = {
  allowedDevOrigins: ['127.0.0.1', 'localhost'],
  output: isProd ? 'standalone' : undefined,
  basePath: isProd ? '/goshop' : '',
  async rewrites() {
    const apiDest = isProd ? 'http://127.0.0.1:8081' : 'http://localhost:8080'
    return [
      { source: '/api/:path*', destination: `${apiDest}/api/:path*` },
      { source: '/uploads/:path*', destination: `${apiDest}/uploads/:path*` },
    ]
  },
}

module.exports = nextConfig
