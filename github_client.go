package main

import (
	"context"

	dbg "github.com/gmlewis/go-httpdebug/httpdebug"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

func initGhClient(token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	tp := dbg.New(dbg.WithTransport(tc.Transport))
	return github.NewClient(tp.Client()), nil
}
