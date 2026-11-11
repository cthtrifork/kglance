// Fixed power states for `vite dev` (see main.js): the three states of the design mockups.
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
