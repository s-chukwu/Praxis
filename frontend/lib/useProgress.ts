'use client';
import { useEffect, useState } from 'react';
import { report } from './errors';
import { getProgress, type Progress } from './baseline';

// Null while loading or if the backend is unreachable; pages then fall back to the first steps.
export function useProgress(): Progress | null {
  const [progress, setProgress] = useState<Progress | null>(null);
  useEffect(() => {
    getProgress().then(setProgress).catch(report);
  }, []);
  return progress;
}
