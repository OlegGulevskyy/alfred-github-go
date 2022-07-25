package main

import (
	"context"
	"flag"

	aw "github.com/deanishe/awgo"
)

func init() {
	initFlags()
	searchOptions = initSortOptions()

	wf = aw.New(
		aw.HelpURL(HELP_URL),
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

	wf.Configure(aw.TextErrors(true))
	ctx := context.Background()
	client, err := initGhClient()
	if err != nil {
		if err.Error() == "token is not set as environment variable" {
			wf.NewItem("Github Personal access token not found").
				Subtitle("Please make sure you added GITHUB_PAT environment variable in Alfred workflow configuration")
		}
	}

	if feature == "repositories" {
		// repositoryFeature := Repository{}
		handleRepositories(ctx, client)
	} else if feature == "pull_requests" {
		handlePullRequests(ctx, client)
	}

	wf.WarnEmpty("No repos found", "Try using different repo name")
	wf.SendFeedback()
}

func main() {
	// Wrap your entry point with Run() to catch and log panics and
	// show an error in Alfred instead of silently dying
	wf.Run(run)
}
