export type Dimension = { label: string; baseline?: number; final?: number };

export const dimensions: Dimension[] = [
  { label: 'Accuracy' },
  { label: 'Judgment and workflow' },
  { label: 'Safety and privacy' },
  { label: 'Communication' },
];

const W = 520;
const H = 300;
const CX = 260;
const CY = 150;
const R = 100;
const MAX = 4;

function point(i: number, n: number, value: number): [number, number] {
  const angle = ((-90 + (360 / n) * i) * Math.PI) / 180;
  const r = (value / MAX) * R;
  return [CX + r * Math.cos(angle), CY + r * Math.sin(angle)];
}

const poly = (values: number[]) =>
  values.map((v, i) => point(i, values.length, v).map((n) => n.toFixed(1)).join(',')).join(' ');

function change(d: Dimension) {
  if (d.baseline === undefined || d.final === undefined) return 'Not available yet';
  if (d.final > d.baseline) return 'Improved';
  if (d.final < d.baseline) return 'Regressed';
  return 'Unchanged';
}

export default function ProgressRadar({ items = dimensions }: { items?: Dimension[] }) {
  const n = items.length;
  const hasBaseline = items.every((d) => d.baseline !== undefined);
  const hasFinal = items.every((d) => d.final !== undefined);
  const anchor = (x: number) => (x < CX - 5 ? 'end' : x > CX + 5 ? 'start' : 'middle');

  return (
    <div className="grid items-center gap-6 md:grid-cols-[1.4fr_1fr]">
      <div>
        <svg
          viewBox={`0 0 ${W} ${H}`}
          role="img"
          aria-label="Chart comparing baseline and final scores for four skill areas, scored 0 to 4"
          className="w-full"
        >
          {[1, 2, 3, 4].map((ring) => (
            <polygon key={ring} points={poly(items.map(() => ring))} fill="none" stroke="#1C222B" />
          ))}
          {items.map((d, i) => {
            const [x, y] = point(i, n, MAX);
            return <line key={d.label} x1={CX} y1={CY} x2={x} y2={y} stroke="#1C222B" />;
          })}
          {hasBaseline && (
            <polygon
              points={poly(items.map((d) => d.baseline as number))}
              fill="none"
              stroke="#9CA3AF"
              strokeWidth="2"
              strokeDasharray="4 3"
            />
          )}
          {hasFinal && (
            <polygon
              points={poly(items.map((d) => d.final as number))}
              fill="rgba(37,99,235,0.25)"
              stroke="#3B82F6"
              strokeWidth="2"
            />
          )}
          {items.map((d, i) => {
            const [x, y] = point(i, n, 4.5);
            return (
              <text key={d.label} x={x} y={y} textAnchor={anchor(x)} dominantBaseline="middle" fontSize="13" fill="#D1D5DB">
                {d.label}
              </text>
            );
          })}
        </svg>

        <div className="flex gap-5 text-xs text-gray-400">
          <span className="flex items-center gap-1.5">
            <i className="inline-block w-5 border-t-2 border-dashed border-gray-400" />
            Baseline
          </span>
          <span className="flex items-center gap-1.5">
            <i className="inline-block w-5 border-t-2 border-blue-500" />
            Final
          </span>
        </div>
        {!hasFinal && <p className="mt-2 text-xs text-gray-400">Your chart appears after your final task.</p>}
      </div>

      <ul>
        {items.map((d) => (
          <li key={d.label} className="flex items-center justify-between gap-3 border-t border-[#161B22] py-3 first:border-t-0">
            <div>
              <p className="text-sm text-white">{d.label}</p>
              {d.baseline !== undefined && d.final !== undefined && (
                <p className="text-xs text-gray-400">
                  Baseline {d.baseline}, final {d.final}
                </p>
              )}
            </div>
            <span className="rounded border border-[#1C222B] px-2 py-0.5 text-xs text-gray-300">{change(d)}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}