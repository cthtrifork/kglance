import { Events } from "@wailsio/runtime";
import { Hide, Info } from "../bindings/github.com/cthtrifork/kglance/internal/popup/service.js";
import { State } from "../bindings/github.com/cthtrifork/kglance/internal/power/service.js";
import { describePower } from "./format.js";

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
function startClock() {
  stopClock();
  const tick = () => {
    renderClock();
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
  $("power-title").textContent = view.title;
  $("power-detail").textContent = view.detail;
  $("power-bar").hidden = view.percent === null;
  $("power-fill").style.width = `${view.percent ?? 0}%`;
}

// How the popup runs, and the shortcut text, which may have been rebound meanwhile.
async function refreshInfo() {
  const info = await Info();
  $("shortcut").textContent = info.shortcut;
  // The user may have cleared the shortcut in System Settings; then there is nothing to advertise.
  $("shortcut-hint").hidden = info.shortcut === "";
  document.documentElement.classList.toggle("layer", info.layer);
}

Events.On("popup:shown", () => {
  startClock();
  refreshInfo();
});
Events.On("popup:hidden", stopClock);
Events.On("power:changed", (event) => renderPower(event.data));

renderClock();
refreshInfo();
State().then(renderPower);
