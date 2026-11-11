'use client';
import { useState } from 'react';
import Link from 'next/link';
import Logo from '@/components/Logo';
import Input from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import AuthShowcase from './AuthShowcase';

export default function AuthScreen({ mode }: { mode: 'signin' | 'signup' }) {
  const signup = mode === 'signup';
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [message, setMessage] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (signup && !name) return setMessage('Enter your full name.');
    if (!email || !password) return setMessage('Enter your email and password.');
    if (signup && password.length < 8) return setMessage('Use at least 8 characters for your password.');
    // Credentials are never logged. Call the backend here once the Go routes exist.
    setMessage(signup ? 'Sign up is not connected to the backend yet.' : 'Sign in is not connected to the backend yet.');
  };

  const tab = (active: boolean) =>
    `flex-1 rounded-md py-2 text-center text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-400 ${
      active ? 'bg-[#1C222B] text-white' : 'text-gray-400 hover:text-white'
    }`;

  return (
    <div className="flex min-h-screen bg-[#0A0D12]">
      <div className="flex w-full flex-col items-center justify-center gap-8 px-4 py-10 lg:w-1/2">
        <Logo />
        <div className="w-full max-w-[400px] rounded-lg border border-[#161B22] bg-[#0D1117] p-7">
          <h1 className="text-xl font-semibold text-white">{signup ? 'Create your account' : 'Welcome back'}</h1>
          <p className="mt-1 text-sm text-gray-400">
            {signup ? 'Create an account to take assessments.' : 'Sign in to continue.'}
          </p>

          <nav aria-label="Account" className="my-5 flex rounded-lg border border-[#1C222B] bg-[#080B0F] p-1">
            <Link href="/login" aria-current={signup ? undefined : 'page'} className={tab(!signup)}>
              Sign in
            </Link>
            <Link href="/signup" aria-current={signup ? 'page' : undefined} className={tab(signup)}>
              Create account
            </Link>
          </nav>

          <form onSubmit={handleSubmit} className="space-y-4" noValidate>
            {message && (
              <div role="alert" className="rounded-md border border-[#1C222B] bg-[#080B0F] p-2 text-sm text-gray-300">
                {message}
              </div>
            )}
            {signup && (
              <Input id="name" label="Full name" autoComplete="name" value={name} onChange={(e) => setName(e.target.value)} />
            )}
            <Input id="email" label="Email address" type="email" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} />
            <Input
              id="password"
              label="Password"
              type="password"
              autoComplete={signup ? 'new-password' : 'current-password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <Button type="submit" className="w-full">
              {signup ? 'Create account' : 'Sign in'}
            </Button>
          </form>
        </div>
      </div>
      <AuthShowcase />
    </div>
  );
}