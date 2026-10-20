import { describe, expect, it } from "vitest";
import type { main } from "../../wailsjs/go/models";
import { diskBreakdown, largest, projects } from "./overview";

const compose = (p: string) => ({ "com.docker.compose.project": p });

const result = {
  inventory: {
    containers: [
      { id: "c1", state: "running", sizeRw: 10, labels: compose("shop") },
      { id: "c2", state: "exited", sizeRw: 5, labels: null },
    ],
    images: [
      { id: "i-used", tags: ["app:1"], size: 100 },
      { id: "i-dangling", tags: null, size: 50 },
      { id: "i-tagged", tags: ["old:1"], size: 40 },
    ],
    volumes: [
      { name: "shop_db", size: 300, labels: compose("shop") },
      { name: "gone_db", size: 200, labels: compose("gone") },
      { name: "unknown", size: -1, labels: null },
    ],
    networks: [{ id: "n1", name: "shop_default", labels: compose("shop") }],
    buildCache: [
      { id: "b1", size: 70, inUse: false, shared: false },
      { id: "b2", size: 30, inUse: false, shared: true },
      { id: "b3", size: 5, inUse: true, shared: false },
    ],
  },
  report: {
    candidates: [
      { kind: "container", id: "c2", verdict: "safe", size: 5, project: "" },
      { kind: "image", id: "i-dangling", verdict: "safe", size: 30, project: "" },
      { kind: "image", id: "i-tagged", verdict: "review", size: 40, project: "" },
      { kind: "volume", id: "gone_db", verdict: "review", size: 200, project: "gone" },
      { kind: "buildCache", id: "build-cache", verdict: "safe", size: 70, project: "" },
    ],
  },
} as unknown as main.ScanResult;

describe("diskBreakdown", () => {
  const byId = Object.fromEntries(diskBreakdown(result).map((b) => [b.id, b]));

  it("counts shared image layers as kept", () => {
    // i-dangling frees 30 of its 50; the other 20 stay.
    expect(byId.images).toMatchObject({ kept: 120, review: 40, safe: 30 });
  });

  it("splits volumes and ignores unknown sizes", () => {
    expect(byId.volumes).toMatchObject({ kept: 300, review: 200, safe: 0 });
  });

  it("leaves shared build cache to the images bar", () => {
    expect(byId.buildCache).toMatchObject({ kept: 5, review: 0, safe: 70 });
  });

  it("splits container writable layers", () => {
    expect(byId.containers).toMatchObject({ kept: 10, review: 0, safe: 5 });
  });

  it("is all zero without a scan", () => {
    for (const b of diskBreakdown(undefined)) {
      expect(b.kept + b.review + b.safe).toBe(0);
    }
  });
});

describe("projects", () => {
  it("groups by Compose project and surfaces projects with only leftovers", () => {
    expect(projects(result)).toEqual([
      {
        name: "gone",
        running: 0,
        containers: 0,
        volumes: 1,
        volumeSize: 200,
        networks: 0,
        reclaimable: 200,
      },
      {
        name: "shop",
        running: 1,
        containers: 1,
        volumes: 1,
        volumeSize: 300,
        networks: 1,
        reclaimable: 0,
      },
    ]);
  });
});

describe("largest", () => {
  it("lists the biggest images and volumes with their status", () => {
    expect(largest(result, 3)).toEqual([
      { kind: "Volume", name: "shop_db", size: 300, status: "in use" },
      { kind: "Volume", name: "gone_db", size: 200, status: "review" },
      { kind: "Image", name: "app:1", size: 100, status: "in use" },
    ]);
  });
});
