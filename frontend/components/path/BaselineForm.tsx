'use client';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { CheckCircle2, Clock, Hourglass } from 'lucide-react';
import Textarea from '@/components/ui/Textarea';
import { Button, ButtonLink } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import { friendly, report } from '@/lib/errors';
import { ago, hoursLeft, REVIEW_HOURS } from '@/lib/journey';
import { getBaseline, getScenario, saveAnswers, startBaseline, submitBaseline, type Answers, type Attempt, type Scenario } from '@/lib/baseline';

type Field = { label: string; hint?: string; rows: number; required?: boolean };

// Keys must match the scenario's fields exactly, or scoring rejects the answer.
const FIELDS: Record<string, Field> = {
  customer_reply: { label: 'Your reply to the customer', hint: 'Write the reply you would send. Use plain language.', rows: 6, required: true },
  internal_note: { label: 'Internal note for your team', hint: 'Add anything your team needs to know about this case.', rows: 3, required: true },
  verification_steps: {
    label: 'How you would check your answer',
    hint: 'Three to five steps, one per line. Explain how you would confirm the reply is correct.',
    rows: 4,
    required: true,
  },
  ai_interaction: { label: 'How you used AI (optional)', hint: 'Paste the AI draft and what you changed.', rows: 3 },
};

const words = (t: string) => t.trim().split(/\s+/).filter(Boolean).length;

const fieldFor = (id: string): Field => FIELDS[id] ?? { label: id.replace(/_/g, ' '), rows: 4, required: true };

type Save = 'idle' | 'saving' | 'saved' | 'failed';

