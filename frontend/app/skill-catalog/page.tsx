'use client';
import { useState } from 'react';
import Link from 'next/link';
import PageHeader from '@/components/ui/PageHeader';


const skills = [
  { name: 'AI fluency', category: 'AI & Emerging Tech', desc: 'Understand how AI tools work and use them responsibly.' },
  { name: 'Data literacy', category: 'Analytical Thinking', desc: 'Read, question and communicate with data.' },
  { name: 'Critical thinking', category: 'Problem Solving', desc: 'Evaluate evidence and test assumptions.' },
  { name: 'Communication', category: 'Human Skills', desc: 'Explain ideas clearly to different audiences.' },
];

const categories = ['All', ...Array.from(new Set(skills.map((s) => s.category)))];

export default function SkillCatalog() {
  const [cat, setCat] = useState('All');
  const shown = cat === 'All' ? skills : skills.filter((s) => s.category === cat);

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <PageHeader title="Skill Catalog" description="Browse skills and start an assessment." />

      <div className="flex flex-wrap gap-2">
        {categories.map((c) => (
          <button
            key={c}
            onClick={() => setCat(c)}
            aria-pressed={cat === c}
            className={`rounded-md border px-3 py-1.5 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-400 ${
              cat === c ? 'border-blue-500 bg-blue-500/10 text-blue-400' : 'border-[#1C222B] text-gray-400 hover:text-white'
            }`}
          >
            {c}
          </button>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        {shown.map((s) => (
          <div key={s.name} className="rounded-lg border border-[#161B22] bg-[#0D1117] p-4">
            <p className="text-xs text-gray-500">{s.category}</p>
            <p className="mt-1 text-base font-medium text-white">{s.name}</p>
            <p className="mt-1 text-sm text-gray-400">{s.desc}</p>
            <Link href="/assessments" className="mt-3 inline-block text-sm text-blue-400 hover:underline">
              View assessments
            </Link>
          </div>
        ))}
      </div>
    </div>
  );
}