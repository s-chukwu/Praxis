import type { ReactNode } from 'react';
import { ButtonLink } from './Button';

export default function EmptyState({
  icon,
  title,
  description,
  actionHref,
  actionLabel,
}: {
  icon?: ReactNode;
  title: string;
  description: string;
  actionHref?: string;
  actionLabel?: string;
}) {
  return (
    <div className="rounded-panel border border-dashed border-line bg-surface px-6 py-12 text-center">
      {icon && <div className="mx-auto mb-4 flex h-11 w-11 items-center justify-center rounded-full bg-page text-muted">{icon}</div>}
      <h2 className="text-lg font-semibold text-ink">{title}</h2>
      <p className="mx-auto mt-1 max-w-sm text-muted">{description}</p>
      {actionHref && actionLabel && (
        <ButtonLink href={actionHref} className="mt-6">
          {actionLabel}
        </ButtonLink>
      )}
    </div>
  );
}
