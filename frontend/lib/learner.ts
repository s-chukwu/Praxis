import { supabase } from './supabase';
import { FriendlyError, report } from './errors';
import type { ProgramId } from './catalog';

export type SkillStatus = 'suggested' | 'confirmed' | 'rejected' | 'verified';

export type Learner = {
  name: string;
  program?: ProgramId;
  onboarded: boolean;
  cvSubmitted: boolean;
  skills: Record<string, SkillStatus>;
};

const API = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';
// Our Go backend has no learner-profile routes yet. Until it does, profile state lives in this browser.
export const DEMO_MODE = process.env.NEXT_PUBLIC_API_MODE !== 'live';

// Dev-only: a sample learner so the screens can be reviewed while sign-up is unavailable.
// NODE_ENV is fixed at build time, so a production build drops this branch entirely.
export const PREVIEW = process.env.NODE_ENV !== 'production' && process.env.NEXT_PUBLIC_PREVIEW === '1';
const PREVIEW_USER = { id: 'preview', user_metadata: { full_name: 'Amari Uwase' } };

async function userId() {
  if (PREVIEW) return PREVIEW_USER;
  const { data } = (await supabase?.auth.getSession()) ?? { data: { session: null } };
  return data.session?.user;
}

export type Role = 'candidate' | 'reviewer' | 'admin' | 'employer';

// The role lives in the profiles table, so only the backend can say who is a reviewer.
export async function getRole(): Promise<Role | null> {
  if (PREVIEW) return 'candidate';
  const user = await userId();
  if (!user) return null;
  try {
    return (await request<{ role: Role }>('GET', '/v1/me')).role;
  } catch (e) {
    report(e);
    return 'candidate';
  }
}

export function clearPreview() {
  if (PREVIEW) localStorage.removeItem(key(PREVIEW_USER.id));
}

function key(id: string) {
  return `praxis.learner.${id}`;
}

export async function getLearner(): Promise<Learner | null> {
  const user = await userId();
  if (!user) return null;
  const blank: Learner = {
    name: (user.user_metadata?.full_name as string) ?? '',
    onboarded: false,
    cvSubmitted: false,
    skills: {},
  };
  if (!DEMO_MODE) return { ...blank, ...(await request<Partial<Learner>>('GET', '/v1/me/learner')) };
  try {
    return { ...blank, ...JSON.parse(localStorage.getItem(key(user.id)) ?? '{}') };
  } catch {
    return blank;
  }
}

export async function saveLearner(patch: Partial<Learner>): Promise<void> {
  const user = await userId();
  if (!user) throw new FriendlyError('Your session ended. Sign in again to save your progress.');
  if (!DEMO_MODE) return void (await request('PATCH', '/v1/me/learner', patch));
  const current = (await getLearner()) ?? {};
  try {
    localStorage.setItem(key(user.id), JSON.stringify({ ...current, ...patch }));
  } catch {
    throw new FriendlyError('Your browser blocked saving. Turn off private browsing and try again.');
  }
}

export async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const { data } = (await supabase?.auth.getSession()) ?? { data: { session: null } };
  const res = await fetch(`${API}${path}`, {
    method,
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${data.session?.access_token ?? ''}` },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (res.status === 401) throw new FriendlyError('Your session ended. Sign in again.');
  if (!res.ok) throw new Error(`${method} ${path} failed with ${res.status}`);
  return (await res.json()) as T;
}
