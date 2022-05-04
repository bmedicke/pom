package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/bmedicke/bhdr/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

//go:embed chordmap.json
var chordmapJSON string

// Config is used for unmarshalling.
type Config struct {
	PomodoroDurationMinutes int    `json:"pomodoroDurationMinutes"`
	BreakDurationMinutes    int    `json:"breakDurationMinutes"`
	DefaultNote             string `json:"defaultNote"`
	DefaultTask             string `json:"defaultTask"`
}

func spawnTUI() {
	app := tview.NewApplication()

	// used for updating the pomodoro from goroutines:
	command := make(chan pomodoroCommand)

	// vim-style key chords:
	chord := util.KeyChord{Active: false, Buffer: "", Action: ""}
	chordmap := map[string]interface{}{}
	json.Unmarshal([]byte(chordmapJSON), &chordmap)

	config := getConfig()
	pom := createPomodoro(config)

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	frame := tview.NewFrame(layout)
	header := tview.NewFlex()
	headerleft := tview.NewTextView()
	headercenter := tview.NewTextView()
	headerright := tview.NewTextView()
	body := tview.NewPages()
	bodytable := tview.NewTable()
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

	// page used for overlaying widgets (edit box):
	body.AddPage("table", bodytable, true, true)

	bodytable.SetSelectable(true, true)
	bodytable.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				util.ResetChord(&chord)
			case tcell.KeyEnter:
				editTableCell(body, bodytable, command)
			}

			if chord.Active {
				util.HandleChords(event.Rune(), &chord, chordmap)
				handleAction(chord.Action, pom, command)
			} else {
				switch event.Rune() {
				case 'q':
					app.Stop()
				case 'c', 'd': // start chord:
					util.HandleChords(event.Rune(), &chord, chordmap)
				}
			}
			statusbar.SetText(chord.Buffer)
			return event
		},
	)

	createBodytable(bodytable, config)
	go handlePomodoroState(&pom, statusbar, command, config)
	go updateHeader(headerleft, headercenter, headerright, &pom)

	app.SetRoot(frame, true)
	app.SetFocus(bodytable).Run()
}

func createBodytable(view *tview.Table, config Config) {
	b := []map[string]string{
		{
			"id":       "current task",
			"onchange": "update_task",
			"type":     "editable",
			"value":    config.DefaultTask,
		},
		{
			"id":       "note",
			"onchange": "update_note",
			"type":     "editable",
			"value":    config.DefaultNote,
		},
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
			if b[row]["type"] == "editable" && col == 2 {
				cell.SetSelectable(true)
				// the ref stores the PomodoroCommand type for sending
				// updates via the command channel:
				cell.SetReference(b[row]["onchange"])
			} else {
				cell.SetSelectable(false)
			}
			view.SetCell(row, col, cell)
		}
	}
}

// TODO: move this to util.
func attachTicker(timer chan time.Time, interval time.Duration) {
	timer <- time.Now() // send one tick immediately.
	t := time.NewTicker(interval)
	for c := range t.C {
		timer <- c
	}
}

func handleAction(action string, pom pomodoro, command chan pomodoroCommand) {
	switch action {
	case "continue":
		command <- pomodoroCommand{commandtype: "continue"}
	case "create_pomodoro":
	case "create_break":
	case "cancel":
	case "delete_pomodoro":
	case "delete_break":
	}
}

func updateHeader(
	left *tview.TextView,
	center *tview.TextView,
	right *tview.TextView,
	pom *pomodoro,
) {
	tick := make(chan time.Time)
	go attachTicker(tick, time.Millisecond*200)

	for {
		<-tick
		timeleft := (*pom).durationLeft.Round(time.Second)
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

func getConfig() Config {
	var config Config
	home, _ := os.UserHomeDir()
	configpath := filepath.Join(home, configfolder, configname)

	file, _ := os.Open(configpath)
	configJSON, _ := ioutil.ReadAll(file)
	json.Unmarshal([]byte(configJSON), &config)

	return config
}

func editTableCell(
	pages *tview.Pages,
	table *tview.Table,
	command chan pomodoroCommand,
) {
	cell := table.GetCell(table.GetSelection())
	x, y, w := cell.GetLastPosition()
	inputField := tview.NewInputField()

	inputField.SetRect(x, y, w, 1)
	inputField.SetText(cell.Text)
	inputField.SetDoneFunc(
		func(key tcell.Key) {
			cell.SetText(inputField.GetText())
			// update pomodoro:
			command <- pomodoroCommand{
				commandtype: cell.GetReference().(string),
				payload:     inputField.GetText(),
			}
			pages.RemovePage("edit")
		},
	)

	pages.AddPage("edit", inputField, false, true)
}
