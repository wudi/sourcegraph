package inventory

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

// Tree computes the inventory of languages for a tree. It caches the inventories of subtrees.
func (c *Context) Tree(ctx context.Context, tree os.FileInfo) (inv Inventory, err error) {
	// Get and set from the cache.
	if c.CacheGet != nil {
		if inv, ok := c.CacheGet(tree); ok {
			return inv, nil // cache hit
		}
	}
	if c.CacheSet != nil {
		defer func() {
			if err == nil {
				c.CacheSet(tree, inv) // store in cache
			}
		}()
	}

	entries, err := c.ReadTree(ctx, tree.Name())
	if err != nil {
		return Inventory{}, err
	}
	langTotalBytes := map[string]uint64{} // language name -> total bytes
	for _, e := range entries {
		switch {
		case e.Mode().IsRegular(): // file
			lang, err := detect(ctx, e, c.ReadFile)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory file %q", e.Name())
			}
			if lang != "" {
				langTotalBytes[lang] += uint64(e.Size())
			}

		case e.Mode().IsDir(): // subtree
			entryInv, err := c.Tree(ctx, e)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory tree %q", e.Name())
			}
			for _, lang := range entryInv.Languages {
				langTotalBytes[lang.Name] += lang.TotalBytes
			}

		default:
			// Skip symlinks, submodules, etc.
		}
	}
	return sum(langTotalBytes), nil
}
