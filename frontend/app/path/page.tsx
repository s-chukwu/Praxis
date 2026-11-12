'use client';
import PageHeader from '@/components/ui/PageHeader';
import PathList from '@/components/path/PathList';
import { useLearner } from '@/components/LearnerProvider';

export default function MyPath() {
  const { learner } = useLearner();
  if (!learner) return null;
  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader title="My path" description="Five steps from your first task to verified skills." />
      <PathList learner={learner} />
    </div>
  );
}
