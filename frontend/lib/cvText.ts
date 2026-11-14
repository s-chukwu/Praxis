import { FriendlyError } from './errors';

export const CV_MAX_CHARS = 20000;
const MAX_BYTES = 5 * 1024 * 1024;

export async function readCvFile(file: File): Promise<string> {
  if (file.size > MAX_BYTES) throw new FriendlyError('That file is over 5 MB. Upload a smaller copy of your CV.');
  // PM decision X-06: PDF only, so every CV is read the same way.
  if (!file.name.toLowerCase().endsWith('.pdf') && file.type !== 'application/pdf') {
    throw new FriendlyError('Upload your CV as a PDF. In Word or Google Docs, use Save as PDF or Download as PDF.');
  }
  const text = await readPdf(file);

  const clean = text.replace(/[ \t]+/g, ' ').replace(/\n{3,}/g, '\n\n').trim();
  if (clean.length < 30) throw new FriendlyError('We could not find text in that file. If it is a scanned image, paste the text instead.');
  return clean.slice(0, CV_MAX_CHARS);
}

async function readPdf(file: File) {
  const pdfjs = await import('pdfjs-dist');
  pdfjs.GlobalWorkerOptions.workerSrc = new URL('pdfjs-dist/build/pdf.worker.min.mjs', import.meta.url).toString();
  const pdf = await pdfjs.getDocument({ data: await file.arrayBuffer() }).promise;
  const pages: string[] = [];
  for (let n = 1; n <= Math.min(pdf.numPages, 10); n++) {
    const content = await (await pdf.getPage(n)).getTextContent();
    // hasEOL marks the end of a visual line; without it a PDF's words run together into one paragraph.
    pages.push(content.items.map((item) => ('str' in item ? item.str + (item.hasEOL ? '\n' : ' ') : '')).join(''));
  }
  return pages.join('\n');
}
