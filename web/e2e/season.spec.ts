import { test, expect } from '@playwright/test';

/**
 * The one end-to-end test that matters: load the page, play a full ten-race
 * season against the WASM simulation, submit, and appear on the leaderboard.
 */
test('play a full season and appear on the leaderboard', async ({ page }) => {
  const failures: string[] = [];
  page.on('pageerror', (e) => failures.push(String(e)));
  page.on('console', (m) => {
    if (m.type() === 'error') failures.push(m.text());
  });

  await page.goto('/');

  // The season descriptor renders before the WASM module has loaded.
  await expect(page.getByText(/Round 1 of 10/)).toBeVisible();

  for (let round = 1; round <= 10; round++) {
    await expect(page.getByText(new RegExp(`Round ${round} of 10`))).toBeVisible();

    // Spend the whole budget: aero first, then split the rest.
    const aero = page.getByTestId('slider-aero');
    await aero.fill('40');
    await page.getByTestId('slider-chassis').fill('30');
    await page.getByTestId('slider-engine').fill('30');

    await page.getByTestId('confirm-race').click();
  }

  // The season resolves locally in the browser via the Go WASM module.
  const complete = page.getByTestId('season-complete');
  await expect(complete).toBeVisible({ timeout: 30_000 });

  const share = await page.getByTestId('share-string').innerText();
  expect(share).toMatch(/^Lights Out · Season \d+/m);
  expect(share.split('\n')).toHaveLength(3);

  const name = `e2e-${Date.now()}`;
  await page.getByTestId('display-name').fill(name);
  await page.getByTestId('submit-run').click();

  await expect(page.getByTestId('submitted')).toBeVisible();

  const board = page.getByTestId('leaderboard');
  await expect(board).toBeVisible();
  await expect(board.getByText(name)).toBeVisible();

  expect(failures, `console/page errors:\n${failures.join('\n')}`).toEqual([]);
});

test('replaying the same day is rejected for the same browser', async ({ page }) => {
  // Each Playwright test gets a fresh context, so localStorage starts empty
  // and this browser is a new player. Both submissions therefore have to
  // happen inside THIS context for the player UUID to be shared.
  const playSeason = async () => {
    for (let round = 1; round <= 10; round++) {
      await expect(page.getByText(new RegExp(`Round ${round} of 10`))).toBeVisible();
      await page.getByTestId('slider-aero').fill('50');
      await page.getByTestId('confirm-race').click();
    }
    await expect(page.getByTestId('season-complete')).toBeVisible({ timeout: 30_000 });
  };

  await page.goto('/');
  await playSeason();
  await page.getByTestId('display-name').fill('replayer');
  await page.getByTestId('submit-run').click();
  await expect(page.getByTestId('submitted')).toBeVisible();

  // Reload and play again. The UUID in localStorage survives, so the
  // server's UNIQUE (season_id, player_id) constraint must reject this.
  await page.reload();
  await playSeason();
  await page.getByTestId('display-name').fill('replayer');
  await page.getByTestId('submit-run').click();

  await expect(page.getByText(/already submitted/i)).toBeVisible();
});
