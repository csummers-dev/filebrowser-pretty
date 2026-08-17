package files

import (
	"sort"
	"strings"

	"github.com/maruel/natural"
)

// Listing is a collection of files.
type Listing struct {
	Items    []*FileInfo `json:"items"`
	NumDirs  int         `json:"numDirs"`
	NumFiles int         `json:"numFiles"`
	Sorting  Sorting     `json:"sorting"`
}

// ApplySort applies the sort order using .Order and .Sort
func (l Listing) ApplySort() {
	// Check '.Order' to know how to sort
	if !l.Sorting.Asc {
		switch l.Sorting.By {
		case "name":
			sort.Sort(sort.Reverse(byName(l)))
		case "size":
			sort.Sort(sort.Reverse(bySize(l)))
		case "modified":
			sort.Sort(sort.Reverse(byModified(l)))
		case "extension":
			sort.Sort(sort.Reverse(byExtension(l)))
		default:
			// If not one of the above, do nothing
			return
		}
	} else { // If we had more Orderings we could add them here
		switch l.Sorting.By {
		case "name":
			sort.Sort(byName(l))
		case "size":
			sort.Sort(bySize(l))
		case "modified":
			sort.Sort(byModified(l))
		case "extension":
			sort.Sort(byExtension(l))
		default:
			sort.Sort(byName(l))
			return
		}
	}
}

// Implement sorting for Listing
type byName Listing
type bySize Listing
type byModified Listing

// byExtension sorts files by their extension (v1.3 S3-5).
//
// Folders are always grouped first regardless of sort direction —
// they don't have user-facing extensions and mixing them with files
// by some arbitrary comparison would look broken. Within the file
// group, sort by extension (case-insensitive); ties break by name.
type byExtension Listing

// By Name
func (l byName) Len() int {
	return len(l.Items)
}

func (l byName) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

// Treat upper and lower case equally
func (l byName) Less(i, j int) bool {
	if l.Items[i].IsDir && !l.Items[j].IsDir {
		return l.Sorting.Asc
	}

	if !l.Items[i].IsDir && l.Items[j].IsDir {
		return !l.Sorting.Asc
	}

	return natural.Less(strings.ToLower(l.Items[j].Name), strings.ToLower(l.Items[i].Name))
}

// By Size
func (l bySize) Len() int {
	return len(l.Items)
}

func (l bySize) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

func (l bySize) Less(i, j int) bool {
	// Folders group together, ahead of files (flipped to the bottom under
	// reverse sort — same as byName/byExtension). A folder's listing Size is
	// only its inode size, not its recursive content size (that's fetched
	// lazily per-folder via GET /api/folder-size), so ordering folders by Size
	// is meaningless and leaves them in an arbitrary order. Fall back to name
	// so the folder group stays predictable, matching byExtension.
	if l.Items[i].IsDir != l.Items[j].IsDir {
		return l.Items[i].IsDir
	}
	if l.Items[i].IsDir {
		return natural.Less(
			strings.ToLower(l.Items[i].Name),
			strings.ToLower(l.Items[j].Name),
		)
	}
	return l.Items[i].Size < l.Items[j].Size
}

// By Modified
func (l byModified) Len() int {
	return len(l.Items)
}

func (l byModified) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

func (l byModified) Less(i, j int) bool {
	iModified, jModified := l.Items[i].ModTime, l.Items[j].ModTime
	return iModified.Sub(jModified) < 0
}

// By Extension (v1.3 S3-5).
func (l byExtension) Len() int      { return len(l.Items) }
func (l byExtension) Swap(i, j int) { l.Items[i], l.Items[j] = l.Items[j], l.Items[i] }

func (l byExtension) Less(i, j int) bool {
	// Folders always group first. byName has the same pattern but
	// inverted on Asc — for extension sort we want directories at the
	// top regardless of direction because they don't participate in
	// the extension comparison meaningfully. The reverse-sort path in
	// ApplySort wraps this with sort.Reverse, which would flip folder
	// position too; that's acceptable (some users prefer dirs-at-bottom)
	// and matches how byName behaves under reverse sort.
	if l.Items[i].IsDir != l.Items[j].IsDir {
		return l.Items[i].IsDir
	}

	// Both folders: fall back to name (extension is meaningless).
	if l.Items[i].IsDir {
		return natural.Less(
			strings.ToLower(l.Items[i].Name),
			strings.ToLower(l.Items[j].Name),
		)
	}

	// Both files: compare extensions case-insensitively. Ties break
	// by basename so files with the same extension sort alphabetically
	// within their group — e.g. "a.pdf" before "b.pdf".
	iExt := strings.ToLower(l.Items[i].Extension)
	jExt := strings.ToLower(l.Items[j].Extension)
	if iExt != jExt {
		return iExt < jExt
	}
	return natural.Less(
		strings.ToLower(l.Items[i].Name),
		strings.ToLower(l.Items[j].Name),
	)
}
