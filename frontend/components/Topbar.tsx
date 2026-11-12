'use client';
import { Menu } from 'lucide-react';
import Logo from './Logo';

// Desktop pages carry their own headings; the bar only exists to open the menu on small screens.
export default function Topbar({ onMenu }: { onMenu: () => void }) {
  return (
    <header className="sticky top-0 z-10 flex h-14 items-center gap-3 border-b border-line bg-surface/95 px-4 backdrop-blur md:hidden">
      <button onClick={onMenu} aria-label="Open menu" className="rounded-control p-2 text-ink hover:bg-page">
        <Menu className="h-5 w-5" aria-hidden />
      </button>
      <Logo />
    </header>
  );
}
