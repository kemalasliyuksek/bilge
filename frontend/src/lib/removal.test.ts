import { describe, expect, it } from "vitest";
import type { sweep } from "../../wailsjs/go/models";
import { needsTypedConfirm, summarize } from "./removal";

describe("needsTypedConfirm", () => {
  it("asks for the name only for named volumes under review", () => {
    expect(needsTypedConfirm({ kind: "volume", verdict: "review" })).toBe(true);
    expect(needsTypedConfirm({ kind: "volume", verdict: "safe" })).toBe(false);
    expect(needsTypedConfirm({ kind: "image", verdict: "review" })).toBe(false);
  });
});

describe("summarize", () => {
  it("counts outcomes and adds up only what was removed", () => {
    const results = [
      { status: "removed", freed: 100 },
      { status: "removed", freed: -1 },
      { status: "skipped", freed: 50 },
      { status: "failed", freed: 70 },
    ] as sweep.Result[];
    expect(summarize(results)).toEqual({ removed: 2, skipped: 1, failed: 1, freed: 100 });
  });
});
