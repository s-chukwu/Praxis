import type { TextareaHTMLAttributes } from 'react';

export default function Textarea({
  label,
  id,
  ...props
}: TextareaHTMLAttributes<HTMLTextAreaElement> & { label: string; id: string }) {
  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-white">
        {label}
      </label>
      <textarea
        id={id}
        {...props}
        className="mt-2 w-full rounded-md border border-[#1C222B] bg-[#080B0F] px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/40"
      />
    </div>
  );
}