package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

// An enricher modifies the row before output.
type enricher struct {
	FieldName string
	// Deps is the set of keys required in the row to enrich the
	// data the first time.
	Deps []string
	// CacheDeps is the set of keys to associate with each cache entry.
	CacheDeps []string
	Run       func(ctx context.Context, row map[string]string) (string, error)
}

func readCommit(repoName string, commitHash string) (*object.Commit, error) {
	if commitHash == "" {
		return nil, fmt.Errorf("ref is empty")
	}

	var (
		gitURI  = fmt.Sprintf("https://github.com/%s.git", repoName)
		refSpec = config.RefSpec(fmt.Sprintf("%s:%s", commitHash, commitHash))
	)

	// Use deterministic storage directory to speedup commit retrieval on common repos.
	storageDir := filepath.Join(os.TempDir(), "github-enricher", repoName)
	err := os.MkdirAll(storageDir, 0750)
	if err != nil {
		return nil, err
	}
	fs := osfs.New(storageDir)

	storage := filesystem.NewStorage(fs, cache.NewObjectLRU(cache.GiByte*4))
	c := &config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{gitURI},
		Fetch: []config.RefSpec{refSpec},
	}
	r := git.NewRemote(storage, c)
	fmt.Fprintf(os.Stderr, "shallow cloning %s to %s\n", gitURI, storageDir)
	err = r.Fetch(&git.FetchOptions{
		Depth:    1,
		RefSpecs: []config.RefSpec{refSpec},
		Progress: os.Stderr,
	})
	if err != git.NoErrAlreadyUpToDate && err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	commit, err := object.GetCommit(storage, plumbing.NewHash(commitHash))
	if err != nil {
		return nil, fmt.Errorf("get commit object: %w", err)
	}
	return commit, nil
}

var allEnrichers = []enricher{
	{
		FieldName: "email",
		Deps:      []string{"repo_name", "commit"},
		CacheDeps: []string{"commit"},
		Run: func(ctx context.Context, row map[string]string) (string, error) {
			commit, err := readCommit(row["repo_name"], row["commit"])
			if err != nil {
				return "", err
			}
			return commit.Author.Email, nil
		},
	},
	{
		FieldName: "name",
		Deps:      []string{"repo_name", "commit"},
		CacheDeps: []string{"commit"},
		Run: func(ctx context.Context, row map[string]string) (string, error) {
			commit, err := readCommit(row["repo_name"], row["commit"])
			if err != nil {
				return "", err
			}
			return commit.Author.Name, nil
		},
	},
}
