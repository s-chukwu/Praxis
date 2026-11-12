'use client';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Logo from '@/components/Logo';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import ProgramDialog from '@/components/onboarding/ProgramDialog';
import CvConsent from '@/components/path/CvConsent';
import { saveLearner } from '@/lib/learner';
import type { ProgramId } from '@/lib/catalog';

const STEPS = ['Welcome', 'Program', 'Your CV'];

const ideas = [
  ['Show what you can do', 'Start with one real support scenario. No quiz, no trick questions.'],
  ['A person reviews it', 'A trained supervisor scores your work and explains the score.'],
  ['Learn only what you need', 'Your plan skips what you already do well, then a final task shows what improved.'],
];

export default function Onboarding() {
  const router = useRouter();
  const { learner, refresh } = useLearner();
  const [step, setStep] = useState(learner?.program ? 2 : 0);
  const firstName = learner?.name.split(' ')[0];

  const chooseProgram = async (program: ProgramId) => {
    await saveLearner({ program });
    await refresh();
    setStep(2);
  };

  const finish = async () => {
    await saveLearner({ onboarded: true });
    await refresh();
    router.replace('/dashboard');
  };

  return (
    <div className="min-h-screen bg-page">
      <header className="border-b border-line bg-surface">
        <div className="mx-auto flex h-16 max-w-3xl items-center justify-between gap-6 px-5">
          <Logo />
          <ol className="flex items-center gap-2 text-[13px] font-semibold" aria-label="Setup progress">
            {STEPS.map((label, i) => (
              <li key={label} aria-current={i === step ? 'step' : undefined} className="flex items-center gap-2">
                <span
                  className={`flex h-6 w-6 items-center justify-center rounded-full text-xs ${
                    i < step ? 'bg-brand text-white' : i === step ? 'bg-ink text-white' : 'bg-line text-muted'
                  }`}
                >
                  {i + 1}
                </span>
                <span className={`hidden sm:inline ${i === step ? 'text-ink' : 'text-muted'}`}>{label}</span>
                {i < STEPS.length - 1 && <span aria-hidden className="mx-1 hidden h-px w-6 bg-line sm:block" />}
              </li>
            ))}
          </ol>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-5 py-12 sm:py-16">
        {step < 2 ? (
          <section>
            <h1 className="text-[34px] font-bold leading-tight sm:text-[40px]">
              {firstName ? `Welcome, ${firstName}.` : 'Welcome.'} Here is how it works.
            </h1>
            <ol className="mt-10 divide-y divide-line rounded-panel bg-surface shadow-panel">
              {ideas.map(([title, text], i) => (
                <li key={title} className="flex gap-5 p-6">
                  <span className="text-[28px] font-bold leading-none text-brand">{i + 1}</span>
                  <div>
                    <p className="text-lg font-bold">{title}</p>
                    <p className="mt-1 text-muted">{text}</p>
                  </div>
                </li>
              ))}
            </ol>
            <Button size="lg" className="mt-10" onClick={() => setStep(1)}>
              Choose my program
            </Button>
            {step === 1 && <ProgramDialog onChoose={chooseProgram} />}
          </section>
        ) : (
          <section>
            <h1 className="text-[34px] font-bold leading-tight">Add your CV</h1>
            <p className="mt-2 max-w-xl text-muted">
              We suggest skills from your CV and you confirm them. Your baseline task still decides your plan.
            </p>
            <div className="mt-8">
              <CvConsent onDone={finish} />
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
