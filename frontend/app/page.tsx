import Link from 'next/link';
import { ArrowRight, BadgeCheck, Bot, Check, Clock, Headset, ShieldCheck, Sparkles, type LucideIcon } from 'lucide-react';
import Logo from '@/components/Logo';
import { ButtonLink } from '@/components/ui/Button';
import { programs, skillNames, type ProgramId } from '@/lib/catalog';

export const metadata = { title: 'Prove what you can do' };

const steps = [
  ['Start with a real case', 'Reply to a customer support scenario. It takes 8 to 12 minutes.'],
  ['A person reviews your work', 'A trained reviewer scores your answers and explains what to improve.'],
  ['Get a plan for your gaps', 'Your plan skips what you already do well and focuses on what you need.'],
  ['Show what changed', 'A final task shows your progress. You leave with evidence, not just a certificate.'],
];

const programDetails: Record<ProgramId, { icon: LucideIcon; outcomes: string[] }> = {
  operator: {
    icon: Headset,
    outcomes: [
      'Write clear, empathetic customer replies',
      'Check AI drafts before they reach a customer',
      'Triage, escalate, and document cases',
      'Protect customer data and spot prompt injection',
    ],
  },
  builder: {
    icon: Bot,
    outcomes: [
      'Build a bounded FAQ chatbot',
      'Design safe human handoff workflows',
      'Test and improve chatbot behaviour',
      'Everything in the customer support agent program',
    ],
  },
};

