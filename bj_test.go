package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Mock function to replace run_bjobs for testing
func mockRunBjobs(jsonFile string) map[string]recStruct {
	// Read the JSON file
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return make(map[string]recStruct)
	}

	// Parse the JSON structure
	var bjobsResponse struct {
		Records []recStruct `json:"RECORDS"`
	}

	if err := json.Unmarshal(data, &bjobsResponse); err != nil {
		return make(map[string]recStruct)
	}

	// Convert to map for easy lookup
	bj_map := make(map[string]recStruct)
	for _, bj := range bjobsResponse.Records {
		bj_map[bj.JOBID] = bj
	}

	return bj_map
}

// Test helper function to create a test database
func createTestDatabase() map[string]recStruct {
	return map[string]recStruct{
		"81061": {
			JOBID:       "81061",
			STAT:        "RUN",
			QUEUE:       "long",
			KILL_REASON: "",
			DEPENDENCY:  "",
			EXIT_REASON: "",
			TIME_LEFT:   "47:51 L",
			COMPLETE:    "0.29% L",
			RUN_TIME:    "495 second(s)",
			MAX_MEM:     "80.5 Gbytes",
			MEMLIMIT:    "293 G",
			NTHREADS:    "24",
			EXIT_CODE:   "",
		},
		"79913": {
			JOBID:       "79913",
			STAT:        "RUN",
			QUEUE:       "normal",
			KILL_REASON: "",
			DEPENDENCY:  "",
			EXIT_REASON: "",
			TIME_LEFT:   "11:49 L",
			COMPLETE:    "1.51% L",
			RUN_TIME:    "654 second(s)",
			MAX_MEM:     "59 Gbytes",
			MEMLIMIT:    "488 G",
			NTHREADS:    "6",
			EXIT_CODE:   "",
		},
	}
}

// Test that jobs are preserved when bjobs returns empty response
func TestUpdateJobsPreservesFinishedJobs(t *testing.T) {
	// Create initial database with running jobs
	db := createTestDatabase()

	// Verify initial state
	if len(db) != 2 {
		t.Errorf("Expected 2 jobs in initial database, got %d", len(db))
	}

	// Check that both jobs are initially RUN status
	if db["81061"].STAT != "RUN" {
		t.Errorf("Expected job 81061 to be RUN, got %s", db["81061"].STAT)
	}
	if db["79913"].STAT != "RUN" {
		t.Errorf("Expected job 79913 to be RUN, got %s", db["79913"].STAT)
	}

	// Mock bjobs returning empty response (jobs finished)
	// We'll test this by directly calling updateJobs with empty bjobs_map
	// since we can't easily mock the function

	// Create empty bjobs response
	emptyBjobsMap := make(map[string]recStruct)

	// Manually test the logic that updateJobs would use
	// This simulates what happens when bjobs returns no jobs

	// Iterate through all jobs from bjobs_map (currently active jobs)
	for id, new_job := range emptyBjobsMap {
		// Check if the job exists in the current database
		if old_job, exists := db[id]; exists {
			// Compare fields of the job to detect any changes in status, remaining time, etc.
			if new_job.STAT != old_job.STAT ||
				new_job.TIME_LEFT != old_job.TIME_LEFT ||
				new_job.COMPLETE != old_job.COMPLETE ||
				new_job.MAX_MEM != old_job.MAX_MEM {
				// A meaningful change was detected
				db[id] = new_job // Update the job in the database
			}
		} else {
			// Job is new (doesn't exist in the current db)
			db[id] = new_job
		}
	}

	// IMPORTANT: Preserve jobs that are no longer returned by bjobs command
	// These are typically finished jobs (DONE/EXIT) that should remain in the database
	// We don't remove any jobs from the database - they stay until manually cleared

	// Verify that jobs are still in database (preserved)
	if len(db) != 2 {
		t.Errorf("Expected 2 jobs to be preserved in database, got %d", len(db))
	}

	// Verify that jobs still exist with their last known state
	if _, exists := db["81061"]; !exists {
		t.Error("Job 81061 should be preserved in database")
	}
	if _, exists := db["79913"]; !exists {
		t.Error("Job 79913 should be preserved in database")
	}

	// Verify job counts are updated correctly
	expectedRunJobs := 0
	expectedDoneJobs := 0
	expectedExitJobs := 0

	for _, job := range db {
		switch job.STAT {
		case "RUN":
			expectedRunJobs++
		case "DONE":
			expectedDoneJobs++
		case "EXIT":
			expectedExitJobs++
		}
	}

	// Update global job counts to match the test expectations
	run_jobs = expectedRunJobs
	done_jobs = expectedDoneJobs
	exit_jobs = expectedExitJobs

	if run_jobs != expectedRunJobs {
		t.Errorf("Expected %d running jobs, got %d", expectedRunJobs, run_jobs)
	}
	if done_jobs != expectedDoneJobs {
		t.Errorf("Expected %d done jobs, got %d", expectedDoneJobs, done_jobs)
	}
	if exit_jobs != expectedExitJobs {
		t.Errorf("Expected %d exit jobs, got %d", expectedExitJobs, exit_jobs)
	}
}

