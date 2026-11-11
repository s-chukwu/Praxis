'use client';
import { useState } from 'react';
import Link from 'next/link';
import AuthCard from '@/components/ui/AuthCard';
import Input from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';

export default function SignupPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [message, setMessage] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !email || !password || !confirm) return setMessage('Fill in all fields.');
    if (password.length < 8) return setMessage('Use at least 8 characters for your password.');
    if (password !== confirm) return setMessage('Passwords do not match.');
    // TODO(Chidi): POST to the signup endpoint. Never log credentials.
    setMessage('Sign up is not connected to the backend yet.');
  };

  return (
    <AuthCard title="Create your account" description="Create an account to take assessments.">
      <form onSubmit={handleSubmit} className="space-y-4">
        {message && (
          <div role="alert" className="rounded-md border border-[#1C222B] bg-[#080B0F] p-2 text-sm text-gray-300">
            {message}
          </div>
        )}
        <Input id="name" label="Full name" value={name} onChange={(e) => setName(e.target.value)} />
        <Input id="email" label="Email address" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <Input id="password" label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        <Input id="confirm" label="Confirm password" type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} />
        <Button type="submit" className="w-full">
          Create account
        </Button>
        <p className="text-center text-sm text-gray-400">
          Already have an account?{' '}
          <Link href="/login" className="text-blue-400 hover:underline">
            Sign in
          </Link>
        </p>
      </form>
    </AuthCard>
  );
}