import { FriendlyError } from './errors';
import { PREVIEW, request } from './learner';
import type { ProgramId } from './catalog';

export type Scenario = {
  scenario_id: string;
  version: string;
  brief: string;
  timebox_minutes: number;
  fields: string[];
};

export type AttemptStatus = 'assigned' | 'in_progress' | 'submitted' | 'under_review' | 'review_pending' | 'reviewed' | 'voided';

export type Attempt = {
  id: string;
  stage: 'baseline' | 'final';
  status: AttemptStatus;
  answers?: Record<string, string>;
  submitted_at: string | null;
};

export type Answers = Record<string, string>;

// Preview has no signed-in user, so it keeps one attempt in this browser and uses the real CS-01-A brief.
const PREVIEW_KEY = 'praxis.baseline.preview';
const PREVIEW_SCENARIO: Scenario = {
  scenario_id: 'CS-01-A',
  version: '1.2.0',
  timebox_minutes: 12,
  fields: ['customer_reply', 'internal_note', 'verification_steps', 'ai_interaction'],
  brief:
    'You work at fictional electronics shop Inyange Tech. A customer writes: "My MoMo was debited twice for my headphones. Send my money back today." Order IT-3307 shows one completed payment of RWF 35,000. The payment provider log shows a second attempt with status \'failed - reversal pending\'. You cannot see the customer\'s wallet statement. Reversals and refunds need finance team confirmation. Write a customer response and an internal note. You may use the AI drafting assistant, but you must verify its suggestions.',
};

function previewAttempt(): Attempt | null {
  try {
    return JSON.parse(localStorage.getItem(PREVIEW_KEY) ?? 'null');
  } catch {
    return null;
  }
}

function savePreview(a: Attempt): Attempt {
  try {
    localStorage.setItem(PREVIEW_KEY, JSON.stringify(a));
  } catch {
    throw new FriendlyError('Your browser blocked saving. Turn off private browsing and try again.');
  }
  return a;
}

export function clearPreviewBaseline() {
  if (PREVIEW) localStorage.removeItem(PREVIEW_KEY);
}

export async function getScenario(): Promise<Scenario> {
  if (PREVIEW) return PREVIEW_SCENARIO;
  return request<Scenario>('GET', '/v1/baseline/scenario');
}

// Returns the candidate's baseline attempt without creating one.
export async function getBaseline(): Promise<Attempt | null> {
  if (PREVIEW) return previewAttempt();
  const { attempts } = await request<{ attempts: Attempt[] }>('GET', '/v1/attempts');
  const found = attempts.find((a) => a.stage === 'baseline' && a.status !== 'voided');
  return found ? request<Attempt>('GET', `/v1/attempts/${found.id}`) : null;
}

// Safe to call more than once: the backend returns the same attempt.
export async function startBaseline(program: ProgramId): Promise<Attempt> {
  if (PREVIEW) return previewAttempt() ?? savePreview({ id: 'preview', stage: 'baseline', status: 'in_progress', answers: {}, submitted_at: null });
  const attempt = await request<Attempt>('POST', '/v1/attempts/baseline', { program });
  // A new attempt starts as assigned; opening it lets the candidate submit later.
  return attempt.status === 'assigned' ? request<Attempt>('POST', `/v1/attempts/${attempt.id}/start`) : attempt;
}

export async function saveAnswers(id: string, answers: Answers): Promise<void> {
  if (PREVIEW) return void savePreview({ ...(previewAttempt() as Attempt), answers });
  await request('PUT', `/v1/attempts/${id}/answers`, { answers });
}

export async function submitBaseline(id: string, answers: Answers): Promise<Attempt> {
  await saveAnswers(id, answers);
  if (PREVIEW) return savePreview({ ...(previewAttempt() as Attempt), status: 'submitted', submitted_at: new Date().toISOString() });
  return request<Attempt>('POST', `/v1/attempts/${id}/submit`);
}

export type PlanReason = { code: string; competency?: string | null };
export type PlanModule = { position: number; course_code: string; title: string; reasons: PlanReason[]; completed_at: string | null };
export type Progress = {
  baseline: { attempt_id: string; status: AttemptStatus; submitted_at: string | null } | null;
  result: { review_id: string; weighted_score: number | null; proposed_outcome: string | null; scores: { dimension: string; score: number }[] } | null;
  plan: { id: string; modules: PlanModule[] } | null;
};

export async function getProgress(): Promise<Progress> {
  if (PREVIEW) {
    const a = previewAttempt();
    return { baseline: a && { attempt_id: a.id, status: a.status, submitted_at: a.submitted_at }, result: null, plan: null };
  }
  return request<Progress>('GET', '/v1/me/progress');
}
