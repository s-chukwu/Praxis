'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import Logo from './Logo';

const items = [
  { label: 'Overview', href: '/', icon: 'M3 12l9-9 9 9M5 10v10h5v-6h4v6h5V10' },
  { label: 'Skill Catalog', href: '/skill-catalog', icon: 'M4 4h6v6H4z M14 4h6v6h-6z M4 14h6v6H4z M14 14h6v6h-6z' },
  {
    label: 'Assessments',
    href: '/assessments',
    icon: 'M9 4h6v3H9z M7 5H6a1 1 0 00-1 1v14a1 1 0 001 1h12a1 1 0 001-1V6a1 1 0 00-1-1h-1 M9 12h6 M9 16h6',
  },
  { label: 'Candidate Review', href: '/candidate-review', icon: 'M12 12a4 4 0 100-8 4 4 0 000 8z M4 21a8 8 0 0116 0' },
];

export default function Sidebar({ open, onClose }: { open: boolean; onClose: () => void }) {
  const path = usePathname();
  const isActive = (href: string) => (href === '/' ? path === '/' : path.startsWith(href));

  return (
    <>
      {open && <div className="fixed inset-0 z-20 bg-black/60 md:hidden" onClick={onClose} aria-hidden="true" />}
      <aside
        className={`fixed inset-y-0 left-0 z-30 flex w-[220px] shrink-0 flex-col border-r border-[#161B22] bg-[#080B0F] p-3 transition-transform md:sticky md:top-0 md:h-screen md:translate-x-0 ${
          open ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="px-2 pb-6 pt-2">
          <Logo />
        </div>
        <nav aria-label="Main" className="space-y-1">
          {items.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              aria-current={isActive(item.href) ? 'page' : undefined}
              className={`flex items-center gap-2.5 rounded-md border px-2.5 py-2 text-sm transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-400 ${
                isActive(item.href)
                  ? 'border-[#1C222B] bg-[#11151B] text-white'
                  : 'border-transparent text-gray-400 hover:bg-[#0F1319] hover:text-white'
              }`}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <path d={item.icon} />
              </svg>
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
    </>
  );
}