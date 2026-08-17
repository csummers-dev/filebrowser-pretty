package files

import (
	"testing"
)

func names(l Listing) []string {
	out := make([]string, len(l.Items))
	for i, it := range l.Items {
		out[i] = it.Name
	}
	return out
}

// bySize must not order folders by their raw inode Size (which is meaningless —
// the real recursive size is fetched lazily per-folder), and must instead fall
// back to name so the folder group is predictable. Files still sort by size.
func TestApplySortBySizeFoldersFallBackToName(t *testing.T) {
	l := Listing{
		Items: []*FileInfo{
			// Folders with arbitrary (near-constant) inode sizes, out of name order.
			{Name: "charlie", Size: 4096, IsDir: true},
			{Name: "alpha", Size: 4096, IsDir: true},
			{Name: "bravo", Size: 128, IsDir: true},
			// Files with real, distinct sizes.
			{Name: "big.bin", Size: 900, IsDir: false},
			{Name: "small.bin", Size: 10, IsDir: false},
		},
		Sorting: Sorting{By: "size", Asc: true},
	}
	l.ApplySort()

	// Folders first (grouped, name order), then files ascending by size.
	got := names(l)
	want := []string{"alpha", "bravo", "charlie", "small.bin", "big.bin"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("bySize asc = %v, want %v", got, want)
		}
	}
}

// Files still sort strictly by size (regression guard for the non-folder path).
func TestApplySortBySizeFilesBySize(t *testing.T) {
	l := Listing{
		Items: []*FileInfo{
			{Name: "a", Size: 30, IsDir: false},
			{Name: "b", Size: 5, IsDir: false},
			{Name: "c", Size: 12, IsDir: false},
		},
		Sorting: Sorting{By: "size", Asc: true},
	}
	l.ApplySort()

	got := names(l)
	want := []string{"b", "c", "a"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("bySize files asc = %v, want %v", got, want)
		}
	}
}
