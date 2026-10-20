import type { cleanup } from "../../wailsjs/go/models";

export type Status = "safe" | "review" | "in use";

// statusIndex maps every resource to its cleanup verdict. The Go side is the
// single source of truth: anything it did not list as a candidate is in use,
// so the UI never re-implements the rules.
export function statusIndex(
  report: cleanup.Report | undefined,
): (kind: string, id: string) => Status {
  const verdicts = new Map<string, Status>();
  for (const c of report?.candidates ?? []) {
    verdicts.set(`${c.kind}:${c.id}`, c.verdict === "safe" ? "safe" : "review");
  }
  return (kind, id) => verdicts.get(`${kind}:${id}`) ?? "in use";
}

// countBy counts how many items reference each key, for "used by N
// containers" columns.
export function countBy<T>(
  items: T[] | null | undefined,
  keys: (item: T) => string[] | null | undefined,
): Map<string, number> {
  const counts = new Map<string, number>();
  for (const item of items ?? []) {
    for (const key of keys(item) ?? []) {
      counts.set(key, (counts.get(key) ?? 0) + 1);
    }
  }
  return counts;
}
