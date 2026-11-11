import PageHeader from '@/components/ui/PageHeader';
import PathList from '@/components/path/PathList';
import ProgressRadar from '@/components/overview/ProgressRadar';

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
    <div className="mx-auto max-w-4xl space-y-8">
      <PageHeader title="Overview" description="Your progress through the four steps." />

      <section>
        <h2 className="mb-3 text-base font-semibold text-white">Your path</h2>
        <PathList />
      </section>

      <section>
        <h2 className="mb-3 text-base font-semibold text-white">Skill progress</h2>
        <div className="rounded-lg border border-[#161B22] bg-[#0D1117] p-5">
          <ProgressRadar />
        </div>
      </section>

      <p className="flex items-center gap-2 text-xs text-gray-400">
        <span className={`h-2 w-2 rounded-full ${health ? 'bg-emerald-400' : 'bg-red-500'}`} />
        Backend: {health ? health.status : 'offline'}
      </p>
    </div>
  );
}