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

	"github.com/fatih/color"
)

var (
	keepRecent  int
	keepDaily   int
	keepWeekly  int
	keepMonthly int
	keepYearly  int

	dryRunFlag     bool
	verboseOneFlag bool
	verboseTwoFlag bool
	helpFlag       bool

	backupDir string
)

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
	flag.BoolVar(&verboseOneFlag, "v", false, "Verbose")
	flag.BoolVar(&verboseTwoFlag, "vv", false, "Verbose")
	flag.BoolVar(&helpFlag, "h", false, "Help")

	flag.Parse()

	if verboseTwoFlag {
		verboseOneFlag = true
	}

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

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
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

	if dryRunFlag {
		fmt.Println("Dry run mode. Not deleting any files.")
	}

	for _, file := range files {
		keep := false

		// Check recent files
		if numRecent < keepRecent {
			keep = true
			numRecent++
		}

		// Check days
		if numDays < keepDaily && file.ModTime().Before(curDay) {
			keep = true

			numDays++
			curDay = curDay.AddDate(0, 0, -1)
		}

		// Check weeks
		if numWeeks < keepWeekly && file.ModTime().Before(curWeek) {
			keep = true

			numWeeks++
			curWeek = curWeek.AddDate(0, 0, -7)
		}

		// Check months
		if numMonths < keepMonthly && file.ModTime().Before(curMonth) {
			keep = true

			numMonths++
			curMonth = curMonth.AddDate(0, -1, 0)
		}

		// Check years
		if numYears < keepYearly && file.ModTime().Before(curYear) {
			keep = true

			numYears++
			curYear = curYear.AddDate(-1, 0, 0)
		}

		if keep {
			if verboseOneFlag || dryRunFlag {
				fmt.Println("[", color.BlueString("Keeping"), " ]", file.Name(), file.ModTime())
			}
		} else {
			if verboseTwoFlag {
				fmt.Println("[", color.RedString("Deleting"), "]", file.Name(), file.ModTime())
			}

			if !dryRunFlag {
				os.Remove(backupDir + "/" + file.Name())
			}
		}
	}
}
