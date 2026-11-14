import { PREVIEW, request, type Role } from './learner';

export type AdminUser = { id: string; full_name: string; role: Role; baseline_status: string };

const SAMPLE_USERS: AdminUser[] = [
  { id: '11111111-aaaa-4aaa-8aaa-000000000001', full_name: 'Amari Uwase', role: 'candidate', baseline_status: 'submitted' },
  { id: '11111111-aaaa-4aaa-8aaa-000000000002', full_name: 'Eric Niyonzima', role: 'reviewer', baseline_status: 'none' },
  { id: '11111111-aaaa-4aaa-8aaa-000000000003', full_name: 'Claudine Uwera', role: 'candidate', baseline_status: 'in_progress' },
  { id: '11111111-aaaa-4aaa-8aaa-000000000004', full_name: 'Aline Mukamana', role: 'admin', baseline_status: 'none' },
];

export async function listUsers(): Promise<AdminUser[]> {
  if (PREVIEW) return SAMPLE_USERS;
  return (await request<{ users: AdminUser[] }>('GET', '/v1/admin/users')).users;
}

// Only candidate and reviewer can be set here; the backend refuses admin and self changes.
export async function setRole(userId: string, role: 'candidate' | 'reviewer'): Promise<void> {
  if (PREVIEW) return;
  await request('PUT', `/v1/admin/users/${userId}/role`, { role });
}