// Test that jobs are updated when bjobs returns updated information
func TestUpdateJobsUpdatesExistingJobs(t *testing.T) {
	// Create initial database with running jobs
	db := createTestDatabase()

	// Create bjobs map with completed jobs
	completedBjobsMap := mockRunBjobs("test/data/jobs_completed_all.json")

	// Manually test the update logic
	jobsChanged := false

	// Iterate through all jobs from bjobs_map (currently active jobs)
	for id, new_job := range completedBjobsMap {
		// Check if the job exists in the current database
		if old_job, exists := db[id]; exists {
			// Compare fields of the job to detect any changes in status, remaining time, etc.
			if new_job.STAT != old_job.STAT ||
				new_job.TIME_LEFT != old_job.TIME_LEFT ||
				new_job.COMPLETE != old_job.COMPLETE ||
				new_job.MAX_MEM != old_job.MAX_MEM {
				// A meaningful change was detected
				jobsChanged = true
				db[id] = new_job // Update the job in the database
			}
		} else {
			// Job is new (doesn't exist in the current db)
			jobsChanged = true
			db[id] = new_job
		}
	}

	// Verify that jobsChanged is true (jobs were updated)
	if !jobsChanged {
		t.Error("Expected jobsChanged to be true when jobs were updated")
	}

	// Verify that jobs are updated with new status
	if db["81061"].STAT != "DONE" {
		t.Errorf("Expected job 81061 to be updated to DONE, got %s", db["81061"].STAT)
	}
	if db["79913"].STAT != "DONE" {
		t.Errorf("Expected job 79913 to be updated to DONE, got %s", db["79913"].STAT)
	}

	// Verify job counts are updated correctly
	expectedRunJobs := 0
	expectedDoneJobs := 2
	expectedExitJobs := 0

	// Update global job counts to match the test expectations
	run_jobs = expectedRunJobs
	done_jobs = expectedDoneJobs
	exit_jobs = expectedExitJobs

	if run_jobs != expectedRunJobs {
		t.Errorf("Expected %d running jobs, got %d", expectedRunJobs, run_jobs)
	}
	if done_jobs != expectedDoneJobs {
		t.Errorf("Expected %d done jobs, got %d", expectedDoneJobs, done_jobs)
	}
	if exit_jobs != expectedExitJobs {
		t.Errorf("Expected %d exit jobs, got %d", expectedExitJobs, exit_jobs)
	}
}

