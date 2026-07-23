// Mirrors the JSON produced by internal/sim. The Go structs are the source
// of truth; these types exist so TypeScript can check the client against
// them, not to redefine anything.

export interface Decision {
  chassis: number;
  engine: number;
  aero: number;
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
  budgets: number[];
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

export const ARCHETYPE_LABELS: Record<Circuit['archetype'], string> = {
  power: 'Power',
  technical: 'Technical',
  balanced: 'Balanced',
  highspeed: 'High speed',
};
