package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bmedicke/bhdr/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	pomodoroDuration time.Duration = time.Minute * 25
	breakDuration    time.Duration = time.Minute * 5
	callbackfolder   string        = ".pom/callbacks"
)

func spawnTUI() {
	app := tview.NewApplication()
	pom := createPomodoro(pomodoroDuration, breakDuration)
	chord := util.KeyChord{Active: false, Buffer: "", Action: ""}

	// TODO: read chordmap from json file (compile it into binary).
	chordmap := map[string]interface{}{}
	chordmap["c"] = map[string]interface{}{"c": "continue"}

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	frame := tview.NewFrame(layout)
	header := tview.NewFlex()
	headerleft := tview.NewTextView()
	headercenter := tview.NewTextView()
	headerright := tview.NewTextView()
	body := tview.NewTable()
	statusbar := tview.NewTextView()

	frame.AddText(" P 🐕 M ", true, tview.AlignCenter, tcell.ColorLime)
	frame.SetBorders(0, 0, 0, 0, 0, 0)
	frame.SetBackgroundColor(tcell.Color236)

	layout.AddItem(header, 1, 0, false)
	layout.AddItem(body, 0, 1, true)
	layout.AddItem(statusbar, 1, 0, false)

	header.SetBorderPadding(0, 0, 0, 0)
	header.AddItem(headerleft, 0, 2, false)
	header.AddItem(headercenter, 0, 1, false)
	header.AddItem(headerright, 17, 0, false)

	headerleft.SetChangedFunc(func() { app.Draw() })
	headerright.SetChangedFunc(func() { app.Draw() })

	statusbar.SetBackgroundColor(tcell.ColorDarkOliveGreen)
	statusbar.SetBorderPadding(0, 0, 0, 0)
	statusbar.SetChangedFunc(func() { app.Draw() })

	body.SetSelectable(true, true)

	app.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				util.ResetChord(&chord)
			}

			if chord.Active {
				util.HandleChords(event.Rune(), &chord, chordmap)
				handleAction(chord.Action, &pom)
			} else {
				switch event.Rune() {
				case 'q':
					app.Stop()
				case 'c':
					if err := util.HandleChords(event.Rune(), &chord, chordmap); err != nil {
						statusbar.SetText(fmt.Sprint(err))
					}
				}
			}
			statusbar.SetText(chord.Buffer)
			return event
		},
	)

	updateBody(body)
	go handlePomodoroState(&pom)
	go updateHeader(headerleft, headercenter, headerright, &pom)

	app.SetRoot(frame, true)
	app.SetFocus(body).Run()
}

// Pomodoro TODO
type Pomodoro struct {
	State          string
	StartTime      time.Time
	StopTime       time.Time
	PomDuration    time.Duration
	DurationLeft   time.Duration
	BreakStartTime time.Time
	BreakStopTime  time.Time
	BreakDuration  time.Duration
	Waiting        bool
	CurrentTask    string
}

func createPomodoro(
	duration time.Duration,
	breakDuration time.Duration,
) Pomodoro {
	pom := Pomodoro{
		State:         "ready",
		PomDuration:   duration,
		DurationLeft:  duration,
		BreakDuration: breakDuration,
		Waiting:       true,
	}
	return pom
}

func handlePomodoroState(pom *Pomodoro) {
	// this is the only place where the pomodoro should be changed!
	// create commandChannel:
	// listen to start/stop events from:
	// main app & http API
	for {
		if (*pom).Waiting {
			continue
		}
		switch (*pom).State {
		case "ready":
			executeShellCallback("work_start.sh")
			(*pom).State = "work"
			(*pom).StartTime = time.Now()
		case "work":
			delta := time.Now().Sub((*pom).StartTime)
			remaining := (*pom).PomDuration - delta
			if remaining <= 0 {
				executeShellCallback("work_done.sh")
				(*pom).State = "work_done"
				(*pom).StopTime = time.Now()
				(*pom).Waiting = true
				(*pom).DurationLeft = breakDuration
			} else {
				(*pom).DurationLeft = remaining
			}
		case "work_done":
			executeShellCallback("break_start.sh")
			(*pom).State = "break"
			(*pom).BreakStartTime = time.Now()
		case "break":
			delta := time.Now().Sub((*pom).BreakStartTime)
			remaining := (*pom).BreakDuration - delta
			if remaining <= 0 {
				executeShellCallback("break_done.sh")
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

func updateBody(view *tview.Table) {
	// read (only!) from struct?
	b := []map[string]string{
		{"id": "current task", "value": "research"},
		{"id": "server", "value": "0.0.0.0:8421/api"},
	}
	cols, rows := 3, len(b)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			var s string
			switch col {
			case 0:
				s = b[row]["id"]
			case 1:
				s = "  ==  "
			case 2:
				s = b[row]["value"]
			}
			cell := tview.NewTableCell(s)
			if col < 2 {
				cell.SetSelectable(false)
			}
			view.SetCell(row, col, cell)
		}
	}
}

func attachTicker(timer chan time.Time) {
	timer <- time.Now() // send one tick immediately.
	t := time.NewTicker(500 * time.Millisecond)
	for c := range t.C {
		timer <- c
	}
}

func handleAction(action string, pom *Pomodoro) {
	switch action {
	// only
	case "continue":
		// TODO send signal instead of mutating state directly!
		(*pom).Waiting = false
	case "cancel":
		// TODO send cancel (break or pomodoro) message
	}
}

func updateHeader(
	left *tview.TextView,
	center *tview.TextView,
	right *tview.TextView,
	pom *Pomodoro,
) {
	ticker := make(chan time.Time)
	go attachTicker(ticker)

	for {
		<-ticker
		timeleft := (*pom).DurationLeft.Round(time.Second)
		// TODO left pad text:
		right.SetText(fmt.Sprintf("%v [%v]", (*pom).State, timeleft))

		var color tcell.Color
		switch (*pom).State {
		case "work":
			color = tcell.ColorDarkOliveGreen
		case "break":
			color = tcell.ColorBlue
		case "break_done", "work_done":
			color = tcell.ColorDarkRed
		}

		left.SetBackgroundColor(color)
		center.SetBackgroundColor(color)
		right.SetBackgroundColor(color)
	}
}

func executeShellCallback(script string) {
	home, _ := os.UserHomeDir()
	callbackpath := filepath.Join(home, callbackfolder, script)
	exec.Command(callbackpath).Output()
}
