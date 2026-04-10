package main

import (
	"flag"
	"os"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func main() {
	path := flag.String("path", "", "path to scan")
	output := flag.String("output", "duplicate_report.txt", "output file path")
	flag.Parse()

	identicalMap, _ := file.GetDuplicateFile(*path)
	identicalReport := file.DuplicateReport(identicalMap)
	// fmt.Println(identicalReport)

	os.WriteFile(*output, []byte(identicalReport), 0644)
}
