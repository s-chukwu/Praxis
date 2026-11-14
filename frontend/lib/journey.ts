import type { Progress } from './baseline';

export type MilestoneState = 'earned' | 'current' | 'locked';
export type Milestone = { id: string; title: string; detail: string; kind: 'step' | 'badge' | 'certificate'; state: MilestoneState };

const SUBMITTED = new Set(['submitted', 'under_review', 'review_pending', 'reviewed']);
export const REVIEW_HOURS = 48;

// The PM's X-02 decision: badge 1 is a reviewed baseline, badge 2 a finished plan, then the certificate.
export function milestones(progress: Progress | null): Milestone[] {
  const status = progress?.baseline?.status;
  const submitted = Boolean(status && SUBMITTED.has(status));
  const reviewed = Boolean(progress?.result);
  const items: Omit<Milestone, 'state'>[] = [
    { id: 'baseline', title: 'Baseline task', detail: 'Answer one real support case', kind: 'step' },
    { id: 'reviewed', title: 'Reviewed', detail: 'A supervisor scores your work', kind: 'badge' },
    { id: 'plan', title: 'Plan complete', detail: 'Finish the courses chosen for you', kind: 'badge' },
    { id: 'final', title: 'Final task', detail: 'A new case shows what improved', kind: 'step' },
    { id: 'certificate', title: 'Certificate', detail: 'Eligible for employer review', kind: 'certificate' },
  ];
  const done = [submitted, reviewed, false, false, false];
  const current = done.indexOf(false);
  return items.map((m, i) => ({ ...m, state: done[i] ? 'earned' : i === current ? 'current' : 'locked' }));
}

// Remembers which milestones this browser has already celebrated, so each one animates exactly once.
const SEEN_KEY = 'praxis.milestones.seen';

export function takeNewlyEarned(list: Milestone[]): Set<string> {
  let seen: string[] = [];
  try {
    seen = JSON.parse(localStorage.getItem(SEEN_KEY) ?? '[]');
  } catch {
    return new Set();
  }
  const earned = list.filter((m) => m.state === 'earned').map((m) => m.id);
  const fresh = earned.filter((id) => !seen.includes(id));
  try {
    localStorage.setItem(SEEN_KEY, JSON.stringify(Array.from(new Set([...seen, ...earned]))));
  } catch {
    // Private browsing: the milestone simply shows without its first-time animation.
  }
  return new Set(fresh);
}

export function hoursLeft(submittedAt: string | null, now = Date.now()): number | null {
  if (!submittedAt) return null;
  return Math.max(0, Math.ceil((new Date(submittedAt).getTime() + REVIEW_HOURS * 3_600_000 - now) / 3_600_000));
}

export function ago(iso: string | null, now = Date.now()): string {
  if (!iso) return '';
  const mins = Math.round((now - new Date(iso).getTime()) / 60_000);
  if (mins < 2) return 'just now';
  if (mins < 60) return `${mins} minutes ago`;
  const hours = Math.round(mins / 60);
  if (hours < 24) return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
  const days = Math.round(hours / 24);
  return days === 1 ? 'yesterday' : `${days} days ago`;
}
