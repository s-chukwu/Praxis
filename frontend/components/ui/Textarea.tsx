import type { TextareaHTMLAttributes } from 'react';
import { fieldControl, fieldError, fieldHint, fieldLabel } from './field';

type Props = TextareaHTMLAttributes<HTMLTextAreaElement> & { label: string; id: string; hint?: string; error?: string };

export default function Textarea({ label, id, hint, error, ...props }: Props) {
  const note = error ? `${id}-error` : hint ? `${id}-hint` : undefined;
  return (
    <div>
      <label htmlFor={id} className={fieldLabel}>
        {label}
      </label>
      <textarea
        id={id}
        aria-invalid={Boolean(error)}
        aria-describedby={note}
        {...props}
        className={`${fieldControl(Boolean(error))} resize-y leading-relaxed`}
      />
      {error ? (
        <p id={note} className={fieldError}>
          {error}
        </p>
      ) : (
        hint && (
          <p id={note} className={fieldHint}>
            {hint}
          </p>
        )
      )}
    </div>
  );
}
