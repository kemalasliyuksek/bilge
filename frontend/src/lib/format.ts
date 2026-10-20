const units = ["B", "kB", "MB", "GB", "TB"];

// Decimal units (1 kB = 1000 B) match what `docker system df` prints, so
// numbers in bilge and in the terminal agree.
export function formatBytes(bytes: number): string {
  if (bytes < 0) {
    return "unknown";
  }
  let value = bytes;
  let unit = 0;
  while (value >= 1000 && unit < units.length - 1) {
    value /= 1000;
    unit++;
  }
  const digits = unit === 0 || value >= 100 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[unit]}`;
}

// Image and container IDs are long digests; 12 hex characters is what the
// docker CLI shows and is unique in practice.
export function shortId(id: string): string {
  const hex = id.startsWith("sha256:") ? id.slice("sha256:".length) : id;
  return hex.slice(0, 12);
}

const relative = new Intl.RelativeTimeFormat("en", { numeric: "auto" });
const steps: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 365 * 24 * 3600],
  ["month", 30 * 24 * 3600],
  ["day", 24 * 3600],
  ["hour", 3600],
  ["minute", 60],
];

// formatAge turns an ISO timestamp from the engine into "3 days ago". Engines
// report the zero time for unknown dates, which becomes an empty string.
export function formatAge(iso: string | undefined, now: Date = new Date()): string {
  if (!iso || iso.startsWith("0001-")) {
    return "";
  }
  const seconds = (new Date(iso).getTime() - now.getTime()) / 1000;
  if (Number.isNaN(seconds)) {
    return "";
  }
  for (const [unit, size] of steps) {
    if (Math.abs(seconds) >= size) {
      return relative.format(Math.round(seconds / size), unit);
    }
  }
  return "just now";
}

const digest = /@sha256:([0-9a-f]{12})[0-9a-f]{52}$/;

// shortRef keeps image references readable. "name:tag@sha256:…" carries the
// tag already, so the digest is dropped; a digest-only reference keeps a
// 12-character digest so it stays distinguishable.
export function shortRef(ref: string): string {
  const match = ref.match(digest);
  if (!match) {
    return ref;
  }
  const name = ref.slice(0, match.index);
  const lastSegment = name.slice(name.lastIndexOf("/") + 1);
  return lastSegment.includes(":") ? name : `${name}@sha256:${match[1]}`;
}
