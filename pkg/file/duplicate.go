package file

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
	"github.com/Mr-xiaotian/CelestialForge/pkg/str"
	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

// ==== Pipeline Stages ====

// getSizeDuplicate 按文件大小分组，返回存在大小重复的文件路径。
func getSizeDuplicate(fileInfoMap FileInfoMap) []string {
	fileSizeMap := make(map[units.HumanBytes][]string)
	for path, fileInfo := range fileInfoMap {
		fileSizeMap[fileInfo.Size] = append(fileSizeMap[fileInfo.Size], path)
	}

	var fileSizeDuplicates []string
	for size, paths := range fileSizeMap {
		if len(paths) > 1 && size > 0 {
			for _, path := range paths {
				fileSizeDuplicates = append(fileSizeDuplicates, path)
			}
		}
	}
	return fileSizeDuplicates
}

// getSnapshotDuplicate 用文件前 4KB 的快照哈希进一步过滤重复候选。
func getSnapshotDuplicate(fileSizeDuplicates []string, numWorkers int) ([]string, error) {
	// 并行计算文件hash
	executor := grow.NewExecutor("SnapshotExecutor", GetFileSnapshotSHA1, numWorkers, grow.NewProgressBar("Snapshoting files"))
	results := executor.Start(fileSizeDuplicates)

	// 收集结果
	fileSnapshotMap := map[string][]string{}
	for _, res := range results {
		fileSnapshotMap[res.Result] = append(fileSnapshotMap[res.Result], res.Task)
	}

	var fileSnapshotDuplicates []string
	for _, paths := range fileSnapshotMap {
		if len(paths) > 1 {
			for _, path := range paths {
				fileSnapshotDuplicates = append(fileSnapshotDuplicates, path)
			}
		}
	}
	return fileSnapshotDuplicates, nil
}

// getHashDuplicate 用完整文件哈希确认最终重复文件。
func getHashDuplicate(fileSnapshotDuplicates []string, fileInfoMap FileInfoMap, numWorkers int) (map[FileInfo][]string, error) {
	// 并行计算文件hash
	executor := grow.NewExecutor("HashExecutor", GetFileSHA1, numWorkers, grow.NewProgressBar("Hashing files"))
	results := executor.Start(fileSnapshotDuplicates)

	// 收集结果
	fileHashMap := map[string][]string{}
	for _, res := range results {
		fileHashMap[res.Result] = append(fileHashMap[res.Result], res.Task)
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

// ==== Public API ====

// ScanDuplicateFile 扫描目录下的重复文件。
// 通过三级过滤（大小 -> 快照哈希 -> 完整哈希）逐步缩小候选集。
func ScanDuplicateFile(path string, numWorkers int) (map[FileInfo][]string, error) {
	fileInfoMap, err := GetFilesInfoRecursive(path)
	if err != nil {
		return nil, err
	}
	// 这里根据文件的大小来初步判断是否可能是重复文件
	fileSizeDuplicates := getSizeDuplicate(fileInfoMap)

	// 利用短hash(4KB)来进行二次判断
	fileSnapshotDuplicates, err := getSnapshotDuplicate(fileSizeDuplicates, numWorkers)
	if err != nil {
		return nil, err
	}

	// 利用hash来进行三次判断
	fileHashDuplicates, err := getHashDuplicate(fileSnapshotDuplicates, fileInfoMap, numWorkers)
	if err != nil {
		return nil, err
	}

	return fileHashDuplicates, nil
}

// ==== Report ====

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
		return entries[i].info.Size*units.NewHumanBytes(int64(len(entries[i].paths))) >
			entries[j].info.Size*units.NewHumanBytes(int64(len(entries[j].paths)))
	})

	var report []string
	var totalSize units.HumanBytes
	var totalItemNum int
	var maxItemNum int
	var maxItemEntry entry

	report = append(report, "\nIdentical items found:\n")

	for idx, e := range entries {
		itemNum := len(e.paths)
		itemsSize := e.info.Size * units.NewHumanBytes(int64(itemNum))
		totalSize += itemsSize
		totalItemNum += itemNum

		if itemNum > maxItemNum {
			maxItemNum = itemNum
			maxItemEntry = e
		}

		data := make([][]string, len(e.paths))
		for i, p := range e.paths {
			data[i] = []string{p, e.info.Size.String()}
		}
		tableText := str.FormatTable(data, []string{"Item", "Size"})

		report = append(report, fmt.Sprintf("%d.Hash: %s (Size: %s)", idx, e.info.Hash, itemsSize))
		report = append(report, tableText)
	}

	report = append(report, fmt.Sprintf("Total size of duplicate items: %s", totalSize))
	report = append(report, fmt.Sprintf("Total number of duplicate items: %d", totalItemNum))
	report = append(report, fmt.Sprintf("Item with the most duplicates: %s(hash) %s(size) %d(number)",
		maxItemEntry.info.Hash, maxItemEntry.info.Size, maxItemNum))

	return strings.Join(report, "\n")
}
