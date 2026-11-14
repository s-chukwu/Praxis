'use client';
import { useState } from 'react';
import { ScanText } from 'lucide-react';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import CvDropzone from './CvDropzone';
import SkillReview, { type Decision } from './SkillReview';
import { programs } from '@/lib/catalog';
import { CV_MAX_CHARS } from '@/lib/cvText';
import { friendly } from '@/lib/errors';
import { DEMO_MODE, saveLearner, suggestSkills, type SkillStatus, type Suggestion } from '@/lib/learner';

export default function CvConsent({ onDone }: { onDone: () => void | Promise<void> }) {
  const { learner, refresh } = useLearner();
  const [consent, setConsent] = useState(false);
  const [text, setText] = useState('');
  const [analysing, setAnalysing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<Suggestion[] | 'unavailable' | null>(null);
  const [decisions, setDecisions] = useState<Record<string, Decision>>({});

  const inProgram = (code: string) => !learner?.program || programs[learner.program].skills.includes(code);
  const sentences = text.split(/[.!?\n]+/).filter((s) => s.trim().length >= 12).length;

  const analyse = async () => {
    setAnalysing(true);
    const res = await suggestSkills(text);
    setAnalysing(false);
    setResult(res.status === 'ok' ? res.suggestions.filter((s) => inProgram(s.competency)) : 'unavailable');
  };

  const complete = async (save: boolean) => {
    setSaving(true);
    setError('');
    try {
      if (save) {
        const skills: Record<string, SkillStatus> = { ...learner?.skills, ...decisions };
        await saveLearner({ cvSubmitted: true, skills });
        await refresh();
      }
      await onDone();
    } catch (e) {
      setError(friendly(e));
      setSaving(false);
    }
  };

  const errorLine = error && (
    <p role="alert" className="text-sm font-medium text-danger">
      {error}
    </p>
  );

  if (analysing) {
    return (
      <div role="status" className="animate-rise rounded-panel bg-surface p-8 text-center shadow-panel">
        <ScanText className="mx-auto h-8 w-8 text-brand" aria-hidden />
        <p className="mt-4 text-lg font-bold">Reading your CV</p>
        <p className="mt-1 text-muted">
          Checking {sentences} {sentences === 1 ? 'sentence' : 'sentences'} for evidence of your skills.
        </p>
        <div className="mx-auto mt-6 h-1.5 max-w-xs overflow-hidden rounded-full bg-line">
          <div className="h-full w-1/2 animate-sweep rounded-full bg-brand" />
        </div>
      </div>
    );
  }

  if (result !== null) {
    const suggestions = result === 'unavailable' ? [] : result;
    const pending = suggestions.filter((s) => !decisions[s.competency]).length;
    return (
      <div className="space-y-6">
        {suggestions.length > 0 ? (
          <SkillReview suggestions={suggestions} decisions={decisions} onDecide={(code, d) => setDecisions((all) => ({ ...all, [code]: d }))} />
        ) : (
          <div role="status" className="animate-rise rounded-panel bg-surface p-6 shadow-panel">
            <p className="font-semibold">{result === 'unavailable' ? 'Suggestions are unavailable right now' : 'No skills found to suggest'}</p>
            <p className="mt-1 text-muted">
              {result === 'unavailable' ? 'A reviewer will read your CV instead. You can carry on.' : 'That is fine. Your baseline task shows what you can do.'}
            </p>
          </div>
        )}
        <p className="text-sm text-muted">Confirmed skills are your claims. They never give you a score or skip a course.</p>
        {errorLine}
        <div className="flex flex-wrap items-center gap-3">
          <Button size="lg" loading={saving} disabled={pending > 0} onClick={() => complete(true)}>
            Continue
          </Button>
          {pending > 0 && <span className="text-sm text-muted">Answer {pending} more to continue.</span>}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="space-y-5 rounded-panel bg-surface p-6 shadow-panel">
        <CvDropzone onText={setText} />
        <Textarea
          id="cv-text"
          label="CV text"
          rows={8}
          maxLength={CV_MAX_CHARS}
          value={text}
          onChange={(e) => setText(e.target.value)}
          hint={`Check or edit the text. Emails, phone and ID numbers are hidden before analysis. ${text.length.toLocaleString()} of ${CV_MAX_CHARS.toLocaleString()} characters.`}
        />
        <label className="flex cursor-pointer items-start gap-3 text-sm">
          <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} className="mt-0.5 h-[18px] w-[18px] rounded accent-[#0F7B5F]" />
          <span>Praxis may analyse my CV to suggest skills. I can change my answers later.</span>
        </label>
      </div>

      {DEMO_MODE && <p className="text-sm text-muted">Demo mode: suggestions come from simple matching, not the AI model.</p>}
      {errorLine}

      <div className="flex flex-wrap gap-3">
        <Button size="lg" disabled={!consent || !text.trim()} onClick={analyse}>
          Find my skills
        </Button>
        <Button size="lg" variant="quiet" loading={saving} onClick={() => complete(false)}>
          Skip for now
        </Button>
      </div>
    </div>
  );
}
