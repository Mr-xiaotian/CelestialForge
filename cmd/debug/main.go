package main

import (
	"fmt"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func debug_info() {
	files, _ := file.GetFilesInfoRecursive(`Q:\Project\CelestialForge\tests\testdata`)
	for _, file := range files {
		fmt.Println(file)
	}
}

func debug_duplicate() {
	identicalMap, _ := file.GetDuplicateFile(`Q:\Project\CelestialForge\tests\testdata\duplicate`)
	identicalReport := file.DuplicateReport(identicalMap)
	fmt.Println(identicalReport)
}

func main() {
	// debug_info()
	debug_duplicate()
}
