'use client';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { CheckCircle2 } from 'lucide-react';
import { Button, ButtonLink } from '@/components/ui/Button';
import ScorePicker from '@/components/reviewer/ScorePicker';
import { programs, skillNames } from '@/lib/catalog';
import { friendly } from '@/lib/errors';
import {
  clearDraft,
  finalizeReview,
  getQueue,
  loadDraft,
  openReview,
  saveScores,
  storeDraft,
  type Draft,
  type QueueItem,
  type Review,
  type ReviewCase,
} from '@/lib/reviewer';

const DIMENSIONS = [
  { id: 'accuracy', title: 'Accuracy', desc: 'States only what the records show. No invented facts or promises.' },
  { id: 'judgment', title: 'Judgment', desc: 'Escalates to the right team with the right context.' },
  { id: 'safety', title: 'Safety', desc: 'Protects customer data and checks AI output before using it.' },
  { id: 'communication', title: 'Communication', desc: 'Clear, respectful and easy for the customer to act on.' },
];

const ANSWER_LABELS: Record<string, string> = {
  customer_reply: 'Reply to the customer',
  internal_note: 'Internal note',
  verification_steps: 'How they would check their answer',
  ai_interaction: 'How they used AI',
};

type Dim = { score?: number; evidence: string };
const squash = (s: string) => s.replace(/\s+/g, ' ').trim().toLowerCase();

