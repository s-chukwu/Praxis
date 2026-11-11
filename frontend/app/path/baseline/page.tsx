import PageHeader from '@/components/ui/PageHeader';
import BaselineForm from '@/components/path/BaselineForm';

export default function BaselinePage() {
  // The scenario will come from the backend. Until then the form shows a placeholder.
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <p className="mb-1 text-xs text-gray-400">Step 2 of 4</p>
        <PageHeader title="Baseline task" description="Write your answers to the customer scenario." />
      </div>
      <BaselineForm />
    </div>
  );
}