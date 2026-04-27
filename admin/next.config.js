/** @type {import('next').NextConfig} */
const nextConfig = {
  /**
   * 开发下访问 dev 资源时校验 Origin。`http://127.0.0.1:端口` 与 `http://localhost:端口` 在浏览器中非同源，同时列出更稳妥（团队可任选其一在地址栏打开）。
   * @see https://nextjs.org/docs/app/api-reference/config/next-config-js/allowedDevOrigins
   */
  allowedDevOrigins: ['127.0.0.1', 'localhost'],
  async rewrites() {
    return [
      { source: '/api/:path*', destination: 'http://localhost:8080/api/:path*' },
      { source: '/uploads/:path*', destination: 'http://localhost:8080/uploads/:path*' },
    ]
  },
}

module.exports = nextConfig
