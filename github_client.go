package main

import (
	"context"

	dbg "github.com/gmlewis/go-httpdebug/httpdebug"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

var GITHUB_PAT = "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"

func initGhClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_PAT},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	tp := dbg.New(dbg.WithTransport(tc.Transport))
	return github.NewClient(tp.Client())
}
