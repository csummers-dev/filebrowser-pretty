import { describe, it, expect } from "vitest";
import { sortListing } from "@/utils/secondarySort";

// sortListing only reads name/size/modified/extension/isDir off each item.
const row = (name: string, size: number, modified: string, extension = "") =>
  ({ name, size, modified, extension }) as unknown as ResourceItem;

// A folder row: its `size` is the (meaningless) inode size, so size-sort must
// ignore it and fall back to name. FileListing sorts dirs and files as
// separate groups, so a folder group never mixes with files here.
const dir = (name: string, size: number) =>
  ({
    name,
    size,
    modified: "",
    extension: "",
    isDir: true,
  }) as unknown as ResourceItem;

// SortCriterion shorthand.
const crit = (by: SortKey, asc: boolean): SortCriterion => ({ by, asc });

describe("sortListing", () => {
  it("returns the same array for fewer than 2 items", () => {
    const items = [row("a", 1, "x")];
    expect(sortListing(items, crit("name", true), null)).toBe(items);
  });

  it("does not mutate the input array", () => {
    const items = [row("b", 1, "x"), row("a", 1, "y")];
    const snapshot = items.map((r) => r.name);
    sortListing(items, crit("name", true), null);
    expect(items.map((r) => r.name)).toEqual(snapshot);
  });

  it("sorts by the primary key ascending", () => {
    const out = sortListing(
      [row("b", 1, "x"), row("a", 1, "y"), row("c", 1, "z")],
      crit("name", true),
      null
    );
    expect(out.map((r) => r.name)).toEqual(["a", "b", "c"]);
  });

  it("sorts by the primary key descending", () => {
    const out = sortListing(
      [row("a", 1, "x"), row("c", 1, "y"), row("b", 1, "z")],
      crit("name", false),
      null
    );
    expect(out.map((r) => r.name)).toEqual(["c", "b", "a"]);
  });

  it("sorts by size numerically (not lexically)", () => {
    const out = sortListing(
      [row("a", 30, "x"), row("b", 5, "y"), row("c", 12, "z")],
      crit("size", true),
      null
    );
    expect(out.map((r) => r.name)).toEqual(["b", "c", "a"]);
  });

  it("sorts folders by name when sorting by size with no resolver (inode size is meaningless)", () => {
    // Folder inode sizes are unrelated to content size; with no folder-size
    // resolver, size-sort falls back to name so folders aren't arbitrary.
    const out = sortListing(
      [dir("charlie", 4096), dir("alpha", 4096), dir("bravo", 128)],
      crit("size", true),
      null
    );
    expect(out.map((r) => r.name)).toEqual(["alpha", "bravo", "charlie"]);
  });

  it("sorts folders by their resolved recursive size when a resolver is given", () => {
    // Inode sizes are identical/meaningless; the resolver supplies real sizes.
    const items = [
      dir("alpha", 4096),
      dir("bravo", 4096),
      dir("charlie", 4096),
    ];
    const real: Record<string, number> = { alpha: 5, bravo: 900, charlie: 50 };
    const resolver = (it: ResourceItem) => real[it.name];

    const asc = sortListing(items, crit("size", true), null, resolver);
    expect(asc.map((r) => r.name)).toEqual(["alpha", "charlie", "bravo"]);

    const desc = sortListing(items, crit("size", false), null, resolver);
    expect(desc.map((r) => r.name)).toEqual(["bravo", "charlie", "alpha"]);
  });

  it("clusters folders with unresolved sizes (name-broken) ahead of known ones ascending", () => {
    const items = [
      dir("charlie", 4096),
      dir("alpha", 4096),
      dir("bravo", 4096),
    ];
    // Only bravo's size has resolved; alpha/charlie are still calculating.
    const resolver = (it: ResourceItem) =>
      it.name === "bravo" ? 900 : undefined;

    const out = sortListing(items, crit("size", true), null, resolver);
    // Unknown (-1) cluster first, name-broken (alpha, charlie), then bravo.
    expect(out.map((r) => r.name)).toEqual(["alpha", "charlie", "bravo"]);
  });

  it("orders names naturally (file_2 before file_10)", () => {
    const out = sortListing(
      [row("file_10", 1, "x"), row("file_2", 1, "y")],
      crit("name", true),
      null
    );
    expect(out.map((r) => r.name)).toEqual(["file_2", "file_10"]);
  });

  it("applies the secondary as a tiebreaker within primary ties", () => {
    const a = row("a", 10, "2020-03-01");
    const b = row("b", 10, "2020-01-01");
    const c = row("c", 20, "2020-02-01");
    // primary size asc → {a,b} tie at 10 then c at 20; secondary modified
    // asc breaks the tie: b before a.
    const out = sortListing(
      [a, b, c],
      crit("size", true),
      crit("modified", true)
    );
    expect(out.map((r) => r.name)).toEqual(["b", "a", "c"]);
  });

  it("honors a descending secondary order", () => {
    const a = row("a", 10, "2020-01-01");
    const b = row("b", 10, "2020-03-01");
    const out = sortListing(
      [a, b],
      crit("size", true),
      crit("modified", false)
    );
    expect(out.map((r) => r.name)).toEqual(["b", "a"]);
  });

  it("treats a null secondary as no tiebreaker (stable input order on ties)", () => {
    const a = row("a", 10, "2020-03-01");
    const b = row("b", 10, "2020-01-01");
    // size ties, no secondary → keeps incoming order.
    const out = sortListing([a, b], crit("size", true), null);
    expect(out.map((r) => r.name)).toEqual(["a", "b"]);
  });

  it("ignores a secondary equal to the primary key (no-op tiebreaker)", () => {
    const a = row("a", 10, "x");
    const b = row("b", 10, "y");
    const out = sortListing([a, b], crit("size", true), crit("size", true));
    expect(out.map((r) => r.name)).toEqual(["a", "b"]);
  });

  it("is stable when both criteria tie (keeps input order)", () => {
    const a = row("a", 10, "2020-01-01");
    const b = row("b", 10, "2020-01-01");
    const out = sortListing([a, b], crit("size", true), crit("modified", true));
    expect(out.map((r) => r.name)).toEqual(["a", "b"]);
  });

  it("breaks a name tie by extension secondary", () => {
    const a = row("file", 1, "x", "zip");
    const b = row("file", 1, "x", "7z");
    const out = sortListing(
      [a, b],
      crit("name", true),
      crit("extension", true)
    );
    expect(out.map((r) => r.extension)).toEqual(["7z", "zip"]);
  });

  it("treats an unparseable modified timestamp as a tie (no NaN comparator)", () => {
    // A malformed date must not make Array.sort's order undefined: it should
    // fall through to the stable (input) order.
    const a = row("a", 5, "not-a-date");
    const b = row("b", 5, "2020-01-01");
    const out = sortListing([a, b], crit("modified", true), null);
    expect(out.map((r) => r.name)).toEqual(["a", "b"]);
  });

  it("does not reorder for an unknown sort key (stale pref)", () => {
    const a = row("b", 1, "x");
    const b = row("a", 1, "y");
    // Cast: simulate a persisted pref naming a removed sort field.
    const out = sortListing(
      [a, b],
      { by: "smart" as unknown as SortKey, asc: true },
      null
    );
    expect(out.map((r) => r.name)).toEqual(["b", "a"]);
  });
});
