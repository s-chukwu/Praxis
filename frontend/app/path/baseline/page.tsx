import PageHeader from '@/components/ui/PageHeader';
import BaselineForm from '@/components/path/BaselineForm';

export const metadata = { title: 'Baseline task' };

export default function BaselinePage() {
  // The scenario will come from our Go backend. Until then the form shows its waiting state.
  return (
    <div className="space-y-8">
      <PageHeader title="Baseline task" description="Answer as you would at work. There are no trick questions." />
      <BaselineForm />
    </div>
  );
}
