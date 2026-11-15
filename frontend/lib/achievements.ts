import type { Learner } from './learner';
import type { Progress } from './baseline';
import { programs } from './catalog';
import { milestones } from './journey';
import { PREVIEW } from './learner';

export type MilestoneId = 'baseline-verified' | 'plan-completed';

export type Milestone = {
  id: MilestoneId;
  title: string;
  message: string;
  actionLabel: string;
 
  actionHref?: string;
};

export type MilestoneState = Milestone & { unlocked: boolean };


export const PREVIEW_PROGRESS_KEY = 'praxis.achievement.preview';

export function withPreviewProgress(progress: Progress | null): Progress | null {
  if (!PREVIEW || typeof window === 'undefined') return progress;
  try {
    const extra = JSON.parse(localStorage.getItem(PREVIEW_PROGRESS_KEY) ?? 'null');
    return extra ? { baseline: null, result: null, plan: null, ...progress, ...extra } : progress;
  } catch {
    return progress;
  }
}

export function trackName(learner: Learner) {
  return programs[learner.program ?? 'operator'].name;
}


export function getMilestones(learner: Learner, progress: Progress | null): MilestoneState[] {
  const track = trackName(learner);
  const earned = (id: string) => milestones(progress).some((m) => m.id === id && m.state === 'earned');
  // Every course in the plan has a completed_at date. lib/journey.ts does not track this yet, so read it directly.
  const planFinished = Boolean(progress?.plan && progress.plan.modules.length > 0 && progress.plan.modules.every((m) => m.completed_at));

  return [
    {
      id: 'baseline-verified',
      title: 'Baseline reviewed',
      message: `A supervisor has scored your ${track} baseline task. Your result and the courses chosen for you are ready.`,
      actionLabel: 'See my result and plan',
      actionHref: '/path/plan',
      unlocked: earned('baseline') && earned('reviewed'),
    },
    {
      id: 'plan-completed',
      title: 'Learning Plan Completed',
      message: `You finished every course in your ${track} learning plan. The final task is next.`,
      actionLabel: 'Continue',
      unlocked: earned('plan') || planFinished,
    },
  ];
}