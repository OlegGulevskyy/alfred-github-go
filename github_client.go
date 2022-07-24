package main

import (
	"context"

	dbg "github.com/gmlewis/go-httpdebug/httpdebug"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

const TOKEN_ENV_KEY = "GITHUB_PAT"

func getToken() string {
	token, _ := wf.Config.Env.Lookup(TOKEN_ENV_KEY)
	return token
}

func initGhClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getToken()},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	tp := dbg.New(dbg.WithTransport(tc.Transport))
	return github.NewClient(tp.Client())
}
