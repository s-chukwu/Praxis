'use client';
import { useRouter, usePathname } from 'next/navigation';

export default function Topbar() {
  const router = useRouter();
  const path = usePathname();
  const employer = path.startsWith('/candidate-review');
  const titles: Record<string, string> = {
  '/': 'Overview',
  '/skill-catalog': 'Skill Catalog',
  '/assessments': 'Assessments',
  '/candidate-review': 'Candidate Verification',
  '/sandbox': 'Practical Sandbox',
  '/ai-compare': 'AI Compare',
  '/team-insights': 'Team insights',
  '/reports': 'Reports',
  '/settings': 'Settings',
  '/help': 'Help & resources',
};
const crumb = titles[path] ?? 'Overview';

  const tab = (label: string, on: boolean, to: string) => (
    <button
      onClick={() => router.push(to)}
      className={`rounded px-2.5 py-1 text-[11px] ${on ? 'bg-[#1C222B] text-white' : 'text-gray-500 hover:text-white'}`}
    >
      {label}
    </button>
  );

  return (
    <header className="sticky top-0 z-10 flex h-12 items-center justify-between border-b border-[#161B22] bg-[#0A0D12]/90 px-8 backdrop-blur">
      <p className="text-[12px] text-gray-500">
        Workspace <span className="mx-1 text-gray-700">/</span>
        <span className="text-gray-200">{crumb}</span>
      </p>
      <div className="flex items-center gap-4">
        <div className="flex w-56 items-center justify-between rounded-md border border-[#1C222B] bg-[#0D1117] px-2.5 py-1.5 text-[11px] text-gray-500">
          <span>Search anything...</span>
          <kbd className="rounded bg-[#1C222B] px-1 font-mono text-[9px]">⌘ K</kbd>
        </div>
        <span className="text-gray-500">🔔︎</span>
        <span className="flex h-5 w-5 items-center justify-center rounded bg-[#1C222B] text-[9px] text-white">JL</span>
        <div className="flex rounded-md border border-[#1C222B] bg-[#0D1117] p-0.5">
          {tab('Learner', !employer, '/')}
          {tab('Employer', employer, '/candidate-review')}
        </div>
      </div>
    </header>
  );
}