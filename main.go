package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v45/github"
	"go.deanishe.net/fuzzy"
)

var (
	wf            *aw.Workflow
	searchOptions []fuzzy.Option

	reposCacheName = "repos.json"
	maxResults     = 600
	maxCacheAge    = 180 * time.Minute

	query      string
	doDownload bool
	reRunTime  = 0.3
)

func init() {
	flag.BoolVar(&doDownload, "download", false, "Fetch list of repositories from Github")

	searchOptions = []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(
		aw.HelpURL("https://github.com/OlegGulevskyy"),
		aw.MaxResults(maxResults),
		aw.SortOptions(searchOptions...),
	)
}

func run() {
	wf.Args() // handle any magic arguments
	flag.Parse()

	if args := wf.Args(); len(args) > 0 {
		query = args[0]
	}

	if doDownload {
		wf.Configure(aw.TextErrors(true))
		log.Printf("Starting repositories fetch...")

		ctx := context.Background()
		client := initGhClient()

		repos := getAllRepos(ctx, client)

		if err := wf.Cache.StoreJSON(reposCacheName, repos); err != nil {
			wf.FatalError(err)
		}

		log.Println("Downloaded repos")
		log.Println(len(repos))
	}

	log.Printf("Search query = %s", query)
	repos := []*github.Repository{}

	if wf.Cache.Exists(reposCacheName) {
		if err := wf.Cache.LoadJSON(reposCacheName, &repos); err != nil {
			wf.FatalError(err)
		}
	}

	if wf.Cache.Expired(reposCacheName, maxCacheAge) {
		wf.Rerun(reRunTime)

		if !wf.IsRunning("download") {
			cmd := exec.Command(os.Args[0], "-download")

			if err := wf.RunInBackground("download", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Println("Download job is already running")
		}

		if len(repos) == 0 {
			wf.NewItem("Downloading repos...").Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}

	log.Println("REPOS", len(repos))

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

	if query != "" {
		res := wf.Filter(query)
		log.Printf(" %d/%d Results matching query %q", len(res), len(repos), query)
	}

	wf.WarnEmpty("No repos found", "Try using different repo name")

	wf.SendFeedback()

}

func main() {
	// Wrap your entry point with Run() to catch and log panics and
	// show an error in Alfred instead of silently dying
	// run()
	wf.Run(run)
}
