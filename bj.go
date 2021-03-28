package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

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
	COMPLETE string `json:"%COMPLETE"`
	RUN_TIME string
	MAX_MEM string
	MEMLIMIT string
	NTHREADS string
	EXIT_CODE string
	}

	func (rec recStruct) mem_usage()  string {
		max_mem := rec.MAX_MEM
		memlimit := rec.MEMLIMIT
		memusage := ""
        if max_mem != "" {
            max_mem = parse_bytes_output(max_mem)
            memlimit = parse_bytes_output(memlimit)
			memusage = max_mem + "/" + memlimit
		}
		return memusage
	}

	func (rec recStruct) atmemlimit() bool {
		max_mem_str := rec.MAX_MEM
		memlimit_str := rec.MEMLIMIT
		atlimit := false
		if max_mem_str != "" {
			max_mem_str = parse_bytes_output(max_mem_str)
			max_mem := parse_human_sizes(max_mem_str)

			memlimit_str = parse_bytes_output(memlimit_str)
			memlimit := parse_human_sizes(memlimit_str)

			if max_mem/memlimit > 0.9 {
				atlimit = true
			}
		}
		return atlimit
	}


func parse_bytes_output(bytes_string string) string {
	bytes_string = strings.ReplaceAll(bytes_string, "Gbytes","G")
	bytes_string = strings.ReplaceAll(bytes_string, "Mbytes","M")
	bytes_string = strings.ReplaceAll(bytes_string, "Kbytes","K")
	bytes_string = strings.ReplaceAll(bytes_string, " ","")

	return bytes_string
}

func parse_human_sizes(human_size_str string) float64 {
	human_size_str = strings.Replace(human_size_str, "G", "000000000", 1)
	human_size_str = strings.Replace(human_size_str, "M", "000000", 1)
	human_size_str = strings.Replace(human_size_str, "K", "000", 1)
	human_size_str = strings.ReplaceAll(human_size_str, " ","",)
	machine_readable_size, _ := strconv.ParseFloat(human_size_str, 64)

	return machine_readable_size
}

func run_bjobs() map[string]recStruct {
	var bjobs_cmd *exec.Cmd

	//if projectBool {
		//bjobs_cmd = exec.Command("bjobs","-Jd",proj_name,"-a","-json","-o","jobid stat queue kill_reason dependency exit_reason time_left %complete run_time max_mem memlimit nthreads exit_code")
	//} else {
		//bjobs_cmd = exec.Command("bjobs","-a","-json","-o","jobid stat queue kill_reason dependency exit_reason time_left %complete run_time max_mem memlimit nthreads exit_code")
	//}
	bjobs_cmd = exec.Command("cat","example.json")

	// 1. fetch current bjobs from shell
	bjobsJson, err := bjobs_cmd.Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1) // if problem with bjobs command then stop here
	}

	// 2. get 'RECORDS' part of JSON
	bjobs_json_raw := make(map[string]json.RawMessage)
	json.Unmarshal([]byte(bjobsJson), &bjobs_json_raw)
	bjobs_recs := bjobs_json_raw["RECORDS"]

	// 2. parse records into a list of bjob structures
	var bjList []recStruct
	json.Unmarshal([]byte(bjobs_recs), &bjList)

	// 3. convert list of records to map for easy lookup by JOBID
	bj_map := make(map[string]recStruct)
	for _, bj := range bjList {
		bj_map[bj.JOBID] = bj
	}


	return bj_map
}

func writeDatabase(usr_home string, usr_config string, db map[string]recStruct) {
	os.MkdirAll(usr_home, 0755)
	b, err := json.Marshal(db)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ioutil.WriteFile(usr_config, b, 0644)
}


func readSavedDatabase(usr_config string) map[string]recStruct{
	var db map[string]recStruct
	if _, err := os.Stat(usr_config); ! os.IsNotExist(err) {
		savedDatabaseJson, _ := ioutil.ReadFile(usr_config)

		json.Unmarshal([]byte(savedDatabaseJson), &db)
	}

	return db
}


