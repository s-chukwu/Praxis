import type { Metadata } from 'next';
import '@fontsource-variable/figtree';
import './globals.css';
import AppShell from '@/components/AppShell';

export const metadata: Metadata = {
  title: { default: 'Praxis', template: '%s · Praxis' },
  description: 'Training for AI-assisted customer support.',
};

const CONTRACT = `<!--
THESIS: A learning platform where your verified progress is the first thing you see; refuses the course-grid storefront.
OWN-WORLD: White surfaces on #F7F8FA, ink #14171F, jade #0F7B5F for action and verified state, amber #C98417 for in progress. Figtree, 10px controls, 14px panels.
STORY: The learner sees what is verified, what is next, and why each course is in their plan, then starts the next step.
FIRST VIEWPORT: Dashboard opens on the skills progress panel beside one primary next-step action; the path runs below.
FORM: Category standard (canon), quality bar Duolingo clarity and Linear precision; seed e3a372d1.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->`;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div hidden dangerouslySetInnerHTML={{ __html: CONTRACT }} />
        <AppShell>{children}</AppShell>
      </body>
    </html>
  );
}
