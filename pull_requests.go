package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v45/github"
)

// func main() {
// 	ghClient, err := initGhClient(); if err != nil {
// 		log.Fatalln(err)
// 	}

// 	getAllPullRequests(context.Background(), ghClient)
// }

func getPullRequestSearchOptions(page int) github.SearchOptions {
	return github.SearchOptions{
		Sort: "updated",
		ListOptions: github.ListOptions{PerPage: 10, Page: page},
	}
}

func getAllPullRequests(ctx context.Context, client *github.Client) ([]*github.Issue, error) {
	prs := []*github.Issue{}
	wg := sync.WaitGroup{}
	opts := getPullRequestSearchOptions(1)
	searchQuery := "is:open is:pr author:OlegGulevskyy archived:false"

	result, response, err := client.Search.Issues(ctx, searchQuery, &opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	prs = append(prs, result.Issues...)

	if response.NextPage == 0 {
		return prs, nil
	}

	for p := response.NextPage; p <= response.LastPage; p++ {

		wg.Add(1)

		go func(page int) {
			options := getPullRequestSearchOptions(page)
			prsPerPage, _, err := client.Search.Issues(ctx, searchQuery, &options)
			if err != nil {
				log.Fatalln(err)
			}
			prs = append(prs, prsPerPage.Issues...)
			wg.Done()
		}(p)
	}

	wg.Wait()

	for _, i := range prs {
		log.Println(*i.Title)
	}
	log.Println("Amount of pull requests fetched ", len(prs))
	return prs, nil
}

func handlePullRequests() {
	log.Println("Handling Pull requests")

	if doDownload {
		wf.Configure(aw.TextErrors(true))
		log.Printf("Starting repositories fetch...")

		ctx := context.Background()
		client, err := initGhClient()
		if err != nil {
			if err.Error() == "token is not set as environment variable" {
				wf.NewItem("Github Personal access token not found").
					Subtitle("Please make sure you added GITHUB_PAT environment variable in Alfred workflow configuration")
			}
		}

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

	// if any global session errors happened
	// such as Bad github token -> this is the moment to handle them
	// before proceeding further
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
	prs := []*github.Issue{}

	if wf.Cache.Exists(pullRequestsCacheName) {
		if err := wf.Cache.LoadJSON(pullRequestsCacheName, &prs); err != nil {
			wf.FatalError(err)
		}
	}

	if wf.Cache.Expired(pullRequestsCacheName, maxCacheAge) {
		wf.Rerun(reRunTime)

		if !wf.IsRunning("download") {
			cmd := exec.Command(os.Args[0], "-download", "-feature=pull_requests")

			if err := wf.RunInBackground("download", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Println("Download job is already running")
		}

		if len(prs) == 0 {
			wf.NewItem("Downloading pull requests...").Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}

	for _, pr := range prs {
		date := pr.CreatedAt.Format("2006-1-2 15:4:5")
		subtitle := fmt.Sprintf("On: %v | Status: %v", date, *pr.State)

		wf.NewItem(*pr.Title).
			Subtitle(subtitle).
			Arg(*pr.HTMLURL).
			UID(string(*pr.ID)).
			Valid(true)
	}

	if len(prs) != 0 {
		res := wf.Filter(query)
		log.Printf(" %d/%d Results matching query %q", len(res), len(prs), query)
	}

	wf.WarnEmpty("No repos found", "Try using different repo name")

	wf.SendFeedback()
}