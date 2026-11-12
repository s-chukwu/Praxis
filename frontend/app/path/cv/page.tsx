'use client';
import { useRouter } from 'next/navigation';
import PageHeader from '@/components/ui/PageHeader';
import CvConsent from '@/components/path/CvConsent';

export default function CvPage() {
  const router = useRouter();
  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader title="Your CV" description="Update your CV and confirm the skills it suggests." />
      <CvConsent onDone={() => router.push('/dashboard')} />
    </div>
  );
}
