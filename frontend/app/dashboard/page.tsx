'use client';
import { Clock } from 'lucide-react';
import { ButtonLink } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import SkillProgress from '@/components/dashboard/SkillProgress';
import PathList, { pathSteps } from '@/components/path/PathList';
import { useProgress } from '@/lib/useProgress';
import { programs } from '@/lib/catalog';

function greeting() {
  const h = new Date().getHours();
  return h < 12 ? 'Good morning' : h < 18 ? 'Good afternoon' : 'Good evening';
}

const NEXT: Record<string, { text: string; waitingTitle?: string; waiting?: string; button?: string }> = {
  'Baseline task': { text: 'Reply to a real customer case. Your answers show which skills you already have.' },
  Review: {
    text: '',
    waitingTitle: 'Your work is with a reviewer',
    waiting: 'A trained reviewer is scoring your baseline task. Your result and learning plan appear here when they finish.',
  },
  'Learning plan': {
    text: 'Your courses are ready, in the order to take them.',
    waitingTitle: 'Your plan is on its way',
    waiting: 'Your review is done. Your learning plan appears here as soon as it is ready.',
    button: 'Open my plan',
  },
};

export default function Dashboard() {
  const { learner } = useLearner();
  const progress = useProgress();
  if (!learner) return null;
  const steps = pathSteps(learner, progress);
  const next = steps.find((s) => s.state === 'next' || s.state === 'waiting') ?? steps[1];
  const copy = NEXT[next.title] ?? NEXT['Baseline task'];
  const firstName = learner.name.split(' ')[0];

  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-[28px] font-bold leading-tight">
          {greeting()}
          {firstName && `, ${firstName}`}
        </h1>
        {learner.program && <p className="mt-1 text-muted">{programs[learner.program].name}</p>}
      </header>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_340px]">
        <SkillProgress learner={learner} />

        <section aria-labelledby="next-title" className="flex flex-col rounded-panel bg-ink p-6 text-white shadow-panel">
          <h2 id="next-title" className="text-sm font-semibold text-white/70">
            Your next step
          </h2>
          <p className="mt-2 text-2xl font-bold leading-snug">{next.state === 'waiting' ? copy.waitingTitle : next.title}</p>
          <p className="mt-2 text-white/75">{next.state === 'waiting' ? copy.waiting : copy.text}</p>
          {next.title === 'Baseline task' && (
            <p className="mt-4 flex items-center gap-2 text-sm text-white/75">
              <Clock className="h-4 w-4" aria-hidden /> 8 to 12 minutes, no timer
            </p>
          )}
          {next.state === 'next' && next.href && (
            <ButtonLink href={next.href} size="lg" className="mt-6 lg:mt-auto">
              {next.title === 'Baseline task' ? `${next.action} baseline task` : copy.button}
            </ButtonLink>
          )}
        </section>
      </div>

      <section aria-labelledby="path-title">
        <h2 id="path-title" className="mb-3 text-lg font-bold">
          Your path
        </h2>
        <PathList learner={learner} progress={progress} />
      </section>
    </div>
  );
}
