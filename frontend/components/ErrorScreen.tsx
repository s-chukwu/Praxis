'use client';
import { RefreshCw } from 'lucide-react';
import Logo from './Logo';
import { Button, ButtonLink } from './ui/Button';

export default function ErrorScreen({ title = 'This page did not load', message, onRetry }: { title?: string; message: string; onRetry?: () => void }) {
  return (
    <div className="flex min-h-screen flex-col bg-page px-6 py-8">
      <Logo />
      <main className="mx-auto flex max-w-md flex-1 flex-col justify-center py-16">
        <h1 className="text-[28px] font-bold leading-tight">{title}</h1>
        <p className="mt-2 text-muted">{message}</p>
        <div className="mt-8 flex flex-wrap gap-3">
          {onRetry && (
            <Button size="lg" onClick={onRetry}>
              <RefreshCw className="h-4 w-4" aria-hidden /> Try again
            </Button>
          )}
          <ButtonLink href="/dashboard" size="lg" variant="secondary">
            Go to dashboard
          </ButtonLink>
        </div>
      </main>
    </div>
  );
}
