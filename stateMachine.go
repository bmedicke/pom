package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rivo/tview"
)

type pomodoro struct {
	Project        string
	Task           string
	Note           string
	Duration       time.Duration
	StartTime      time.Time
	State          string
	StopTime       time.Time
	breakDuration  time.Duration
	breakStartTime time.Time
	breakStopTime  time.Time
	durationLeft   time.Duration
	waiting        bool
}

type pomodoroCommand struct {
	commandtype string
	payload     string
}

func handlePomodoroState(
	pom *pomodoro,
	view *tview.TextView,
	command chan pomodoroCommand,
	config Config,
) {
	// this is the only place where the pomodoro should be changed,
	// all external changes should be triggered via channels!
	// TODO: use channel for changing the statusbar.
	// TODO: listen to start/stop events from: main app & http API.
	tick := make(chan time.Time)
	go attachTicker(tick, time.Millisecond*200)
	go writeTmuxFile(pom)

	for {
		<-tick

		// non-blocking command handling:
		select {
		case cmd := <-command:
			switch cmd.commandtype {
			case "continue":
				(*pom).waiting = false
			case "update_project":
				(*pom).Project = cmd.payload
			case "update_task":
				(*pom).Task = cmd.payload
			case "update_note":
				(*pom).Note = cmd.payload
			}
		default:
		}

		if (*pom).waiting {
			continue
		}
		switch (*pom).State {
		case "ready":
			view.SetText(executeShellHook("work_start"))
			(*pom).State = "work"
			(*pom).StartTime = time.Now()
		case "work":
			delta := time.Now().Sub((*pom).StartTime)
			remaining := (*pom).Duration - delta
			if remaining <= 0 {
				view.SetText(executeShellHook("work_done"))
				(*pom).State = "work_done"
				(*pom).StopTime = time.Now()
				(*pom).waiting = true
				(*pom).durationLeft = (*pom).breakDuration
				go logPomodoro(*pom)
			} else {
				(*pom).durationLeft = remaining
			}
		case "work_done":
			view.SetText(executeShellHook("break_start"))
			(*pom).State = "break"
			(*pom).breakStartTime = time.Now()
		case "break":
			delta := time.Now().Sub((*pom).breakStartTime)
			remaining := (*pom).breakDuration - delta
			if remaining <= 0 {
				view.SetText(executeShellHook("break_done"))
				(*pom).State = "break_done"
				(*pom).breakStopTime = time.Now()
				(*pom).durationLeft = (*pom).Duration
				(*pom).waiting = true
			} else {
				(*pom).durationLeft = remaining
			}
		case "break_done":
			(*pom).State = "ready"
			(*pom).durationLeft = (*pom).Duration
		}
	}
}

func writeTmuxFile(pom *pomodoro) {
	tick := make(chan time.Time)
	go attachTicker(tick, time.Second*1)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	file := filepath.Join(home, configfolder, "tmux")

	for {
		<-tick
		f, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		f.WriteString(
			fmt.Sprintf(
				"%s %s",
				(*pom).State,
				(*pom).durationLeft.Round(time.Second),
			),
		)
	}
}

func clearTmuxFile() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	file := filepath.Join(home, configfolder, "tmux")
	f, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	f.WriteString("")
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

func logPomodoro(newPomodoro pomodoro) {
	home, _ := os.UserHomeDir()
	log := filepath.Join(home, configfolder, "log.json")
	var pomodoros []pomodoro

	file, _ := os.Open(log)
	content, _ := ioutil.ReadAll(file)
	json.Unmarshal(content, &pomodoros)

	pomodoros = append(pomodoros, newPomodoro)
	newJSON, _ := json.MarshalIndent(pomodoros, "", "  ")
	file, _ = os.OpenFile(log, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	file.WriteString(string(newJSON))
}

func createPomodoro(config Config) pomodoro {
	pomodoroDuration := time.Duration(
		config.PomodoroDurationMinutes,
	) * time.Minute
	breakDuration := time.Duration(
		config.BreakDurationMinutes,
	) * time.Minute

	// set sensible default durations
	// (in case of missing config file):
	if pomodoroDuration == 0 {
		pomodoroDuration = 25 * time.Minute
	}
	if breakDuration == 0 {
		breakDuration = 5 * time.Minute
	}

	pom := pomodoro{
		State:         "ready",
		Project:       config.DefaultProject,
		Task:          config.DefaultNote,
		Note:          config.DefaultTask,
		Duration:      pomodoroDuration,
		breakDuration: breakDuration,
		durationLeft:  pomodoroDuration,
		waiting:       true,
	}
	return pom
}
