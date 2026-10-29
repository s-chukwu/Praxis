import Link from 'next/link';

export default function Sidebar() {
  return (
    <aside className="w-64 bg-slate-900 border-r border-slate-800 flex flex-col justify-between p-4 min-h-screen text-slate-100">
      <div className="space-y-6">
        {/* Logo / Header */}
        <div className="flex items-center space-x-2 px-2">
          <span className="font-bold text-xl tracking-tight text-indigo-400">Praxis</span>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-indigo-950 text-indigo-300 border border-indigo-800">
            v0.1.0
          </span>
        </div>

        {/* Role Selector Toggle */}
        <div className="bg-slate-950 p-1 rounded-lg border border-slate-800 flex text-xs font-medium">
          <button className="flex-1 py-1.5 rounded-md bg-indigo-600 text-white font-semibold text-center transition-colors">
            Learner
          </button>
          <button className="flex-1 py-1.5 rounded-md text-slate-400 hover:text-slate-200 text-center transition-colors">
            Employer
          </button>
        </div>

        {/* Navigation Links */}
        <nav className="space-y-1">
          <Link
            href="/overview"
            className="flex items-center space-x-3 px-3 py-2.5 rounded-lg text-sm font-medium text-slate-300 hover:bg-slate-800 hover:text-white transition-colors"
          >
            <svg
              className="w-5 h-5 text-slate-400"
              width="20"
              height="20"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 00-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
              />
            </svg>
            <span>Overview</span>
          </Link>

          <Link
            href="/skill-catalog"
            className="flex items-center space-x-3 px-3 py-2.5 rounded-lg text-sm font-medium text-slate-300 hover:bg-slate-800 hover:text-white transition-colors"
          >
            <svg
              className="w-5 h-5 text-slate-400"
              width="20"
              height="20"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
              />
            </svg>
            <span>Skill Catalog</span>
          </Link>

          <Link
            href="/assessments"
            className="flex items-center space-x-3 px-3 py-2.5 rounded-lg text-sm font-medium text-slate-300 hover:bg-slate-800 hover:text-white transition-colors"
          >
            <svg
              className="w-5 h-5 text-slate-400"
              width="20"
              height="20"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
              />
            </svg>
            <span>Assessments</span>
          </Link>
        </nav>
      </div>

      {/* Footer / Account links */}
      <div className="pt-4 border-t border-slate-800 space-y-1">
        <Link
          href="/login"
          className="block px-3 py-2 text-sm text-slate-400 hover:text-slate-200 transition-colors"
        >
          Login
        </Link>
        <Link
          href="/signup"
          className="block px-3 py-2 text-sm text-slate-400 hover:text-slate-200 transition-colors"
        >
          Sign Up
        </Link>
      </div>
    </aside>
  );
}