'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import PageHeader from '@/components/ui/PageHeader';
import Input from '@/components/ui/Input';
import Textarea from '@/components/ui/Textarea';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import CvDropzone from '@/components/path/CvDropzone';
import { programs, type ProgramId } from '@/lib/catalog';
import { CV_MAX_CHARS } from '@/lib/cvText';
import { friendly } from '@/lib/errors';
import { deleteAccount, saveLearner } from '@/lib/learner';

const card = 'rounded-panel bg-surface p-6 shadow-panel';

export default function ProfilePage() {
  const router = useRouter();
  const { learner, refresh } = useLearner();
  const [name, setName] = useState('');
  const [program, setProgram] = useState<ProgramId>('operator');
  const [availability, setAvailability] = useState('');
  const [saved, setSaved] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState('');
  const [draftCv, setDraftCv] = useState<string | null>(null);
  const [confirmText, setConfirmText] = useState('');

  useEffect(() => {
    if (!learner) return;
    setName(learner.name);
    setProgram(learner.program ?? 'operator');
    setAvailability(learner.availability ?? '');
  }, [learner]);

  if (!learner) return null;

  const run = async (what: string, task: () => Promise<void>, done: string) => {
    setBusy(what);
    setError('');
    setSaved('');
    try {
      await task();
      await refresh();
      setSaved(done);
    } catch (e) {
      setError(friendly(e));
    }
    setBusy('');
  };

  const saveDetails = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return setError('Enter your name.');
    void run('details', () => saveLearner({ name: name.trim(), program, availability: availability.trim() }), 'Your details are saved.');
  };

  const removeAccount = async () => {
    setBusy('delete');
    setError('');
    try {
      await deleteAccount();
      router.replace('/');
    } catch (e) {
      setError(friendly(e));
      setBusy('');
    }
  };

  return (
    <div className="max-w-3xl space-y-6">
      <PageHeader title="Profile" description="Your details, your CV, and your account." />
      <p className="text-sm" aria-live="polite">
        {error ? <span className="font-medium text-danger">{error}</span> : <span className="text-brand-ink">{saved}</span>}
      </p>

      <form onSubmit={saveDetails} className={`${card} space-y-5`} aria-labelledby="details-title">
        <h2 id="details-title" className="font-bold">
          Your details
        </h2>
        <Input id="name" label="Full name" autoComplete="name" value={name} onChange={(e) => setName(e.target.value)} maxLength={120} />
        <fieldset>
          <legend className="text-sm font-semibold">Programme</legend>
          <div className="mt-2 grid gap-2 sm:grid-cols-2">
            {(Object.keys(programs) as ProgramId[]).map((id) => (
              <label key={id} className={`flex cursor-pointer gap-3 rounded-control border p-3 text-sm ${program === id ? 'border-brand bg-brand-soft/40' : 'border-line'}`}>
                <input type="radio" name="program" checked={program === id} onChange={() => setProgram(id)} className="mt-0.5 accent-[#0F7B5F]" />
                <span className="font-medium">{programs[id].name}</span>
              </label>
            ))}
          </div>
          <p className="mt-2 text-[13px] text-muted">Changing programme does not change a baseline task you have already started.</p>
        </fieldset>
        <Input
          id="availability"
          label="When you can train"
          hint="For example: weekday evenings, Saturday mornings."
          value={availability}
          maxLength={120}
          onChange={(e) => setAvailability(e.target.value)}
        />
        <Button type="submit" loading={busy === 'details'}>
          Save details
        </Button>
      </form>

      <section className={`${card} space-y-4`} aria-labelledby="cv-title">
        <h2 id="cv-title" className="font-bold">
          Your CV
        </h2>
        {draftCv === null ? (
          <>
            <p className="text-sm text-muted">
              {learner.cvSubmitted ? 'A CV is saved to your profile. It is not analysed by AI.' : 'No CV saved. Adding one is optional.'}
            </p>
            {learner.cvText && <p className="line-clamp-4 whitespace-pre-line rounded-control bg-page p-3 text-sm">{learner.cvText}</p>}
            <div className="flex flex-wrap gap-3">
              <Button variant="secondary" onClick={() => setDraftCv('')}>
                {learner.cvSubmitted ? 'Replace with a new PDF' : 'Upload a PDF'}
              </Button>
              {learner.cvSubmitted && (
                <Button variant="quiet" loading={busy === 'cv-delete'} onClick={() => run('cv-delete', () => saveLearner({ cvText: null, cvSubmitted: false }), 'Your CV is deleted. Your account is unchanged.')}>
                  Delete my CV
                </Button>
              )}
            </div>
          </>
        ) : (
          <>
            <CvDropzone onText={setDraftCv} />
            {draftCv && (
              <Textarea id="cv-review" label="Check the text before saving" rows={7} maxLength={CV_MAX_CHARS} value={draftCv} onChange={(e) => setDraftCv(e.target.value)} />
            )}
            <div className="flex flex-wrap gap-3">
              <Button
                disabled={!draftCv?.trim()}
                loading={busy === 'cv'}
                onClick={() =>
                  run('cv', async () => {
                    await saveLearner({ cvText: (draftCv ?? '').trim(), cvSubmitted: true });
                    setDraftCv(null);
                  }, 'Your new CV is saved.')
                }
              >
                Save this CV
              </Button>
              <Button variant="quiet" onClick={() => setDraftCv(null)}>
                Cancel
              </Button>
            </div>
          </>
        )}
      </section>

      <section className="space-y-3 rounded-panel border border-danger/30 bg-surface p-6" aria-labelledby="danger-title">
        <h2 id="danger-title" className="font-bold text-danger">
          Delete your account
        </h2>
        <p className="text-sm">
          This deletes your profile, your CV and your name, and you can no longer sign in. Your task scores stay for the pilot evaluation, with nothing that identifies you.
          This cannot be undone.
        </p>
        <Input id="confirm-delete" label='Type "DELETE" to confirm' value={confirmText} onChange={(e) => setConfirmText(e.target.value)} autoComplete="off" />
        <Button variant="danger" disabled={confirmText !== 'DELETE'} loading={busy === 'delete'} onClick={removeAccount}>
          Delete my account
        </Button>
      </section>
    </div>
  );
}
