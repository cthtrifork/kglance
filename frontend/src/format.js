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
 * kind is "ac" (no battery), "battery", "charging" or "full"; percent is null when there is no bar.
 */
export function describePower(state) {
  if (!state?.present) {
    return { kind: "ac", title: "On AC", detail: "No battery detected", percent: null };
  }
  const percent = Math.max(0, Math.min(100, Math.round(state.percentage)));
  const title = `${percent}%`;
  const left = formatDuration(state.secondsLeft);
  switch (state.status) {
    case "charging":
      return { kind: "charging", title, detail: left ? `Charging · ${left} until full` : "Charging", percent };
    case "discharging":
      return { kind: "battery", title, detail: left ? `On battery · ${left} left` : "On battery", percent };
    case "fully-charged":
      return { kind: "full", title, detail: "Fully charged", percent };
    case "empty":
      return { kind: "battery", title, detail: "Empty", percent };
    case "pending-charge":
    case "pending-discharge":
      return { kind: "battery", title, detail: "Not charging", percent };
    default:
      return { kind: "battery", title, detail: "", percent };
  }
}