export default function Home() {
  return (
    <div className="min-h-screen bg-page">
      <header className="border-b border-line bg-surface">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-6 px-5 sm:px-8">
          <Logo />
          <nav className="flex items-center gap-2" aria-label="Account">
            <Link href="/login" className="rounded-control px-3 py-2 text-sm font-semibold text-ink hover:bg-page">
              Sign in
            </Link>
            <ButtonLink href="/signup">Create account</ButtonLink>
          </nav>
        </div>
      </header>

      <main>
        <section className="mx-auto grid max-w-6xl gap-10 px-5 py-16 sm:px-8 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)] lg:items-center lg:py-24">
          <div>
            <p className="inline-flex items-center gap-2 rounded-full bg-brand-soft px-3 py-1 text-[13px] font-semibold text-brand-ink">
              <Sparkles className="h-4 w-4" aria-hidden />
              Training for AI-assisted customer support
            </p>
            <h1 className="mt-5 max-w-2xl text-[42px] font-bold leading-[1.08] tracking-tight sm:text-[56px]">
              Prove what you can do. Skip what you already know.
            </h1>
            <p className="mt-5 max-w-xl text-lg text-muted">
              Praxis starts with a real support scenario, not a quiz. A person reviews your work, then you get a short plan for the skills that still need practice.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <ButtonLink href="/signup" size="lg">
                Start with a baseline task
                <ArrowRight className="h-4 w-4" aria-hidden />
              </ButtonLink>
              <ButtonLink href="/login" variant="secondary" size="lg">
                Sign in
              </ButtonLink>
            </div>
            <p className="mt-4 flex items-center gap-2 text-sm text-muted">
              <BadgeCheck className="h-4 w-4 text-brand" aria-hidden />
              No quiz. No trick questions. Your work is reviewed by a person.
            </p>
          </div>

          <div className="rounded-[24px] bg-brand p-7 text-white shadow-lift sm:p-9">
            <p className="text-sm font-semibold text-white/70">How it works</p>
            <ol className="mt-6 space-y-6">
              {steps.map(([title, text], i) => (
                <li key={title} className="flex gap-4">
                  <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white text-sm font-bold text-brand-ink">
                    {i + 1}
                  </span>
                  <div>
                    <p className="font-semibold">{title}</p>
                    <p className="mt-1 text-sm text-white/75">{text}</p>
                  </div>
                </li>
              ))}
            </ol>
            <div className="mt-8 flex items-center gap-2 border-t border-white/15 pt-5 text-sm text-white/80">
              <Clock className="h-4 w-4" aria-hidden />
              Baseline task: 8 to 12 minutes
            </div>
          </div>
        </section>

        <section className="border-y border-line bg-surface">
          <div className="mx-auto max-w-6xl px-5 py-16 sm:px-8">
            <div className="max-w-2xl">
              <p className="text-sm font-semibold text-brand-ink">Two programs</p>
              <h2 className="mt-2 text-[32px] font-bold leading-tight">Choose the job you are training for.</h2>
              <p className="mt-3 text-muted">
                Both programs start with the same baseline task. You choose your program after you create an account.
              </p>
            </div>

            <div className="mt-8 grid gap-5 lg:grid-cols-2">
              {(Object.keys(programs) as ProgramId[]).map((id) => {
                const program = programs[id];
                const detail = programDetails[id];
                const Icon = detail.icon;
                return (
                  <article key={id} className="flex flex-col rounded-panel border border-line bg-page p-6 sm:p-7">
                    <div className="flex items-start gap-4">
                      <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-control bg-brand text-white">
                        <Icon className="h-5 w-5" aria-hidden />
                      </span>
                      <div>
                        <h3 className="text-xl font-bold">{program.name}</h3>
                        <p className="mt-1 text-sm text-muted">{program.summary}</p>
                      </div>
                    </div>

                    <ul className="mt-6 space-y-2.5">
                      {detail.outcomes.map((outcome) => (
                        <li key={outcome} className="flex gap-2.5 text-sm">
                          <Check className="mt-0.5 h-4 w-4 shrink-0 text-brand" aria-hidden />
                          <span>{outcome}</span>
                        </li>
                      ))}
                    </ul>

                    <div className="mt-6 border-t border-line pt-5">
                      <p className="text-[13px] font-semibold text-ink">Skills assessed</p>
                      <ul className="mt-2 flex flex-wrap gap-1.5">
                        {program.skills.map((skill) => (
                          <li key={skill} className="rounded-full bg-surface px-2.5 py-1 text-[12px] font-medium text-muted">
                            {skillNames[skill]}
                          </li>
                        ))}
                      </ul>
                    </div>

                    <p className="mt-5 text-[13px] font-semibold text-brand-ink">{program.skills.length} skills assessed</p>
                  </article>
                );
              })}
            </div>
          </div>
        </section>

        <section className="mx-auto max-w-6xl px-5 py-16 sm:px-8">
          <div className="grid gap-8 lg:grid-cols-3">
            <div>
              <p className="text-sm font-semibold text-brand-ink">Why Praxis</p>
              <h2 className="mt-2 text-[32px] font-bold leading-tight">Evidence, not just a certificate.</h2>
            </div>
            <div className="rounded-panel bg-surface p-6 shadow-panel">
              <ShieldCheck className="h-5 w-5 text-brand" aria-hidden />
              <p className="mt-4 font-bold">Human review</p>
              <p className="mt-1 text-sm text-muted">A trained reviewer checks your work and explains the score. AI helps, but a person decides.</p>
            </div>
            <div className="rounded-panel bg-surface p-6 shadow-panel">
              <BadgeCheck className="h-5 w-5 text-brand" aria-hidden />
              <p className="mt-4 font-bold">A plan for your gaps</p>
              <p className="mt-1 text-sm text-muted">You do not sit through courses you already know. Your plan focuses on what the baseline shows you need.</p>
            </div>
          </div>
        </section>

        <section className="mx-auto max-w-6xl px-5 pb-20 sm:px-8">
          <div className="rounded-[24px] bg-ink px-6 py-10 text-white sm:px-10 sm:py-12">
            <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
              <div className="max-w-xl">
                <h2 className="text-[30px] font-bold leading-tight">Start with one real customer case.</h2>
                <p className="mt-2 text-white/75">Create an account, choose your program, and show what you can already do.</p>
              </div>
              <ButtonLink href="/signup" size="lg" className="shrink-0">
                Create your account
                <ArrowRight className="h-4 w-4" aria-hidden />
              </ButtonLink>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t border-line bg-surface">
        <div className="mx-auto flex max-w-6xl flex-col gap-3 px-5 py-8 text-sm text-muted sm:flex-row sm:items-center sm:justify-between sm:px-8">
          <Logo />
          <p>Results are checked by people and show your work on practice tasks. They are not a job offer.</p>
        </div>
      </footer>
    </div>
  );
}
