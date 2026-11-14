'use client';
import { createContext, useCallback, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { supabase } from '@/lib/supabase';
import { PREVIEW, clearPreview, getLearner, type Learner } from '@/lib/learner';
import { friendly } from '@/lib/errors';

type State = {
  learner: Learner | null;
  ready: boolean;
  loadError: string;
  refresh: () => Promise<void>;
  signOut: () => Promise<void>;
};

const LearnerContext = createContext<State | null>(null);

export function LearnerProvider({ children }: { children: ReactNode }) {
  const [learner, setLearner] = useState<Learner | null>(null);
  const [ready, setReady] = useState(false);
  const [loadError, setLoadError] = useState('');

  const refresh = useCallback(async () => {
    try {
      setLearner(await getLearner());
      setLoadError('');
    } catch (e) {
      setLoadError(friendly(e, 'We could not load your training. Try again in a minute.'));
    }
    setReady(true);
  }, []);

  useEffect(() => {
    void refresh();
    const sub = supabase?.auth.onAuthStateChange(() => void refresh());
    return () => sub?.data.subscription.unsubscribe();
  }, [refresh]);

  const signOut = useCallback(async () => {
    // In preview there is no account to leave, so signing out starts the sample learner over.
    if (PREVIEW) {
      clearPreview();
      return refresh();
    }
    await supabase?.auth.signOut();
    setLearner(null);
  }, [refresh]);

  return <LearnerContext.Provider value={{ learner, ready, loadError, refresh, signOut }}>{children}</LearnerContext.Provider>;
}

export function useLearner() {
  const state = useContext(LearnerContext);
  if (!state) throw new Error('useLearner must be used inside LearnerProvider');
  return state;
}
