'use client';
import { useState, type FormEvent } from 'react';
import { Clock } from 'lucide-react';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';

export type Scenario = { id: string; brief: string; timeboxMinutes: number };

type Field = {
  id: string;
  label: string;
  hint?: string;
  rows: number;
  required?: boolean;
};

const fields: Field[] = [
  {
    id: 'customer_reply',
    label: 'Your reply to the customer',
    hint: 'Write the reply you would send. Use plain language.',
    rows: 6,
    required: true,
  },
  {
    id: 'internal_note',
    label: 'Internal note for your team',
    hint: 'Add anything your team needs to know about this case.',
    rows: 3,
    required: true,
  },
  {
    id: 'verification_steps',
    label: 'How you would check your answer',
    hint: 'Three to five steps, one per line. Explain how you would confirm the reply is correct.',
    rows: 4,
    required: true,
  },
  {
    id: 'ai_record',
    label: 'How you used AI (optional)',
    hint: 'Paste the AI draft and what you changed.',
    rows: 3,
  },
];

export default function BaselineForm({ scenario }: { scenario?: Scenario }) {
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState('');
  const [message, setMessage] = useState('');

  const submit = (e: FormEvent) => {
    e.preventDefault();
    const found: Record<string, string> = {};
    for (const field of fields) {
      if (field.required && !answers[field.id]?.trim()) {
        found[field.id] = 'This answer is required.';
      }
    }

    setErrors(found);
    if (Object.keys(found).length) {
      setFormError('Answer the required questions before submitting.');
      setMessage('');
      const first = fields.find((field) => found[field.id]);
      if (first) document.getElementById(first.id)?.focus();
      return;
    }

    setFormError('');
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
        {fields.map((field) => (
          <Textarea
            key={field.id}
            id={field.id}
            label={field.label}
            hint={field.hint}
            error={errors[field.id]}
            rows={field.rows}
            value={answers[field.id] ?? ''}
            onChange={(e) => {
              setAnswers((current) => ({ ...current, [field.id]: e.target.value }));
              if (errors[field.id]) {
                setErrors((current) => {
                  const next = { ...current };
                  delete next[field.id];
                  return next;
                });
              }
            }}
          />
        ))}
        {formError && (
          <p role="alert" className="rounded-control bg-danger-soft px-4 py-3 text-sm font-medium text-danger">
            {formError}
          </p>
        )}
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
