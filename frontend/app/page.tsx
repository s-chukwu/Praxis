import PageHeader from '@/components/ui/PageHeader';
import EmptyState from '@/components/ui/EmptyState';

type Health = { status: string; service: string; version: string };


async function getBackendHealth(): Promise<Health | null> {
  const base = process.env.API_INTERNAL_URL ?? 'http://localhost:8080';
  try {
    const res = await fetch(`${base}/health`, { cache: 'no-store' });
    if (!res.ok) return null;
    return (await res.json()) as Health;
  } catch {
    return null;
  }
}

export default async function Home() {
  const health = await getBackendHealth();
  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <PageHeader title="Overview" description="Your assessments and progress." />
      
      <EmptyState
        title="No assessments yet"
        description="Start your first one."
        actionHref="/skill-catalog"
        actionLabel="Browse skills"
      />
      <p className="flex items-center gap-2 text-xs text-gray-500">
        <span className={`h-2 w-2 rounded-full ${health ? 'bg-emerald-400' : 'bg-red-500'}`} />
        Backend: {health ? health.status : 'offline'}
      </p>
    </div>
  );
}