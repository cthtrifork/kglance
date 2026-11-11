// Fixed states for `vite dev` (see main.js): the power states of the design mockups, each with a memory state.
const GiB = 2 ** 30;
const memory = {
  battery: { totalBytes: 15.6 * GiB, usedBytes: 6.1 * GiB, swapUsedBytes: 0 },
  charging: { totalBytes: 15.6 * GiB, usedBytes: 9.5 * GiB, swapUsedBytes: 0 },
  low: { totalBytes: 15.6 * GiB, usedBytes: 8.1 * GiB, swapUsedBytes: 0 },
  ac: { totalBytes: 15.6 * GiB, usedBytes: 14.2 * GiB, swapUsedBytes: 1.1 * GiB },
};

const demos = {
  battery: { present: true, percentage: 72, status: "discharging", secondsLeft: 11520 },
  charging: { present: true, percentage: 38, status: "charging", secondsLeft: 3900 },
  low: { present: true, percentage: 9, status: "discharging", secondsLeft: 1500 },
  ac: { present: false, percentage: 0, status: "unknown", secondsLeft: 0 },
};

/** The state named by ?demo=<name> in a query string, or null. */
export function demoState(search) {
  return demos[new URLSearchParams(search).get("demo")] ?? null;
}

/** The memory state that goes with ?demo=<name>, or null. */
export function demoMemory(search) {
  return memory[new URLSearchParams(search).get("demo")] ?? null;
}

/** The OS update state that goes with ?update=staged, or null. */
export function demoUpdate(search) {
  return new URLSearchParams(search).get("update") === "staged" ? { staged: true, version: "latest.20261111" } : null;
}
