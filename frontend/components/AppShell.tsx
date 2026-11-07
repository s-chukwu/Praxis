'use client';
import { usePathname } from 'next/navigation';
import Sidebar from './Sidebar';
import Topbar from './Topbar';

const NO_SHELL = ['/login', '/signup'];

export default function AppShell({ children }: { children: React.ReactNode }) {
  const path = usePathname();
  if (NO_SHELL.includes(path)) return <>{children}</>;
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <Topbar />
        <main className="flex-1 px-8 py-6">{children}</main>
      </div>
    </div>
  );
}