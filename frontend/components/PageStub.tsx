import Link from 'next/link';

export default function PageStub({ title, description }: { title: string; description: string }) {
  return (
    <div className="mx-auto max-w-[900px] space-y-6">
      <div>
        <h1 className="text-[22px] font-semibold text-white">{title}</h1>
        <p className="mt-1 text-[12px] text-gray-500">{description}</p>
      </div>
      <div className="rounded-lg border border-[#161B22] bg-[#0D1117] p-10 text-center">
        <p className="text-[13px] font-medium text-white">Nothing here yet</p>
        <p className="mt-1 text-[11px] text-gray-500">This section is part of the next release.</p>
        <Link href="/" className="mt-4 inline-block text-[11px] text-blue-400 hover:underline">
          Back to overview
        </Link>
      </div>
    </div>
  );
}