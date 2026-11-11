import PageHeader from '@/components/ui/PageHeader';
import PathList from '@/components/path/PathList';

export default function MyPath() {
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <PageHeader title="My path" description="Four steps, in order." />
      <PathList />
    </div>
  );
}