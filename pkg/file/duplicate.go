package file

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
	"github.com/Mr-xiaotian/CelestialForge/pkg/str"
	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

// executor
type SnapshotExecutor struct {
	*grow.Executor[string, string]
}
type HashExecutor struct {
	*grow.Executor[string, string]
}

func NewSnapshotExecutor(processor func(string) (string, error), numWorkers int, observers ...grow.Observer) *SnapshotExecutor {
	executor := &SnapshotExecutor{
		Executor: grow.NewExecutor("SnapshotExecutor", processor, numWorkers, observers...),
	}
	return executor
}
func NewHashExecutor(processor func(string) (string, error), numWorkers int, observers ...grow.Observer) *HashExecutor {
	executor := &HashExecutor{
		Executor: grow.NewExecutor("HashExecutor", processor, numWorkers, observers...),
	}
	return executor
}

// func
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

func getSnapshotDuplicate(fileSizeDuplicates []string, numWorkers int) ([]string, error) {
	// 利用短hash(4KB)来进行二次判断
	origin := make(map[int]string, len(fileSizeDuplicates))
	for i, path := range fileSizeDuplicates {
		origin[i] = path
	}

	// 并行计算文件hash
	executor := NewSnapshotExecutor(GetFileSnapshotSHA1, numWorkers, grow.NewProgressBar("Snapshoting files"))
	go executor.Start(fileSizeDuplicates)

	// 收集结果
	fileSnapshotMap := map[string][]string{}
	executor.Drain(func(res grow.Payload[string]) {
		fileSnapshotMap[res.Value] = append(fileSnapshotMap[res.Value], origin[res.ID])
	})

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

func getHashDuplicate(fileSnapshotDuplicates []string, fileInfoMap FileInfoMap, numWorkers int) (map[FileInfo][]string, error) {
	// 利用hash来进行三次判断
	origin := make(map[int]string, len(fileSnapshotDuplicates))
	for i, path := range fileSnapshotDuplicates {
		origin[i] = path
	}

	// 并行计算文件hash
	executor := NewHashExecutor(GetFileSHA1, numWorkers, grow.NewProgressBar("Hashing files"))
	go executor.Start(fileSnapshotDuplicates)

	// 收集结果
	fileHashMap := map[string][]string{}
	executor.Drain(func(res grow.Payload[string]) {
		fileHashMap[res.Value] = append(fileHashMap[res.Value], origin[res.ID])
	})
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
