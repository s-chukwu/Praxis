export const fieldLabel = 'block text-sm font-semibold text-ink';
export const fieldHint = 'mt-1.5 text-[13px] text-muted';
export const fieldError = 'mt-1.5 text-[13px] font-medium text-danger';

export function fieldControl(invalid?: boolean) {
  return (
    'mt-2 w-full rounded-control border bg-surface px-3.5 py-2.5 text-[15px] text-ink placeholder:text-muted/80 ' +
    'transition-[border-color,box-shadow] focus:outline-none focus:ring-4 disabled:bg-page disabled:text-muted ' +
    (invalid ? 'border-danger focus:ring-danger/15' : 'border-line hover:border-ink/25 focus:border-brand focus:ring-brand/15')
  );
}
