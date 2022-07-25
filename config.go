package main

import (
	"flag"
	"time"

	aw "github.com/deanishe/awgo"
	"go.deanishe.net/fuzzy"
)

// FIXME
// refactor to use a struct if it's better than just global variables

var (
	wf            *aw.Workflow
	searchOptions []fuzzy.Option

	reposCacheName        = "repos.json"
	pullRequestsCacheName = "pull_requests.json"
	maxResults            = 600
	maxCacheAge           = 180 * time.Minute

	query      string
	doDownload bool
	feature    string
	reRunTime  = 0.3
)

func initFlags() {
	flag.BoolVar(&doDownload, "download", false, "Fetch list of repositories from Github")
	flag.StringVar(&feature, "feature", "", "Defines which Github feature will be queried and handled")
}

func initSortOptions() []fuzzy.Option {
	return []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
}
