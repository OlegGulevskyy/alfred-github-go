package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	aw "github.com/deanishe/awgo"
)

func runInBackground(feature string) {
	wf.Rerun(reRunTime)
	if !wf.IsRunning("download") {
		cmd := exec.Command(
			os.Args[0],
			"-download",
			fmt.Sprintf("-feature=%v", feature),
		)

		if err := wf.RunInBackground("download", cmd); err != nil {
			wf.FatalError(err)
		}
	} else {
		log.Println("Download job is already running")
	}
}

func hasSessionErrors() bool {
	if wf.Session.Exists(SESSION_ERROR_KEY) {
		sessionStatus, err := wf.Session.Load(SESSION_ERROR_KEY)
		if err != nil {
			wf.FatalError(err)
		}
		sessionStatusStr := string(sessionStatus)
		log.Println(sessionStatusStr)

		if strings.Contains(sessionStatusStr, "401") {
			wf.NewItem("Bad github token").
				Subtitle("Please make sure you have a valid Github Personal Access Token set in Alfred workflow configuration with correct scopes (at least 'Repos')").
				Icon(aw.IconError)
			wf.SendFeedback()
			return true
		}

		wf.NewItem("Error").Subtitle(sessionStatusStr)
		wf.SendFeedback()
		return true
	}

	return false
}
