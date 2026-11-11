import PageHeader from '@/components/ui/PageHeader';
import PlanList from '@/components/path/PlanList';

export default function PlanPage() {
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <p className="mb-1 text-xs text-gray-400">Step 3 of 4</p>
        <PageHeader title="Learning plan" description="Your courses, in order." />
      </div>
      <PlanList />
    </div>
  );
}