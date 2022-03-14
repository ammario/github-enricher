package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/hstove/gender/classifier"
	"github.com/samber/lo"
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

func setupEnrichers() ([]enricher, error) {
	githubClient, err := newGitHubClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create github client: %w", err)
	}

	var enrichers []enricher
	enrichers = append(enrichers, []enricher{
		{
			FieldName: "email",
			Deps:      []string{"repo_name", "commit"},
			CacheDeps: []string{"commit"},
			Run: func(ctx context.Context, row map[string]string) (string, error) {
				commit, err := readCommit(ctx, githubClient, row["repo_name"], row["commit"])
				if err != nil {
					return "", err
				}
				return commit.email, nil
			},
		},
		{
			FieldName: "name",
			Deps:      []string{"repo_name", "commit"},
			CacheDeps: []string{"commit"},
			Run: func(ctx context.Context, row map[string]string) (string, error) {
				commit, err := readCommit(ctx, githubClient, row["repo_name"], row["commit"])
				if err != nil {
					return "", err
				}
				return commit.name, nil
			},
		},
		{
			FieldName: "lastname",
			Deps:      []string{"name"},
			Run: func(ctx context.Context, row map[string]string) (string, error) {
				name, _ := lo.Last(strings.Split(row["name"], " "))
				return cleanName(name), nil
			},
		},
		{
			FieldName: "firstname",
			Deps:      []string{"name"},
			Run: func(ctx context.Context, row map[string]string) (string, error) {
				name := strings.Split(row["name"], " ")[0]
				return cleanName(name), nil
			},
		},
	}...)

	genderClassifier := classifier.Classifier()
	enrichers = append(enrichers, enricher{
		FieldName: "gender",
		Deps:      []string{"name"},
		CacheDeps: []string{"name"},
		Run: func(ctx context.Context, row map[string]string) (string, error) {
			firstName := strings.Split(row["name"], " ")[0]
			gender, _ := classifier.Classify(genderClassifier, firstName)
			return gender, nil
		},
	})
	return enrichers, nil
}
