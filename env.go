package main

import (
	"errors"
	"log"
)

const (
	TOKEN_ENV_KEY          = "GITHUB_PAT"
	GITHUB_HANDLER_ENV_KEY = "GITHUB_HANDLER"
)

type Env struct {
	GITHUB_HOST    *string
	GITHUB_PAT     *string
	GITHUB_HANDLER *string
}

func (e Env) getToken() (*string, error) {
	token, isSet := wf.Config.Env.Lookup(TOKEN_ENV_KEY)
	if !isSet {
		return nil, errors.New("token is not set as environment variable")
	}
	return &token, nil
}

func (e Env) getHandler() (*string, error) {
	handler, ok := wf.Config.Env.Lookup(GITHUB_HANDLER_ENV_KEY)
	if !ok {
		return nil, errors.New("github handler is not specified")
	}
	return &handler, nil
}

func (e Env) New() Env {
	env := Env{}
	token, err := env.getToken()
	if err != nil {
		log.Fatalln(err)
	}
	env.GITHUB_PAT = token

	handler, err := env.getHandler()
	if err != nil {
		log.Fatalln(err)
	}
	env.GITHUB_HANDLER = handler
	return env
}
