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

	"github.com/bmedicke/bhdr/util"
	"github.com/rivo/tview"
)

type pomodoro struct {
	Project                     string
	Task                        string
	Note                        string
	Duration                    time.Duration
	Remaining                   time.Duration
	StartTime                   time.Time
	State                       string
	StopTime                    time.Time
	breakDuration               time.Duration
	breakStartTime              time.Time
	breakStopTime               time.Time
	longBreakDuration           time.Duration
	pomodorosUntilLongBreak     int
	pomodorosUntilLongBreakLeft int
	durationLeft                time.Duration
	waiting                     bool
}

type pomodoroCommand struct {
	commandtype string
	payload     string
}

func handlePomodoroCommand(
	command chan pomodoroCommand,
	pom *pomodoro,
	config Config,
	app *tview.Application,
) {
	select {
	case cmd := <-command:
		switch cmd.commandtype {
		case "continue":
			(*pom).waiting = false
		case "quit_app_save", "quit_app_nosave":
			if (*pom).State == "work" {
				(*pom).State = "incomplete"
				(*pom).StopTime = time.Now()
				if config.LogJSON && cmd.commandtype == "quit_app_save" {
					logPomodoro(*pom)
				}
				executeShellHook("pomodoro_cancelled")
			}
			app.Stop()
		case "update_project":
			(*pom).Project = cmd.payload
		case "update_task":
			(*pom).Task = cmd.payload
		case "update_note":
			(*pom).Note = cmd.payload
		}
	default:
	}
}

func handlePomodoroState(
	pom *pomodoro,
	statusbar *tview.TextView,
	app *tview.Application,
	command chan pomodoroCommand,
	config Config,
) {
	// this is the only place where the pomodoro should be changed,
	// all external changes should be triggered via channels!
	tick := make(chan time.Time)
	go util.AttachTicker(tick, time.Millisecond*200)
	if config.WriteTmuxFile {
		go writeTmuxFile(pom)
	}

	for {
		<-tick

		// non-blocking command handling:
		handlePomodoroCommand(command, pom, config, app)

		// state handling:
		if (*pom).waiting {
			continue
		}
		switch (*pom).State {
		case "ready":
			statusbar.SetText(executeShellHook("work_start"))
			(*pom).State = "work"
			(*pom).StartTime = time.Now()
		case "work":
			delta := time.Now().Sub((*pom).StartTime)
			(*pom).Remaining = (*pom).Duration - delta
			if (*pom).Remaining <= 0 {
				statusbar.SetText(executeShellHook("work_done"))
				(*pom).State = "work_done"
				(*pom).StopTime = time.Now()
				(*pom).waiting = true
				(*pom).Remaining = 0

				(*pom).pomodorosUntilLongBreakLeft--
				if (*pom).pomodorosUntilLongBreakLeft == 0 {
					(*pom).durationLeft = (*pom).longBreakDuration
				} else {
					(*pom).durationLeft = (*pom).breakDuration
				}

				if config.LogJSON {
					go logPomodoro(*pom)
				}
			} else {
				(*pom).durationLeft = (*pom).Remaining
			}
		case "work_done":
			if (*pom).pomodorosUntilLongBreakLeft == 0 {
				statusbar.SetText(executeShellHook("longbreak_start"))
				(*pom).State = "longbreak"
			} else {
				statusbar.SetText(executeShellHook("break_start"))
				(*pom).State = "break"
			}

			(*pom).breakStartTime = time.Now()
		case "break", "longbreak":
			delta := time.Now().Sub((*pom).breakStartTime)
			var remaining time.Duration

			if (*pom).State == "longbreak" {
				remaining = (*pom).longBreakDuration - delta
			} else {
				remaining = (*pom).breakDuration - delta
			}

			if remaining <= 0 {
				if (*pom).State == "longbreak" {
					statusbar.SetText(executeShellHook("longbreak_done"))
					(*pom).State = "longbreak_done"
				} else {
					statusbar.SetText(executeShellHook("break_done"))
					(*pom).State = "break_done"
				}

				(*pom).breakStopTime = time.Now()

				(*pom).durationLeft = (*pom).Duration
				if (*pom).pomodorosUntilLongBreakLeft == 0 {
					(*pom).pomodorosUntilLongBreakLeft = (*pom).pomodorosUntilLongBreak
				}

				(*pom).waiting = true
			} else {
				(*pom).durationLeft = remaining
			}
		case "break_done", "longbreak_done":
			(*pom).State = "ready"
			(*pom).durationLeft = (*pom).Duration
		}
	}
}

func writeTmuxFile(pom *pomodoro) {
	tick := make(chan time.Time)
	go util.AttachTicker(tick, time.Second*1)

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

func createPomodoro(config Config, longBreakIn int) pomodoro {
	pomodoroDuration := time.Duration(
		config.PomodoroDurationMinutes,
	) * time.Minute
	breakDuration := time.Duration(
		config.BreakDurationMinutes,
	) * time.Minute
	longBreakDuration := time.Duration(
		config.LongBreakDurationMinutes,
	) * time.Minute
	longBreakAfterPomodoros := config.LongBreakAfterPomodoros

	// set sensible defaults (in case of missing config file):
	if pomodoroDuration == 0 {
		pomodoroDuration = 25 * time.Minute
	}
	if breakDuration == 0 {
		breakDuration = 5 * time.Minute
	}
	if longBreakDuration == 0 {
		longBreakDuration = 30 * time.Minute
	}
	if longBreakAfterPomodoros == 0 {
		longBreakAfterPomodoros = 3
	}
	if longBreakIn < 1 {
		longBreakIn = longBreakAfterPomodoros
	}

	pom := pomodoro{
		State:                       "ready",
		Project:                     config.DefaultProject,
		Task:                        config.DefaultTask,
		Note:                        config.DefaultNote,
		Duration:                    pomodoroDuration,
		breakDuration:               breakDuration,
		longBreakDuration:           longBreakDuration,
		durationLeft:                pomodoroDuration,
		pomodorosUntilLongBreak:     longBreakAfterPomodoros,
		pomodorosUntilLongBreakLeft: longBreakIn,
		waiting:                     true,
	}
	return pom
}
