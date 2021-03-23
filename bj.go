package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)
type recStruct struct {
	JOBID string
	STAT string
	QUEUE string
	KILL_REASON string
	DEPENDENCY string
	EXIT_REASON string
	TIME_LEFT string
	COMPLETE string
	RUN_TIME string
	MAX_MEM string
	MEMLIMIT string
	NTHREADS string
	EXIT_CODE string
	}

type bjobsStruct struct {
  COMMAND string
  JOBS int
  RECORDS []recStruct
}



func main() {
	//code that generated the json
	//bjobs -a -json -o 'jobid stat queue kill_reason dependency exit_reason time_left %complete run_time max_mem memlimit nthreads exit_code'
	bjobsJson, _ := ioutil.ReadFile("example.json")

	//unmarshal (parse) json into struct
	var bjobs bjobsStruct
	json.Unmarshal([]byte(bjobsJson), &bjobs)


	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// get dimensions of current terminal window
	termWidth, termHeight := ui.TerminalDimensions()



	// Setup the main grid with columns for bjobs
	table1 := widgets.NewTable()
	table1.SetRect(0, 0, termWidth, termHeight-3)
	table1.TextAlignment = ui.AlignCenter
	table1.RowSeparator = false

	// set table headers
	table1.Rows = [][]string{[]string{"JOB ID", "STATUS", "QUEUE", "RAM USAGE", "%COMPLETE"}}
	table1.RowStyles[0] = ui.NewStyle(ui.ColorYellow, ui.ColorClear, ui.ModifierBold)

	// add job info to rows
	pend_jobs := 0
	run_jobs := 0
	done_jobs := 0
	exit_jobs := 0
	for i:=0; i < bjobs.JOBS; i++ {
		bjob := bjobs.RECORDS[i]

		switch bjob.STAT {
		case "PEND":
			pend_jobs++
		case "DONE":
			done_jobs++
			table1.Rows = append(table1.Rows, []string{bjob.JOBID, bjob.STAT, bjob.QUEUE })
			table1.RowStyles[(i+1)] = ui.NewStyle(ui.ColorGreen, ui.ColorClear)
		case "EXIT":
			exit_jobs++
			table1.Rows = append(table1.Rows, []string{bjob.JOBID, bjob.STAT, bjob.QUEUE })
			table1.RowStyles[(i+1)] = ui.NewStyle(ui.ColorRed, ui.ColorClear)
		case "RUN":
			run_jobs++
			table1.Rows = append(table1.Rows, []string{bjob.JOBID, bjob.STAT, bjob.QUEUE })
			table1.RowStyles[(i+1)] = ui.NewStyle(ui.ColorClear, ui.ColorClear)
		}

	}
	ui.Render(table1)



	stats_grid := ui.NewGrid()
	stats_grid.SetRect(0, termHeight-3, termWidth, termHeight-2)
	run_jobs_p := widgets.NewParagraph()
	run_jobs_p.Text = "Running: " + strconv.Itoa(run_jobs)
	run_jobs_p.Border = false
	pend_jobs_p := widgets.NewParagraph()
	pend_jobs_p.Text = "Pending: " + strconv.Itoa(pend_jobs)
	pend_jobs_p.Border = false
	done_jobs_p := widgets.NewParagraph()
	done_jobs_p.Text = "Done: " + strconv.Itoa(done_jobs)
	done_jobs_p.Border = false
	exit_jobs_p := widgets.NewParagraph()
	exit_jobs_p.Text = "Exited: " + strconv.Itoa(exit_jobs)
	exit_jobs_p.Border = false

	stats_grid.Set(ui.NewRow(1.0/1.0,
				ui.NewCol(1.0/4, run_jobs_p),
				ui.NewCol(1.0/4, pend_jobs_p),
				ui.NewCol(1.0/4, done_jobs_p),
				ui.NewCol(1.0/4, exit_jobs_p)))
	ui.Render(stats_grid)

	// Make grid layout for the buttons
	// on the bottom of the screen
	button_grid := ui.NewGrid()
	button_grid.SetRect(0, termHeight-2, termWidth, termHeight+1)

	quit_btn := widgets.NewParagraph()
	quit_btn.Text = "Quit [q] "
	quit_btn.Border = false
	quit_btn.TextStyle.Fg = ui.ColorClear

	email_btn := widgets.NewParagraph()
	email_btn.Text = "Email On All Ending [e] "
	email_btn.Border = false
	email_btn.TextStyle.Fg = ui.ColorClear

	killall_btn := widgets.NewParagraph()
	killall_btn.Text = "Kill All Jobs [k] "
	killall_btn.Border = false
	killall_btn.TextStyle.Fg = ui.ColorClear


	button_grid.Set(ui.NewRow(1.0/1.0,
				ui.NewCol(1.0/3, quit_btn),
				ui.NewCol(1.0/3, email_btn),
				ui.NewCol(1.0/3, killall_btn)))
	ui.Render(button_grid)


	//keep ui running until keystroke
	for e := range ui.PollEvents() {
		if e.ID == "q" {
			break
		}
		if e.ID == "e" {
			email_btn.TextStyle.Fg = ui.ColorGreen
			ui.Render(button_grid)
		}
		if e.ID == "k" {
			killall_btn.TextStyle.Fg = ui.ColorRed
			ui.Render(button_grid)
		}
	}

}

