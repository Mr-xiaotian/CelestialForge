package main

import (
	"fmt"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
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

func debug_bytes() {
	a := units.HumanBytes(1500 * 1024 * 1024) // 1.5 GB
	b := 500 * units.HumanBytes(1024)         // 500 MB

	sum := a.Add(b)
	fmt.Printf("a = %s\n", a)     // 1.50 GB
	fmt.Printf("b = %s\n", b)     // 500.00 MB
	fmt.Printf("sum = %s\n", sum) // 1.95 GB

	// 原始值
	fmt.Printf("bytes = %d\n", sum.Int64()) // 2097152000
}

func debug_size() {
	dirSize, err := file.GetDirSize(`D:\Project\CelestialForge`)
	if err != nil {
		fmt.Printf("GetDirSize error: %v\n", err)
		return
	}
	fmt.Printf("dirSize = %s\n", dirSize)
}

func debug_mtime() {
	dirMtime, err := file.GetDirMtime(`D:\Project\CelestialForge`)
	if err != nil {
		fmt.Printf("GetDirMtime error: %v\n", err)
		return
	}
	fmt.Printf("dirMtime = %s\n", dirMtime)
}

func main() {
	// debug_info()
	// debug_duplicate()
	// debug_bytes()
	// debug_size()
	debug_mtime()
}
