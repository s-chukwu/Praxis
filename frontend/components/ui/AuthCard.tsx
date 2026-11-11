import type { ReactNode } from 'react';

export default function AuthCard({
  title,
  description,
  children,
}: {
  title: string;
  description: string;
  children: ReactNode;
}) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[#0A0D12] px-4">
      <div className="w-full max-w-[380px] rounded-lg border border-[#161B22] bg-[#0D1117] p-6">
        <h1 className="text-xl font-semibold text-white">{title}</h1>
        <p className="mt-1 text-sm text-gray-400">{description}</p>
        <div className="mt-5">{children}</div>
      </div>
    </div>
  );
}