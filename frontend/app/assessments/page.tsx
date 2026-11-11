import PageHeader from '@/components/ui/PageHeader';
import EmptyState from '@/components/ui/EmptyState';


export default function Assessments() {
  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <PageHeader title="Assessments" description="Your results and assessments in progress." />
      <EmptyState
        title="No assessments yet"
        description="Pick a skill to start your first one."
        actionHref="/skill-catalog"
        actionLabel="Browse skills"
      />
    </div>
  );
}