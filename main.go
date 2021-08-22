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

type backup struct {
	file fs.FileInfo
	keep bool
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

	if keepRecent < 3 {
		// Keep at least 3 just to be safe
		keepRecent = 3
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

	if dryRunFlag {
		fmt.Println("Dry run mode. Not deleting any files.")
	}

	// Initialize backups
	candidates := make([]*backup, 0, len(files))
	for _, file := range files {
		candidates = append(candidates, &backup{
			file: file,
		})
	}

	// Mark backups to keep
	keep(candidates, keepRecent, nil)
	keep(candidates, keepDaily, func(backup *backup) string {
		return backup.file.ModTime().Format("2006-01-02")
	})
	keep(candidates, keepWeekly, func(backup *backup) string {
		yy, ww := backup.file.ModTime().ISOWeek()
		return fmt.Sprintf("%04d-%02d", yy, ww)
	})
	keep(candidates, keepMonthly, func(backup *backup) string {
		return backup.file.ModTime().Format("2006-01")
	})
	keep(candidates, keepYearly, func(backup *backup) string {
		return backup.file.ModTime().Format("2006")
	})

	// Process all backups
	for _, backup := range candidates {
		if backup.keep {
			fmt.Println("[", color.BlueString("Keeping"), " ]", backup.file.Name(), backup.file.ModTime())
		} else if !dryRunFlag {
			fmt.Println("[", color.RedString("Removing"), " ]", backup.file.Name(), backup.file.ModTime())
			os.Remove(backupDir + "/" + backup.file.Name())
		}
	}
}

func keep(candidates []*backup, max int, compare func(*backup) string) {
	if max < 1 {
		return
	}

	grouped := make(map[string]*backup, max*2)
	for _, backup := range candidates {
		if backup.keep {
			continue
		}

		key := backup.file.Name()
		if compare != nil {
			key = compare(backup)
		}

		previous := grouped[key]
		if previous == nil || previous.file.ModTime().Before(backup.file.ModTime()) {
			grouped[key] = backup
		}
	}

	finalCandidates := make([]*backup, 0, len(grouped))
	for _, backup := range grouped {
		finalCandidates = append(finalCandidates, backup)
	}

	if max > len(finalCandidates) {
		// keep all
		//
		// no need to sort through the results and only keep "max"
		// we're going to keep them all anyway
		for _, backup := range finalCandidates {
			backup.keep = true
		}

		return
	}

	sort.Slice(finalCandidates, func(i, j int) bool {
		return finalCandidates[i].file.ModTime().After(finalCandidates[j].file.ModTime())
	})

	for i := 0; i < max; i++ {
		finalCandidates[i].keep = true
	}
}
