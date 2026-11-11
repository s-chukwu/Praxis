'use client';
import { useState } from 'react';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';
import { skillNames } from './skills';

type Suggestion = {
  competency: string;
  evidence_quote: string;
  status: 'suggested' | 'confirmed' | 'rejected';
};

const MAX = 20000;

export default function CvConsent() {
  const [consent, setConsent] = useState(false);
  const [text, setText] = useState('');
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [unavailable, setUnavailable] = useState(false);

  const canSubmit = consent && text.trim().length > 0 && text.length <= MAX;

  const getSuggestions = () => {
    // Suggestions will come from the backend. Until that route exists, a reviewer checks the CV.
    setSuggestions([]);
    setUnavailable(true);
  };

  const decide = (code: string, status: 'confirmed' | 'rejected') =>
    setSuggestions((list) => list.map((s) => (s.competency === code ? { ...s, status } : s)));

  return (
    <div className="space-y-6">
      <section className="space-y-4 rounded-lg border border-[#161B22] bg-[#0D1117] p-5">
        <label className="flex items-start gap-3 text-sm text-gray-300">
          <input
            type="checkbox"
            checked={consent}
            onChange={(e) => setConsent(e.target.checked)}
            className="mt-0.5 h-4 w-4 accent-blue-600"
          />
          I agree that Praxis may analyse my CV to suggest skills.
        </label>

        <Textarea
          id="cv-text"
          label="CV text"
          rows={8}
          maxLength={MAX}
          value={text}
          onChange={(e) => setText(e.target.value)}
        />
        <div className="flex justify-between text-xs text-gray-400">
          <span>Paste the text of your CV. Emails, phone numbers and ID numbers are hidden before analysis.</span>
          <span>
            {text.length.toLocaleString()} / {MAX.toLocaleString()}
          </span>
        </div>

        <Button type="button" onClick={getSuggestions} disabled={!canSubmit} className="disabled:cursor-not-allowed disabled:opacity-50">
          Get suggestions
        </Button>
      </section>

      {suggestions.length > 0 && (
        <section>
          <h2 className="mb-3 text-base font-semibold text-white">Suggested skills</h2>
          <ul className="divide-y divide-[#161B22] rounded-lg border border-[#161B22] bg-[#0D1117]">
            {suggestions.map((s) => (
              <li key={s.competency} className="flex items-center justify-between gap-4 p-4">
                <div>
                  <p className="text-sm font-medium text-white">
                    {s.competency} {skillNames[s.competency]}
                  </p>
                  <p className="mt-1 text-xs text-gray-400">&ldquo;{s.evidence_quote}&rdquo;</p>
                </div>
                {s.status === 'suggested' ? (
                  <div className="flex gap-2">
                    <Button type="button" onClick={() => decide(s.competency, 'confirmed')}>Confirm</Button>
                    <Button type="button" variant="secondary" onClick={() => decide(s.competency, 'rejected')}>Reject</Button>
                  </div>
                ) : (
                  <span className="rounded border border-[#1C222B] px-2 py-0.5 text-xs text-gray-300">
                    {s.status === 'confirmed' ? 'Confirmed' : 'Rejected'}
                  </span>
                )}
              </li>
            ))}
          </ul>
          <p className="mt-3 text-xs text-gray-400">
            Suggestions are claims from your CV. They do not give you a score or skip any course.
          </p>
        </section>
      )}

      {unavailable && (
        <div role="status" className="rounded-lg border border-[#161B22] bg-[#0D1117] p-4">
          <p className="text-sm font-medium text-white">Suggestions are unavailable</p>
          <p className="mt-1 text-sm text-gray-400">A reviewer will check your CV.</p>
        </div>
      )}
    </div>
  );
}