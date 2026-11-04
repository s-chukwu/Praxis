import Link from 'next/link';

type Health = { status: string; service: string; version: string };

// Teammate's backend health check - kept as is
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

const skills = [
  { area: 'AI & Emerging Tech', name: 'AI fluency', status: 'In progress', done: 3, pct: 75, cta: 'Continue assessment' },
  { area: 'Analytical Thinking', name: 'Data literacy', status: 'In progress', done: 2, pct: 50, cta: 'Continue assessment' },
  { area: 'Problem Solving', name: 'Critical thinking', status: 'Completed', done: 4, pct: 100, cta: 'View skill report' },
  { area: 'Human Skills', name: 'Communication', status: 'In progress', done: 1, pct: 25, cta: 'Continue assessment' },
];

const recent = [
  { name: 'AI foundations & responsible use', area: 'AI fluency', date: 'Oct 28, 2026', score: 84, level: 'Proficient' },
  { name: 'Evaluating evidence & assumptions', area: 'Critical thinking', date: 'Oct 26, 2026', score: 92, level: 'Advanced' },
  { name: 'Interpreting data & visualizations', area: 'Data literacy', date: 'Oct 24, 2026', score: 67, level: 'Developing' },
];

const badge: Record<string, string> = {
  'In progress': 'bg-amber-500/10 text-amber-400',
  Completed: 'bg-emerald-500/10 text-emerald-400',
  Proficient: 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20',
  Advanced: 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20',
  Developing: 'bg-amber-500/10 text-amber-400 border border-amber-500/20',
};

const card = 'rounded-lg border border-[#161B22] bg-[#0D1117]';
const mono = 'font-mono text-[10px] uppercase tracking-widest text-gray-500';

export default async function Home() {
  const health = await getBackendHealth();
  return (
    <div className="mx-auto max-w-[900px] space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <p className={mono}>Your growth, in focus</p>
          <h1 className="mt-1 text-[22px] font-semibold text-white">Welcome back, Learner</h1>
          <p className="mt-1 text-[12px] text-gray-500">Know where you stand. Build what comes next.</p>
        </div>
        <div className="text-right">
          <p className={mono}>Thursday, October 29, 2026</p>
          <Link href="/skill-catalog" className="mt-2 inline-block rounded-md bg-blue-600 px-3 py-1.5 text-[11px] font-medium text-white hover:bg-blue-500">
            + Explore assessments
          </Link>
        </div>
      </div>

      <section>
        <div className="mb-2 flex items-center justify-between">
          <p className="text-[12px] font-medium text-white">
            Establish your baseline <span className={`${mono} ml-1`}>4 skill areas</span>
          </p>
          <Link href="/skill-catalog" className="text-[11px] text-blue-400">View all skills ↗</Link>
        </div>
        <div className="grid grid-cols-2 gap-3">
          {skills.map((s) => (
            <div key={s.name} className={`${card} p-4`}>
              <div className="flex items-start justify-between">
                <span className="text-gray-500">▦</span>
                <span className={`rounded px-2 py-0.5 text-[10px] ${badge[s.status]}`}>{s.status}</span>
              </div>
              <p className={`${mono} mt-2`}>{s.area}</p>
              <p className="text-[14px] font-medium text-white">{s.name}</p>
              <div className="mt-3 flex justify-between text-[10px] text-gray-500">
                <span>{s.done} of 4 assessments</span>
                <span className="font-mono text-blue-400">{s.pct}%</span>
              </div>
              <div className="mt-1 h-1 rounded bg-[#1C222B]">
                <div className="h-1 rounded bg-blue-500" style={{ width: `${s.pct}%` }} />
              </div>
              <p className="mt-3 text-[11px] text-blue-400">{s.cta} →</p>
            </div>
          ))}
        </div>
      </section>

      <section>
        <p className="mb-2 text-[12px] font-medium text-white">Your assessment snapshot</p>
        <div className="grid grid-cols-3 gap-3">
          {[['Assessments completed', '12', '+4 vs. last month'], ['Average proficiency', '81%', '3 results recent assessments'], ['Skills benchmarked', '8', '']].map(([l, v, n]) => (
            <div key={l} className={`${card} p-4`}>
              <p className="text-[11px] text-gray-500">{l}</p>
              <p className="mt-1 text-[22px] font-medium text-white">{v}</p>
              {n && <p className="text-[10px] text-blue-400">{n}</p>}
            </div>
          ))}
        </div>
      </section>

      <section className="grid grid-cols-[1.6fr_1fr] gap-3">
        <div className={`${card} p-4`}>
          <p className="text-[11px] text-gray-400">
            <span className="font-medium text-white">Skill growth</span> Average proficiency score
          </p>
          <svg viewBox="0 0 400 120" className="mt-3 w-full">
            <line x1="0" x2="400" y1="45" y2="45" stroke="#2A3340" strokeDasharray="3 3" />
            <polyline fill="none" stroke="#3B82F6" strokeWidth="2"
              points="0,100 40,92 80,88 110,70 150,72 190,55 230,50 270,35 310,38 350,22 400,15" />
          </svg>
          <div className="mt-1 flex justify-between font-mono text-[9px] text-gray-600">
            {['Sep 18', 'Sep 25', 'Oct 02', 'Oct 09', 'Oct 16', 'Oct 23'].map((d) => <span key={d}>{d}</span>)}
          </div>
        </div>
        <div className={`${card} p-4`}>
          <p className="font-mono text-[10px] uppercase tracking-widest text-blue-400">Praxis Intelligence</p>
          <p className="mt-2 text-[13px] font-medium text-white">Turn AI fluency into your next advantage.</p>
          <p className="mt-1 text-[11px] leading-snug text-gray-500">
            You&apos;re in the top 24% for critical thinking. Pair it with prompt engineering to strengthen your AI skills.
          </p>
          <div className="mt-3 rounded-md border border-[#1C222B] p-3">
            <p className="text-[11px] text-gray-500">Learning time invested</p>
            <p className="font-mono text-[20px] text-white">6.4h</p>
          </div>
        </div>
      </section>

      <section>
        <div className="mb-2 flex items-center justify-between">
          <p className="text-[12px] font-medium text-white">Recent assessments</p>
          <Link href="/assessments" className="text-[11px] text-blue-400">View assessment history ↗</Link>
        </div>
        <div className={card}>
          <div className={`${mono} grid grid-cols-[2.4fr_1fr_1fr_0.8fr_1fr] px-4 py-2`}>
            <span>Assessment</span><span>Skill area</span><span>Completed</span><span>Score</span><span>Proficiency</span>
          </div>
          {recent.map((r) => (
            <div key={r.name} className="grid grid-cols-[2.4fr_1fr_1fr_0.8fr_1fr] items-center border-t border-[#161B22] px-4 py-3 text-[12px] text-gray-300">
              <span className="text-white">{r.name}</span>
              <span className="text-gray-400">{r.area}</span>
              <span className="font-mono text-[11px] text-gray-500">{r.date}</span>
              <span className="font-mono text-white">{r.score}<span className="text-gray-600"> / 100</span></span>
              <span><span className={`rounded px-2 py-0.5 text-[10px] ${badge[r.level]}`}>{r.level}</span></span>
            </div>
          ))}
        </div>
      </section>

      <p className="flex items-center justify-between text-[10px] text-gray-600">
        <span className="flex items-center gap-1.5">
          <span className={`h-1.5 w-1.5 rounded-full ${health ? 'bg-emerald-400' : 'bg-red-500'}`} />
          Backend: {health ? health.status : 'offline'}
        </span>
        <span>Your growth. Your data. Always private.</span>
      </p>
    </div>
  );
}