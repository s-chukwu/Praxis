import PageHeader from '@/components/ui/PageHeader';
import CvConsent from '@/components/path/CvConsent';

export default function CvPage() {
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <p className="mb-1 text-xs text-gray-400">Step 1 of 4</p>
        <PageHeader title="CV and consent" description="Add your CV to get skill suggestions." />
      </div>
      <CvConsent />
    </div>
  );
}