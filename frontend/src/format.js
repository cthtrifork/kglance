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
  const old = age && age !== "now" ? ` · image ${age} old` : "";
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
