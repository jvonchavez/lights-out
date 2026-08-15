import { test, expect, type Page } from '@playwright/test';

const ROLLS = 5;

/** A complete legal draft: car, both drivers, engineer, principal. */
const LEGAL = [0, 1, 2, 3, 4];

/**
 * Draft a team, taking one item per roll and waiting for each roll to come
 * round. The reel is skipped: it is presentation over a result the client
 * already holds, and forty-five seconds per test is not a useful wait.
 */
async function draft(page: Page, picks: number[] = LEGAL) {
  for (let r = 1; r <= ROLLS; r++) {
    await expect(page.getByText(new RegExp(`Roll ${r} of ${ROLLS}`))).toBeVisible({
      timeout: 30_000,
    });
    await page.getByTestId(`item-${picks[r - 1]}`).click();
    await page.getByTestId('confirm-pick').click();
  }
  await expect(page.getByTestId('reel')).toBeVisible({ timeout: 30_000 });
  await page.getByTestId('skip-reel').click();
  await expect(page.getByTestId('season-complete')).toBeVisible({ timeout: 30_000 });
}

/**
 * The one end-to-end test that matters: load the page, draft a team out of
 * Formula 1 history against the WASM simulation, watch it race, submit, and
 * appear on the all-time board.
 */
test('draft a team, race the season, and appear on the leaderboard', async ({ page }) => {
  const failures: string[] = [];
  page.on('pageerror', (e) => failures.push(String(e)));
  page.on('console', (m) => {
    if (m.type() === 'error') failures.push(m.text());
  });

  await page.goto('/');

  // The roll renders from the API before the WASM module has finished
  // loading: the server mints the seed and is the authority on what was
  // offered.
  await expect(page.getByTestId('team-era')).toBeVisible();
  await expect(page.getByTestId('roll-team')).not.toBeEmpty();
  for (let k = 0; k < 5; k++) await expect(page.getByTestId(`item-${k}`)).toBeVisible();

  // Confirm is disabled until an item is chosen: a roll cannot be skipped.
  await expect(page.getByTestId('confirm-pick')).toBeDisabled();

  await draft(page);

  // The lineup is the artifact the draft produces.
  const lineup = page.getByTestId('lineup');
  await expect(lineup).toBeVisible();
  await expect(lineup.getByTestId('lineup-row')).toHaveCount(ROLLS);

  const share = await page.getByTestId('share-string').innerText();
  expect(share).toMatch(/^Lights Out · Season \d+/m);
  expect(share).toMatch(/P\d+ of \d+ · \d+ pts/);
  expect(share.split('\n')).toHaveLength(3);

  // Both drivers score, so the drivers' table is part of the result.
  await expect(page.getByTestId('driver-standings')).toBeVisible();

  const name = `e2e-${Date.now()}`;
  await page.getByTestId('display-name').fill(name);
  await page.getByTestId('submit-run').click();

  await expect(page.getByTestId('submitted')).toBeVisible();
  const board = page.getByTestId('leaderboard');
  await expect(board).toBeVisible();
  await expect(board.getByText(name)).toBeVisible();

  expect(failures, `console/page errors:\n${failures.join('\n')}`).toEqual([]);
});

/**
 * A slot with no room left cannot be filled again. This is the draft's own
 * rule made visible: taking the car closes the car slot, and the sim would
 * reject a submission that broke it anyway.
 */
test('a filled slot closes and the draft cannot go illegal', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByTestId('item-0')).toBeEnabled();

  await page.getByTestId('item-0').click(); // the car
  await page.getByTestId('confirm-pick').click();

  await expect(page.getByText(/Roll 2 of 5/)).toBeVisible();
  await expect(page.getByTestId('item-0')).toBeDisabled();
  await expect(page.getByTestId('item-1')).toBeEnabled();
});

/**
 * Play is unlimited: the same player can submit as many seasons as they
 * like, because each one is a separate issued season. What they cannot do
 * is submit the SAME season twice -- UNIQUE (season_id, player_id) settles
 * that in the database rather than in application logic that can race.
 *
 * Each Playwright test gets a fresh context, so localStorage starts empty
 * and this browser is one new player throughout.
 */
