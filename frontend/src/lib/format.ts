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
