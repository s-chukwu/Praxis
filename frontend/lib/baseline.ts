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
  updated_at?: string;
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
type AttemptSummary = { attempt_id: string; status: AttemptStatus; submitted_at: string | null };
type ResultSummary = { review_id: string; weighted_score: number | null; proposed_outcome: string | null; scores: { dimension: string; score: number }[] };
export type Progress = {
  baseline: AttemptSummary | null;
  result: ResultSummary | null;
  plan: { id: string; modules: PlanModule[] } | null;
  final?: AttemptSummary | null;
  final_result?: ResultSummary | null;
};

export async function getProgress(): Promise<Progress> {
  if (PREVIEW) {
    const a = previewAttempt();
    return { baseline: a && { attempt_id: a.id, status: a.status, submitted_at: a.submitted_at }, result: null, plan: null };
  }
  return request<Progress>('GET', '/v1/me/progress');
}

export type Stage = 'baseline' | 'final';

// Preview keeps course completions in this browser so the plan can be walked end to end.
const PREVIEW_DONE_KEY = 'praxis.courses.preview';

export async function completeCourse(code: string): Promise<void> {
  if (PREVIEW) {
    try {
      const done: string[] = JSON.parse(localStorage.getItem(PREVIEW_DONE_KEY) ?? '[]');
      localStorage.setItem(PREVIEW_DONE_KEY, JSON.stringify(Array.from(new Set([...done, code]))));
    } catch {
      // Private browsing: nothing to remember.
    }
    return;
  }
  await request('POST', `/v1/me/plan/courses/${code}/complete`);
}

export function previewCompleted(): string[] {
  try {
    return JSON.parse(localStorage.getItem(PREVIEW_DONE_KEY) ?? '[]');
  } catch {
    return [];
  }
}

// The final task opens only when the plan is done; the backend refuses it otherwise.
export async function getFinalScenario(): Promise<Scenario> {
  if (PREVIEW) return { ...PREVIEW_SCENARIO, scenario_id: 'CS-01-B' };
  return request<Scenario>('GET', '/v1/final/scenario');
}

export async function getFinal(): Promise<Attempt | null> {
  if (PREVIEW) return null;
  const { attempts } = await request<{ attempts: Attempt[] }>('GET', '/v1/attempts');
  const found = attempts.find((a) => a.stage === 'final' && a.status !== 'voided');
  return found ? request<Attempt>('GET', `/v1/attempts/${found.id}`) : null;
}

export async function startFinal(): Promise<Attempt> {
  if (PREVIEW) return { id: 'preview-final', stage: 'final', status: 'in_progress', answers: {}, submitted_at: null };
  return request<Attempt>('POST', '/v1/attempts/final');
}

export async function submitAttempt(id: string, answers: Answers): Promise<Attempt> {
  if (PREVIEW && id === 'preview-final') return { id, stage: 'final', status: 'submitted', answers, submitted_at: new Date().toISOString() };
  return submitBaseline(id, answers);
}
