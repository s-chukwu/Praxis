import { redirect } from 'next/navigation';

// Placeholder until the landing page lands (owned by another developer); the dashboard replaced the old overview.
export default function Home() {
  redirect('/dashboard');
}
