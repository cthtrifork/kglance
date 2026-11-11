// Pure helpers for the page; no DOM, so `node --test` can run them.

/** "3 h 12 min", "1 h 05 min", "42 min"; "" when the time is unknown. */
export function formatDuration(seconds) {
  if (!Number.isFinite(seconds) || seconds <= 0) return "";
  const minutes = Math.max(1, Math.round(seconds / 60));
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return h > 0 ? `${h} h ${String(m).padStart(2, "0")} min` : `${m} min`;
}

const RED_BELOW = 30;
const GREEN_FROM = 70;
const FULL_FROM = 95;

/** The lit bars of the battery icon for a percentage: 3 from 70%, 2 from 30%, else 1. */
export function batteryBars(percent) {
  if (percent >= GREEN_FROM) return 3;
  if (percent >= RED_BELOW) return 2;
  return 1;
}

/** The colour of a battery that is running down: green from 70%, orange from 30%, red below. */
export function batteryTone(percent) {
  if (percent >= GREEN_FROM) return "good";
  if (percent >= RED_BELOW) return "caution";
  return "warn";
}

/**
 * What the power row shows for a state from the Go side ({present, percentage, status, secondsLeft}).
 * kind is "ac" (no battery), "battery", "charging" or "full"; percent is null when there is no bar.
 * bars (0 to 3) is how many bars the battery icon lights: 3 from 70%, 2 from 30%, else 1 (0 without a battery).
 * tone is the colour. Running on the battery it follows the bars: "good" (green, 3), "caution" (orange, 2) or "warn" (red, 1).
 * A charging battery from FULL_FROM (95%) counts as fully charged, as does the "fully-charged" status.
 * Charging, fully charged and no battery are "charge" (blue), but a charging battery below 30% is "warn" (red).
 */
export function describePower(state) {
  if (!state?.present) {
    return { kind: "ac", title: "On AC", detail: "No battery detected", percent: null, tone: "charge", bars: 0 };
  }
  const percent = Math.max(0, Math.min(100, Math.round(state.percentage)));
  const title = `${percent}%`;
  const left = formatDuration(state.secondsLeft);
  const bars = batteryBars(percent);
  const tone = batteryTone(percent);
  const chargingTone = tone === "warn" ? "warn" : "charge";
  const full = { kind: "full", title, detail: "Fully charged", percent, tone: "charge", bars };
  switch (state.status) {
    case "charging":
      if (percent >= FULL_FROM) return full;
      return {
        kind: "charging",
        title,
        detail: left ? `Charging · ${left} until full` : "Charging",
        percent,
        tone: chargingTone,
        bars,
      };
    case "discharging":
      return { kind: "battery", title, detail: left ? `On battery · ${left} left` : "On battery", percent, tone, bars };
    case "fully-charged":
      return full;
    case "empty":
      return { kind: "battery", title, detail: "Empty", percent, tone: "warn", bars };
    case "pending-charge":
    case "pending-discharge":
      return { kind: "battery", title, detail: "Not charging", percent, tone, bars };
    default:
      return { kind: "battery", title, detail: "", percent, tone, bars };
  }
}

/** "now", "12m", "2h", "3d": how long ago; "" when the time is unknown. */
export function formatAge(seconds) {
  if (!Number.isFinite(seconds) || seconds <= 0) return "";
  if (seconds < 60) return "now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  return hours < 48 ? `${hours}h` : `${Math.floor(hours / 24)}d`;
}

/**
 * What the OS update pill shows for a state from the Go side ({known, staged, version, timestamp}); now is
 * the current time in Unix seconds. kind is "staged" (a new image waits for a reboot), "none" (nothing
 * staged) or "unknown" (rpm-ostree could not be asked). text is the first line; detail is a second line,
 * only for a staged update ("" otherwise). The version is shown only for a staged update: the running image's is
 * not news.
 */
export function describeUpdate(state, now) {
  if (!state?.known) {
    return { kind: "unknown", text: "OS update status unavailable", detail: "", title: "rpm-ostree could not be read." };
  }
  const age = state.timestamp ? formatAge(now - state.timestamp) : "";
  if (state.staged) {
    const built = age === "now" ? "Built just now" : age ? `Built ${age} ago` : "";
    return {
      kind: "staged",
      text: state.version ? `OS update staged · ${state.version}` : "OS update staged",
      detail: [built, "reboot to apply"]
        .filter(Boolean)
        .join(" · ")
        .replace(/^./, (c) => c.toUpperCase()),
      title: "A new OS image is staged and applies on the next reboot.",
    };
  }
  const old = age === "now" ? " · built just now" : age ? ` · built ${age} ago` : "";
  return {
    kind: "none",
    text: `No OS update staged${old}`,
    detail: "",
    title: "Nothing is staged. The age is when the running image was built.",
  };
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
