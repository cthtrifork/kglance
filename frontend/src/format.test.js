import assert from "node:assert/strict";
import { test } from "node:test";
import { batteryLevel, describeMemory, describePower, describeUpdate, formatAge, formatDuration } from "./format.js";

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
    low: false,
    level: "over-half",
  });
});

test("charging (mockup 2)", () => {
  const view = describePower({ present: true, percentage: 38, status: "charging", secondsLeft: 3900 });
  assert.deepEqual(view, {
    kind: "charging",
    title: "38%",
    detail: "Charging · 1 h 05 min until full",
    percent: 38,
    low: false,
    level: "under-half",
  });
});

test("desktop without a battery (mockup 3)", () => {
  const view = describePower({ present: false, percentage: 0, status: "unknown", secondsLeft: 0 });
  assert.deepEqual(view, { kind: "ac", title: "On AC", detail: "No battery detected", percent: null, low: false, level: "none" });
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

test("battery level bands", () => {
  const bands = [
    [100, "full"],
    [80, "full"],
    [79, "over-half"],
    [60, "over-half"],
    [59, "half"],
    [40, "half"],
    [39, "under-half"],
    [16, "under-half"],
    [15, "low"],
    [0, "low"],
  ];
  for (const [percent, level] of bands) assert.equal(batteryLevel(percent), level, `${percent}%`);
  assert.equal(describePower({ present: true, percentage: 90, status: "charging", secondsLeft: 0 }).level, "full");
  assert.equal(describePower({ present: true, percentage: 50, status: "pending-discharge", secondsLeft: 0 }).level, "half");
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
