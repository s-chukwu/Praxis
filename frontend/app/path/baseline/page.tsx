import PageHeader from '@/components/ui/PageHeader';
import BaselineForm from '@/components/path/BaselineForm';

export const metadata = { title: 'Baseline task' };

export default function BaselinePage() {
  return (
    <div className="space-y-8">
      <PageHeader
        title="Baseline task"
        description="Read the customer case, write your reply, and explain how you would check your answer. There are no trick questions."
      />
      <BaselineForm />
    </div>
  );
}