export default function BaselineForm() {
  const { learner } = useLearner();
  const [scenario, setScenario] = useState<Scenario | null>(null);
  const [attempt, setAttempt] = useState<Attempt | null>(null);
  const [answers, setAnswers] = useState<Answers>({});
  const [loadError, setLoadError] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState('');
  const [busy, setBusy] = useState(false);
  const [save, setSave] = useState<Save>('idle');
  const [resumedAt, setResumedAt] = useState<string | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    Promise.all([getScenario(), getBaseline()])
      .then(([s, a]) => {
        setScenario(s);
        setAttempt(a);
        setAnswers(a?.answers ?? {});
        if (a?.updated_at && Object.values(a.answers ?? {}).some((v) => v.trim())) setResumedAt(a.updated_at);
      })
      .catch((e) => setLoadError(friendly(e)));
    return () => clearTimeout(timer.current);
  }, []);

  const open = attempt?.status === 'in_progress' || attempt?.status === 'assigned';
  const required = scenario ? scenario.fields.filter((id) => fieldFor(id).required) : [];

  const start = async () => {
    if (!learner?.program) return setFormError('Choose your program first, then start the task.');
    setBusy(true);
    setFormError('');
    try {
      const a = await startBaseline(learner.program);
      setAttempt(a);
      setAnswers(a.answers ?? {});
    } catch (e) {
      setFormError(friendly(e));
    }
    setBusy(false);
  };

  // Saves 1.5 seconds after the last keystroke, so typing never waits on the network.
  const change = (id: string, value: string) => {
    const next = { ...answers, [id]: value };
    setAnswers(next);
    if (errors[id]) setErrors(({ [id]: _, ...rest }) => rest);
    if (!attempt) return;
    clearTimeout(timer.current);
    setSave('saving');
    timer.current = setTimeout(() => {
      saveAnswers(attempt.id, next)
        .then(() => setSave('saved'))
        .catch((e) => {
          report(e);
          setSave('failed');
        });
    }, 1500);
  };

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (!attempt || !scenario) return;
    const found: Record<string, string> = {};
    for (const id of scenario.fields) {
      if (fieldFor(id).required && !answers[id]?.trim()) found[id] = 'This answer is required.';
    }
    setErrors(found);
    const first = scenario.fields.find((id) => found[id]);
    if (first) {
      setFormError('Answer the required questions before submitting.');
      return document.getElementById(first)?.focus();
    }

    clearTimeout(timer.current);
    setBusy(true);
    setFormError('');
    try {
      setAttempt(await submitBaseline(attempt.id, answers));
    } catch (err) {
      setFormError(friendly(err));
    }
    setBusy(false);
  };

  if (loadError) {
    return (
      <p role="alert" className="rounded-panel bg-danger-soft p-6 font-medium text-danger">
        {loadError}
      </p>
    );
  }
  if (!scenario) {
    return <div className="h-64 animate-pulse rounded-panel bg-surface shadow-panel" aria-label="Loading your task" />;
  }

  if (attempt && !open) {
    const left = hoursLeft(attempt.submitted_at);
    return (
      <div role="status" className="animate-rise mx-auto max-w-2xl rounded-panel bg-surface p-8 shadow-panel">
        <CheckCircle2 className="h-11 w-11 animate-pop text-brand" aria-hidden />
        <h2 className="mt-4 text-2xl font-bold">Case closed on your side</h2>
        <p className="mt-2 max-w-prose text-muted">Your answers are locked and with a trained support supervisor. This is the first entry in your training record.</p>
        <div className="mt-6 rounded-control bg-page p-4">
          <p className="flex items-center gap-2 font-semibold">
            <Hourglass className="h-4 w-4 text-progress" aria-hidden />
            {left === null ? 'Reviewed within 48 hours' : left > 0 ? `Review due within ${left} ${left === 1 ? 'hour' : 'hours'}` : 'Taking a little longer than usual'}
          </p>
          {left !== null && (
            <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-line" aria-hidden>
              <div className="h-full rounded-full bg-progress" style={{ width: `${((REVIEW_HOURS - left) / REVIEW_HOURS) * 100}%` }} />
            </div>
          )}
        </div>
        <ol className="mt-6 space-y-3 text-sm">
          {['A reviewer scores your four rubric areas and each skill.', 'Your result shows your strongest area and where to start.', 'Your learning plan lists only the courses you need.'].map((t, i) => (
            <li key={t} className="flex gap-3">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-brand-soft text-xs font-bold text-brand-ink">{i + 1}</span>
              <span className="pt-0.5">{t}</span>
            </li>
          ))}
        </ol>
        <ButtonLink href="/dashboard" size="lg" className="mt-7">
          Back to dashboard
        </ButtonLink>
      </div>
    );
  }

  return (
    <form onSubmit={submit} className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)] lg:items-start" noValidate>
      <section aria-labelledby="case-title" className="rounded-panel bg-surface p-6 shadow-panel lg:sticky lg:top-8">
        <div className="flex items-center justify-between gap-3">
          <h2 id="case-title" className="font-bold">
            The customer case
          </h2>
          <span className="flex items-center gap-1.5 text-sm text-muted">
            <Clock className="h-4 w-4" aria-hidden /> About {scenario.timebox_minutes} minutes
          </span>
        </div>
        <p className="mt-3 leading-relaxed text-ink/90">{scenario.brief}</p>
      </section>

      <div className="space-y-6">
        {!attempt ? (
          <div className="rounded-panel bg-surface p-6 shadow-panel">
            <p className="font-semibold">Ready when you are</p>
            <p className="mt-1 text-sm text-muted">There is no timer. Your answers save as you type, so you can leave and come back.</p>
            <Button size="lg" className="mt-5" loading={busy} onClick={start}>
              Start the task
            </Button>
          </div>
        ) : (
          <>
            {resumedAt && (
              <p role="status" className="animate-rise rounded-control bg-brand-soft px-4 py-3 text-sm text-brand-ink">
                Welcome back. Everything you wrote was saved {ago(resumedAt)}.
              </p>
            )}
            <div className="sticky top-16 z-10 -mx-1 rounded-control bg-page/95 px-1 py-2 backdrop-blur md:top-4">
              <div className="flex items-center justify-between text-sm">
                <span className="font-semibold">
                  <span className="tabular-nums">{required.filter((id) => answers[id]?.trim()).length}</span> of {required.length} answers written
                </span>
                <span className="text-muted" aria-live="polite">
                  {save === 'saving' ? 'Saving…' : save === 'saved' ? 'Saved' : save === 'failed' ? 'Not saved yet' : ''}
                </span>
              </div>
              <div className="mt-2 flex gap-1.5" aria-hidden>
                {required.map((id) => (
                  <span key={id} className={`h-1.5 flex-1 rounded-full transition-colors duration-500 ${answers[id]?.trim() ? 'bg-brand' : 'bg-line'}`} />
                ))}
              </div>
            </div>
            {scenario.fields.map((id) => {
              const f = fieldFor(id);
              return (
                <Textarea
                  key={id}
                  id={id}
                  label={f.label}
                  hint={answers[id]?.trim() ? `${f.hint ?? ''} ${words(answers[id])} words.`.trim() : f.hint}
                  error={errors[id]}
                  rows={f.rows}
                  value={answers[id] ?? ''}
                  onChange={(e) => change(id, e.target.value)}
                />
              );
            })}
            <div className="flex flex-wrap items-center gap-4">
              <Button type="submit" size="lg" loading={busy}>
                Submit my answers
              </Button>
              <span className="text-sm text-muted">
                {save === 'failed' ? 'Not saved yet. Check your connection.' : ''}
              </span>
            </div>
            <p className="text-sm text-muted">After you submit, your answers go to a reviewer and can no longer be changed.</p>
          </>
        )}
        {formError && (
          <p role="alert" className="rounded-control bg-danger-soft px-4 py-3 text-sm font-medium text-danger">
            {formError}
          </p>
        )}
      </div>
    </form>
  );
}
