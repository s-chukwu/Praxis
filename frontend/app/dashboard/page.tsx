'use client';
import { useLearner } from '@/components/LearnerProvider';
import SkillProgress from '@/components/dashboard/SkillProgress';
import NextStep from '@/components/dashboard/NextStep';
import CaseTrail from '@/components/dashboard/CaseTrail';
import ResultStory from '@/components/dashboard/ResultStory';
import { programs } from '@/lib/catalog';
import { useProgress } from '@/lib/useProgress';

function greeting() {
  const h = new Date().getHours();
  return h < 12 ? 'Good morning' : h < 18 ? 'Good afternoon' : 'Good evening';
}

export default function Dashboard() {
  const { learner } = useLearner();
  const progress = useProgress();
  if (!learner) return null;
  const firstName = learner.name.split(' ')[0];

  return (
    <div className="space-y-6">
      <header className="pb-2">
        <h1 className="text-[28px] font-bold leading-tight">
          {greeting()}
          {firstName && `, ${firstName}`}
        </h1>
        {learner.program && <p className="mt-1 text-muted">{programs[learner.program].name}</p>}
      </header>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
        <SkillProgress learner={learner} />
        <NextStep progress={progress} />
      </div>

      {progress?.result && <ResultStory result={progress.result} />}
      <CaseTrail progress={progress} />
    </div>
  );
}
