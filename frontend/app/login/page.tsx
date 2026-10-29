'use client';

import { useState } from 'react';
import Link from 'next/link';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log('Logging in with:', { email, password });
  };

  return (
    <div className="max-w-md mx-auto mt-12 bg-slate-900 border border-slate-800 rounded-xl p-8 shadow-xl">
      <h1 className="text-2xl font-bold text-white mb-2">Welcome Back</h1>
      <p className="text-sm text-slate-400 mb-6">Log in to continue your workforce transition journey.</p>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-slate-300 mb-1">Email Address</label>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
            placeholder="you@example.com"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-slate-300 mb-1">Password</label>
          <input
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
            placeholder="••••••••"
          />
        </div>

        <button
          type="submit"
          className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2 rounded-lg transition-colors text-sm"
        >
          Sign In
        </button>
      </form>

      <p className="text-xs text-slate-400 text-center mt-6">
        Don't have an account?{' '}
        <Link href="/signup" className="text-blue-400 hover:underline">
          Sign up
        </Link>
      </p>
    </div>
  );
}