export default function ReviewPage() {
  const { attemptId } = useParams<{ attemptId: string }>();
  const [queue, setQueue] = useState<QueueItem[]>([]);
  const [review, setReview] = useState<Review | null>(null);
  const [rc, setRc] = useState<ReviewCase | null>(null);
  const [dims, setDims] = useState<Record<string, Dim>>({});
  const [draft, setDraft] = useState<Draft>({ competencies: {}, notes: '' });
  const [loadError, setLoadError] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState('');
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState<'save' | 'submit' | null>(null);
  const [done, setDone] = useState(false);

  useEffect(() => {
    (async () => {
      const items = await getQueue();
      setQueue(items);
      const item = items.find((i) => i.attempt_id === attemptId);
      if (!item) throw new Error('This case is no longer in the queue. It may have been reviewed already.');
      const opened = await openReview(item);
      setReview(opened.review);
      setRc(opened.reviewCase);
      setDims(Object.fromEntries(opened.review.scores.map((s) => [s.dimension, { score: s.score, evidence: s.evidence[0] ?? '' }])));
      setDraft(loadDraft(opened.review.id));
    })().catch((e) => setLoadError(e instanceof Error && e.message.startsWith('This case') ? e.message : friendly(e)));
  }, [attemptId]);

  if (loadError) {
    return (
      <div className="mx-auto max-w-2xl px-5 py-12">
        <p role="alert" className="rounded-panel bg-danger-soft p-6 font-medium text-danger">
          {loadError}
        </p>
        <ButtonLink href="/reviewer" variant="secondary" className="mt-4">
          Back to queue
        </ButtonLink>
      </div>
    );
  }
  if (!review || !rc) return <div className="mx-auto mt-10 h-96 max-w-[1280px] animate-pulse rounded-panel bg-surface" aria-label="Loading the case" />;

  const competencies = programs[rc.program].skills;
  const allAnswers = squash(Object.values(rc.answers).join(' '));
  const position = queue.findIndex((i) => i.attempt_id === attemptId);
  const open = queue.filter((i) => i.status === 'submitted' || i.mine);
  const nextItem = open.find((i, n) => n > open.findIndex((o) => o.attempt_id === attemptId) && i.attempt_id !== attemptId);

  const setDim = (id: string, patch: Partial<Dim>) => {
    setDims((all) => ({ ...all, [id]: { ...(all[id] ?? { evidence: '' }), ...patch } }));
    setErrors(({ [id]: _, ...rest }) => rest);
  };
  const setCompetency = (code: string, score: number) => {
    const next = { ...draft, competencies: { ...draft.competencies, [code]: score } };
    setDraft(next);
    storeDraft(review.id, next);
    setErrors(({ [code]: _, ...rest }) => rest);
  };

  // The AI service checks every quote against the answers, so catching a typo here saves a rejected review.
  const checkDim = (id: string, d?: Dim) => {
    if (d?.score === undefined) return 'Choose a score.';
    const quote = d.evidence.trim();
    if (quote.length < 12 || quote.length > 500) return 'Paste a quote of 12 to 500 characters.';
    if (!allAnswers.includes(squash(quote))) return 'This quote is not in the candidate’s answers. Copy it exactly.';
    return '';
  };
  const scored = () =>
    DIMENSIONS.filter((d) => dims[d.id]?.score !== undefined && !checkDim(d.id, dims[d.id])).map((d) => ({
      dimension: d.id,
      score: dims[d.id].score as number,
      evidence: [dims[d.id].evidence.trim()],
    }));

  const saveDraft = async () => {
    setBusy('save');
    setFormError('');
    storeDraft(review.id, draft);
    try {
      const ready = scored();
      if (ready.length) await saveScores(review.id, ready);
      setNotice('Draft saved');
    } catch (e) {
      setFormError(friendly(e));
    }
    setBusy(null);
  };

  const submit = async () => {
    const found: Record<string, string> = {};
    for (const d of DIMENSIONS) {
      const msg = checkDim(d.id, dims[d.id]);
      if (msg) found[d.id] = msg;
    }
    for (const c of competencies) if (draft.competencies[c] === undefined) found[c] = 'Score this skill.';
    setErrors(found);
    if (Object.keys(found).length) return setFormError(`Finish ${Object.keys(found).length} more item(s) before submitting.`);

    setBusy('submit');
    setFormError('');
    try {
      await saveScores(review.id, scored());
      await finalizeReview(review.id, { competencies: Object.fromEntries(competencies.map((c) => [c, draft.competencies[c]])), notes: draft.notes });
      clearDraft(review.id);
      setDone(true);
    } catch (e) {
      setFormError(friendly(e));
    }
    setBusy(null);
  };

  if (done) {
    return (
      <div className="mx-auto max-w-xl px-5 py-16 text-center">
        <CheckCircle2 className="mx-auto h-12 w-12 animate-pop text-brand" aria-hidden />
        <h1 className="mt-4 text-2xl font-bold">Review submitted</h1>
        <p className="mt-2 text-muted">The score is locked and {rc.candidateRef}’s learning plan has been created.</p>
        <div className="mt-6 flex justify-center gap-3">
          {nextItem && <ButtonLink href={`/reviewer/review/${nextItem.attempt_id}`}>Next review</ButtonLink>}
          <ButtonLink href="/reviewer" variant="secondary">
            Back to queue
          </ButtonLink>
        </div>
      </div>
    );
  }

  return (
    <div className="pb-28">
      <div className="border-b border-line bg-surface">
        <div className="mx-auto flex max-w-[1280px] items-center gap-4 px-5 py-2.5 text-sm">
          <span className="text-muted">
            Review {position + 1} of {queue.length}
          </span>
          <div className="h-1.5 w-40 overflow-hidden rounded-full bg-line" aria-hidden>
            <div className="h-full rounded-full bg-brand" style={{ width: `${((position + 1) / Math.max(queue.length, 1)) * 100}%` }} />
          </div>
          <Link href="/reviewer" className="ml-auto font-semibold text-brand-ink hover:underline">
            Back to queue
          </Link>
        </div>
      </div>

      <div className="mx-auto max-w-[1280px] px-5 pt-6">
        <h1 className="text-2xl font-bold tabular-nums">{rc.candidateRef}</h1>
        <p className="mt-1 text-muted">{programs[rc.program].name} · Baseline task</p>

        <div className="mt-6 grid gap-6 lg:grid-cols-2 lg:items-start">
          <section aria-labelledby="evidence-title" className="space-y-4 lg:sticky lg:top-20">
            <h2 id="evidence-title" className="sr-only">
              Candidate evidence
            </h2>
            <div className="rounded-panel bg-surface p-5 shadow-panel">
              <p className="text-sm font-semibold text-muted">The customer case</p>
              <p className="mt-2 leading-relaxed">{rc.brief}</p>
            </div>
            {Object.entries(rc.answers).map(([key, text]) => (
              <div key={key} className="rounded-panel bg-surface p-5 shadow-panel">
                <p className="text-sm font-semibold text-muted">{ANSWER_LABELS[key] ?? key}</p>
                <p className="mt-2 whitespace-pre-line leading-relaxed">{text || <span className="text-muted">No answer</span>}</p>
              </div>
            ))}
            <div className="rounded-panel border border-progress/30 bg-progress-soft p-5">
              <p className="text-sm font-semibold text-[#7A4E09]">Marking guide (reviewers only)</p>
              <ul className="mt-2 list-disc space-y-1 pl-5 text-sm">
                {rc.checks.map((c) => (
                  <li key={c}>{c}</li>
                ))}
              </ul>
            </div>
          </section>

          <section aria-labelledby="rubric-title" className="space-y-6">
            <div className="rounded-panel bg-surface p-5 shadow-panel">
              <h2 id="rubric-title" className="font-bold">
                Rubric
              </h2>
              <p className="text-sm text-muted">Score 0 to 4. Paste a quote from the answers as evidence for each.</p>
              <div className="mt-4 space-y-5">
                {DIMENSIONS.map((d) => (
                  <div key={d.id}>
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div className="min-w-0 flex-1">
                        <p className="font-semibold">{d.title}</p>
                        <p className="text-sm text-muted">{d.desc}</p>
                      </div>
                      <ScorePicker label={`${d.title} score`} value={dims[d.id]?.score} onChange={(n) => setDim(d.id, { score: n })} />
                    </div>
                    <textarea
                      aria-label={`${d.title} evidence quote`}
                      rows={2}
                      value={dims[d.id]?.evidence ?? ''}
                      onChange={(e) => setDim(d.id, { evidence: e.target.value })}
                      placeholder="Quote from the candidate’s answer"
                      className="mt-2 w-full resize-y rounded-control border border-line px-3 py-2 text-sm focus:border-brand focus:outline-none"
                    />
                    {errors[d.id] && <p className="mt-1 text-sm font-medium text-danger">{errors[d.id]}</p>}
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-panel bg-surface p-5 shadow-panel">
              <h2 className="font-bold">Skills</h2>
              <p className="text-sm text-muted">These scores build the candidate’s learning plan. Every skill needs a score.</p>
              <div className="mt-4 space-y-3">
                {competencies.map((c) => (
                  <div key={c}>
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <p className="font-medium">
                        <span className="text-muted">{c}</span> {skillNames[c]}
                      </p>
                      <ScorePicker label={`${skillNames[c]} score`} value={draft.competencies[c]} onChange={(n) => setCompetency(c, n)} />
                    </div>
                    {errors[c] && <p className="mt-1 text-sm font-medium text-danger">{errors[c]}</p>}
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-panel bg-surface p-5 shadow-panel">
              <label htmlFor="notes" className="font-bold">
                Notes
              </label>
              <textarea
                id="notes"
                rows={3}
                maxLength={2000}
                value={draft.notes}
                onChange={(e) => {
                  const next = { ...draft, notes: e.target.value };
                  setDraft(next);
                  storeDraft(review.id, next);
                }}
                placeholder="Anything the next reviewer or the candidate’s coach should know"
                className="mt-2 w-full resize-y rounded-control border border-line px-3 py-2 text-sm focus:border-brand focus:outline-none"
              />
            </div>
          </section>
        </div>
      </div>

      <div className="fixed inset-x-0 bottom-0 z-20 border-t border-line bg-surface">
        <div className="mx-auto flex max-w-[1280px] flex-wrap items-center gap-3 px-5 py-3">
          <Button variant="secondary" loading={busy === 'save'} disabled={busy !== null} onClick={saveDraft}>
            Save draft
          </Button>
          <span className="text-sm text-muted" aria-live="polite">
            {formError ? <span className="font-medium text-danger">{formError}</span> : notice}
          </span>
          <Button className="ml-auto" loading={busy === 'submit'} disabled={busy !== null} onClick={submit}>
            Submit review
          </Button>
        </div>
      </div>
    </div>
  );
}
