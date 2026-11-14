'use client';
import { useEffect, useState } from 'react';
import PageHeader from '@/components/ui/PageHeader';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import { supabase } from '@/lib/supabase';
import { PREVIEW } from '@/lib/learner';

export default function ReviewerProfile() {
  const { learner, role, signOut } = useLearner();
  const [email, setEmail] = useState(PREVIEW ? 'reviewer@example.com' : '');

  useEffect(() => {
    if (PREVIEW) return;
    supabase?.auth.getUser().then(({ data }) => setEmail(data.user?.email ?? ''));
  }, []);

  return (
    <div className="max-w-2xl space-y-6">
      <PageHeader title="Profile" />
      <dl className="divide-y divide-line rounded-panel bg-surface shadow-panel">
        {[
          ['Name', learner?.name || 'Not set'],
          ['Email', email || 'Loading…'],
          ['Role', role === 'admin' ? 'Admin' : 'Reviewer'],
        ].map(([k, v]) => (
          <div key={k} className="flex justify-between gap-4 p-5">
            <dt className="text-muted">{k}</dt>
            <dd className="font-medium">{v}</dd>
          </div>
        ))}
      </dl>
      <p className="text-sm text-muted">Reviewer access is given by an admin. To change your name or email, ask your programme lead.</p>
      <Button variant="secondary" onClick={() => void signOut()}>
        Sign out
      </Button>
    </div>
  );
}
