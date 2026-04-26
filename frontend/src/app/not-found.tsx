import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="w-20 h-20 mx-auto rounded-full bg-atom-mist flex items-center justify-center mb-4">
          <span className="text-3xl font-bold text-atom-core">404</span>
        </div>
        <h1 className="text-xl font-bold text-foreground mb-2">页面不存在</h1>
        <p className="text-sm text-muted-foreground mb-6">你访问的页面可能已被移除或不存在</p>
        <Link
          href="/"
          className="inline-flex px-4 py-2 bg-atom-core text-white rounded-lg text-sm hover:bg-atom-nucleus transition-colors"
        >
          返回首页
        </Link>
      </div>
    </div>
  );
}
