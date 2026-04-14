package grow

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

type FailRecord struct {
	FormatTime   string
	ExecutorName string
	TaskID       int
	TaskValue    any
	ErrorMessage string
}

type FailRecordHandler struct {
	FailPath string
	FailFile *os.File
}

func (f *FailRecordHandler) BeforeStart() error {
	var err error

	today := time.Now().Format("2006-01-02")
	now := time.Now().Format("15-04-05.000")
	f.FailPath = fmt.Sprintf("fallback/%s/grow_fail(%s).jsonl", today, now)
	if err = os.MkdirAll(fmt.Sprintf("fallback/%s", today), 0755); err != nil {
		return fmt.Errorf("创建失败记录目录失败: %w", err)
	}

	f.FailFile, err = os.OpenFile(f.FailPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开失败记录文件失败: %w", err)
	}

	return nil
}

func (f *FailRecordHandler) HandleRecord(record FailRecord) error {
	var err error
	var b []byte

	b, err = json.Marshal(record)
	if err != nil {
		return fmt.Errorf("序列化失败记录失败: %w", err)
	}

	_, err = f.FailFile.Write(append(b, '\n'))
	if err != nil {
		return fmt.Errorf("写入失败记录文件失败: %w", err)
	}

	return err
}

func (f *FailRecordHandler) AfterStop() error {
	err := f.FailFile.Close()
	if err != nil {
		return fmt.Errorf("关闭失败记录文件失败: %w", err)
	}

	return nil
}

// FailInlet 失败记录生产端，内嵌 Inlet 并提供领域方法
type FailInlet struct {
	funnel.Inlet[FailRecord]
}

func NewFailInlet(ch chan<- FailRecord, timeout time.Duration) *FailInlet {
	return &FailInlet{
		Inlet: *funnel.NewInlet(ch, timeout),
	}
}

func (f *FailInlet) TaskError(executorName string, taskID int, task any, err error) {
	f.Send(FailRecord{
		FormatTime:   time.Now().Format("2006-01-02 15:04:05"),
		ExecutorName: executorName,
		TaskID:       taskID,
		TaskValue:    task,
		ErrorMessage: fmt.Sprintf("%v", err),
	})
}
