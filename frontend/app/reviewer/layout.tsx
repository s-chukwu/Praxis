'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { LogOut } from 'lucide-react';
import Logo from '@/components/Logo';
import { useLearner } from '@/components/LearnerProvider';
import { PREVIEW } from '@/lib/learner';

const NAV = [
  { label: 'Queue', href: '/reviewer' },
  { label: 'My reviews', href: '/reviewer/reviews' },
  { label: 'Calibration', href: '/reviewer/calibration' },
  { label: 'Guidelines', href: '/reviewer/guidelines' },
  { label: 'Profile', href: '/reviewer/profile' },
];

// Focus mode: no learner sidebar, one slim bar, and the review screen hides even the nav.
export default function ReviewerLayout({ children }: { children: React.ReactNode }) {
  const path = usePathname();
  const { signOut } = useLearner();
  const reviewing = path.startsWith('/reviewer/review/');

  return (
    <div className="min-h-screen bg-page">
      <header className="sticky top-0 z-20 border-b border-line bg-surface">
        <div className="mx-auto flex h-14 max-w-[1280px] items-center gap-6 px-5">
          <Link href="/reviewer" className="flex items-center gap-2" aria-label="Reviewer queue">
            <Logo />
            <span className="rounded-full bg-brand-soft px-2 py-0.5 text-xs font-semibold text-brand-ink">Reviewer</span>
          </Link>
          {!reviewing && (
            <nav aria-label="Reviewer" className="hidden flex-1 gap-1 md:flex">
              {NAV.map((n) => {
                const active = n.href === '/reviewer' ? path === n.href : path.startsWith(n.href);
                return (
                  <Link
                    key={n.href}
                    href={n.href}
                    aria-current={active ? 'page' : undefined}
                    className={`rounded-control px-3 py-1.5 text-sm font-medium ${active ? 'bg-page text-ink' : 'text-muted hover:text-ink'}`}
                  >
                    {n.label}
                  </Link>
                );
              })}
            </nav>
          )}
          <button
            onClick={() => void signOut()}
            className="ml-auto flex items-center gap-1.5 rounded-control px-2.5 py-1.5 text-sm font-medium text-muted hover:bg-page hover:text-ink"
          >
            <LogOut className="h-4 w-4" aria-hidden /> Sign out
          </button>
        </div>
        {!reviewing && (
          <nav aria-label="Reviewer mobile" className="flex gap-1 overflow-x-auto px-3 pb-2 md:hidden">
            {NAV.map((n) => (
              <Link key={n.href} href={n.href} className="shrink-0 rounded-control px-3 py-1.5 text-sm font-medium text-muted">
                {n.label}
              </Link>
            ))}
          </nav>
        )}
      </header>
      {PREVIEW && (
        <p className="fixed bottom-4 right-4 z-40 rounded-full bg-ink px-3.5 py-1.5 text-xs font-semibold text-white shadow-lift">
          Preview: sample cases, nothing is saved online
        </p>
      )}
      <main className={reviewing ? '' : 'mx-auto w-full max-w-[1080px] px-5 py-8 md:py-10'}>{children}</main>
    </div>
  );
}
