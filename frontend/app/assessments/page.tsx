// TODO(Chidi): replace with GET /assessments
const assessments = [
  { name: 'AI foundations & responsible use', area: 'AI fluency', status: 'Completed', score: 84 },
  { name: 'Evaluating evidence & assumptions', area: 'Critical thinking', status: 'Completed', score: 92 },
  { name: 'Interpreting data & visualizations', area: 'Data literacy', status: 'Completed', score: 67 },
  { name: 'Prompt engineering fundamentals', area: 'AI fluency', status: 'In progress', score: null },
  { name: 'Writing for different audiences', area: 'Communication', status: 'Not started', score: null },
];

const badge: Record<string, string> = {
  Completed: 'bg-emerald-500/10 text-emerald-400',
  'In progress': 'bg-amber-500/10 text-amber-400',
  'Not started': 'bg-gray-500/10 text-gray-400',
};

const mono = 'font-mono text-[10px] uppercase tracking-widest text-gray-500';

export default function Assessments() {
  return (
    <div className="mx-auto max-w-[900px] space-y-6">
      <div>
        <h1 className="text-[22px] font-semibold text-white">Assessments</h1>
        <p className="mt-1 text-[12px] text-gray-500">Track your results and pick up where you left off.</p>
      </div>

      <div className="rounded-lg border border-[#161B22] bg-[#0D1117]">
        <div className={`${mono} grid grid-cols-[2.4fr_1.2fr_1fr_0.8fr_1fr] px-4 py-2`}>
          <span>Assessment</span><span>Skill area</span><span>Status</span><span>Score</span><span />
        </div>
        {assessments.map((a) => (
          <div key={a.name} className="grid grid-cols-[2.4fr_1.2fr_1fr_0.8fr_1fr] items-center border-t border-[#161B22] px-4 py-3 text-[12px]">
            <span className="text-white">{a.name}</span>
            <span className="text-gray-400">{a.area}</span>
            <span><span className={`rounded px-2 py-0.5 text-[10px] ${badge[a.status]}`}>{a.status}</span></span>
            <span className="font-mono text-white">{a.score ?? '-'}</span>
            <span className="text-right">
              <button className="rounded-md border border-[#1C222B] px-2.5 py-1 text-[11px] text-gray-300 hover:text-white">
                {a.status === 'Completed' ? 'View report' : a.status === 'In progress' ? 'Continue' : 'Start'}
              </button>
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}