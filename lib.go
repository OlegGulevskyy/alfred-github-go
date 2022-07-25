package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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