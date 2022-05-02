package main

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)

func spawnTUI() {
	app := tview.NewApplication()

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	frame := tview.NewFrame(layout)
	header := tview.NewFlex()
	headerleft := tview.NewTextView()
	headercenter := tview.NewTextView()
	headerright := tview.NewTextView()
	body := tview.NewTable()
	footer := tview.NewTextView()

	frame.AddText(" P 🐕 M ", true, tview.AlignCenter, tcell.ColorOlive)
	frame.SetBorders(0, 0, 0, 0, 0, 0)

	layout.AddItem(header, 1, 0, false)
	layout.AddItem(body, 0, 1, true)
	layout.AddItem(footer, 1, 0, false)

	header.SetBorderPadding(0, 0, 2, 2)
	header.AddItem(headerleft, 0, 1, false)
	header.AddItem(headercenter	, 0, 1, false)
	header.AddItem(headerright, 9, 0, false)

	body.SetBorder(true)

	fmt.Fprint(headerleft, "WORKING 15:04/25:00")
	fmt.Fprint(headerright, "[09:56]")
	fmt.Fprint(footer, "[0]  1   2")

	app.SetRoot(frame, true)
	app.SetFocus(body).Run()
}
