'use client';
import { useEffect } from 'react';
import ErrorScreen from '@/components/ErrorScreen';
import { report } from '@/lib/errors';

// Catches anything a page throws, so a learner never sees a stack trace or framework error text.
export default function PageError({ error, reset }: { error: Error; reset: () => void }) {
  useEffect(() => report(error), [error]);
  return <ErrorScreen message="Something went wrong on our side. Try again in a minute." onRetry={reset} />;
}
