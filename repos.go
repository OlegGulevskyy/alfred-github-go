package main

import (
	"context"
	"log"
	"sync"

	"github.com/google/go-github/v45/github"
)

// func main() {

// 		ctx := context.Background()
// 		client := initGhClient()

// 		getAllRepos(ctx, client)
// }

func getAllRepos(ctx context.Context, client *github.Client) []*github.Repository {
	repos := []*github.Repository{}

	wg := sync.WaitGroup{}

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 50, Page: 1},
	}

	result, response, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		log.Fatalln(err)
	}
	repos = append(repos, result...)

	// there is just one page of repositories
	// no need for further fetching
	if response.NextPage == 0 {
		return repos
	}

	// first page is always retrieved above, so we start iteration
	// from 2 - which is the page number after 1 :) (c) Cap.Obvious
	for p := response.NextPage; p < response.LastPage + 1; p++ {
		wg.Add(1)

		go func(opt *github.RepositoryListOptions, page int) {
			opt.ListOptions.Page = page
			reposPerPage, _, err := client.Repositories.List(ctx, "", opt)
			if err != nil {
				log.Fatalln(err)
			}
			// lastPage = res.LastPage
			repos = append(repos, reposPerPage...)
			wg.Done()
		}(opt, p)
	}

	wg.Wait()

	for _, i := range repos {
		log.Println(*i.FullName)
	}
	log.Println("LENGTH ", len(repos))
	return repos
}
