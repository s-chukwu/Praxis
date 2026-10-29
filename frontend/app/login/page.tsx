'use client';

import { useState } from 'react';
import Link from 'next/link';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(''); // Added error state

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Basic validation check
    if (!email || !password) {
      setError('Please fill in all fields.');
      return;
    }
    
    setError('');
    console.log('Logging in with:', { email, password });
    // TODO: Hook up to backend auth endpoint
  };

  return (
    <div className="max-w-md mx-auto mt-12 bg-slate-900 border border-slate-800 rounded-xl p-8 shadow-xl">
      <h1 className="text-2xl font-bold text-white mb-2">Welcome Back</h1>
      <p className="text-sm text-slate-400 mb-6">Log in to continue your workforce transition journey.</p>

      {/* Error message banner */}
      {error && (
        <div className="mb-4 p-3 bg-red-950/50 border border-red-800 rounded-lg text-red-200 text-sm" role="alert">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-slate-300 mb-1">Email Address</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
            placeholder="you@example.com"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-slate-300 mb-1">Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
            placeholder="••••••••"
          />
        </div>

        <button
          type="submit"
          className="w-full py-2 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-lg text-sm transition-colors"
        >
          Sign In
        </button>
      </form>

      <p className="text-xs text-slate-400 mt-6 text-center">
        Don&apos;t have an account?{' '}
        <Link href="/signup" className="text-indigo-400 hover:underline">
          Sign up
        </Link>
      </p>
    </div>
  );
}