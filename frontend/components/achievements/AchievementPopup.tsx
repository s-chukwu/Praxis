'use client';
import { useEffect, useMemo, useRef, useState } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { useLearner } from '@/components/LearnerProvider';
import { PREVIEW } from '@/lib/learner';
import { supabase } from '@/lib/supabase';
import { useProgress } from '@/lib/useProgress';
import { getMilestones, withPreviewProgress, type MilestoneId } from '@/lib/achievements';

const seenKey = (uid: string, program: string, id: MilestoneId) => `praxis.achievement.seen.${uid}.${program}.${id}`;

function hasSeen(key: string) {
  try {
    return localStorage.getItem(key) === '1';
  } catch {
    return false;
  }
}

function markSeen(key: string) {
  try {
    localStorage.setItem(key, '1');
  } catch {
    // Storage is blocked. The pop-up may show again on the next visit, which is harmless.
  }
}

export default function AchievementPopup() {
  const { learner } = useLearner();
  const real = useProgress();
  const progress = useMemo(() => withPreviewProgress(real), [real]);
  const router = useRouter();
  const path = usePathname();
  const [uid, setUid] = useState<string | null>(null);
  const [activeId, setActiveId] = useState<MilestoneId | null>(null);
  const [shown, setShown] = useState(false);
  const button = useRef<HTMLButtonElement>(null);

  // The learner object has no user id, so read it from the session the same way lib/learner.ts does.
  useEffect(() => {
    let live = true;
    (async () => {
      if (PREVIEW) return live && setUid('preview');
      const { data } = (await supabase?.auth.getSession()) ?? { data: { session: null } };
      if (live) setUid(data.session?.user.id ?? null);
    })();
    return () => {
      live = false;
    };
  }, []);

  // Find the first unlocked milestone this learner has not seen. After a dismiss, this finds the next one.
  useEffect(() => {
    if (!learner || !uid || activeId) return;
    const program = learner.program ?? 'operator';
    const next = getMilestones(learner, progress).find((m) => m.unlocked && !hasSeen(seenKey(uid, program, m.id)));
    if (next) {
      markSeen(seenKey(uid, program, next.id));
      setActiveId(next.id);
    }
  }, [learner, progress, uid, activeId]);

  const milestone = learner && activeId ? getMilestones(learner, progress).find((m) => m.id === activeId) : undefined;

  useEffect(() => {
    if (!milestone) return setShown(false);
    const frame = requestAnimationFrame(() => setShown(true));
    button.current?.focus();
    document.body.style.overflow = 'hidden';
    return () => {
      cancelAnimationFrame(frame);
      document.body.style.overflow = '';
    };
  }, [milestone]);

  if (!milestone) return null;

  const dismiss = () => setActiveId(null);
  const proceed = () => {
    dismiss();
    if (milestone.actionHref && milestone.actionHref !== path) router.push(milestone.actionHref);
  };

  return (
    <div
      className={`fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4 backdrop-blur-[1px] transition-opacity duration-200 ${shown ? 'opacity-100' : 'opacity-0'}`}
      onClick={dismiss}
      onKeyDown={(e) => {
        if (e.key === 'Escape') dismiss();
        if (e.key === 'Tab') {
          e.preventDefault();
          button.current?.focus();
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="achievement-heading"
        aria-describedby="achievement-message"
        onClick={(e) => e.stopPropagation()}
        className={`w-full max-w-md overflow-hidden rounded-2xl bg-white shadow-xl transition-all duration-300 ${shown ? 'translate-y-0 scale-100' : 'translate-y-2 scale-95'}`}
      >
        <div className="relative">
          <div className="relative h-44 overflow-hidden rounded-b-[50%_44px] bg-gradient-to-br from-[#0F766E] to-[#065F46]">
            <Confetti />
          </div>
          <div className="absolute left-1/2 top-[128px] -translate-x-1/2">
            <Rosette />
          </div>
        </div>

        <div className="px-8 pb-8 pt-24 text-center">
          <h2 id="achievement-heading" className="text-2xl font-bold text-slate-900">
            Congratulations!
          </h2>
          <p className="mt-1 text-lg font-semibold text-slate-800">{milestone.title}</p>
          <p id="achievement-message" className="mt-3 text-sm leading-relaxed text-slate-600">
            {milestone.message}
          </p>
          <button
            ref={button}
            type="button"
            onClick={proceed}
            className="mt-6 rounded-lg bg-[#0F766E] px-6 py-2.5 font-medium text-white hover:bg-[#065F46]"
          >
            {milestone.actionLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

function Confetti() {
  return (
    <svg aria-hidden viewBox="0 0 400 176" preserveAspectRatio="xMidYMid slice" className="absolute inset-0 h-full w-full">
      <rect x="46" y="30" width="10" height="10" rx="1.5" fill="#FBBF24" transform="rotate(18 51 35)" />
      <rect x="332" y="26" width="9" height="9" rx="1.5" fill="#F59E0B" transform="rotate(-24 336 30)" />
      <rect x="92" y="96" width="7" height="7" rx="1" fill="#F59E0B" transform="rotate(32 95 99)" />
      <rect x="300" y="92" width="8" height="8" rx="1" fill="#FBBF24" transform="rotate(10 304 96)" />
      <polygon points="150,22 160,40 140,40" fill="#FBBF24" transform="rotate(14 150 32)" />
      <polygon points="262,34 271,50 253,50" fill="#F59E0B" transform="rotate(-18 262 44)" />
      <polygon points="30,108 38,122 22,122" fill="#FBBF24" transform="rotate(-10 30 116)" />
      <polygon points="372,108 380,122 364,122" fill="#F59E0B" transform="rotate(20 372 116)" />
      <circle cx="116" cy="46" r="3" fill="#FFFFFF" fillOpacity="0.85" />
      <circle cx="210" cy="20" r="2.5" fill="#FFFFFF" fillOpacity="0.8" />
      <circle cx="360" cy="64" r="3" fill="#FFFFFF" fillOpacity="0.85" />
      <circle cx="64" cy="70" r="2.5" fill="#FFFFFF" fillOpacity="0.8" />
      <circle cx="326" cy="140" r="2.5" fill="#FFFFFF" fillOpacity="0.7" />
      <line x1="188" y1="52" x2="200" y2="62" stroke="#FFFFFF" strokeOpacity="0.7" strokeWidth="2" strokeLinecap="round" />
      <line x1="236" y1="82" x2="248" y2="74" stroke="#FFFFFF" strokeOpacity="0.7" strokeWidth="2" strokeLinecap="round" />
      <line x1="112" y1="132" x2="122" y2="140" stroke="#FFFFFF" strokeOpacity="0.6" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

function starPoints(cx: number, cy: number, outer: number, inner: number) {
  return Array.from({ length: 10 }, (_, i) => {
    const r = i % 2 === 0 ? outer : inner;
    const a = (-90 + i * 36) * (Math.PI / 180);
    return `${(cx + r * Math.cos(a)).toFixed(1)},${(cy + r * Math.sin(a)).toFixed(1)}`;
  }).join(' ');
}

function Rosette() {
  const bumps = Array.from({ length: 14 }, (_, i) => {
    const a = (i * 360) / 14 * (Math.PI / 180);
    return { x: 60 + 33 * Math.cos(a), y: 60 + 33 * Math.sin(a) };
  });
  return (
    <svg aria-hidden viewBox="0 0 120 146" className="h-[117px] w-[96px] drop-shadow-md">
      <polygon points="40,80 62,88 52,140 44,130 34,140" fill="#065F46" />
      <polygon points="80,80 58,88 68,140 76,130 86,140" fill="#065F46" />
      {bumps.map((b, i) => (
        <circle key={i} cx={b.x.toFixed(1)} cy={b.y.toFixed(1)} r="9" fill="#F59E0B" />
      ))}
      <circle cx="60" cy="60" r="34" fill="#FBBF24" />
      <circle cx="60" cy="60" r="26" fill="none" stroke="#F59E0B" strokeWidth="1.5" />
      <polygon points={starPoints(60, 60, 15, 6.5)} fill="#FFFFFF" />
    </svg>
  );
}