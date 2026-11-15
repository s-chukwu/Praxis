'use client';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { Check, CheckCircle2, Lock, Sparkles, X } from 'lucide-react';
import Logo from '@/components/Logo';
import Textarea from '@/components/ui/Textarea';
import { Button, ButtonLink } from '@/components/ui/Button';
import { COURSES, CONTENT_STATUS } from '@/lib/courses';
import { courseNames } from '@/lib/catalog';
import { completeCourse, previewCompleted, type PlanModule } from '@/lib/baseline';
import { useProgress } from '@/lib/useProgress';
import { friendly } from '@/lib/errors';
import { PREVIEW } from '@/lib/learner';

const STEPS = ['What you will learn', 'Worked example', 'Using AI safely', 'Your practice', 'Check yourself'];
const MIN_WORDS = 15;

const practiceKey = (code: string) => `praxis.practice.${code}`;

export default function CoursePage() {
  const { code } = useParams<{ code: string }>();
  const course = COURSES[code];
  const progress = useProgress();
  const [step, setStep] = useState(0);
  const [practice, setPractice] = useState('');
  const [ticked, setTicked] = useState<Set<number>>(new Set());
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [done, setDone] = useState(false);

  // The practice answer stays in this browser so leaving mid-course never loses it.
  useEffect(() => {
    try {
      setPractice(localStorage.getItem(practiceKey(code)) ?? '');
    } catch {
      // Private browsing: start empty.
    }
  }, [code]);

  const modules: PlanModule[] = (progress?.plan?.modules ?? []).map((m) =>
    PREVIEW && previewCompleted().includes(m.course_code) ? { ...m, completed_at: m.completed_at ?? 'preview' } : m,
  );
  const index = modules.findIndex((m) => m.course_code === code);
  const mod = modules[index];
  const locked = index > 0 && modules.slice(0, index).some((m) => !m.completed_at);
  const already = Boolean(mod?.completed_at);
  const next = modules.slice(index + 1).find((m) => !m.completed_at);
  const words = practice.trim().split(/\s+/).filter(Boolean).length;
  const title = mod?.title || courseNames[code] || code;

  if (!course) {
    return (
      <div className="mx-auto max-w-xl px-5 py-16 text-center">
        <p className="font-semibold">This course is not in the catalogue.</p>
        <ButtonLink href="/path/plan" variant="secondary" className="mt-4">
          Back to my plan
        </ButtonLink>
      </div>
    );
  }

  const savePractice = (text: string) => {
    setPractice(text);
    try {
      localStorage.setItem(practiceKey(code), text);
    } catch {
      // Kept in memory for this visit.
    }
  };

  const finish = async () => {
    setBusy(true);
    setError('');
    try {
      await completeCourse(code);
      setDone(true);
    } catch (e) {
      setError(friendly(e));
    }
    setBusy(false);
  };

  const canFinish = words >= MIN_WORDS && ticked.size === course.checks.length;

  if (done) {
    return (
      <div className="min-h-screen bg-page">
        <div className="mx-auto max-w-xl px-5 py-20 text-center">
          <CheckCircle2 className="mx-auto h-12 w-12 animate-pop text-brand" aria-hidden />
          <h1 className="mt-4 text-2xl font-bold">{title} complete</h1>
          <p className="mt-2 text-muted">
            {next ? `Next in your plan: ${next.title || courseNames[next.course_code]}.` : 'That was the last course in your plan. Your final task is open.'}
          </p>
          <div className="mt-7 flex justify-center gap-3">
            {next ? <ButtonLink href={`/learn/${next.course_code}`}>Start the next course</ButtonLink> : <ButtonLink href="/path/final">Open the final task</ButtonLink>}
            <ButtonLink href="/path/plan" variant="secondary">
              Back to my plan
            </ButtonLink>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-page pb-28">
      <header className="sticky top-0 z-20 border-b border-line bg-surface">
        <div className="mx-auto flex h-14 max-w-[960px] items-center gap-4 px-5">
          <Logo />
          <span className="hidden text-sm text-muted sm:inline">
            {index >= 0 ? `Course ${index + 1} of ${modules.length}` : 'Course'} · Step {step + 1} of {STEPS.length}
          </span>
          <div className="hidden h-1.5 w-32 overflow-hidden rounded-full bg-line sm:block" aria-hidden>
            <div className="h-full rounded-full bg-brand transition-[width] duration-500" style={{ width: `${((step + 1) / STEPS.length) * 100}%` }} />
          </div>
          <Link href="/path/plan" className="ml-auto flex items-center gap-1.5 text-sm font-semibold text-brand-ink hover:underline">
            <X className="h-4 w-4" aria-hidden /> Close
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-[960px] px-5 pt-8">
        <h1 className="text-[28px] font-bold leading-tight">{title}</h1>
        <p className="mt-1 text-sm text-muted">{CONTENT_STATUS}</p>

        {locked ? (
          <div className="mt-8 flex gap-3 rounded-panel bg-surface p-6 shadow-panel">
            <Lock className="mt-0.5 h-5 w-5 text-muted" aria-hidden />
            <div>
              <p className="font-semibold">Finish the earlier courses first</p>
              <p className="mt-1 text-sm text-muted">Your plan puts courses in order, so each one builds on the last.</p>
              <ButtonLink href="/path/plan" variant="secondary" className="mt-4">
                Back to my plan
              </ButtonLink>
            </div>
          </div>
        ) : (
          <div className="mt-8 grid gap-6 lg:grid-cols-[200px_minmax(0,1fr)]">
            <ol className="hidden space-y-1 lg:block" aria-label="Course steps">
              {STEPS.map((label, i) => (
                <li key={label}>
                  <button
                    type="button"
                    onClick={() => setStep(i)}
                    aria-current={i === step ? 'step' : undefined}
                    className={`flex w-full items-center gap-2.5 rounded-control px-3 py-2 text-left text-sm ${i === step ? 'bg-surface font-semibold shadow-panel' : 'text-muted hover:text-ink'}`}
                  >
                    <span className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-[11px] font-bold ${i < step ? 'bg-brand text-white' : 'bg-page ring-1 ring-line'}`}>
                      {i < step ? <Check className="h-3 w-3" strokeWidth={3} aria-hidden /> : i + 1}
                    </span>
                    {label}
                  </button>
                </li>
              ))}
            </ol>

            <section key={step} className="animate-rise rounded-panel bg-surface p-6 shadow-panel" aria-live="polite">
              <h2 className="text-lg font-bold">{STEPS[step]}</h2>
              {step === 0 && <p className="mt-3 text-lg leading-relaxed">{course.objective}</p>}
              {step === 1 && (
                <div className="mt-3 space-y-4">
                  <p className="leading-relaxed">{course.example.situation}</p>
                  <div className="rounded-control bg-danger-soft p-4">
                    <p className="text-sm font-semibold text-danger">Weak</p>
                    <p className="mt-1">{course.example.weak}</p>
                  </div>
                  <div className="rounded-control bg-brand-soft p-4">
                    <p className="text-sm font-semibold text-brand-ink">Strong</p>
                    <p className="mt-1">{course.example.strong}</p>
                  </div>
                  <p className="text-muted">{course.example.why}</p>
                </div>
              )}
              {step === 2 && (
                <ul className="mt-3 space-y-3">
                  {course.ai.map((tip) => (
                    <li key={tip} className="flex gap-3">
                      <Sparkles className="mt-0.5 h-4 w-4 shrink-0 text-progress" aria-hidden />
                      <span className="leading-relaxed">{tip}</span>
                    </li>
                  ))}
                </ul>
              )}
              {step === 3 && (
                <div className="mt-3 space-y-4">
                  <p className="leading-relaxed">{course.practice}</p>
                  <Textarea
                    id="practice"
                    label="Your answer"
                    rows={7}
                    value={practice}
                    onChange={(e) => savePractice(e.target.value)}
                    hint={`${words} words${words < MIN_WORDS ? `, at least ${MIN_WORDS} to finish` : ''}. Saved in this browser as you type.`}
                  />
                </div>
              )}
              {step === 4 && (
                <div className="mt-3 space-y-3">
                  <p className="text-muted">Read your practice answer again. Tick each line only if it is true.</p>
                  {practice.trim() && <p className="whitespace-pre-line rounded-control bg-page p-4 text-sm">{practice}</p>}
                  {course.checks.map((c, i) => (
                    <label key={c} className="flex cursor-pointer items-start gap-3 rounded-control border border-line p-3">
                      <input
                        type="checkbox"
                        checked={ticked.has(i)}
                        onChange={() =>
                          setTicked((all) => {
                            const next = new Set(all);
                            if (next.has(i)) next.delete(i);
                            else next.add(i);
                            return next;
                          })
                        }
                        className="mt-0.5 h-4 w-4 accent-[#0F7B5F]"
                      />
                      <span>{c}</span>
                    </label>
                  ))}
                  {words < MIN_WORDS && <p className="text-sm text-[#7A4E09]">Write your practice answer first (step 4).</p>}
                </div>
              )}
            </section>
          </div>
        )}
      </main>

      {!locked && (
        <div className="fixed inset-x-0 bottom-0 z-20 border-t border-line bg-surface">
          <div className="mx-auto flex max-w-[960px] items-center gap-3 px-5 py-3">
            <Button variant="secondary" disabled={step === 0} onClick={() => setStep((s) => s - 1)}>
              Back
            </Button>
            {error && <span className="text-sm font-medium text-danger">{error}</span>}
            {step < STEPS.length - 1 ? (
              <Button className="ml-auto" onClick={() => setStep((s) => s + 1)}>
                Next
              </Button>
            ) : (
              <Button className="ml-auto" disabled={!canFinish} loading={busy} onClick={finish}>
                {already ? 'Done' : 'Mark course complete'}
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
