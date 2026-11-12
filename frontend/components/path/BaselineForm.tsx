'use client';
import { useState } from 'react';
import { Clock } from 'lucide-react';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';

export type Scenario = { id: string; brief: string; timeboxMinutes: number };

const fields = [
  { id: 'customer_reply', label: 'Reply to the customer', rows: 6 },
  { id: 'internal_note', label: 'Internal note for your team', rows: 3 },
  { id: 'verification_steps', label: 'Verification steps', hint: 'Three to five steps, one per line.', rows: 4 },
  { id: 'ai_record', label: 'How you used AI (optional)', hint: 'Paste the AI draft and what you changed.', rows: 3 },
];

export default function BaselineForm({ scenario }: { scenario?: Scenario }) {
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [message, setMessage] = useState('');

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setMessage('Submitting opens once the assessment service is connected. Your answers stay on this page until then.');
  };

  return (
    <form onSubmit={submit} className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)] lg:items-start">
      <section aria-labelledby="case-title" className="rounded-panel bg-surface p-6 shadow-panel lg:sticky lg:top-8">
        <div className="flex items-center justify-between gap-3">
          <h2 id="case-title" className="font-bold">
            The customer case
          </h2>
          {scenario && (
            <span className="flex items-center gap-1.5 text-sm text-muted">
              <Clock className="h-4 w-4" aria-hidden /> {scenario.timeboxMinutes} minutes
            </span>
          )}
        </div>
        <p className="mt-3 leading-relaxed text-ink/90">
          {scenario ? scenario.brief : 'Your case appears here when you start the task. Take your time to read it before you answer.'}
        </p>
      </section>

      <div className="space-y-6">
        {fields.map((f) => (
          <Textarea
            key={f.id}
            id={f.id}
            label={f.label}
            hint={f.hint}
            rows={f.rows}
            value={answers[f.id] ?? ''}
            onChange={(e) => setAnswers((a) => ({ ...a, [f.id]: e.target.value }))}
          />
        ))}
        {message && (
          <p role="status" className="rounded-control bg-progress-soft px-4 py-3 text-sm text-[#7A4E09]">
            {message}
          </p>
        )}
        <Button type="submit" size="lg">
          Submit my answers
        </Button>
      </div>
    </form>
  );
}
