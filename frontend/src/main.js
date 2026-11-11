import { Events } from "@wailsio/runtime";
import { Hide, Info } from "../bindings/github.com/cthtrifork/kglance/internal/popup/service.js";

const backdrop = document.getElementById("backdrop");
const card = document.getElementById("card");
const shortcut = document.getElementById("shortcut");
const shortcutHint = document.getElementById("shortcut-hint");

// Close on Escape, and on a click outside the card (the card swallows its own clicks).
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") Hide();
});
backdrop.addEventListener("click", (event) => {
  if (!card.contains(event.target)) Hide();
});

// Whenever the popup is shown, ask how it runs: the shortcut text may have been rebound meanwhile.
async function refresh() {
  const info = await Info();
  shortcut.textContent = info.shortcut;
  // The user may have cleared the shortcut in System Settings; then there is nothing to advertise.
  shortcutHint.hidden = info.shortcut === "";
  document.documentElement.classList.toggle("layer", info.layer);
}

Events.On("popup:shown", refresh);
refresh();
