/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: [],
  output: 'standalone',  // Enable standalone output for Docker deployment
}

module.exports = nextConfig