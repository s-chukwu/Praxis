import Link from 'next/link';
import { Loader2 } from 'lucide-react';
import type { ButtonHTMLAttributes, ReactNode } from 'react';

const base =
  'inline-flex select-none items-center justify-center gap-2 rounded-control font-semibold transition ' +
  'active:translate-y-px disabled:pointer-events-none disabled:opacity-50';
const variants = {
  primary: 'bg-brand text-white shadow-[0_1px_0_rgba(10,86,66,0.6),0_2px_6px_-2px_rgba(10,86,66,0.45)] hover:bg-brand-hover',
  secondary: 'border border-line bg-surface text-ink hover:border-ink/25 hover:bg-page',
  quiet: 'text-brand-ink hover:bg-brand-soft',
  danger: 'border border-danger/30 bg-surface text-danger hover:bg-danger-soft',
};
const sizes = { md: 'h-10 px-4 text-sm', lg: 'h-12 px-6 text-[15px]' };

type Look = { variant?: keyof typeof variants; size?: keyof typeof sizes };

function look({ variant = 'primary', size = 'md' }: Look, className = '') {
  return `${base} ${variants[variant]} ${sizes[size]} ${className}`;
}

export function Button({
  variant,
  size,
  loading = false,
  className,
  children,
  disabled,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & Look & { loading?: boolean }) {
  return (
    <button {...props} disabled={disabled || loading} aria-busy={loading} className={look({ variant, size }, className)}>
      {loading && <Loader2 aria-hidden className="h-4 w-4 animate-spin" />}
      {children}
    </button>
  );
}

export function ButtonLink({
  href,
  children,
  variant,
  size,
  className,
}: Look & { href: string; children: ReactNode; className?: string }) {
  return (
    <Link href={href} className={look({ variant, size }, className)}>
      {children}
    </Link>
  );
}