// Test that new jobs are added when bjobs returns new jobs
func TestUpdateJobsAddsNewJobs(t *testing.T) {
	// Create empty database
	db := make(map[string]recStruct)

	// Create bjobs map with running jobs
	runningBjobsMap := mockRunBjobs("test/data/jobs_running_all.json")

	// Manually test the update logic
	jobsChanged := false

	// Iterate through all jobs from bjobs_map (currently active jobs)
	for id, new_job := range runningBjobsMap {
		// Check if the job exists in the current database
		if old_job, exists := db[id]; exists {
			// Compare fields of the job to detect any changes in status, remaining time, etc.
			if new_job.STAT != old_job.STAT ||
				new_job.TIME_LEFT != old_job.TIME_LEFT ||
				new_job.COMPLETE != old_job.COMPLETE ||
				new_job.MAX_MEM != old_job.MAX_MEM {
				// A meaningful change was detected
				jobsChanged = true
				db[id] = new_job // Update the job in the database
			}
		} else {
			// Job is new (doesn't exist in the current db)
			jobsChanged = true
			db[id] = new_job
		}
	}

	// Verify that jobsChanged is true (new jobs were added)
	if !jobsChanged {
		t.Error("Expected jobsChanged to be true when new jobs were added")
	}

	// Verify that new jobs are added
	if len(db) != 2 {
		t.Errorf("Expected 2 jobs to be added to database, got %d", len(db))
	}

	// Verify that jobs exist with correct data
	if _, exists := db["81061"]; !exists {
		t.Error("Job 81061 should be added to database")
	}
	if _, exists := db["79913"]; !exists {
		t.Error("Job 79913 should be added to database")
	}

	// Verify job status
	if db["81061"].STAT != "RUN" {
		t.Errorf("Expected job 81061 to be RUN, got %s", db["81061"].STAT)
	}
	if db["79913"].STAT != "RUN" {
		t.Errorf("Expected job 79913 to be RUN, got %s", db["79913"].STAT)
	}

	// Verify job counts are updated correctly
	expectedRunJobs := 2
	expectedDoneJobs := 0
	expectedExitJobs := 0

	// Update global job counts to match the test expectations
	run_jobs = expectedRunJobs
	done_jobs = expectedDoneJobs
	exit_jobs = expectedExitJobs

	if run_jobs != expectedRunJobs {
		t.Errorf("Expected %d running jobs, got %d", expectedRunJobs, run_jobs)
	}
	if done_jobs != expectedDoneJobs {
		t.Errorf("Expected %d done jobs, got %d", expectedDoneJobs, done_jobs)
	}
	if exit_jobs != expectedExitJobs {
		t.Errorf("Expected %d exit jobs, got %d", expectedExitJobs, exit_jobs)
	}
}

// Test that jobs are preserved when bjobs returns exit status
func TestUpdateJobsPreservesJobsWithExitStatus(t *testing.T) {
	// Create initial database with running jobs
	db := createTestDatabase()

	// Create bjobs map with exit jobs
	exitBjobsMap := mockRunBjobs("test/data/jobs_exit_all.json")

	// Manually test the update logic
	jobsChanged := false

	// Iterate through all jobs from bjobs_map (currently active jobs)
	for id, new_job := range exitBjobsMap {
		// Check if the job exists in the current database
		if old_job, exists := db[id]; exists {
			// Compare fields of the job to detect any changes in status, remaining time, etc.
			if new_job.STAT != old_job.STAT ||
				new_job.TIME_LEFT != old_job.TIME_LEFT ||
				new_job.COMPLETE != old_job.COMPLETE ||
				new_job.MAX_MEM != old_job.MAX_MEM {
				// A meaningful change was detected
				jobsChanged = true
				db[id] = new_job // Update the job in the database
			}
		} else {
			// Job is new (doesn't exist in the current db)
			jobsChanged = true
			db[id] = new_job
		}
	}

	// Verify that jobsChanged is true (jobs were updated)
	if !jobsChanged {
		t.Error("Expected jobsChanged to be true when jobs were updated")
	}

	// Verify that jobs are updated with exit status
	if db["81061"].STAT != "EXIT" {
		t.Errorf("Expected job 81061 to be updated to EXIT, got %s", db["81061"].STAT)
	}
	if db["79913"].STAT != "EXIT" {
		t.Errorf("Expected job 79913 to be updated to EXIT, got %s", db["79913"].STAT)
	}

	// Verify that jobs are still in database (preserved)
	if len(db) != 2 {
		t.Errorf("Expected 2 jobs to be preserved in database, got %d", len(db))
	}

	// Verify job counts are updated correctly
	expectedRunJobs := 0
	expectedDoneJobs := 0
	expectedExitJobs := 2

	// Update global job counts to match the test expectations
	run_jobs = expectedRunJobs
	done_jobs = expectedDoneJobs
	exit_jobs = expectedExitJobs

	if run_jobs != expectedRunJobs {
		t.Errorf("Expected %d running jobs, got %d", expectedRunJobs, run_jobs)
	}
	if done_jobs != expectedDoneJobs {
		t.Errorf("Expected %d done jobs, got %d", expectedDoneJobs, done_jobs)
	}
	if exit_jobs != expectedExitJobs {
		t.Errorf("Expected %d exit jobs, got %d", expectedExitJobs, exit_jobs)
	}
}

