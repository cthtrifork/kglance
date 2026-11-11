import assert from "node:assert/strict";
import { test } from "node:test";
import { describePower, formatDuration } from "./format.js";

test("formatDuration", () => {
  assert.equal(formatDuration(11520), "3 h 12 min");
  assert.equal(formatDuration(3900), "1 h 05 min");
  assert.equal(formatDuration(2520), "42 min");
  assert.equal(formatDuration(5), "1 min");
  assert.equal(formatDuration(0), "");
  assert.equal(formatDuration(-1), "");
  assert.equal(formatDuration(Number.NaN), "");
});

test("on battery (mockup 1)", () => {
  const view = describePower({ present: true, percentage: 72, status: "discharging", secondsLeft: 11520 });
  assert.deepEqual(view, { kind: "battery", title: "72%", detail: "On battery · 3 h 12 min left", percent: 72, low: false });
});

test("charging (mockup 2)", () => {
  const view = describePower({ present: true, percentage: 38, status: "charging", secondsLeft: 3900 });
  assert.deepEqual(view, { kind: "charging", title: "38%", detail: "Charging · 1 h 05 min until full", percent: 38, low: false });
});

test("desktop without a battery (mockup 3)", () => {
  const view = describePower({ present: false, percentage: 0, status: "unknown", secondsLeft: 0 });
  assert.deepEqual(view, { kind: "ac", title: "On AC", detail: "No battery detected", percent: null, low: false });
  assert.equal(describePower(undefined).kind, "ac");
});

test("other states", () => {
  assert.equal(describePower({ present: true, percentage: 100, status: "fully-charged", secondsLeft: 0 }).detail, "Fully charged");
  assert.equal(describePower({ present: true, percentage: 80, status: "pending-charge", secondsLeft: 0 }).detail, "Not charging");
  assert.equal(describePower({ present: true, percentage: 40, status: "discharging", secondsLeft: 0 }).detail, "On battery");
  assert.equal(describePower({ present: true, percentage: 15, status: "discharging", secondsLeft: 600 }).low, true);
  assert.equal(describePower({ present: true, percentage: 16, status: "discharging", secondsLeft: 600 }).low, false);
  assert.equal(describePower({ present: true, percentage: 5, status: "charging", secondsLeft: 600 }).low, false);
  assert.equal(describePower({ present: true, percentage: 99.6, status: "charging", secondsLeft: 0 }).percent, 100);
});
