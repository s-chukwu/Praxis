import PageHeader from '@/components/ui/PageHeader';
import BaselineForm from '@/components/path/BaselineForm';

export const metadata = { title: 'Final task' };

export default function FinalPage() {
  return (
    <div className="space-y-8">
      <PageHeader title="Final task" description="A new case, like your baseline. It shows what your training changed. There is no timer." />
      <BaselineForm stage="final" />
    </div>
  );
}
