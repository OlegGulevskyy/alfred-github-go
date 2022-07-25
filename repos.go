package main

import (
	"context"
	"log"
	"sync"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v45/github"
)

func getOptions(page int) *github.RepositoryListOptions {
	return &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: fetchResultsPerPage, Page: page},
	}
}

func getAllRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	repos := []*github.Repository{}

	wg := sync.WaitGroup{}

	opt := getOptions(1)

	result, response, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	repos = append(repos, result...)

	// there is just one page of repositories
	// no need for further fetching
	if response.NextPage == 0 {
		return repos, nil
	}

	// first page is always retrieved above, so we start iteration
	// from 2 - which is the page number after 1 :) (c) Cap.Obvious
	for p := response.NextPage; p <= response.LastPage; p++ {
		wg.Add(1)

		go func(opt *github.RepositoryListOptions, page int) {
			options := getOptions(page)
			reposPerPage, _, err := client.Repositories.List(ctx, "", options)
			if err != nil {
				log.Fatalln(err)
			}
			repos = append(repos, reposPerPage...)
			wg.Done()
		}(opt, p)
	}

	wg.Wait()
	log.Println("Amount of repos fetched ", len(repos))
	return repos, nil
}

func handleRepositories(ctx context.Context, client *github.Client) {
	if doDownload {
		repos, err := getAllRepos(ctx, client)
		if err != nil {
			err := wf.Session.Store(SESSION_ERROR_KEY, []byte(err.Error()))
			if err != nil {
				wf.FatalError(err)
			}
			return
		}
		// reset any session errors
		wf.Session.Store(SESSION_ERROR_KEY, nil)

		if err := wf.Cache.StoreJSON(reposCacheName, repos); err != nil {
			wf.FatalError(err)
		}

		log.Println("Downloaded repos")
		log.Println(len(repos))
	}

	sessionErrors := hasSessionErrors()
	if sessionErrors {
		return
	}

	log.Printf("Search query = %s", query)
	repos := []*github.Repository{}

	if wf.Cache.Exists(reposCacheName) {
		if err := wf.Cache.LoadJSON(reposCacheName, &repos); err != nil {
			wf.FatalError(err)
		}
	}

	if wf.Cache.Expired(reposCacheName, maxCacheAge) {
		runInBackground(feature)

		if len(repos) == 0 {
			wf.NewItem("Downloading repos...").Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}

	for _, repo := range repos {
		var sub string

		if repo.Description != nil {
			sub = *repo.Description
		}

		wf.NewItem(*repo.FullName).
			Subtitle(sub).
			Arg(*repo.HTMLURL).
			UID(*repo.FullName).
			Valid(true)
	}

	if len(repos) != 0 {
		res := wf.Filter(query)
		log.Printf(" %d/%d Results matching query %q", len(res), len(repos), query)
	}
}
