<script lang="ts" module>
import type { SortValue } from "../lib/table";

export type Column<T> = {
  id: string;
  label: string;
  // Initial width in px. The last column has none and takes the rest.
  width?: number;
  align?: "right";
  // Fixed columns keep their width; no resize handle.
  fixed?: boolean;
  title?: string;
  sort?: (row: T) => SortValue;
};
</script>

<script lang="ts" generics="T">
import { type Snippet, untrack } from "svelte";
import { loadPref, savePref } from "../lib/prefs";
import { type SortDir, sortRows } from "../lib/table";

type Props = {
  // Key for remembering widths and sort order across sessions.
  id: string;
  columns: Column<T>[];
  rows: T[];
  key: (row: T) => string;
  row: Snippet<[T]>;
  sortBy?: string;
  sortDir?: SortDir;
};

let { id, columns, rows, key, row, sortBy = "", sortDir = "asc" }: Props = $props();

const minWidth = 48;
// Columns, id and the initial sort are read once: they define the table, and
// the user's later changes live in widths and sort.
const defaults = untrack(() => Object.fromEntries(columns.map((c) => [c.id, c.width ?? 0])));
const saved = untrack(() => loadPref(`table:${id}`, { widths: defaults, sortBy, sortDir }));

let widths = $state<Record<string, number>>({ ...defaults, ...saved.widths });
let sort = $state({ by: saved.sortBy, dir: saved.sortDir as SortDir });

const persist = () => savePref(`table:${untrack(() => id)}`, { widths, sortBy: sort.by, sortDir: sort.dir });

const sorted = $derived.by(() => {
  const col = columns.find((c) => c.id === sort.by);
  return col?.sort ? sortRows(rows, col.sort, sort.dir) : rows;
});

// The table is at least as wide as its fixed columns, so widening a column
// scrolls horizontally instead of squeezing the others.
const fixedWidth = $derived(columns.reduce((t, c) => t + (c.width ? widths[c.id] : 100), 0));

function toggleSort(col: Column<T>) {
  if (!col.sort) return;
  sort = sort.by === col.id ? { by: col.id, dir: sort.dir === "asc" ? "desc" : "asc" } : { by: col.id, dir: col.align === "right" ? "desc" : "asc" };
  persist();
}

function startResize(event: PointerEvent, col: Column<T>) {
  event.preventDefault();
  event.stopPropagation();
  const handle = event.currentTarget as HTMLElement;
  const startX = event.clientX;
  const startWidth = widths[col.id];
  handle.setPointerCapture(event.pointerId);
  const move = (e: PointerEvent) => {
    widths[col.id] = Math.max(minWidth, Math.round(startWidth + e.clientX - startX));
  };
  const up = () => {
    handle.removeEventListener("pointermove", move);
    handle.removeEventListener("pointerup", up);
    persist();
  };
  handle.addEventListener("pointermove", move);
  handle.addEventListener("pointerup", up);
}

function resetWidth(col: Column<T>) {
  widths[col.id] = defaults[col.id];
  persist();
}
</script>

<div class="scroller">
  <table class="data" style:width="max(100%, {fixedWidth}px)">
    <colgroup>
      {#each columns as col (col.id)}
        <col style:width={col.width ? `${widths[col.id]}px` : undefined} />
      {/each}
    </colgroup>
    <thead>
      <tr>
        {#each columns as col (col.id)}
          <th class:num={col.align === "right"} title={col.title} aria-sort={sort.by === col.id ? (sort.dir === "asc" ? "ascending" : "descending") : undefined}>
            {#if col.sort}
              <button type="button" class="sort" onclick={() => toggleSort(col)}>
                {col.label}
                <span class="arrow" aria-hidden="true">{sort.by === col.id ? (sort.dir === "asc" ? "▲" : "▼") : ""}</span>
              </button>
            {:else}
              {col.label}
            {/if}
            {#if col.width && !col.fixed}
              <button
                type="button"
                class="grip"
                aria-label="Resize {col.label}"
                title="Drag to resize, double-click to reset"
                onpointerdown={(e) => startResize(e, col)}
                ondblclick={() => resetWidth(col)}
              ></button>
            {/if}
          </th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each sorted as item (key(item))}
        <tr>{@render row(item)}</tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  .scroller {
    overflow-x: auto;
  }

  table {
    table-layout: fixed;
  }

  th {
    position: relative;
    user-select: none;
  }

  .sort {
    padding: 0;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    cursor: pointer;
  }

  .sort:hover {
    color: var(--fg);
  }

  .arrow {
    font-size: 0.6rem;
  }

  .grip {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: none;
    position: absolute;
    top: 0;
    right: -3px;
    bottom: 0;
    width: 7px;
    cursor: col-resize;
    z-index: 1;
  }

  .grip::after {
    content: "";
    position: absolute;
    top: 25%;
    bottom: 25%;
    left: 3px;
    width: 1px;
    background: var(--line);
  }

  th:hover .grip::after,
  .grip:active::after {
    background: var(--accent);
  }
</style>
