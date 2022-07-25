package main

import (
	"flag"

	aw "github.com/deanishe/awgo"
	"go.deanishe.net/fuzzy"
)

func init() {
	flag.BoolVar(&doDownload, "download", false, "Fetch list of repositories from Github")
	flag.StringVar(&feature, "feature", "", "Defines which Github feature will be queried and handled")

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

	if args := wf.Args(); len(args) > 1 {
		query = args[len(args)-1]
	}

	if feature == "repositories" {
		handleRepositories()
		return
	}

	if feature == "pull_requests" {
		handlePullRequests()
		return
	}
}

func main() {
	// Wrap your entry point with Run() to catch and log panics and
	// show an error in Alfred instead of silently dying
	wf.Run(run)
}
