package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rivo/tview"
)

func handlePomodoroState(pom *pomodoro, view *tview.TextView) {
	// this is the only place where the pomodoro should be changed,
	// all external changes should be triggered via channels!
	// TODO: use channel for changing the statusbar.
	// TODO: listen to start/stop events from: main app & http API.
	// TODO: listen for change-current-focus event.
	tick := make(chan time.Time)
	go attachTicker(tick)

	for {
		<-tick
		if (*pom).Waiting {
			continue
		}
		switch (*pom).State {
		case "ready":
			view.SetText(executeShellHook("work_start"))
			(*pom).State = "work"
			(*pom).StartTime = time.Now()
		case "work":
			delta := time.Now().Sub((*pom).StartTime)
			remaining := (*pom).PomDuration - delta
			if remaining <= 0 {
				view.SetText(executeShellHook("work_done"))
				(*pom).State = "work_done"
				(*pom).StopTime = time.Now()
				(*pom).Waiting = true
				(*pom).DurationLeft = breakDuration
			} else {
				(*pom).DurationLeft = remaining
			}
		case "work_done":
			view.SetText(executeShellHook("break_start"))
			(*pom).State = "break"
			(*pom).BreakStartTime = time.Now()
		case "break":
			delta := time.Now().Sub((*pom).BreakStartTime)
			remaining := (*pom).BreakDuration - delta
			if remaining <= 0 {
				view.SetText(executeShellHook("break_done"))
				(*pom).State = "break_done"
				(*pom).BreakStopTime = time.Now()
				(*pom).DurationLeft = pomodoroDuration
				(*pom).Waiting = true
			} else {
				(*pom).DurationLeft = remaining
			}
		case "break_done":
			(*pom).State = "ready"
			(*pom).PomDuration = pomodoroDuration
			(*pom).DurationLeft = pomodoroDuration
			(*pom).BreakDuration = breakDuration
		}
	}
}

func executeShellHook(script string) string {
	home, _ := os.UserHomeDir()
	hookpath := filepath.Join(home, configfolder, hookfolder, script)
	_, err := exec.Command(hookpath).Output()
	if err != nil {
		return fmt.Sprintf("hook error: [%s]", err)
	}
	return ""
}
