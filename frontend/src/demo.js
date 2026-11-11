// Fixed states for `vite dev` (see main.js): the power states of the design mockups, each with a memory state.
const GiB = 2 ** 30;
const memory = {
  battery: { totalBytes: 15.6 * GiB, usedBytes: 6.1 * GiB, swapUsedBytes: 0 },
  charging: { totalBytes: 15.6 * GiB, usedBytes: 9.5 * GiB, swapUsedBytes: 0 },
  low: { totalBytes: 15.6 * GiB, usedBytes: 8.1 * GiB, swapUsedBytes: 0 },
  ac: { totalBytes: 15.6 * GiB, usedBytes: 14.2 * GiB, swapUsedBytes: 1.1 * GiB },
  full: { totalBytes: 15.6 * GiB, usedBytes: 6.1 * GiB, swapUsedBytes: 0 },
  "over-half": { totalBytes: 15.6 * GiB, usedBytes: 6.1 * GiB, swapUsedBytes: 0 },
  half: { totalBytes: 15.6 * GiB, usedBytes: 6.1 * GiB, swapUsedBytes: 0 },
  "under-half": { totalBytes: 15.6 * GiB, usedBytes: 6.1 * GiB, swapUsedBytes: 0 },
};

const demos = {
  battery: { present: true, percentage: 72, status: "discharging", secondsLeft: 11520 },
  charging: { present: true, percentage: 38, status: "charging", secondsLeft: 3900 },
  low: { present: true, percentage: 9, status: "discharging", secondsLeft: 1500 },
  ac: { present: false, percentage: 0, status: "unknown", secondsLeft: 0 },
  full: { present: true, percentage: 94, status: "discharging", secondsLeft: 16200 },
  "over-half": { present: true, percentage: 68, status: "discharging", secondsLeft: 11000 },
  half: { present: true, percentage: 50, status: "discharging", secondsLeft: 8400 },
  "under-half": { present: true, percentage: 27, status: "discharging", secondsLeft: 4500 },
};

/** The state named by ?demo=<name> in a query string, or null. */
export function demoState(search) {
  return demos[new URLSearchParams(search).get("demo")] ?? null;
}

/** The memory state that goes with ?demo=<name>, or null. */
export function demoMemory(search) {
  return memory[new URLSearchParams(search).get("demo")] ?? null;
}

const updates = {
  // 2026-11-10 20:50:47 UTC: two hours before the clock the screenshots fix.
  staged: { known: true, staged: true, version: "latest.20261111", timestamp: 1794343847 },
  none: { known: true, staged: false, version: "latest.20261110", timestamp: 1794247200 },
  unknown: { known: false, staged: false, version: "", timestamp: 0 },
};

/**
 * The OS update state named by &update=<name> (staged|none|unknown). Without it, the charging demo shows a
 * staged update and the others none, so the screenshots cover both without extra images.
 */
export function demoUpdate(search) {
  const params = new URLSearchParams(search);
  const name = params.get("update") ?? (params.get("demo") === "charging" ? "staged" : "none");
  return updates[name] ?? updates.none;
}
