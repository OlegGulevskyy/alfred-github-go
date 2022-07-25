package main

import (
	"time"

	aw "github.com/deanishe/awgo"
	"go.deanishe.net/fuzzy"
)

// FIXME
// refactor to use a struct if it's better than just global variables

var (
	wf            *aw.Workflow
	searchOptions []fuzzy.Option

	reposCacheName = "repos.json"
	maxResults     = 600
	maxCacheAge    = 180 * time.Minute

	query      string
	doDownload bool
	feature    string
	reRunTime  = 0.3
)
