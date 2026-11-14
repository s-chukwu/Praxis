'use client';
import { useEffect, useMemo, useState } from 'react';
import { Search } from 'lucide-react';
import PageHeader from '@/components/ui/PageHeader';
import { Button } from '@/components/ui/Button';
import { useLearner } from '@/components/LearnerProvider';
import { friendly } from '@/lib/errors';
import { PREVIEW } from '@/lib/learner';
import { listUsers, setRole, type AdminUser } from '@/lib/admin';

const BASELINE: Record<string, string> = {
  none: 'Not started',
  assigned: 'Not started',
  in_progress: 'In progress',
  submitted: 'Waiting for review',
  under_review: 'Being reviewed',
  review_pending: 'Held for checking',
  reviewed: 'Reviewed',
  voided: 'Closed by admin',
};

const ROLE_LABEL: Record<string, string> = { candidate: 'Candidate', reviewer: 'Reviewer', admin: 'Admin', employer: 'Employer' };

export default function AdminUsers() {
  const { role } = useLearner();
  const [users, setUsers] = useState<AdminUser[] | null>(null);
  const [query, setQuery] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState('');
  const [notice, setNotice] = useState('');

  useEffect(() => {
    listUsers()
      .then(setUsers)
      .catch((e) => setError(friendly(e)));
  }, []);

  const shown = useMemo(() => {
    const q = query.trim().toLowerCase();
    return (users ?? []).filter((u) => !q || u.full_name.toLowerCase().includes(q) || u.id.startsWith(q));
  }, [users, query]);

  if (!PREVIEW && role !== 'admin') {
    return <p className="rounded-panel bg-surface p-6 shadow-panel">This page is for admins only.</p>;
  }

  const change = async (u: AdminUser, next: 'candidate' | 'reviewer') => {
    // Making a candidate a reviewer removes them from the pilot cohort, so ask first.
    if (next === 'reviewer' && u.baseline_status !== 'none' && !window.confirm(`${u.full_name || 'This user'} has started the baseline task. Make them a reviewer anyway?`)) return;
    setBusy(u.id);
    setError('');
    try {
      await setRole(u.id, next);
      setUsers((all) => all?.map((x) => (x.id === u.id ? { ...x, role: next } : x)) ?? null);
      setNotice(`${u.full_name || 'User'} is now a ${next}.`);
    } catch (e) {
      setError(friendly(e));
    }
    setBusy('');
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Users" description="Reviewers are invite only. Make someone a reviewer after they have signed up." />

      <label className="relative block max-w-sm">
        <span className="sr-only">Search users</span>
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" aria-hidden />
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search by name"
          className="h-10 w-full rounded-control border border-line bg-surface pl-9 pr-3 text-sm focus:border-brand focus:outline-none"
        />
      </label>

      <p className="text-sm text-muted" aria-live="polite">
        {error ? <span className="font-medium text-danger">{error}</span> : notice}
      </p>

      {!users && !error && <div className="h-48 animate-pulse rounded-panel bg-surface shadow-panel" aria-label="Loading users" />}

      {users && (
        <div className="overflow-x-auto rounded-panel bg-surface shadow-panel">
          <table className="w-full min-w-[560px] text-left text-sm">
            <thead className="border-b border-line text-muted">
              <tr>
                <th className="px-5 py-3 font-medium">Name</th>
                <th className="px-5 py-3 font-medium">Role</th>
                <th className="px-5 py-3 font-medium">Baseline</th>
                <th className="px-5 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {shown.map((u) => (
                <tr key={u.id}>
                  <td className="px-5 py-3">
                    <p className="font-medium">{u.full_name || 'No name'}</p>
                    <p className="text-xs tabular-nums text-muted">{u.id.slice(0, 8)}</p>
                  </td>
                  <td className="px-5 py-3">{ROLE_LABEL[u.role] ?? u.role}</td>
                  <td className="px-5 py-3 text-muted">{BASELINE[u.baseline_status] ?? u.baseline_status}</td>
                  <td className="px-5 py-3 text-right">
                    {u.role === 'candidate' && (
                      <Button variant="secondary" loading={busy === u.id} onClick={() => change(u, 'reviewer')}>
                        Make reviewer
                      </Button>
                    )}
                    {u.role === 'reviewer' && (
                      <Button variant="quiet" loading={busy === u.id} onClick={() => change(u, 'candidate')}>
                        Remove reviewer
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
              {shown.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-5 py-8 text-center text-muted">
                    No users match “{query}”.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
