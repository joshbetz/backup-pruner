package main

import (
	"flag"
	"fmt"
	"io/fs"
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
	curDate := time.Now()
	numRecent := 0
	numDays := 0
	numWeeks := 0
	numMonths := 0
	numYears := 0

	if dryRunFlag {
		fmt.Println("Dry run mode. Not deleting any files.")
	}

	for _, file := range files {

		// Check recent files
		if numRecent < keepRecent {
			numRecent++
			curDate = file.ModTime()

			printKeeping(file, color.RedString("(Recent) "))
			continue
		}

		// Check days
		if numDays < keepDaily {
			if file.ModTime().Before(curDate.AddDate(0, 0, -1)) {
				numDays++
				curDate = file.ModTime()

				printKeeping(file, color.YellowString("(Daily)  "))
				continue
			}

			deleteFile(file)
			continue
		}

		// Check weeks
		if numWeeks < keepWeekly {
			if file.ModTime().Before(curDate.AddDate(0, 0, -7)) {
				numWeeks++
				curDate = file.ModTime()

				printKeeping(file, color.CyanString("(Weekly) "))
				continue
			}

			deleteFile(file)
			continue
		}

		// Check months
		if numMonths < keepMonthly {
			if file.ModTime().Before(curDate.AddDate(0, -1, 0)) {
				numMonths++
				curDate = file.ModTime()

				printKeeping(file, color.BlueString("(Monthly)"))
				continue
			}

			deleteFile(file)
			continue
		}

		// Check years
		if numYears < keepYearly {
			if file.ModTime().Before(curDate.AddDate(-1, 0, 0)) {
				numYears++
				curDate = file.ModTime()

				printKeeping(file, color.MagentaString("(Yearly) "))
				continue
			}

			deleteFile(file)
			continue
		}
	}
}

func printKeeping(file fs.FileInfo, detail string) {
	if verboseOneFlag || dryRunFlag {
		fmt.Println("[", "Keeping", detail, "]", file.Name(), file.ModTime())
	}
}

func deleteFile(file fs.FileInfo) {
	// Delete
	if verboseTwoFlag {
		fmt.Println("[", color.RedString("Deleting         "), "]", file.Name(), file.ModTime())
	}

	if !dryRunFlag {
		os.Remove(backupDir + "/" + file.Name())
	}
}