func updateDatabase(db map[string]recStruct, bjobs_map map[string]recStruct) map[string]recStruct {
	// get list of jobids whose data needs udpating
	updateList := make([]string, len(bjobs_map))
	i := 0
	for id, new_job := range bjobs_map {
		// when no entry exists for new job id add it to list
		if _, ok := db[id]; !ok {
			updateList[i] = id
		} else {
			// check if jobs are identical
			if (new_job != db[id]) {
				updateList[i] = id
			}
		}
		i++
	}

	for i := 0; i < len(updateList); i++ {
		if (updateList[i] != "") {
			db[updateList[i]] = bjobs_map[updateList[i]]
		}
	}

	return db
}

// set job counts / statistics line
func statsGrid(pend_jobs int, done_jobs int, exit_jobs int) {
	stats_grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	stats_grid.SetRect(0, termHeight-3, termWidth, termHeight-2)
	run_jobs_p := widgets.NewParagraph()
	//run_jobs_p.Text = "Running: " + strconv.Itoa(run_jobs)
	randgen := rand.New(rand.NewSource(time.Now().UnixNano()))
	run_jobs_p.Text = "Running: " + strconv.Itoa(randgen.Intn(200))
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
}

func refreshInterface(db map[string]recStruct,table2 **widgets.Table) {
	table1 := widgets.NewTable()
	termWidth, termHeight := ui.TerminalDimensions()
	table1.SetRect(0, 0, termWidth, termHeight-3)
	table1.TextAlignment = ui.AlignCenter
	table1.RowSeparator = false

	// set table headers
	table1.Rows = [][]string{[]string{"JOB ID", "STATUS", "QUEUE", "RAM USAGE", "%Time Limit"}}
	table1.RowStyles[0] = ui.NewStyle(ui.ColorYellow, ui.ColorClear, ui.ModifierBold)


	// fetch and parse output from bjobs command
	bjobs_map := run_bjobs()

	// merge with existing cache
	db = updateDatabase(db, bjobs_map)

	// set initial counts
	var all_run_jobs_list []string
	var exit_jobs_list []string
	var done_jobs_list []string
	var remaining_run_jobs_list []string
	pend_jobs := 0
	run_jobs := 0
	done_jobs := 0
	exit_jobs := 0
	// add job info to rows
	for _, bjob := range db {
		switch bjob.STAT {
		case "PEND":
			pend_jobs++
		case "DONE":
			done_jobs++
			done_jobs_list = append(done_jobs_list, bjob.JOBID)
		case "EXIT":
			exit_jobs++
			exit_jobs_list = append(exit_jobs_list, bjob.JOBID)
		case "RUN":
			run_jobs++
			all_run_jobs_list = append(all_run_jobs_list, bjob.JOBID)
		}
	}
	sort.Strings(done_jobs_list)
	sort.Strings(exit_jobs_list)

	sort.Strings(all_run_jobs_list)
	for _, id := range all_run_jobs_list {
		job := db[id]
		completion_perc, _ := strconv.Atoi(strings.Replace(job.COMPLETE, "% L","",1))
		if completion_perc > 95 {
			table1 = danger_alert(table1, db, id, "nearly at time limit")
		} else if job.atmemlimit() {
			table1 = danger_alert(table1, db, id, "at memory limit")
		} else {
			remaining_run_jobs_list = append(remaining_run_jobs_list, id)
		}
	}
	sort.Strings(remaining_run_jobs_list)

	for _, id := range remaining_run_jobs_list {
		table1.Rows = append(table1.Rows, []string{db[id].JOBID, db[id].STAT, db[id].QUEUE, db[id].mem_usage(), strings.Replace(db[id].COMPLETE, " L","",1)})
		table1.RowStyles[(len(table1.Rows)-1)] = ui.NewStyle(ui.Color(248), ui.ColorClear)
	}
	for _, id := range exit_jobs_list {
		table1.Rows = append(table1.Rows, []string{db[id].JOBID, db[id].STAT, db[id].QUEUE, db[id].mem_usage(), db[id].EXIT_REASON})
		table1.RowStyles[(len(table1.Rows)-1)] = ui.NewStyle(ui.ColorRed, ui.ColorClear)
	}
	for _, id := range done_jobs_list {
		table1.Rows = append(table1.Rows, []string{db[id].JOBID, db[id].STAT, db[id].QUEUE, db[id].mem_usage()})
		table1.RowStyles[(len(table1.Rows)-1)] = ui.NewStyle(ui.ColorGreen, ui.ColorClear)
	}



	statsGrid(pend_jobs, done_jobs, exit_jobs)
	ui.Render(table1) // display constructed table
}


