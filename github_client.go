package main

import (
	"context"
	"errors"

	dbg "github.com/gmlewis/go-httpdebug/httpdebug"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

const TOKEN_ENV_KEY = "GITHUB_PAT"

func getToken() (*string, error) {
	token, isSet := wf.Config.Env.Lookup(TOKEN_ENV_KEY)
	if !isSet {
		return nil, errors.New("token is not set as environment variable")
	}
	return &token, nil
}

func initGhClient() (*github.Client, error) {
	githubToken, err := getToken(); if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	tp := dbg.New(dbg.WithTransport(tc.Transport))
	return github.NewClient(tp.Client()), nil
}
