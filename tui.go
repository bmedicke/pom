package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bmedicke/bhdr/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

//go:embed chordmap.json
var chordmapJSON string

func spawnTUI(config Config) {
	app := tview.NewApplication()
	pom := createPomodoro(config)

	// used for updating the pomodoro from goroutines:
	command := make(chan pomodoroCommand)

	// vim-style key chords:
	chord := util.KeyChord{Active: false, Buffer: "", Action: ""}
	chordmap := map[string]interface{}{}
	json.Unmarshal([]byte(chordmapJSON), &chordmap)

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
	header.AddItem(headerright, 24, 0, false)

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
				editTableCell(body, bodytable, command, "append_cell")
			}

			if chord.Active {
				util.HandleChords(event.Rune(), &chord, chordmap)
				handleChordAction(chord.Action, command, body, bodytable)
			} else {
				switch event.Rune() {
				case 'a', 'A':
					editTableCell(body, bodytable, command, "append_cell")
				case ';': // continue with next state:
					command <- pomodoroCommand{commandtype: "continue"}
				case 'q', 'Q': // quit the app:
					command <- pomodoroCommand{commandtype: "quit_app"}
				case 'c', 'd': // start chord:
					util.HandleChords(event.Rune(), &chord, chordmap)
				}
			}
			statusbar.SetText(chord.Buffer)
			return event
		},
	)

	createBodytable(bodytable, config)
	go handlePomodoroState(&pom, statusbar, app, command, config)
	go updateHeader(headerleft, headercenter, headerright, &pom)

	app.SetRoot(frame, true)
	app.SetFocus(bodytable).Run()
}

func createBodytable(bodytable *tview.Table, config Config) {
	b := []map[string]string{
		{
			"id":       "projekt",
			"onchange": "update_project",
			"type":     "editable",
			"value":    config.DefaultProject,
		},
		{
			"id":       "task",
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
			bodytable.SetCell(row, col, cell)
		}
	}
}

func handleChordAction(
	action string,
	command chan pomodoroCommand,
	body *tview.Pages,
	bodytable *tview.Table,
) {
	switch action {
	case "change_cell", "delete_cell":
		editTableCell(body, bodytable, command, action)
	}
}

func updateHeader(
	left *tview.TextView,
	center *tview.TextView,
	right *tview.TextView,
	pom *pomodoro,
) {
	tick := make(chan time.Time)
	go util.AttachTicker(tick, time.Millisecond*200)

	for {
		<-tick
		timeleft := (*pom).durationLeft.Round(time.Second)
		timer := fmt.Sprintf("[%6v]", timeleft)
		right.SetText(fmt.Sprintf("%12v %10v", (*pom).State, timer))

		var color tcell.Color
		switch (*pom).State {
		case "work":
			color = tcell.ColorDarkOliveGreen
		case "break":
			color = tcell.ColorBlue
		case "longbreak":
			color = tcell.ColorDarkBlue
		case "break_done", "longbreak_done", "work_done":
			color = tcell.ColorDarkRed
		case "ready":
			color = tcell.ColorBlack
		}

		left.SetBackgroundColor(color)
		center.SetBackgroundColor(color)
		right.SetBackgroundColor(color)
	}
}

func editTableCell(
	pages *tview.Pages,
	table *tview.Table,
	command chan pomodoroCommand,
	action string,
) {
	cell := table.GetCell(table.GetSelection())
	x, y, w := cell.GetLastPosition()
	inputField := tview.NewInputField()
	inputField.SetRect(x, y, w, 1)

	switch action {
	case "append_cell":
		inputField.SetText(cell.Text)
	case "change_cell":
		inputField.SetText("")
	case "delete_cell":
		cell.SetText("")
		command <- pomodoroCommand{
			commandtype: cell.GetReference().(string), payload: "",
		}
		return
	}

	inputField.SetDoneFunc(
		func(key tcell.Key) {
			cell.SetText(inputField.GetText())
			pages.RemovePage("edit")
			command <- pomodoroCommand{
				commandtype: cell.GetReference().(string),
				payload:     inputField.GetText(),
			}
		},
	)

	pages.AddPage("edit", inputField, false, true)
}
