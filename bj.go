package main

import (
	"encoding/json"
	"os/exec"
	"time"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

type jobLog struct {
	NOTIFY bool
	IDS []string
	LENGTH int
	BJOBS []recStruct
}


func writeDatabase(usr_home string, usr_config string, db jobLog) {
	os.MkdirAll(usr_home, 0755)
	b, _ := json.Marshal(db)
	ioutil.WriteFile(usr_config, b, 0644)
}


func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func readSavedDatabase(usr_config string) jobLog {

	var savedDatabase jobLog
	if _, err := os.Stat(usr_config); ! os.IsNotExist(err) {
		savedDatabaseJson, _ := ioutil.ReadFile(usr_config)

		json.Unmarshal([]byte(savedDatabaseJson), &savedDatabase)
	}
	return savedDatabase
}

func updateDatabase(currentDB *jobLog, BJob *bjobsStruct) {

	for i:=0; i < BJob.JOBS; i++ {
		if ! stringInSlice(BJob.RECORDS[i].JOBID ,currentDB.IDS) {
			currentDB.BJOBS = append(currentDB.BJOBS, BJob.RECORDS[i])
			currentDB.IDS = append(currentDB.IDS, BJob.RECORDS[i].JOBID)
			currentDB.LENGTH = currentDB.LENGTH+1
		}
	}

}



func main() {
	usr_home, _ := os.UserHomeDir()
	usr_home = usr_home + "/.config/better-bjobs/"
	usr_config := usr_home + "savedDatabase.json"

	// initiate default values to be later changed by different user interactions
	projectBool := false
	kill_menu := false
	email_on := false

	bjobs_cmd := exec.Command("bjobs","-a","-json","-o","jobid stat queue kill_reason dependency exit_reason time_left %complete run_time max_mem memlimit nthreads exit_code")
	bjobsJson, err := bjobs_cmd.Output()

	// if problem with bjobs then stop here
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}


	//unmarshal (parse) json into struct
	var bjobs bjobsStruct
	json.Unmarshal([]byte(bjobsJson), &bjobs)


	// load config with previous session data
	db := readSavedDatabase(usr_config)
	if db.LENGTH > 0 {
		//fmt.Println(db)
		updateDatabase(&db, &bjobs)
	} else {
		db.LENGTH = bjobs.JOBS
		db.BJOBS = bjobs.RECORDS
	}

	// start curses terminal interface
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// get dimensions of current terminal window
	termWidth, termHeight := ui.TerminalDimensions()



	// Setup the main table with columns for bjobs
	table1 := widgets.NewTable()
	table1.SetRect(0, 0, termWidth, termHeight-3)
	table1.TextAlignment = ui.AlignCenter
	table1.RowSeparator = false

	// set table headers
	table1.Rows = [][]string{[]string{"JOB ID", "STATUS", "QUEUE", "RAM USAGE", "%COMPLETE"}}
	table1.RowStyles[0] = ui.NewStyle(ui.ColorYellow, ui.ColorClear, ui.ModifierBold)

	// set initial counts
	pend_jobs := 0
	run_jobs := 0
	done_jobs := 0
	exit_jobs := 0

	// add job info to rows
	for i:=0; i < db.LENGTH; i++ {
		bjob := db.BJOBS[i]

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
	ui.Render(table1) // display constructed table



	// set job counts / statistics line
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


	// set statusline to be same location as buttons
	// so as to hide the buttons when we display a status
	statusline_grid := ui.NewGrid()
	statusline_grid.SetRect(0, termHeight-2, termWidth, termHeight+1)
	statusline := widgets.NewParagraph()
	statusline.Border = false
	statusline_grid.Set(ui.NewRow(1.0/1.0, statusline))


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

	clear_btn := widgets.NewParagraph()
	clear_btn.Text = "Clear Job Cache [c] "
	clear_btn.Border = false
	clear_btn.TextStyle.Fg = ui.ColorClear

	button_grid.Set(ui.NewRow(1.0/1.0,
				ui.NewCol(1.0/4, quit_btn),
				ui.NewCol(1.0/4, email_btn),
				ui.NewCol(1.0/4, killall_btn),
				ui.NewCol(1.0/4, clear_btn)))
	ui.Render(button_grid)



	// setup keyboard input to process user actions
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {

			// quit on pressing q or contrl-c
			case "q", "<C-c>":
			writeDatabase(usr_home, usr_config, db)
				return

			// TODO: fix this to actually send email on completion
			case "e":
				email_on = !(email_on)
				if email_on {
					email_btn.TextStyle.Fg = ui.ColorGreen
					email_btn.Text = "Email notification scheduled"
				} else {
					email_btn.TextStyle.Fg = ui.ColorClear
					email_btn.Text = "Email On All Ending [e] "
				}
				ui.Render(button_grid)

			// clear the cache of saved jobs
			case "c", "<C-l>":
				clear_btn.TextStyle.Fg = ui.ColorYellow
				clear_btn.Text = "Clearing cahed job info"
				ui.Render(button_grid)

				// replace savedDatabase with an empty one on pressing clear
				var emptyDB jobLog
				writeDatabase(usr_home, usr_config, emptyDB)
				ui.Render(table1)

				// pause long enough for user to see whats happening
				time.Sleep(1 * time.Second)

				clear_btn.TextStyle.Fg = ui.ColorClear
				clear_btn.Text = "Clear Job Cache [c] "
				ui.Render(button_grid)


			// re-render all elements on resizing terminal window
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				table1.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(table1)
				ui.Render(button_grid)
				ui.Render(stats_grid)


			// loop over all cached bjob ids killing each one on "k"
			case "k":
				// specify that only project ids will be killed if we have a project subview
				projectText := ""
				if projectBool {
					projectText = " for this project"
				}

				kill_menu = true
				statusline.Text = "Are you sure you want to kill all unfinished bjobs"+projectText+"? [Yn] "
				statusline.TextStyle.Fg = ui.ColorRed

				ui.Render(statusline_grid)


			// manage yes and no prompts initiated by other cases
			case "n":
				if kill_menu {
					// if we say no to all-kill menu then reset statusline and put back buttons
					statusline.TextStyle.Fg = ui.ColorClear
					statusline.Text = ""
					ui.Render(button_grid)
				}

			case "y":
				if kill_menu {
					// if we say yes to all-kill menu then alert user
					statusline.Text = "KILLING ALL BJOBS.."
					ui.Render(statusline_grid)

					for jobnb:=0; jobnb < len(db.IDS); jobnb++ {
						_, err := exec.Command("bkill", db.IDS[jobnb]).Output()
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					}

					killall_btn.TextStyle.Fg = ui.ColorClear
					ui.Render(button_grid)
				}

			}
		}
	}

}

