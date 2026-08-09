import type { RaceResult } from '../game/types';

/**
 * Hand-rolled SVG. The only visualisation in the game is this sparkline of
 * per-race points, and a charting library would be more code than the chart.
 */
export function Sparkline({ races }: { races: RaceResult[] }) {
  // Two cars per team now, so a round's contribution is what the
  // CONSTRUCTOR scored -- which is the number the championship uses.
  const pts = races.map((r) => {
    const mine = r.entries.filter((c) => c.team_id === 0);
    return {
      points: mine.reduce((n, c) => n + c.points, 0),
      dnf: mine.length > 0 && mine.every((c) => c.dnf),
    };
  });
  const w = 260;
  const h = 44;
  const max = 43; // both cars scoring 25 and 18
  const step = pts.length > 1 ? w / (pts.length - 1) : w;

  const path = pts
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${(i * step).toFixed(1)} ${(h - ((p?.points ?? 0) / max) * h).toFixed(1)}`)
    .join(' ');

  return (
    <svg viewBox={`0 0 ${w} ${h + 8}`} className="w-full" role="img" aria-label="Points per race">
      <line x1="0" y1={h} x2={w} y2={h} stroke="var(--color-edge)" strokeWidth="1" />
      <path d={path} fill="none" stroke="var(--color-accent)" strokeWidth="2" />
      {pts.map((p, i) => (
        <circle
          key={i}
          cx={i * step}
          cy={h - ((p?.points ?? 0) / max) * h}
          r={p?.dnf ? 3.5 : 2.5}
          fill={p?.dnf ? 'var(--color-accent)' : '#e8ecf1'}
        />
      ))}
    </svg>
  );
}
