'use client';
import { useEffect, useRef, useState } from 'react';
import { Bot, Headset } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { programs, type ProgramId } from '@/lib/catalog';
import { friendly } from '@/lib/errors';

const icons = { operator: Headset, builder: Bot };

export default function ProgramDialog({ onChoose }: { onChoose: (program: ProgramId) => Promise<void> }) {
  const ref = useRef<HTMLDialogElement>(null);
  const [choice, setChoice] = useState<ProgramId | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  // showModal() gives the focus trap, Escape handling and inert background for free.
  useEffect(() => ref.current?.showModal(), []);

  return (
    <dialog
      ref={ref}
      aria-labelledby="program-title"
      onCancel={(e) => e.preventDefault()}
      className="w-[min(640px,calc(100vw-32px))] rounded-[20px] bg-surface p-0 text-ink shadow-lift backdrop:bg-ink/40 backdrop:backdrop-blur-[2px]"
    >
      <div className="p-6 sm:p-8">
        <h2 id="program-title" className="text-2xl font-bold">
          Choose your program
        </h2>
        <p className="mt-1 text-muted">You can switch before your baseline task.</p>

        <div role="radiogroup" aria-labelledby="program-title" className="mt-6 grid gap-3 sm:grid-cols-2">
          {(Object.keys(programs) as ProgramId[]).map((id) => {
            const p = programs[id];
            const Icon = icons[id];
            const selected = choice === id;
            return (
              <button
                key={id}
                role="radio"
                aria-checked={selected}
                onClick={() => setChoice(id)}
                className={`flex flex-col rounded-panel border-2 p-5 text-left transition-colors ${
                  selected ? 'border-brand bg-brand-soft/50' : 'border-line hover:border-ink/25'
                }`}
              >
                <span className={`flex h-10 w-10 items-center justify-center rounded-control ${selected ? 'bg-brand text-white' : 'bg-page text-ink'}`}>
                  <Icon className="h-5 w-5" aria-hidden />
                </span>
                <span className="mt-4 text-lg font-bold leading-snug">{p.name}</span>
                <span className="mt-1.5 text-sm text-muted">{p.summary}</span>
                <span className="mt-4 text-[13px] font-semibold text-ink">{p.skills.length} skills assessed</span>
              </button>
            );
          })}
        </div>

        {error && (
          <p role="alert" className="mt-4 text-sm font-medium text-danger">
            {error}
          </p>
        )}
        <Button
          size="lg"
          className="mt-6 w-full sm:w-auto"
          disabled={!choice}
          loading={busy}
          onClick={async () => {
            if (!choice) return;
            setBusy(true);
            try {
              await onChoose(choice);
              ref.current?.close();
            } catch (e) {
              setError(friendly(e));
              setBusy(false);
            }
          }}
        >
          {choice ? `Start ${programs[choice].name}` : 'Select a program'}
        </Button>
      </div>
    </dialog>
  );
}
