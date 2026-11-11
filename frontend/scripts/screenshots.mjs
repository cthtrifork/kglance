// Renders the popup in every demo state in light, and in one state in dark, into the directory given as the first argument
// (default ../screenshots). It starts the Vite dev server itself, so the demo states (?demo=..., dev only)
// are available, and fixes the clock so the pictures do not change from run to run.
//
//   npm ci && npx playwright-core install --with-deps chromium && node scripts/screenshots.mjs
import { mkdirSync } from "node:fs";
import { resolve } from "node:path";
import { chromium } from "playwright-core";
import { createServer } from "vite";

const out = resolve(process.argv[2] ?? "../screenshots");
const states = ["battery", "charging", "low", "ac", "full", "over-half", "half", "under-half"];
// Every state in light; dark only for one example (charging, which also shows the staged OS update pill).
const darkStates = ["charging"];
// The time of the design mockups.
const now = new Date("2026-11-10T23:50:47+01:00");

mkdirSync(out, { recursive: true });

const server = await createServer({ server: { host: "127.0.0.1", port: 0 }, logLevel: "warn" });
await server.listen();
const base = server.resolvedUrls.local[0];

// A desktop-like backdrop, so the card and its shadow can be judged (the real window is transparent).
const backdrop = "html { background: linear-gradient(160deg, #2b3a52, #4a3f5f) !important; }";

const browser = await chromium.launch();
let failed = false;
try {
  for (const colorScheme of ["light", "dark"]) {
    const context = await browser.newContext({
      viewport: { width: 340, height: 400 },
      deviceScaleFactor: 2,
      colorScheme,
      locale: "en-GB",
      timezoneId: "Europe/Copenhagen",
    });
    for (const state of colorScheme === "dark" ? darkStates : states) {
      const page = await context.newPage();
      const errors = [];
      page.on("pageerror", (error) => errors.push(error.message));
      await page.clock.setFixedTime(now);
      await page.goto(`${base}?demo=${state}`);
      await page.waitForFunction(() => document.getElementById("power-title")?.textContent !== "");
      await page.addStyleTag({ content: backdrop });
      const file = `${out}/popup-${state}-${colorScheme}.png`;
      await page.screenshot({ path: file });
      console.log(file);
      if (errors.length) {
        failed = true;
        console.error(`${state}/${colorScheme}: page errors:\n  ${errors.join("\n  ")}`);
      }
      await page.close();
    }
    await context.close();
  }
} finally {
  await browser.close();
  await server.close();
}
if (failed) process.exit(1);
