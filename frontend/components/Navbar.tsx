'use client';

import Link from 'next/link';
import { useState } from 'react';

export default function Navbar() {
  const [role, setRole] = useState<'learner' | 'employer'>('learner');

  return (
    <header className="w-full bg-slate-900 text-white border-b border-slate-800 px-6 py-4 flex items-center justify-between">
      <div className="flex items-center space-x-3">
        <span className="text-xl font-bold tracking-tight text-blue-400">Praxis</span>
        <span className="text-xs bg-slate-800 text-slate-400 px-2 py-1 rounded">v0.1.0</span>
      </div>

      <div className="flex items-center space-x-6">
        {/* Role Switcher */}
        <div className="flex bg-slate-800 p-1 rounded-lg border border-slate-700 text-xs">
          <button
            onClick={() => setRole('learner')}
            className={`px-3 py-1 rounded-md transition-all ${
              role === 'learner' ? 'bg-blue-600 text-white font-medium' : 'text-slate-400 hover:text-white'
            }`}
          >
            Learner
          </button>
          <button
            onClick={() => setRole('employer')}
            className={`px-3 py-1 rounded-md transition-all ${
              role === 'employer' ? 'bg-blue-600 text-white font-medium' : 'text-slate-400 hover:text-white'
            }`}
          >
            Employer
          </button>
        </div>

        {/* Auth Navigation Links */}
        <div className="flex items-center space-x-3 text-sm">
          <Link href="/login" className="text-slate-300 hover:text-white transition-colors">
            Login
          </Link>
          <Link
            href="/signup"
            className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-1.5 rounded-md font-medium transition-colors"
          >
            Sign Up
          </Link>
        </div>
      </div>
    </header>
  );
}