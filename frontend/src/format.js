// Pure helpers for the page; no DOM, so `node --test` can run them.

/** "3 h 12 min", "1 h 05 min", "42 min"; "" when the time is unknown. */
export function formatDuration(seconds) {
  if (!Number.isFinite(seconds) || seconds <= 0) return "";
  const minutes = Math.max(1, Math.round(seconds / 60));
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return h > 0 ? `${h} h ${String(m).padStart(2, "0")} min` : `${m} min`;
}

/**
 * What the power row shows for a state from the Go side ({present, percentage, status, secondsLeft}).
 * kind is "ac" (no battery), "battery", "charging" or "full"; percent is null when there is no bar;
 * low is true on a battery at or below LOW_PERCENT.
 */
const LOW_PERCENT = 15;

export function describePower(state) {
  if (!state?.present) {
    return { kind: "ac", title: "On AC", detail: "No battery detected", percent: null, low: false };
  }
  const percent = Math.max(0, Math.min(100, Math.round(state.percentage)));
  const title = `${percent}%`;
  const left = formatDuration(state.secondsLeft);
  const low = percent <= LOW_PERCENT;
  switch (state.status) {
    case "charging":
      return { kind: "charging", title, detail: left ? `Charging · ${left} until full` : "Charging", percent, low: false };
    case "discharging":
      return { kind: "battery", title, detail: left ? `On battery · ${left} left` : "On battery", percent, low };
    case "fully-charged":
      return { kind: "full", title, detail: "Fully charged", percent, low: false };
    case "empty":
      return { kind: "battery", title, detail: "Empty", percent, low: true };
    case "pending-charge":
    case "pending-discharge":
      return { kind: "battery", title, detail: "Not charging", percent, low: false };
    default:
      return { kind: "battery", title, detail: "", percent, low: false };
  }
}

const GIB = 2 ** 30;
const HIGH_MEMORY_PERCENT = 85;
const gib = (bytes) => (bytes / GIB).toFixed(1);

/**
 * What the memory row shows for a state from the Go side ({totalBytes, usedBytes, swapUsedBytes}), or null
 * when the state is unknown (nothing is shown then). high is true at HIGH_MEMORY_PERCENT or more.
 */
export function describeMemory(state) {
  if (!state?.totalBytes) return null;
  const percent = Math.max(0, Math.min(100, Math.round((state.usedBytes / state.totalBytes) * 100)));
  const swap = state.swapUsedBytes > 0 ? ` · swap ${gib(state.swapUsedBytes)} GiB` : "";
  return {
    title: `${gib(state.usedBytes)} GiB`,
    detail: `of ${gib(state.totalBytes)} GiB used · ${percent} %${swap}`,
    percent,
    high: percent >= HIGH_MEMORY_PERCENT,
  };
}
