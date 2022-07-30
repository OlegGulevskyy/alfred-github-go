package main

import (
	"context"
	"log"

	flag "github.com/spf13/pflag"

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

	env = Env{}.New()
}

func run() {
	wf.Args() // handle any magic arguments
	flag.Parse()

	wf.Configure(aw.TextErrors(true))
	ctx := context.Background()
	client, err := initGhClient(*env.GITHUB_PAT)
	if err != nil {
		if err.Error() == "token is not set as environment variable" {
			wf.NewItem("Github Personal access token not found").
				Subtitle("Please make sure you added GITHUB_PAT environment variable in Alfred workflow configuration")
		}
	}

	if feature == REPOSITORIES {
		handleRepositories(ctx, client)
	} else if feature == PULL_REQUESTS {
		handlePullRequests(ctx, client)
	}

	if query == "" && markVisited != "" {
		// TODO
		log.Println("Marking Item as visited", markVisited)
	}

	wf.WarnEmpty("No repos found", "Try using different repo name")
	wf.SendFeedback()
}

func main() {
	// Wrap your entry point with Run() to catch and log panics and
	// show an error in Alfred instead of silently dying
	wf.Run(run)
}
