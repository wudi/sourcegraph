package graphqlbackend

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/src-d/enry/v2"
)

func (srs *searchResultsStats) Languages(ctx context.Context) ([]*LanguageStatistics, error) {
	srr, err := srs.getResults(ctx)
	if err != nil {
		return nil, err
	}

	getFiles := func(ctx context.Context, repo gitserver.Repo, commitID api.CommitID, path string) ([]os.FileInfo, error) {
		return git.ReadDir(ctx, repo, commitID, "", true)
	}

	type repoCommit struct {
		repo     api.RepoID
		commitID api.CommitID
	}
	var (
		repos          = map[api.RepoID]*types.Repo{}
		filesMap       = map[repoCommit][]os.FileInfo{}
		allInventories []inventory.Inventory
	)
	sawRepo := func(repo *types.Repo) {
		if _, ok := repos[repo.ID]; !ok {
			repos[repo.ID] = repo
		}
	}
	for _, res := range srr.Results() {
		if fileMatch, ok := res.ToFileMatch(); ok {
			sawRepo(fileMatch.Repository().repo)
			key := repoCommit{repo: fileMatch.Repository().repo.ID, commitID: fileMatch.commitID}
			if fileMatch.File().IsDirectory() {
				repo := gitserver.Repo{Name: fileMatch.Repository().repo.Name}
				treeFiles, err := getFiles(ctx, repo, fileMatch.commitID, fileMatch.JPath)
				if err != nil {
					return nil, err
				}
				filesMap[key] = append(filesMap[key], treeFiles...)
			} else {
				var lines int64
				if len(fileMatch.LineMatches()) > 0 {
					lines = int64(len(fileMatch.LineMatches()))
				} else {
					content, err := fileMatch.File().Content(ctx)
					if err != nil {
						return nil, err
					}
					lines = int64(strings.Count(content, "\n"))
				}
				filesMap[key] = append(filesMap[key], &fileInfo{path: fileMatch.JPath, isDir: fileMatch.File().IsDirectory(), size: lines})
			}
		} else if repo, ok := res.ToRepository(); ok {
			sawRepo(repo.repo)
			// TODO!(sqs): dedupe if a repo matches and some of its files/diffs match, so we don't
			// double-count.
			branchRef, err := repo.DefaultBranch(ctx)
			if err != nil {
				return nil, err
			}
			if branchRef == nil || branchRef.Target() == nil {
				continue
			}
			target, err := branchRef.Target().OID(ctx)
			if err != nil {
				return nil, err
			}
			inv, err := backend.Repos.GetInventory(ctx, repo.repo, api.CommitID(target))
			if err != nil {
				return nil, err
			}
			{
				// TODO!(sqs): hack adjust for lines
				for _, l := range inv.Languages {
					l.TotalBytes = l.TotalBytes / 31
				}
			}
			allInventories = append(allInventories, *inv)
		} else if commit, ok := res.ToCommitSearchResult(); ok {
			if commit.raw.Diff == nil {
				continue
			}
			sawRepo(commit.commit.Repository().repo)
			key := repoCommit{repo: commit.commit.Repository().repo.ID, commitID: api.CommitID(commit.commit.oid)}
			fileDiffs, err := diff.ParseMultiFileDiff([]byte(commit.raw.Diff.Raw))
			if err != nil {
				return nil, err
			}
			for _, fileDiff := range fileDiffs {
				var lines int64
				for _, hunk := range fileDiff.Hunks {
					c := int64(hunk.NewLines - hunk.OrigLines)
					if c < 0 {
						c = c * -1
					}
					lines += c
				}
				filesMap[key] = append(filesMap[key], &fileInfo{path: fileDiff.NewName, isDir: false, size: lines})
			}
		}
	}

	for key, files := range filesMap {
		cachedRepo, err := backend.CachedGitRepo(ctx, repos[key.repo])
		if err != nil {
			return nil, err
		}
		invCtx, err := backend.InventoryContext(*cachedRepo, key.commitID)
		if err != nil {
			return nil, err
		}
		inv, err := invCtx.Files(ctx, files)
		if err != nil {
			return nil, err
		}
		allInventories = append(allInventories, inv)
	}

	byLang := map[string]inventory.Lang{}
	for _, inv := range allInventories {
		for _, lang := range inv.Languages {
			langInv, ok := byLang[lang.Name]
			if ok {
				langInv.TotalBytes += lang.TotalBytes
			} else {
				byLang[lang.Name] = lang
			}
		}
	}

	langStats := make([]*LanguageStatistics, 0, len(byLang))
	for _, langInv := range byLang {
		langStats = append(langStats, &LanguageStatistics{langInv})
	}
	sort.Slice(langStats, func(i, j int) bool {
		return langStats[i].Lang.TotalBytes > langStats[j].Lang.TotalBytes
	})
	return langStats, nil
}

type LanguageStatistics struct{ inventory.Lang }

func (v *LanguageStatistics) Name() string      { return v.Lang.Name }
func (v *LanguageStatistics) TotalBytes() int32 { return int32(v.Lang.TotalBytes) }
func (v *LanguageStatistics) Type() string      { return typeToString(enry.GetLanguageType(v.Lang.Name)) }

func typeToString(t enry.Type) string {
	switch t {
	case enry.Data:
		return "data"
	case enry.Programming:
		return "programming"
	case enry.Markup:
		return "markup"
	case enry.Prose:
		return "prose"
	default:
		return ""
	}
}
