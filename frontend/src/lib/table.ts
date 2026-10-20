export type SortValue = string | number;
export type SortDir = "asc" | "desc";

// matchesQuery is true when every whitespace-separated term appears in at
// least one field, so "pg running" finds a running postgres container.
export function matchesQuery(query: string, fields: (string | null | undefined)[]): boolean {
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
  if (terms.length === 0) {
    return true;
  }
  const haystack = fields.filter(Boolean).join("\n").toLowerCase();
  return terms.every((t) => haystack.includes(t));
}

// sortRows returns a sorted copy. Equal keys keep their input order, so a
// secondary order chosen by the caller survives.
export function sortRows<T>(rows: T[], key: (row: T) => SortValue, dir: SortDir): T[] {
  const sign = dir === "asc" ? 1 : -1;
  return rows
    .map((row, index) => ({ row, index, k: key(row) }))
    .sort((a, b) => {
      const c =
        typeof a.k === "number" && typeof b.k === "number"
          ? a.k - b.k
          : String(a.k).localeCompare(String(b.k), undefined, { numeric: true });
      return c === 0 ? a.index - b.index : c * sign;
    })
    .map((x) => x.row);
}
