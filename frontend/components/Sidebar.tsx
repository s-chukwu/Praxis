
'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import Logo from './Logo';

const workspace = [
  { label: 'Overview', href: '/' },
  { label: 'Skill Catalog', href: '/skill-catalog' },
  { label: 'Practical Sandbox', href: '/sandbox' },
  { label: 'AI Compare', href: '/ai-compare' },
  { label: 'Talent Matching', href: '/candidate-review' },
];
const organization = [
  { label: 'Team insights', href: '/team-insights' },
  { label: 'Reports', href: '/reports' },
];

function Icon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
      <rect x="4" y="4" width="7" height="7" rx="1.5" />
      <rect x="13" y="4" width="7" height="7" rx="1.5" />
      <rect x="4" y="13" width="7" height="7" rx="1.5" />
      <rect x="13" y="13" width="7" height="7" rx="1.5" />
    </svg>
  );
}

function NavItem({ label, href, active }: { label: string; href: string; active: boolean }) {
  return (
    <Link
      href={href}
      className={`flex items-center gap-2.5 rounded-md border px-2.5 py-2 text-[13px] transition-colors ${
        active
          ? 'border-[#1C222B] bg-[#11151B] text-white'
          : 'border-transparent text-gray-400 hover:bg-[#0F1319] hover:text-white'
      }`}
    >
      <Icon />
      {label}
    </Link>
  );
}

export default function Sidebar() {
  const path = usePathname();
  const isActive = (href: string) => (href === '/' ? path === '/' : path.startsWith(href));
  return (

    <aside className="sticky top-0 flex h-screen w-[220px] shrink-0 flex-col border-r border-[#161B22] bg-[#080B0F] p-3">
      <div className="px-2 pb-4 pt-2"><Logo /></div>

      <button className="mb-5 flex items-center gap-2 rounded-md border border-[#1C222B] bg-[#0D1117] px-2 py-2 text-left">
        <span className="flex h-5 w-5 items-center justify-center rounded bg-[#1C222B] text-[10px] text-white">A</span>
        <span className="flex-1 leading-tight">
          <span className="block text-[12px] font-medium text-white">Praxis Engine</span>
          <span className="block text-[10px] text-gray-500">Enterprise plan</span>
        </span>
        <span className="text-xs text-gray-500">⌄</span>
      </button>

      <p className="mb-1.5 px-2 font-mono text-[10px] uppercase tracking-widest text-gray-600">Workspace</p>
      <nav className="space-y-1">
        {workspace.map((n) => <NavItem key={n.label} {...n} active={isActive(n.href)} />)}
      </nav>

      <p className="mb-1.5 mt-5 px-2 font-mono text-[10px] uppercase tracking-widest text-gray-600">Organization</p>
      <nav className="space-y-1">
        {organization.map((n) => <NavItem key={n.label} {...n} active={isActive(n.href)} />)}
      </nav>

      <div className="mt-auto space-y-3">
        <div className="rounded-lg border border-[#1C222B] bg-[#0D1117] p-3">
          <p className="text-[12px] font-medium text-white">Your next breakthrough</p>
          <p className="mt-1 text-[11px] leading-snug text-gray-500">
            A clearer picture of your skills starts with a baseline.
          </p>
          <Link href="/skill-catalog" className="mt-2 inline-block text-[11px] text-gray-300 hover:text-white">
            Explore assessments ↗
          </Link>
        </div>
        <div className="space-y-1">
          <NavItem label="Settings" href="/settings" active={false} />
          <NavItem label="Help & resources" href="/help" active={false} />
        </div>
        <div className="flex items-center gap-2 border-t border-[#161B22] px-1 pt-3">
          <span className="flex h-6 w-6 items-center justify-center rounded bg-[#1C222B] text-[10px] text-white">JL</span>
          <span className="flex-1 leading-tight">
            <span className="block text-[12px] text-white">Jamie Lee</span>
            <span className="block text-[10px] text-gray-500">Product designer</span>
          </span>
        </div>
      </div>
    </aside>
  );
}