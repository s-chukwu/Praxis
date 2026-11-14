'use client';
import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import Logo from '@/components/Logo';
import Input from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import { supabase } from '@/lib/supabase';
import { report } from '@/lib/errors';
import AuthShowcase from './AuthShowcase';

type Errors = Partial<Record<'name' | 'email' | 'password' | 'form', string>>;

// Supabase's messages are written for developers; learners get what happened and what to do next.
function authMessage(raw: string) {
  if (/email not confirmed/i.test(raw)) return 'Confirm your email first. Open the link we sent you, then sign in.';
  if (/invalid login/i.test(raw)) return 'That email and password do not match.';
  if (/already registered|already exists/i.test(raw)) return 'An account with this email already exists. Sign in instead.';
  if (/rate limit|too many/i.test(raw)) return 'Too many attempts. Wait a minute, then try again.';
  if (/password/i.test(raw)) return 'Choose a stronger password: at least 8 characters, including a letter and a number.';
  return 'We could not create or open your account just now. Try again in a minute.';
}

export default function AuthScreen({ mode }: { mode: 'signin' | 'signup' }) {
  const signup = mode === 'signup';
  const router = useRouter();
  const { refresh } = useLearner();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<Errors>({});
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);
  const [unconfirmed, setUnconfirmed] = useState(false);
  const [resent, setResent] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const found: Errors = {};
    if (signup && !name.trim()) found.name = 'Enter your full name.';
    if (!/^\S+@\S+\.\S+$/.test(email)) found.email = 'Enter a valid email address.';
    if (signup) {
      const hasLetter = /[A-Za-z]/.test(password);
      const hasNumber = /\d/.test(password);
      if (password.length < 8 || !hasLetter || !hasNumber) {
        found.password = 'Use at least 8 characters, including a letter and a number.';
      }
    } else if (!password) {
      found.password = 'Enter your password.';
    }
    setErrors(found);
    if (Object.keys(found).length) return;
    if (!supabase) return setErrors({ form: 'Sign-in is not configured on this server yet.' });

    setBusy(true);
    const { data, error } = signup
      ? await supabase.auth.signUp({ email, password, options: { data: { full_name: name.trim() } } })
      : await supabase.auth.signInWithPassword({ email, password });
    setBusy(false);

    if (error) {
      report(error);
      setUnconfirmed(/email not confirmed/i.test(error.message));
      return setErrors({ form: authMessage(error.message) });
    }
    if (signup && !data.session) return setNotice(`We sent a confirmation link to ${email}. Open it, then sign in.`);
    await refresh();
    router.replace(signup ? '/onboarding' : '/dashboard');
  };

  const resend = async () => {
    const { error } = (await supabase?.auth.resend({ type: 'signup', email })) ?? {};
    if (error) {
      report(error);
      return setErrors({ form: authMessage(error.message) });
    }
    setResent(true);
  };

  return (
    <div className="flex min-h-screen bg-surface">
      <div className="flex w-full flex-col px-6 py-8 sm:px-10 lg:w-[46%]">
        <Logo />
        <div className="mx-auto flex w-full max-w-[400px] flex-1 flex-col justify-center py-12">
          <h1 className="text-[32px] font-bold leading-tight">{signup ? 'Start your training' : 'Welcome back'}</h1>
          <p className="mt-2 text-muted">
            {signup ? 'Create your account. Setup takes about five minutes.' : 'Sign in to pick up where you left off.'}
          </p>

          {notice ? (
            <div role="status" className="mt-8 rounded-panel bg-brand-soft p-5 text-brand-ink">
              <p className="font-semibold">Check your email</p>
              <p className="mt-1 text-sm">{notice}</p>
            </div>
          ) : (
            <form onSubmit={submit} className="mt-8 space-y-5" noValidate>
              {errors.form && (
                <div role="alert" className="rounded-control bg-danger-soft px-4 py-3 text-sm font-medium text-danger">
                  {errors.form}
                </div>
              )}
              {unconfirmed && (
                <button type="button" onClick={resend} disabled={resent} className="text-sm font-semibold text-brand-ink underline disabled:no-underline disabled:opacity-70">
                  {resent ? `Confirmation email sent to ${email}` : 'Resend confirmation email'}
                </button>
              )}
              {signup && (
                <Input id="name" label="Full name" autoComplete="name" value={name} error={errors.name} onChange={(e) => setName(e.target.value)} />
              )}
              <Input id="email" label="Email" type="email" autoComplete="email" value={email} error={errors.email} onChange={(e) => setEmail(e.target.value)} />
              <Input
                id="password"
                label="Password"
                type="password"
                autoComplete={signup ? 'new-password' : 'current-password'}
                hint={signup ? 'At least 8 characters, including a letter and a number.' : undefined}
                value={password}
                error={errors.password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <Button type="submit" size="lg" loading={busy} className="w-full">
                {signup ? 'Create account' : 'Sign in'}
              </Button>
            </form>
          )}

          <p className="mt-8 text-sm text-muted">
            {signup ? 'Already have an account? ' : 'New to Praxis? '}
            <Link href={signup ? '/login' : '/signup'} className="font-semibold text-brand-ink underline">
              {signup ? 'Sign in' : 'Create an account'}
            </Link>
          </p>
        </div>
      </div>
      <AuthShowcase />
    </div>
  );
}