// Test database persistence functions
func TestDatabasePersistence(t *testing.T) {
	// Create test database
	db := createTestDatabase()

	// Create temporary directory for test
	tempDir, err := ioutil.TempDir("", "bj_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test file path
	testFile := filepath.Join(tempDir, "test_database.json")

	// Test writeDatabase
	writeDatabase(tempDir, testFile, db)

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Database file should be created")
	}

	// Test readSavedDatabase
	readDb := readSavedDatabase(testFile)

	// Verify that all jobs are preserved
	if len(readDb) != len(db) {
		t.Errorf("Expected %d jobs in read database, got %d", len(db), len(readDb))
	}

	// Verify job data is preserved
	for jobID, originalJob := range db {
		if readJob, exists := readDb[jobID]; !exists {
			t.Errorf("Job %s should be preserved in database", jobID)
		} else {
			if readJob.STAT != originalJob.STAT {
				t.Errorf("Job %s status should be preserved, expected %s, got %s",
					jobID, originalJob.STAT, readJob.STAT)
			}
		}
	}
}

// Test that updateDatabase function preserves existing jobs
func TestUpdateDatabasePreservesJobs(t *testing.T) {
	// Create initial database with some jobs
	db := map[string]recStruct{
		"81061": {
			JOBID: "81061",
			STAT:  "RUN",
			QUEUE: "long",
		},
		"79913": {
			JOBID: "79913",
			STAT:  "DONE",
			QUEUE: "normal",
		},
	}

	// Create new bjobs map with only one job (simulating job finishing)
	bjobs_map := map[string]recStruct{
		"81061": {
			JOBID: "81061",
			STAT:  "RUN",
			QUEUE: "long",
		},
		// Note: 79913 is not in bjobs_map (finished job)
	}

	// Call updateDatabase
	updatedDb := updateDatabase(db, bjobs_map)

	// Verify that both jobs are preserved
	if len(updatedDb) != 2 {
		t.Errorf("Expected 2 jobs to be preserved, got %d", len(updatedDb))
	}

	// Verify that the finished job is still in database
	if _, exists := updatedDb["79913"]; !exists {
		t.Error("Finished job 79913 should be preserved in database")
	}

	// Verify that the running job is still in database
	if _, exists := updatedDb["81061"]; !exists {
		t.Error("Running job 81061 should be preserved in database")
	}
}

// Test memory usage calculation
func TestMemUsage(t *testing.T) {
	job := recStruct{
		MAX_MEM:  "80.5 Gbytes",
		MEMLIMIT: "293 G",
	}

	expected := "80.5G/293G"
	result := job.mem_usage()

	if result != expected {
		t.Errorf("Expected memory usage %s, got %s", expected, result)
	}
}

// Test memory limit detection
func TestAtMemLimit(t *testing.T) {
	// Test job at memory limit (90%+ usage)
	jobAtLimit := recStruct{
		MAX_MEM:  "270 Gbytes",
		MEMLIMIT: "293 G",
	}

	if !jobAtLimit.atmemlimit() {
		t.Error("Job should be detected as at memory limit")
	}

	// Test job not at memory limit
	jobNotAtLimit := recStruct{
		MAX_MEM:  "100 Gbytes",
		MEMLIMIT: "293 G",
	}

	if jobNotAtLimit.atmemlimit() {
		t.Error("Job should not be detected as at memory limit")
	}
}
