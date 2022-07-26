package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v45/github"
)

type PullRequest struct {
	TimesVisited int
	Data *github.Issue
}

func (p PullRequest) New(issue *github.Issue) PullRequest {
	return PullRequest{Data: issue}
}

func getPullRequestSearchOptions(page int) github.SearchOptions {
	return github.SearchOptions{
		Sort:        "updated",
		ListOptions: github.ListOptions{PerPage: fetchResultsPerPage, Page: page},
	}
}

func getQuery() *string {
	query := ""
	author := *env.GITHUB_HANDLER
	query = fmt.Sprintf("is:open is:pr author:%v archived:false", author)
	return &query
}

func getAllPullRequests(ctx context.Context, client *github.Client) ([]PullRequest, error) {
	prs := []*github.Issue{}
	wg := sync.WaitGroup{}
	opts := getPullRequestSearchOptions(1)
	searchQuery := getQuery()

	result, response, err := client.Search.Issues(ctx, *searchQuery, &opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	prs = append(prs, result.Issues...)

	if response.NextPage != 0 {
		for p := response.NextPage; p <= response.LastPage; p++ {
	
			wg.Add(1)
	
			go func(page int) {
				options := getPullRequestSearchOptions(page)
				prsPerPage, _, err := client.Search.Issues(ctx, *searchQuery, &options)
				if err != nil {
					log.Fatalln(err)
				}
				prs = append(prs, prsPerPage.Issues...)
				wg.Done()
			}(p)
		}

		wg.Wait()
	}

	log.Println("Amount of pull requests fetched ", len(prs))

	normalizedPullRequests := []PullRequest{}

	for _, i := range prs {
		pr := PullRequest{}.New(i)
		normalizedPullRequests = append(normalizedPullRequests, pr)
	}

	return normalizedPullRequests, nil
}

func handlePullRequests(ctx context.Context, client *github.Client) {
	if doDownload {
		prs, err := getAllPullRequests(ctx, client)
		if err != nil {
			err := wf.Session.Store(SESSION_ERROR_KEY, []byte(err.Error()))
			if err != nil {
				wf.FatalError(err)
			}
			return
		}
		// reset any session errors
		wf.Session.Store(SESSION_ERROR_KEY, nil)

		if err := wf.Cache.StoreJSON(pullRequestsCacheName, prs); err != nil {
			wf.FatalError(err)
		}

		log.Println("Downloaded pull requests")
		log.Println(len(prs))
	}

	sessionErrors := hasSessionErrors()
	if sessionErrors {
		return
	}

	if wf.Session.Exists(SESSION_ERROR_KEY) {
		sessionStatus, err := wf.Session.Load(SESSION_ERROR_KEY)
		if err != nil {
			wf.FatalError(err)
		}
		sessionStatusStr := string(sessionStatus)
		if strings.Contains(sessionStatusStr, "401") {
			wf.NewItem("Bad github token").
				Subtitle("Please make sure you have a valid Github Personal Access Token set in Alfred workflow configuration with correct scopes (at least 'Repos')").
				Icon(aw.IconError)
			wf.SendFeedback()
			return
		}

		wf.NewItem("Error").Subtitle(sessionStatusStr)
		wf.SendFeedback()
		return
	}

	log.Printf("Search query = %s", query)
	prs := []PullRequest{}

	if wf.Cache.Exists(pullRequestsCacheName) {
		if err := wf.Cache.LoadJSON(pullRequestsCacheName, &prs); err != nil {
			wf.FatalError(err)
		}
	}

	if wf.Cache.Expired(pullRequestsCacheName, maxCacheAge) {
		runInBackground(feature)

		if len(prs) == 0 {
			wf.NewItem("Downloading pull requests...").Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}

	for _, pr := range prs {
		date := pr.Data.CreatedAt.Format("2006-1-2 15:4:5")
		subtitle := fmt.Sprintf("On: %v | Status: %v", date, *pr.Data.State)

		wf.NewItem(*pr.Data.Title).
			Subtitle(subtitle).
			Arg(*pr.Data.HTMLURL).
			UID(fmt.Sprintf("%v", *pr.Data.ID)).
			Valid(true)
	}

	if len(prs) != 0 {
		res := wf.Filter(query)
		log.Printf(" %d/%d Results matching query %q", len(res), len(prs), query)
	}
}
