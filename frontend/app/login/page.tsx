'use client';
import { useState } from 'react';
import Link from 'next/link';
import AuthCard from '@/components/ui/AuthCard';
import Input from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [message, setMessage] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) {
      setMessage('Enter your email and password.');
      return;
    }
   
    setMessage('Sign in is not connected to the backend yet.');
  };

  return (
    <AuthCard title="Sign in" description="Sign in to see your assessments.">
      <form onSubmit={handleSubmit} className="space-y-4">
        {message && (
          <div role="alert" className="rounded-md border border-[#1C222B] bg-[#080B0F] p-2 text-sm text-gray-300">
            {message}
          </div>
        )}
        <Input id="email" label="Email address" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <Input id="password" label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        <Button type="submit" className="w-full">
          Sign in
        </Button>
        <p className="text-center text-sm text-gray-400">
          Don&apos;t have an account?{' '}
          <Link href="/signup" className="text-blue-400 hover:underline">
            Create one
          </Link>
        </p>
      </form>
    </AuthCard>
  );
}