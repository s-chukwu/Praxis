import type { Learner, SkillStatus } from '@/lib/learner';
import { programs, skillNames } from '@/lib/catalog';

const label: Record<SkillStatus | 'none', { text: string; tone: string }> = {
  none: { text: 'Not assessed', tone: 'bg-page text-muted' },
  rejected: { text: 'Not assessed', tone: 'bg-page text-muted' },
  suggested: { text: 'Suggested', tone: 'bg-progress-soft text-[#7A4E09]' },
  confirmed: { text: 'Your claim', tone: 'bg-progress-soft text-[#7A4E09]' },
  verified: { text: 'Verified', tone: 'bg-brand-soft text-brand-ink' },
};

export default function SkillProgress({ learner }: { learner: Learner }) {
  const codes = programs[learner.program ?? 'operator'].skills;
  const status = (code: string) => learner.skills[code] ?? 'none';
  const verified = codes.filter((c) => status(c) === 'verified').length;
  const claimed = codes.filter((c) => status(c) === 'confirmed').length;

  return (
    <section aria-labelledby="skills-title" className="rounded-panel bg-surface p-6 shadow-panel">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h2 id="skills-title" className="text-lg font-bold">
          Your skills
        </h2>
        <p className="text-sm text-muted">
          <span className="font-semibold text-ink">{verified}</span> of {codes.length} verified
          {claimed > 0 && `, ${claimed} claimed from your CV`}
        </p>
      </div>

      <div
        className="mt-4 flex h-2.5 overflow-hidden rounded-full bg-page"
        role="img"
        aria-label={`${verified} verified, ${claimed} claimed, ${codes.length - verified - claimed} not assessed`}
      >
        <span className="bg-brand transition-[width] duration-700 ease-out" style={{ width: `${(verified / codes.length) * 100}%` }} />
        <span className="bg-progress/60 transition-[width] duration-700 ease-out" style={{ width: `${(claimed / codes.length) * 100}%` }} />
      </div>

      <ul className="mt-5 divide-y divide-line">
        {codes.map((code) => {
          const s = label[status(code)];
          return (
            <li key={code} className="flex items-center justify-between gap-4 py-3">
              <span className="min-w-0 truncate">{skillNames[code]}</span>
              <span className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold ${s.tone}`}>{s.text}</span>
            </li>
          );
        })}
      </ul>
      <p className="mt-4 text-[13px] text-muted">Skills become verified when a reviewer scores your baseline task.</p>
    </section>
  );
}
