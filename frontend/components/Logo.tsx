export default function Logo({ className = '' }: { className?: string }) {
  return (
    <span className={`inline-flex items-center gap-2 text-ink ${className}`}>
      <svg width="26" height="26" viewBox="0 0 26 26" aria-hidden="true">
        <rect width="26" height="26" rx="7" className="fill-brand" />
        <path d="M9 20V7.5h5.25a4 4 0 010 8H9" fill="none" stroke="#fff" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M13 11.5l1.6 1.6 3-3.2" fill="none" stroke="#FBF1DF" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
      <span className="text-[19px] font-bold tracking-tight">praxis</span>
    </span>
  );
}
