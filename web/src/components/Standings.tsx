import type { Standing } from '../game/types';

export function Standings({ standings }: { standings: Standing[] }) {
  return (
    <div className="rounded-lg border border-edge bg-panel p-4">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-widest text-muted">
        Constructors
      </h2>
      <table className="w-full text-sm">
        <thead>
          <tr className="text-left text-xs uppercase tracking-wide text-muted">
            <th className="pb-2 font-medium">#</th>
            <th className="pb-2 font-medium">Team</th>
            <th className="pb-2 text-right font-medium">Pts</th>
            <th className="pb-2 text-right font-medium">Win</th>
            <th className="pb-2 text-right font-medium">DNF</th>
          </tr>
        </thead>
        <tbody>
          {standings.map((s, i) => (
            <tr
              key={s.team_id}
              className={s.team_id === 0 ? 'font-semibold text-accent' : 'text-slate-200'}
            >
              <td className="py-1 font-mono text-xs text-muted">{i + 1}</td>
              <td className="py-1">{s.name}</td>
              <td className="py-1 text-right font-mono tabular-nums">{s.points}</td>
              <td className="py-1 text-right font-mono tabular-nums text-muted">{s.wins}</td>
              <td className="py-1 text-right font-mono tabular-nums text-muted">{s.dnfs}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
