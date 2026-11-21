import { describe, expect, it } from "vitest";
import type { sweep } from "../../wailsjs/go/models";
import { bulkSelectable, isSelected, needsTypedConfirm, setAll, summarize } from "./removal";

describe("named volumes", () => {
  it("need the typed name and stay out of bulk selection", () => {
    expect(needsTypedConfirm({ confirmByName: true })).toBe(true);
    expect(bulkSelectable({ confirmByName: true })).toBe(false);
    expect(needsTypedConfirm({ confirmByName: false })).toBe(false);
    expect(bulkSelectable({ confirmByName: false })).toBe(true);
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

describe("isSelected", () => {
  const image = { kind: "image", id: "sha:1", verdict: "safe" };
  const cache = { kind: "buildCache", id: "build-cache", verdict: "safe" };
  const tagged = { kind: "image", id: "sha:2", verdict: "review" };
  const stack = { kind: "container", id: "c1", verdict: "review" };

  it("selects safe items but the build cache by default", () => {
    expect(isSelected(image, new Set())).toBe(true);
    expect(isSelected(cache, new Set())).toBe(false);
  });

  it("selects no review item by default", () => {
    expect(isSelected(tagged, new Set())).toBe(false);
    expect(isSelected(stack, new Set())).toBe(false);
  });

  it("flips the default for toggled items", () => {
    const toggled = new Set(["image:sha:1", "buildCache:build-cache", "image:sha:2"]);
    expect(isSelected(image, toggled)).toBe(false);
    expect(isSelected(cache, toggled)).toBe(true);
    expect(isSelected(tagged, toggled)).toBe(true);
  });
});

describe("setAll", () => {
  const tagged = { kind: "image", id: "sha:2", verdict: "review" };
  const stack = { kind: "container", id: "c1", verdict: "review" };

  it("selects only the items it is given", () => {
    // The image filter is on: the container is hidden and must stay unselected.
    const next = setAll([tagged], new Set(), true);
    expect(isSelected(tagged, next)).toBe(true);
    expect(isSelected(stack, next)).toBe(false);
  });

  it("clears the given items and keeps hidden choices", () => {
    const both = setAll([tagged, stack], new Set(), true);
    const next = setAll([tagged], both, false);
    expect(isSelected(tagged, next)).toBe(false);
    expect(isSelected(stack, next)).toBe(true);
  });
});
