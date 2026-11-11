import assert from "node:assert/strict";
import { test } from "node:test";
import { batteryBars, batteryTone, describeMemory, describePower, describeUpdate, formatAge, formatDuration } from "./format.js";

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
  assert.deepEqual(view, {
    kind: "battery",
    title: "72%",
    detail: "On battery · 3 h 12 min left",
    percent: 72,
    tone: "good",
    bars: 3,
  });
});

test("charging (mockup 2)", () => {
  const view = describePower({ present: true, percentage: 38, status: "charging", secondsLeft: 3900 });
  assert.deepEqual(view, {
    kind: "charging",
    title: "38%",
    detail: "Charging · 1 h 05 min until full",
    percent: 38,
    tone: "charge",
    bars: 2,
  });
});

test("desktop without a battery (mockup 3)", () => {
  const view = describePower({ present: false, percentage: 0, status: "unknown", secondsLeft: 0 });
  assert.deepEqual(view, {
    kind: "ac",
    title: "On AC",
    detail: "No battery detected",
    percent: null,
    tone: "charge",
    bars: 0,
  });
  assert.equal(describePower(undefined).kind, "ac");
});

test("other states", () => {
  assert.equal(describePower({ present: true, percentage: 100, status: "fully-charged", secondsLeft: 0 }).detail, "Fully charged");
  assert.equal(describePower({ present: true, percentage: 80, status: "pending-charge", secondsLeft: 0 }).detail, "Not charging");
  assert.equal(describePower({ present: true, percentage: 40, status: "discharging", secondsLeft: 0 }).detail, "On battery");
  assert.equal(describePower({ present: true, percentage: 99.6, status: "charging", secondsLeft: 0 }).percent, 100);
});

test("battery bars and colour", () => {
  const bands = [
    [100, 3, "good"],
    [70, 3, "good"],
    [69, 2, "caution"],
    [30, 2, "caution"],
    [29, 1, "warn"],
    [0, 1, "warn"],
  ];
  for (const [percent, bars, tone] of bands) {
    assert.equal(batteryBars(percent), bars, `${percent}% bars`);
    assert.equal(batteryTone(percent), tone, `${percent}% tone`);
  }
  const on = (percentage, status) => describePower({ present: true, percentage, status, secondsLeft: 0 });
  // On a battery the colour follows the bars: green 3, orange 2, red 1.
  for (const status of ["discharging", "pending-charge", "pending-discharge", "unknown"]) {
    assert.deepEqual([on(80, status).bars, on(80, status).tone], [3, "good"], status);
    assert.deepEqual([on(45, status).bars, on(45, status).tone], [2, "caution"], status);
    assert.deepEqual([on(10, status).bars, on(10, status).tone], [1, "warn"], status);
  }
  // Charging is blue at every level but red below 30%; the bars still show the level. Fully charged is blue too.
  for (const [percent, bars, tone] of [
    [95, 3, "charge"],
    [70, 3, "charge"],
    [45, 2, "charge"],
    [30, 2, "charge"],
    [29, 1, "warn"],
    [0, 1, "warn"],
  ]) {
    assert.deepEqual([on(percent, "charging").bars, on(percent, "charging").tone], [bars, tone], `charging ${percent}%`);
  }
  assert.deepEqual([on(95, "fully-charged").bars, on(95, "fully-charged").tone], [3, "charge"]);
  assert.equal(on(10, "fully-charged").tone, "charge");
  // From 95% a charging battery counts as fully charged.
  assert.deepEqual(
    [on(95, "charging").kind, on(95, "charging").detail, on(94, "charging").kind, on(94, "charging").detail],
    ["full", "Fully charged", "charging", "Charging"],
  );
  assert.equal(on(97, "discharging").detail, "On battery");
  assert.equal(on(90, "empty").tone, "warn");
});

const GiB = 2 ** 30;

test("memory (mockup values)", () => {
  const view = describeMemory({ totalBytes: 15.6 * GiB, usedBytes: 6.1 * GiB, swapUsedBytes: 0 });
  assert.deepEqual(view, { title: "6.1 GiB", detail: "of 15.6 GiB used · 39 %", percent: 39, high: false });
});

test("memory nearly full, with swap", () => {
  const view = describeMemory({ totalBytes: 15.6 * GiB, usedBytes: 14.2 * GiB, swapUsedBytes: 1.1 * GiB });
  assert.deepEqual(view, { title: "14.2 GiB", detail: "of 15.6 GiB used · 91 % · swap 1.1 GiB", percent: 91, high: true });
});

test("memory threshold and unknown state", () => {
  assert.equal(describeMemory({ totalBytes: 100, usedBytes: 85, swapUsedBytes: 0 }).high, true);
  assert.equal(describeMemory({ totalBytes: 100, usedBytes: 84, swapUsedBytes: 0 }).high, false);
  assert.equal(describeMemory({ totalBytes: 0, usedBytes: 0, swapUsedBytes: 0 }), null);
  assert.equal(describeMemory(undefined), null);
});

test("formatAge", () => {
  assert.equal(formatAge(10), "now");
  assert.equal(formatAge(12 * 60), "12m");
  assert.equal(formatAge(2 * 3600 + 5), "2h");
  assert.equal(formatAge(47 * 3600), "47h");
  assert.equal(formatAge(3 * 86400), "3d");
  assert.equal(formatAge(0), "");
  assert.equal(formatAge(-5), "");
});

test("OS update pill", () => {
  const now = 1000000;
  assert.equal(describeUpdate(null, now).kind, "unknown");
  assert.equal(describeUpdate({ known: false }, now).text, "OS update status unavailable");
  const staged = { known: true, staged: true, version: "latest.20261111", timestamp: now - 7200 };
  assert.deepEqual(
    { ...describeUpdate(staged, now), title: "" },
    { kind: "staged", text: "OS update staged · latest.20261111", detail: "Built 2h ago · reboot to apply", title: "" },
  );
  assert.equal(describeUpdate({ ...staged, timestamp: now - 5 }, now).detail, "Built just now · reboot to apply");
  assert.equal(describeUpdate({ ...staged, version: "", timestamp: 0 }, now).text, "OS update staged");
  assert.equal(describeUpdate({ ...staged, version: "", timestamp: 0 }, now).detail, "Reboot to apply");
  const none = { known: true, staged: false, version: "latest.20261110", timestamp: now - 86400 * 3 };
  assert.equal(describeUpdate(none, now).text, "No OS update staged · built 3d ago");
  assert.equal(describeUpdate(none, now).detail, "");
  assert.equal(describeUpdate({ known: true, staged: false, version: "", timestamp: 0 }, now).text, "No OS update staged");
});
