import { describe, expect, it } from "vitest";
import { matchesQuery, sortRows } from "./table";

describe("matchesQuery", () => {
  it("matches every term against any field, case-insensitively", () => {
    expect(matchesQuery("PG run", ["postgres-pg18", "running"])).toBe(true);
    expect(matchesQuery("pg exited", ["postgres-pg18", "running"])).toBe(false);
  });

  it("accepts an empty query and skips missing fields", () => {
    expect(matchesQuery("  ", ["x"])).toBe(true);
    expect(matchesQuery("x", [null, undefined, "x"])).toBe(true);
  });
});

describe("sortRows", () => {
  const rows = [
    { name: "web-10", size: 3 },
    { name: "web-2", size: 1 },
    { name: "db", size: 3 },
  ];

  it("sorts numbers numerically in both directions", () => {
    expect(sortRows(rows, (r) => r.size, "desc").map((r) => r.name)).toEqual([
      "web-10",
      "db",
      "web-2",
    ]);
    expect(sortRows(rows, (r) => r.size, "asc").map((r) => r.name)).toEqual([
      "web-2",
      "web-10",
      "db",
    ]);
  });

  it("sorts strings with natural number order", () => {
    expect(sortRows(rows, (r) => r.name, "asc").map((r) => r.name)).toEqual([
      "db",
      "web-2",
      "web-10",
    ]);
  });

  it("does not mutate the input", () => {
    sortRows(rows, (r) => r.size, "asc");
    expect(rows[0].name).toBe("web-10");
  });
});
