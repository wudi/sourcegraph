package backend

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var inventoryCache = rcache.New("inv")

// Feature flag for enhanced (but much slower) language detection that uses file contents, not just
// filenames.
var useEnhancedLanguageDetection, _ = strconv.ParseBool(os.Getenv("USE_ENHANCED_LANGUAGE_DETECTION"))

// InventoryContext returns the inventory context for computing the inventory for the repository at
// the given commit.
func InventoryContext(repo gitserver.Repo, commitID api.CommitID) (inventory.Context, error) {
	if !git.IsAbsoluteRevision(string(commitID)) {
		return inventory.Context{}, errors.Errorf("refusing to compute inventory for non-absolute commit ID %q", commitID)
	}

	cacheKey := func(e os.FileInfo) string {
		return e.Sys().(git.ObjectInfo).OID().String()
	}
	invCtx := inventory.Context{
		ReadTree: func(ctx context.Context, path string) ([]os.FileInfo, error) {
			// TODO: As a perf optimization, we could read multiple levels of the Git tree at once
			// to avoid sequential tree traversal calls.
			return git.ReadDir(ctx, repo, commitID, path, false)
		},
		ReadFile: func(ctx context.Context, path string, minBytes int64) ([]byte, error) {
			return git.ReadFile(ctx, repo, commitID, path, minBytes)
		},
		CacheGet: func(e os.FileInfo) (inventory.Inventory, bool) {
			if b, ok := inventoryCache.Get(cacheKey(e)); ok {
				var inv inventory.Inventory
				if err := json.Unmarshal(b, &inv); err != nil {
					log15.Warn("Failed to unmarshal cached JSON inventory.", "repo", repo.Name, "commitID", commitID, "path", e.Name(), "err", err)
					return inventory.Inventory{}, false
				}
				return inv, true
			}
			return inventory.Inventory{}, false
		},
		CacheSet: func(e os.FileInfo, inv inventory.Inventory) {
			b, err := json.Marshal(&inv)
			if err != nil {
				log15.Warn("Failed to marshal JSON inventory for cache.", "repo", repo.Name, "commitID", commitID, "path", e.Name(), "err", err)
				return
			}
			inventoryCache.Set(cacheKey(e), b)
		},
	}

	if !useEnhancedLanguageDetection {
		// If USE_ENHANCED_LANGUAGE_DETECTION is disabled, do not read file contents to determine
		// the language.
		invCtx.ReadFile = func(ctx context.Context, path string, minBytes int64) ([]byte, error) {
			return nil, nil
		}
	}

	return invCtx, nil
}
