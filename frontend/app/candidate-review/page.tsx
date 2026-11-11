import PageHeader from '@/components/ui/PageHeader';
import EmptyState from '@/components/ui/EmptyState';


export default function CandidateReview() {
  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <PageHeader title="Candidate Review" description="Review candidate skills, interview notes and recommendations." />
      <EmptyState title="No candidates yet" description="Candidates appear here after they complete assessments." />
    </div>
  );
}