func danger_alert(table1 *widgets.Table, db map[string]recStruct, id string, alert string) *widgets.Table {
	table1.Rows = append(table1.Rows, []string{db[id].JOBID, db[id].STAT, db[id].QUEUE, "Job is "+alert})
	table1.RowStyles[(len(table1.Rows)-1)] = ui.NewStyle(ui.Color(197), ui.ColorClear, ui.ModifierUnderline)
	return table1
}



func main() {
	// initiate default values to be later changed by different user interactions
	projectBool := false
	proj_name := ""
	kill_menu := false
	email_on := false

	// load config and cached job information
	usr_home, _ := os.UserHomeDir()
	usr_home = usr_home + "/.config/better-bjobs/"
	usr_config := usr_home + "savedDatabase.json"


	if len(os.Args) > 2 {
	fmt.Println("Error more than one argument passed, give zero arguments to select all bjobs or one argument to specify a specific project name")
		os.Exit(1)
	} else if len(os.Args) == 2 {
		proj_name = os.Args[1]
		projectBool = true
	}

	// start curses terminal interface
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// get dimensions of current terminal window
	termWidth, termHeight := ui.TerminalDimensions()



	// load config with previous session data
	db := readSavedDatabase(usr_config)



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
	quit_btn.TextStyle.Fg = ui.Color(248)

	email_btn := widgets.NewParagraph()
	email_btn.Text = "Email On All Ending [e] "
	email_btn.Border = false
	email_btn.TextStyle.Fg = ui.Color(248)

	killall_btn := widgets.NewParagraph()
	killall_btn.Text = "Kill All Jobs [k] "
	killall_btn.Border = false
	killall_btn.TextStyle.Fg = ui.Color(248)

	clear_btn := widgets.NewParagraph()
	clear_btn.Text = "Clear Job Cache [c] "
	clear_btn.Border = false
	clear_btn.TextStyle.Fg = ui.Color(248)

	button_grid.Set(ui.NewRow(1.0/1.0,
				ui.NewCol(1.0/4, quit_btn),
				ui.NewCol(1.0/4, email_btn),
				ui.NewCol(1.0/4, killall_btn),
				ui.NewCol(1.0/4, clear_btn)))
	ui.Render(button_grid)



	job_table := widgets.NewTable()
	job_table.SetRect(0, 0, termWidth, termHeight-3)
	job_table.TextAlignment = ui.AlignCenter
	job_table.RowSeparator = false

	// set table headers
	job_table.Rows = [][]string{[]string{"JOB ID", "STATUS", "QUEUE", "RAM USAGE", "%Time Limit"}}
	job_table.RowStyles[0] = ui.NewStyle(ui.ColorYellow, ui.ColorClear, ui.ModifierBold)

	refreshInterface(db, &job_table)


	// setup keyboard input to process user actions
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(500*time.Millisecond).C // update interface every second
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
				clear_btn.Text = "Clearing cached job info"
				ui.Render(button_grid)

				// replace savedDatabase with an empty one on pressing clear
				var emptyDB map[string]recStruct
				writeDatabase(usr_home, usr_config, emptyDB)
				//ui.Render(table1)
				//refreshInterface(db)

				// pause long enough for user to see whats happening
				time.Sleep(1 * time.Second)

				clear_btn.TextStyle.Fg = ui.ColorClear
				clear_btn.Text = "Clear Job Cache [c] "
				ui.Render(button_grid)


			// re-render all elements on resizing terminal window
			//case "<Resize>":
				//payload := e.Payload.(ui.Resize)
				//table1.SetRect(0, 0, payload.Width, payload.Height)
				//ui.Clear()
				//ui.Render(table1)
				//ui.Render(button_grid)
				//refreshInterface(db)


			// loop over all cached bjob ids killing each one on "k"
			case "k":
				// specify that only project ids will be killed if we have a project subview
				projectText := ""
				if projectBool {
					projectText = " for project " + proj_name
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
					statusline.Text = "Killing all bjobs"
					ui.Render(statusline_grid)

					for jobid, _ := range db {
						_, err := exec.Command("bkill", jobid).Output()
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					}

					killall_btn.TextStyle.Fg = ui.ColorClear
					ui.Render(button_grid)
				}

			}
			case <-ticker:
				refreshInterface(db, &job_table)
		}
	}

}

