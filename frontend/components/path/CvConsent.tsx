'use client';
import { useState } from 'react';
import { Check } from 'lucide-react';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import CvDropzone from './CvDropzone';
import { programs, skillNames } from '@/lib/catalog';
import { CV_MAX_CHARS } from '@/lib/cvText';
import { friendly } from '@/lib/errors';
import { saveLearner, type SkillStatus } from '@/lib/learner';

// PM decision X-06: the CV is optional, never sent to the AI service, and only creates claimed skills.
export default function CvConsent({ onDone }: { onDone: () => void | Promise<void> }) {
  const { learner, refresh } = useLearner();
  const codes = programs[learner?.program ?? 'operator'].skills;
  const [text, setText] = useState('');
  const [consent, setConsent] = useState(false);
  const [claimed, setClaimed] = useState<Set<string>>(
    () => new Set(codes.filter((c) => learner?.skills[c] === 'confirmed' || learner?.skills[c] === 'verified')),
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const toggle = (code: string) =>
    setClaimed((all) => {
      const next = new Set(all);
      if (next.has(code)) next.delete(code);
      else next.add(code);
      return next;
    });

  const complete = async (save: boolean) => {
    if (save && text.trim() && !consent) return setError('Tick the consent box to keep your CV, or clear the CV text.');
    setSaving(true);
    setError('');
    try {
      if (save) {
        // Verified skills come only from a reviewer, so a claim never overwrites one.
        const skills: Record<string, SkillStatus> = { ...learner?.skills };
        for (const c of codes) {
          if (skills[c] === 'verified') continue;
          if (claimed.has(c)) skills[c] = 'confirmed';
          else delete skills[c];
        }
        await saveLearner({ cvSubmitted: Boolean(text.trim() && consent), skills });
        await refresh();
      }
      await onDone();
    } catch (e) {
      setError(friendly(e));
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <section aria-labelledby="cv-title" className="space-y-5 rounded-panel bg-surface p-6 shadow-panel">
        <div>
          <h2 id="cv-title" className="font-bold">
            Your CV <span className="font-normal text-muted">(optional)</span>
          </h2>
          <p className="mt-1 text-sm text-muted">Upload a PDF, then check the text. Edit anything that came out wrong.</p>
        </div>
        <CvDropzone onText={setText} />
        <Textarea
          id="cv-text"
          label="CV text"
          rows={7}
          maxLength={CV_MAX_CHARS}
          value={text}
          onChange={(e) => setText(e.target.value)}
          hint={`${text.length.toLocaleString()} of ${CV_MAX_CHARS.toLocaleString()} characters.`}
        />
        {text.trim() && (
          <label className="flex cursor-pointer items-start gap-3 text-sm">
            <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} className="mt-0.5 h-[18px] w-[18px] rounded accent-[#0F7B5F]" />
            <span>Praxis may keep this CV text in my profile. It is not analysed by AI, and I can delete it at any time.</span>
          </label>
        )}
      </section>

      <section aria-labelledby="claims-title" className="rounded-panel bg-surface p-6 shadow-panel">
        <h2 id="claims-title" className="font-bold">
          Skills you already use
        </h2>
        <p className="mt-1 text-sm text-muted">Tick the ones you use at work today. These are your claims; your baseline task shows a reviewer what you can do.</p>
        <ul className="mt-4 grid gap-2 sm:grid-cols-2">
          {codes.map((c) => {
            const on = claimed.has(c);
            const verified = learner?.skills[c] === 'verified';
            return (
              <li key={c}>
                <button
                  type="button"
                  aria-pressed={on || verified}
                  disabled={verified}
                  onClick={() => toggle(c)}
                  className={`flex w-full items-center gap-3 rounded-control border p-3 text-left text-sm transition ${
                    on || verified ? 'border-brand/50 bg-brand-soft/40' : 'border-line hover:border-brand/40'
                  }`}
                >
                  <span className={`flex h-5 w-5 shrink-0 items-center justify-center rounded ${on || verified ? 'bg-brand text-white' : 'border border-line'}`}>
                    {(on || verified) && <Check key={c} className="h-3 w-3 animate-pop" strokeWidth={3} aria-hidden />}
                  </span>
                  <span className="font-medium">{skillNames[c]}</span>
                  {verified && <span className="ml-auto text-xs text-brand-ink">Verified</span>}
                </button>
              </li>
            );
          })}
        </ul>
      </section>

      {error && (
        <p role="alert" className="text-sm font-medium text-danger">
          {error}
        </p>
      )}
      <div className="flex flex-wrap gap-3">
        <Button size="lg" loading={saving} onClick={() => complete(true)}>
          Continue
        </Button>
        <Button size="lg" variant="quiet" disabled={saving} onClick={() => complete(false)}>
          Skip for now
        </Button>
      </div>
    </div>
  );
}
