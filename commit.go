package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/coder/flog"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

type commit struct {
	email string
	name  string
}

func newGitHubClient(ctx context.Context) (*github.Client, error) {
	const envName = "GITHUB_TOKEN"
	token, ok := os.LookupEnv(envName)
	if !ok {
		return nil, fmt.Errorf("%q not set", envName)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, ts)
	client := github.NewClient(httpClient)

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	me, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get self: %w", err)
	}
	flog.Info("authenticated as %v", *me.Login)
	return client, nil
}

func readCommitAPI(ctx context.Context, cli *github.Client, repoName string, commitHash string) (*commit, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%v/commits/%v", repoName, commitHash), nil)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	var ct github.CommitResult
	resp, err := cli.Do(ctx, req, &ct)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer func() {
		if recover() != nil {
			fmt.Printf("commit: %+v\n", ct)
		}
		err = fmt.Errorf("paniced")
	}()

	return &commit{
		email: *ct.Commit.Author.Email,
		name:  *ct.Commit.Author.Name,
	}, err
}

func readCommitFetch(repoName string, commitHash string) (*commit, error) {
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

	ct, err := object.GetCommit(storage, plumbing.NewHash(commitHash))
	if err != nil {
		return nil, fmt.Errorf("get commit object: %w", err)
	}
	return &commit{
		email: ct.Author.Email,
		name:  ct.Author.Name,
	}, nil
}

func readCommit(ctx context.Context, githubClient *github.Client, repoName string, commitHash string) (*commit, error) {
	commit, err := readCommitAPI(ctx, githubClient, repoName, commitHash)
	if err == nil {
		return commit, nil
	}
	flog.Info("github API errored with %v, falling back to fetch", err)

	return readCommitFetch(repoName, commitHash)
}
