'use client';
import { useState } from 'react';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';

export type Scenario = { id: string; brief: string; timeboxMinutes: number };

const fields = [
  { id: 'customer_reply', label: 'Customer reply', rows: 5 },
  { id: 'internal_note', label: 'Internal note', rows: 3 },
  { id: 'verification_steps', label: 'Verification steps (3 to 5)', rows: 4 },
  { id: 'ai_record', label: 'AI interaction record (optional)', rows: 3 },
];

export default function BaselineForm({ scenario }: { scenario?: Scenario }) {
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [message, setMessage] = useState('');

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setMessage('Submitting is not connected to the backend yet.');
  };

  return (
    <form onSubmit={submit} className="space-y-5">
      <section className="rounded-lg border border-[#161B22] bg-[#0D1117] p-5">
        <p className="text-xs text-gray-400">
          Scenario{scenario ? ` · Time guide: ${scenario.timeboxMinutes} minutes` : ''}
        </p>
        <p className="mt-2 text-sm leading-relaxed text-gray-300">
          {scenario ? scenario.brief : 'Your scenario will appear here.'}
        </p>
      </section>

      {fields.map((f) => (
        <Textarea
          key={f.id}
          id={f.id}
          label={f.label}
          rows={f.rows}
          value={answers[f.id] ?? ''}
          onChange={(e) => setAnswers((a) => ({ ...a, [f.id]: e.target.value }))}
        />
      ))}

      {message && (
        <div role="status" className="rounded-md border border-[#1C222B] bg-[#080B0F] p-2 text-sm text-gray-300">
          {message}
        </div>
      )}
      <Button type="submit">Submit answers</Button>
    </form>
  );
}