'use client';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { ClipboardPen, LayoutDashboard, ListOrdered, LogOut, Route, X } from 'lucide-react';
import Logo from './Logo';
import { useLearner } from './LearnerProvider';
import { programs } from '@/lib/catalog';
import { PREVIEW } from '@/lib/learner';

const items = [
  { label: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { label: 'My path', href: '/path', icon: Route, exact: true },
  { label: 'Baseline task', href: '/path/baseline', icon: ClipboardPen },
  { label: 'Learning plan', href: '/path/plan', icon: ListOrdered },
];

export default function Sidebar({ open, onClose }: { open: boolean; onClose: () => void }) {
  const path = usePathname();
  const router = useRouter();
  const { learner, signOut } = useLearner();
  const program = learner?.program ? programs[learner.program].name : null;

  return (
    <>
      {open && <div className="fixed inset-0 z-20 bg-ink/30 md:hidden" onClick={onClose} aria-hidden="true" />}
      <aside
        className={`fixed inset-y-0 left-0 z-30 flex w-[248px] shrink-0 flex-col border-r border-line bg-surface transition-transform duration-300 ease-out md:sticky md:top-0 md:h-screen md:translate-x-0 ${
          open ? 'translate-x-0 shadow-lift' : '-translate-x-full'
        }`}
      >
        <div className="flex h-16 items-center justify-between px-5">
          <Link href="/dashboard" aria-label="Praxis dashboard">
            <Logo />
          </Link>
          <button onClick={onClose} aria-label="Close menu" className="rounded-control p-1.5 text-muted hover:bg-page md:hidden">
            <X className="h-5 w-5" aria-hidden />
          </button>
        </div>

        <nav aria-label="Main" className="flex-1 space-y-0.5 px-3 pt-2">
          {items.map(({ label, href, icon: Icon, exact }) => {
            const active = exact ? path === href : path.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                aria-current={active ? 'page' : undefined}
                className={`flex h-10 items-center gap-3 rounded-control px-3 text-sm font-semibold transition-colors ${
                  active ? 'bg-brand-soft text-brand-ink' : 'text-muted hover:bg-page hover:text-ink'
                }`}
              >
                <Icon className="h-[18px] w-[18px]" strokeWidth={2} aria-hidden />
                {label}
              </Link>
            );
          })}
        </nav>

        <div className="border-t border-line p-3">
          <div className="px-3 py-2">
            <p className="truncate text-sm font-semibold text-ink">{learner?.name || 'Learner'}</p>
            {program && <p className="truncate text-[13px] text-muted">{program}</p>}
          </div>
          <button
            onClick={async () => {
              await signOut();
              router.replace(PREVIEW ? '/onboarding' : '/login');
            }}
            className="flex h-10 w-full items-center gap-3 rounded-control px-3 text-sm font-semibold text-muted hover:bg-page hover:text-ink"
          >
            <LogOut className="h-[18px] w-[18px]" aria-hidden />
            {PREVIEW ? 'Restart preview' : 'Sign out'}
          </button>
        </div>
      </aside>
    </>
  );
}
