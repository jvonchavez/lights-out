import { useState } from 'react';
import type { SeasonResult } from '../game/types';
import { Sparkline } from './Sparkline';

export function SeasonComplete({
  result,
  onSubmit,
  submitting,
  submitted,
  error,
  name,
  onNameChange,
}: {
  result: SeasonResult;
  onSubmit: () => void;
  submitting: boolean;
  submitted: boolean;
  error: string | null;
  name: string;
  onNameChange: (v: string) => void;
}) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(result.share);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      setCopied(false);
    }
  }

  return (
    <div className="rounded-lg border border-edge bg-panel p-6" data-testid="season-complete">
      <p className="text-xs font-semibold uppercase tracking-widest text-muted">Season complete</p>
      <p className="mt-1 text-4xl font-bold tabular-nums">
        P{result.player_position}
        <span className="ml-3 text-xl font-normal text-muted">{result.player.points} pts</span>
      </p>

      <div className="mt-4">
        <Sparkline races={result.races} />
      </div>

      <pre
        data-testid="share-string"
        className="mt-4 overflow-x-auto whitespace-pre rounded border border-edge bg-track p-3 text-sm leading-relaxed"
        style={{
          // The emoji row is the point of the share card, so name the colour
          // emoji fonts explicitly rather than relying on the monospace
          // stack, which on many systems has no emoji coverage and renders
          // the season as tofu boxes.
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, monospace, "Apple Color Emoji", "Segoe UI Emoji", "Noto Color Emoji"',
        }}
      >
        {result.share}
      </pre>

      <button
        onClick={copy}
        className="mt-2 rounded border border-edge px-3 py-1.5 text-xs text-muted transition hover:border-muted hover:text-slate-100"
      >
        {copied ? 'Copied' : 'Copy result'}
      </button>

      {!submitted && (
        <div className="mt-6 border-t border-edge pt-4">
          <label htmlFor="name" className="text-xs uppercase tracking-widest text-muted">
            Name for the leaderboard
          </label>
          <div className="mt-2 flex gap-2">
            <input
              id="name"
              data-testid="display-name"
              value={name}
              maxLength={32}
              onChange={(e) => onNameChange(e.target.value)}
              placeholder="Anonymous"
              className="flex-1 rounded border border-edge bg-track px-3 py-2 text-sm outline-none focus:border-muted"
            />
            <button
              data-testid="submit-run"
              onClick={onSubmit}
              disabled={submitting}
              className="rounded bg-accent px-4 py-2 text-sm font-semibold text-white transition hover:brightness-110 disabled:opacity-50"
            >
              {submitting ? 'Verifying…' : 'Submit'}
            </button>
          </div>
          <p className="mt-2 text-xs text-muted">
            Only your ten decisions are sent. The server replays them and computes the score itself,
            so the leaderboard measures decisions and nothing else.
          </p>
        </div>
      )}

      {submitted && (
        <p className="mt-6 border-t border-edge pt-4 text-sm text-emerald-400" data-testid="submitted">
          Submitted and verified.
        </p>
      )}
      {error && <p className="mt-3 text-sm text-accent">{error}</p>}
    </div>
  );
}
