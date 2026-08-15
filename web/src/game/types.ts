// Mirrors the JSON produced by internal/sim. The Go structs are the source
// of truth; these types exist so TypeScript can check the client against
// them, not to redefine anything.

export interface CarSpec {
  id: string;
  name: string;
  power: number;
  cornering: number;
  aero: number;
  reliability: number;
}

export interface DriverSpec {
  id: string;
  name: string;
  pace: number;
  racecraft: number;
  consistency: number;
  composure: number;
}

export interface EngineerSpec {
  id: string;
  name: string;
  setup: number;
  strategy: number;
  ops: number;
}

export interface PrincipalSpec {
  id: string;
  name: string;
  development: number;
  leadership: number;
  nerve: number;
}

/** One team-season: the unit a roll lands on. */
export interface TeamEra {
  id: string;
  team: string;
  year: number;
  era_id: string;
  livery: string;
  car: CarSpec;
  drivers: [DriverSpec, DriverSpec];
  engineer: EngineerSpec;
  principal: PrincipalSpec;
}

export interface Lineup {
  car: CarSpec;
  drivers: [DriverSpec, DriverSpec];
  engineer: EngineerSpec;
  principal: PrincipalSpec;
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

export interface Team {
  id: number;
  name: string;
  livery: string;
  lineup: Lineup;
}

export interface EntryResult {
  team_id: number;
  entry: number;
  driver_id: string;
  driver: string;
  grid: number;
  finish: number;
  dnf: boolean;
  dnf_reason: '' | 'mechanical' | 'driver';
  points: number;
}

export interface RaceResult {
  round: number;
  circuit: string;
  safety_car: boolean;
  entries: EntryResult[];
}

export interface Standing {
  team_id: number;
  name: string;
  points: number;
  wins: number;
  podiums: number;
  dnfs: number;
}

export interface DriverStanding {
  driver_id: string;
  name: string;
  team_id: number;
  points: number;
  wins: number;
}

export interface SeasonResult {
  sim_version: string;
  seed: number;
  rolls: TeamEra[];
  picks: number[];
  lineup: Lineup;
  races: RaceResult[];
  standings: Standing[];
  drivers: DriverStanding[];
  player: Standing;
  player_position: number;
  share: string;
}

/**
 * A season issued by POST /api/seasons.
 *
 * The seed is minted server-side and never chosen by the client -- a client
 * that could nominate its own seed could nominate one it had already solved
 * offline, and the submission would verify perfectly.
 */
export interface SeasonDescriptor {
  id: number;
  /** A string, not a number: JS floats lose precision above 2^53. */
  seed: string;
  sim_version: string;
  calendar: Circuit[];
  /** The 2026 grid. Identical in every season. */
  field: Team[];
  /** The five team-eras this seed offers, derived from the seed. */
  rolls: TeamEra[];
}

/** One all-time row: a player's best season, and how many they have played. */
export interface LeaderboardEntry {
  rank: number;
  player_id: string;
  display_name: string;
  points: number;
  wins: number;
  podiums: number;
  dnfs: number;
  runs: number;
}

// ---------------------------------------------------------------------------
// The draft. These five constants mirror sim.ItemKind and must stay in the
// same order: a pick is the integer index the server replays.
// ---------------------------------------------------------------------------

export const ITEM_CAR = 0;
export const ITEM_DRIVER_A = 1;
export const ITEM_DRIVER_B = 2;
export const ITEM_ENGINEER = 3;
export const ITEM_PRINCIPAL = 4;

export type Slot = 'car' | 'driver' | 'engineer' | 'principal';

export const SLOT_OF: Record<number, Slot> = {
  [ITEM_CAR]: 'car',
  [ITEM_DRIVER_A]: 'driver',
  [ITEM_DRIVER_B]: 'driver',
  [ITEM_ENGINEER]: 'engineer',
  [ITEM_PRINCIPAL]: 'principal',
};

/** How many of each slot a complete team has. Sums to ROLL_COUNT. */
export const SLOT_CAPACITY: Record<Slot, number> = {
  car: 1,
  driver: 2,
  engineer: 1,
  principal: 1,
};

export const ROLL_COUNT = 5;
export const TEAM_COUNT = 12;

export const SLOT_LABELS: Record<Slot, string> = {
  car: 'Car',
  driver: 'Driver',
  engineer: 'Race engineer',
  principal: 'Team principal',
};

/**
 * Overall is COMPUTED here exactly as internal/sim computes it, never read
 * from the wire, because the sim does not send it. The weights are mirrored
 * from roster.go; a divergence would show the player a number the race does
 * not use, which is the one thing the rule about derived display forbids.
 */
export function carOverall(c: CarSpec): number {
  return Math.floor((30 * c.power + 30 * c.cornering + 30 * c.aero + 10 * c.reliability) / 100);
}

export function driverOverall(d: DriverSpec): number {
  return Math.floor((40 * d.pace + 25 * d.racecraft + 20 * d.consistency + 15 * d.composure) / 100);
}

export function engineerOverall(e: EngineerSpec): number {
  return Math.floor((35 * e.setup + 40 * e.strategy + 25 * e.ops) / 100);
}

export function principalOverall(p: PrincipalSpec): number {
  return Math.floor((45 * p.development + 30 * p.leadership + 25 * p.nerve) / 100);
}

export interface Item {
  kind: number;
  slot: Slot;
  name: string;
  overall: number;
  /** The two or three ratings worth showing on the card face. */
  stats: Array<{ label: string; value: number }>;
}

/** The five things a rolled team-era offers. */
export function itemsOf(te: TeamEra): Item[] {
  return [
    {
      kind: ITEM_CAR,
      slot: 'car',
      name: te.car.name,
      overall: carOverall(te.car),
      stats: [
        { label: 'PWR', value: te.car.power },
        { label: 'COR', value: te.car.cornering },
        { label: 'AER', value: te.car.aero },
        { label: 'REL', value: te.car.reliability },
      ],
    },
    ...te.drivers.map((d, i) => ({
      kind: i === 0 ? ITEM_DRIVER_A : ITEM_DRIVER_B,
      slot: 'driver' as Slot,
      name: d.name,
      overall: driverOverall(d),
      stats: [
        { label: 'PAC', value: d.pace },
        { label: 'RCE', value: d.racecraft },
        { label: 'CON', value: d.consistency },
        { label: 'CMP', value: d.composure },
      ],
    })),
    {
      kind: ITEM_ENGINEER,
      slot: 'engineer',
      name: te.engineer.name,
      overall: engineerOverall(te.engineer),
      stats: [
        { label: 'SET', value: te.engineer.setup },
        { label: 'STR', value: te.engineer.strategy },
        { label: 'OPS', value: te.engineer.ops },
      ],
    },
    {
      kind: ITEM_PRINCIPAL,
      slot: 'principal',
      name: te.principal.name,
      overall: principalOverall(te.principal),
      stats: [
        { label: 'DEV', value: te.principal.development },
        { label: 'LED', value: te.principal.leadership },
        { label: 'NRV', value: te.principal.nerve },
      ],
    },
  ];
}

export const ARCHETYPE_LABELS: Record<Circuit['archetype'], string> = {
  power: 'Power',
  technical: 'Technical',
  balanced: 'Balanced',
  highspeed: 'High speed',
};

export const DNF_LABELS: Record<'mechanical' | 'driver', string> = {
  mechanical: 'Mechanical',
  driver: 'Driver error',
};
