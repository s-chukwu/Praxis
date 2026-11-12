import PageHeader from '@/components/ui/PageHeader';
import PlanList from '@/components/path/PlanList';

export const metadata = { title: 'Learning plan' };

export default function PlanPage() {
  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader title="Learning plan" description="Only the courses you need, in the order to take them." />
      <PlanList />
    </div>
  );
}
