package file

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mr-xiaotian/CelestialForge/pkg/number"
	"github.com/Mr-xiaotian/CelestialForge/pkg/str"
)

type flowResult struct {
	path string
	hash string
}

type flowError struct {
	path string
	err  error
}

func worker(tasks <-chan string, results chan<- flowResult, errors chan<- flowError) {
	for task := range tasks {
		hash, err := GetFileSHA1(task)
		if err != nil {
			errors <- flowError{path: task, err: err}
		} else {
			results <- flowResult{path: task, hash: hash}
		}
	}
}

func processSuccess(result flowResult) []string {
	return []string{result.path, result.hash}
}

func GetDuplicateFile(path string) (map[FileInfo][]string, error) {
	fileInfoMap, err := GetFilesInfoRecursive(path)
	if err != nil {
		return nil, err
	}

	// 这里根据文件的大小来初步判断是否可能是重复文件
	fileSizeMap := make(map[int64][]string)
	for path, fileInfo := range fileInfoMap {
		fileSizeMap[fileInfo.Size] = append(fileSizeMap[fileInfo.Size], path)
	}

	var fileSizeDuplicates []string
	for _, paths := range fileSizeMap {
		if len(paths) > 1 {
			for _, path := range paths {
				fileSizeDuplicates = append(fileSizeDuplicates, path)
			}
		}
	}

	// 利用hash来进行二次判断
	tasks := make(chan string, len(fileSizeDuplicates))
	results := make(chan flowResult, len(fileSizeDuplicates))
	errors := make(chan flowError, len(fileSizeDuplicates))
	numWorkers := 3

	// 启动工作协程
	for i := 0; i < numWorkers; i++ {
		go worker(tasks, results, errors)
	}

	// 发送任务
	for _, path := range fileSizeDuplicates {
		tasks <- path
	}
	close(tasks)

	// 收集结果
	fileHashMap := map[string][]string{}
	for i := 0; i < len(fileSizeDuplicates); i++ {
		select {
		case res := <-results:
			fileHashMap[res.hash] = append(fileHashMap[res.hash], res.path)
		case err := <-errors:
			return nil, err.err
		}
	}

	fileHashDuplicates := map[FileInfo][]string{}
	for hash, paths := range fileHashMap {
		if len(paths) > 1 {
			fileInfo := FileInfo{
				Hash: hash,
				Size: fileInfoMap[paths[0]].Size,
			}
			fileHashDuplicates[fileInfo] = paths
		}
	}

	return fileHashDuplicates, nil
}

// DuplicateReport 生成重复文件的详细报告
func DuplicateReport(identicalMap map[FileInfo][]string) string {
	if len(identicalMap) == 0 {
		fmt.Println("\nNo identical items found.")
		return ""
	}

	// 按总占用大小降序排序
	type entry struct {
		info  FileInfo
		paths []string
	}
	entries := make([]entry, 0, len(identicalMap))
	for info, paths := range identicalMap {
		entries = append(entries, entry{info: info, paths: paths})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].info.Size*int64(len(entries[i].paths)) >
			entries[j].info.Size*int64(len(entries[j].paths))
	})

	var report []string
	var totalSize int64
	var totalItemNum int
	var maxItemNum int
	var maxItemEntry entry

	report = append(report, "\nIdentical items found:\n")

	for idx, e := range entries {
		itemNum := len(e.paths)
		itemsSize := e.info.Size * int64(itemNum)
		totalSize += itemsSize
		totalItemNum += itemNum

		if itemNum > maxItemNum {
			maxItemNum = itemNum
			maxItemEntry = e
		}

		data := make([][]string, len(e.paths))
		for i, p := range e.paths {
			data[i] = []string{p, number.HumanBytes(e.info.Size)}
		}
		tableText := str.FormatTable(data, []string{"Item", "Size"})

		report = append(report, fmt.Sprintf("%d.Hash: %s (Size: %s)", idx, e.info.Hash, number.HumanBytes(itemsSize)))
		report = append(report, tableText)
	}

	report = append(report, fmt.Sprintf("Total size of duplicate items: %s", number.HumanBytes(totalSize)))
	report = append(report, fmt.Sprintf("Total number of duplicate items: %d", totalItemNum))
	report = append(report, fmt.Sprintf("Item with the most duplicates: %s(hash) %s(size) %d(number)",
		maxItemEntry.info.Hash, number.HumanBytes(maxItemEntry.info.Size), maxItemNum))

	return strings.Join(report, "\n")
}
