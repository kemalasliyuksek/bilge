// Per-viewer conveniences only (column widths, sort order). Storage can be
// missing or throw in a webview, and the UI must work without it.
export function loadPref<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(`bilge:${key}`);
    return raw ? { ...fallback, ...JSON.parse(raw) } : fallback;
  } catch {
    return fallback;
  }
}

export function savePref(key: string, value: unknown): void {
  try {
    localStorage.setItem(`bilge:${key}`, JSON.stringify(value));
  } catch {
    // Not worth surfacing: the table still works, it just forgets.
  }
}
