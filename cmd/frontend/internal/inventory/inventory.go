package inventory

import (
	"context"
	"os"
	"sort"
)

// Inventory summarizes a tree's contents (e.g., which programming
// languages are used).
type Inventory struct {
	// Languages are the programming languages used in the tree.
	Languages []Lang `json:"Languages,omitempty"`
}

// Lang represents a programming language used in a directory tree.
type Lang struct {
	// Name is the name of a programming language (e.g., "Go" or
	// "Java").
	Name string `json:"Name,omitempty"`
	// TotalBytes is the total number of bytes of code written in the
	// programming language.
	TotalBytes uint64 `json:"TotalBytes,omitempty"`
}

func sum(langTotalBytes map[string]uint64) Inventory {
	sum := Inventory{Languages: make([]Lang, 0, len(langTotalBytes))}
	for lang, totalBytes := range langTotalBytes {
		sum.Languages = append(sum.Languages, Lang{Name: lang, TotalBytes: totalBytes})
	}
	sort.Slice(sum.Languages, func(i, j int) bool {
		return sum.Languages[i].TotalBytes > sum.Languages[j].TotalBytes || (sum.Languages[i].TotalBytes == sum.Languages[j].TotalBytes && sum.Languages[i].Name < sum.Languages[j].Name)
	})
	return sum
}

// Context defines the environment in which the inventory is computed.
type Context struct {
	// ReadTree is called to list the immediate children of a tree at path. The returned os.FileInfo
	// values' Name method must return the full path (that can be passed to another ReadTree or
	// ReadFile call), not just the basename.
	ReadTree func(ctx context.Context, path string) ([]os.FileInfo, error)

	// ReadFile is called to read the partial contents of the file at path. At least the specified
	// number of bytes must be returned (or the entire file, if it is smaller).
	ReadFile func(ctx context.Context, path string, minBytes int64) ([]byte, error)

	// CacheGet, if set, returns the cached inventory and true for the given tree, or false for a cache miss.
	CacheGet func(os.FileInfo) (Inventory, bool)

	// CacheSet, if set, stores the inventory in the cache for the given tree.
	CacheSet func(os.FileInfo, Inventory)
}
