'use client';
import { useState } from 'react';
import Link from 'next/link';

export default function SignupPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !email || !password || !confirm) return setError('Please fill in all fields.');
    if (password.length < 8) return setError('Password must be at least 8 characters.');
    if (password !== confirm) return setError('Passwords do not match.');
    setError('');
    console.log('Signing up:', { name, email });
    // TODO(Chidi): POST to the signup endpoint
  };

  const input =
    'mt-1 w-full rounded-md border border-[#1C222B] bg-[#080B0F] px-3 py-2 text-[12px] text-white placeholder:text-gray-600 focus:border-blue-500 focus:outline-none';

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#0A0D12]">
      <form onSubmit={handleSubmit} className="w-[380px] space-y-4 rounded-lg border border-[#161B22] bg-[#0D1117] p-6">
        <div>
          <h1 className="text-[15px] font-semibold text-white">Create your account</h1>
          <p className="mt-1 text-[11px] text-gray-500">Start building your skills baseline</p>
        </div>
        {error && <div role="alert" className="rounded-md border border-red-800 bg-red-950/50 p-2 text-[11px] text-red-200">{error}</div>}
        <label className="block text-[11px] text-gray-300">Full name
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Jamie Lee" className={input} />
        </label>
        <label className="block text-[11px] text-gray-300">Email address
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="name@company.com" className={input} />
        </label>
        <label className="block text-[11px] text-gray-300">Password
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" className={input} />
        </label>
        <label className="block text-[11px] text-gray-300">Confirm password
          <input type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} placeholder="••••••••" className={input} />
        </label>
        <button type="submit" className="w-full rounded-md bg-blue-500 py-2 text-[12px] font-medium text-white hover:bg-blue-400">Create account</button>
        <p className="text-center text-[11px] text-gray-500">
          Already have an account? <Link href="/login" className="text-blue-400 hover:underline">Sign in</Link>
        </p>
      </form>
    </div>
  );
}