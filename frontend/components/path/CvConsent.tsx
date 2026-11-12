'use client';
import { useState } from 'react';
import { Check, FileText, Quote, X } from 'lucide-react';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import { programs, skillNames } from '@/lib/catalog';
import { friendly } from '@/lib/errors';
import { DEMO_MODE, saveLearner, suggestSkills, type SkillStatus, type Suggestion } from '@/lib/learner';

const MAX = 20000;
type Decision = 'confirmed' | 'rejected';

export default function CvConsent({ onDone }: { onDone: () => void | Promise<void> }) {
  const { learner, refresh } = useLearner();
  const [consent, setConsent] = useState(false);
  const [text, setText] = useState('');
  const [fileNote, setFileNote] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<Suggestion[] | 'unavailable' | null>(null);
  const [decisions, setDecisions] = useState<Record<string, Decision>>({});

  const inProgram = (code: string) => !learner?.program || programs[learner.program].skills.includes(code);

  const readFile = async (file?: File) => {
    if (!file) return;
    // Our Go backend will extract text from PDF and Word files; in the browser only plain text is safe to read.
    if (!file.name.toLowerCase().endsWith('.txt')) {
      return setFileNote('PDF and Word upload is coming soon. Copy the text from your CV and paste it below.');
    }
    setFileNote('');
    setText((await file.text()).slice(0, MAX));
  };

  const analyse = async () => {
    setBusy(true);
    const res = await suggestSkills(text);
    setBusy(false);
    setResult(res.status === 'ok' ? res.suggestions.filter((s) => inProgram(s.competency)) : 'unavailable');
  };

  const finish = async () => {
    const skills: Record<string, SkillStatus> = { ...learner?.skills };
    for (const [code, d] of Object.entries(decisions)) skills[code] = d;
    try {
      await saveLearner({ cvSubmitted: true, skills });
      await refresh();
      await onDone();
    } catch (e) {
      setError(friendly(e));
    }
  };

  const skip = async () => {
    try {
      await onDone();
    } catch (e) {
      setError(friendly(e));
    }
  };

  if (result !== null) {
    const suggestions = result === 'unavailable' ? [] : result;
    const pending = suggestions.filter((s) => !decisions[s.competency]).length;
    return (
      <div className="space-y-6">
        {result === 'unavailable' ? (
          <div role="status" className="rounded-panel border border-line bg-surface p-6">
            <p className="font-semibold">Suggestions are unavailable right now</p>
            <p className="mt-1 text-muted">A reviewer will read your CV instead. You can carry on.</p>
          </div>
        ) : suggestions.length === 0 ? (
          <div role="status" className="rounded-panel border border-line bg-surface p-6">
            <p className="font-semibold">No skills found to suggest</p>
            <p className="mt-1 text-muted">That is fine. Your baseline task shows what you can do.</p>
          </div>
        ) : (
          <ul className="space-y-3">
            {suggestions.map((s) => {
              const d = decisions[s.competency];
              return (
                <li key={s.competency} className={`rounded-panel border bg-surface p-5 transition-colors ${d === 'confirmed' ? 'border-brand/40' : 'border-line'}`}>
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div className="min-w-0 flex-1">
                      <p className="font-semibold">{skillNames[s.competency]}</p>
                      <p className="mt-2 flex gap-2 text-sm text-muted">
                        <Quote className="mt-0.5 h-3.5 w-3.5 shrink-0" aria-hidden />
                        <span>{s.evidence_quote}</span>
                      </p>
                    </div>
                    <div className="flex gap-2" role="group" aria-label={`Is ${skillNames[s.competency]} right?`}>
                      <Button
                        variant={d === 'confirmed' ? 'primary' : 'secondary'}
                        aria-pressed={d === 'confirmed'}
                        onClick={() => setDecisions((all) => ({ ...all, [s.competency]: 'confirmed' }))}
                      >
                        <Check className="h-4 w-4" aria-hidden /> Yes, that's me
                      </Button>
                      <Button
                        variant="secondary"
                        aria-pressed={d === 'rejected'}
                        className={d === 'rejected' ? 'border-ink/40 bg-page' : ''}
                        onClick={() => setDecisions((all) => ({ ...all, [s.competency]: 'rejected' }))}
                      >
                        <X className="h-4 w-4" aria-hidden /> No
                      </Button>
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
        <p className="text-sm text-muted">Confirmed skills are your claims. They never give you a score or skip a course.</p>
        {error && <p role="alert" className="text-sm font-medium text-danger">{error}</p>}
        <div className="flex flex-wrap items-center gap-3">
          <Button size="lg" onClick={finish} disabled={pending > 0}>
            Continue
          </Button>
          {pending > 0 && <span className="text-sm text-muted">Answer {pending} more to continue.</span>}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="rounded-panel border border-line bg-surface p-6">
        <label className="flex cursor-pointer items-center gap-3 rounded-control border border-dashed border-line px-4 py-4 hover:border-brand/50 hover:bg-brand-soft/40">
          <FileText className="h-5 w-5 text-brand" aria-hidden />
          <span className="flex-1 text-sm">
            <span className="font-semibold">Upload a text file</span>
            <span className="text-muted"> or paste your CV below</span>
          </span>
          <input type="file" accept=".txt,.pdf,.doc,.docx" className="sr-only" onChange={(e) => readFile(e.target.files?.[0])} />
        </label>
        {fileNote && <p className="mt-2 text-sm text-progress">{fileNote}</p>}

        <div className="mt-5">
          <Textarea
            id="cv-text"
            label="CV text"
            rows={9}
            maxLength={MAX}
            value={text}
            onChange={(e) => setText(e.target.value)}
            hint={`Emails, phone and ID numbers are hidden before analysis. ${text.length.toLocaleString()} of ${MAX.toLocaleString()} characters.`}
          />
        </div>

        <label className="mt-5 flex cursor-pointer items-start gap-3 text-sm">
          <input
            type="checkbox"
            checked={consent}
            onChange={(e) => setConsent(e.target.checked)}
            className="mt-0.5 h-[18px] w-[18px] rounded accent-[#0F7B5F]"
          />
          <span>Praxis may analyse my CV to suggest skills. I can change my answers later.</span>
        </label>
      </div>

      {DEMO_MODE && <p className="text-sm text-muted">Demo mode: suggestions come from simple matching, not the AI model.</p>}
      {error && (
        <p role="alert" className="text-sm font-medium text-danger">
          {error}
        </p>
      )}

      <div className="flex flex-wrap gap-3">
        <Button size="lg" loading={busy} disabled={!consent || !text.trim()} onClick={analyse}>
          Find my skills
        </Button>
        <Button size="lg" variant="quiet" onClick={skip}>
          Skip for now
        </Button>
      </div>
    </div>
  );
}
