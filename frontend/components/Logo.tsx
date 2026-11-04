export default function Logo() {
  return (
    <div className="flex items-center gap-2">
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
        <path d="M5 3h8a6 6 0 010 12H9v6H5V3z" fill="#fff" />
        <path d="M9 7l5 2.5L9 12V7z" fill="#0A0D12" />
      </svg>
      <span className="text-[17px] font-semibold tracking-tight text-white">praxis</span>
    </div>
  );
}