test('a player can submit season after season', async ({ page }) => {
  // The board is all-time and shared, so scope every assertion to THIS
  // player's row -- a bare /1 run/ matches whoever else has played.
  const name = `grinder-${Date.now()}`;
  const myRow = () =>
    page.getByTestId('leaderboard').getByTestId('board-row').filter({ hasText: name });

  await page.goto('/');
  await draft(page);
  await page.getByTestId('display-name').fill(name);
  await page.getByTestId('submit-run').click();
  await expect(page.getByTestId('submitted')).toBeVisible();
  await expect(myRow()).toHaveText(/1 run/);

  await page.getByTestId('play-again').click();
  await expect(page.getByText(/Roll 1 of 5/)).toBeVisible({ timeout: 30_000 });
  await draft(page);
  await page.getByTestId('display-name').fill(name);
  await page.getByTestId('submit-run').click();
  await expect(page.getByTestId('submitted')).toBeVisible();

  // The board keeps the BEST season and counts them both.
  await expect(myRow()).toHaveText(/2 runs/);
});

test('the default share stays spoiler-free and copying the team is opt-in', async ({ page }) => {
  await page.goto('/');
  await draft(page, [1, 0, 2, 3, 4]);

  const share = await page.getByTestId('share-string').innerText();
  const names = await page.getByTestId('lineup-name').allInnerTexts();

  expect(names).toHaveLength(ROLLS);

  // The default copy is a boast anyone can read cold; the team that
  // produced it is a second, deliberate click.
  for (const raw of names) {
    const n = raw.trim();
    // Guard: toContain('') is vacuously true and would hide a real leak.
    expect(n.length).toBeGreaterThan(0);
    expect(share).not.toContain(n);
  }
  await expect(page.getByTestId('copy-plain')).toBeVisible();
  await expect(page.getByTestId('copy-build')).toBeVisible();
});

/**
 * Playing again must actually deal a different season -- a new seed, a new
 * calendar of real circuits, and a new set of five rolls. If it did not,
 * "unlimited" would just mean replaying one puzzle.
 */
test('playing again deals a genuinely different season', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByTestId('team-era')).toBeVisible();

  const firstLabel = await page.getByTestId('season-label').innerText();
  const firstTeam = await page.getByTestId('roll-team').innerText();
  const firstCircuits = await page.getByTestId('circuit-name').allInnerTexts();

  await draft(page);
  await page.getByTestId('play-again').click();
  await expect(page.getByText(/Roll 1 of 5/)).toBeVisible({ timeout: 30_000 });

  // A different issued season, so a different id.
  await expect(page.getByTestId('season-label')).not.toHaveText(firstLabel);

  const secondCircuits = await page.getByTestId('circuit-name').allInnerTexts();
  expect(secondCircuits).toHaveLength(10);
  // The calendar is drawn from a pool of real circuits, so two seasons
  // running should not be identical. Allow the roll to coincide -- 33
  // team-eras means that happens -- but not the whole calendar.
  const secondTeam = await page.getByTestId('roll-team').innerText();
  expect(
    secondCircuits.join() !== firstCircuits.join() || secondTeam !== firstTeam,
    'a new season reproduced the previous calendar AND the previous opening roll',
  ).toBe(true);
});

/** The calendar must be real circuits, not the old fictional placeholders. */
test('the calendar is real Formula 1 circuits', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByTestId('team-era')).toBeVisible();

  const names = await page.getByTestId('circuit-name').allInnerTexts();
  expect(names).toHaveLength(10);

  // None of the fictional placeholders survive.
  for (const gone of ['Vellmar Straight', 'Kestrel Park', 'Sable Bay', 'Mont Aubade']) {
    expect(names).not.toContain(gone);
  }
  // And the pool is recognisable: over a handful of seasons a famous
  // circuit will turn up. Checking one season would be flaky, so this
  // asserts every name is drawn from the real pool instead.
  const real = new Set([
    'Monza', 'Spa-Francorchamps', 'Baku City', 'Las Vegas Strip', 'Mexico City',
    'Red Bull Ring', 'Gilles Villeneuve', 'Hockenheimring', 'Paul Ricard',
    'Monaco', 'Hungaroring', 'Zandvoort', 'Marina Bay', 'Imola', 'Magny-Cours',
    'Adelaide', 'Valencia Street', 'Shanghai', 'Miami', 'Barcelona-Catalunya',
    'Madring', 'Bahrain', 'Circuit of the Americas', 'Interlagos', 'Yas Marina',
    'Sepang', 'Nurburgring', 'Silverstone', 'Suzuka', 'Albert Park', 'Lusail',
    'Istanbul Park', 'Kyalami', 'Buddh International',
  ]);
  for (const n of names) {
    expect(real.has(n.trim()), `${n} is not a real circuit in the pool`).toBe(true);
  }
});
