package main

import (
	"context"
	"log"
	"sync"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v45/github"
)

type Repository struct {
	TimesVisited int

	// data will be taken from API
	Id int64
	HtmlUrl string
	FullName string
	Description string
}

type Repositories []Repository

func (r Repository) New(repo *github.Repository) Repository {
	rep := Repository{HtmlUrl: *repo.HTMLURL, FullName: *repo.FullName, Id: *repo.ID}
	if repo.Description != nil {
		rep.Description = *repo.Description
	} else {
		rep.Description = ""
	}

	return rep
}

func (r Repository) MarkVisited() {
 // TODO maybe
}

func (r *Repositories) LoadCacheData() {
	// data := []*Repository{}
	if wf.Cache.Exists(reposCacheName) {
		if err := wf.Cache.LoadJSON(reposCacheName, r); err != nil {
			wf.FatalError(err)
		}
	}
}

func (Repositories) ApiOptions(page int) *github.RepositoryListOptions {
	return &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: fetchResultsPerPage, Page: page},
	}
}

func (r Repositories) ApiData(ctx context.Context, client *github.Client) ([]Repository, error) {
	repos := []*github.Repository{}

	wg := sync.WaitGroup{}

	opt := r.ApiOptions(1)

	result, response, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	repos = append(repos, result...)

	// there is just one page of repositories
	// no need for further fetching
	if response.NextPage != 0 {
		for p := response.NextPage; p <= response.LastPage; p++ {
			wg.Add(1)
	
			go func(opt *github.RepositoryListOptions, page int) {
				options := r.ApiOptions(page)
				reposPerPage, _, err := client.Repositories.List(ctx, "", options)
				if err != nil {
					log.Fatalln(err)
				}
				repos = append(repos, reposPerPage...)
				wg.Done()
			}(opt, p)
		}
		wg.Wait()
	}

	normalizedRepositories := []Repository{}

	for _, i := range repos {
		normalizedRepositories = append(normalizedRepositories, Repository{}.New(i))
	}

	log.Println("Amount of repos fetched ", len(repos))
	return normalizedRepositories, nil
}

func handleRepositories(ctx context.Context, client *github.Client) {
	repositories := Repositories{}

	if doDownload {
		reposData, err := repositories.ApiData(ctx, client)
		if err != nil {
			err := wf.Session.Store(SESSION_ERROR_KEY, []byte(err.Error()))
			if err != nil {
				wf.FatalError(err)
			}
			return
		}
		// reset any session errors
		wf.Session.Store(SESSION_ERROR_KEY, nil)

		if err := wf.Cache.StoreJSON(reposCacheName, reposData); err != nil {
			wf.FatalError(err)
		}

		log.Println("Downloaded repos")
		log.Println(len(reposData))
	}

	sessionErrors := hasSessionErrors()
	if sessionErrors {
		return
	}

	log.Printf("Search query = %s", query)

	repositories.LoadCacheData()

	if wf.Cache.Expired(reposCacheName, maxCacheAge) {
		runInBackground(feature)

		if len(repositories) == 0 {
			wf.NewItem("Downloading repos...").Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}

	for _, repo := range repositories {
		wf.NewItem(repo.FullName).
			Subtitle(repo.Description).
			Arg(repo.HtmlUrl).
			UID(repo.FullName).
			Valid(true)
	}

	if len(repositories) != 0 {
		res := wf.Filter(query)
		log.Printf(" %d/%d Results matching query %q", len(res), len(repositories), query)
	}
}
