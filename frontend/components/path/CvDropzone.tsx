'use client';
import { useState } from 'react';
import { FileCheck2, Loader2, Upload } from 'lucide-react';
import { readCvFile } from '@/lib/cvText';
import { friendly } from '@/lib/errors';

export default function CvDropzone({ onText }: { onText: (text: string) => void }) {
  const [dragging, setDragging] = useState(false);
  const [reading, setReading] = useState(false);
  const [fileName, setFileName] = useState('');
  const [error, setError] = useState('');

  const accept = async (file?: File) => {
    if (!file) return;
    setError('');
    setReading(true);
    try {
      onText(await readCvFile(file));
      setFileName(file.name);
    } catch (e) {
      setFileName('');
      setError(friendly(e));
    }
    setReading(false);
  };

  return (
    <div>
      <label
        onDragOver={(e) => {
          e.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(e) => {
          e.preventDefault();
          setDragging(false);
          void accept(e.dataTransfer.files[0]);
        }}
        className={`flex cursor-pointer flex-col items-center rounded-panel border-2 border-dashed px-6 py-8 text-center transition-colors focus-within:ring-4 focus-within:ring-brand/20 ${
          dragging ? 'border-brand bg-brand-soft' : fileName ? 'border-brand/40 bg-brand-soft/40' : 'border-line hover:border-brand/50 hover:bg-page'
        }`}
      >
        <span className={`flex h-12 w-12 items-center justify-center rounded-full ${fileName ? 'bg-brand text-white' : 'bg-page text-brand'}`}>
          {reading ? (
            <Loader2 className="h-5 w-5 animate-spin" aria-hidden />
          ) : fileName ? (
            <FileCheck2 key={fileName} className="h-5 w-5 animate-pop" aria-hidden />
          ) : (
            <Upload className="h-5 w-5" aria-hidden />
          )}
        </span>
        <span className="mt-3 font-semibold">
          {reading ? 'Reading your file…' : fileName ? fileName : dragging ? 'Drop it here' : 'Drop your CV here or choose a file'}
        </span>
        <span className="mt-1 text-sm text-muted">{fileName ? 'Text added below. Choose another file to replace it.' : 'PDF, Word (.docx) or text, up to 5 MB'}</span>
        <input
          type="file"
          accept=".pdf,.docx,.txt,application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document,text/plain"
          className="sr-only"
          onChange={(e) => void accept(e.target.files?.[0])}
        />
      </label>
      {error && (
        <p role="alert" className="mt-2 text-sm font-medium text-danger">
          {error}
        </p>
      )}
    </div>
  );
}
