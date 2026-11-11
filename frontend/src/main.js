import { Events } from "@wailsio/runtime";
import { State as MemoryState } from "../bindings/github.com/cthtrifork/kglance/internal/memory/service.js";
import { State as UpdateState } from "../bindings/github.com/cthtrifork/kglance/internal/osupdate/service.js";
import { Hide, Info } from "../bindings/github.com/cthtrifork/kglance/internal/popup/service.js";
import { State } from "../bindings/github.com/cthtrifork/kglance/internal/power/service.js";
import { demoMemory, demoState, demoUpdate } from "./demo.js";
import { describeMemory, describePower, describeUpdate } from "./format.js";

const $ = (id) => document.getElementById(id);
const card = $("card");

// Close on Escape, and on a click outside the card (the card swallows its own clicks).
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") Hide();
});
$("backdrop").addEventListener("click", (event) => {
  if (!card.contains(event.target)) Hide();
});

// Clock and date. Intl follows the system locale. The timer only runs while the popup is shown.
const hourMinute = new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" });
const second = new Intl.DateTimeFormat(undefined, { second: "2-digit" });
const date = new Intl.DateTimeFormat(undefined, { weekday: "long", day: "numeric", month: "long", year: "numeric" });

function renderClock() {
  const now = new Date();
  $("time").textContent = hourMinute.format(now);
  $("seconds").textContent = second.format(now);
  $("date").textContent = date.format(now);
}

let timer = null;
let ticks = 0;
function startClock() {
  stopClock();
  ticks = 0;
  const tick = () => {
    renderClock();
    if (ticks++ % 2 === 0 && !demo) refreshMemory();
    // Aim at the start of the next second so the display does not drift.
    timer = setTimeout(tick, 1000 - (Date.now() % 1000) + 5);
  };
  tick();
}
function stopClock() {
  clearTimeout(timer);
  timer = null;
}

// Power row.
function renderPower(state) {
  const view = describePower(state);
  $("power").dataset.kind = view.kind;
  $("power").dataset.tone = view.tone;
  $("power").dataset.bars = String(view.bars);
  $("power-title").textContent = view.title;
  $("power-detail").textContent = view.detail;
  $("power-bar").hidden = view.percent === null;
  $("power-fill").style.width = `${view.percent ?? 0}%`;
}

// Memory row. There is no signal for memory, so it is read every two seconds, only while the popup is shown.
function renderMemory(state) {
  const view = describeMemory(state);
  $("memory").hidden = view === null;
  $("memory-rule").hidden = view === null;
  if (view === null) return;
  $("memory").dataset.level = view.high ? "high" : "normal";
  $("memory-title").textContent = view.title;
  $("memory-detail").textContent = view.detail;
  $("memory-fill").style.width = `${view.percent}%`;
}

async function refreshMemory() {
  try {
    renderMemory(await MemoryState());
  } catch {
    // The row keeps its last value; the next tick tries again.
  }
}

// OS update pill. The host already knows whether an image is staged (rpm-ostree reads local state, nothing is
// fetched), so it is asked each time the popup is shown. The pill is always there, so nothing shifts when it fills in.
function renderUpdate(state) {
  const view = describeUpdate(state, Date.now() / 1000);
  $("update").dataset.kind = view.kind;
  $("update").title = view.title;
  $("update-text").textContent = view.text;
  $("update-detail").textContent = view.detail;
}

async function refreshUpdate() {
  try {
    renderUpdate(await UpdateState());
  } catch {
    // The pill keeps what it showed; the next time the popup is shown tries again.
  }
}

// How the popup runs, and the shortcut text, which may have been rebound meanwhile.
async function refreshInfo() {
  const info = await Info();
  $("shortcut").textContent = info.shortcut;
  // The user may have cleared the shortcut in System Settings; then there is nothing to advertise.
  $("shortcut-hint").hidden = info.shortcut === "";
  document.documentElement.classList.toggle("layer", info.layer);
}

// `vite dev` with ?demo=battery|charging|low|ac shows a fixed power state without the Go side
// (scripts/screenshots.mjs, design work). The branch is dead code in a production build.
const demo = import.meta.env.DEV ? demoState(location.search) : null;

if (!demo) {
  Events.On("popup:shown", () => {
    startClock();
    refreshInfo();
    refreshUpdate();
  });
  Events.On("popup:hidden", stopClock);
  Events.On("power:changed", (event) => renderPower(event.data));
}

renderClock();
if (demo) {
  renderPower(demo);
  renderMemory(demoMemory(location.search));
  renderUpdate(demoUpdate(location.search));
} else {
  refreshInfo();
  State().then(renderPower);
  refreshMemory();
  refreshUpdate();
}
