import type { LeaderboardEntry, SeasonDescriptor, SeasonResult } from './types';

const PLAYER_ID_KEY = 'lightsout.playerId';
const NAME_KEY = 'lightsout.displayName';

/**
 * Player identity is a UUID in localStorage plus a display name. No
 * passwords means no reset flow, no email, no session management and no
 * PII. Losing your device loses your streak, which for a daily browser
 * game is a reasonable trade rather than a defect.
 */
export function playerId(): string {
  let id = safeGet(PLAYER_ID_KEY);
  if (!id) {
    id = crypto.randomUUID();
    safeSet(PLAYER_ID_KEY, id);
  }
  return id;
}

export function displayName(): string {
  return safeGet(NAME_KEY) ?? '';
}

export function setDisplayName(name: string): void {
  safeSet(NAME_KEY, name);
}

function safeGet(k: string): string | null {
  try {
    return localStorage.getItem(k);
  } catch {
    return null;
  }
}

function safeSet(k: string, v: string): void {
  try {
    localStorage.setItem(k, v);
  } catch {
    /* private browsing, blocked storage: the game still plays */
  }
}

async function json<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `request failed with ${res.status}`);
  }
  return res.json() as Promise<T>;
}

/**
 * startSeason asks the server for a new season. Play is unlimited, so this
 * can be called as often as the player likes -- each call mints a fresh
 * seed, records it, and returns the calendar, field and rolls it produces.
 *
 * The seed is deliberately NOT generated here. The server is the sole
 * source, which is what keeps it authoritative when a player can have as
 * many attempts as they want.
 */
export function startSeason(): Promise<SeasonDescriptor> {
  return fetch('/api/seasons', { method: 'POST' }).then(json<SeasonDescriptor>);
}

/** Re-read an issued season, so a reload resumes the same rolls. */
export function fetchSeason(id: number): Promise<SeasonDescriptor> {
  return fetch(`/api/seasons/${id}`).then(json<SeasonDescriptor>);
}

export interface SubmitResponse {
  rank: number;
  points: number;
  wins: number;
  podiums: number;
  dnfs: number;
  position: number;
  share: string;
  result: SeasonResult;
}

/**
 * submitRun posts the five pick indices and nothing else. The score is
 * deliberately not sent: the server re-derives the rolls from the season's
 * seed, replays these picks, and computes the authoritative result itself.
 */
export function submitRun(
  seasonId: number,
  picks: number[],
  name: string,
): Promise<SubmitResponse> {
  return fetch('/api/runs', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      season_id: seasonId,
      player_id: playerId(),
      display_name: name,
      picks,
    }),
  }).then(json<SubmitResponse>);
}

/** The all-time board: one row per player, their best season. */
export function fetchLeaderboard(): Promise<LeaderboardEntry[]> {
  return fetch('/api/leaderboard?limit=50')
    .then(json<{ entries: LeaderboardEntry[] }>)
    .then((b) => b.entries);
}
