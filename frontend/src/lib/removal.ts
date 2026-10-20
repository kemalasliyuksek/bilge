import type { cleanup, sweep } from "../../wailsjs/go/models";

export const kindLabel: Record<string, string> = {
  container: "Container",
  image: "Image",
  volume: "Volume",
  network: "Network",
  buildCache: "Build cache",
};

// A named volume is the one kind of candidate that can hold data nobody can
// recreate, so it is the one the user must confirm by typing its name.
export function needsTypedConfirm(c: Pick<cleanup.Candidate, "kind" | "verdict">): boolean {
  return c.kind === "volume" && c.verdict === "review";
}

export function requestFor(c: cleanup.Candidate): sweep.Request {
  return { kind: c.kind, id: c.id, verdict: c.verdict };
}

export function resultKey(r: { kind: string; id: string }): string {
  return `${r.kind}:${r.id}`;
}

export type Summary = { removed: number; skipped: number; failed: number; freed: number };

export function summarize(results: sweep.Result[]): Summary {
  const s: Summary = { removed: 0, skipped: 0, failed: 0, freed: 0 };
  for (const r of results) {
    if (r.status === "removed") {
      s.removed++;
      s.freed += Math.max(r.freed, 0);
    } else if (r.status === "skipped") {
      s.skipped++;
    } else {
      s.failed++;
    }
  }
  return s;
}

// The build cache is opt-in: removing it loses no data, but the next builds
// that relied on it start from scratch, so the user picks it explicitly.
export function selectedByDefault(c: Pick<cleanup.Candidate, "kind">): boolean {
  return c.kind !== "buildCache";
}

// toggled holds the items whose selection differs from their default. A
// rescan then keeps the defaults for new items and the user's choices for
// known ones.
export function isSelected(
  c: Pick<cleanup.Candidate, "kind" | "id">,
  toggled: Set<string>,
): boolean {
  return selectedByDefault(c) !== toggled.has(resultKey(c));
}
