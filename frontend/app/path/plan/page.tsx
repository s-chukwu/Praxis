'use client';
import PageHeader from '@/components/ui/PageHeader';
import PlanList from '@/components/path/PlanList';
import { useProgress } from '@/lib/useProgress';

export default function PlanPage() {
  const progress = useProgress();
  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader title="Learning plan" description="Only the courses you need, in the order to take them." />
      <PlanList progress={progress} />
    </div>
  );
}
