import type { Metadata } from 'next';
import Navbar from '@/components/Navbar';
import Sidebar from '@/components/Sidebar';

export const metadata: Metadata = {
  title: 'Praxis - AI Workforce Platform',
  description: 'AI-driven workforce transition prototype',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-slate-950 text-slate-100 min-h-screen flex flex-col font-sans">
        <Navbar />
        <div className="flex flex-1">
          <Sidebar />
          <main className="flex-1 p-6 overflow-y-auto">{children}</main>
        </div>
      </body>
    </html>
  );
}