package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func main() {
	scanPath := flag.String("scan-path", "", "path to scan")
	outputPath := flag.String("output-path", "duplicate_report.txt", "output file path")
	numTends := flag.Int("num-tends", 4, "number of tends to use")
	flag.Parse()

	var err error

	identicalMap, err := file.ScanDuplicateFile(*scanPath, *numTends)
	if err != nil {
		fmt.Println("Error scanning duplicate files:", err)
		return
	}
	identicalReport := file.DuplicateReport(identicalMap)

	err = os.WriteFile(*outputPath, []byte(identicalReport), 0644)
	if err != nil {
		fmt.Println("Error writing output file:", err)
		return
	}
}
