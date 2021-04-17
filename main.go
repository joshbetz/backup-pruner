package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var (
	keepRecent  int
	keepDaily   int
	keepWeekly  int
	keepMonthly int
	keepYearly  int
	dryRunFlag  bool
	helpFlag    bool
	backupDir   string
)

type BackupFile struct {
	name    string
	modTime time.Time
	keep    bool
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <dir>\n", os.Args[0])
	flag.PrintDefaults()
}

func init() {
	flag.IntVar(&keepRecent, "keep-recent", 0, "Recent backups to keep. Default is 0")
	flag.IntVar(&keepDaily, "keep-daily", 0, "Daily backups to keep. Default is 0")
	flag.IntVar(&keepWeekly, "keep-weekly", 0, "Weekly backups to keep. Default is 0")
	flag.IntVar(&keepMonthly, "keep-monthly", 0, "Monthly backups to keep. Default is 0")
	flag.IntVar(&keepYearly, "keep-yearly", 0, "Yearly backups to keep. Default is 0")
	flag.BoolVar(&dryRunFlag, "dry-run", false, "Dry run mode")
	flag.BoolVar(&helpFlag, "h", false, "Help")

	flag.Parse()

	if helpFlag {
		usage()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
	}

	if keepRecent == 0 && keepDaily == 0 && keepWeekly == 0 && keepMonthly == 0 && keepYearly == 0 {
		fmt.Println("ERROR: Must specify some backups to keep")
		fmt.Println()

		usage()
		os.Exit(1)
	}

	var err error
	backupDir = flag.Arg(0)
	backupDir, err = filepath.Abs(backupDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}

func main() {
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		log.Fatal(err)
	}

	var backupFiles []BackupFile
	for _, file := range files {
		backupFiles = append(backupFiles, BackupFile{
			name:    file.Name(),
			modTime: file.ModTime(),
			keep:    false,
		})
	}

	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].modTime.After(backupFiles[j].modTime)
	})

	// Process files
	numRecent := 0

	numDays := 0
	curDay := time.Now()

	numWeeks := 0
	curWeek := time.Now()

	numMonths := 0
	curMonth := time.Now()

	numYears := 0
	curYear := time.Now()

	for i, file := range backupFiles {

		// Check recent files
		if numRecent < keepRecent {
			backupFiles[i].keep = true
			numRecent++
		}

		// Check days
		if numDays < keepDaily && file.modTime.Before(curDay) {
			backupFiles[i].keep = true

			numDays++
			curDay = curDay.AddDate(0, 0, -1)
		}

		// Check weeks
		if numWeeks < keepWeekly && file.modTime.Before(curWeek) {
			backupFiles[i].keep = true

			numWeeks++
			curWeek = curWeek.AddDate(0, 0, -7)
		}

		// Check months
		if numMonths < keepMonthly && file.modTime.Before(curMonth) {
			backupFiles[i].keep = true

			numMonths++
			curMonth = curMonth.AddDate(0, -1, 0)
		}

		// Check years
		if numYears < keepYearly && file.modTime.Before(curYear) {
			backupFiles[i].keep = true

			numYears++
			curYear = curYear.AddDate(-1, 0, 0)
		}
	}

	if dryRunFlag {
		fmt.Println("Dry run mode. Not deleting any files.")
	}

	for _, file := range backupFiles {
		if file.keep {
			fmt.Println(file.name, file.modTime)
		} else if !dryRunFlag {
			os.Remove(backupDir + "/" + file.name)
		}
	}
}
