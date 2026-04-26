import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import dynamic from 'next/dynamic';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

const ClientLayout = dynamic(() => import('@/components/layout/AppLayout'), {
  ssr: false,
  loading: () => (
    <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div style={{ width: 40, height: 40, borderRadius: '50%', background: '#06B6D4' }} />
    </div>
  ),
});

export const metadata: Metadata = {
  title: 'atom - 小原子，大能量',
  description: 'atom 企业级 AI 智能助手平台',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body className={inter.className}>
        <ClientLayout>{children}</ClientLayout>
      </body>
    </html>
  );
}
