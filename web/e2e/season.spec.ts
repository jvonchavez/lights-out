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
 * appear on the board.
 */
test('draft a team, race the season, and appear on the leaderboard', async ({ page }) => {
  const failures: string[] = [];
  page.on('pageerror', (e) => failures.push(String(e)));
  page.on('console', (m) => {
    if (m.type() === 'error') failures.push(m.text());
  });

  await page.goto('/');

  // The roll renders from the API before the WASM module has finished
  // loading: the server is the authority on what was offered.
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

test('replaying the same day is rejected for the same browser', async ({ page }) => {
  // Each Playwright test gets a fresh context, so localStorage starts empty
  // and this browser is a new player. Both submissions therefore have to
  // happen inside THIS context for the player UUID to be shared.
  await page.goto('/');
  await draft(page);
  await page.getByTestId('display-name').fill('replayer');
  await page.getByTestId('submit-run').click();
  await expect(page.getByTestId('submitted')).toBeVisible();

  await page.reload();
  await draft(page);
  await page.getByTestId('display-name').fill('replayer');
  await page.getByTestId('submit-run').click();

  await expect(page.getByText(/already submitted/i)).toBeVisible();
});

test('the default share stays spoiler-free and copying the team is opt-in', async ({ page }) => {
  await page.goto('/');
  await draft(page, [1, 0, 2, 3, 4]);

  const share = await page.getByTestId('share-string').innerText();
  const names = await page.getByTestId('lineup-name').allInnerTexts();

  expect(names).toHaveLength(ROLLS);

  // On a shared daily seed your LINEUP is the strategy, so naming it gives
  // the day away and the default copy must not do it.
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
 * Free play is the replayability answer, and it needs no backend: the sim
 * generates a season in the browser and the run is simply never posted.
 */
test('free play starts a new roll and is not scored', async ({ page }) => {
  await page.goto('/');
  await draft(page);
  await expect(page.getByTestId('submit-run')).toBeVisible();

  await page.getByTestId('play-again').click();

  await expect(page.getByText(/Roll 1 of 5/)).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText(/Free play/)).toBeVisible();

  await draft(page);
  // No leaderboard submission in free play: nothing to forge, nothing to
  // score, and the daily seed stays the only thing measured.
  await expect(page.getByTestId('submit-run')).toHaveCount(0);
  await expect(page.getByText(/not scored/i)).toBeVisible();
});
