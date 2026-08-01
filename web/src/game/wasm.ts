import type { SeasonDescriptor, SeasonResult } from './types';

/** The functions the Go WASM module installs on globalThis. */
declare global {
  interface Window {
    Go: new () => {
      importObject: WebAssembly.Imports;
      run(instance: WebAssembly.Instance): Promise<void>;
    };
    lightsOutRunSeason?: (seed: string, picksJSON: string) => string;
    lightsOutRunPartial?: (seed: string, picksJSON: string) => string;
    lightsOutGenerateSeason?: (seed: string) => string;
    lightsOutVersion?: string;
  }
}

export interface SimAPI {
  version: string;
  runSeason(seed: string, picks: number[]): SeasonResult;
  /** Resolves only the races a prefix of picks unlocks. */
  runPartial(seed: string, picks: number[]): SeasonResult;
  generateSeason(seed: string): SeasonDescriptor;
}

let pending: Promise<SimAPI> | null = null;

/**
 * loadSim fetches and instantiates the Go simulation.
 *
 * This is the same code the server runs natively to verify a submission,
 * compiled to WebAssembly. Because there is one implementation, the browser
 * and the server cannot disagree about the rules -- a parity test asserts
 * they produce byte-identical results across 3000 seeds.
 *
 * Loaded lazily after first paint: the module is ~1.25 MB gzipped, and the
 * calendar renders from the API response without it.
 */
export function loadSim(): Promise<SimAPI> {
  if (pending) return pending;
  pending = (async () => {
    await loadScript('/wasm_exec.js');
    const go = new window.Go();
    const { instance } = await WebAssembly.instantiateStreaming(
      fetch('/sim.wasm'),
      go.importObject,
    );
    // main() ends in select{}, so this promise never settles. The globals
    // are installed during the synchronous part of startup.
    void go.run(instance);

    if (typeof window.lightsOutRunSeason !== 'function') {
      throw new Error('simulation module failed to initialise');
    }
    return {
      version: window.lightsOutVersion ?? 'unknown',
      runSeason(seed, picks) {
        return unwrap<SeasonResult>(window.lightsOutRunSeason!(seed, JSON.stringify(picks)));
      },
      runPartial(seed, picks) {
        return unwrap<SeasonResult>(window.lightsOutRunPartial!(seed, JSON.stringify(picks)));
      },
      generateSeason(seed) {
        return unwrap<SeasonDescriptor>(window.lightsOutGenerateSeason!(seed));
      },
    };
  })();
  return pending;
}

function unwrap<T>(raw: string): T {
  const parsed = JSON.parse(raw) as T & { error?: string };
  if (parsed.error) throw new Error(parsed.error);
  return parsed;
}

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) return resolve();
    const el = document.createElement('script');
    el.src = src;
    el.onload = () => resolve();
    el.onerror = () => reject(new Error(`could not load ${src}`));
    document.head.appendChild(el);
  });
}
