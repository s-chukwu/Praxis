'use client';
import { usePathname } from 'next/navigation';

const titles: Record<string, string> = {
  '/': 'Overview',
  '/skill-catalog': 'Skill Catalog',
  '/assessments': 'Assessments',
  '/candidate-review': 'Candidate Review',
};

export default function Topbar({ onMenu }: { onMenu: () => void }) {
  const path = usePathname();
  return (
    <header className="sticky top-0 z-10 flex h-12 items-center gap-3 border-b border-[#161B22] bg-[#0A0D12]/90 px-4 backdrop-blur md:px-8">
      <button
        onClick={onMenu}
        aria-label="Open menu"
        className="rounded-md p-1 text-gray-400 hover:text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-400 md:hidden"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
          <path d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>
      <p className="text-sm text-gray-200">{titles[path] ?? 'Praxis'}</p>
    </header>
  );
}