import Link from 'next/link';
import type { ButtonHTMLAttributes, ReactNode } from 'react';

const base =
  'inline-flex items-center justify-center rounded-md px-3 py-2 text-sm font-medium transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400';
const styles = {
  primary: 'bg-blue-600 text-white hover:bg-blue-500',
  secondary: 'border border-[#1C222B] text-gray-300 hover:text-white',
};

export function Button({
  variant = 'primary',
  className = '',
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: keyof typeof styles }) {
  return <button {...props} className={`${base} ${styles[variant]} ${className}`} />;
}

export function ButtonLink({
  href,
  children,
  variant = 'primary',
}: {
  href: string;
  children: ReactNode;
  variant?: keyof typeof styles;
}) {
  return (
    <Link href={href} className={`${base} ${styles[variant]}`}>
      {children}
    </Link>
  );
}