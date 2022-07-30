package main

import (
	"time"

	flag "github.com/spf13/pflag"

	aw "github.com/deanishe/awgo"
	"go.deanishe.net/fuzzy"
)

var (
	wf            *aw.Workflow
	searchOptions []fuzzy.Option

	reposCacheName        = "repos.json"
	pullRequestsCacheName = "pull_requests.json"
	maxResults            = 600
	maxCacheAge           = 180 * time.Minute

	fetchResultsPerPage = 30

	query      string
	doDownload bool
	feature    string
	markVisited string
	reRunTime  = 0.3

	env Env
)

func initFlags() {
	flag.StringVar(&query, "query", "", "Main query input")
	flag.BoolVar(&doDownload, "download", false, "Fetch list of repositories from Github")
	flag.StringVar(&feature, "feature", "", "Defines which Github feature will be queried and handled")
	flag.StringVar(&markVisited, "mark_visited", "", "Mark item as visited")
}

func initSortOptions() []fuzzy.Option {
	return []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
}
