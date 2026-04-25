/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: [],
  output: 'standalone',  // Enable standalone output for Docker deployment
  experimental: {
    serverActions: true,
  },
}

module.exports = nextConfig