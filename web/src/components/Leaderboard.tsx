import type { LeaderboardEntry } from '../game/types';
import { playerId } from '../game/api';

/**
 * The all-time board: one row per player, their best season.
 *
 * The run count is shown on purpose. When play is unlimited a best-of-N is
 * partly a measure of N, and a player who found 260 points in three seasons
 * did something different from one who found it in two hundred. Hiding the
 * N would make the board look like it measures only skill.
 */
export function Leaderboard({ entries }: { entries: LeaderboardEntry[] }) {
  const me = playerId();
  return (
    <div className="rounded-lg border border-edge bg-panel p-4" data-testid="leaderboard">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-widest text-muted">
        All-time best seasons
      </h2>
      {entries.length === 0 ? (
        <p className="text-sm text-muted">Nobody has submitted yet.</p>
      ) : (
        <ol className="space-y-1">
          {entries.map((e) => (
            <li
              key={e.player_id}
              data-testid="board-row"
              className={`flex items-baseline gap-3 rounded px-2 py-1 text-sm ${
                e.player_id === me ? 'bg-accent/15 font-semibold' : ''
              }`}
            >
              <span className="w-6 text-right font-mono text-xs text-muted">{e.rank}</span>
              <span className="flex-1 truncate">{e.display_name}</span>
              <span className="font-mono tabular-nums">{e.points}</span>
              <span className="w-16 text-right font-mono text-xs text-muted">
                {e.runs} {e.runs === 1 ? 'run' : 'runs'}
              </span>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}
