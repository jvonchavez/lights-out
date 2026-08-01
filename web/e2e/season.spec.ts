import { test, expect } from '@playwright/test';

const WINDOWS = 5;

/** Take one card per window, waiting for each window to come round. */
async function playSeason(page: import('@playwright/test').Page, choose: (w: number) => number) {
  for (let w = 1; w <= WINDOWS; w++) {
    await expect(page.getByText(new RegExp(`Development window ${w} of ${WINDOWS}`))).toBeVisible({
      timeout: 30_000,
    });
    await page.getByTestId(`card-${choose(w)}`).click();
    await page.getByTestId('confirm-pick').click();
  }
  await expect(page.getByTestId('season-complete')).toBeVisible({ timeout: 30_000 });
}

/**
 * The one end-to-end test that matters: load the page, draft a car across
 * five windows against the WASM simulation, submit, appear on the board.
 */
test('draft a car, run the season, and appear on the leaderboard', async ({ page }) => {
  const failures: string[] = [];
  page.on('pageerror', (e) => failures.push(String(e)));
  page.on('console', (m) => {
    if (m.type() === 'error') failures.push(m.text());
  });

  await page.goto('/');

  // Deals render from the API before the WASM module has finished loading.
  await expect(page.getByTestId('card-0')).toBeVisible();
  await expect(page.getByTestId('card-2')).toBeVisible();

  // Confirm is disabled until a card is chosen: a window cannot be skipped.
  await expect(page.getByTestId('confirm-pick')).toBeDisabled();

  await playSeason(page, (w) => w % 3);

  // The build is the artifact the draft produces.
  const build = page.getByTestId('build');
  await expect(build).toBeVisible();
  await expect(build.locator('li')).toHaveCount(WINDOWS);

  const share = await page.getByTestId('share-string').innerText();
  expect(share).toMatch(/^Lights Out · Season \d+/m);
  expect(share).toMatch(/P\d+ of \d+ · \d+ pts/);
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
  await page.goto('/');
  await playSeason(page, () => 0);
  await page.getByTestId('display-name').fill('replayer');
  await page.getByTestId('submit-run').click();
  await expect(page.getByTestId('submitted')).toBeVisible();

  await page.reload();
  await playSeason(page, () => 0);
  await page.getByTestId('display-name').fill('replayer');
  await page.getByTestId('submit-run').click();

  await expect(page.getByText(/already submitted/i)).toBeVisible();
});

test('the default share stays spoiler-free and the build copy is opt-in', async ({ page }) => {
  await page.goto('/');
  await playSeason(page, () => 1);

  const share = await page.getByTestId('share-string').innerText();
  const parts = await page.getByTestId('build-part-name').allInnerTexts();

  expect(parts).toHaveLength(WINDOWS);

  // Naming your parts on a shared daily seed gives the day away, so the
  // default copy must not do it.
  for (const raw of parts) {
    const partName = raw.trim();
    // Guard: toContain('') is vacuously true and would hide a real leak.
    expect(partName.length).toBeGreaterThan(0);
    expect(share).not.toContain(partName);
  }
  await expect(page.getByTestId('copy-plain')).toBeVisible();
  await expect(page.getByTestId('copy-build')).toBeVisible();
});
