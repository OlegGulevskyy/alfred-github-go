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

	Id int64
	HtmlUrl string
	FullName string
	Description string
	CreatedAt string
	State string
}

type PullRequests []PullRequest

func (p PullRequest) New(issue *github.Issue) PullRequest {
	pr := PullRequest{
		HtmlUrl: *issue.HTMLURL,
		FullName: *issue.Title,
		Id: *issue.ID,
		CreatedAt: issue.CreatedAt.Format("2006-1-2 15:4:5"),
		State: *issue.State,
	}
	return pr
}

func (p PullRequest) MarkVisited() {
	// TODO
}

func (prs *PullRequests) LoadCacheData() {
	if wf.Cache.Exists(pullRequestsCacheName) {
		if err := wf.Cache.LoadJSON(pullRequestsCacheName, prs); err != nil {
			wf.FatalError(err)
		}
	}
}

func (PullRequests) ApiOptions(page int) github.SearchOptions {
	return github.SearchOptions{
		Sort:        "updated",
		ListOptions: github.ListOptions{PerPage: fetchResultsPerPage, Page: page},
	}
}

func (PullRequests) ApiSearchQuery() *string {
	query := ""
	author := *env.GITHUB_HANDLER
	query = fmt.Sprintf("is:open is:pr author:%v archived:false", author)
	return &query
}

func (prs PullRequests) ApiData(ctx context.Context, client *github.Client) ([]PullRequest, error) {
	data := []*github.Issue{}
	wg := sync.WaitGroup{}
	opts := prs.ApiOptions(1)
	searchQuery := prs.ApiSearchQuery()

	result, response, err := client.Search.Issues(ctx, *searchQuery, &opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	data = append(data, result.Issues...)

	if response.NextPage != 0 {
		for p := response.NextPage; p <= response.LastPage; p++ {
	
			wg.Add(1)
	
			go func(page int) {
				options := prs.ApiOptions(page)
				prsPerPage, _, err := client.Search.Issues(ctx, *searchQuery, &options)
				if err != nil {
					log.Fatalln(err)
				}
				data = append(data, prsPerPage.Issues...)
				wg.Done()
			}(p)
		}

		wg.Wait()
	}

	log.Println("Amount of pull requests fetched ", len(prs))

	normalizedPullRequests := []PullRequest{}

	for _, i := range data {
		pr := PullRequest{}.New(i)
		normalizedPullRequests = append(normalizedPullRequests, pr)
	}

	return normalizedPullRequests, nil
}

func handlePullRequests(ctx context.Context, client *github.Client) {
	prs := PullRequests{}
	if doDownload {
		prs, err := prs.ApiData(ctx, client)
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
	prs.LoadCacheData()

	if wf.Cache.Expired(pullRequestsCacheName, maxCacheAge) {
		runInBackground(feature)

		if len(prs) == 0 {
			wf.NewItem("Downloading pull requests...").Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}

	for _, pr := range prs {
		subtitle := fmt.Sprintf("On: %v | Status: %v", pr.CreatedAt, pr.State)

		wf.NewItem(pr.FullName).
			Subtitle(subtitle).
			Arg(pr.HtmlUrl).
			UID(fmt.Sprintf("%v", pr.Id)).
			Valid(true)
	}

	if len(prs) != 0 {
		res := wf.Filter(query)
		log.Printf(" %d/%d Results matching query %q", len(res), len(prs), query)
	}
}
