import { describe, expect, it } from "vitest";
import type { cleanup } from "../../wailsjs/go/models";
import { countBy, statusIndex } from "./status";

describe("statusIndex", () => {
  const report = {
    candidates: [
      { kind: "volume", id: "old_db", verdict: "review" },
      { kind: "image", id: "sha:1", verdict: "safe" },
    ],
  } as cleanup.Report;

  it("returns the verdict for candidates and in use for everything else", () => {
    const status = statusIndex(report);
    expect(status("volume", "old_db")).toBe("review");
    expect(status("image", "sha:1")).toBe("safe");
    expect(status("volume", "shop_db")).toBe("in use");
  });

  it("does not mix up kinds that share an id", () => {
    expect(statusIndex(report)("container", "sha:1")).toBe("in use");
  });

  it("handles a null candidate list from Go", () => {
    const status = statusIndex({ candidates: null } as unknown as cleanup.Report);
    expect(status("image", "x")).toBe("in use");
  });
});

describe("countBy", () => {
  it("counts references and tolerates null lists from Go", () => {
    const containers = [{ volumes: ["a", "b"] }, { volumes: ["a"] }, { volumes: null }];
    const counts = countBy(containers, (c) => c.volumes);
    expect(counts.get("a")).toBe(2);
    expect(counts.get("b")).toBe(1);
    expect(counts.get("c")).toBeUndefined();
  });
});
