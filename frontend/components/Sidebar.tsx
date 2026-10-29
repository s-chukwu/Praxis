'use client';

import Link from 'next/link';

export default function Sidebar() {
  return (
    <aside className="w-64 bg-slate-900 text-slate-300 min-h-[calc(100vh-73px)] border-r border-slate-800 p-4 flex flex-col justify-between">
      <nav className="space-y-1">
        <Link
          href="/"
          className="flex items-center space-x-3 px-3 py-2 rounded-md bg-slate-800 text-white font-medium"
        >
          <span>🏠</span>
          <span>Overview</span>
        </Link>
        <Link
          href="/catalog"
          className="flex items-center space-x-3 px-3 py-2 rounded-md hover:bg-slate-800/60 hover:text-white transition-colors"
        >
          <span>🎯</span>
          <span>Skill Catalog</span>
        </Link>
        <Link
          href="/assessments"
          className="flex items-center space-x-3 px-3 py-2 rounded-md hover:bg-slate-800/60 hover:text-white transition-colors"
        >
          <span>📊</span>
          <span>Assessments</span>
        </Link>
      </nav>

      <div className="p-3 bg-slate-800/40 rounded-lg border border-slate-800 text-xs text-slate-400">
        <p className="font-semibold text-slate-300 mb-1">Praxis Stack</p>
        <p>Go • FastAPI • Next.js</p>
      </div>
    </aside>
  );
}