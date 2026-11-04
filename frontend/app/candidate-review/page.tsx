const comps = [
  ['AI fluency', 92, 'Advanced'], ['Strategic decision making', 88, 'Advanced'], ['Data literacy', 76, 'Proficient'],
  ['Communication', 71, 'Developing'], ['Stakeholder alignment', 68, 'Developing'], ['Change management', 82, 'Proficient'],
] as const;

const tone: Record<string, string> = {
  Advanced: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  Proficient: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  Developing: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
};
const card = 'rounded-lg border border-[#161B22] bg-[#0D1117]';
const mono = 'font-mono text-[10px] uppercase tracking-widest text-gray-500';

export default function CandidateReview() {
  return (
    <div className="mx-auto max-w-[800px] space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <p className={mono}>Candidate review</p>
          <h1 className="mt-1 text-[22px] font-semibold text-white">Candidate Skill Verification</h1>
          <p className="mt-1 text-[12px] text-gray-500">Review the candidate profile, competency evidence, interview notes, and recommendation status before advancing the review.</p>
        </div>
        <div className="shrink-0 text-right">
          <p className={mono}>Review window · October 2026</p>
          <button className="mt-2 rounded-md bg-blue-600 px-3 py-1.5 text-[11px] text-white hover:bg-blue-500">Open full review</button>
        </div>
      </div>

      <section className={`${card} p-4`}>
        <div className="flex justify-between">
          <div>
            <p className="text-[12px] font-medium text-white">Candidate profile overview</p>
            <p className="text-[10px] text-gray-500">Example candidate · Illustrative review data only</p>
          </div>
          <span className="h-fit rounded bg-blue-500/10 px-2 py-0.5 text-[10px] text-blue-400">In review</span>
        </div>
        <div className="mt-4 grid grid-cols-[1.6fr_1fr] gap-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[#1C222B] text-[10px] text-white">EC</span>
              <div><p className="text-[12px] text-white">Elena Carter</p><p className="text-[10px] text-gray-500">Senior Product Strategist</p></div>
            </div>
            <div className="mt-3 grid grid-cols-3 gap-2 text-[11px]">
              {[['Source', 'Praxis AI talent pipeline'], ['Location', 'Remote · Pacific Time'], ['Availability', '2 weeks']].map(([k, v]) => (
                <div key={k}><p className={mono}>{k}</p><p className="text-gray-300">{v}</p></div>
              ))}
            </div>
          </div>
          <div className="space-y-2 text-right">
            {[['Overall score', '84%'], ['Interview rating', '88%'], ['Competencies reviewed', '6']].map(([k, v]) => (
              <div key={k}><p className={mono}>{k}</p><p className="text-[18px] font-medium text-white">{v}</p></div>
            ))}
          </div>
        </div>
        <div className="mt-4 grid grid-cols-3 gap-2">
          {[['Strongest fit', 'AI fluency and strategic decision making'], ['Growth area', 'Data storytelling and stakeholder alignment'], ['Next step', 'Final panel review and offer decision']].map(([k, v]) => (
            <div key={k} className="rounded-md border border-[#1C222B] bg-[#0A0D12] p-2.5">
              <p className={mono}>{k}</p><p className="mt-1 text-[11px] text-gray-300">{v}</p>
            </div>
          ))}
        </div>
      </section>

      <section className={`${card} p-4`}>
        <div className="mb-3 flex justify-between"><p className="text-[12px] font-medium text-white">Competency score breakdown</p><p className={mono}>6 competencies</p></div>
        <div className="space-y-3">
          {comps.map(([n, s, l]) => (
            <div key={n} className="grid grid-cols-[1.4fr_0.5fr_2fr_0.8fr] items-center gap-3 text-[11px]">
              <span className="text-gray-300">{n}</span>
              <span className="font-mono text-gray-300">{s}<span className="text-gray-600"> / 100</span></span>
              <div className="h-1 rounded bg-[#1C222B]"><div className="h-1 rounded bg-blue-500" style={{ width: `${s}%` }} /></div>
              <span className={`rounded border px-2 py-0.5 text-center text-[10px] ${tone[l]}`}>{l}</span>
            </div>
          ))}
        </div>
      </section>

      <section className={`${card} p-4`}>
        <div className="mb-3 flex justify-between"><p className="text-[12px] font-medium text-white">Interview notes</p><p className={mono}>2 sessions</p></div>
        <div className="space-y-2">
          {[['Technical session', 'Strong clarity on AI use cases, prompt design, and model risk. Candidate connected technical decisions to business outcomes effectively.'],
            ['Behavioral session', 'Candidate demonstrated strong problem framing and structured thinking. Stakeholder storytelling and alignment could be more polished in high-pressure scenarios.']].map(([k, v]) => (
            <div key={k} className="rounded-md border border-[#1C222B] bg-[#0A0D12] p-3">
              <p className={mono}>{k}</p><p className="mt-1 text-[11px] text-gray-300">{v}</p>
            </div>
          ))}
        </div>
        <div className="mt-3 flex items-center justify-between border-t border-[#161B22] pt-3">
          <div>
            <p className="text-[12px] text-white">Hiring recommendation status</p>
            <p className="text-[10px] text-gray-500">Candidate is currently under review. Final recommendation is pending panel feedback and offer package alignment.</p>
          </div>
          <span className="rounded bg-blue-500/10 px-2 py-0.5 text-[10px] text-blue-400">Under review</span>
        </div>
      </section>
    </div>
  );
}