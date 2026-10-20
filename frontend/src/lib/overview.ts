import type { main } from "../../wailsjs/go/models";
import { statusIndex } from "./status";

export const composeLabel = "com.docker.compose.project";

export type Share = { kept: number; review: number; safe: number };
export type Breakdown = { id: string; label: string } & Share;

const pos = (n: number) => Math.max(n, 0);

// diskBreakdown splits each storage category into what stays, what needs
// review and what is safe to remove. Sizes of candidates come from the cleanup
// report, so shared image layers count as kept, exactly as cleanup decided.
export function diskBreakdown(result: main.ScanResult | undefined): Breakdown[] {
  const inv = result?.inventory;
  const candidates = new Map(
    (result?.report.candidates ?? []).map((c) => [`${c.kind}:${c.id}`, c] as const),
  );
  const add = (share: Share, kind: string, id: string, total: number) => {
    const c = candidates.get(`${kind}:${id}`);
    if (!c) {
      share.kept += pos(total);
      return;
    }
    const freed = Math.min(pos(c.size), pos(total));
    share[c.verdict === "safe" ? "safe" : "review"] += freed;
    share.kept += pos(total) - freed;
  };

  const images: Share = { kept: 0, review: 0, safe: 0 };
  for (const i of inv?.images ?? []) add(images, "image", i.id, i.size);

  const volumes: Share = { kept: 0, review: 0, safe: 0 };
  for (const v of inv?.volumes ?? []) add(volumes, "volume", v.name, v.size);

  const containers: Share = { kept: 0, review: 0, safe: 0 };
  for (const c of inv?.containers ?? []) add(containers, "container", c.id, c.sizeRw);

  // Shared build cache records are layers that images own too; they are
  // already in the images bar, and counting them here would show the same
  // bytes twice. Build cache is one candidate in the report; the rest of the
  // unshared records are in use.
  const cache: Share = { kept: 0, review: 0, safe: 0 };
  const cacheTotal = (inv?.buildCache ?? [])
    .filter((r) => !r.shared)
    .reduce((t, r) => t + pos(r.size), 0);
  const cacheFreed = pos(candidates.get("buildCache:build-cache")?.size ?? 0);
  cache.safe = Math.min(cacheFreed, cacheTotal);
  cache.kept = cacheTotal - cache.safe;

  return [
    { id: "images", label: "Images", ...images },
    { id: "buildCache", label: "Build cache", ...cache },
    { id: "volumes", label: "Volumes", ...volumes },
    { id: "containers", label: "Container layers", ...containers },
  ];
}

export type Project = {
  name: string;
  running: number;
  containers: number;
  volumes: number;
  volumeSize: number;
  networks: number;
  reclaimable: number;
};

// projects groups resources by Compose project label. A project whose
// containers are all gone but whose volumes remain shows up with zero
// containers: that is the leftover bilge exists to surface.
export function projects(result: main.ScanResult | undefined): Project[] {
  const inv = result?.inventory;
  const byName = new Map<string, Project>();
  const get = (labels: Record<string, string> | null | undefined) => {
    const name = labels?.[composeLabel];
    if (!name) return undefined;
    let p = byName.get(name);
    if (!p) {
      p = {
        name,
        running: 0,
        containers: 0,
        volumes: 0,
        volumeSize: 0,
        networks: 0,
        reclaimable: 0,
      };
      byName.set(name, p);
    }
    return p;
  };
  for (const c of inv?.containers ?? []) {
    const p = get(c.labels);
    if (!p) continue;
    p.containers++;
    if (c.state === "running") p.running++;
  }
  for (const v of inv?.volumes ?? []) {
    const p = get(v.labels);
    if (!p) continue;
    p.volumes++;
    p.volumeSize += pos(v.size);
  }
  for (const n of inv?.networks ?? []) {
    const p = get(n.labels);
    if (p) p.networks++;
  }
  for (const c of result?.report.candidates ?? []) {
    if (c.project) {
      const p = byName.get(c.project);
      if (p) p.reclaimable += pos(c.size);
    }
  }
  return [...byName.values()].sort((a, b) => a.name.localeCompare(b.name));
}

export type Item = { kind: string; name: string; size: number; status: string };

// largest lists the biggest images and volumes, the usual suspects when a
// disk fills up.
export function largest(result: main.ScanResult | undefined, n: number): Item[] {
  const inv = result?.inventory;
  const status = statusIndex(result?.report);
  const items: Item[] = [
    ...(inv?.images ?? []).map((i) => ({
      kind: "Image",
      name: (i.tags ?? []).find((t) => t !== "<none>:<none>") ?? i.id,
      size: i.size,
      status: status("image", i.id),
    })),
    ...(inv?.volumes ?? []).map((v) => ({
      kind: "Volume",
      name: v.name,
      size: v.size,
      status: status("volume", v.name),
    })),
  ];
  return items.sort((a, b) => b.size - a.size).slice(0, n);
}
