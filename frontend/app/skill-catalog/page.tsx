'use client';
import { useState } from 'react';
import Link from 'next/link';

// TODO(Chidi): replace with GET /skills
const skills = [
  { name: 'AI fluency', category: 'AI & Emerging Tech', level: 'Intermediate', mins: 45, count: 4, desc: 'Understand how AI tools work and use them responsibly.' },
  { name: 'Prompt engineering', category: 'AI & Emerging Tech', level: 'Advanced', mins: 30, count: 3, desc: 'Write prompts that produce reliable, useful output.' },
  { name: 'Data literacy', category: 'Analytical Thinking', level: 'Intermediate', mins: 40, count: 4, desc: 'Read, question and communicate with data.' },
  { name: 'Critical thinking', category: 'Problem Solving', level: 'Intermediate', mins: 50, count: 4, desc: 'Evaluate evidence and test assumptions.' },
  { name: 'Communication', category: 'Human Skills', level: 'Beginner', mins: 35, count: 4, desc: 'Explain ideas clearly to different audiences.' },
  { name: 'Change management', category: 'Human Skills', level: 'Advanced', mins: 40, count: 3, desc: 'Guide teams through new ways of working.' },
];

const categories = ['All', 'AI & Emerging Tech', 'Analytical Thinking', 'Problem Solving', 'Human Skills'];

export default function SkillCatalog() {
  const [cat, setCat] = useState('All');
  const shown = cat === 'All' ? skills : skills.filter((s) => s.category === cat);

  return (
    <div className="mx-auto max-w-[900px] space-y-6">
      <div>
        <h1 className="text-[22px] font-semibold text-white">Skill Catalog</h1>
        <p className="mt-1 text-[12px] text-gray-500">Browse skills and start an assessment to see where you stand.</p>
      </div>

      <div className="flex flex-wrap gap-2">
        {categories.map((c) => (
          <button
            key={c}
            onClick={() => setCat(c)}
            className={`rounded-md border px-3 py-1.5 text-[11px] ${
              cat === c ? 'border-blue-500 bg-blue-500/10 text-blue-400' : 'border-[#1C222B] text-gray-400 hover:text-white'
            }`}
          >
            {c}
          </button>
        ))}
      </div>

      <div className="grid grid-cols-2 gap-3">
        {shown.map((s) => (
          <div key={s.name} className="rounded-lg border border-[#161B22] bg-[#0D1117] p-4">
            <p className="font-mono text-[10px] uppercase tracking-widest text-gray-500">{s.category}</p>
            <p className="mt-1 text-[14px] font-medium text-white">{s.name}</p>
            <p className="mt-1 text-[11px] leading-snug text-gray-500">{s.desc}</p>
            <div className="mt-3 flex items-center justify-between text-[10px] text-gray-500">
              <span>{s.count} assessments · {s.mins} min</span>
              <span className="rounded border border-[#1C222B] px-2 py-0.5">{s.level}</span>
            </div>
            <Link href="/assessments" className="mt-3 inline-block text-[11px] text-blue-400 hover:underline">
              View assessments →
            </Link>
          </div>
        ))}
      </div>
    </div>
  );
}