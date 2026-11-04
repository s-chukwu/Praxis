import type { Metadata } from 'next';
import './globals.css';
import AppShell from '@/components/AppShell';

export const metadata: Metadata = {
  title: 'Praxis - AI Workforce Platform',
  description: 'AI-driven workforce transition prototype',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="dark">
      <body className="min-h-screen bg-[#0A0D12] text-gray-100 antialiased">
        <AppShell>{children}</AppShell>
      </body>
    </html>
  );
}