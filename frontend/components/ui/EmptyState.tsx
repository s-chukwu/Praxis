import { ButtonLink } from './Button';

export default function EmptyState({
  title,
  description,
  actionHref,
  actionLabel,
}: {
  title: string;
  description: string;
  actionHref?: string;
  actionLabel?: string;
}) {
  return (
    <div className="rounded-lg border border-[#161B22] bg-[#0D1117] p-8 text-center">
      <h2 className="text-base font-medium text-white">{title}</h2>
      <p className="mt-1 text-sm text-gray-400">{description}</p>
      {actionHref && actionLabel && (
        <div className="mt-4">
          <ButtonLink href={actionHref}>{actionLabel}</ButtonLink>
        </div>
      )}
    </div>
  );
}