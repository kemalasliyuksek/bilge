import { describe, expect, it } from "vitest";
import { formatBytes } from "./format";

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
