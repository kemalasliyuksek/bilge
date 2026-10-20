import { describe, expect, it } from "vitest";
import { formatAge, formatBytes, shortId, shortRef } from "./format";

describe("formatBytes", () => {
  it("uses decimal units like docker system df", () => {
    expect(formatBytes(0)).toBe("0 B");
    expect(formatBytes(999)).toBe("999 B");
    expect(formatBytes(1000)).toBe("1.0 kB");
    expect(formatBytes(211_400_000)).toBe("211 MB");
    expect(formatBytes(6_587_000_000)).toBe("6.6 GB");
  });

  it("reports sizes the engine did not compute", () => {
    expect(formatBytes(-1)).toBe("unknown");
  });
});

describe("shortId", () => {
  it("drops the digest prefix and keeps 12 characters like the docker CLI", () => {
    expect(shortId("sha256:3f1c5d0e8a9b7c6d5e4f3a2b1c0d")).toBe("3f1c5d0e8a9b");
    expect(shortId("abcdef0123456789")).toBe("abcdef012345");
  });
});

describe("formatAge", () => {
  const now = new Date("2026-10-20T12:00:00Z");

  it("formats past timestamps relative to now", () => {
    expect(formatAge("2026-10-17T12:00:00Z", now)).toBe("3 days ago");
    expect(formatAge("2026-10-20T11:00:00Z", now)).toBe("1 hour ago");
    expect(formatAge("2026-10-20T11:59:50Z", now)).toBe("just now");
  });

  it("returns an empty string for unknown dates", () => {
    expect(formatAge("0001-01-01T00:00:00Z", now)).toBe("");
    expect(formatAge("", now)).toBe("");
    expect(formatAge("not a date", now)).toBe("");
  });
});

describe("shortRef", () => {
  const hex = "9dede0e3ccc071cf71935b17f76bf243331df0b1575338c8ac294640fcf12a36";

  it("drops the digest when the reference already has a tag", () => {
    expect(shortRef(`timescale/timescaledb:2.30.1-pg18@sha256:${hex}`)).toBe(
      "timescale/timescaledb:2.30.1-pg18",
    );
  });

  it("shortens a digest-only reference", () => {
    expect(shortRef(`timescale/timescaledb@sha256:${hex}`)).toBe(
      "timescale/timescaledb@sha256:9dede0e3ccc0",
    );
  });

  it("does not mistake a registry port for a tag", () => {
    expect(shortRef(`localhost:5000/app@sha256:${hex}`)).toBe(
      "localhost:5000/app@sha256:9dede0e3ccc0",
    );
  });

  it("leaves plain references alone", () => {
    expect(shortRef("postgres:16-alpine")).toBe("postgres:16-alpine");
  });
});
