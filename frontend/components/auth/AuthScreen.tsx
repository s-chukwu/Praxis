'use client';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import Logo from '@/components/Logo';
import Input from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import { supabase } from '@/lib/supabase';
import { report } from '@/lib/errors';
import AuthShowcase from './AuthShowcase';

type Errors = Partial<Record<'name' | 'email' | 'password' | 'confirm' | 'form', string>>;
type View = 'signin' | 'forgot' | 'recovery';

// Supabase's messages are written for developers; learners get what happened and what to do next.
function authMessage(raw: string) {
  if (/email not confirmed/i.test(raw)) return 'Confirm your email first. Open the link we sent you, then sign in.';
  if (/invalid login/i.test(raw)) return 'That email and password do not match.';
  if (/already registered|already exists/i.test(raw)) return 'An account with this email already exists. Sign in instead.';
  if (/rate limit|too many/i.test(raw)) return 'Too many attempts. Wait a minute, then try again.';
  if (/password/i.test(raw)) return 'Choose a stronger password: at least 8 characters, including a letter and a number.';
  return 'We could not create or open your account just now. Try again in a minute.';
}

function resetMessage(raw: string) {
  if (/rate limit|too many|security purposes/i.test(raw)) return 'Too many reset requests. Wait a minute, then try again.';
  if (/same password|different from the old/i.test(raw)) return 'Choose a password you have not used before.';
  if (/expired|invalid|token/i.test(raw)) return 'This reset link is invalid or has expired. Request a new one.';
  if (/password/i.test(raw)) return 'Choose a stronger password: at least 8 characters, including a letter and a number.';
  return 'We could not update your password just now. Try again in a minute.';
}

function Notice({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div role="status" className="mt-8 rounded-panel bg-brand-soft p-5 text-brand-ink">
      <p className="font-semibold">{title}</p>
      <div className="mt-1 text-sm">{children}</div>
    </div>
  );
}

function FormError({ children }: { children: React.ReactNode }) {
  return (
    <div role="alert" className="rounded-control bg-danger-soft px-4 py-3 text-sm font-medium text-danger">
      {children}
    </div>
  );
}

