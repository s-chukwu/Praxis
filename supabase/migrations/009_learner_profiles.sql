
CREATE TABLE public.learner_profiles (
  user_id         uuid PRIMARY KEY REFERENCES public.profiles(id) ON DELETE CASCADE,
  programme       text,
  availability    text,
  onboarding_done boolean NOT NULL DEFAULT false,
  cv_text         text,
  claimed_skills  jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE public.learner_profiles ENABLE ROW LEVEL SECURITY;

CREATE POLICY learner_profiles_select_own ON public.learner_profiles
  FOR SELECT TO authenticated USING (user_id = auth.uid());
CREATE POLICY learner_profiles_insert_own ON public.learner_profiles
  FOR INSERT TO authenticated WITH CHECK (user_id = auth.uid());
CREATE POLICY learner_profiles_update_own ON public.learner_profiles
  FOR UPDATE TO authenticated
  USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());
