'use client';
import { useEffect, useState } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { LearnerProvider, useLearner } from './LearnerProvider';
import Sidebar from './Sidebar';
import Topbar from './Topbar';
import ErrorScreen from './ErrorScreen';
import { PREVIEW } from '@/lib/learner';

// The landing page belongs to another team; these routes render without the learner shell.
const PUBLIC = ['/', '/login', '/signup'];

export default function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <LearnerProvider>
      <Frame>{children}</Frame>
    </LearnerProvider>
  );
}

function Frame({ children }: { children: React.ReactNode }) {
  const path = usePathname();
  const router = useRouter();
  const { learner, role, ready, loadError, refresh } = useLearner();
  const [menuOpen, setMenuOpen] = useState(false);
  const isPublic = PUBLIC.includes(path);
  const isOnboarding = path.startsWith('/onboarding');
  const isReviewerArea = path.startsWith('/reviewer');
  const staff = role === 'reviewer' || role === 'admin';

  useEffect(() => setMenuOpen(false), [path]);

  useEffect(() => {
    if (!ready || isPublic || loadError) return;
    if (!learner) return router.replace('/login');
    // Preview has no real roles, so both areas stay open for demos.
    if (PREVIEW) {
      if (!isReviewerArea && !learner.onboarded && !isOnboarding) router.replace('/onboarding');
      return;
    }
    if (staff && !isReviewerArea) router.replace('/reviewer');
    else if (!staff && isReviewerArea) router.replace('/dashboard');
    else if (!staff && !learner.onboarded && !isOnboarding) router.replace('/onboarding');
  }, [ready, learner, loadError, isPublic, isOnboarding, isReviewerArea, staff, path, router]);

  if (isPublic) return <>{children}</>;
  if (loadError) return <ErrorScreen message={loadError} onRetry={() => void refresh()} />;
  if (!ready || !learner) return <div className="min-h-screen bg-page" aria-busy="true" />;
  // The reviewer workspace has its own focus layout, separate from the learner app.
  if (isOnboarding || isReviewerArea) return <>{children}</>;

  return (
    <div className="flex min-h-screen">
      {PREVIEW && (
        <p className="fixed bottom-4 right-4 z-40 rounded-full bg-ink px-3.5 py-1.5 text-xs font-semibold text-white shadow-lift">
          Preview: sample learner, nothing is saved online
        </p>
      )}
      <Sidebar open={menuOpen} onClose={() => setMenuOpen(false)} />
      <div className="flex min-w-0 flex-1 flex-col">
        <Topbar onMenu={() => setMenuOpen(true)} />
        <main className="mx-auto w-full max-w-[1120px] flex-1 px-5 py-8 md:px-10 md:py-10">{children}</main>
      </div>
    </div>
  );
}
