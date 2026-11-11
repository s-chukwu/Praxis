import type { InputHTMLAttributes } from 'react';

export default function Input({
  label,
  id,
  ...props
}: InputHTMLAttributes<HTMLInputElement> & { label: string; id: string }) {
  return (
    <div>
      <label htmlFor={id} className="block text-sm text-gray-300">
        {label}
      </label>
      <input
        id={id}
        {...props}
        className="mt-1 w-full rounded-md border border-[#1C222B] bg-[#080B0F] px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/40"
      />
    </div>
  );
}