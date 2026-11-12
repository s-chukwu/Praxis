import type { TextareaHTMLAttributes } from 'react';
import { fieldControl, fieldHint, fieldLabel } from './field';

type Props = TextareaHTMLAttributes<HTMLTextAreaElement> & { label: string; id: string; hint?: string };

export default function Textarea({ label, id, hint, ...props }: Props) {
  return (
    <div>
      <label htmlFor={id} className={fieldLabel}>
        {label}
      </label>
      <textarea
        id={id}
        aria-describedby={hint ? `${id}-hint` : undefined}
        {...props}
        className={`${fieldControl()} resize-y leading-relaxed`}
      />
      {hint && (
        <p id={`${id}-hint`} className={fieldHint}>
          {hint}
        </p>
      )}
    </div>
  );
}
