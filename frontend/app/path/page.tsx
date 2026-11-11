import Link from 'next/link';
import PageHeader from '@/components/ui/PageHeader';

const steps = [
  { title: 'CV and consent', text: 'Confirm the skills suggested from your CV.', href: '/path/cv' },
  { title: 'Baseline task', text: 'A customer support scenario.', href: '/path/baseline' },
  { title: 'Learning plan', text: 'Courses ordered by your gaps.', href: '/path/plan' },
  { title: 'Final task', text: 'A different scenario, then see what improved.' },
];

export default function MyPath() {
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <PageHeader title="My path" description="Four steps, in order." />
      <ol className="divide-y divide-[#161B22] rounded-lg border border-[#161B22] bg-[#0D1117]">
        {steps.map((s, i) => (
          <li key={s.title} className="flex items-center gap-4 p-4">
            <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-[#1C222B] text-xs text-gray-300">
              {i + 1}
            </span>
            <div className="flex-1">
              <p className="text-sm font-medium text-white">{s.title}</p>
              <p className="mt-0.5 text-xs text-gray-400">{s.text}</p>
            </div>
            {s.href ? (
              <Link href={s.href} className="text-sm text-blue-400 hover:underline">
                Open
              </Link>
            ) : (
              <span className="text-xs text-gray-400">Not available yet</span>
            )}
          </li>
        ))}
      </ol>
    </div>
  );
}