export default function AuthScreen({ mode }: { mode: 'signin' | 'signup' }) {
  const signup = mode === 'signup';
  const router = useRouter();
  const { refresh } = useLearner();
  const [view, setView] = useState<View>('signin');
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<Errors>({});
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);
  const [unconfirmed, setUnconfirmed] = useState(false);
  const [resent, setResent] = useState(false);
  const [resetEmail, setResetEmail] = useState('');
  const [resetSent, setResetSent] = useState(false);
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [recoveryError, setRecoveryError] = useState('');
  const [recoveryState, setRecoveryState] = useState<'checking' | 'ready' | 'invalid'>('checking');
  const [recoveryAttempt, setRecoveryAttempt] = useState(0);
  const [recoveryDone, setRecoveryDone] = useState(false);

  useEffect(() => {
    if (signup) return;

    const sync = () => {
      const params = new URLSearchParams(window.location.search);
      const hash = new URLSearchParams(window.location.hash.replace(/^#/, ''));
      const error = hash.get('error_description') || params.get('error_description');
      const type = hash.get('type') || params.get('type');
      const recovery =
        params.get('recovery') === '1' || type === 'recovery' || (params.has('code') && type !== 'signup') || Boolean(error);

      setErrors({});
      if (recovery) {
        setView('recovery');
        setRecoveryAttempt((n) => n + 1);
        if (error) {
          setRecoveryError('This reset link is invalid or has expired. Request a new one.');
          setRecoveryState('invalid');
        } else {
          setRecoveryError('');
          setRecoveryState('checking');
        }
        return;
      }
      setView(params.get('forgot') === '1' ? 'forgot' : 'signin');
    };

    sync();
    window.addEventListener('popstate', sync);
    return () => window.removeEventListener('popstate', sync);
  }, [signup]);

  useEffect(() => {
    if (view !== 'recovery' || recoveryError || !supabase) return;

    let active = true;
    const finish = (hasSession: boolean) => {
      if (!active) return;
      setRecoveryState((current) => (current === 'ready' ? current : hasSession ? 'ready' : 'invalid'));
    };

    void supabase.auth.getSession().then(({ data }) => finish(Boolean(data.session)));
    const { data: sub } = supabase.auth.onAuthStateChange((event, session) => {
      if (event === 'PASSWORD_RECOVERY' || event === 'SIGNED_IN' || event === 'TOKEN_REFRESHED') {
        finish(Boolean(session));
      }
    });
    const timer = window.setTimeout(() => finish(false), 2500);

    return () => {
      active = false;
      sub.subscription.unsubscribe();
      window.clearTimeout(timer);
    };
  }, [view, recoveryError, recoveryAttempt]);

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

  const resetSignInState = () => {
    setView('signin');
    setErrors({});
    setResetSent(false);
  };

  const goToSignIn = () => {
    resetSignInState();
    router.push('/login');
  };

  const goToForgot = (seedEmail: string) => {
    setResetEmail(seedEmail);
    setResetSent(false);
    setErrors({});
    setView('forgot');
    router.push('/login?forgot=1');
  };

  const sendReset = async (e: React.FormEvent) => {
    e.preventDefault();
    const found: Errors = {};
    if (!/^\S+@\S+\.\S+$/.test(resetEmail)) found.email = 'Enter a valid email address.';
    setErrors(found);
    if (Object.keys(found).length) return;
    if (!supabase) return setErrors({ form: 'Sign-in is not configured on this server yet.' });

    setBusy(true);
    const { error } = await supabase.auth.resetPasswordForEmail(resetEmail, {
      redirectTo: `${window.location.origin}/login?recovery=1`,
    });
    setBusy(false);

    if (error) {
      report(error);
      if (/rate limit|too many|security purposes/i.test(error.message)) {
        return setErrors({ form: 'Too many reset requests. Wait a minute, then try again.' });
      }
      if (/network|fetch/i.test(error.message)) {
        return setErrors({ form: 'We could not send the email. Check your connection and try again.' });
      }
    }
    setResetSent(true);
  };

  const updatePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    const found: Errors = {};
    const hasLetter = /[A-Za-z]/.test(newPassword);
    const hasNumber = /\d/.test(newPassword);
    if (newPassword.length < 8 || !hasLetter || !hasNumber) {
      found.password = 'Use at least 8 characters, including a letter and a number.';
    }
    if (confirmPassword !== newPassword) found.confirm = 'The passwords do not match.';
    setErrors(found);
    if (Object.keys(found).length) return;
    if (!supabase) return setErrors({ form: 'Sign-in is not configured on this server yet.' });

    setBusy(true);
    const { error } = await supabase.auth.updateUser({ password: newPassword });
    setBusy(false);

    if (error) {
      report(error);
      return setErrors({ form: resetMessage(error.message) });
    }
    setRecoveryDone(true);
    window.history.replaceState({}, '', '/login');
    await supabase.auth.signOut();
  };

  const title = signup
    ? 'Start your training'
    : view === 'forgot'
      ? 'Reset your password'
      : view === 'recovery'
        ? 'Set a new password'
        : 'Welcome back';

  const subtitle = signup
    ? 'Create your account. Setup takes about five minutes.'
    : view === 'forgot'
      ? 'Enter the email you use for Praxis. We will send you a link to set a new password.'
      : view === 'recovery'
        ? 'Choose a new password for your account.'
        : 'Sign in to pick up where you left off.';

  return (
    <div className="flex min-h-screen bg-surface">
      <div className="flex w-full flex-col px-6 py-8 sm:px-10 lg:w-[46%]">
        <Logo />
        <div className="mx-auto flex w-full max-w-[400px] flex-1 flex-col justify-center py-12">
          <h1 className="text-[32px] font-bold leading-tight">{title}</h1>
          <p className="mt-2 text-muted">{subtitle}</p>

          {view === 'forgot' ? (
            resetSent ? (
              <>
                <Notice title="Check your email">
                  If an account exists for {resetEmail}, we sent a link to reset your password.
                </Notice>
                <Button type="button" variant="secondary" size="lg" className="mt-6 w-full" onClick={goToSignIn}>
                  Back to sign in
                </Button>
              </>
            ) : (
              <form onSubmit={sendReset} className="mt-8 space-y-5" noValidate>
                {errors.form && <FormError>{errors.form}</FormError>}
                <Input
                  id="reset-email"
                  label="Email"
                  type="email"
                  autoComplete="email"
                  value={resetEmail}
                  error={errors.email}
                  onChange={(e) => setResetEmail(e.target.value)}
                />
                <Button type="submit" size="lg" loading={busy} className="w-full">
                  Send reset link
                </Button>
              </form>
            )
          ) : view === 'recovery' ? (
            recoveryDone ? (
              <>
                <Notice title="Password updated">Your password has been updated. Sign in with your new password.</Notice>
                <Button type="button" size="lg" className="mt-6 w-full" onClick={goToSignIn}>
                  Sign in
                </Button>
              </>
            ) : !supabase ? (
              <Notice title="Password reset is not available">Sign-in is not configured on this server yet.</Notice>
            ) : recoveryState === 'checking' ? (
              <div role="status" className="mt-8 rounded-panel border border-line bg-page p-5 text-sm text-muted">
                Checking your reset link...
              </div>
            ) : recoveryState === 'invalid' ? (
              <>
                <Notice title="Link expired or invalid">
                  {recoveryError || 'This reset link is invalid or has expired. Request a new one.'}
                </Notice>
                <Button type="button" size="lg" className="mt-6 w-full" onClick={() => goToForgot(resetEmail)}>
                  Request a new link
                </Button>
              </>
            ) : (
              <form onSubmit={updatePassword} className="mt-8 space-y-5" noValidate>
                {errors.form && <FormError>{errors.form}</FormError>}
                <Input
                  id="new-password"
                  label="New password"
                  type="password"
                  autoComplete="new-password"
                  hint="At least 8 characters, including a letter and a number."
                  value={newPassword}
                  error={errors.password}
                  onChange={(e) => setNewPassword(e.target.value)}
                />
                <Input
                  id="confirm-password"
                  label="Confirm new password"
                  type="password"
                  autoComplete="new-password"
                  value={confirmPassword}
                  error={errors.confirm}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                />
                <Button type="submit" size="lg" loading={busy} className="w-full">
                  Update password
                </Button>
              </form>
            )
          ) : notice ? (
            <Notice title="Check your email">{notice}</Notice>
          ) : (
            <form onSubmit={submit} className="mt-8 space-y-5" noValidate>
              {errors.form && <FormError>{errors.form}</FormError>}
              {unconfirmed && (
                <button
                  type="button"
                  onClick={resend}
                  disabled={resent}
                  className="text-sm font-semibold text-brand-ink underline disabled:no-underline disabled:opacity-70"
                >
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
              {!signup && (
                <div className="text-right">
                  <button type="button" onClick={() => goToForgot(email)} className="text-sm font-semibold text-brand-ink underline">
                    Forgot password?
                  </button>
                </div>
              )}
              <Button type="submit" size="lg" loading={busy} className="w-full">
                {signup ? 'Create account' : 'Sign in'}
              </Button>
            </form>
          )}

          {view !== 'recovery' && (
            <p className="mt-8 text-sm text-muted">
              {signup ? 'Already have an account? ' : view === 'forgot' ? 'Remembered your password? ' : 'New to Praxis? '}
              {signup || view === 'forgot' ? (
                <Link href="/login" onClick={signup ? undefined : resetSignInState} className="font-semibold text-brand-ink underline">
                  Sign in
                </Link>
              ) : (
                <Link href="/signup" className="font-semibold text-brand-ink underline">
                  Create an account
                </Link>
              )}
            </p>
          )}
        </div>
      </div>
      <AuthShowcase />
    </div>
  );
}
