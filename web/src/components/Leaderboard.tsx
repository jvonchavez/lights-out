import type { LeaderboardEntry } from '../game/types';
import { playerId } from '../game/api';

export function Leaderboard({ entries }: { entries: LeaderboardEntry[] }) {
  const me = playerId();
  return (
    <div className="rounded-lg border border-edge bg-panel p-4" data-testid="leaderboard">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-widest text-muted">
        Today&rsquo;s leaderboard
      </h2>
      {entries.length === 0 ? (
        <p className="text-sm text-muted">Nobody has submitted yet.</p>
      ) : (
        <ol className="space-y-1">
          {entries.map((e) => (
            <li
              key={e.player_id}
              className={`flex items-baseline gap-3 rounded px-2 py-1 text-sm ${
                e.player_id === me ? 'bg-accent/15 font-semibold' : ''
              }`}
            >
              <span className="w-6 text-right font-mono text-xs text-muted">{e.rank}</span>
              <span className="flex-1 truncate">{e.display_name}</span>
              <span className="font-mono tabular-nums">{e.points}</span>
              <span className="w-10 text-right font-mono text-xs text-muted">{e.dnfs} DNF</span>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}
