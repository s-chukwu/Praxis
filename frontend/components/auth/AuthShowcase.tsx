const steps = [
  ['CV and consent', 'Confirm skills suggested from your CV.'],
  ['Baseline task', 'A support scenario, 8 to 12 minutes.'],
  ['Learning plan', 'Courses ordered by your gaps.'],
  ['Final task', 'A new scenario, then see what improved.'],
];

const skills = [
  ['C1', 'Fact checking and comprehension'],
  ['C2', 'Communication'],
  ['C3', 'Workflow and escalation'],
  ['C4', 'Privacy and AI safety'],
  ['C5', 'AI verification'],
  ['C6', 'Triage and documentation'],
];

export default function AuthShowcase() {
  return (
    <div className="relative hidden flex-1 overflow-hidden border-l border-[#161B22] bg-[#080B0F] lg:flex lg:flex-col lg:justify-center lg:px-12">
      <div
        aria-hidden="true"
        className="absolute inset-0 bg-[radial-gradient(circle_at_30%_30%,rgba(37,99,235,0.2),transparent_60%)]"
      />
      <div
        aria-hidden="true"
        className="absolute inset-0 opacity-70 [background-image:linear-gradient(#161B22_1px,transparent_1px),linear-gradient(90deg,#161B22_1px,transparent_1px)] [background-size:48px_48px]"
      />

      <div className="relative">
        <h2 className="text-3xl font-semibold leading-tight text-white">Check your customer support skills</h2>
        <p className="mt-3 text-base text-gray-300">
          Take a baseline task, follow a learning plan, then take a final task to see what improved.
        </p>

        <div className="relative mt-8 grid grid-cols-[0.9fr_1.3fr] gap-6 rounded-xl border border-[#1C222B] bg-[#0A0D12] p-5">
          <div>
            <p className="mb-3 text-xs text-gray-400">Your path</p>
            <ol>
              {steps.map(([title, text], i) => (
                <li key={title} className="relative pb-5 pl-10 last:pb-0">
                  <span className="absolute left-0 top-0 flex h-7 w-7 items-center justify-center rounded-full border border-[#1C222B] bg-[#0A0D12] text-xs text-gray-300">
                    {i + 1}
                  </span>
                  {i < steps.length - 1 && (
                    <span aria-hidden="true" className="absolute bottom-0 left-3.5 top-7 w-px bg-[#1C222B]" />
                  )}
                  <p className="text-sm font-medium text-white">{title}</p>
                  <p className="mt-0.5 text-xs text-gray-400">{text}</p>
                </li>
              ))}
            </ol>
          </div>

          <div>
            <div className="mb-3 flex justify-between text-xs text-gray-400">
              <span>Skills measured</span>
              <span>Scored 0 to 4</span>
            </div>
            {skills.map(([code, name]) => (
              <div key={code} className="flex items-center justify-between gap-2 border-t border-[#161B22] py-2 text-xs text-gray-300">
                <span>
                  <span className="mr-2 rounded border border-[#1C222B] px-1.5 py-0.5">{code}</span>
                  {name}
                </span>
                <span className="flex gap-[3px]" aria-hidden="true">
                  {[0, 1, 2, 3, 4].map((n) => (
                    <i key={n} className="block h-2 w-4 rounded-sm border border-[#2A3340]" />
                  ))}
                </span>
              </div>
            ))}
            <p className="mt-3 text-xs text-gray-400">Builder track adds B1 to B3.</p>
          </div>

          <div className="absolute -bottom-11 -right-4 w-[300px] rounded-lg border border-[#1C222B] bg-[#0D1117] p-3.5 text-xs leading-relaxed text-gray-300">
            Reviewed by people. Results show your work on practice tasks and do not promise a job.
          </div>
        </div>
      </div>
    </div>
  );
}