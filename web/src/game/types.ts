// Mirrors the JSON produced by internal/sim. The Go structs are the source
// of truth; these types exist so TypeScript can check the client against
// them, not to redefine anything.

export interface Decision {
  chassis: number;
  engine: number;
  aero: number;
}

/** One development part. Effect is the budget allocation it represents. */
export interface Card {
  id: string;
  name: string;
  blurb: string;
  effect: Decision;
}

export interface CircuitProfile {
  chassis: number;
  engine: number;
  aero: number;
  overtake_difficulty: number;
}

export interface Circuit {
  name: string;
  archetype: 'power' | 'technical' | 'balanced' | 'highspeed';
  profile: CircuitProfile;
}

export interface Ratings {
  chassis: number;
  engine: number;
  aero: number;
}

export interface Team {
  id: number;
  name: string;
  archetype: string;
  start: Ratings;
  driver_skill: number;
}

export interface CarResult {
  team_id: number;
  grid: number;
  finish: number;
  dnf: boolean;
  points: number;
}

export interface RaceResult {
  round: number;
  circuit: string;
  safety_car: boolean;
  cars: CarResult[];
}

export interface Standing {
  team_id: number;
  name: string;
  points: number;
  wins: number;
  podiums: number;
  dnfs: number;
}

export interface SeasonResult {
  sim_version: string;
  seed: number;
  /** The cards the player took, in window order. */
  build: Card[];
  races: RaceResult[];
  standings: Standing[];
  player: Standing;
  player_position: number;
  share: string;
}

/** The season descriptor from GET /api/seasons/today. */
export interface SeasonDescriptor {
  id: number;
  /** A string, not a number: JS floats lose precision above 2^53. */
  seed: string;
  sim_version: string;
  calendar: Circuit[];
  field: Team[];
  /** Three cards per development window, derived from the seed. */
  deals: Card[][];
  /** 0-based rounds a window precedes. */
  window_rounds: number[];
  closes_at: string;
}

export interface LeaderboardEntry {
  rank: number;
  player_id: string;
  display_name: string;
  points: number;
  wins: number;
  podiums: number;
  dnfs: number;
}

export const AREA_LABELS: Record<keyof Decision, string> = {
  chassis: 'Chassis',
  engine: 'Engine',
  aero: 'Aero',
};

/** Card cost bounds, mirroring MinCardCost/MaxCardCost in the sim. */
export const MIN_CARD_COST = 140;
export const MAX_CARD_COST = 260;

/**
 * riskPips scores a card 1-5 for display. Risk is quadratic in cumulative
 * spend and aero is credited back 30% of the pressure its own share caused,
 * so an expensive card is riskier and an aero-heavy one less so. The pips
 * are derived from the effect, never authored: what you see is the real
 * risk the simulation will apply.
 */
export function riskPips(c: Card): number {
  const cost = c.effect.chassis + c.effect.engine + c.effect.aero;
  if (cost === 0) return 1;
  const aeroShare = c.effect.aero / cost;
  const relief = 1 - 0.3 * aeroShare;
  const span = MAX_CARD_COST - MIN_CARD_COST;
  const norm = ((cost - MIN_CARD_COST) / span) * relief;
  return Math.max(1, Math.min(5, Math.round(1 + norm * 4)));
}

/** Human-readable effect, e.g. "+12 Engine · +8 Aero". Units are tenths. */
export function effectLabel(c: Card): string {
  const parts: string[] = [];
  if (c.effect.chassis > 0) parts.push(`+${Math.round(c.effect.chassis / 10)} Chassis`);
  if (c.effect.engine > 0) parts.push(`+${Math.round(c.effect.engine / 10)} Engine`);
  if (c.effect.aero > 0) parts.push(`+${Math.round(c.effect.aero / 10)} Aero`);
  return parts.join(' · ');
}

export const ARCHETYPE_LABELS: Record<Circuit['archetype'], string> = {
  power: 'Power',
  technical: 'Technical',
  balanced: 'Balanced',
  highspeed: 'High speed',
};
