'use client';
import PageHeader from '@/components/ui/PageHeader';
import PathList from '@/components/path/PathList';
import { useLearner } from '@/components/LearnerProvider';
import { useProgress } from '@/lib/useProgress';

export default function MyPath() {
  const { learner } = useLearner();
  const progress = useProgress();
  if (!learner) return null;
  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader title="My path" description="Five steps from your first task to verified skills." />
      <PathList learner={learner} progress={progress} />
    </div>
  );
}
