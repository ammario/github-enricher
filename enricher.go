package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/coder/flog"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/samber/lo"
)

type engine struct {
	Log *flog.Logger
}

// An enricher modifies the row before output.
type enricher struct {
	FieldName    string
	Dependencies []string
	Run          func(ctx context.Context, row map[string]string) (string, error)
}

func readCommit(repoName string, commitHash string) (*object.Commit, error) {

	var (
		gitURI  = fmt.Sprintf("https://github.com/%s.git")
		storage = memory.NewStorage()
		refSpec = config.RefSpec(fmt.Sprintf("%s:%s", commitHash, commitHash))
	)

	if commitHash == "" {
		return nil, fmt.Errorf("ref is empty")
	}
	c := &config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{gitURI},
		Fetch: []config.RefSpec{refSpec},
	}
	r := git.NewRemote(storage, c)
	err := r.Fetch(&git.FetchOptions{
		Depth:    1,
		RefSpecs: []config.RefSpec{refSpec},
		Progress: os.Stderr,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	commit, err := object.GetCommit(storage, plumbing.NewHash(commitHash))
	if err != nil {
		return nil, fmt.Errorf("get commit object: %w", err)
	}
	return commit, nil
}

var enrichers = []enricher{
	{
		FieldName:    "email",
		Dependencies: []string{"repo_name", "commit"},
		Run: func(ctx context.Context, row map[string]string) (string, error) {
			commit, err := readCommit(row["repo_name"], row["commit"])
			if err != nil {
				return "", err
			}
			return commit.Author.Email, nil
		},
	},
}

// Run is the main enrichment loop
func (eng engine) Run(w io.Writer, r io.Reader) error {
	var (
		csvReader = csv.NewReader(r)
		csvWriter = csv.NewWriter(w)
	)

	inputHeader, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	// Form output header from input header and additional possible enrichers
	outputHeader := append([]string(nil), inputHeader...)
	var usedEnrichers []enricher
findEnrichers:
	for _, enricher := range enrichers {
		if lo.Contains(inputHeader, enricher.FieldName) {
			continue
		}
		// Break if dependencies don't exist in input
		for _, dep := range enricher.Dependencies {
			if !lo.Contains(inputHeader, dep) {
				continue findEnrichers
			}
		}
		outputHeader = append(outputHeader, enricher.FieldName)
		usedEnrichers = append(usedEnrichers, enricher)
	}

	err = csvWriter.Write(outputHeader)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	csvWriter.Flush()

	for i := 0; ; i++ {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read row: %w", err)
		}

		rowMap := make(map[string]string, len(row))
		for i, v := range row {
			rowMap[inputHeader[i]] = v
		}

		for _, e := range usedEnrichers {
			v, err := e.Run(context.Background(), rowMap)
			row = append(row, v)
			if err != nil {
				eng.Log.Error("%q enrich failed: %+v", e.FieldName, err)
			}
		}
		err = csvWriter.Write(row)
		if err != nil {
			return fmt.Errorf("write row: %w", err)
		}
		csvWriter.Flush()
	}
}
